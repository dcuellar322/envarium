package application

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	actions "envarium.dev/envarium/internal/actions/domain"
	runtime "envarium.dev/envarium/internal/runtime/domain"
	git "envarium.dev/envarium/internal/sourcecontrol/domain"
)

type reads struct {
	runtime atomic.Int32
	git     atomic.Int32
	actions atomic.Int32
	current atomic.Int32
	maximum atomic.Int32
	started chan struct{}
	release chan struct{}
}

func (r *reads) Inspect(ctx context.Context, id string) (runtime.Observation, error) {
	r.runtime.Add(1)
	if r.started != nil {
		current := r.current.Add(1)
		defer r.current.Add(-1)
		for {
			old := r.maximum.Load()
			if current <= old || r.maximum.CompareAndSwap(old, current) {
				break
			}
		}
		r.started <- struct{}{}
		select {
		case <-r.release:
		case <-ctx.Done():
			return runtime.Observation{}, ctx.Err()
		}
	}
	return runtime.Observation{ProjectID: id}, nil
}
func (r *reads) GetStatus(_ context.Context, id string) (git.State, error) {
	r.git.Add(1)
	if id == "denied" {
		return git.State{}, errors.New("private denied detail")
	}
	return git.State{ProjectID: id}, nil
}
func (r *reads) List(_ context.Context, id string) (actions.ProjectActions, error) {
	r.actions.Add(1)
	return actions.ProjectActions{ProjectID: id}, nil
}
func TestBatchKindsAndPermissionFailuresRemainIsolated(t *testing.T) {
	t.Parallel()
	reader := &reads{}
	service := NewService(reader, reader, reader, nil)
	for _, kind := range []string{"runtime", "git-status", "actions"} {
		result, err := service.Read(context.Background(), kind, []string{"one"})
		if err != nil || len(result.Projects) != 1 || result.Projects[0].ProjectID != "one" {
			t.Fatalf("%s: %#v %v", kind, result, err)
		}
	}
	result, err := service.Read(context.Background(), "git-status", []string{"denied", "one"})
	if err != nil || result.Projects[0].Warning == "" || result.Projects[0].Git != nil || result.Projects[1].Git == nil {
		t.Fatalf("permission: %#v %v", result, err)
	}
	if reader.runtime.Load() != 1 || reader.actions.Load() != 1 || reader.git.Load() != 3 {
		t.Fatal("unrelated kinds were read")
	}
	if result.Projects[0].Warning == "private denied detail" {
		t.Fatal("private error leaked")
	}
	for _, ids := range [][]string{nil, {"one", "one"}, make([]string, 65)} {
		if _, err := service.Read(context.Background(), "runtime", ids); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("bounds: %v", err)
		}
	}
	if _, err := service.Read(context.Background(), "execute", []string{"one"}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatal("unrestricted kind")
	}
}
func TestBatchConcurrencyAndCancellation(t *testing.T) {
	t.Parallel()
	for _, cancelEarly := range []bool{false, true} {
		reader := &reads{started: make(chan struct{}, 16), release: make(chan struct{})}
		service := NewService(reader, reader, reader, nil)
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		ids := make([]string, 11)
		for i := range ids {
			ids[i] = fmt.Sprint(i)
		}
		go func() { _, err := service.Read(ctx, "runtime", ids); done <- err }()
		for range 4 {
			<-reader.started
		}
		if cancelEarly {
			cancel()
		} else {
			close(reader.release)
		}
		err := <-done
		cancel()
		if cancelEarly && !errors.Is(err, context.Canceled) || !cancelEarly && err != nil {
			t.Fatalf("cancel=%v err=%v", cancelEarly, err)
		}
		if reader.maximum.Load() > 4 || reader.current.Load() != 0 {
			t.Fatal("worker bound/shutdown violated")
		}
	}
}
