package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/config"
)

func TestLoadFromEnv_Defaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_LEVEL", "")

	cfg, err := config.LoadFromEnv()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	if assert.NotNil(t, cfg.Telemetry) {
		assert.Equal(t, "debug", cfg.Telemetry.LogLevel)
	}

	if assert.NotNil(t, cfg.HTTPServer) {
		assert.Equal(t, ":8080", cfg.HTTPServer.Addr)
	}
}

func TestLoadFromEnv_CustomAddr(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("LOG_LEVEL", "")

	cfg, err := config.LoadFromEnv()
	require.NoError(t, err)
	assert.Equal(t, ":9090", cfg.HTTPServer.Addr)
}

func TestLoadFromEnv_LogLevel(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_LEVEL", "info")

	cfg, err := config.LoadFromEnv()
	require.NoError(t, err)
	assert.Equal(t, "info", cfg.Telemetry.LogLevel)
}

func TestLoadFromEnv_LogLevel_DefaultWhenEmpty(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_LEVEL", "")

	cfg, err := config.LoadFromEnv()
	require.NoError(t, err)
	assert.Equal(t, "debug", cfg.Telemetry.LogLevel)
}

func TestLoadFromEnv_LogLevel_Warn(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_LEVEL", "warn")

	cfg, err := config.LoadFromEnv()
	require.NoError(t, err)
	assert.Equal(t, "warn", cfg.Telemetry.LogLevel)
}

func TestLoadFromEnv_LogLevel_Error(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_LEVEL", "error")

	cfg, err := config.LoadFromEnv()
	require.NoError(t, err)
	assert.Equal(t, "error", cfg.Telemetry.LogLevel)
}

func TestConfig_Validate_Success(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		Telemetry: &config.TelemetryConfig{LogLevel: "info"},
		Database:  &config.DatabaseConfig{MasterURL: "postgres://localhost/test"},
	}

	err := cfg.Validate(logger)
	assert.NoError(t, err)
}

func TestConfig_Validate_InvalidLogLevel(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		Telemetry: &config.TelemetryConfig{LogLevel: "invalid"},
		Database:  &config.DatabaseConfig{MasterURL: "postgres://localhost/test"},
	}

	err := cfg.Validate(logger)
	require.Error(t, err)
	assert.ErrorContains(t, err, "unrecognized level")
}
