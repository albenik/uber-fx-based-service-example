package config

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("config",
		fx.Provide(LoadFromEnv, fx.Private),
		fx.Provide(splitConfig),
	)
}

func splitConfig(conf *Config) (*TelemetryConfig, *HTTPServerConfig) {
	return conf.Telemetry, conf.HTTPServer
}
