package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

const featureToggleHashKey = "feature_toggles"

// featureToggleDTO is the JSON value stored per toggle in the Redis hash.
type featureToggleDTO struct {
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

// FeatureToggleProvider implements ports.FeatureToggleProvider with Redis as
// the storage backend and project-local evaluation logic. Toggles are kept in
// a single Redis hash ("feature_toggles") where each field is the toggle name
// and the value is a JSON object carrying the enabled flag and description.
//
// Unknown (missing) toggles are treated as disabled rather than as errors,
// which is the safest default for feature-flag systems.
type FeatureToggleProvider struct {
	client *redis.Client
	logger *zap.Logger
}

// NewFeatureToggleProvider creates a new Redis-backed feature toggle provider.
func NewFeatureToggleProvider(client *redis.Client, logger *zap.Logger) *FeatureToggleProvider {
	return &FeatureToggleProvider{client: client, logger: logger}
}

func (p *FeatureToggleProvider) IsEnabled(ctx context.Context, name string) (bool, error) {
	raw, err := p.client.HGet(ctx, featureToggleHashKey, name).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		p.logger.Error("Redis HGet failed", zap.String("name", name), zap.Error(err))
		return false, fmt.Errorf("%w: %v", domain.ErrToggleProviderUnavailable, err)
	}
	var dto featureToggleDTO
	if err := json.Unmarshal([]byte(raw), &dto); err != nil {
		p.logger.Error("Failed to decode toggle value", zap.String("name", name), zap.Error(err))
		return false, fmt.Errorf("%w: corrupt toggle data", domain.ErrToggleProviderUnavailable)
	}
	return dto.Enabled, nil
}

func (p *FeatureToggleProvider) GetToggle(ctx context.Context, name string) (*domain.FeatureToggle, error) {
	raw, err := p.client.HGet(ctx, featureToggleHashKey, name).Result()
	if err == redis.Nil {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		p.logger.Error("Redis HGet failed", zap.String("name", name), zap.Error(err))
		return nil, fmt.Errorf("%w: %v", domain.ErrToggleProviderUnavailable, err)
	}
	var dto featureToggleDTO
	if err := json.Unmarshal([]byte(raw), &dto); err != nil {
		p.logger.Error("Failed to decode toggle value", zap.String("name", name), zap.Error(err))
		return nil, fmt.Errorf("%w: corrupt toggle data", domain.ErrToggleProviderUnavailable)
	}
	return &domain.FeatureToggle{
		Name:        name,
		Enabled:     dto.Enabled,
		Description: dto.Description,
	}, nil
}

func (p *FeatureToggleProvider) ListToggles(ctx context.Context) ([]*domain.FeatureToggle, error) {
	all, err := p.client.HGetAll(ctx, featureToggleHashKey).Result()
	if err != nil {
		p.logger.Error("Redis HGetAll failed", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", domain.ErrToggleProviderUnavailable, err)
	}
	toggles := make([]*domain.FeatureToggle, 0, len(all))
	for name, raw := range all {
		var dto featureToggleDTO
		if err := json.Unmarshal([]byte(raw), &dto); err != nil {
			p.logger.Warn("Skipping corrupt toggle entry", zap.String("name", name), zap.Error(err))
			continue
		}
		toggles = append(toggles, &domain.FeatureToggle{
			Name:        name,
			Enabled:     dto.Enabled,
			Description: dto.Description,
		})
	}
	return toggles, nil
}

var _ ports.FeatureToggleProvider = (*FeatureToggleProvider)(nil)
