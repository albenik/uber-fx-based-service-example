package featuretoggle_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap/zaptest"

	"github.com/albenik/uber-fx-based-service-example/internal/config"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services/featuretoggle"
)

func TestModule_NoopProvider(t *testing.T) {
	var provider ports.FeatureToggleProvider

	app := fxtest.New(t,
		fx.Supply(zaptest.NewLogger(t)),
		fx.Supply(&config.FeatureToggleConfig{}),
		featuretoggle.Module(),
		fx.Populate(&provider),
	)
	app.RequireStart()
	defer app.RequireStop()

	require.NotNil(t, provider)
	enabled, err := provider.IsEnabled(t.Context(), "any-toggle")
	require.NoError(t, err)
	require.False(t, enabled)
}
