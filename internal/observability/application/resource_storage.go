package application

import (
	"context"
	"time"

	"envarium.dev/envarium/internal/observability/domain"
)

// runStorage keeps slow disk enrichment off the metric sampling deadline.
func (s *ResourceService) runStorage(ctx context.Context) {
	refresh := func() {
		projects, err := s.projects.ListResourceProjects(ctx)
		if err != nil {
			return
		}
		_, _ = s.inspectStorage(ctx, projects)
	}
	refresh()
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			refresh()
		}
	}
}

func (s *ResourceService) inspectStorage(ctx context.Context, projects []domain.ProjectDescriptor) (domain.StorageInventory, error) {
	inventory, err := s.storage.InspectStorage(ctx, projects)
	if ctx.Err() != nil {
		return inventory, ctx.Err()
	}
	if err != nil {
		inventory = domain.StorageInventory{Summary: domain.StorageSummary{Classification: domain.StorageUnknown}, Warnings: []string{"Docker storage is unavailable."}}
	}
	s.stateMu.Lock()
	s.storageState = inventory
	s.stateMu.Unlock()
	return inventory, err
}

func (s *ResourceService) storageSnapshot() domain.StorageInventory {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	return s.storageState
}

// CachedStorageSummary lets host presentation reuse disk enrichment without
// starting a second Docker filesystem traversal. Pending/unknown data is explicit.
func (s *ResourceService) CachedStorageSummary() (domain.StorageSummary, bool) {
	inventory := s.storageSnapshot()
	return inventory.Summary, inventory.Connected && !inventory.ObservedAt.IsZero() && inventory.Summary.Classification != "" && inventory.Summary.Classification != domain.StorageUnknown
}
