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
				NewMemoryFooEntityRepository,
				fx.As(new(ports.FooEntityRepository)),
			),
		),
	)
}
