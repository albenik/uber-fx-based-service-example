package postgres

import (
	"context"
	"database/sql"
	"errors"

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
	row := vehicleAssignmentToRow(entity)
	const query = `
		INSERT INTO vehicle_assignments (id, driver_id, vehicle_id, contract_id, start_time, end_time, deleted_at)
		VALUES (:id, :driver_id, :vehicle_id, :contract_id, :start_time, :end_time, :deleted_at)
		ON CONFLICT (id) DO UPDATE SET
			driver_id = EXCLUDED.driver_id,
			vehicle_id = EXCLUDED.vehicle_id,
			contract_id = EXCLUDED.contract_id,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			deleted_at = EXCLUDED.deleted_at
	`
	_, err := r.db.Master().NamedExecContext(ctx, query, row)
	return err
}

// FindByID returns a vehicle assignment by ID, excluding soft-deleted.
func (r *VehicleAssignmentRepository) FindByID(ctx context.Context, id string) (*domain.VehicleAssignment, error) {
	var row vehicleAssignmentRow
	const query = `
		SELECT id::text, driver_id::text, vehicle_id::text, contract_id::text, start_time, end_time, deleted_at
		FROM vehicle_assignments
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

// FindByContractID returns all non-deleted assignments for a contract, sorted by StartTime.
func (r *VehicleAssignmentRepository) FindByContractID(
	ctx context.Context,
	contractID string,
) ([]*domain.VehicleAssignment, error) {
	var rows []vehicleAssignmentRow
	const query = `
		SELECT id::text, driver_id::text, vehicle_id::text, contract_id::text, start_time, end_time, deleted_at
		FROM vehicle_assignments
		WHERE contract_id = $1 AND deleted_at IS NULL
		ORDER BY start_time
	`

	if err := r.db.Replica().SelectContext(ctx, &rows, query, contractID); err != nil {
		return nil, err
	}
	result := make([]*domain.VehicleAssignment, len(rows))
	for i := range rows {
		result[i] = rows[i].toDomain()
	}
	return result, nil
}

// FindActiveByDriverID returns all active (end_time IS NULL) assignments for a driver.
func (r *VehicleAssignmentRepository) FindActiveByDriverID(
	ctx context.Context,
	driverID string,
) ([]*domain.VehicleAssignment, error) {
	var rows []vehicleAssignmentRow
	const query = `
		SELECT id::text, driver_id::text, vehicle_id::text, contract_id::text, start_time, end_time, deleted_at
		FROM vehicle_assignments
		WHERE driver_id = $1 AND end_time IS NULL AND deleted_at IS NULL
	`

	if err := r.db.Replica().SelectContext(ctx, &rows, query, driverID); err != nil {
		return nil, err
	}

	result := make([]*domain.VehicleAssignment, len(rows))
	for i := range rows {
		result[i] = rows[i].toDomain()
	}

	return result, nil
}

// FindActiveByDriverIDAndFleetID returns the active assignment for a driver in a fleet (if any).
// Returns (nil, nil) when no active assignment exists.
func (r *VehicleAssignmentRepository) FindActiveByDriverIDAndFleetID(
	ctx context.Context,
	driverID, fleetID string,
) (*domain.VehicleAssignment, error) {
	var row vehicleAssignmentRow
	const query = `
		SELECT
			va.id::text,
			va.driver_id::text,
			va.vehicle_id::text,
			va.contract_id::text,
			va.start_time,
			va.end_time,
			va.deleted_at -- no comma after last field
		FROM vehicle_assignments va
		JOIN vehicles v ON v.id = va.vehicle_id
		WHERE va.driver_id = $1 AND v.fleet_id = $2
			AND va.end_time IS NULL AND va.deleted_at IS NULL AND v.deleted_at IS NULL
		LIMIT 1
	`

	if err := r.db.Replica().GetContext(ctx, &row, query, driverID, fleetID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return row.toDomain(), nil
}

// SoftDelete marks a vehicle assignment as deleted.
func (r *VehicleAssignmentRepository) SoftDelete(ctx context.Context, id string) error {
	const query = `
		UPDATE vehicle_assignments
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
		const query = `SELECT 1 FROM vehicle_assignments WHERE id = $1 AND deleted_at IS NOT NULL`
		err := r.db.Master().GetContext(ctx, &n2, query, id)
		if err == nil {
			return domain.ErrAlreadyDeleted
		}
		return domain.ErrNotFound
	}

	return nil
}

// Undelete restores a soft-deleted vehicle assignment.
func (r *VehicleAssignmentRepository) Undelete(ctx context.Context, id string) error {
	const query = `UPDATE vehicle_assignments SET deleted_at = NULL WHERE id = $1`
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
