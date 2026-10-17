package sockets

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	gnet "github.com/shirou/gopsutil/v4/net"
)

func TestTCPObservationExpiresAndDoesNotRetainFailedEvidence(t *testing.T) {
	now := time.Now()
	calls := 0
	failure := errors.New("permission denied")
	observer := NewTCPObserver()
	observer.now = func() time.Time { return now }
	observer.read = func(context.Context, string) ([]gnet.ConnectionStat, error) {
		calls++
		if calls == 2 {
			return nil, failure
		}
		return []gnet.ConnectionStat{{Pid: int32(calls), Laddr: gnet.Addr{Port: 3000}}}, nil
	}
	values, err := observer.Connections(context.Background())
	if err != nil || values[0].Pid != 1 {
		t.Fatalf("first inventory: %v %v", values, err)
	}
	values[0].Pid = 99
	values, err = observer.Connections(context.Background())
	if err != nil || values[0].Pid != 1 || calls != 1 {
		t.Fatalf("mutable or repeated inventory: %v %v calls=%d", values, err, calls)
	}
	now = now.Add(observationWindow)
	for range 10 {
		values, err = observer.Connections(context.Background())
		if !errors.Is(err, failure) || len(values) != 0 {
			t.Fatalf("failed evidence: %v %v", values, err)
		}
	}
	if calls != 2 {
		t.Fatalf("failed scan was repeated %d times", calls)
	}
	now = now.Add(observationWindow)
	values, err = observer.Connections(context.Background())
	if err != nil || values[0].Pid != 3 {
		t.Fatalf("recovery: %v %v", values, err)
	}
}

func TestConcurrentTCPObserversShareOneInventory(t *testing.T) {
	var calls atomic.Int32
	entered, release := make(chan struct{}), make(chan struct{})
	observer := NewTCPObserver()
	observer.read = func(context.Context, string) ([]gnet.ConnectionStat, error) {
		if calls.Add(1) == 1 {
			close(entered)
		}
		<-release
		return []gnet.ConnectionStat{{Pid: 42}}, nil
	}
	var workers sync.WaitGroup
	for range 32 {
		workers.Go(func() {
			values, err := observer.Connections(context.Background())
			if err != nil || len(values) != 1 || values[0].Pid != 42 {
				t.Errorf("inventory: %v %v", values, err)
			}
		})
	}
	<-entered
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := observer.Connections(canceled); !errors.Is(err, context.Canceled) {
		t.Fatalf("wait cancellation: %v", err)
	}
	close(release)
	workers.Wait()
	if calls.Load() != 1 {
		t.Fatalf("got %d scans", calls.Load())
	}
}

func TestCanceledTCPScanIsNotCached(t *testing.T) {
	observer := NewTCPObserver()
	ctx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int32
	finished := make(chan struct{})
	observer.read = func(ctx context.Context, _ string) ([]gnet.ConnectionStat, error) {
		if calls.Add(1) == 1 {
			cancel()
			defer close(finished)
			return nil, ctx.Err()
		}
		return []gnet.ConnectionStat{{Pid: 2}}, nil
	}
	if _, err := observer.Connections(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
	<-finished
	// A waiter may still share the cancelled flight until its result is published.
	deadline := time.Now().Add(time.Second)
	for {
		values, err := observer.Connections(context.Background())
		if err == nil && len(values) == 1 && values[0].Pid == 2 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("retry: %v %v", values, err)
		}
	}
}
