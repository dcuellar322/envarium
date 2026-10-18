// Package repositoryfile owns bounded, rooted reads of untrusted repository files.
package repositoryfile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Open rejects symlinks and special files, pins directory identities while
// walking, and verifies the opened descriptor. It never executes file content.
func Open(root, relative string, limit int64) (*os.File, error) {
	if limit < 1 || !filepath.IsLocal(relative) {
		return nil, errors.New("invalid repository file request")
	}
	directory, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer func() { _ = directory.Close() }()
	parts := strings.Split(filepath.Clean(relative), string(filepath.Separator))
	for _, part := range parts[:len(parts)-1] {
		before, err := directory.Lstat(part)
		if err != nil {
			return nil, err
		}
		if !before.IsDir() || before.Mode()&os.ModeSymlink != 0 {
			return nil, errors.New("repository path must contain real directories")
		}
		next, err := directory.OpenRoot(part)
		if err != nil {
			return nil, err
		}
		after, err := next.Stat(".")
		if err != nil || !os.SameFile(before, after) {
			_ = next.Close()
			return nil, errors.New("repository directory changed during open")
		}
		_ = directory.Close()
		directory = next
	}
	name := parts[len(parts)-1]
	before, err := directory.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() || before.Size() > limit {
		return nil, fmt.Errorf("repository source must be a regular file no larger than %d bytes", limit)
	}
	file, err := directory.OpenFile(name, os.O_RDONLY|openSafetyFlags, 0)
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

// Read returns one complete bounded file, including a post-read growth check.
func Read(root, relative string, limit int64) ([]byte, error) {
	file, err := Open(root, relative, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	contents, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(contents)) > limit {
		return nil, fmt.Errorf("repository source exceeds %d bytes", limit)
	}
	return contents, nil
}
