package toggle

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

type Service struct {
	provider ports.FeatureToggleProvider
	logger   *zap.Logger
}

func New(provider ports.FeatureToggleProvider, logger *zap.Logger) *Service {
	return &Service{provider: provider, logger: logger}
}

func (s *Service) IsEnabled(ctx context.Context, name string) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("%w: toggle name is required", domain.ErrInvalidInput)
	}
	enabled, err := s.provider.IsEnabled(ctx, name)
	if err != nil {
		s.logger.Error("Failed to check toggle", zap.String("name", name), zap.Error(err))
		return false, err
	}
	return enabled, nil
}

func (s *Service) GetToggle(ctx context.Context, name string) (*domain.FeatureToggle, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: toggle name is required", domain.ErrInvalidInput)
	}
	toggle, err := s.provider.GetToggle(ctx, name)
	if err != nil {
		s.logger.Error("Failed to get toggle", zap.String("name", name), zap.Error(err))
		return nil, err
	}
	return toggle, nil
}

func (s *Service) ListToggles(ctx context.Context) ([]*domain.FeatureToggle, error) {
	toggles, err := s.provider.ListToggles(ctx)
	if err != nil {
		s.logger.Error("Failed to list toggles", zap.Error(err))
		return nil, err
	}
	return toggles, nil
}

var _ ports.FeatureToggleService = (*Service)(nil)
