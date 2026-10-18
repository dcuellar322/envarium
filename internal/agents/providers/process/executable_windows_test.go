//go:build windows

package process

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWindowsNativeExecutableResolution(t *testing.T) {
	directory := t.TempDir()
	executable := filepath.Join(directory, "provider.exe")
	if err := os.WriteFile(executable, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", ".;relative;"+directory)
	t.Setenv("PATHEXT", ".CMD;.EXE;.COM")
	expected, err := os.Stat(executable)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"provider", executable} {
		got, err := CanonicalExecutable(name)
		if err != nil {
			t.Fatalf("resolve %q: %v", name, err)
		}
		actual, statErr := os.Stat(got)
		if statErr != nil || !filepath.IsAbs(got) || !os.SameFile(expected, actual) {
			t.Fatalf("resolve %q=%q err=%v; expected file %q", name, got, statErr, executable)
		}
	}
	batch := filepath.Join(directory, "batch.cmd")
	if err := os.WriteFile(batch, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"batch", batch, `.\provider.exe`} {
		if _, err := CanonicalExecutable(name); err == nil {
			t.Fatalf("accepted unsafe executable %s", name)
		}
	}
}
