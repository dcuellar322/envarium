package adapters

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGitBranchSwitchPreservesIgnoredFilesAndDoesNotRunHooks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell hook fixture is Unix-only")
	}
	t.Parallel()
	root := t.TempDir()
	runGit(t, root, "init", "-b", "main")
	runGit(t, root, "config", "user.name", "Test")
	runGit(t, root, "config", "user.email", "test@example.invalid")
	write := func(name, contents string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, name), []byte(contents), 0700); err != nil {
			t.Fatal(err)
		}
	}
	write(".gitignore", "local.env\n")
	write("file", "main")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "main")
	runGit(t, root, "checkout", "-b", "feature")
	write("local.env", "target tracked value")
	runGit(t, root, "add", "-f", "local.env")
	runGit(t, root, "commit", "-m", "feature")
	runGit(t, root, "checkout", "main")
	write("local.env", "private local value")
	write(".git/hooks/post-checkout", "#!/bin/sh\ntouch hook-was-run\n")
	git := NewGit()
	if err := git.SwitchBranch(context.Background(), root, "feature"); err == nil {
		t.Fatal("ignored local file could be overwritten")
	}
	value, _ := os.ReadFile(filepath.Join(root, "local.env"))
	if string(value) != "private local value" {
		t.Fatal("ignored file modified")
	}
	if _, err := os.Stat(filepath.Join(root, "hook-was-run")); !os.IsNotExist(err) {
		t.Fatal("repository hook executed")
	}
	if err := os.Remove(filepath.Join(root, "local.env")); err != nil {
		t.Fatal(err)
	}
	if err := git.SwitchBranch(context.Background(), root, "feature"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "hook-was-run")); !os.IsNotExist(err) {
		t.Fatal("repository hook executed during successful switch")
	}
	runGit(t, root, "config", "filter.custom.smudge", "touch filter-was-run")
	if err := git.CheckSwitchSafety(context.Background(), root); err == nil {
		t.Fatal("checkout filter accepted")
	}
	if err := git.SwitchBranch(context.Background(), root, "main"); err == nil {
		t.Fatal("configured filter accepted during execution")
	}
	if _, err := os.Stat(filepath.Join(root, "filter-was-run")); !os.IsNotExist(err) {
		t.Fatal("configured filter executed")
	}
}
