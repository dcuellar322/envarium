package application

import (
	"context"
	"errors"
	"testing"

	"envarium.dev/envarium/internal/runtime/domain"
)

type monitoringSource struct {
	calls   int
	err     error
	project domain.ProjectRuntime
}

func (s *monitoringSource) ResolveRuntime(ctx context.Context, _ string) (domain.ProjectRuntime, error) {
	s.calls++
	if ctx.Err() != nil {
		return domain.ProjectRuntime{}, ctx.Err()
	}
	return s.project, s.err
}
func (*monitoringSource) ListRuntimeProjectIDs(context.Context) ([]string, error) {
	return []string{"p"}, nil
}

type monitoringDriver struct {
	Driver
	state             domain.ProjectState
	inspects, metrics int
	err               error
	received          domain.ProjectRuntime
}

func (*monitoringDriver) Kind() domain.Kind { return domain.KindProcess }
func (d *monitoringDriver) Inspect(context.Context, domain.ProjectRuntime) (domain.Observation, error) {
	d.inspects++
	return domain.Observation{ProjectID: "p", State: d.state, Driver: domain.KindProcess}, nil
}
func (d *monitoringDriver) StreamMetrics(ctx context.Context, request domain.MetricRequest, sink domain.MetricSink) error {
	d.metrics++
	d.received = request.Project
	if d.err != nil {
		return d.err
	}
	return sink.WriteMetric(ctx, domain.MetricSample{ProjectID: "p", ServiceID: "web", CPUAvailable: true, MemoryAvailable: true})
}
func TestCombinedMonitoringResolvesOnceAndStillChecksFreshTrust(t *testing.T) {
	t.Parallel()
	source := &monitoringSource{project: domain.ProjectRuntime{ProjectID: "p", Root: "/repo", ManifestHash: "current", Kind: domain.KindProcess}}
	driver := &monitoringDriver{state: domain.StateRunning}
	service := NewService(source, driver)
	observation, samples, err := service.ObserveMetrics(context.Background(), "p")
	if err != nil || source.calls != 1 || driver.metrics != 1 || len(samples) != 1 || observation.State != domain.StateRunning || driver.received.ManifestHash != "current" {
		t.Fatalf("combined metrics: %#v %v calls=%d", samples, err, source.calls)
	}
	project, _, err := service.InspectProject(context.Background(), "p")
	if err != nil || source.calls != 2 || project.ManifestHash != "current" {
		t.Fatal("health project not current")
	}
	denied := errors.New("trust revoked")
	source.err = denied
	if _, _, err := service.InspectProject(context.Background(), "p"); !errors.Is(err, denied) {
		t.Fatal("cached status bypassed fresh trust")
	}
	if _, _, err := service.ObserveMetrics(context.Background(), "p"); !errors.Is(err, denied) || driver.metrics != 1 {
		t.Fatal("revoked project sampled")
	}
}
func TestCombinedMonitoringInactiveFailureAndCancellation(t *testing.T) {
	t.Parallel()
	source := &monitoringSource{project: domain.ProjectRuntime{ProjectID: "p", Kind: domain.KindProcess}}
	driver := &monitoringDriver{state: domain.StateStopped}
	service := NewService(source, driver)
	if _, samples, err := service.ObserveMetrics(context.Background(), "p"); err != nil || len(samples) != 0 || driver.metrics != 0 {
		t.Fatal("inactive driver sampled")
	}
	service.observations.invalidate("p")
	driver.state = domain.StateRunning
	failure := errors.New("stats unavailable")
	driver.err = failure
	if observation, _, err := service.ObserveMetrics(context.Background(), "p"); !errors.Is(err, failure) || observation.State != domain.StateRunning {
		t.Fatal("statistics failure hidden")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := service.ObserveMetrics(ctx, "p"); !errors.Is(err, context.Canceled) || driver.metrics != 1 {
		t.Fatal("canceled observation ran")
	}
}
