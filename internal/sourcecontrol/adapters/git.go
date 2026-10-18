// Package adapters provides installed-Git observation using stable porcelain formats.
package adapters

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"envarium.dev/envarium/internal/sourcecontrol/domain"
)

type gitRunner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}

type installedGit struct{ executable string }

func (g installedGit) Run(ctx context.Context, root string, arguments ...string) ([]byte, error) {
	args := append([]string{"-c", "core.fsmonitor=false", "-C", root}, arguments...)
	executable := g.executable
	if executable == "" {
		executable = "git"
	}
	output, err := boundedGitOutput(exec.CommandContext(ctx, executable, args...))
	// A selected toolchain may be removed while the daemon remains running.
	if executable != "git" && errors.Is(err, os.ErrNotExist) {
		return boundedGitOutput(exec.CommandContext(ctx, "git", args...))
	}
	return output, err
}

func boundedGitOutput(command *exec.Cmd) ([]byte, error) {
	output := &gitOutput{remaining: 4 << 20}
	command.Stdout = output
	err := command.Run()
	if output.exceeded {
		return nil, errors.New("Git observation exceeded the four-MiB output limit")
	}
	return output.Bytes(), err
}

type gitOutput struct {
	bytes.Buffer
	remaining int
	exceeded  bool
}

func (o *gitOutput) Write(p []byte) (int, error) {
	if len(p) > o.remaining {
		o.exceeded = true
		return 0, io.ErrShortBuffer
	}
	o.remaining -= len(p)
	return o.Buffer.Write(p)
}

// Git observes one repository without invoking hooks or repository commands.
type Git struct {
	runner   gitRunner
	now      func() time.Time
	metadata *gitMetadata
}

// NewGit creates an observer backed by the installed Git executable.
func NewGit() *Git {
	return &Git{runner: newInstalledGit(), now: time.Now, metadata: &gitMetadata{entries: make(map[string]gitMetadataEntry)}}
}

// Observe returns a fresh, read-only repository snapshot for one trusted root.
func (g *Git) Observe(ctx context.Context, projectID, root string) (domain.State, error) {
	state, err := g.ObserveStatus(ctx, projectID, root)
	if err != nil || !state.Repository {
		return state, err
	}
	state.LastCommit = g.commitAtHead(ctx, root, state.Head)
	state.Remotes = g.remotes(ctx, root)
	state.Worktrees, err = g.ObserveWorktrees(ctx, root)
	return state, err
}

// ObserveInventory adds file and local branch inventory only on explicit demand.
func (g *Git) ObserveInventory(ctx context.Context, projectID, root string) (domain.State, error) {
	state, err := g.Observe(ctx, projectID, root)
	if err != nil || !state.Repository {
		return state, err
	}
	details, err := g.details(ctx, root)
	state.Details = &details
	return state, err
}

// ObserveStatus reads all mutable porcelain status without unused enrichment.
func (g *Git) ObserveStatus(ctx context.Context, projectID, root string) (domain.State, error) {
	state := domain.State{ProjectID: projectID, Remotes: []domain.Remote{}, Worktrees: []domain.Worktree{}, ObservedAt: g.now().UTC()}
	directory, err := g.gitDirectory(ctx, root)
	if err != nil {
		return state, gitProbeError(ctx, err)
	}
	if directory == "" {
		return state, nil
	}
	state.Repository = true
	status, err := g.runner.Run(ctx, root, "status", "--porcelain=v2", "--branch", "--show-stash", "--untracked-files=normal")
	if err != nil {
		return domain.State{}, fmt.Errorf("read Git porcelain status: %w", err)
	}
	parseStatus(status, &state)
	state.OperationState = operationState(directory)
	return state, nil
}

func gitProbeError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return nil
	}
	return fmt.Errorf("inspect Git repository: %w", err)
}

