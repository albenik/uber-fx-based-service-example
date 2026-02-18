package services

import (
	"github.com/google/uuid"
	"go.uber.org/fx"

	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services/fooservice"
)

// Module provides core business logic services.
func Module() fx.Option {
	return fx.Module("services",
		fx.Provide(
			fx.Private,
			func() fooservice.IDGenerator { return uuid.NewString },
		),
		fx.Provide(
			fx.Annotate(
				fooservice.New,
				fx.As(new(ports.FooEntityService)),
			),
		),
	)
}
