package repositoryfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSiblingReadsStayBoundedAndFreshAcrossObservations(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "inputs")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "portable"), []byte("first"), 0600); err != nil {
		t.Fatal(err)
	}
	directory, err := OpenDirectory(root, "inputs")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = directory.Close() }()
	if got, err := directory.Read("portable", 5); err != nil || string(got) != "first" {
		t.Fatalf("portable=%q err=%v", got, err)
	}
	if _, err := directory.Read("local", 5); !os.IsNotExist(err) {
		t.Fatalf("missing overlay=%v", err)
	}
	for _, name := range []string{"../inputs/portable", "inputs/portable", ".", "/dev/zero"} {
		if _, err := directory.Read(name, 5); err == nil {
			t.Errorf("accepted non-child %q", name)
		}
	}
	if _, err := directory.Read("portable", 4); err == nil {
		t.Fatal("accepted oversized source")
	}
	if err := os.WriteFile(filepath.Join(path, "portable"), []byte("fresh"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "local"), []byte("local"), 0600); err != nil {
		t.Fatal(err)
	}
	if got, err := directory.Read("portable", 5); err != nil || string(got) != "fresh" {
		t.Fatalf("cached old bytes=%q err=%v", got, err)
	}
	if got, err := directory.Read("local", 5); err != nil || string(got) != "local" {
		t.Fatalf("cached absence=%q err=%v", got, err)
	}
	if err := directory.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := directory.Read("portable", 5); err == nil {
		t.Fatal("read survived directory closure")
	}
}

func TestDirectoryRejectsSymlinkParentsAndSources(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "real"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "source"), []byte("valid"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "real"), filepath.Join(root, "alias")); err != nil {
		t.Skip(err)
	}
	if _, err := OpenDirectory(root, "alias"); err == nil {
		t.Fatal("accepted symlink directory")
	}
	if err := os.Symlink(filepath.Join(root, "source"), filepath.Join(root, "real", "alias")); err != nil {
		t.Fatal(err)
	}
	directory, err := OpenDirectory(root, "real")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = directory.Close() }()
	if _, err := directory.Read("alias", 5); err == nil {
		t.Fatal("accepted symlink source")
	}
}

func TestPinnedDirectoryCannotFollowReplacementAlias(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "inputs"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "inputs", "source"), []byte("approved"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "source"), []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	directory, err := OpenDirectory(root, "inputs")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = directory.Close() }()
	if err := os.Rename(filepath.Join(root, "inputs"), filepath.Join(root, "pinned")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "inputs")); err != nil {
		t.Skip(err)
	}
	contents, err := directory.Read("source", 20)
	if err != nil || string(contents) != "approved" {
		t.Fatalf("escaped pinned directory: %q %v", contents, err)
	}
	if replacement, err := OpenDirectory(root, "inputs"); err == nil {
		_ = replacement.Close()
		t.Fatal("followed replacement directory symlink")
	}
}
