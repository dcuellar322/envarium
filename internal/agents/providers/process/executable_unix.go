//go:build !windows

package process

import "os"

func executableCandidates(path string) (string, error) { return executableFile(path) }
func executableMode(mode os.FileMode, _ string) bool   { return mode.Perm()&0o111 != 0 }
