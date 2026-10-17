package application

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"envarium.dev/envarium/internal/sourcecontrol/domain"
)

type heldObserver struct {
	calls   atomic.Int64
	started chan struct{}
	release chan struct{}
}

func (o *heldObserver) Observe(ctx context.Context, id, _ string) (domain.State, error) {
	if o.calls.Add(1) == 1 {
		close(o.started)
	}
	select {
	case <-ctx.Done():
		return domain.State{}, ctx.Err()
	case <-o.release:
	}
	return domain.State{ProjectID: id, Remotes: []domain.Remote{{Name: "origin"}}, LastCommit: &domain.Commit{Subject: "original"}}, nil
}
func TestConcurrentGitReadsShareWorkButNotResultsOrWaiterCancellation(t *testing.T) {
	t.Parallel()
	observer := &heldObserver{started: make(chan struct{}), release: make(chan struct{})}
	service := NewService(projectSourceStub{root: "/repo"}, observer)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	first := make(chan error, 1)
	go func() { _, err := service.Get(ctx, "one"); first <- err }()
	<-observer.started
	second := make(chan domain.State, 1)
	secondErr := make(chan error, 1)
	joined := make(chan struct{})
	go func() {
		state, err := service.Get(&observedWaiterContext{Context: context.Background(), joined: joined}, "one")
		second <- state
		secondErr <- err
	}()
	<-joined
	// The observer remains blocked; cancellation must not cancel the shared read.
	cancel()
	if err := <-first; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
	close(observer.release)
	state := <-second
	if err := <-secondErr; err != nil {
		t.Fatal(err)
	}
	if observer.calls.Load() != 1 {
		t.Fatalf("duplicate reads: %d", observer.calls.Load())
	}
	state.LastCommit.Subject = "caller mutation"
	state.Remotes[0].Name = "caller mutation"
	next, err := service.Get(context.Background(), "one")
	if err != nil {
		t.Fatal(err)
	}
	if observer.calls.Load() != 2 || next.LastCommit.Subject != "original" || next.Remotes[0].Name != "origin" {
		t.Fatalf("fresh next read: %#v", next)
	}
}

func TestCanceledAndUntrustedReadsDoNotInvokeGit(t *testing.T) {
	t.Parallel()
	o := &heldObserver{started: make(chan struct{}), release: make(chan struct{})}
	service := NewService(projectSourceStub{err: ErrProjectUntrusted}, o)
	if _, err := service.Get(context.Background(), "one"); !errors.Is(err, ErrProjectUntrusted) {
		t.Fatal(err)
	}
	service = NewService(projectSourceStub{root: "/repo"}, o)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.Get(ctx, "one"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if o.calls.Load() != 0 {
		t.Fatal("Git ran without a valid caller")
	}
}

func TestGitQueueCancellationIsBounded(t *testing.T) {
	t.Parallel()
	o := &heldObserver{started: make(chan struct{}), release: make(chan struct{})}
	defer close(o.release)
	service := NewService(projectSourceStub{root: "/repo"}, o)
	for range cap(service.gitReads) {
		service.gitReads <- struct{}{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := service.Get(ctx, "queued"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	// Release capacity so the bounded shared invocation can complete too.
	for range cap(service.gitReads) {
		<-service.gitReads
	}
}

type observedWaiterContext struct {
	context.Context
	once   sync.Once
	joined chan struct{}
}

func (c *observedWaiterContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.joined) })
	return c.Context.Done()
}
