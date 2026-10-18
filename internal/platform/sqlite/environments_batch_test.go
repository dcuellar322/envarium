package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"envarium.dev/envarium/internal/environments/domain"
)

func TestEnvironmentSnapshotKeepsLeaseOwnershipOrderAndFreshness(t *testing.T) {
	ctx := context.Background()
	database, err := Open(ctx, filepath.Join(t.TempDir(), "environments.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = database.Close() }()
	now := time.Now().UTC()
	_, err = database.connection.ExecContext(ctx, `INSERT INTO projects
        (id,slug,display_name,description,trust_state,primary_location,manifest_revision,created_at,updated_at)
        VALUES ('project','project','Project','','trusted',?,1,?,?)`, t.TempDir(), formatTime(now), formatTime(now))
	if err != nil {
		t.Fatal(err)
	}
	repository := NewEnvironmentRepository(database)
	if got, err := repository.List(ctx); err != nil || len(got) != 0 {
		t.Fatalf("empty inventory=%v err=%v", got, err)
	}
	var inputs []domain.Environment
	for index, name := range []string{"primary", "feature", "empty"} {
		inputs = append(inputs, domain.Environment{
			ID: "env-" + name, ProjectID: "project", Name: name, Path: t.TempDir(), Primary: index == 0,
			Availability: domain.AvailabilityAvailable, State: domain.StateRegistered, Hostname: name + ".localhost",
			RegisteredAt: now, LastObservedAt: now, UpdatedAt: now,
			Allocation: domain.RuntimeAllocation{ComposeProjectName: "sy-" + name, PortLeaseNamespace: "worktree:" + name, PortOffset: index + 1},
		})
	}
	inputs[0].Allocation.PortLeases = []domain.PortLease{
		{PortID: "web", Protocol: "tcp", TargetPort: 80, HostPort: 15000},
		{PortID: "api", Protocol: "tcp", TargetPort: 8080, HostPort: 15001},
	}
	inputs[1].Allocation.PortLeases = []domain.PortLease{{PortID: "web", Protocol: "tcp", TargetPort: 80, HostPort: 15002}}
	if err := repository.ReplaceProject(ctx, "project", inputs); err != nil {
		t.Fatal(err)
	}
	items, err := repository.List(ctx)
	if err != nil || len(items) != 3 || items[0].Name != "primary" || items[1].Name != "empty" || items[2].Name != "feature" {
		t.Fatalf("stable inventory=%v err=%v", items, err)
	}
	for _, item := range items {
		one, err := repository.Get(ctx, item.ID)
		if err != nil || !reflect.DeepEqual(one, item) {
			t.Fatalf("batch/single mismatch: %v %v err=%v", item, one, err)
		}
	}
	if items[0].Allocation.PortLeases[0].PortID != "api" || len(items[1].Allocation.PortLeases) != 0 || items[2].Allocation.PortLeases[0].HostPort != 15002 {
		t.Fatal("lease ownership or ordering changed")
	}
	inputs[1].Allocation.PortLeases[0].HostPort = 15003
	if err := repository.Update(ctx, inputs[1]); err != nil {
		t.Fatal(err)
	}
	items, err = repository.List(ctx)
	if err != nil || items[2].Allocation.PortLeases[0].HostPort != 15003 {
		t.Fatal("batch retained a previous allocation")
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := repository.List(canceled); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation=%v", err)
	}
}
