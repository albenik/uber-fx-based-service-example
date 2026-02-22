package main

import (
	"go.uber.org/fx"

	httpAdapter "github.com/albenik/uber-fx-based-service-example/internal/adapters/in/http"
	"github.com/albenik/uber-fx-based-service-example/internal/adapters/out/postgres"
	"github.com/albenik/uber-fx-based-service-example/internal/config"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services"
	"github.com/albenik/uber-fx-based-service-example/internal/telemetry"
)

func main() {
	fx.New(AppModules()...).Run()
}

func AppModules() []fx.Option {
	return []fx.Option{
		// Telemetry and monitoring
		telemetry.Module(),

		// Configuration
		config.Module(),

		// Output adapters (driven/secondary)
		postgres.Module(),

		// Core business logic
		services.Module(),

		// Input adapters (driving/primary)
		httpAdapter.Module(),
	}
}
