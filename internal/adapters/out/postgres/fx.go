package postgres

import (
	"context"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/config"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

// Module provides PostgreSQL-backed output adapters with master/replica read splitting.
func Module() fx.Option {
	return fx.Module("postgres",
		fx.Provide(
			newDB,
			fx.Annotate(
				NewLegalEntityRepository,
				fx.As(new(ports.LegalEntityRepository)),
			),
			fx.Annotate(
				NewFleetRepository,
				fx.As(new(ports.FleetRepository)),
			),
			fx.Annotate(
				NewVehicleRepository,
				fx.As(new(ports.VehicleRepository)),
			),
			fx.Annotate(
				NewDriverRepository,
				fx.As(new(ports.DriverRepository)),
			),
			fx.Annotate(
				NewContractRepository,
				fx.As(new(ports.ContractRepository)),
			),
			fx.Annotate(
				NewVehicleAssignmentRepository,
				fx.As(new(ports.VehicleAssignmentRepository)),
			),
		),
		fx.Invoke(runMigrationsLifecycle),
	)
}

func newDB(lc fx.Lifecycle, cfg *config.DatabaseConfig, logger *zap.Logger) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := NewDB(ctx, cfg)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing PostgreSQL connection pools")
			return db.Close()
		},
	})

	return db, nil
}

func runMigrationsLifecycle(lc fx.Lifecycle, cfg *config.DatabaseConfig, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if cfg == nil || cfg.MasterURL == "" {
				logger.Warn("DATABASE_MASTER_URL not set, skipping migrations")
				return nil
			}
			logger.Info("Running database migrations")
			if err := RunMigrations(ctx, cfg); err != nil {
				return err
			}
			logger.Info("Migrations completed successfully")
			return nil
		},
	})
}

