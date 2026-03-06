package grpc

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

// Module provides gRPC output adapters for external services.
func Module() fx.Option {
	return fx.Module("grpc",
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
		return noopLicenseValidator{}, nil
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

	return NewDriverLicenseClient(conn, logger), nil
}

// ProvideFeatureToggleClient creates a FeatureToggleClient backed by a gRPC
// connection. It is intended to be called from the feature toggle backend
// selection logic rather than being auto-provided by the Module.
func ProvideFeatureToggleClient(lc fx.Lifecycle, cfg *config.FeatureToggleConfig, logger *zap.Logger) (ports.FeatureToggleProvider, error) {
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	if cfg.GRPCTLSEnabled {
		creds = grpc.WithTransportCredentials(credentials.NewTLS(nil))
	}
	conn, err := grpc.NewClient(cfg.GRPCAddr, creds)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing feature toggle gRPC connection")
			return conn.Close()
		},
	})

	logger.Info("Feature toggle provider: gRPC", zap.String("addr", cfg.GRPCAddr))
	return NewFeatureToggleClient(conn, logger), nil
}
