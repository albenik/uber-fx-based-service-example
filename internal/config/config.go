package config

import (
	"os"
	"strconv"
)

type Config struct {
	Telemetry       *TelemetryConfig
	Database        *DatabaseConfig
	HTTPServer      *HTTPServerConfig
	DriverLicenseGRPC *DriverLicenseGRPCConfig
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
		Database: &DatabaseConfig{
			MasterURL:  os.Getenv("DATABASE_MASTER_URL"),
			ReplicaURL: os.Getenv("DATABASE_REPLICA_URL"),
		},
		HTTPServer: &HTTPServerConfig{
			Addr: addr,
		},
		DriverLicenseGRPC: &DriverLicenseGRPCConfig{
			Addr: os.Getenv("DRIVER_LICENSE_GRPC_ADDR"),
		},
	}
}

func parseBool(s string) bool {
	v, _ := strconv.ParseBool(s)
	return v
}
