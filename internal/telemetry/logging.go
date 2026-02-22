package telemetry

import (
	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/config"
)

func NewLogger(cfg *config.TelemetryConfig) (*zap.Logger, error) {
	if cfg.Development {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
