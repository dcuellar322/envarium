// Package repositoryfile owns bounded, rooted reads of untrusted repository files.
package repositoryfile

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Open rejects symlinks and special files, pins directory identities while
// walking, and verifies the opened descriptor. It never executes file content.
func Open(root, relative string, limit int64) (*os.File, error) {
	if limit < 1 || !filepath.IsLocal(relative) {
		return nil, errors.New("invalid repository file request")
	}
	relative = filepath.Clean(relative)
	directory, err := OpenDirectory(root, filepath.Dir(relative))
	if err != nil {
		return nil, err
	}
	defer func() { _ = directory.Close() }()
	return directory.open(filepath.Base(relative), limit)
}

// Read returns one complete bounded file, including a post-read growth check.
func Read(root, relative string, limit int64) ([]byte, error) {
	file, err := Open(root, relative, limit)
	if err != nil {
		return nil, err
	}
	return read(file, limit)
}

func read(file *os.File, limit int64) ([]byte, error) {
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