func parseStatus(output []byte, state *domain.State) {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "# branch.oid "):
			state.Head = strings.TrimPrefix(line, "# branch.oid ")
		case strings.HasPrefix(line, "# branch.head "):
			state.Branch = strings.TrimPrefix(line, "# branch.head ")
			state.Detached = state.Branch == "(detached)"
		case strings.HasPrefix(line, "# branch.upstream "):
			state.Upstream = strings.TrimPrefix(line, "# branch.upstream ")
		case strings.HasPrefix(line, "# branch.ab "):
			_, _ = fmt.Sscanf(strings.TrimPrefix(line, "# branch.ab "), "+%d -%d", &state.Ahead, &state.Behind)
		case strings.HasPrefix(line, "# stash "):
			state.Stashes, _ = strconv.Atoi(strings.TrimPrefix(line, "# stash "))
		case strings.HasPrefix(line, "? "):
			state.Changes.Untracked++
		case strings.HasPrefix(line, "u "):
			state.Changes.Conflicted++
		case strings.HasPrefix(line, "1 "), strings.HasPrefix(line, "2 "):
			fields := strings.Fields(line)
			if len(fields) > 1 && len(fields[1]) >= 2 {
				if fields[1][0] != '.' {
					state.Changes.Staged++
				}
				if fields[1][1] != '.' {
					state.Changes.Modified++
				}
			}
		}
	}
}

func (g *Git) lastCommit(ctx context.Context, root string) *domain.Commit {
	output, err := g.runner.Run(ctx, root, "log", "-1", "--format=%H%x00%h%x00%an%x00%aI%x00%s")
	if err != nil {
		return nil
	}
	fields := strings.Split(strings.TrimSpace(string(output)), "\x00")
	if len(fields) != 5 {
		return nil
	}
	committed, err := time.Parse(time.RFC3339, fields[3])
	if err != nil {
		return nil
	}
	return &domain.Commit{Hash: fields[0], ShortHash: fields[1], Author: fields[2], CommittedAt: committed, Subject: fields[4]}
}

func (g *Git) remotes(ctx context.Context, root string) []domain.Remote {
	output, err := g.runner.Run(ctx, root, "remote", "-v")
	if err != nil {
		return []domain.Remote{}
	}
	var result []domain.Remote
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 3 {
			result = append(result, domain.Remote{Name: fields[0], URL: remoteURL(fields[1]), Kind: strings.Trim(fields[2], "()")})
		}
	}
	return result
}

// ObserveWorktrees reads stable Git porcelain without invoking repository
// hooks or repository-defined commands.
func (g *Git) ObserveWorktrees(ctx context.Context, root string) ([]domain.Worktree, error) {
	output, err := g.runner.Run(ctx, root, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("read Git worktree inventory: %w", err)
	}
	return parseWorktrees(output), nil
}

func parseWorktrees(output []byte) []domain.Worktree {
	var result []domain.Worktree
	var current *domain.Worktree
	separator := "\n"
	if bytes.IndexByte(output, 0) >= 0 {
		separator = "\x00"
	}
	for _, line := range strings.Split(string(output), separator) {
		switch {
		case strings.HasPrefix(line, "worktree "):
			result = append(result, domain.Worktree{Path: strings.TrimPrefix(line, "worktree ")})
			current = &result[len(result)-1]
		case current != nil && strings.HasPrefix(line, "HEAD "):
			current.Head = strings.TrimPrefix(line, "HEAD ")
		case current != nil && strings.HasPrefix(line, "branch "):
			current.Branch = strings.TrimPrefix(strings.TrimPrefix(line, "branch "), "refs/heads/")
		case current != nil && line == "detached":
			current.Detached = true
		case current != nil && line == "bare":
			current.Bare = true
		case current != nil && strings.HasPrefix(line, "locked"):
			current.Locked = true
		}
	}
	return result
}

func operationState(directory string) string {
	for _, candidate := range []struct{ path, state string }{{"MERGE_HEAD", "merge"}, {"rebase-merge", "rebase"}, {"rebase-apply", "rebase"}, {"CHERRY_PICK_HEAD", "cherry_pick"}, {"REVERT_HEAD", "revert"}, {"BISECT_LOG", "bisect"}, {"sequencer", "sequencer"}, {"index.lock", "unknown"}} {
		path := filepath.Join(directory, candidate.path)
		if _, err := os.Stat(path); err == nil {
			return candidate.state
		} else if !errors.Is(err, os.ErrNotExist) {
			return "unknown"
		}
	}
	return ""
}
