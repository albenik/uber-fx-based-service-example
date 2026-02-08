package repository

import (
	"context"
	"sync"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// MemoryFooEntityRepository is an in-memory implementation of FooEntityRepository.
type MemoryFooEntityRepository struct {
	mu         sync.RWMutex
	FooEntitys map[string]*domain.FooEntity
}

func NewMemoryFooEntityRepository() *MemoryFooEntityRepository {
	return &MemoryFooEntityRepository{
		FooEntitys: make(map[string]*domain.FooEntity),
	}
}

func (r *MemoryFooEntityRepository) Save(_ context.Context, FooEntity *domain.FooEntity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.FooEntitys[FooEntity.ID] = FooEntity
	return nil
}

func (r *MemoryFooEntityRepository) FindByID(_ context.Context, id string) (*domain.FooEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	FooEntity, ok := r.FooEntitys[id]
	if !ok {
		return nil, domain.ErrEntityNotFound
	}
	return FooEntity, nil
}

func (r *MemoryFooEntityRepository) FindAll(_ context.Context) ([]*domain.FooEntity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	FooEntitys := make([]*domain.FooEntity, 0, len(r.FooEntitys))
	for _, u := range r.FooEntitys {
		FooEntitys = append(FooEntitys, u)
	}
	return FooEntitys, nil
}

func (r *MemoryFooEntityRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.FooEntitys[id]; !ok {
		return domain.ErrEntityNotFound
	}
	delete(r.FooEntitys, id)
	return nil
}
