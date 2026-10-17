package adapters

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"envarium.dev/envarium/internal/sourcecontrol/domain"
)

const maximumGitMetadata = 128

type gitMetadata struct {
	mu      sync.Mutex
	entries map[string]gitMetadataEntry
}
type gitMetadataEntry struct {
	marker     os.FileInfo
	markerText string
	directory  string
	head       string
	commit     *domain.Commit
}

func (g *Git) gitDirectory(ctx context.Context, root string) (string, error) {
	markerPath := filepath.Join(root, ".git")
	marker, statErr := os.Stat(markerPath)
	text := ""
	cacheable := statErr == nil && os.Getenv("GIT_DIR") == "" && os.Getenv("GIT_WORK_TREE") == "" && os.Getenv("GIT_COMMON_DIR") == ""
	if cacheable && !marker.IsDir() {
		raw, err := readGitMarker(markerPath, 4096)
		cacheable = err == nil && len(raw) < 4096
		text = string(raw)
	}
	if cacheable && g.metadata != nil {
		g.metadata.mu.Lock()
		entry, ok := g.metadata.entries[root]
		g.metadata.mu.Unlock()
		if ok && entry.marker != nil && os.SameFile(marker, entry.marker) && text == entry.markerText {
			return entry.directory, nil
		}
	}
	raw, err := g.runner.Run(ctx, root, "rev-parse", "--absolute-git-dir", "--is-inside-work-tree")
	if err != nil {
		return "", err
	}
	parts := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(parts) != 2 || parts[1] != "true" {
		return "", nil
	}
	directory := parts[0]
	if cacheable && g.metadata != nil && directory != "" {
		g.metadata.mu.Lock()
		if len(g.metadata.entries) >= maximumGitMetadata {
			clear(g.metadata.entries)
		}
		g.metadata.entries[root] = gitMetadataEntry{marker: marker, markerText: text, directory: directory}
		g.metadata.mu.Unlock()
	}
	return directory, nil
}

func (g *Git) commitAtHead(ctx context.Context, root, head string) *domain.Commit {
	if g.metadata == nil || head == "" || head == "(initial)" || !commitMemoAllowed(root) {
		return g.lastCommit(ctx, root)
	}
	g.metadata.mu.Lock()
	entry, exists := g.metadata.entries[root]
	if exists && entry.head == head && entry.commit != nil {
		commit := *entry.commit
		g.metadata.mu.Unlock()
		return &commit
	}
	g.metadata.mu.Unlock()
	commit := g.lastCommit(ctx, root)
	if commit != nil && commit.Hash == head && len(commit.Subject)+len(commit.Author)+len(commit.Hash)+len(commit.ShortHash) <= 8192 && ctx.Err() == nil && exists {
		g.metadata.mu.Lock()
		latest := g.metadata.entries[root]
		if latest.directory == entry.directory {
			commitCopy := *commit
			latest.head, latest.commit = head, &commitCopy
			g.metadata.entries[root] = latest
		}
		g.metadata.mu.Unlock()
	}
	return commit
}

func readGitMarker(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	return io.ReadAll(io.LimitReader(file, limit+1))
}

// Replacement objects and grafts can change metadata without changing HEAD.
// Preserve installed Git semantics by falling back to a fresh log read.
func commitMemoAllowed(root string) bool {
	directory := filepath.Join(root, ".git")
	info, err := os.Stat(directory)
	if err != nil || !info.IsDir() || os.Getenv("GIT_REPLACE_REF_BASE") != "" {
		return false
	}
	for _, name := range []string{"refs/replace", "info/grafts", "commondir"} {
		if _, err := os.Stat(filepath.Join(directory, name)); !os.IsNotExist(err) {
			return false
		}
	}
	raw, err := readGitMarker(filepath.Join(directory, "packed-refs"), 1<<20)
	return (os.IsNotExist(err) || err == nil && len(raw) <= 1<<20) && !bytes.Contains(raw, []byte("refs/replace/"))
}
