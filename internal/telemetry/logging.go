package telemetry

import (
	"os"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/albenik/uber-fx-based-service-example/internal/config"
)

func NewLogger() (*zap.Logger, zap.AtomicLevel, error) {
	level := zap.NewAtomicLevelAt(zap.DebugLevel)

	devlogEnv := os.Getenv("DEVLOG")
	if devlogEnv == "" {
		devlogEnv = "false"
	}

	dev, err := strconv.ParseBool(devlogEnv)
	if err != nil {
		return nil, zap.AtomicLevel{}, err
	}

	var cfg zap.Config
	if dev {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}
	cfg.Level = level

	logger, err := cfg.Build()
	if err != nil {
		return nil, zap.AtomicLevel{}, err
	}
	return logger, level, nil
}

// ReconfigureLogLevel sets the logger level from config. Called after config validation.
func ReconfigureLogLevel(level zap.AtomicLevel, cfg *config.TelemetryConfig) error {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		return err
	}
	level.SetLevel(lvl)
	return nil
}
