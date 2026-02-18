package telemetry_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/albenik/uber-fx-based-service-example/internal/config"
	"github.com/albenik/uber-fx-based-service-example/internal/telemetry"
)

func TestNewLogger_Production(t *testing.T) {
	logger, err := telemetry.NewLogger(&config.TelemetryConfig{Development: false})
	require.NoError(t, err)
	require.NotNil(t, logger)
}

func TestNewLogger_Development(t *testing.T) {
	logger, err := telemetry.NewLogger(&config.TelemetryConfig{Development: true})
	require.NoError(t, err)
	require.NotNil(t, logger)
}
