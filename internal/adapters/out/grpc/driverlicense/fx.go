package driverlicense

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/albenik/uber-fx-based-service-example/internal/config"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

// Module provides the driver license validation gRPC client.
// When DRIVER_LICENSE_GRPC_ADDR is empty, a no-op validator is used that returns
// an error on ValidateLicense calls.
func Module() fx.Option {
	return fx.Module("driverlicense",
		fx.Provide(newDriverLicenseValidator),
	)
}

func newDriverLicenseValidator(
	lc fx.Lifecycle,
	cfg *config.DriverLicenseGRPCConfig,
	logger *zap.Logger,
) (ports.DriverLicenseValidator, error) {
	if cfg == nil || cfg.Addr == "" {
		logger.Info("DRIVER_LICENSE_GRPC_ADDR not set, using no-op license validator")
		return noopValidator{}, nil
	}

	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	if cfg.TLSEnabled {
		creds = grpc.WithTransportCredentials(credentials.NewTLS(nil))
	}
	conn, err := grpc.NewClient(cfg.Addr, creds)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing driver license gRPC connection")
			return conn.Close()
		},
	})

	return NewClient(conn, logger), nil
}
