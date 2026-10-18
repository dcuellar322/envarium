package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"envarium.dev/envarium/internal/transport/contract/generated"
	"envarium.dev/envarium/internal/transport/httpclient"
)

type desktopStatusFake struct {
	chunks                     [][]string
	batchErr                   error
	omitHealth, omitRuntime    map[string]bool
	runtimeCalls, healthCalls  int
	cancelBatch, cancelRuntime context.CancelFunc
}

func (f *desktopStatusFake) StatusBatch(_ context.Context, ids []string) (generated.ProjectObservationBatch, error) {
	f.chunks = append(f.chunks, append([]string(nil), ids...))
	if f.cancelBatch != nil {
		f.cancelBatch()
	}
	items := []map[string]any{}
	for _, id := range ids {
		item := map[string]any{"projectId": id}
		if !f.omitRuntime[id] {
			item["runtime"] = map[string]any{"state": "degraded"}
		}
		if !f.omitHealth[id] {
			item["health"] = map[string]any{"status": "unhealthy"}
		}
		items = append(items, item)
	}
	// A response outside the requested scope must never populate another project.
	items = append(items, map[string]any{"projectId": "untrusted", "runtime": map[string]any{"state": "running"}})
	raw, _ := json.Marshal(map[string]any{"kind": "status", "projects": items})
	var result generated.ProjectObservationBatch
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, err
	}
	return result, f.batchErr
}

func (f *desktopStatusFake) Runtime(context.Context, string) (generated.RuntimeObservation, error) {
	f.runtimeCalls++
	if f.cancelRuntime != nil {
		f.cancelRuntime()
	}
	return generated.RuntimeObservation{State: "running"}, nil
}

func (f *desktopStatusFake) Health(context.Context, string) (generated.ProjectHealth, error) {
	f.healthCalls++
	return generated.ProjectHealth{Status: "healthy"}, nil
}

func TestDesktopProjectsBatchesTrustedIDsWithoutChangingOrder(t *testing.T) {
	t.Parallel()
	projects := []generated.Project{{Id: "untrusted"}}
	for i := range 130 {
		projects = append(projects, generated.Project{Id: fmt.Sprint(i), TrustState: generated.ProjectTrustStateTrusted})
	}
	client := &desktopStatusFake{}
	result, err := desktopProjects(context.Background(), client, projects)
	if err != nil || len(result) != len(projects) || len(client.chunks) != 3 {
		t.Fatalf("snapshot count=%d chunks=%d err=%v", len(result), len(client.chunks), err)
	}
	if len(client.chunks[0]) != 64 || len(client.chunks[1]) != 64 || len(client.chunks[2]) != 2 || client.runtimeCalls != 0 || client.healthCalls != 0 {
		t.Fatalf("batch boundaries or legacy calls: %#v", client)
	}
	for i, item := range result {
		if item.Project.Id != projects[i].Id {
			t.Fatal("project order changed")
		}
		if i == 0 {
			if item.Runtime != nil || item.Health != nil {
				t.Fatal("untrusted project was populated")
			}
		} else if item.Runtime == nil || item.Health == nil || len(item.Warnings) != 0 || item.Runtime.State != "degraded" || item.Health.Status != "unhealthy" {
			t.Fatalf("status %d = %#v", i, item)
		}
	}
}

func TestDesktopProjectsRetainsPartialWarningsAndDoesNotRetryDeniedReads(t *testing.T) {
	t.Parallel()
	projects := []generated.Project{{Id: "one", TrustState: generated.ProjectTrustStateTrusted}}
	client := &desktopStatusFake{omitHealth: map[string]bool{"one": true}}
	result, err := desktopProjects(context.Background(), client, projects)
	if err != nil || result[0].Runtime == nil || result[0].Health != nil || len(result[0].Warnings) != 1 {
		t.Fatalf("partial = %#v, %v", result, err)
	}
	for _, status := range []int{http.StatusForbidden, http.StatusServiceUnavailable} {
		client = &desktopStatusFake{batchErr: &httpclient.APIError{Status: status, Detail: "private failure"}}
		result, err = desktopProjects(context.Background(), client, projects)
		if err != nil || len(result[0].Warnings) != 2 || result[0].Runtime != nil || client.runtimeCalls != 0 || client.healthCalls != 0 {
			t.Fatalf("failed batch = %#v, %v", result, err)
		}
	}
}

func TestDesktopProjectsDoesNotFanOutForOldDaemonAndHonorsCancellation(t *testing.T) {
	t.Parallel()
	projects := []generated.Project{{Id: "one", TrustState: generated.ProjectTrustStateTrusted}}
	client := &desktopStatusFake{batchErr: &httpclient.APIError{Status: http.StatusNotFound}}
	result, err := desktopProjects(context.Background(), client, projects)
	if err != nil || len(result[0].Warnings) != 2 || client.runtimeCalls != 0 || client.healthCalls != 0 {
		t.Fatalf("unsupported=%#v err=%v client=%#v", result, err, client)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client = &desktopStatusFake{cancelBatch: cancel}
	result, err = desktopProjects(ctx, client, projects)
	if !errors.Is(err, context.Canceled) || result != nil {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}
