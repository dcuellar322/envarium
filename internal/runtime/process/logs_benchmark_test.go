package process

import (
	"testing"

	"envarium.dev/envarium/internal/runtime/domain"
)

func BenchmarkNativeLogAppend(b *testing.B) {
	buffer := newLogBuffer(defaultLogCapacity)
	entry := domain.LogEntry{Message: "development server request completed"}
	for range defaultLogCapacity {
		buffer.add(entry)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		buffer.add(entry)
	}
}
