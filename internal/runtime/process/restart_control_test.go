package process

import (
	"context"
	"envarium.dev/envarium/internal/runtime/domain"
	"path/filepath"
	"testing"
	"time"
)

func TestStoppingNativeBackoffJoinsRestartBeforeReportingStopped(t *testing.T) {
	for _, action := range []domain.Action{domain.ActionStop, domain.ActionKill} {
		t.Run(string(action), func(t *testing.T) {
			driver, store, project, cancel := nativeTestDriver(t, "crash-once", freePort(t), domain.RestartPolicy{Mode: "on-failure", MaxRetries: 1, BackoffSeconds: 60})
			defer cancel()
			driver.now = func() time.Time { return time.Now().UTC().Add(time.Hour) } // Exercise observation beyond identity-handoff grace.
			project.Process.Processes[0].Environment["MARKER"] = filepath.Join(t.TempDir(), "crash")
			executeProcessAction(t, driver, project, domain.ActionStart)
			var managed *managedRun
			var done chan struct{}
			deadline := time.Now().Add(5 * time.Second)
			for time.Now().Before(deadline) {
				driver.mu.RLock()
				managed = driver.managed[serviceKey(project.ProjectID, "web")]
				driver.mu.RUnlock()
				if managed != nil {
					managed.mu.Lock()
					done = managed.restartDone
					managed.mu.Unlock()
					if done != nil {
						break
					}
				}
				time.Sleep(10 * time.Millisecond)
			}
			if done == nil {
				t.Fatal("helper never entered restart backoff")
			}
			observation, err := driver.Inspect(context.Background(), project)
			if err != nil || observation.State != domain.StateStarting {
				t.Fatalf("pending restart appeared stopped: %#v %v", observation, err)
			}
			executeProcessAction(t, driver, project, domain.ActionStart) // Must not duplicate a pending owned run.
			runs, _ := store.ListProjectRuns(context.Background(), project.ProjectID)
			if len(runs) != 1 || runs[0].EndedAt != nil {
				t.Fatal("observation/start lost pending ownership", runs)
			}
			started := time.Now()
			executeProcessAction(t, driver, project, action)
			if time.Since(started) > 5*time.Second {
				t.Fatal("stop waited for the backoff deadline")
			}
			select {
			case <-done:
			default:
				t.Fatal("stop returned with a restart goroutine still pending")
			}
			runs, _ = store.ListProjectRuns(context.Background(), project.ProjectID)
			if len(runs) != 1 || runs[0].EndedAt == nil || runs[0].RestartCount != 0 {
				t.Fatal("pending restart was not cancelled", runs)
			}
			waitForProcessState(t, driver, project, domain.StateStopped)
		})
	}
}
