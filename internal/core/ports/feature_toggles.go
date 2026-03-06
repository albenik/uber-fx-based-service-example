package ports

//go:generate go tool mockgen -destination=mocks/mock_feature_toggles.go -package=mocks . FeatureToggleProvider

import (
	"context"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// FeatureToggleProvider is the output port for reading feature toggle state.
// Implementations can back onto different systems: a dedicated gRPC toggle
// service, Redis with project-local evaluation logic, etc.
type FeatureToggleProvider interface {
	IsEnabled(ctx context.Context, name string) (bool, error)
	GetToggle(ctx context.Context, name string) (*domain.FeatureToggle, error)
	ListToggles(ctx context.Context) ([]*domain.FeatureToggle, error)
}
