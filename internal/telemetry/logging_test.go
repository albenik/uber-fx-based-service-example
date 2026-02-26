package telemetry_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/albenik/uber-fx-based-service-example/internal/config"
	"github.com/albenik/uber-fx-based-service-example/internal/telemetry"
)

func TestNewLogger_Production(t *testing.T) {
	t.Setenv("DEVLOG", "")

	logger, level, err := telemetry.NewLogger()
	require.NoError(t, err)
	require.NotNil(t, logger)
	require.Equal(t, zapcore.DebugLevel, level.Level())
}

func TestNewLogger_Development(t *testing.T) {
	t.Setenv("DEVLOG", "true")

	logger, level, err := telemetry.NewLogger()
	require.NoError(t, err)
	require.NotNil(t, logger)
	require.Equal(t, zapcore.DebugLevel, level.Level())
}

func TestReconfigureLogLevel(t *testing.T) {
	level := zap.NewAtomicLevelAt(zap.DebugLevel)
	cfg := &config.TelemetryConfig{LogLevel: "warn"}

	err := telemetry.ReconfigureLogLevel(level, cfg)
	require.NoError(t, err)
	require.Equal(t, zapcore.WarnLevel, level.Level())
}

func TestReconfigureLogLevel_Invalid(t *testing.T) {
	level := zap.NewAtomicLevelAt(zap.DebugLevel)
	cfg := &config.TelemetryConfig{LogLevel: "invalid"}

	err := telemetry.ReconfigureLogLevel(level, cfg)
	require.Error(t, err)
	require.Equal(t, zapcore.DebugLevel, level.Level()) // unchanged
}
