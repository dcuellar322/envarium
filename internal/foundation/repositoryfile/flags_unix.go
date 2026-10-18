//go:build !windows

package repositoryfile

import "syscall"

const openSafetyFlags = syscall.O_NONBLOCK | syscall.O_NOFOLLOW
