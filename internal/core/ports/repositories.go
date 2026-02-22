package ports

//go:generate mockgen -destination=mocks/mock_repositories.go -package=mocks . LegalEntityRepository,FleetRepository,VehicleRepository,DriverRepository,ContractRepository,VehicleAssignmentRepository

import (
	"context"
	"time"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// LegalEntityRepository is the output port for LegalEntity persistence.
type LegalEntityRepository interface {
	Save(ctx context.Context, entity *domain.LegalEntity) error
	FindByID(ctx context.Context, id string) (*domain.LegalEntity, error)
	FindAll(ctx context.Context) ([]*domain.LegalEntity, error)
	SoftDelete(ctx context.Context, id string) error
	Undelete(ctx context.Context, id string) error
}

// FleetRepository is the output port for Fleet persistence.
type FleetRepository interface {
	Save(ctx context.Context, entity *domain.Fleet) error
	FindByID(ctx context.Context, id string) (*domain.Fleet, error)
	FindByLegalEntityID(ctx context.Context, legalEntityID string) ([]*domain.Fleet, error)
	SoftDelete(ctx context.Context, id string) error
	Undelete(ctx context.Context, id string) error
}

// VehicleRepository is the output port for Vehicle persistence.
type VehicleRepository interface {
	Save(ctx context.Context, entity *domain.Vehicle) error
	FindByID(ctx context.Context, id string) (*domain.Vehicle, error)
	FindByFleetID(ctx context.Context, fleetID string) ([]*domain.Vehicle, error)
	SoftDelete(ctx context.Context, id string) error
	Undelete(ctx context.Context, id string) error
}

// DriverRepository is the output port for Driver persistence.
type DriverRepository interface {
	Save(ctx context.Context, entity *domain.Driver) error
	FindByID(ctx context.Context, id string) (*domain.Driver, error)
	FindAll(ctx context.Context) ([]*domain.Driver, error)
	SoftDelete(ctx context.Context, id string) error
	Undelete(ctx context.Context, id string) error
}

// ContractRepository is the output port for Contract persistence.
type ContractRepository interface {
	Save(ctx context.Context, entity *domain.Contract) error
	FindByID(ctx context.Context, id string) (*domain.Contract, error)
	FindByDriverID(ctx context.Context, driverID string) ([]*domain.Contract, error)
	FindOverlapping(ctx context.Context, driverID, legalEntityID, fleetID string, startDate, endDate time.Time, excludeID string) ([]*domain.Contract, error)
	SoftDelete(ctx context.Context, id string) error
	Undelete(ctx context.Context, id string) error
}

// VehicleAssignmentRepository is the output port for VehicleAssignment persistence.
type VehicleAssignmentRepository interface {
	Save(ctx context.Context, entity *domain.VehicleAssignment) error
	FindByID(ctx context.Context, id string) (*domain.VehicleAssignment, error)
	FindByContractID(ctx context.Context, contractID string) ([]*domain.VehicleAssignment, error)
	FindActiveByDriverID(ctx context.Context, driverID string) ([]*domain.VehicleAssignment, error)
	FindActiveByDriverIDAndFleetID(ctx context.Context, driverID, fleetID string) (*domain.VehicleAssignment, error)
	SoftDelete(ctx context.Context, id string) error
	Undelete(ctx context.Context, id string) error
}
