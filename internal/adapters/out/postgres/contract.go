package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

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
	row := contractToRow(entity)
	const query = `
		INSERT INTO contracts (id, driver_id, legal_entity_id, fleet_id, start_date, end_date, terminated_at, terminated_by, deleted_at)
		VALUES (:id, :driver_id, :legal_entity_id, :fleet_id, :start_date, :end_date, :terminated_at, :terminated_by, :deleted_at)
		ON CONFLICT (id) DO UPDATE SET
			driver_id = EXCLUDED.driver_id,
			legal_entity_id = EXCLUDED.legal_entity_id,
			fleet_id = EXCLUDED.fleet_id,
			start_date = EXCLUDED.start_date,
			end_date = EXCLUDED.end_date,
			terminated_at = EXCLUDED.terminated_at,
			terminated_by = EXCLUDED.terminated_by,
			deleted_at = EXCLUDED.deleted_at
	`
	_, err := r.db.Master().NamedExecContext(ctx, query, row)
	return err
}

// FindByID returns a contract by ID, excluding soft-deleted.
func (r *ContractRepository) FindByID(ctx context.Context, id string) (*domain.Contract, error) {
	var row contractRow
	const query = `
		SELECT id::text, driver_id::text, legal_entity_id::text, fleet_id::text,
			start_date, end_date, terminated_at, terminated_by, deleted_at
		FROM contracts
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

// FindByDriverID returns all non-deleted contracts for a driver, sorted by StartDate.
func (r *ContractRepository) FindByDriverID(ctx context.Context, driverID string) ([]*domain.Contract, error) {
	var rows []contractRow
	const query = `
		SELECT id::text, driver_id::text, legal_entity_id::text, fleet_id::text,
			start_date, end_date, terminated_at, terminated_by, deleted_at
		FROM contracts
		WHERE driver_id = $1 AND deleted_at IS NULL
		ORDER BY start_date
	`
	if err := r.db.Replica().SelectContext(ctx, &rows, query, driverID); err != nil {
		return nil, err
	}
	result := make([]*domain.Contract, len(rows))
	for i := range rows {
		result[i] = rows[i].toDomain()
	}
	return result, nil
}

// FindOverlapping returns contracts that overlap with the given date range for the same driver/legal/fleet.
func (r *ContractRepository) FindOverlapping(ctx context.Context, driverID, legalEntityID, fleetID string, startDate, endDate time.Time, excludeID string) ([]*domain.Contract, error) {
	var rows []contractRow
	const query = `
		SELECT id::text, driver_id::text, legal_entity_id::text, fleet_id::text,
			start_date, end_date, terminated_at, terminated_by, deleted_at
		FROM contracts
		WHERE driver_id = $1 AND legal_entity_id = $2 AND fleet_id = $3
			AND id != $4 AND deleted_at IS NULL
			AND $5 < COALESCE(terminated_at::date, end_date)
			AND $6 > start_date
	`
	if err := r.db.Replica().SelectContext(ctx, &rows, query, driverID, legalEntityID, fleetID, excludeID, startDate, endDate); err != nil {
		return nil, err
	}
	result := make([]*domain.Contract, len(rows))
	for i := range rows {
		result[i] = rows[i].toDomain()
	}
	return result, nil
}

// SoftDelete marks a contract as deleted.
func (r *ContractRepository) SoftDelete(ctx context.Context, id string) error {
	const query = `
		UPDATE contracts
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
		const checkQuery = `SELECT 1 FROM contracts WHERE id = $1 AND deleted_at IS NOT NULL`
		if err := r.db.Master().GetContext(ctx, &n2, checkQuery, id); err == nil {
			return domain.ErrAlreadyDeleted
		}
		return domain.ErrNotFound
	}
	return nil
}

// Undelete restores a soft-deleted contract.
func (r *ContractRepository) Undelete(ctx context.Context, id string) error {
	const query = `UPDATE contracts SET deleted_at = NULL WHERE id = $1`
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
