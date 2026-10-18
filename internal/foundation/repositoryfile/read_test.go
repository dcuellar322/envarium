package repositoryfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadBoundsAndFileTypes(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "manifest"), []byte("valid"), 0600); err != nil {
		t.Fatal(err)
	}
	if got, err := Read(root, "manifest", 5); err != nil || string(got) != "valid" {
		t.Fatalf("read=%q err=%v", got, err)
	}
	for _, name := range []string{"manifest", ".", "../escape", "/dev/zero", "missing"} {
		if _, err := Read(root, name, 4); err == nil {
			t.Errorf("accepted %q", name)
		}
	}
	if err := os.Symlink(filepath.Join(root, "manifest"), filepath.Join(root, "alias")); err != nil {
		t.Skip(err)
	}
	if _, err := Read(root, "alias", 5); err == nil {
		t.Fatal("accepted file symlink")
	}
	if err := os.Mkdir(filepath.Join(root, "nested"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "nested", "manifest"), []byte("valid"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "nested"), filepath.Join(root, "directory")); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(root, "directory/manifest", 5); err == nil {
		t.Fatal("accepted directory symlink")
	}
}
