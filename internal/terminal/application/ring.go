package application

import "sync"

type byteRing struct {
	mu        sync.Mutex
	data      []byte
	limit     int
	head      int
	size      int
	truncated bool
}

func newByteRing(limit int) *byteRing { return &byteRing{limit: limit} }

// Write copies only incoming bytes and reports whether any scrollback was discarded.
func (r *byteRing) Write(value []byte) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(value) == 0 {
		return r.truncated
	}
	if r.data == nil {
		r.data = make([]byte, r.limit)
	}
	if len(value) >= r.limit {
		r.truncated = r.truncated || r.size+len(value) > r.limit
		copy(r.data, value[len(value)-r.limit:])
		r.head, r.size = 0, r.limit
		return r.truncated
	}
	overflow := max(0, r.size+len(value)-r.limit)
	if overflow > 0 {
		r.head = (r.head + overflow) % r.limit
		r.size -= overflow
		r.truncated = true
	}
	start := (r.head + r.size) % r.limit
	copied := copy(r.data[start:], value)
	copy(r.data, value[copied:])
	r.size += len(value)
	return r.truncated
}

func (r *byteRing) Snapshot() ([]byte, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]byte, r.size)
	copied := copy(result, r.data[r.head:min(len(r.data), r.head+r.size)])
	copy(result[copied:], r.data[:r.size-copied])
	return result, r.truncated
}
