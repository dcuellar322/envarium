package compose

import (
	"errors"
	"strings"
	"sync"
)

const maximumIdleEngineClients = 8

// engineClients retains transports, never observations or negotiation results.
// Only explicit Unix endpoints qualify; environment/TLS clients remain fresh.
type engineClients struct {
	mu     sync.Mutex
	idle   map[dockerConnection][]engineClient
	count  int
	closed bool
}

func newEngineClients() *engineClients {
	return &engineClients{idle: make(map[dockerConnection][]engineClient)}
}

func (p *engineClients) borrow(connection dockerConnection, create func() (engineClient, error)) (engineClient, error) {
	if p == nil || connection.FromEnv || !strings.HasPrefix(connection.Host, "unix://") {
		return create()
	}
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, errors.New("docker client pool is closed")
	}
	var engine engineClient
	if clients := p.idle[connection]; len(clients) > 0 {
		engine = clients[len(clients)-1]
		p.count--
		if len(clients) == 1 {
			delete(p.idle, connection)
		} else {
			p.idle[connection] = clients[:len(clients)-1]
		}
	}
	p.mu.Unlock()
	if engine == nil {
		var err error
		engine, err = create()
		if err != nil {
			return nil, err
		}
	}
	return &engineLease{engineClient: engine, pool: p, connection: connection}, nil
}

func (p *engineClients) release(connection dockerConnection, engine engineClient) error {
	p.mu.Lock()
	if p.closed || p.count >= maximumIdleEngineClients {
		p.mu.Unlock()
		return engine.Close()
	}
	p.idle[connection] = append(p.idle[connection], engine)
	p.count++
	p.mu.Unlock()
	return nil
}

func (p *engineClients) close() error {
	p.mu.Lock()
	p.closed = true
	idle := p.idle
	p.idle = make(map[dockerConnection][]engineClient)
	p.count = 0
	p.mu.Unlock()
	var failures []error
	for _, clients := range idle {
		for _, engine := range clients {
			failures = append(failures, engine.Close())
		}
	}
	return errors.Join(failures...)
}

type engineLease struct {
	engineClient
	pool       *engineClients
	connection dockerConnection
	once       sync.Once
	err        error
}

func (c *engineLease) Close() error {
	c.once.Do(func() { c.err = c.pool.release(c.connection, c.engineClient) })
	return c.err
}

func (c *engineLease) discard() error {
	c.once.Do(func() { c.err = c.engineClient.Close() })
	return c.err
}
