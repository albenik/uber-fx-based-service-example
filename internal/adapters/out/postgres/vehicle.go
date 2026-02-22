package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// VehicleRepository implements ports.VehicleRepository.
type VehicleRepository struct {
	db *DB
}

// NewVehicleRepository creates a new VehicleRepository.
func NewVehicleRepository(db *DB) *VehicleRepository {
	return &VehicleRepository{db: db}
}

// Save inserts or updates a vehicle.
func (r *VehicleRepository) Save(ctx context.Context, entity *domain.Vehicle) error {
	row := vehicleToRow(entity)
	const query = `
		INSERT INTO vehicles (id, fleet_id, make, model, year, license_plate, deleted_at)
		VALUES (:id, :fleet_id, :make, :model, :year, :license_plate, :deleted_at)
		ON CONFLICT (id) DO UPDATE SET
			fleet_id = EXCLUDED.fleet_id,
			make = EXCLUDED.make,
			model = EXCLUDED.model,
			year = EXCLUDED.year,
			license_plate = EXCLUDED.license_plate,
			deleted_at = EXCLUDED.deleted_at
	`

	_, err := r.db.Master().NamedExecContext(ctx, query, row)
	return err
}

// FindByID returns a vehicle by ID, excluding soft-deleted.
func (r *VehicleRepository) FindByID(ctx context.Context, id string) (*domain.Vehicle, error) {
	var row vehicleRow
	const query = `
		SELECT id::text, fleet_id::text, make, model, year, license_plate, deleted_at
		FROM vehicles
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

// FindByFleetID returns all non-deleted vehicles for a fleet, sorted by ID.
func (r *VehicleRepository) FindByFleetID(ctx context.Context, fleetID string) ([]*domain.Vehicle, error) {
	var rows []vehicleRow
	const query = `
		SELECT id::text, fleet_id::text, make, model, year, license_plate, deleted_at
		FROM vehicles
		WHERE fleet_id = $1 AND deleted_at IS NULL
		ORDER BY id
	`

	if err := r.db.Replica().SelectContext(ctx, &rows, query, fleetID); err != nil {
		return nil, err
	}

	result := make([]*domain.Vehicle, len(rows))
	for i := range rows {
		result[i] = rows[i].toDomain()
	}

	return result, nil
}

// SoftDelete marks a vehicle as deleted.
func (r *VehicleRepository) SoftDelete(ctx context.Context, id string) error {
	const query = `
		UPDATE vehicles
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
		const query = `SELECT 1 FROM vehicles WHERE id = $1 AND deleted_at IS NOT NULL`
		if err := r.db.Master().GetContext(ctx, &n2, query, id); err == nil {
			return domain.ErrAlreadyDeleted
		}

		return domain.ErrNotFound
	}

	return nil
}

// Undelete restores a soft-deleted vehicle.
func (r *VehicleRepository) Undelete(ctx context.Context, id string) error {
	const query = `UPDATE vehicles SET deleted_at = NULL WHERE id = $1`

	res, err := r.db.Master().ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}

	return nil
}
