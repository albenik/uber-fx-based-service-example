package repository

import (
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
	"go.uber.org/fx"
)

// Module provides output adapters (driven adapters).
func Module() fx.Option {
	return fx.Module("repository",
		fx.Provide(
			fx.Annotate(
				NewMemoryUserRepository,
				fx.As(new(ports.UserRepository)),
			),
		),
	)
}
