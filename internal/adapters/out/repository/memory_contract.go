package repository

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

type MemoryContractRepository struct {
	mu       sync.RWMutex
	entities map[string]domain.Contract
}

func NewMemoryContractRepository() *MemoryContractRepository {
	return &MemoryContractRepository{
		entities: make(map[string]domain.Contract),
	}
}

func (r *MemoryContractRepository) Save(_ context.Context, entity *domain.Contract) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entities[entity.ID] = *entity
	return nil
}

func (r *MemoryContractRepository) FindByID(_ context.Context, id string) (*domain.Contract, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entity, ok := r.entities[id]
	if !ok || entity.DeletedAt != nil {
		return nil, domain.ErrNotFound
	}
	return &entity, nil
}

func (r *MemoryContractRepository) FindByDriverID(_ context.Context, driverID string) ([]*domain.Contract, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Contract
	for _, e := range r.entities {
		if e.DriverID == driverID && e.DeletedAt == nil {
			cp := e
			result = append(result, &cp)
		}
	}
	slices.SortFunc(result, func(a, b *domain.Contract) int {
		if a.StartDate.Before(b.StartDate) {
			return -1
		}
		if a.StartDate.After(b.StartDate) {
			return 1
		}
		return 0
	})
	return result, nil
}

func (r *MemoryContractRepository) FindOverlapping(_ context.Context, driverID, legalEntityID, fleetID string, startDate, endDate time.Time, excludeID string) ([]*domain.Contract, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Contract
	for _, e := range r.entities {
		if e.DeletedAt != nil || e.DriverID != driverID || e.LegalEntityID != legalEntityID || e.FleetID != fleetID || e.ID == excludeID {
			continue
		}
		effectiveEnd := e.EndDate
		if e.TerminatedAt != nil && e.TerminatedAt.Before(effectiveEnd) {
			effectiveEnd = *e.TerminatedAt
		}
		if startDate.Before(effectiveEnd) && endDate.After(e.StartDate) {
			cp := e
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (r *MemoryContractRepository) SoftDelete(_ context.Context, id string) error {
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

func (r *MemoryContractRepository) Undelete(_ context.Context, id string) error {
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
