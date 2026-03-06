package featuretoggle

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/adapters/out/featuretoggle/grpcprovider"
	"github.com/albenik/uber-fx-based-service-example/internal/adapters/out/featuretoggle/redisprovider"
	"github.com/albenik/uber-fx-based-service-example/internal/config"
	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

// Module provides the feature toggle output adapter. The concrete backend is
// selected at startup via FEATURE_TOGGLE_BACKEND ("grpc", "redis", or empty
// for a no-op provider that treats every toggle as disabled).
func Module() fx.Option {
	return fx.Module("featuretoggle",
		fx.Provide(newProvider),
	)
}

func newProvider(lc fx.Lifecycle, cfg *config.FeatureToggleConfig, logger *zap.Logger) (ports.FeatureToggleProvider, error) {
	if cfg == nil || cfg.Backend == "" {
		logger.Info("FEATURE_TOGGLE_BACKEND not set, using no-op provider (all toggles disabled)")
		return noopProvider{}, nil
	}

	switch cfg.Backend {
	case "grpc":
		if cfg.GRPCAddr == "" {
			return nil, fmt.Errorf("FEATURE_TOGGLE_GRPC_ADDR is required when backend is grpc")
		}
		return grpcprovider.Provide(lc, cfg, logger)
	case "redis":
		if cfg.RedisAddr == "" {
			return nil, fmt.Errorf("FEATURE_TOGGLE_REDIS_ADDR is required when backend is redis")
		}
		return redisprovider.Provide(lc, cfg, logger)
	default:
		return nil, fmt.Errorf("unsupported FEATURE_TOGGLE_BACKEND: %q (must be grpc, redis, or empty)", cfg.Backend)
	}
}

// noopProvider returns all toggles as disabled when no backend is configured.
type noopProvider struct{}

func (noopProvider) IsEnabled(context.Context, string) (bool, error) {
	return false, nil
}

func (noopProvider) GetToggle(_ context.Context, _ string) (*domain.FeatureToggle, error) {
	return nil, domain.ErrToggleProviderUnavailable
}

func (noopProvider) ListToggles(context.Context) ([]*domain.FeatureToggle, error) {
	return nil, domain.ErrToggleProviderUnavailable
}
