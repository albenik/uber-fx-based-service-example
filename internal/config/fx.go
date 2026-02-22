package config

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("config",
		fx.Provide(LoadFromEnv, fx.Private),
		fx.Provide(splitConfig),
	)
}

func splitConfig(conf *Config) (*TelemetryConfig, *DatabaseConfig, *HTTPServerConfig, *DriverLicenseGRPCConfig) {
	return conf.Telemetry, conf.Database, conf.HTTPServer, conf.DriverLicenseGRPC
}
