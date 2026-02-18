package repository

import (
	"context"
	"slices"
	"sync"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// MemoryFooEntityRepository is an in-memory implementation of FooEntityRepository.
type MemoryFooEntityRepository struct {
	mu       sync.RWMutex
	entities map[string]domain.FooEntity
}

func NewMemoryFooEntityRepository() *MemoryFooEntityRepository {
	return &MemoryFooEntityRepository{
		entities: make(map[string]domain.FooEntity),
	}
}

func (r *MemoryFooEntityRepository) Save(_ context.Context, entity *domain.FooEntity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.entities[entity.ID] = *entity
	return nil
}

func (r *MemoryFooEntityRepository) FindByID(_ context.Context, id string) (*domain.FooEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entity, ok := r.entities[id]
	if !ok {
		return nil, domain.ErrEntityNotFound
	}
	return &entity, nil
}

func (r *MemoryFooEntityRepository) FindAll(_ context.Context) ([]*domain.FooEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// NOTE: This returns pointers to copies, which is safe as long as
	// domain.FooEntity contains only value-type fields (no pointers/slices/maps).
	entities := make([]*domain.FooEntity, 0, len(r.entities))
	for _, e := range r.entities {
		entities = append(entities, &e)
	}
	slices.SortFunc(entities, func(a, b *domain.FooEntity) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	return entities, nil
}

func (r *MemoryFooEntityRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.entities[id]; !ok {
		return domain.ErrEntityNotFound
	}
	delete(r.entities, id)
	return nil
}
