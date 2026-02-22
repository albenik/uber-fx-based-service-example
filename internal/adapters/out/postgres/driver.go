package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// DriverRepository implements ports.DriverRepository.
type DriverRepository struct {
	db *DB
}

// NewDriverRepository creates a new DriverRepository.
func NewDriverRepository(db *DB) *DriverRepository {
	return &DriverRepository{db: db}
}

// Save inserts or updates a driver.
func (r *DriverRepository) Save(ctx context.Context, entity *domain.Driver) error {
	row := driverToRow(entity)
	const query = `
		INSERT INTO drivers (id, first_name, last_name, license_number, deleted_at)
		VALUES (:id, :first_name, :last_name, :license_number, :deleted_at)
		ON CONFLICT (id) DO UPDATE SET
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			license_number = EXCLUDED.license_number,
			deleted_at = EXCLUDED.deleted_at
	`
	_, err := r.db.Master().NamedExecContext(ctx, query, row)
	return err
}

// FindByID returns a driver by ID, excluding soft-deleted.
func (r *DriverRepository) FindByID(ctx context.Context, id string) (*domain.Driver, error) {
	var row driverRow
	const query = `
		SELECT id::text, first_name, last_name, license_number, deleted_at
		FROM drivers
		WHERE id = $1 AND deleted_at IS NULL
	`
	if err := r.db.Replica().GetContext(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return row.toDomain(), nil
}

// FindAll returns all non-deleted drivers, sorted by ID.
func (r *DriverRepository) FindAll(ctx context.Context) ([]*domain.Driver, error) {
	var rows []driverRow
	const query = `
		SELECT id::text, first_name, last_name, license_number, deleted_at
		FROM drivers
		WHERE deleted_at IS NULL
		ORDER BY id
	`
	if err := r.db.Replica().SelectContext(ctx, &rows, query); err != nil {
		return nil, err
	}
	result := make([]*domain.Driver, len(rows))
	for i := range rows {
		result[i] = rows[i].toDomain()
	}
	return result, nil
}

// SoftDelete marks a driver as deleted.
func (r *DriverRepository) SoftDelete(ctx context.Context, id string) error {
	const query = `
		UPDATE drivers
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	res, err := r.db.Master().ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		var n2 int
		const checkQuery = `SELECT 1 FROM drivers WHERE id = $1 AND deleted_at IS NOT NULL`
		switch err := r.db.Master().GetContext(ctx, &n2, checkQuery, id); {
		case err == nil:
			return domain.ErrAlreadyDeleted
		case errors.Is(err, sql.ErrNoRows):
			return domain.ErrNotFound
		default:
			return err
		}
	}
	return nil
}

// Undelete restores a soft-deleted driver.
func (r *DriverRepository) Undelete(ctx context.Context, id string) error {
	const query = `UPDATE drivers SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL`
	res, err := r.db.Master().ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		var n2 int
		const checkQuery = `SELECT 1 FROM drivers WHERE id = $1 AND deleted_at IS NULL`
		switch err := r.db.Master().GetContext(ctx, &n2, checkQuery, id); {
		case err == nil:
			return fmt.Errorf("%w: entity is not deleted", domain.ErrConflict)
		case errors.Is(err, sql.ErrNoRows):
			return domain.ErrNotFound
		default:
			return err
		}
	}
	return nil
}
