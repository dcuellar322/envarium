//go:build !darwin && !linux

package repositoryfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type directoryHandle struct{ root *os.Root }

func openDirectory(root, relative string) (*directoryHandle, error) {
	directory, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	if relative != "." {
		for _, part := range strings.Split(relative, string(filepath.Separator)) {
			next, err := openChildDirectory(directory, part)
			_ = directory.Close()
			if err != nil {
				return nil, err
			}
			directory = next
		}
	}
	return &directoryHandle{root: directory}, nil
}
func openChildDirectory(parent *os.Root, name string) (*os.Root, error) {
	before, err := parent.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !before.IsDir() || before.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("repository path must contain real directories")
	}
	next, err := parent.OpenRoot(name)
	if err != nil {
		return nil, err
	}
	after, err := next.Stat(".")
	if err != nil || !os.SameFile(before, after) {
		_ = next.Close()
		return nil, errors.New("repository directory changed during open")
	}
	return next, nil
}
func (d *directoryHandle) open(name string, limit int64) (*os.File, error) {
	before, err := d.root.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() || before.Size() > limit {
		return nil, fmt.Errorf("repository source must be a regular file no larger than %d bytes", limit)
	}
	file, err := d.root.OpenFile(name, os.O_RDONLY|openSafetyFlags, 0)
	if err != nil {
		return nil, err
	}
	after, err := file.Stat()
	if err != nil || !after.Mode().IsRegular() || after.Size() > limit || !os.SameFile(before, after) {
		_ = file.Close()
		return nil, errors.New("repository file changed or is not a bounded regular file")
	}
	return file, nil
}
func (d *directoryHandle) close() error { return d.root.Close() }
