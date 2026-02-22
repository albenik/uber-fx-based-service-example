package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// ContractRepository implements ports.ContractRepository.
type ContractRepository struct {
	db *DB
}

// NewContractRepository creates a new ContractRepository.
func NewContractRepository(db *DB) *ContractRepository {
	return &ContractRepository{db: db}
}

// Save inserts or updates a contract.
func (r *ContractRepository) Save(ctx context.Context, entity *domain.Contract) error {
	_, err := r.db.Master().Exec(ctx, `
		INSERT INTO contracts (id, driver_id, legal_entity_id, fleet_id, start_date, end_date, terminated_at, terminated_by, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			driver_id = EXCLUDED.driver_id,
			legal_entity_id = EXCLUDED.legal_entity_id,
			fleet_id = EXCLUDED.fleet_id,
			start_date = EXCLUDED.start_date,
			end_date = EXCLUDED.end_date,
			terminated_at = EXCLUDED.terminated_at,
			terminated_by = EXCLUDED.terminated_by,
			deleted_at = EXCLUDED.deleted_at
	`, entity.ID, entity.DriverID, entity.LegalEntityID, entity.FleetID,
		entity.StartDate, entity.EndDate, entity.TerminatedAt, entity.TerminatedBy, entity.DeletedAt)
	return err
}

// FindByID returns a contract by ID, excluding soft-deleted.
func (r *ContractRepository) FindByID(ctx context.Context, id string) (*domain.Contract, error) {
	row := r.db.Replica().QueryRow(ctx, `
		SELECT id::text, driver_id::text, legal_entity_id::text, fleet_id::text,
			start_date, end_date, terminated_at, terminated_by, deleted_at
		FROM contracts
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	var e domain.Contract
	if err := row.Scan(&e.ID, &e.DriverID, &e.LegalEntityID, &e.FleetID,
		&e.StartDate, &e.EndDate, &e.TerminatedAt, &e.TerminatedBy, &e.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

// FindByDriverID returns all non-deleted contracts for a driver, sorted by StartDate.
func (r *ContractRepository) FindByDriverID(ctx context.Context, driverID string) ([]*domain.Contract, error) {
	rows, err := r.db.Replica().Query(ctx, `
		SELECT id::text, driver_id::text, legal_entity_id::text, fleet_id::text,
			start_date, end_date, terminated_at, terminated_by, deleted_at
		FROM contracts
		WHERE driver_id = $1 AND deleted_at IS NULL
		ORDER BY start_date
	`, driverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.Contract
	for rows.Next() {
		var e domain.Contract
		if err := rows.Scan(&e.ID, &e.DriverID, &e.LegalEntityID, &e.FleetID,
			&e.StartDate, &e.EndDate, &e.TerminatedAt, &e.TerminatedBy, &e.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, &e)
	}
	return result, rows.Err()
}

// FindOverlapping returns contracts that overlap with the given date range for the same driver/legal/fleet.
func (r *ContractRepository) FindOverlapping(ctx context.Context, driverID, legalEntityID, fleetID string, startDate, endDate time.Time, excludeID string) ([]*domain.Contract, error) {
	rows, err := r.db.Replica().Query(ctx, `
		SELECT id::text, driver_id::text, legal_entity_id::text, fleet_id::text,
			start_date, end_date, terminated_at, terminated_by, deleted_at
		FROM contracts
		WHERE driver_id = $1 AND legal_entity_id = $2 AND fleet_id = $3
			AND id != $4 AND deleted_at IS NULL
			AND $5 < COALESCE(terminated_at::date, end_date)
			AND $6 > start_date
	`, driverID, legalEntityID, fleetID, excludeID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.Contract
	for rows.Next() {
		var e domain.Contract
		if err := rows.Scan(&e.ID, &e.DriverID, &e.LegalEntityID, &e.FleetID,
			&e.StartDate, &e.EndDate, &e.TerminatedAt, &e.TerminatedBy, &e.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, &e)
	}
	return result, rows.Err()
}

// SoftDelete marks a contract as deleted.
func (r *ContractRepository) SoftDelete(ctx context.Context, id string) error {
	res, err := r.db.Master().Exec(ctx, `
		UPDATE contracts
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		var n int
		err := r.db.Master().QueryRow(ctx, `SELECT 1 FROM contracts WHERE id = $1 AND deleted_at IS NOT NULL`, id).Scan(&n)
		if err == nil {
			return domain.ErrAlreadyDeleted
		}
		return domain.ErrNotFound
	}
	return nil
}

// Undelete restores a soft-deleted contract.
func (r *ContractRepository) Undelete(ctx context.Context, id string) error {
	res, err := r.db.Master().Exec(ctx, `
		UPDATE contracts SET deleted_at = NULL WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
