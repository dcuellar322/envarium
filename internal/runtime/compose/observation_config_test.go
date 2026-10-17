package compose

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
)

func TestObservationConfigCachesSuccessFailureAndExpires(t *testing.T) {
	for _, fail := range []bool{false, true} {
		cache := newObservationConfigs()
		now := time.Now()
		cache.now = func() time.Time { return now }
		calls := 0
		expected := errors.New("Docker unavailable")
		normalize := func() (normalizedConfig, error) {
			calls++
			if fail {
				return normalizedConfig{}, expected
			}
			return normalizedConfig{ProjectName: "fixture", Services: []string{"api"}, BaseArgs: []string{"compose"}}, nil
		}
		for range 5 {
			config, err := cache.load(context.Background(), domain.ProjectRuntime{ProjectID: "one"}, normalize)
			if fail && !errors.Is(err, expected) || !fail && err != nil {
				t.Fatal(err)
			}
			if !fail {
				if config.Services[0] != "api" {
					t.Fatal("cached config was mutated")
				}
				config.Services[0] = "changed"
			}
		}
		if calls != 1 {
			t.Fatalf("normalizations=%d", calls)
		}
		now = now.Add(observationConfigTTL)
		fail = false
		config, err := cache.load(context.Background(), domain.ProjectRuntime{ProjectID: "one"}, normalize)
		if err != nil || calls != 2 || config.ProjectName != "fixture" {
			t.Fatalf("expiry/recovery: %v calls %d", err, calls)
		}
	}
}

func TestObservationConfigKeyAndCapacity(t *testing.T) {
	cache := newObservationConfigs()
	calls := 0
	normalize := func() (normalizedConfig, error) { calls++; return normalizedConfig{}, nil }
	projects := []domain.ProjectRuntime{
		{ProjectID: "one", ManifestHash: "before", Root: "/one"},
		{ProjectID: "one", ManifestHash: "after", Root: "/one"},
		{ProjectID: "one", ManifestHash: "after", Root: "/two"},
		{ProjectID: "one", ManifestHash: "after", Root: "/two", Compose: &domain.ComposeRuntime{Context: "other"}},
	}
	for _, project := range projects {
		if _, err := cache.load(context.Background(), project, normalize); err != nil {
			t.Fatal(err)
		}
	}
	if calls != len(projects) {
		t.Fatal("changed project input reused stale config")
	}
	for index := range maximumObservationConfigs * 2 {
		if _, err := cache.load(context.Background(), domain.ProjectRuntime{ProjectID: fmt.Sprint(index)}, normalize); err != nil {
			t.Fatal(err)
		}
	}
	if len(cache.entries) > maximumObservationConfigs {
		t.Fatal("cache grows with historical project count")
	}
}

func TestObservationConfigCoalescesConcurrentReaders(t *testing.T) {
	cache := newObservationConfigs()
	var calls atomic.Int32
	started, release := make(chan struct{}), make(chan struct{})
	normalize := func() (normalizedConfig, error) {
		if calls.Add(1) == 1 {
			close(started)
		}
		<-release
		return normalizedConfig{ProjectName: "fixture"}, nil
	}
	var group sync.WaitGroup
	for range 20 {
		group.Add(1)
		go func() {
			defer group.Done()
			config, err := cache.load(context.Background(), domain.ProjectRuntime{}, normalize)
			if err != nil || config.ProjectName != "fixture" {
				t.Errorf("config=%+v error=%v", config, err)
			}
		}()
	}
	<-started
	close(release)
	group.Wait()
	if calls.Load() != 1 {
		t.Fatalf("normalizations=%d", calls.Load())
	}
}

func TestObservationConfigDoesNotCacheCanceledRequests(t *testing.T) {
	cache := newObservationConfigs()
	ctx, cancel := context.WithCancel(context.Background())
	started, finished := make(chan struct{}), make(chan struct{})
	result := make(chan error, 1)
	go func() {
		_, err := cache.load(ctx, domain.ProjectRuntime{}, func() (normalizedConfig, error) {
			close(started)
			<-ctx.Done()
			close(finished)
			return normalizedConfig{}, ctx.Err()
		})
		result <- err
	}()
	<-started
	cancel()
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	<-finished
	// Join the singleflight operation before testing recovery.
	_, joinErr, _ := cache.requests.Do(mustObservationKey(t), func() (any, error) { return nil, nil })
	if joinErr != nil {
		t.Fatal(joinErr)
	}
	calls := 0
	_, err := cache.load(context.Background(), domain.ProjectRuntime{}, func() (normalizedConfig, error) { calls++; return normalizedConfig{}, nil })
	if err != nil || calls != 1 {
		t.Fatalf("canceled normalization retained: calls=%d error=%v", calls, err)
	}
	canceled, cancelAgain := context.WithCancel(context.Background())
	cancelAgain()
	if _, err := cache.load(canceled, domain.ProjectRuntime{}, func() (normalizedConfig, error) {
		t.Fatal("canceled request invoked CLI")
		return normalizedConfig{}, nil
	}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func mustObservationKey(t *testing.T) string {
	t.Helper()
	// The empty runtime document is encoded exactly as it is for the cache key.
	document, err := json.Marshal(domain.ProjectRuntime{})
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(document)
	return hex.EncodeToString(digest[:])
}

func TestComposeObservationCacheDoesNotCacheLifecyclePlans(t *testing.T) {
	var normalizations atomic.Int32
	runner := scriptedRunner{run: func(command domain.Command, stdout, _ io.Writer) error {
		joined := strings.Join(command.Arguments, " ")
		if strings.Contains(joined, "context inspect default") {
			_, err := io.WriteString(stdout, `[{"Name":"selected","Endpoints":{"docker":{"Host":"unix:///tmp/docker.sock"}},"Storage":{}}]`)
			return err
		}
		if strings.Contains(joined, "config --format json") {
			normalizations.Add(1)
			_, err := io.WriteString(stdout, `{"name":"fixture","services":{"api":{}}}`)
			return err
		}
		return fmt.Errorf("unexpected command %s", joined)
	}}
	driver := NewDriver()
	driver.config = configReader{runner: runner, contexts: contextResolver{runner: runner}}
	project := domain.ProjectRuntime{ProjectID: "fixture", Root: t.TempDir(), Compose: &domain.ComposeRuntime{Files: []string{"compose.yaml"}, Context: "default"}}
	if err := os.WriteFile(filepath.Join(project.Root, "compose.yaml"), []byte("services:\n  api:\n    image: alpine\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		if _, err := driver.observationConfig(context.Background(), project); err != nil {
			t.Fatal(err)
		}
	}
	if normalizations.Load() != 0 {
		t.Fatal("read observation invoked CLI")
	}
	for range 2 {
		if _, err := driver.Plan(context.Background(), domain.PlanRequest{Project: project, Action: domain.ActionStart}); err != nil {
			t.Fatal(err)
		}
	}
	if normalizations.Load() != 2 {
		t.Fatal("lifecycle plan reused observation cache")
	}
}
