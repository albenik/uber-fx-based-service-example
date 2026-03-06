package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
	featuretogglev1 "github.com/albenik/uber-fx-based-service-example/internal/gen/featuretoggle/v1"
)

// FeatureToggleClient implements ports.FeatureToggleProvider by calling an external gRPC
// feature-toggle service. All evaluation logic lives on the remote side.
type FeatureToggleClient struct {
	grpcClient featuretogglev1.FeatureToggleServiceClient
	logger     *zap.Logger
}

// NewFeatureToggleClient creates a new feature toggle gRPC client.
func NewFeatureToggleClient(conn grpc.ClientConnInterface, logger *zap.Logger) *FeatureToggleClient {
	return &FeatureToggleClient{
		grpcClient: featuretogglev1.NewFeatureToggleServiceClient(conn),
		logger:     logger,
	}
}

func (c *FeatureToggleClient) IsEnabled(ctx context.Context, name string) (bool, error) {
	resp, err := c.grpcClient.IsEnabled(ctx, &featuretogglev1.IsEnabledRequest{Name: name})
	if err != nil {
		c.logger.Error("gRPC feature toggle IsEnabled failed", zap.String("name", name), zap.Error(err))
		return false, domain.ErrToggleProviderUnavailable
	}
	return resp.Enabled, nil
}

func (c *FeatureToggleClient) GetToggle(ctx context.Context, name string) (*domain.FeatureToggle, error) {
	resp, err := c.grpcClient.GetToggle(ctx, &featuretogglev1.GetToggleRequest{Name: name})
	if err != nil {
		c.logger.Error("gRPC feature toggle GetToggle failed", zap.String("name", name), zap.Error(err))
		return nil, domain.ErrToggleProviderUnavailable
	}
	if resp.Toggle == nil {
		return nil, domain.ErrNotFound
	}
	return featureToggleProtoToDomain(resp.Toggle), nil
}

func (c *FeatureToggleClient) ListToggles(ctx context.Context) ([]*domain.FeatureToggle, error) {
	resp, err := c.grpcClient.ListToggles(ctx, &featuretogglev1.ListTogglesRequest{})
	if err != nil {
		c.logger.Error("gRPC feature toggle ListToggles failed", zap.Error(err))
		return nil, domain.ErrToggleProviderUnavailable
	}
	out := make([]*domain.FeatureToggle, 0, len(resp.Toggles))
	for _, t := range resp.Toggles {
		out = append(out, featureToggleProtoToDomain(t))
	}
	return out, nil
}

var _ ports.FeatureToggleProvider = (*FeatureToggleClient)(nil)

func featureToggleProtoToDomain(t *featuretogglev1.FeatureToggle) *domain.FeatureToggle {
	return &domain.FeatureToggle{
		Name:        t.Name,
		Enabled:     t.Enabled,
		Description: t.Description,
	}
}
