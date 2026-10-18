//go:build windows

package process

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Only native executable formats may be launched without an explicit shell.
// PATHEXT commonly includes batch scripts; those are deliberately excluded.
func executableCandidates(path string) (string, error) {
	if filepath.Ext(path) != "" {
		return executableFile(path)
	}
	for _, extension := range filepath.SplitList(os.Getenv("PATHEXT")) {
		if !nativeExtension(extension) {
			continue
		}
		if resolved, err := executableFile(path + extension); err == nil {
			return resolved, nil
		}
	}
	for _, extension := range []string{".exe", ".com"} {
		if resolved, err := executableFile(path + extension); err == nil {
			return resolved, nil
		}
	}
	return "", errors.New("native provider executable not found")
}
func nativeExtension(extension string) bool {
	return strings.EqualFold(extension, ".exe") || strings.EqualFold(extension, ".com")
}
func executableMode(_ os.FileMode, path string) bool { return nativeExtension(filepath.Ext(path)) }
