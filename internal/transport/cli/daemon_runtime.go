package cli

import (
	"os"
	"runtime"
)

// configureProcessRuntime avoids excess CPU-worker wakeups in the I/O-oriented
// control plane and thin clients. This changes no child environment variables;
// project processes retain their configured Go parallelism. Goroutine/I/O
// concurrency and monitoring cadence are unchanged.
// An explicit Go runtime override always wins. See ADR-0022.
func configureProcessRuntime() func() {
	if os.Getenv("GOMAXPROCS") != "" {
		return func() {}
	}
	previous := runtime.GOMAXPROCS(1)
	return func() { runtime.GOMAXPROCS(previous) }
}
