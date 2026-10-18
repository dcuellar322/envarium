package application

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
)

type feedSource struct {
	mu    sync.Mutex
	calls int
	value string
	err   error
}

func (s *feedSource) ResolveRuntime(context.Context, string) (domain.ProjectRuntime, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return domain.ProjectRuntime{Root: s.value}, s.err
}
func (s *feedSource) ListRuntimeProjectIDs(context.Context) ([]string, error) {
	return []string{"p"}, nil
}
func (s *feedSource) ResolveRuntimeProjects(context.Context) (map[string]domain.ProjectRuntime, map[string]error, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	return map[string]domain.ProjectRuntime{"p": {Root: s.value, Ports: map[string]domain.PortDeclaration{"web": {Host: 1234}}}}, map[string]error{}, s.err
}

func receiveFeed(t *testing.T, ch <-chan domain.ProjectSnapshot) domain.ProjectSnapshot {
	t.Helper()
	select {
	case value, ok := <-ch:
		if !ok {
			t.Fatal("feed closed unexpectedly")
		}
		return value
	case <-time.After(3 * time.Second):
		t.Fatal("feed did not publish")
		return domain.ProjectSnapshot{}
	}
}

func TestSchedulerFeedFreshSharedCopiedBoundedAndCanceled(t *testing.T) {
	source := &feedSource{value: "first"}
	feed := NewSchedulerSource(source)
	first, unsub1 := feed.SubscribeRuntimeProjects()
	defer unsub1()
	slow, unsub2 := feed.SubscribeRuntimeProjects()
	defer unsub2()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { feed.Run(ctx); close(done) }()
	one := receiveFeed(t, first)
	two := receiveFeed(t, slow)
	source.mu.Lock()
	calls := source.calls
	source.value = "changed"
	source.mu.Unlock()
	if calls != 1 || one.Projects["p"].Root != "first" || two.Projects["p"].Root != "first" {
		t.Fatalf("not one shared cycle: calls=%d", calls)
	}
	one.Projects["p"].Ports["web"] = domain.PortDeclaration{Host: 9999}
	if two.Projects["p"].Ports["web"].Host != 1234 {
		t.Fatal("subscriber maps alias")
	}
	// A direct read never reuses the feed, even between periodic observations.
	direct, err := feed.ResolveRuntime(ctx, "p")
	if err != nil || direct.Root != "changed" {
		t.Fatalf("stale direct read: %#v %v", direct, err)
	}
	updated := receiveFeed(t, first)
	if updated.Projects["p"].Root != "changed" {
		t.Fatal("periodic configuration stale")
	}
	source.mu.Lock()
	source.value = "latest"
	source.err = errors.New("source unavailable")
	source.mu.Unlock()
	failed := receiveFeed(t, first)
	if failed.Err == nil {
		t.Fatal("failed cycle hidden")
	}
	// The unconsumed subscriber retains only the latest cycle and cannot block us.
	latest := receiveFeed(t, slow)
	if latest.Err == nil || latest.Projects["p"].Root != "latest" {
		t.Fatalf("slow subscriber queued stale cycles: %#v", latest)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("feed failed to cancel")
	}
	unsub1()
	unsub1()
	unsub2()
	closed, unsubscribe := feed.SubscribeRuntimeProjects()
	defer unsubscribe()
	if _, ok := <-closed; ok {
		t.Fatal("new subscriber after shutdown remained open")
	}
}
