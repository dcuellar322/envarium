package application

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"envarium.dev/envarium/internal/sourcecontrol/domain"
)

var (
	// ErrUnsafeCheckout requires local resolution of edits or repository extensions.
	ErrUnsafeCheckout = errors.New("branch switch requires a clean checkout with no in-progress operation or checkout filters")
	// ErrCheckoutChanged binds consent to the exact observed source and target.
	ErrCheckoutChanged = errors.New("repository or target branch changed; review a new branch switch")
)

// BranchSwitcher is the narrow installed-Git mutation boundary.
type BranchSwitcher interface {
	CheckSwitchSafety(context.Context, string) error
	SwitchBranch(context.Context, string, string) error
}

// StoppedCheckoutGate excludes lifecycle changes while a checkout is changed.
type StoppedCheckoutGate interface {
	WithStoppedProject(context.Context, string, func() error) error
	WithCheckoutMutation(context.Context, string, func() error) error
}

// BranchService coordinates explicit local checkout changes, without executing
// new branch runtime commands or inventing an automatic rollback.
type BranchService struct {
	git      *Service
	switcher BranchSwitcher
	runtime  StoppedCheckoutGate
}

// NewBranchService composes trusted Git with stopped-checkout serialization.
func NewBranchService(git *Service, switcher BranchSwitcher, runtime StoppedCheckoutGate) *BranchService {
	return &BranchService{git: git, switcher: switcher, runtime: runtime}
}

// Prepare observes exact refs after proving every runtime using this root stopped.
func (s *BranchService) Prepare(ctx context.Context, projectID, branch string) (domain.BranchSwitchPlan, error) {
	var plan domain.BranchSwitchPlan
	err := s.runtime.WithStoppedProject(ctx, projectID, func() error {
		var err error
		plan, err = s.prepare(ctx, projectID, branch)
		return err
	})
	return plan, err
}
func (s *BranchService) prepare(ctx context.Context, projectID, branch string) (domain.BranchSwitchPlan, error) {
	readCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	root, err := s.git.projects.Root(readCtx, projectID)
	if err != nil {
		return domain.BranchSwitchPlan{}, err
	}
	// Status can consult clean filters as well as checkout's smudge/process
	// filters. Check configuration before asking Git to inspect file contents.
	if err := s.switcher.CheckSwitchSafety(readCtx, root); err != nil {
		return domain.BranchSwitchPlan{}, fmt.Errorf("%w: %v", ErrUnsafeCheckout, err)
	}
	observer, ok := s.git.observer.(interface {
		ObserveInventory(context.Context, string, string) (domain.State, error)
	})
	if !ok {
		return domain.BranchSwitchPlan{}, ErrInventoryUnsupported
	}
	state, err := observer.ObserveInventory(readCtx, projectID, root)
	if err != nil {
		return domain.BranchSwitchPlan{}, err
	}
	if !cleanCheckout(state) {
		return domain.BranchSwitchPlan{}, ErrUnsafeCheckout
	}
	var target *domain.Branch
	for _, candidate := range state.Details.Branches {
		if candidate.Name == branch && !candidate.Current {
			selected := candidate
			target = &selected
			break
		}
	}
	if target == nil {
		return domain.BranchSwitchPlan{}, fmt.Errorf("%w: choose another existing local branch", ErrUnsafeCheckout)
	}
	for _, worktree := range state.Worktrees {
		if worktree.Branch == branch && filepath.Clean(worktree.Path) != filepath.Clean(root) {
			return domain.BranchSwitchPlan{}, fmt.Errorf("%w: branch is already checked out in another worktree", ErrUnsafeCheckout)
		}
	}
	if err := s.switcher.CheckSwitchSafety(readCtx, root); err != nil {
		return domain.BranchSwitchPlan{}, fmt.Errorf("%w: %v", ErrUnsafeCheckout, err)
	}
	return domain.BranchSwitchPlan{ProjectID: projectID, Root: root, FromBranch: state.Branch, FromHead: state.Head, ToBranch: target.Name, ToHead: target.Head, Risk: "mutating", ReviewRestart: true}, nil
}

// Execute revalidates under the same root gate used by runtime execution.
func (s *BranchService) Execute(ctx context.Context, expected domain.BranchSwitchPlan) error {
	return s.runtime.WithCheckoutMutation(ctx, expected.ProjectID, func() error {
		current, err := s.prepare(ctx, expected.ProjectID, expected.ToBranch)
		if err != nil {
			return err
		}
		if current != expected {
			return ErrCheckoutChanged
		}
		switchCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		if err := s.switcher.SwitchBranch(switchCtx, current.Root, current.ToBranch); err != nil {
			return fmt.Errorf("branch switch failed; inspect the repository before retrying: %w", err)
		}
		state, err := s.git.observer.Observe(switchCtx, expected.ProjectID, current.Root)
		if err != nil {
			return err
		}
		if state.Branch != current.ToBranch || state.Head != current.ToHead {
			return ErrCheckoutChanged
		}
		return nil
	})
}

func cleanCheckout(state domain.State) bool {
	return state.Repository && state.Head != "" && state.Head != "(initial)" && !state.Detached && state.OperationState == "" && state.Changes == (domain.ChangeCounts{}) && state.Details != nil && len(state.Details.Files) == 0 && !state.Details.FilesTruncated
}
