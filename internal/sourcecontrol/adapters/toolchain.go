package adapters

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Select the operator's Git once per observer. Apple's /usr/bin/git is a
// developer-tool launcher; resolve its selected toolchain instead of paying
// launcher discovery for every background status request. Custom PATH Git
// installations and developer-directory overrides retain their selection.
func newInstalledGit() installedGit {
	path, err := exec.LookPath("git")
	if err != nil {
		return installedGit{executable: "git"}
	}
	if runtime.GOOS != "darwin" || path != "/usr/bin/git" {
		return installedGit{executable: path}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	raw, err := exec.CommandContext(ctx, "/usr/bin/xcrun", "--find", "git").Output()
	if err != nil {
		return installedGit{executable: path}
	}
	selected := strings.TrimSuffix(string(raw), "\n")
	info, err := os.Stat(selected)
	if err != nil || !filepath.IsAbs(selected) || filepath.Base(selected) != "git" || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		return installedGit{executable: path}
	}
	return installedGit{executable: selected}
}
