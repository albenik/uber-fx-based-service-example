package grpcprovider

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

func Provide(lc fx.Lifecycle, cfg *config.FeatureToggleConfig, logger *zap.Logger) (ports.FeatureToggleProvider, error) {
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
	return NewClient(conn, logger), nil
}
