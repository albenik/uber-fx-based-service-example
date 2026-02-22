package repository

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

type MemoryVehicleAssignmentRepository struct {
	mu            sync.RWMutex
	entities      map[string]domain.VehicleAssignment
	vehicleRepo   ports.VehicleRepository
}

func NewMemoryVehicleAssignmentRepository(vehicleRepo ports.VehicleRepository) *MemoryVehicleAssignmentRepository {
	return &MemoryVehicleAssignmentRepository{
		entities:    make(map[string]domain.VehicleAssignment),
		vehicleRepo: vehicleRepo,
	}
}

func (r *MemoryVehicleAssignmentRepository) Save(_ context.Context, entity *domain.VehicleAssignment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entities[entity.ID] = *entity
	return nil
}

func (r *MemoryVehicleAssignmentRepository) FindByID(_ context.Context, id string) (*domain.VehicleAssignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entity, ok := r.entities[id]
	if !ok || entity.DeletedAt != nil {
		return nil, domain.ErrNotFound
	}
	return &entity, nil
}

func (r *MemoryVehicleAssignmentRepository) FindActiveByDriverID(_ context.Context, driverID string) ([]*domain.VehicleAssignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.VehicleAssignment
	for _, e := range r.entities {
		if e.DriverID == driverID && e.DeletedAt == nil && e.EndTime == nil {
			cp := e
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (r *MemoryVehicleAssignmentRepository) FindByContractID(_ context.Context, contractID string) ([]*domain.VehicleAssignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.VehicleAssignment
	for _, e := range r.entities {
		if e.ContractID == contractID && e.DeletedAt == nil {
			cp := e
			result = append(result, &cp)
		}
	}
	slices.SortFunc(result, func(a, b *domain.VehicleAssignment) int {
		if a.StartTime.Before(b.StartTime) {
			return -1
		}
		if a.StartTime.After(b.StartTime) {
			return 1
		}
		return 0
	})
	return result, nil
}

func (r *MemoryVehicleAssignmentRepository) FindActiveByDriverIDAndFleetID(ctx context.Context, driverID, fleetID string) (*domain.VehicleAssignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.entities {
		if e.DriverID != driverID || e.DeletedAt != nil || e.EndTime != nil {
			continue
		}
		vehicle, err := r.vehicleRepo.FindByID(ctx, e.VehicleID)
		if err != nil || vehicle == nil || vehicle.FleetID != fleetID {
			continue
		}
		cp := e
		return &cp, nil
	}
	return nil, nil
}

func (r *MemoryVehicleAssignmentRepository) SoftDelete(_ context.Context, id string) error {
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

func (r *MemoryVehicleAssignmentRepository) Undelete(_ context.Context, id string) error {
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
