package application

import (
	"context"
	"errors"
)

// A scoped collector can share global read models for one scheduler cycle.
// Individual project trust, health, logs and action checks remain current.
type scheduledCollectionSource interface {
	BeginScheduledCollection(context.Context) (Collector, error)
}

func (s *Service) scheduledScope(ctx context.Context) (*Service, error) {
	source, ok := s.collector.(scheduledCollectionSource)
	if !ok {
		return s, nil
	}
	collector, err := source.BeginScheduledCollection(ctx)
	if err != nil {
		return nil, err
	}
	if collector == nil {
		return nil, errors.New("scheduled evidence scope is unavailable")
	}
	result := *s
	result.collector = collector
	return &result, nil
}
