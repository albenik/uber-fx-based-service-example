package toggle_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports/mocks"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services/toggle"
)

func TestService_IsEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := mocks.NewMockFeatureToggleProvider(ctrl)

	provider.EXPECT().IsEnabled(gomock.Any(), "dark-mode").Return(true, nil)

	svc := toggle.New(provider, zaptest.NewLogger(t))
	enabled, err := svc.IsEnabled(t.Context(), "dark-mode")
	require.NoError(t, err)
	assert.True(t, enabled)
}

func TestService_IsEnabled_EmptyName(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := mocks.NewMockFeatureToggleProvider(ctrl)

	svc := toggle.New(provider, zaptest.NewLogger(t))
	_, err := svc.IsEnabled(t.Context(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestService_IsEnabled_ProviderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := mocks.NewMockFeatureToggleProvider(ctrl)

	provider.EXPECT().IsEnabled(gomock.Any(), "dark-mode").Return(false, domain.ErrToggleProviderUnavailable)

	svc := toggle.New(provider, zaptest.NewLogger(t))
	_, err := svc.IsEnabled(t.Context(), "dark-mode")
	assert.ErrorIs(t, err, domain.ErrToggleProviderUnavailable)
}

func TestService_GetToggle(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := mocks.NewMockFeatureToggleProvider(ctrl)

	expected := &domain.FeatureToggle{Name: "dark-mode", Enabled: true, Description: "Enable dark mode"}
	provider.EXPECT().GetToggle(gomock.Any(), "dark-mode").Return(expected, nil)

	svc := toggle.New(provider, zaptest.NewLogger(t))
	result, err := svc.GetToggle(t.Context(), "dark-mode")
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestService_GetToggle_EmptyName(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := mocks.NewMockFeatureToggleProvider(ctrl)

	svc := toggle.New(provider, zaptest.NewLogger(t))
	_, err := svc.GetToggle(t.Context(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestService_GetToggle_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := mocks.NewMockFeatureToggleProvider(ctrl)

	provider.EXPECT().GetToggle(gomock.Any(), "nonexistent").Return(nil, domain.ErrNotFound)

	svc := toggle.New(provider, zaptest.NewLogger(t))
	_, err := svc.GetToggle(t.Context(), "nonexistent")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestService_ListToggles(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := mocks.NewMockFeatureToggleProvider(ctrl)

	expected := []*domain.FeatureToggle{
		{Name: "dark-mode", Enabled: true, Description: "Enable dark mode"},
		{Name: "beta-feature", Enabled: false, Description: "Beta feature"},
	}
	provider.EXPECT().ListToggles(gomock.Any()).Return(expected, nil)

	svc := toggle.New(provider, zaptest.NewLogger(t))
	result, err := svc.ListToggles(t.Context())
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestService_ListToggles_ProviderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	provider := mocks.NewMockFeatureToggleProvider(ctrl)

	provider.EXPECT().ListToggles(gomock.Any()).Return(nil, domain.ErrToggleProviderUnavailable)

	svc := toggle.New(provider, zaptest.NewLogger(t))
	_, err := svc.ListToggles(t.Context())
	assert.ErrorIs(t, err, domain.ErrToggleProviderUnavailable)
}
