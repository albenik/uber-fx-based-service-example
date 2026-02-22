package repository

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

type MemoryFleetRepository struct {
	mu       sync.RWMutex
	entities map[string]domain.Fleet
}

func NewMemoryFleetRepository() *MemoryFleetRepository {
	return &MemoryFleetRepository{
		entities: make(map[string]domain.Fleet),
	}
}

func (r *MemoryFleetRepository) Save(_ context.Context, entity *domain.Fleet) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entities[entity.ID] = *entity
	return nil
}

func (r *MemoryFleetRepository) FindByID(_ context.Context, id string) (*domain.Fleet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entity, ok := r.entities[id]
	if !ok || entity.DeletedAt != nil {
		return nil, domain.ErrNotFound
	}
	return &entity, nil
}

func (r *MemoryFleetRepository) FindByLegalEntityID(_ context.Context, legalEntityID string) ([]*domain.Fleet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Fleet
	for _, e := range r.entities {
		if e.LegalEntityID == legalEntityID && e.DeletedAt == nil {
			cp := e
			result = append(result, &cp)
		}
	}
	slices.SortFunc(result, func(a, b *domain.Fleet) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	return result, nil
}

func (r *MemoryFleetRepository) SoftDelete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	entity, ok := r.entities[id]
	if !ok {
		return domain.ErrNotFound
	}
	if entity.DeletedAt != nil {
		return domain.ErrAlreadyDeleted
	}
	now := time.Now()
	entity.DeletedAt = &now
	r.entities[id] = entity
	return nil
}

func (r *MemoryFleetRepository) Undelete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	entity, ok := r.entities[id]
	if !ok {
		return domain.ErrNotFound
	}
	if entity.DeletedAt == nil {
		return nil
	}
	entity.DeletedAt = nil
	r.entities[id] = entity
	return nil
}
