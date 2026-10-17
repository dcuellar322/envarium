package compose

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"envarium.dev/envarium/internal/runtime/domain"
	"golang.org/x/sync/singleflight"
)

const observationConfigTTL = 15 * time.Second
const maximumObservationConfigs = 128

type observedConfig struct {
	config normalizedConfig
	err    error
	until  time.Time
}

// observationConfigs coalesces in-process observation loading, including failures.
// Plans always normalize afresh. File/context edits reach reads within 15 seconds.
type observationConfigs struct {
	mu       sync.Mutex
	entries  map[string]observedConfig
	requests singleflight.Group
	now      func() time.Time
}

func newObservationConfigs() *observationConfigs {
	return &observationConfigs{entries: make(map[string]observedConfig), now: time.Now}
}

func (d *Driver) observationConfig(ctx context.Context, project domain.ProjectRuntime) (normalizedConfig, error) {
	if d.observed == nil {
		return d.config.loadObservation(ctx, project)
	}
	return d.observed.load(ctx, project, func() (normalizedConfig, error) { return d.config.loadObservation(ctx, project) })
}

func (c *observationConfigs) load(ctx context.Context, project domain.ProjectRuntime, normalize func() (normalizedConfig, error)) (normalizedConfig, error) {
	if err := ctx.Err(); err != nil {
		return normalizedConfig{}, err
	}
	document, err := json.Marshal(project)
	if err != nil {
		return normalizedConfig{}, err
	}
	digest := sha256.Sum256(document)
	key := hex.EncodeToString(digest[:])
	if entry, ok := c.cached(key); ok {
		return cloneConfig(entry.config), entry.err
	}
	result := c.requests.DoChan(key, func() (any, error) {
		if entry, ok := c.cached(key); ok {
			return entry, nil
		}
		config, err := normalize()
		entry := observedConfig{config: cloneConfig(config), err: err, until: c.now().Add(observationConfigTTL)}
		if ctx.Err() == nil {
			c.store(key, entry)
		}
		return entry, nil
	})
	select {
	case <-ctx.Done():
		return normalizedConfig{}, ctx.Err()
	case result := <-result:
		entry := result.Val.(observedConfig)
		return cloneConfig(entry.config), entry.err
	}
}

func (c *observationConfigs) cached(key string) (observedConfig, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if ok && !c.now().Before(entry.until) {
		delete(c.entries, key)
		ok = false
	}
	return entry, ok
}

func (c *observationConfigs) store(key string, entry observedConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= maximumObservationConfigs {
		oldestKey := ""
		var oldest time.Time
		for candidate, value := range c.entries {
			if oldestKey == "" || value.until.Before(oldest) {
				oldestKey, oldest = candidate, value.until
			}
		}
		delete(c.entries, oldestKey)
	}
	c.entries[key] = entry
}

func cloneConfig(config normalizedConfig) normalizedConfig {
	config.Services = append([]string(nil), config.Services...)
	config.BaseArgs = append([]string(nil), config.BaseArgs...)
	return config
}
