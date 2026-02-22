package repository

import (
	"go.uber.org/fx"

	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

// Module provides output adapters (driven adapters).
func Module() fx.Option {
	return fx.Module("repository",
		fx.Provide(
			fx.Annotate(
				NewMemoryLegalEntityRepository,
				fx.As(new(ports.LegalEntityRepository)),
			),
			fx.Annotate(
				NewMemoryFleetRepository,
				fx.As(new(ports.FleetRepository)),
			),
			fx.Annotate(
				NewMemoryVehicleRepository,
				fx.As(new(ports.VehicleRepository)),
			),
			fx.Annotate(
				NewMemoryDriverRepository,
				fx.As(new(ports.DriverRepository)),
			),
			fx.Annotate(
				NewMemoryContractRepository,
				fx.As(new(ports.ContractRepository)),
			),
			fx.Annotate(
				NewMemoryVehicleAssignmentRepository,
				fx.As(new(ports.VehicleAssignmentRepository)),
			),
		),
	)
}
