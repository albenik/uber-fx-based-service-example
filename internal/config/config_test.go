package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/albenik/uber-fx-based-service-example/internal/config"
)

func TestLoadFromEnv_Defaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_DEV", "")

	cfg := config.LoadFromEnv()
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.Telemetry)
	require.NotNil(t, cfg.HTTPServer)
	assert.Equal(t, ":8080", cfg.HTTPServer.Addr)
	assert.False(t, cfg.Telemetry.Development)
}

func TestLoadFromEnv_CustomAddr(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("LOG_DEV", "")

	cfg := config.LoadFromEnv()
	assert.Equal(t, ":9090", cfg.HTTPServer.Addr)
}

func TestLoadFromEnv_DevLogging(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_DEV", "true")

	cfg := config.LoadFromEnv()
	assert.True(t, cfg.Telemetry.Development)
}

func TestLoadFromEnv_DevLogging_One(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_DEV", "1")

	cfg := config.LoadFromEnv()
	assert.True(t, cfg.Telemetry.Development)
}

func TestLoadFromEnv_DevLogging_False(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_DEV", "false")

	cfg := config.LoadFromEnv()
	assert.False(t, cfg.Telemetry.Development)
}

func TestLoadFromEnv_DevLogging_Zero(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_DEV", "0")

	cfg := config.LoadFromEnv()
	assert.False(t, cfg.Telemetry.Development)
}

func TestLoadFromEnv_DevLogging_InvalidValue(t *testing.T) {
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_DEV", "invalid")

	cfg := config.LoadFromEnv()
	assert.False(t, cfg.Telemetry.Development)
}
