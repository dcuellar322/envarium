package process

import (
	"context"
	"sort"
	"sync"

	"envarium.dev/envarium/internal/runtime/domain"
)

type logBuffer struct {
	mu          sync.RWMutex
	capacity    int
	head        int
	size        int
	bytes       int
	entries     []domain.LogEntry
	subscribers map[chan domain.LogEntry]struct{}
}

func newLogBuffer(capacity int) *logBuffer {
	return &logBuffer{capacity: capacity, subscribers: make(map[chan domain.LogEntry]struct{})}
}

func (b *logBuffer) add(entry domain.LogEntry) {
	b.mu.Lock()
	// Count metadata conservatively as well as message/attribute payloads.
	size := logEntryBytes(entry)
	if size <= defaultLogByteCapacity {
		for b.size > 0 && (b.size == b.capacity || b.bytes+size > defaultLogByteCapacity) {
			b.bytes -= logEntryBytes(b.entries[b.head])
			b.entries[b.head] = domain.LogEntry{}
			b.head = (b.head + 1) % len(b.entries)
			b.size--
		}
		if b.size == len(b.entries) {
			entries := make([]domain.LogEntry, min(b.capacity, max(16, len(b.entries)*2)))
			for index := 0; index < b.size; index++ {
				entries[index] = b.entries[(b.head+index)%len(b.entries)]
			}
			b.entries, b.head = entries, 0
		}
		b.entries[(b.head+b.size)%len(b.entries)] = entry
		b.size++
		b.bytes += size
	}
	for subscriber := range b.subscribers {
		select {
		case subscriber <- entry:
		default:
		}
	}
	b.mu.Unlock()
}

func (b *logBuffer) snapshot(tail int) []domain.LogEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()
	count := min(b.size, max(0, tail))
	result := make([]domain.LogEntry, 0, count)
	for index := b.size - count; index < b.size; index++ {
		result = append(result, b.entries[(b.head+index)%len(b.entries)])
	}
	return result
}

func (b *logBuffer) subscribe() (<-chan domain.LogEntry, func()) {
	channel := make(chan domain.LogEntry, 128)
	b.mu.Lock()
	b.subscribers[channel] = struct{}{}
	b.mu.Unlock()
	return channel, func() {
		b.mu.Lock()
		delete(b.subscribers, channel)
		close(channel)
		b.mu.Unlock()
	}
}

func (d *Driver) streamLogs(ctx context.Context, request domain.LogRequest, sink domain.LogSink) error {
	selected, err := d.selectLogBuffers(ctx, request)
	if err != nil {
		return err
	}
	if err := writeLogSnapshot(ctx, selected, request.Tail, sink); err != nil {
		return err
	}
	if !request.Follow || len(selected) == 0 {
		return nil
	}
	return followLogBuffers(ctx, selected, sink)
}

func (d *Driver) selectLogBuffers(ctx context.Context, request domain.LogRequest) ([]*logBuffer, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	selected := []*logBuffer{}
	for _, service := range request.Project.Services {
		if request.Service != "" && service.ID != request.Service {
			continue
		}
		if request.Follow {
			// Subscribe before the first run as well as across restarts. An idle
			// service needs no repeated run-history queries or collector reconnects.
			selected = append(selected, d.serviceLogBuffer(request.Project.ProjectID, service.ID))
			continue
		}
		d.mu.RLock()
		buffer := d.logs[serviceKey(request.Project.ProjectID, service.ID)]
		d.mu.RUnlock()
		if buffer != nil {
			selected = append(selected, buffer)
		}
	}
	return selected, nil
}

func writeLogSnapshot(ctx context.Context, selected []*logBuffer, tail int, sink domain.LogSink) error {
	entries := []domain.LogEntry{}
	for _, buffer := range selected {
		entries = append(entries, buffer.snapshot(tail)...)
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Timestamp.Before(entries[j].Timestamp) })
	if len(entries) > tail {
		entries = entries[len(entries)-tail:]
	}
	for _, entry := range entries {
		if err := sink.WriteLog(ctx, entry); err != nil {
			return err
		}
	}
	return nil
}

func followLogBuffers(ctx context.Context, selected []*logBuffer, sink domain.LogSink) error {
	ctx, cancelContext := context.WithCancel(ctx)
	var workers sync.WaitGroup
	type subscription struct {
		entries <-chan domain.LogEntry
		cancel  func()
	}
	subscriptions := make([]subscription, 0, len(selected))
	merged := make(chan domain.LogEntry, 128)
	for _, buffer := range selected {
		entries, cancel := buffer.subscribe()
		subscriptions = append(subscriptions, subscription{entries: entries, cancel: cancel})
		workers.Add(1)
		go func(source <-chan domain.LogEntry) {
			defer workers.Done()
			for entry := range source {
				select {
				case merged <- entry:
				case <-ctx.Done():
					return
				}
			}
		}(entries)
	}
	defer func() {
		cancelContext()
		for _, item := range subscriptions {
			item.cancel()
		}
		workers.Wait()
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case entry := <-merged:
			if err := sink.WriteLog(ctx, entry); err != nil {
				return err
			}
		}
	}
}

// logEntryBytes bounds retained payload plus fixed entry/map overhead.
func logEntryBytes(entry domain.LogEntry) int {
	size := 256 + len(entry.Message) + len(entry.ProjectID) + len(entry.ServiceID) + len(entry.RunID) + len(entry.OperationID) + len(entry.Source) + len(entry.Stream) + len(entry.Level)
	for key, value := range entry.Attributes {
		size += 32 + len(key) + len(value)
	}
	return size
}

// serviceLogBuffer shares one bounded history and follow stream across restarts.
func (d *Driver) serviceLogBuffer(projectID, serviceID string) *logBuffer {
	d.mu.Lock()
	defer d.mu.Unlock()
	key := serviceKey(projectID, serviceID)
	if buffer := d.logs[key]; buffer != nil {
		return buffer
	}
	buffer := newLogBuffer(d.logCapacity)
	d.logs[key] = buffer
	return buffer
}
