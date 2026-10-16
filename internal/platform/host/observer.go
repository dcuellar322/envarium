// Package host observes local capacity and aggregate Docker storage.
package host

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/moby/moby/client"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"envarium.dev/envarium/internal/platform/dockerengine"
	"envarium.dev/envarium/internal/system/application"
)

type dockerUsage func(context.Context) (int64, int64, error)

var errDockerStorageUnavailable = errors.New("docker storage inspection is unavailable")

// Observer reads current OS capacity and best-effort Docker disk usage.
type Observer struct {
	cpu            func(context.Context, time.Duration, bool) ([]float64, error)
	memory         func(context.Context) (*mem.VirtualMemoryStat, error)
	docker         dockerUsage
	now            func() time.Time
	mu             sync.Mutex
	cached         *application.HostObservation
	until          time.Time
	dockerCached   *application.DockerObservation
	dockerWarnings []string
	dockerUntil    time.Time
}

// NewObserver creates the production host observer.
func NewObserver() *Observer {
	return &Observer{cpu: cpu.PercentWithContext, memory: mem.VirtualMemoryWithContext, docker: readDockerUsage, now: time.Now}
}

// Observe returns partial data with explicit warnings when an optional source is unavailable.
func (o *Observer) Observe(ctx context.Context) application.HostObservation {
	o.mu.Lock()
	defer o.mu.Unlock()
	now := o.now().UTC()
	if o.cached != nil && now.Before(o.until) {
		return cloneObservation(*o.cached)
	}
	result := application.HostObservation{
		Docker: application.DockerObservation{Attribution: "unknown"}, ObservedAt: now, Warnings: []string{},
	}
	percentages, err := o.cpu(ctx, 0, false)
	if err != nil || len(percentages) == 0 {
		result.Warnings = append(result.Warnings, "Host CPU usage is unavailable.")
	} else {
		result.CPUPercent = percentages[0]
	}
	memory, err := o.memory(ctx)
	if err != nil {
		result.Warnings = append(result.Warnings, "Host memory usage is unavailable.")
	} else {
		result.MemoryUsedBytes, result.MemoryTotalBytes = memory.Used, memory.Total
	}
	if o.dockerCached == nil || !now.Before(o.dockerUntil) {
		docker := application.DockerObservation{Attribution: "unknown"}
		o.dockerWarnings = nil
		dockerContext, cancel := context.WithTimeout(ctx, 3*time.Second)
		storage, reclaimable, usageErr := o.docker(dockerContext)
		cancel()
		if usageErr != nil {
			docker.Connected = errors.Is(usageErr, errDockerStorageUnavailable)
			o.dockerWarnings = []string{"Docker storage is unavailable."}
		} else {
			docker.Connected = true
			docker.StorageBytes, docker.ReclaimableBytes = &storage, &reclaimable
			docker.Attribution = "shared"
		}
		o.dockerCached, o.dockerUntil = &docker, o.now().UTC().Add(5*time.Minute)
	}
	result.Docker = *o.dockerCached
	result.Warnings = append(result.Warnings, o.dockerWarnings...)
	o.cached = &result
	o.until = now.Add(5 * time.Second)
	return cloneObservation(result)
}

func cloneObservation(value application.HostObservation) application.HostObservation {
	value.Warnings = append([]string(nil), value.Warnings...)
	return value
}

func readDockerUsage(ctx context.Context) (int64, int64, error) {
	engine, err := dockerengine.NewClient(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = engine.Close() }()
	if _, err := engine.Ping(ctx, client.PingOptions{NegotiateAPIVersion: true}); err != nil {
		return 0, 0, err
	}
	usage, err := engine.DiskUsage(ctx, client.DiskUsageOptions{Containers: true, Images: true, Volumes: true, BuildCache: true})
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %v", errDockerStorageUnavailable, err)
	}
	storage := usage.Containers.TotalSize + usage.Images.TotalSize + usage.Volumes.TotalSize + usage.BuildCache.TotalSize
	reclaimable := usage.Containers.Reclaimable + usage.Images.Reclaimable + usage.Volumes.Reclaimable + usage.BuildCache.Reclaimable
	return storage, reclaimable, nil
}
