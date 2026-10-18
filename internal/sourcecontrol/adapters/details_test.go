package adapters

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInventoryLiteralPathsPartialStagingRenameAndBranches(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	runGit(t, root, "init", "-b", "main")
	runGit(t, root, "config", "user.name", "Test")
	runGit(t, root, "config", "user.email", "test@example.invalid")
	name := "space and name.txt"
	untracked := "untracked file"
	if runtime.GOOS != "windows" {
		name = "space and\nnewline.txt"
		untracked = "untracked\tfile"
	}
	file := filepath.Join(root, name)
	if err := os.WriteFile(file, []byte("one"), 0600); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "--", name)
	runGit(t, root, "commit", "-m", "first")
	runGit(t, root, "branch", "feature")
	if err := os.WriteFile(file, []byte("two"), 0600); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "--", name)
	if err := os.WriteFile(file, []byte("three"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, untracked), []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	observer := NewGit()
	state, err := observer.ObserveInventory(context.Background(), "p", root)
	if err != nil {
		t.Fatal(err)
	}
	details := state.Details
	if details == nil || len(details.Files) != 2 || len(details.Branches) != 2 {
		t.Fatalf("inventory=%#v", state)
	}
	change := details.Files[0]
	if change.Path != name || change.Index != "M" || change.WorkingTree != "M" {
		t.Fatalf("literal partial stage=%#v", change)
	}
	if details.Branches[0].Name != "feature" || details.Branches[1].Name != "main" || !details.Branches[1].Current {
		t.Fatalf("branches=%#v", details.Branches)
	}
	// Basic observations retain their cheaper contract, including diagnostics.
	basic, err := observer.Observe(context.Background(), "p", root)
	if err != nil || basic.Details != nil {
		t.Fatalf("basic unexpectedly read inventory: %#v %v", basic, err)
	}
	runGit(t, root, "reset", "--hard", "HEAD")
	runGit(t, root, "mv", "--", name, "renamed file")
	state, err = observer.ObserveInventory(context.Background(), "p", root)
	if err != nil {
		t.Fatal(err)
	}
	change = state.Details.Files[0]
	if change.Path != "renamed file" || change.OriginalPath != name || change.Index != "R" {
		t.Fatalf("rename=%#v", change)
	}
}

func TestParseFilesBoundsAndConflict(t *testing.T) {
	t.Parallel()
	files, truncated := parseFiles([]byte("UU conflict\x00?? new\x00R  destination\x00old\nname\x00"))
	if truncated || len(files) != 3 || !files[0].Conflicted || !files[1].Untracked || files[2].OriginalPath != "old\nname" {
		t.Fatalf("files=%#v truncated=%v", files, truncated)
	}
	files, truncated = parseFiles([]byte(strings.Repeat("?? file\x00", fileLimit+1)))
	if !truncated || len(files) != fileLimit {
		t.Fatal("unbounded file inventory")
	}
	output := &gitOutput{remaining: 4}
	if _, err := output.Write([]byte("12345")); err == nil || !output.exceeded || output.Len() != 0 {
		t.Fatal("output cap failed")
	}
}

func TestWorktreeInventoryPreservesLiteralNames(t *testing.T) {
	t.Parallel()
	path := "/repo/space and\nnewline"
	trees := parseWorktrees([]byte("worktree " + path + "\x00HEAD abc\x00branch refs/heads/feature\x00\x00"))
	if len(trees) != 1 || trees[0].Path != path || trees[0].Branch != "feature" {
		t.Fatalf("worktree path corrupted: %#v", trees)
	}
}

func TestOperationStateBlocksInterruptedGitSequences(t *testing.T) {
	t.Parallel()
	for _, state := range []struct{ name, want string }{{"CHERRY_PICK_HEAD", "cherry_pick"}, {"REVERT_HEAD", "revert"}, {"BISECT_LOG", "bisect"}, {"sequencer", "sequencer"}, {"index.lock", "unknown"}} {
		t.Run(state.name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, state.name), []byte("pending"), 0600); err != nil {
				t.Fatal(err)
			}
			if got := operationState(root); got != state.want {
				t.Fatalf("state=%s", got)
			}
		})
	}
}
