package process

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
)

func TestLogBufferCountByteBoundsAndOrdering(t *testing.T) {
	for _, capacity := range []int{1, 7, 10000} {
		buffer := newLogBuffer(capacity)
		var want []domain.LogEntry
		for index := 0; index < 500; index++ {
			entry := domain.LogEntry{Message: fmt.Sprintf("%d:%s", index, strings.Repeat("x", index*31)), Attributes: map[string]string{"source": "fixture"}}
			want = append(want, entry)
			total := 0
			for _, item := range want {
				total += logEntryBytes(item)
			}
			for len(want) > capacity || total > defaultLogByteCapacity {
				total -= logEntryBytes(want[0])
				want = want[1:]
			}
			buffer.add(entry)
			got := buffer.snapshot(10000)
			if len(got) != len(want) || buffer.bytes != total {
				t.Fatalf("capacity %d append %d: count %d/%d bytes %d/%d", capacity, index, len(got), len(want), buffer.bytes, total)
			}
			for i := range got {
				if got[i].Message != want[i].Message {
					t.Fatalf("order differs at %d", i)
				}
			}
			if buffer.size > capacity || buffer.bytes > defaultLogByteCapacity {
				t.Fatal("capture bounds exceeded")
			}
		}
		if got := buffer.snapshot(2); len(got) != min(2, len(want)) || got[len(got)-1].Message != want[len(want)-1].Message {
			t.Fatal("tail ordering differs")
		}
	}
}

func TestLogBufferOversizedEntryIsLiveWithoutRetainingPayload(t *testing.T) {
	buffer := newLogBuffer(10)
	live, cancel := buffer.subscribe()
	defer cancel()
	entry := domain.LogEntry{Message: strings.Repeat("x", defaultLogByteCapacity+1)}
	buffer.add(entry)
	if got := <-live; got.Message != entry.Message {
		t.Fatal("live entry changed")
	}
	if len(buffer.snapshot(10)) != 0 || buffer.bytes != 0 {
		t.Fatal("oversized entry retained")
	}
}

func TestNativeCaptureReusesServiceBufferAcrossRuns(t *testing.T) {
	store := newMemoryRunStore()
	driver := newDriver(context.Background(), store, nil, nil)
	buffer := driver.serviceLogBuffer("project", "api")
	live, cancel := buffer.subscribe()
	defer cancel()
	for index := range 100 {
		id := fmt.Sprintf("run-%d", index)
		if err := store.CreateRun(context.Background(), domain.RunRecord{ID: id, ProjectID: "project", ServiceID: "api"}); err != nil {
			t.Fatal(err)
		}
		next := driver.serviceLogBuffer("project", "api")
		if next != buffer {
			t.Fatal("restart allocated another capture buffer")
		}
		next.add(domain.LogEntry{RunID: id, Message: id})
		if got := <-live; got.RunID != id {
			t.Fatal("follow lost restarted run")
		}
	}
	selected, err := driver.selectLogBuffers(context.Background(), domain.LogRequest{Project: domain.ProjectRuntime{ProjectID: "project", Services: []domain.ServiceDeclaration{{ID: "api"}}}})
	if err != nil || len(selected) != 1 || len(driver.logs) != 1 {
		t.Fatalf("buffers: %d retained: %d, %v", len(selected), len(driver.logs), err)
	}
	if driver.serviceLogBuffer("project", "worker") == buffer || driver.serviceLogBuffer("other", "api") == buffer {
		t.Fatal("service histories mixed")
	}
}

type failingLogSink struct{ err error }

func (s failingLogSink) WriteLog(context.Context, domain.LogEntry) error { return s.err }

func TestFollowLogBuffersStopsWorkersOnSinkFailureAndCancellation(t *testing.T) {
	for _, failure := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		buffers := []*logBuffer{newLogBuffer(10), newLogBuffer(10)}
		expected := context.Canceled
		if failure {
			expected = errors.New("sink failed")
		}
		done := make(chan error, 1)
		go func() { done <- followLogBuffers(ctx, buffers, failingLogSink{expected}) }()
		deadline := time.Now().Add(time.Second)
		for {
			buffers[0].mu.RLock()
			count := len(buffers[0].subscribers)
			buffers[0].mu.RUnlock()
			if count > 0 {
				break
			}
			if time.Now().After(deadline) {
				cancel()
				t.Fatal("subscriber not attached")
			}
			time.Sleep(time.Millisecond)
		}
		if failure {
			for range 1000 {
				for _, buffer := range buffers {
					buffer.add(domain.LogEntry{Message: "line"})
				}
			}
		} else {
			cancel()
		}
		select {
		case err := <-done:
			if !errors.Is(err, expected) {
				t.Fatalf("error=%v", err)
			}
		case <-time.After(time.Second):
			cancel()
			t.Fatal("workers failed to stop")
		}
		cancel()
		for _, buffer := range buffers {
			buffer.mu.RLock()
			count := len(buffer.subscribers)
			buffer.mu.RUnlock()
			if count != 0 {
				t.Fatal("subscriber leaked")
			}
		}
	}
}

func TestConfiguredNativeLogEntryLimit(t *testing.T) {
	driver := NewDriverWithLogCapacity(context.Background(), newMemoryRunStore(), 2)
	buffer := driver.serviceLogBuffer("project", "api")
	for index := range 3 {
		buffer.add(domain.LogEntry{Message: fmt.Sprint(index)})
	}
	entries := buffer.snapshot(10)
	if len(entries) != 2 || entries[0].Message != "1" || entries[1].Message != "2" {
		t.Fatalf("configured tail=%+v", entries)
	}
}

func TestNativeLogFollowWaitsForFirstRunWithoutHistoryQueries(t *testing.T) {
	// A nil store makes any accidental history query fail immediately.
	driver := newDriver(context.Background(), nil, nil, nil)
	project := domain.ProjectRuntime{ProjectID: "project", Services: []domain.ServiceDeclaration{{ID: "api"}, {ID: "worker"}}}
	request := domain.LogRequest{Project: project, Service: "api", Follow: true, Tail: 100}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	entries := make(chan domain.LogEntry, 1)
	go func() { result <- driver.StreamLogs(ctx, request, channelLogSink{entries}) }()
	buffer := driver.serviceLogBuffer("project", "api")
	deadline := time.Now().Add(time.Second)
	for {
		buffer.mu.RLock()
		count := len(buffer.subscribers)
		buffer.mu.RUnlock()
		if count > 0 {
			break
		}
		select {
		case err := <-result:
			t.Fatalf("follow ended before first run: %v", err)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal("subscriber did not attach")
		}
		time.Sleep(time.Millisecond)
	}
	buffer.add(domain.LogEntry{RunID: "first", Message: "started"})
	select {
	case entry := <-entries:
		if entry.RunID != "first" {
			t.Fatal("wrong run")
		}
	case <-time.After(time.Second):
		t.Fatal("first run log was not forwarded")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("follow failed to cancel")
	}
	if len(driver.logs) != 1 {
		t.Fatal("unselected services allocated capture buffers")
	}
	canceled, cancelNow := context.WithCancel(context.Background())
	cancelNow()
	request.Service = "worker"
	if _, err := driver.selectLogBuffers(canceled, request); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if len(driver.logs) != 1 {
		t.Fatal("canceled selection allocated capture buffer")
	}
}

type channelLogSink struct{ entries chan<- domain.LogEntry }

func (s channelLogSink) WriteLog(ctx context.Context, entry domain.LogEntry) error {
	select {
	case s.entries <- entry:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
