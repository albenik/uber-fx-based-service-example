package main

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"

	grpcAdapter "github.com/albenik/uber-fx-based-service-example/internal/adapters/out/grpc"
	redisAdapter "github.com/albenik/uber-fx-based-service-example/internal/adapters/out/redis"
	"github.com/albenik/uber-fx-based-service-example/internal/config"
	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

// newFeatureToggleProvider selects the feature toggle backend based on
// FEATURE_TOGGLE_BACKEND ("grpc", "redis", or empty for a no-op provider
// that treats every toggle as disabled).
func newFeatureToggleProvider(lc fx.Lifecycle, cfg *config.FeatureToggleConfig, logger *zap.Logger) (ports.FeatureToggleProvider, error) {
	if cfg == nil || cfg.Backend == "" {
		logger.Info("FEATURE_TOGGLE_BACKEND not set, using no-op provider (all toggles disabled)")
		return noopFeatureToggleProvider{}, nil
	}

	switch cfg.Backend {
	case "grpc":
		if cfg.GRPCAddr == "" {
			return nil, fmt.Errorf("FEATURE_TOGGLE_GRPC_ADDR is required when backend is grpc")
		}
		return grpcAdapter.ProvideFeatureToggleClient(lc, cfg, logger)
	case "redis":
		if cfg.RedisAddr == "" {
			return nil, fmt.Errorf("FEATURE_TOGGLE_REDIS_ADDR is required when backend is redis")
		}
		return redisAdapter.ProvideFeatureToggleProvider(lc, cfg, logger)
	default:
		return nil, fmt.Errorf("unsupported FEATURE_TOGGLE_BACKEND: %q (must be grpc, redis, or empty)", cfg.Backend)
	}
}

// noopFeatureToggleProvider returns all toggles as disabled when no backend is configured.
type noopFeatureToggleProvider struct{}

func (noopFeatureToggleProvider) IsEnabled(context.Context, string) (bool, error) {
	return false, nil
}

func (noopFeatureToggleProvider) GetToggle(_ context.Context, _ string) (*domain.FeatureToggle, error) {
	return nil, domain.ErrToggleProviderUnavailable
}

func (noopFeatureToggleProvider) ListToggles(context.Context) ([]*domain.FeatureToggle, error) {
	return nil, domain.ErrToggleProviderUnavailable
}
