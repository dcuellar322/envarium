package adapters

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstalledGitPreservesCustomPathSelection(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX PATH/symlink fixture")
	}
	original, err := exec.LookPath("git")
	if err != nil {
		t.Skip("Git unavailable")
	}
	directory := t.TempDir()
	selected := filepath.Join(directory, "git")
	if err := os.Symlink(original, selected); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory)
	if got := newInstalledGit(); got.executable != selected {
		t.Fatalf("custom Git changed: %q", got.executable)
	}
}

func TestInstalledGitHandlesCancellationAndRemovedToolchain(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git unavailable")
	}
	runner := newInstalledGit()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := runner.Run(ctx, t.TempDir(), "--version"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
	runner.executable = filepath.Join(t.TempDir(), "removed-git")
	output, err := runner.Run(context.Background(), t.TempDir(), "--version")
	if err != nil || !strings.HasPrefix(string(output), "git version ") {
		t.Fatalf("fallback=%q err=%v", output, err)
	}
}
