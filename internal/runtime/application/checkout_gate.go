package application

import (
	"context"
	"envarium.dev/envarium/internal/foundation/identifier"
	"errors"
	"fmt"
	"sync"

	"envarium.dev/envarium/internal/runtime/domain"
)

// ErrCheckoutActive rejects changes that could alter files used by live services.
var ErrCheckoutActive = errors.New("stop this checkout's project and primary environment before switching branches")

// ErrPlanStale rejects a preview whose underlying checkout/configuration changed.
var ErrPlanStale = errors.New("runtime configuration changed; prepare a new lifecycle preview")

type checkoutGate struct {
	mu          sync.Mutex
	entries     map[string]*checkoutEntry
	newRevision func() (string, error)
}
type checkoutEntry struct {
	token    chan struct{}
	users    int
	revision string
}

func (g *checkoutGate) acquire(ctx context.Context, root string) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	g.mu.Lock()
	e, err := g.entryLocked(root)
	if err != nil {
		g.mu.Unlock()
		return nil, err
	}
	e.users++
	g.mu.Unlock()
	remove := func() {
		g.mu.Lock()
		defer g.mu.Unlock()
		e.users--
	}
	select {
	case e.token <- struct{}{}:
		return func() { <-e.token; remove() }, nil
	case <-ctx.Done():
		remove()
		return nil, ctx.Err()
	}
}

// WithStoppedProject serializes checkout mutation with all lifecycle execution
// using that same root, and observes drivers directly rather than a UI cache.
func (s *Service) WithStoppedProject(ctx context.Context, id string, action func() error) error {
	project, _, err := s.resolve(ctx, id)
	if err != nil {
		return err
	}
	release, err := s.checkouts.acquire(ctx, project.Root)
	if err != nil {
		return err
	}
	defer release()
	current, _, err := s.resolve(ctx, id)
	if err != nil {
		return err
	}
	if current.Root != project.Root {
		return ErrPlanStale
	}
	projects, failures, err := snapshotProjects(ctx, s.projects)
	if err != nil {
		return err
	}
	if failure := failures[id]; failure != nil {
		return failure
	}
	projects[id] = current
	for _, candidate := range projects {
		if candidate.Root != current.Root {
			continue
		}
		driver := s.drivers[candidate.Kind]
		if driver == nil {
			return domain.ErrUnsupportedDriver
		}
		observation, err := driver.Inspect(ctx, candidate)
		if err != nil {
			return fmt.Errorf("verify stopped checkout: %w", err)
		}
		if observation.State != domain.StateStopped && observation.State != domain.StateFailed {
			return ErrCheckoutActive
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return action()
}

// ErrCheckoutBusy bounds retained coordination state without evicting live roots.
var ErrCheckoutBusy = errors.New("too many checkout operations are in flight; wait and retry")

func (g *checkoutGate) freshRevision() (string, error) {
	if g.newRevision != nil {
		return g.newRevision()
	}
	return identifier.New("checkout")
}
func (g *checkoutGate) entryLocked(root string) (*checkoutEntry, error) {
	if g.entries == nil {
		g.entries = map[string]*checkoutEntry{}
	}
	if entry := g.entries[root]; entry != nil {
		if entry.revision == "" {
			revision, err := g.freshRevision()
			if err != nil {
				return nil, err
			}
			entry.revision = revision
		}
		return entry, nil
	}
	if len(g.entries) >= 128 {
		for key, entry := range g.entries {
			if entry.users == 0 {
				delete(g.entries, key)
				break
			}
		}
	}
	if len(g.entries) >= 128 {
		return nil, ErrCheckoutBusy
	}
	revision, err := g.freshRevision()
	if err != nil {
		return nil, err
	}
	entry := &checkoutEntry{token: make(chan struct{}, 1), revision: revision}
	g.entries[root] = entry
	return entry, nil
}
func (g *checkoutGate) revision(root string) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	entry, err := g.entryLocked(root)
	if err != nil {
		return "", err
	}
	return entry.revision, nil
}
func (g *checkoutGate) changed(root string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	entry := g.entries[root]
	if entry == nil {
		return ErrPlanStale
	}
	entry.revision = ""
	revision, err := g.freshRevision()
	if err != nil {
		return err
	}
	entry.revision = revision
	return nil
}

// WithCheckoutMutation invalidates outstanding runtime reviews even if the
// checkout operation returns an error after modifying some files.
func (s *Service) WithCheckoutMutation(ctx context.Context, id string, action func() error) error {
	return s.WithStoppedProject(ctx, id, func() error {
		project, _, err := s.resolve(ctx, id)
		if err != nil {
			return err
		}
		actionErr := action()
		return errors.Join(actionErr, s.checkouts.changed(project.Root))
	})
}

// CheckReview rejects a changed checkout or manifest before queuing/executing
// an action confirmed against a previous runtime preview.
func CheckReview(plan domain.Plan, checkoutRevision, manifestHash string) error {
	if checkoutRevision != "" && checkoutRevision != plan.SourceCheckoutRevision || manifestHash != "" && manifestHash != plan.SourceManifestHash {
		return ErrPlanStale
	}
	return nil
}
