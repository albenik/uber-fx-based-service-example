package main

import (
	"go.uber.org/fx"

	httpAdapter "github.com/albenik/uber-fx-based-service-example/internal/adapters/in/http"
	"github.com/albenik/uber-fx-based-service-example/internal/adapters/out/repository"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services"
	"github.com/albenik/uber-fx-based-service-example/internal/telemetry"
)

func main() {
	fx.New(
		// Telemetry and monitoring
		telemetry.Module(),

		// Output adapters (driven/secondary)
		repository.Module(),

		// Core business logic
		services.Module(),

		// Input adapters (driving/primary)
		httpAdapter.Module(),
	).Run()
}
