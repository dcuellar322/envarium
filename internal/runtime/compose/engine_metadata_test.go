package compose

import (
	"context"
	"errors"
	"testing"

	"envarium.dev/envarium/internal/runtime/domain"
	"github.com/moby/moby/client"
)

func TestInspectRefreshesEngineVersionAndReportsMetadataFailure(t *testing.T) {
	t.Parallel()
	engine := &fakeEngine{version: client.ServerVersionResult{Version: "first"}}
	driver := &Driver{engine: fakeConnector{engine: engine}, managed: newManagedContainers()}
	project := domain.ProjectRuntime{ProjectID: "one"}
	for _, version := range []string{"first", "second"} {
		engine.version.Version = version
		observation, err := driver.inspect(context.Background(), project, normalizedConfig{})
		if err != nil || !observation.Engine.Connected || observation.Engine.ServerVersion != version {
			t.Fatalf("fresh engine metadata = %#v, %v", observation, err)
		}
	}
	engine.versionErr = errors.New("version metadata unavailable")
	observation, err := driver.inspect(context.Background(), project, normalizedConfig{})
	if err != nil || observation.Engine.Connected || observation.State != domain.StateUnknown {
		t.Fatalf("metadata failure = %#v, %v", observation, err)
	}
	if engine.versionCalls != 3 {
		t.Fatalf("version calls = %d, want 3", engine.versionCalls)
	}
}

func TestMetricSamplingDoesNotRequireUnusedServerMetadata(t *testing.T) {
	t.Parallel()
	engine := &fakeEngine{versionErr: errors.New("unused server metadata failed")}
	driver := &Driver{engine: fakeConnector{engine: engine}}
	sink := &composeMetricSink{}
	if err := driver.streamMetrics(context.Background(), domain.MetricRequest{}, normalizedConfig{}, sink); err != nil {
		t.Fatal(err)
	}
	if engine.versionCalls != 0 {
		t.Fatalf("metric-only version calls = %d, want 0", engine.versionCalls)
	}
}
