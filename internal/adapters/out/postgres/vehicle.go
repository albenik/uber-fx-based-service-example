package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

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
	_, err := r.db.Master().Exec(ctx, `
		INSERT INTO vehicles (id, fleet_id, make, model, year, license_plate, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			fleet_id = EXCLUDED.fleet_id,
			make = EXCLUDED.make,
			model = EXCLUDED.model,
			year = EXCLUDED.year,
			license_plate = EXCLUDED.license_plate,
			deleted_at = EXCLUDED.deleted_at
	`, entity.ID, entity.FleetID, entity.Make, entity.Model, entity.Year, entity.LicensePlate, entity.DeletedAt)
	return err
}

// FindByID returns a vehicle by ID, excluding soft-deleted.
func (r *VehicleRepository) FindByID(ctx context.Context, id string) (*domain.Vehicle, error) {
	row := r.db.Replica().QueryRow(ctx, `
		SELECT id::text, fleet_id::text, make, model, year, license_plate, deleted_at
		FROM vehicles
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	var e domain.Vehicle
	if err := row.Scan(&e.ID, &e.FleetID, &e.Make, &e.Model, &e.Year, &e.LicensePlate, &e.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

// FindByFleetID returns all non-deleted vehicles for a fleet, sorted by ID.
func (r *VehicleRepository) FindByFleetID(ctx context.Context, fleetID string) ([]*domain.Vehicle, error) {
	rows, err := r.db.Replica().Query(ctx, `
		SELECT id::text, fleet_id::text, make, model, year, license_plate, deleted_at
		FROM vehicles
		WHERE fleet_id = $1 AND deleted_at IS NULL
		ORDER BY id
	`, fleetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.Vehicle
	for rows.Next() {
		var e domain.Vehicle
		if err := rows.Scan(&e.ID, &e.FleetID, &e.Make, &e.Model, &e.Year, &e.LicensePlate, &e.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, &e)
	}
	return result, rows.Err()
}

// SoftDelete marks a vehicle as deleted.
func (r *VehicleRepository) SoftDelete(ctx context.Context, id string) error {
	res, err := r.db.Master().Exec(ctx, `
		UPDATE vehicles
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		var n int
		err := r.db.Master().QueryRow(ctx, `SELECT 1 FROM vehicles WHERE id = $1 AND deleted_at IS NOT NULL`, id).Scan(&n)
		if err == nil {
			return domain.ErrAlreadyDeleted
		}
		return domain.ErrNotFound
	}
	return nil
}

// Undelete restores a soft-deleted vehicle.
func (r *VehicleRepository) Undelete(ctx context.Context, id string) error {
	res, err := r.db.Master().Exec(ctx, `
		UPDATE vehicles SET deleted_at = NULL WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
