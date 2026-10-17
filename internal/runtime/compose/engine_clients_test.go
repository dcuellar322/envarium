package compose

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

type countedEngine struct {
	fakeEngine
	closed atomic.Int32
}

func (c *countedEngine) Close() error { c.closed.Add(1); return nil }

func TestEngineClientsReuseOnlyIdleClientsWithMatchingLocalIdentity(t *testing.T) {
	t.Parallel()
	pool := newEngineClients()
	t.Cleanup(func() { _ = pool.close() })
	connection := dockerConnection{Host: "unix:///engine.sock", ContextName: "one"}
	created := []*countedEngine{}
	create := func() (engineClient, error) {
		value := &countedEngine{}
		created = append(created, value)
		return value, nil
	}
	first, err := pool.borrow(connection, create)
	if err != nil {
		t.Fatal(err)
	}
	second, err := pool.borrow(connection, create)
	if err != nil || len(created) != 2 {
		t.Fatalf("active lease reused: created=%d err=%v", len(created), err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	_ = first.Close()
	if pool.count != 1 || created[0].closed.Load() != 0 {
		t.Fatal("lease returned twice or closed before reuse")
	}
	again, err := pool.borrow(connection, create)
	if err != nil || again.(*engineLease).engineClient != created[0] {
		t.Fatal("matching idle client not reused")
	}
	other := connection
	other.ContextName = "two"
	third, err := pool.borrow(other, create)
	if err != nil || len(created) != 3 {
		t.Fatal("different endpoint identity reused")
	}
	_ = again.Close()
	_ = second.Close()
	_ = third.Close()
	if err := pool.close(); err != nil {
		t.Fatal(err)
	}
	for _, client := range created {
		if client.closed.Load() != 1 {
			t.Fatal("retained transport not closed exactly once")
		}
	}
}

func TestEngineClientsBoundIdleRetentionAndCloseOutstandingLeases(t *testing.T) {
	t.Parallel()
	pool := newEngineClients()
	var created atomic.Int32
	var clientsMu sync.Mutex
	clients := []*countedEngine{}
	create := func() (engineClient, error) {
		value := &countedEngine{}
		created.Add(1)
		clientsMu.Lock()
		clients = append(clients, value)
		clientsMu.Unlock()
		return value, nil
	}
	var workers sync.WaitGroup
	for i := range 32 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			lease, err := pool.borrow(dockerConnection{Host: fmt.Sprintf("unix:///engine-%d.sock", i)}, create)
			if err == nil {
				_ = lease.Close()
			}
		}()
	}
	workers.Wait()
	if pool.count > maximumIdleEngineClients || len(pool.idle) > maximumIdleEngineClients {
		t.Fatal("idle transport retention is unbounded")
	}
	active, err := pool.borrow(dockerConnection{Host: "unix:///active.sock"}, create)
	if err != nil {
		t.Fatal(err)
	}
	if err := pool.close(); err != nil {
		t.Fatal(err)
	}
	_ = active.Close()
	_ = active.Close()
	if _, err := pool.borrow(dockerConnection{Host: "unix:///closed.sock"}, create); err == nil {
		t.Fatal("borrowed after shutdown")
	}
	if int(created.Load()) != len(clients) {
		t.Fatal("unexpected creation count")
	}
	for _, client := range clients {
		if client.closed.Load() != 1 {
			t.Fatal("transport leaked or closed twice")
		}
	}
}

func TestEngineClientsDoNotPoolEnvironmentOrTLSClientsAndDiscardFailures(t *testing.T) {
	t.Parallel()
	pool := newEngineClients()
	t.Cleanup(func() { _ = pool.close() })
	for _, connection := range []dockerConnection{{Host: "unix:///env.sock", FromEnv: true}, {Host: "tcp://tls.example:2376"}} {
		value := &countedEngine{}
		lease, err := pool.borrow(connection, func() (engineClient, error) { return value, nil })
		if err != nil {
			t.Fatal(err)
		}
		_ = lease.Close()
		if pool.count != 0 || value.closed.Load() != 1 {
			t.Fatal("environment/TLS client retained")
		}
	}
	connection := dockerConnection{Host: "unix:///one.sock"}
	failure := errors.New("creation failed")
	if _, err := pool.borrow(connection, func() (engineClient, error) { return nil, failure }); !errors.Is(err, failure) {
		t.Fatal("creation failure hidden")
	}
	value := &countedEngine{}
	lease, err := pool.borrow(connection, func() (engineClient, error) { return value, nil })
	if err != nil {
		t.Fatal(err)
	}
	_ = lease.(*engineLease).discard()
	_ = lease.Close()
	if pool.count != 0 || value.closed.Load() != 1 {
		t.Fatal("failed negotiation client retained")
	}
}
