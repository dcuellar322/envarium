package application

import (
	"context"
	"errors"
	"testing"

	"envarium.dev/envarium/internal/sourcecontrol/domain"
)

type branchFixture struct {
	state    domain.State
	safety   error
	calls    int
	switched bool
}

func (f *branchFixture) Observe(ctx context.Context, _ string, _ string) (domain.State, error) {
	if err := ctx.Err(); err != nil {
		return domain.State{}, err
	}
	return f.state, nil
}
func (f *branchFixture) ObserveInventory(ctx context.Context, id, root string) (domain.State, error) {
	return f.Observe(ctx, id, root)
}
func (f *branchFixture) CheckSwitchSafety(context.Context, string) error { return f.safety }
func (f *branchFixture) SwitchBranch(_ context.Context, _ string, branch string) error {
	f.calls++
	f.switched = true
	f.state.Branch = branch
	f.state.Head = "target"
	return nil
}

type stoppedFixture struct {
	err   error
	calls int
}

func (f *stoppedFixture) WithStoppedProject(ctx context.Context, _ string, run func() error) error {
	f.calls++
	if err := ctx.Err(); err != nil {
		return err
	}
	if f.err != nil {
		return f.err
	}
	return run()
}
func cleanBranchFixture() *branchFixture {
	return &branchFixture{state: domain.State{Repository: true, Branch: "main", Head: "source", Details: &domain.Details{Files: []domain.FileChange{}, Branches: []domain.Branch{{Name: "main", Head: "source", Current: true}, {Name: "feature", Head: "target"}}}}}
}

func TestBranchSwitchBindsReviewedRefsAndChecksTrustAgain(t *testing.T) {
	t.Parallel()
	f := cleanBranchFixture()
	gate := &stoppedFixture{}
	git := NewService(projectSourceStub{root: "/repo"}, f)
	service := NewBranchService(git, f, gate)
	plan, err := service.Prepare(context.Background(), "p", "feature")
	if err != nil {
		t.Fatal(err)
	}
	f.state.Details.Branches[1].Head = "moved"
	if err := service.Execute(context.Background(), plan); !errors.Is(err, ErrCheckoutChanged) || f.calls != 0 {
		t.Fatalf("stale target accepted: %v", err)
	}
	f.state.Details.Branches[1].Head = "target"
	git.projects = projectSourceStub{err: ErrProjectUntrusted}
	if err := service.Execute(context.Background(), plan); !errors.Is(err, ErrProjectUntrusted) || f.calls != 0 {
		t.Fatalf("trust revocation ignored: %v", err)
	}
	git.projects = projectSourceStub{root: "/repo"}
	if err := service.Execute(context.Background(), plan); err != nil || f.calls != 1 || gate.calls != 4 {
		t.Fatalf("switch failed: %v calls=%d gates=%d", err, f.calls, gate.calls)
	}
}
func TestBranchSwitchRejectsUnsafeAndCanceledCheckouts(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"dirty", "untracked", "conflict", "merge", "detached", "other-worktree", "filters", "unknown-target", "active", "cancel"} {
		t.Run(kind, func(t *testing.T) {
			f := cleanBranchFixture()
			gate := &stoppedFixture{}
			ctx := context.Background()
			branch := "feature"
			switch kind {
			case "dirty":
				f.state.Changes.Modified = 1
			case "untracked":
				f.state.Details.Files = []domain.FileChange{{Path: "new", Untracked: true}}
			case "conflict":
				f.state.Changes.Conflicted = 1
			case "merge":
				f.state.OperationState = "merge"
			case "detached":
				f.state.Detached = true
			case "other-worktree":
				f.state.Worktrees = []domain.Worktree{{Path: "/other", Branch: "feature"}}
			case "filters":
				f.safety = errors.New("filter configured")
			case "unknown-target":
				branch = "missing"
			case "active":
				gate.err = errors.New("runtime active")
			case "cancel":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			service := NewBranchService(NewService(projectSourceStub{root: "/repo"}, f), f, gate)
			if _, err := service.Prepare(ctx, "p", branch); err == nil || f.calls != 0 {
				t.Fatal("unsafe checkout accepted")
			}
		})
	}
}

func (f *stoppedFixture) WithCheckoutMutation(ctx context.Context, id string, run func() error) error {
	return f.WithStoppedProject(ctx, id, run)
}
