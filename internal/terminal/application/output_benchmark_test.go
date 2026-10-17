package application

import (
	"testing"
	"time"

	"envarium.dev/envarium/internal/terminal/domain"
)

func BenchmarkTerminalOutput(b *testing.B) {
	managed := newManagedSession(domain.Session{}, nil, 1<<20)
	managed.scrollback.Write(make([]byte, 1<<20))
	chunk := make([]byte, 32<<10)
	at := time.Now()
	b.ReportAllocs()
	b.SetBytes(int64(len(chunk)))
	b.ResetTimer()
	for b.Loop() {
		managed.appendOutput(chunk, at)
	}
}
