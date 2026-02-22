package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

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
	_, err := r.db.Master().Exec(ctx, `
		INSERT INTO drivers (id, first_name, last_name, license_number, deleted_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			license_number = EXCLUDED.license_number,
			deleted_at = EXCLUDED.deleted_at
	`, entity.ID, entity.FirstName, entity.LastName, entity.LicenseNumber, entity.DeletedAt)
	return err
}

// FindByID returns a driver by ID, excluding soft-deleted.
func (r *DriverRepository) FindByID(ctx context.Context, id string) (*domain.Driver, error) {
	row := r.db.Replica().QueryRow(ctx, `
		SELECT id::text, first_name, last_name, license_number, deleted_at
		FROM drivers
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	var e domain.Driver
	if err := row.Scan(&e.ID, &e.FirstName, &e.LastName, &e.LicenseNumber, &e.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

// FindAll returns all non-deleted drivers, sorted by ID.
func (r *DriverRepository) FindAll(ctx context.Context) ([]*domain.Driver, error) {
	rows, err := r.db.Replica().Query(ctx, `
		SELECT id::text, first_name, last_name, license_number, deleted_at
		FROM drivers
		WHERE deleted_at IS NULL
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.Driver
	for rows.Next() {
		var e domain.Driver
		if err := rows.Scan(&e.ID, &e.FirstName, &e.LastName, &e.LicenseNumber, &e.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, &e)
	}
	return result, rows.Err()
}

// SoftDelete marks a driver as deleted.
func (r *DriverRepository) SoftDelete(ctx context.Context, id string) error {
	res, err := r.db.Master().Exec(ctx, `
		UPDATE drivers
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		var n int
		err := r.db.Master().QueryRow(ctx, `SELECT 1 FROM drivers WHERE id = $1 AND deleted_at IS NOT NULL`, id).Scan(&n)
		if err == nil {
			return domain.ErrAlreadyDeleted
		}
		return domain.ErrNotFound
	}
	return nil
}

// Undelete restores a soft-deleted driver.
func (r *DriverRepository) Undelete(ctx context.Context, id string) error {
	res, err := r.db.Master().Exec(ctx, `
		UPDATE drivers SET deleted_at = NULL WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
