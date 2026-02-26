package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Telemetry         *TelemetryConfig
	Database          *DatabaseConfig
	HTTPServer        *HTTPServerConfig
	DriverLicenseGRPC *DriverLicenseGRPCConfig
}

func LoadFromEnv() (*Config, error) {
	tlsEnabledEnv := getEnv("DRIVER_LICENSE_GRPC_TLS", "false")
	tlsEnabled, err := strconv.ParseBool(tlsEnabledEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DRIVER_LICENSE_GRPC_TLS: %w", err)
	}

	cfg := &Config{
		Telemetry: &TelemetryConfig{
			LogLevel: getEnv("LOG_LEVEL", "debug"),
		},
		Database: &DatabaseConfig{
			MasterURL:  getEnv("DATABASE_MASTER_URL", ""),
			ReplicaURL: getEnv("DATABASE_REPLICA_URL", ""),
		},
		HTTPServer: &HTTPServerConfig{
			Addr: getEnv("HTTP_ADDR", ":8080"),
		},
		DriverLicenseGRPC: &DriverLicenseGRPCConfig{
			Addr:       getEnv("DRIVER_LICENSE_GRPC_ADDR", ""),
			TLSEnabled: tlsEnabled,
		},
	}

	return cfg, nil
}

// Validate checks config consistency and required fields. Logs each error and returns
// an aggregated error if any validations failed (app will stop).
func (c *Config) Validate(logger *zap.Logger) error {
	var errs []error

	// Validate LOG_LEVEL
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(c.Telemetry.LogLevel)); err != nil {
		logger.Error("invalid LOG_LEVEL", zap.String("value", c.Telemetry.LogLevel), zap.Error(err))
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func getEnv(env, defaultValue string) string {
	envValue := os.Getenv(env)
	if envValue == "" {
		return defaultValue
	}

	return envValue
}
