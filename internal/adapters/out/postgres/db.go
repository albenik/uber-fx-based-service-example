package postgres

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/albenik/uber-fx-based-service-example/internal/config"
)

var errMissingMasterURL = errors.New("database config with MasterURL is required")

// DB holds master and replica connection pools for read/write splitting.
type DB struct {
	master  *sqlx.DB
	replica *sqlx.DB
}

// NewDB creates master and optionally replica pools. If ReplicaURL is empty, master is used for both.
func NewDB(ctx context.Context, cfg *config.DatabaseConfig) (*DB, error) {
	if cfg == nil || cfg.MasterURL == "" {
		return nil, errMissingMasterURL
	}

	master, err := sqlx.ConnectContext(ctx, "pgx", cfg.MasterURL)
	if err != nil {
		return nil, err
	}

	if err := master.PingContext(ctx); err != nil {
		_ = master.Close()
		return nil, err
	}

	replica := master
	if cfg.ReplicaURL != "" {
		replica, err = sqlx.ConnectContext(ctx, "pgx", cfg.ReplicaURL)
		if err != nil {
			_ = master.Close()
			return nil, err
		}
		if err := replica.PingContext(ctx); err != nil {
			_ = master.Close()
			_ = replica.Close()
			return nil, err
		}
	}

	return &DB{master: master, replica: replica}, nil
}

// Master returns the pool for write operations.
func (db *DB) Master() *sqlx.DB {
	return db.master
}

// Replica returns the pool for read operations. Same as Master if no replica is configured.
func (db *DB) Replica() *sqlx.DB {
	return db.replica
}

// Close closes both pools.
func (db *DB) Close() error {
	if err := db.master.Close(); err != nil {
		return err
	}
	if db.replica != db.master {
		return db.replica.Close()
	}
	return nil
}
