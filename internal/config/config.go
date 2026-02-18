package config

import (
	"os"
	"strconv"
)

type Config struct {
	Telemetry  *TelemetryConfig
	HTTPServer *HTTPServerConfig
}

func LoadFromEnv() *Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return &Config{
		Telemetry: &TelemetryConfig{
			Development: parseBool(os.Getenv("LOG_DEV")),
		},
		HTTPServer: &HTTPServerConfig{
			Addr: addr,
		},
	}
}

func parseBool(s string) bool {
	v, _ := strconv.ParseBool(s)
	return v
}
