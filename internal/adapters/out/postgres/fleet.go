package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

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
	_, err := r.db.Master().Exec(ctx, `
		INSERT INTO fleets (id, legal_entity_id, name, deleted_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			legal_entity_id = EXCLUDED.legal_entity_id,
			name = EXCLUDED.name,
			deleted_at = EXCLUDED.deleted_at
	`, entity.ID, entity.LegalEntityID, entity.Name, entity.DeletedAt)
	return err
}

// FindByID returns a fleet by ID, excluding soft-deleted.
func (r *FleetRepository) FindByID(ctx context.Context, id string) (*domain.Fleet, error) {
	row := r.db.Replica().QueryRow(ctx, `
		SELECT id::text, legal_entity_id::text, name, deleted_at
		FROM fleets
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	var e domain.Fleet
	if err := row.Scan(&e.ID, &e.LegalEntityID, &e.Name, &e.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

// FindByLegalEntityID returns all non-deleted fleets for a legal entity, sorted by ID.
func (r *FleetRepository) FindByLegalEntityID(ctx context.Context, legalEntityID string) ([]*domain.Fleet, error) {
	rows, err := r.db.Replica().Query(ctx, `
		SELECT id::text, legal_entity_id::text, name, deleted_at
		FROM fleets
		WHERE legal_entity_id = $1 AND deleted_at IS NULL
		ORDER BY id
	`, legalEntityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.Fleet
	for rows.Next() {
		var e domain.Fleet
		if err := rows.Scan(&e.ID, &e.LegalEntityID, &e.Name, &e.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, &e)
	}
	return result, rows.Err()
}

// SoftDelete marks a fleet as deleted.
func (r *FleetRepository) SoftDelete(ctx context.Context, id string) error {
	res, err := r.db.Master().Exec(ctx, `
		UPDATE fleets
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		var n int
		err := r.db.Master().QueryRow(ctx, `SELECT 1 FROM fleets WHERE id = $1 AND deleted_at IS NOT NULL`, id).Scan(&n)
		if err == nil {
			return domain.ErrAlreadyDeleted
		}
		return domain.ErrNotFound
	}
	return nil
}

// Undelete restores a soft-deleted fleet.
func (r *FleetRepository) Undelete(ctx context.Context, id string) error {
	res, err := r.db.Master().Exec(ctx, `
		UPDATE fleets SET deleted_at = NULL WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
