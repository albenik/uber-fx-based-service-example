package repository

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

type MemoryLegalEntityRepository struct {
	mu       sync.RWMutex
	entities map[string]domain.LegalEntity
}

func NewMemoryLegalEntityRepository() *MemoryLegalEntityRepository {
	return &MemoryLegalEntityRepository{
		entities: make(map[string]domain.LegalEntity),
	}
}

func (r *MemoryLegalEntityRepository) Save(_ context.Context, entity *domain.LegalEntity) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entities[entity.ID] = *entity
	return nil
}

func (r *MemoryLegalEntityRepository) FindByID(_ context.Context, id string) (*domain.LegalEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entity, ok := r.entities[id]
	if !ok || entity.DeletedAt != nil {
		return nil, domain.ErrNotFound
	}
	return &entity, nil
}

func (r *MemoryLegalEntityRepository) FindAll(_ context.Context) ([]*domain.LegalEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.LegalEntity
	for _, e := range r.entities {
		if e.DeletedAt == nil {
			cp := e
			result = append(result, &cp)
		}
	}
	slices.SortFunc(result, func(a, b *domain.LegalEntity) int {
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

func (r *MemoryLegalEntityRepository) SoftDelete(_ context.Context, id string) error {
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

func (r *MemoryLegalEntityRepository) Undelete(_ context.Context, id string) error {
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
