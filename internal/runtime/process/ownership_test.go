package process

import (
	"context"
	"math"
	"testing"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
)

func TestBoundedPIDRejectsInvalidNativeValues(t *testing.T) {
	t.Parallel()
	for _, value := range []int{0, -1} {
		if _, err := boundedPID(value); err == nil {
			t.Fatalf("boundedPID(%d) accepted an invalid PID", value)
		}
	}
	if int64(math.MaxInt) > math.MaxInt32 {
		if _, err := boundedPID(math.MaxInt); err == nil {
			t.Fatal("boundedPID(MaxInt) accepted a PID wider than the durable format")
		}
	}
	if got, err := boundedPID(42); err != nil || got != 42 {
		t.Fatalf("boundedPID(42) = %d, %v", got, err)
	}
}

func TestGroupMembersPrefersOwnedProcessMembership(t *testing.T) {
	t.Parallel()
	startedAt := time.Now().UTC()
	inspector := inspectorFake{
		snapshots: map[int32]domain.ProcessIdentity{
			10: {PID: 10, ProcessGroup: 10, StartedAt: startedAt, Fingerprint: "parent"},
			11: {PID: 11, ProcessGroup: 11, StartedAt: startedAt, Fingerprint: "child"},
		},
		groups: map[int32][]domain.ProcessIdentity{
			10: {{PID: 99, ProcessGroup: 10, Fingerprint: "fallback"}},
		},
	}
	driver := newDriver(context.Background(), newMemoryRunStore(), inspector, &secretResolverFake{})
	ownership := &memberSourceOwnershipFake{group: 10, pids: []int32{10, 11}}

	members, err := driver.groupMembers(context.Background(), ownership, ownership.group)
	if err != nil {
		t.Fatal(err)
	}
	if ownership.calls != 1 || len(members) != 2 {
		t.Fatalf("membership calls = %d, members = %#v", ownership.calls, members)
	}
	for _, member := range members {
		if member.ProcessGroup != ownership.group || member.PID == 99 {
			t.Fatalf("member = %#v", member)
		}
	}
}

type memberSourceOwnershipFake struct {
	group int32
	pids  []int32
	calls int
}

func TestGroupTrackingPersistsNewAndChangedIdentitiesOnly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	started := time.Now().UTC()
	identity := domain.ProcessIdentity{PID: 10, ProcessGroup: 10, StartedAt: started, Fingerprint: "initial"}
	inspector := inspectorFake{snapshots: map[int32]domain.ProcessIdentity{10: identity}}
	store := &countingRunStore{memoryRunStore: newMemoryRunStore()}
	if err := store.CreateRun(ctx, domain.RunRecord{ID: "run-1"}); err != nil {
		t.Fatal(err)
	}
	driver := newDriver(ctx, store, inspector, &secretResolverFake{})
	managed := &managedRun{run: domain.RunRecord{ID: "run-1"}, group: 10,
		ownership: &memberSourceOwnershipFake{group: 10, pids: []int32{10}}}
	driver.recordGroupMembers(ctx, managed)
	driver.recordGroupMembers(ctx, managed)
	if store.writes != 1 {
		t.Fatalf("unchanged identity writes = %d", store.writes)
	}
	identity.Fingerprint = "after-exec"
	inspector.snapshots[10] = identity
	driver.recordGroupMembers(ctx, managed)
	if store.writes != 2 {
		t.Fatalf("changed identity was not persisted: writes = %d", store.writes)
	}
}

type countingRunStore struct {
	*memoryRunStore
	writes int
}

func (s *countingRunStore) RecordProcess(ctx context.Context, identity domain.ProcessIdentity) error {
	s.writes++
	return s.memoryRunStore.RecordProcess(ctx, identity)
}

func (o *memberSourceOwnershipFake) Group() int32    { return o.group }
func (*memberSourceOwnershipFake) Running() bool     { return true }
func (*memberSourceOwnershipFake) Signal(bool) error { return nil }
func (*memberSourceOwnershipFake) Close() error      { return nil }
func (o *memberSourceOwnershipFake) MemberPIDs() ([]int32, error) {
	o.calls++
	return append([]int32(nil), o.pids...), nil
}
