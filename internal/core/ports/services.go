package ports

//go:generate go tool mockgen -destination=mocks/mock_services.go -package=mocks . LegalEntityService,FleetService,VehicleService,DriverService,ContractService,VehicleAssignmentService

import (
	"context"
	"time"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// LegalEntityService is the input port for LegalEntity operations.
type LegalEntityService interface {
	Create(ctx context.Context, name, taxID string) (*domain.LegalEntity, error)
	Get(ctx context.Context, id string) (*domain.LegalEntity, error)
	List(ctx context.Context) ([]*domain.LegalEntity, error)
	Delete(ctx context.Context, id string) error
	Undelete(ctx context.Context, id string) error
}

// FleetService is the input port for Fleet operations.
type FleetService interface {
	Create(ctx context.Context, legalEntityID, name string) (*domain.Fleet, error)
	Get(ctx context.Context, id string) (*domain.Fleet, error)
	ListByLegalEntity(ctx context.Context, legalEntityID string) ([]*domain.Fleet, error)
	Delete(ctx context.Context, id string) error
	Undelete(ctx context.Context, id string) error
}

// VehicleService is the input port for Vehicle operations.
type VehicleService interface {
	Create(ctx context.Context, fleetID, make, model, licensePlate string, year int) (*domain.Vehicle, error)
	Get(ctx context.Context, id string) (*domain.Vehicle, error)
	ListByFleet(ctx context.Context, fleetID string) ([]*domain.Vehicle, error)
	Delete(ctx context.Context, id string) error
	Undelete(ctx context.Context, id string) error
}

// DriverService is the input port for Driver operations.
type DriverService interface {
	Create(ctx context.Context, firstName, lastName, licenseNumber string) (*domain.Driver, error)
	Get(ctx context.Context, id string) (*domain.Driver, error)
	List(ctx context.Context) ([]*domain.Driver, error)
	Delete(ctx context.Context, id string) error
	Undelete(ctx context.Context, id string) error
	ValidateLicense(ctx context.Context, id string) (domain.LicenseValidationResult, error)
}

// ContractService is the input port for Contract operations.
type ContractService interface {
	Create(ctx context.Context, driverID, legalEntityID, fleetID string, startDate, endDate time.Time) (*domain.Contract, error)
	Get(ctx context.Context, id string) (*domain.Contract, error)
	ListByDriver(ctx context.Context, driverID string) ([]*domain.Contract, error)
	Terminate(ctx context.Context, id, terminatedBy string) (*domain.Contract, error)
	Delete(ctx context.Context, id string) error
	Undelete(ctx context.Context, id string) error
}

// VehicleAssignmentService is the input port for VehicleAssignment operations.
type VehicleAssignmentService interface {
	Assign(ctx context.Context, contractID, vehicleID string) (*domain.VehicleAssignment, error)
	Get(ctx context.Context, id string) (*domain.VehicleAssignment, error)
	ListByContract(ctx context.Context, contractID string) ([]*domain.VehicleAssignment, error)
	Return(ctx context.Context, id string) (*domain.VehicleAssignment, error)
	Delete(ctx context.Context, id string) error
	Undelete(ctx context.Context, id string) error
}
