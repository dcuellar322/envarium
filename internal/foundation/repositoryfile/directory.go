package repositoryfile

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// Directory pins a real directory for one bounded set of sibling file reads.
// Callers close it after the observation, not retain it between resolutions.
type Directory struct {
	mu     sync.RWMutex
	handle *directoryHandle
	closed bool
}

// OpenDirectory rejects symlink parents and pins the opened directory identity.
func OpenDirectory(root, relative string) (*Directory, error) {
	if !filepath.IsLocal(relative) {
		return nil, errors.New("invalid repository directory request")
	}
	handle, err := openDirectory(root, filepath.Clean(relative))
	if err != nil {
		return nil, err
	}
	return &Directory{handle: handle}, nil
}
func (d *Directory) open(name string, limit int64) (*os.File, error) {
	if d == nil {
		return nil, os.ErrClosed
	}
	if limit < 1 || !filepath.IsLocal(name) || name != filepath.Base(name) {
		return nil, errors.New("invalid repository file request")
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.closed || d.handle == nil {
		return nil, os.ErrClosed
	}
	return d.handle.open(name, limit)
}

// Read returns fresh, regular, bounded bytes from one direct child file.
func (d *Directory) Read(name string, limit int64) ([]byte, error) {
	file, err := d.open(name, limit)
	if err != nil {
		return nil, err
	}
	return read(file, limit)
}

// Close excludes concurrent opens before releasing the pinned directory handle.
func (d *Directory) Close() error {
	if d == nil {
		return os.ErrClosed
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed || d.handle == nil {
		return os.ErrClosed
	}
	d.closed = true
	return d.handle.close()
}
