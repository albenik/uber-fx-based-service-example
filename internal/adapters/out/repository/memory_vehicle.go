package repository

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

type MemoryVehicleRepository struct {
	mu       sync.RWMutex
	entities map[string]domain.Vehicle
}

func NewMemoryVehicleRepository() *MemoryVehicleRepository {
	return &MemoryVehicleRepository{
		entities: make(map[string]domain.Vehicle),
	}
}

func (r *MemoryVehicleRepository) Save(_ context.Context, entity *domain.Vehicle) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entities[entity.ID] = *entity
	return nil
}

func (r *MemoryVehicleRepository) FindByID(_ context.Context, id string) (*domain.Vehicle, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entity, ok := r.entities[id]
	if !ok || entity.DeletedAt != nil {
		return nil, domain.ErrNotFound
	}
	return &entity, nil
}

func (r *MemoryVehicleRepository) FindByFleetID(_ context.Context, fleetID string) ([]*domain.Vehicle, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Vehicle
	for _, e := range r.entities {
		if e.FleetID == fleetID && e.DeletedAt == nil {
			cp := e
			result = append(result, &cp)
		}
	}
	slices.SortFunc(result, func(a, b *domain.Vehicle) int {
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

func (r *MemoryVehicleRepository) SoftDelete(_ context.Context, id string) error {
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

func (r *MemoryVehicleRepository) Undelete(_ context.Context, id string) error {
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
