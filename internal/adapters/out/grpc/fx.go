package grpc

import (
	"go.uber.org/fx"

	"github.com/albenik/uber-fx-based-service-example/internal/adapters/out/grpc/driverlicense"
)

// Module provides gRPC output adapters for external services.
func Module() fx.Option {
	return fx.Module("grpc",
		driverlicense.Module(),
	)
}
