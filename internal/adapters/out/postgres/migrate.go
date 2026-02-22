package postgres

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/albenik/uber-fx-based-service-example/internal/config"
	"github.com/albenik/uber-fx-based-service-example/migrations"
)

// RunMigrations runs goose Up migrations against the master database.
func RunMigrations(ctx context.Context, cfg *config.DatabaseConfig) error {
	if cfg == nil || cfg.MasterURL == "" {
		return nil
	}

	connConfig, err := pgx.ParseConfig(cfg.MasterURL)
	if err != nil {
		return err
	}
	connector := stdlib.GetConnector(*connConfig)
	db := sql.OpenDB(connector)
	defer db.Close() //nolint:errcheck // it's ok here

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.SQL)
	if err != nil {
		return err
	}

	_, err = provider.Up(ctx)
	return err
}
