//go:build !windows

package repositoryfile

import (
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestFIFOIsRejectedWithoutWaitingForWriter(t *testing.T) {
	root := t.TempDir()
	if err := syscall.Mkfifo(filepath.Join(root, "pipe"), 0600); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := Read(root, "pipe", 1024); done <- err }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("accepted FIFO")
		}
	case <-time.After(time.Second):
		t.Fatal("blocked opening FIFO")
	}
}
