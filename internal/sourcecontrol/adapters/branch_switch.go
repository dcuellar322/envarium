package adapters

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// CheckSwitchSafety rejects checkout filters, including inherited Git LFS
// programs. A user can switch those repositories locally and then refresh.
func (g *Git) CheckSwitchSafety(ctx context.Context, root string) error {
	output, err := g.runner.Run(ctx, root, "config", "--get-regexp", `^filter\..*\.(clean|smudge|process)$`)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return nil
	}
	if err != nil {
		return fmt.Errorf("cannot inspect checkout filter configuration")
	}
	if len(output) > 0 {
		return errors.New("checkout filters are configured; switch this repository with Git locally")
	}
	return nil
}

// SwitchBranch never forces, overwrites ignored files, invokes hooks or
// recursively updates submodules. Git itself protects other worktree checkouts.
func (g *Git) SwitchBranch(ctx context.Context, root, branch string) error {
	if err := g.CheckSwitchSafety(ctx, root); err != nil {
		return err
	}
	if _, err := g.runner.Run(ctx, root, "check-ref-format", "--branch", branch); err != nil {
		return errors.New("invalid local branch")
	}
	hooks, err := os.MkdirTemp("", "envarium-empty-hooks-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(hooks) }()
	_, err = g.runner.Run(ctx, root, "-c", "core.hooksPath="+hooks, "-c", "submodule.recurse=false", "checkout", "--no-guess", "--no-overwrite-ignore", "--no-recurse-submodules", branch, "--")
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return errors.New("Git refused checkout; edits, ignored files or another worktree may need attention")
	}
	return nil
}
