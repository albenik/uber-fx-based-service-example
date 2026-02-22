package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// VehicleAssignmentRepository implements ports.VehicleAssignmentRepository.
type VehicleAssignmentRepository struct {
	db *DB
}

// NewVehicleAssignmentRepository creates a new VehicleAssignmentRepository.
func NewVehicleAssignmentRepository(db *DB) *VehicleAssignmentRepository {
	return &VehicleAssignmentRepository{db: db}
}

// Save inserts or updates a vehicle assignment.
func (r *VehicleAssignmentRepository) Save(ctx context.Context, entity *domain.VehicleAssignment) error {
	_, err := r.db.Master().Exec(ctx, `
		INSERT INTO vehicle_assignments (id, driver_id, vehicle_id, contract_id, start_time, end_time, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			driver_id = EXCLUDED.driver_id,
			vehicle_id = EXCLUDED.vehicle_id,
			contract_id = EXCLUDED.contract_id,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			deleted_at = EXCLUDED.deleted_at
	`, entity.ID, entity.DriverID, entity.VehicleID, entity.ContractID, entity.StartTime, entity.EndTime, entity.DeletedAt)
	return err
}

// FindByID returns a vehicle assignment by ID, excluding soft-deleted.
func (r *VehicleAssignmentRepository) FindByID(ctx context.Context, id string) (*domain.VehicleAssignment, error) {
	row := r.db.Replica().QueryRow(ctx, `
		SELECT id::text, driver_id::text, vehicle_id::text, contract_id::text, start_time, end_time, deleted_at
		FROM vehicle_assignments
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	var e domain.VehicleAssignment
	if err := row.Scan(&e.ID, &e.DriverID, &e.VehicleID, &e.ContractID, &e.StartTime, &e.EndTime, &e.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

// FindByContractID returns all non-deleted assignments for a contract, sorted by StartTime.
func (r *VehicleAssignmentRepository) FindByContractID(ctx context.Context, contractID string) ([]*domain.VehicleAssignment, error) {
	rows, err := r.db.Replica().Query(ctx, `
		SELECT id::text, driver_id::text, vehicle_id::text, contract_id::text, start_time, end_time, deleted_at
		FROM vehicle_assignments
		WHERE contract_id = $1 AND deleted_at IS NULL
		ORDER BY start_time
	`, contractID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.VehicleAssignment
	for rows.Next() {
		var e domain.VehicleAssignment
		if err := rows.Scan(&e.ID, &e.DriverID, &e.VehicleID, &e.ContractID, &e.StartTime, &e.EndTime, &e.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, &e)
	}
	return result, rows.Err()
}

// FindActiveByDriverID returns all active (end_time IS NULL) assignments for a driver.
func (r *VehicleAssignmentRepository) FindActiveByDriverID(ctx context.Context, driverID string) ([]*domain.VehicleAssignment, error) {
	rows, err := r.db.Replica().Query(ctx, `
		SELECT id::text, driver_id::text, vehicle_id::text, contract_id::text, start_time, end_time, deleted_at
		FROM vehicle_assignments
		WHERE driver_id = $1 AND end_time IS NULL AND deleted_at IS NULL
	`, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.VehicleAssignment
	for rows.Next() {
		var e domain.VehicleAssignment
		if err := rows.Scan(&e.ID, &e.DriverID, &e.VehicleID, &e.ContractID, &e.StartTime, &e.EndTime, &e.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, &e)
	}
	return result, rows.Err()
}

// FindActiveByDriverIDAndFleetID returns the active assignment for a driver in a fleet (if any).
// Returns (nil, nil) when no active assignment exists.
func (r *VehicleAssignmentRepository) FindActiveByDriverIDAndFleetID(ctx context.Context, driverID, fleetID string) (*domain.VehicleAssignment, error) {
	row := r.db.Replica().QueryRow(ctx, `
		SELECT va.id::text, va.driver_id::text, va.vehicle_id::text, va.contract_id::text, va.start_time, va.end_time, va.deleted_at
		FROM vehicle_assignments va
		JOIN vehicles v ON v.id = va.vehicle_id
		WHERE va.driver_id = $1 AND v.fleet_id = $2
			AND va.end_time IS NULL AND va.deleted_at IS NULL AND v.deleted_at IS NULL
		LIMIT 1
	`, driverID, fleetID)
	var e domain.VehicleAssignment
	if err := row.Scan(&e.ID, &e.DriverID, &e.VehicleID, &e.ContractID, &e.StartTime, &e.EndTime, &e.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

// SoftDelete marks a vehicle assignment as deleted.
func (r *VehicleAssignmentRepository) SoftDelete(ctx context.Context, id string) error {
	res, err := r.db.Master().Exec(ctx, `
		UPDATE vehicle_assignments
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		var n int
		err := r.db.Master().QueryRow(ctx, `SELECT 1 FROM vehicle_assignments WHERE id = $1 AND deleted_at IS NOT NULL`, id).Scan(&n)
		if err == nil {
			return domain.ErrAlreadyDeleted
		}
		return domain.ErrNotFound
	}
	return nil
}

// Undelete restores a soft-deleted vehicle assignment.
func (r *VehicleAssignmentRepository) Undelete(ctx context.Context, id string) error {
	res, err := r.db.Master().Exec(ctx, `
		UPDATE vehicle_assignments SET deleted_at = NULL WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
