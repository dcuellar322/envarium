// Package sockets observes operating-system socket bindings without product policy.
package sockets

import (
	"context"
	"sync"
	"time"

	gnet "github.com/shirou/gopsutil/v4/net"
	"golang.org/x/sync/singleflight"
)

const observationWindow = 2 * time.Second

// TCPObserver coalesces read-only TCP inventories across projects and ports.
// Lifecycle decisions must call Connections directly instead. Identities are
// never cached here: consumers still revalidate the current process fingerprint.
type TCPObserver struct {
	read     func(context.Context, string) ([]gnet.ConnectionStat, error)
	now      func() time.Time
	mu       sync.Mutex
	values   []gnet.ConnectionStat
	err      error
	until    time.Time
	requests singleflight.Group
}

// NewTCPObserver creates one bounded inventory owned by its consuming driver.
func NewTCPObserver() *TCPObserver { return &TCPObserver{read: Connections, now: time.Now} }

// Connections returns a snapshot no more than two seconds old, including errors.
// Failed refreshes replace previous evidence; cancellation is never cached.
func (o *TCPObserver) Connections(ctx context.Context) ([]gnet.ConnectionStat, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if values, ok, err := o.cached(); ok {
		return values, err
	}
	result := o.requests.DoChan("tcp", func() (any, error) {
		if values, ok, err := o.cached(); ok {
			return values, err
		}
		values, err := o.read(ctx, "tcp")
		if ctx.Err() == nil {
			o.mu.Lock()
			o.values, o.err, o.until = append([]gnet.ConnectionStat(nil), values...), err, o.now().Add(observationWindow)
			o.mu.Unlock()
		}
		return values, err
	})
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-result:
		if result.Err != nil {
			return nil, result.Err
		}
		return append([]gnet.ConnectionStat(nil), result.Val.([]gnet.ConnectionStat)...), nil
	}
}

func (o *TCPObserver) cached() ([]gnet.ConnectionStat, bool, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if !o.now().Before(o.until) {
		return nil, false, nil
	}
	return append([]gnet.ConnectionStat(nil), o.values...), true, o.err
}
