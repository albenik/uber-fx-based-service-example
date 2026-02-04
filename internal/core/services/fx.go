package services

import (
	"go.uber.org/fx"

	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services/userservice"
)

// Module provides core business logic services.
func Module() fx.Option {
	return fx.Module("services",
		fx.Provide(
			fx.Annotate(
				userservice.New,
				fx.As(new(ports.UserService)),
			),
		),
	)
}
