package adapters

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

type countingGitRunner struct {
	mu    sync.Mutex
	calls map[string]int
}

func (r *countingGitRunner) Run(ctx context.Context, root string, args ...string) ([]byte, error) {
	r.mu.Lock()
	r.calls[args[0]]++
	r.mu.Unlock()
	return (installedGit{}).Run(ctx, root, args...)
}

func TestMetadataReuseKeepsWorkingTreeRemotesAndOperationsFresh(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	runGit(t, root, "init", "-b", "main")
	runGit(t, root, "config", "user.name", "Envarium Test")
	runGit(t, root, "config", "user.email", "envarium@example.invalid")
	file := filepath.Join(root, "tracked.txt")
	if err := os.WriteFile(file, []byte("first\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "tracked.txt")
	runGit(t, root, "commit", "-m", "initial")
	observer := NewGit()
	runner := &countingGitRunner{calls: map[string]int{}}
	observer.runner = runner
	first, err := observer.Observe(context.Background(), "project", root)
	if err != nil {
		t.Fatal(err)
	}
	original, _ := os.Stat(file)
	if err := os.WriteFile(file, []byte("other\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(file, original.ModTime(), original.ModTime()); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "remote", "add", "origin", "https://example.invalid/repository")
	if err := os.WriteFile(filepath.Join(root, ".git", "MERGE_HEAD"), []byte(first.Head), 0o600); err != nil {
		t.Fatal(err)
	}
	first.LastCommit.Subject = "caller mutation"
	second, err := observer.Observe(context.Background(), "project", root)
	if err != nil {
		t.Fatal(err)
	}
	if second.Changes.Modified != 1 || len(second.Remotes) != 2 || second.OperationState != "merge" || second.LastCommit.Subject != "initial" {
		t.Fatalf("stale or shared result: %#v", second)
	}
	if runner.calls["rev-parse"] != 1 || runner.calls["log"] != 1 || runner.calls["status"] != 2 || runner.calls["worktree"] != 2 || runner.calls["remote"] != 2 {
		t.Fatalf("commands: %#v", runner.calls)
	}
	// Remove the synthetic marker before making a real new commit.
	if err := os.Remove(filepath.Join(root, ".git", "MERGE_HEAD")); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "tracked.txt")
	runGit(t, root, "commit", "-m", "next")
	next, err := observer.Observe(context.Background(), "project", root)
	if err != nil {
		t.Fatal(err)
	}
	if next.Head == first.Head || next.LastCommit.Subject != "next" || runner.calls["log"] != 2 {
		t.Fatalf("new HEAD not refreshed: %#v", next)
	}
}

func TestDirectoryObservationPreservesBareAndNonRepositoryBehavior(t *testing.T) {
	t.Parallel()
	for _, bare := range []bool{false, true} {
		root := t.TempDir()
		if bare {
			runGit(t, root, "init", "--bare")
		}
		state, err := NewGit().Observe(context.Background(), "project", root)
		if err != nil || state.Repository {
			t.Fatalf("bare=%v state=%#v error=%v", bare, state, err)
		}
	}
}

func TestCommitMetadataFallbackForReplacementObjects(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git", "refs", "replace"), 0o700); err != nil {
		t.Fatal(err)
	}
	if commitMemoAllowed(root) {
		t.Fatal("replacement refs must retain fresh Git semantics")
	}
}

func TestStatusOnlyKeepsTrackedEditsFreshWithoutInventoryCommands(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	runGit(t, root, "init", "-b", "main")
	file := filepath.Join(root, "tracked")
	if err := os.WriteFile(file, []byte("first"), 0600); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "tracked")
	observer := NewGit()
	runner := &countingGitRunner{calls: map[string]int{}}
	observer.runner = runner
	first, err := observer.ObserveStatus(context.Background(), "one", root)
	if err != nil || first.Changes.Staged != 1 {
		t.Fatalf("first=%#v err=%v", first, err)
	}
	if err := os.WriteFile(file, []byte("other"), 0600); err != nil {
		t.Fatal(err)
	}
	second, err := observer.ObserveStatus(context.Background(), "one", root)
	if err != nil || second.Changes.Modified != 1 || second.LastCommit != nil || len(second.Remotes)+len(second.Worktrees) != 0 {
		t.Fatalf("second=%#v err=%v", second, err)
	}
	if runner.calls["status"] != 2 || runner.calls["rev-parse"] != 1 || runner.calls["log"]+runner.calls["remote"]+runner.calls["worktree"] != 0 {
		t.Fatalf("unneeded commands: %#v", runner.calls)
	}
}
