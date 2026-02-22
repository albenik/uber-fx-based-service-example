package services

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"

	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services/assignment"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services/contract"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services/driver"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services/fleet"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services/legalentity"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services/vehicle"
)

// Module provides core business logic services.
func Module() fx.Option {
	return fx.Module("services",
		fx.Provide(
			fx.Private,
			func() legalentity.IDGenerator { return uuid.NewString },
			func() fleet.IDGenerator { return uuid.NewString },
			func() vehicle.IDGenerator { return uuid.NewString },
			func() driver.IDGenerator { return uuid.NewString },
			func() contract.IDGenerator { return uuid.NewString },
			func() assignment.IDGenerator { return uuid.NewString },
			func() driver.Clock { return time.Now },
			func() contract.Clock { return time.Now },
			func() assignment.Clock { return time.Now },
		),
		fx.Provide(
			fx.Annotate(
				legalentity.New,
				fx.As(new(ports.LegalEntityService)),
			),
			fx.Annotate(
				fleet.New,
				fx.As(new(ports.FleetService)),
			),
			fx.Annotate(
				vehicle.New,
				fx.As(new(ports.VehicleService)),
			),
			fx.Annotate(
				driver.New,
				fx.As(new(ports.DriverService)),
			),
			fx.Annotate(
				contract.New,
				fx.As(new(ports.ContractService)),
			),
			fx.Annotate(
				assignment.New,
				fx.As(new(ports.VehicleAssignmentService)),
			),
		),
	)
}
