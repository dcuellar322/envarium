//go:build darwin || linux

package repositoryfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sys/unix"
)

type directoryHandle struct {
	fd      int
	cleanup runtime.Cleanup
}

func openDirectory(root, relative string) (*directoryHandle, error) {
	// The caller supplies the approved root. As with os.OpenRoot, that root
	// itself may be an alias; every repository-relative component must be real.
	fd, err := unix.Open(root, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open directory", Path: root, Err: err}
	}
	if relative != "." {
		for _, part := range strings.Split(relative, string(filepath.Separator)) {
			next, openErr := unix.Openat(fd, part, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
			_ = unix.Close(fd)
			if openErr != nil {
				return nil, &os.PathError{Op: "open repository directory", Path: relative, Err: openErr}
			}
			fd = next
		}
	}
	handle := &directoryHandle{fd: fd}
	handle.cleanup = runtime.AddCleanup(handle, func(fd int) { _ = unix.Close(fd) }, fd)
	return handle, nil
}

func (d *directoryHandle) open(name string, limit int64) (*os.File, error) {
	defer runtime.KeepAlive(d)
	fd := d.fd // Directory's read lock excludes Close/descriptor reuse.
	var before unix.Stat_t
	if err := unix.Fstatat(fd, name, &before, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return nil, &os.PathError{Op: "stat repository file", Path: name, Err: err}
	}
	if before.Mode&unix.S_IFMT != unix.S_IFREG || before.Size > limit {
		return nil, fmt.Errorf("repository source must be a regular file no larger than %d bytes", limit)
	}
	opened, err := unix.Openat(fd, name, unix.O_RDONLY|unix.O_CLOEXEC|openSafetyFlags, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open repository file", Path: name, Err: err}
	}
	file := os.NewFile(uintptr(opened), name)
	var after unix.Stat_t
	if err := unix.Fstat(opened, &after); err != nil || after.Mode&unix.S_IFMT != unix.S_IFREG || after.Size > limit || before.Dev != after.Dev || before.Ino != after.Ino {
		_ = file.Close()
		return nil, errors.New("repository file changed or is not a bounded regular file")
	}
	return file, nil
}
func (d *directoryHandle) close() error {
	d.cleanup.Stop()
	return unix.Close(d.fd)
}
