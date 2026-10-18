package domain

// IsActive identifies states for which current resource samples are meaningful.
func (s ProjectState) IsActive() bool {
	switch s {
	case StateStarting, StateRunning, StateRunningExternal, StatePartiallyRunning, StateDegraded, StatePaused:
		return true
	case StateUnknown, StateStopped, StateStopping, StateFailed:
		return false
	}
	return false
}
