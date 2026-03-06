package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/config"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

// ProvideFeatureToggleProvider creates a FeatureToggleProvider backed by Redis.
// It is intended to be called from the feature toggle backend selection logic
// rather than being auto-provided by an FX module.
func ProvideFeatureToggleProvider(lc fx.Lifecycle, cfg *config.FeatureToggleConfig, logger *zap.Logger) (ports.FeatureToggleProvider, error) {
	opts, err := redis.ParseURL(cfg.RedisAddr)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := client.Ping(ctx).Err(); err != nil {
				logger.Error("Redis ping failed", zap.Error(err))
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing feature toggle Redis connection")
			return client.Close()
		},
	})

	logger.Info("Feature toggle provider: redis", zap.String("addr", cfg.RedisAddr))
	return NewFeatureToggleProvider(client, logger), nil
}
