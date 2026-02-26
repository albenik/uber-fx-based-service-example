package config

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module("config",
		fx.Provide(LoadFromEnv, fx.Private),
		fx.Provide(splitConfig),
		fx.Invoke(func(c *Config, l *zap.Logger) error { return c.Validate(l) }),
	)
}

func splitConfig(conf *Config) (*TelemetryConfig, *DatabaseConfig, *HTTPServerConfig, *DriverLicenseGRPCConfig) {
	return conf.Telemetry, conf.Database, conf.HTTPServer, conf.DriverLicenseGRPC
}
