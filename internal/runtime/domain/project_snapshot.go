package domain

// ProjectSnapshot is one background configuration observation. It contains no
// runtime status cache: drivers still observe health, events and metrics live.
type ProjectSnapshot struct {
	Projects map[string]ProjectRuntime
	Failures map[string]error
	Err      error
}
