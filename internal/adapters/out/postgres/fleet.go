package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// FleetRepository implements ports.FleetRepository.
type FleetRepository struct {
	db *DB
}

// NewFleetRepository creates a new FleetRepository.
func NewFleetRepository(db *DB) *FleetRepository {
	return &FleetRepository{db: db}
}

// Save inserts or updates a fleet.
func (r *FleetRepository) Save(ctx context.Context, entity *domain.Fleet) error {
	row := fleetToRow(entity)
	const query = `
		INSERT INTO fleets (id, legal_entity_id, name, deleted_at)
		VALUES (:id, :legal_entity_id, :name, :deleted_at)
		ON CONFLICT (id) DO UPDATE SET
			legal_entity_id = EXCLUDED.legal_entity_id,
			name = EXCLUDED.name,
			deleted_at = EXCLUDED.deleted_at
	`
	_, err := r.db.Master().NamedExecContext(ctx, query, row)
	return err
}

// FindByID returns a fleet by ID, excluding soft-deleted.
func (r *FleetRepository) FindByID(ctx context.Context, id string) (*domain.Fleet, error) {
	var row fleetRow
	const query = `
		SELECT id::text, legal_entity_id::text, name, deleted_at
		FROM fleets
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

// FindByLegalEntityID returns all non-deleted fleets for a legal entity, sorted by ID.
func (r *FleetRepository) FindByLegalEntityID(ctx context.Context, legalEntityID string) ([]*domain.Fleet, error) {
	var rows []fleetRow
	const query = `
		SELECT id::text, legal_entity_id::text, name, deleted_at
		FROM fleets
		WHERE legal_entity_id = $1 AND deleted_at IS NULL
		ORDER BY id
	`
	if err := r.db.Replica().SelectContext(ctx, &rows, query, legalEntityID); err != nil {
		return nil, err
	}
	result := make([]*domain.Fleet, len(rows))
	for i := range rows {
		result[i] = rows[i].toDomain()
	}
	return result, nil
}

// SoftDelete marks a fleet as deleted.
func (r *FleetRepository) SoftDelete(ctx context.Context, id string) error {
	const query = `
		UPDATE fleets
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
		const checkQuery = `SELECT 1 FROM fleets WHERE id = $1 AND deleted_at IS NOT NULL`
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

// Undelete restores a soft-deleted fleet.
func (r *FleetRepository) Undelete(ctx context.Context, id string) error {
	const query = `UPDATE fleets SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL`
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
		const checkQuery = `SELECT 1 FROM fleets WHERE id = $1 AND deleted_at IS NULL`
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
