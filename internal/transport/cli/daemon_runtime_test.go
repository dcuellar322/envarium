package cli

import (
	"runtime"
	"testing"
)

func TestDaemonRuntimeDefaultOverrideAndRestoration(t *testing.T) {
	// Runtime/environment changes are confined to this nonparallel test.
	previous := runtime.GOMAXPROCS(4)
	defer runtime.GOMAXPROCS(previous)
	t.Setenv("GOMAXPROCS", "")
	restore := configureProcessRuntime()
	if got := runtime.GOMAXPROCS(0); got != 1 {
		t.Fatalf("default=%d", got)
	}
	restore()
	if got := runtime.GOMAXPROCS(0); got != 4 {
		t.Fatalf("restored=%d", got)
	}
	t.Setenv("GOMAXPROCS", "4")
	restore = configureProcessRuntime()
	if got := runtime.GOMAXPROCS(0); got != 4 {
		t.Fatalf("explicit override changed=%d", got)
	}
	restore()
	if got := runtime.GOMAXPROCS(0); got != 4 {
		t.Fatalf("explicit override restoration=%d", got)
	}
}
