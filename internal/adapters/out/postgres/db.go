package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/albenik/uber-fx-based-service-example/internal/config"
)

var errMissingMasterURL = errors.New("database config with MasterURL is required")

// DB holds master and replica connection pools for read/write splitting.
type DB struct {
	master  *pgxpool.Pool
	replica *pgxpool.Pool
}

// NewDB creates master and optionally replica pools. If ReplicaURL is empty, master is used for both.
func NewDB(ctx context.Context, cfg *config.DatabaseConfig) (*DB, error) {
	if cfg == nil || cfg.MasterURL == "" {
		return nil, errMissingMasterURL
	}

	master, err := pgxpool.New(ctx, cfg.MasterURL)
	if err != nil {
		return nil, err
	}

	if err := master.Ping(ctx); err != nil {
		master.Close()
		return nil, err
	}

	replica := master
	if cfg.ReplicaURL != "" {
		replica, err = pgxpool.New(ctx, cfg.ReplicaURL)
		if err != nil {
			master.Close()
			return nil, err
		}
		if err := replica.Ping(ctx); err != nil {
			master.Close()
			replica.Close()
			return nil, err
		}
	}

	return &DB{master: master, replica: replica}, nil
}

// Master returns the pool for write operations.
func (db *DB) Master() *pgxpool.Pool {
	return db.master
}

// Replica returns the pool for read operations. Same as Master if no replica is configured.
func (db *DB) Replica() *pgxpool.Pool {
	return db.replica
}

// Close closes both pools.
func (db *DB) Close() {
	db.master.Close()
	if db.replica != db.master {
		db.replica.Close()
	}
}
