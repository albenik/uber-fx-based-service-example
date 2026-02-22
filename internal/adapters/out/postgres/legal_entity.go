package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// LegalEntityRepository implements ports.LegalEntityRepository.
type LegalEntityRepository struct {
	db *DB
}

// NewLegalEntityRepository creates a new LegalEntityRepository.
func NewLegalEntityRepository(db *DB) *LegalEntityRepository {
	return &LegalEntityRepository{db: db}
}

// Save inserts or updates a legal entity.
func (r *LegalEntityRepository) Save(ctx context.Context, entity *domain.LegalEntity) error {
	row := legalEntityToRow(entity)
	const query = `
		INSERT INTO legal_entities (id, name, tax_id, deleted_at)
		VALUES (:id, :name, :tax_id, :deleted_at)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			tax_id = EXCLUDED.tax_id,
			deleted_at = EXCLUDED.deleted_at
	`

	_, err := r.db.Master().NamedExecContext(ctx, query, row)
	return err
}

// FindByID returns a legal entity by ID, excluding soft-deleted.
func (r *LegalEntityRepository) FindByID(ctx context.Context, id string) (*domain.LegalEntity, error) {
	var row legalEntityRow
	const query = `
		SELECT id::text, name, tax_id, deleted_at
		FROM legal_entities
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

// FindAll returns all non-deleted legal entities, sorted by ID.
func (r *LegalEntityRepository) FindAll(ctx context.Context) ([]*domain.LegalEntity, error) {
	var rows []legalEntityRow
	const query = `
		SELECT id::text, name, tax_id, deleted_at
		FROM legal_entities
		WHERE deleted_at IS NULL
		ORDER BY id
	`

	if err := r.db.Replica().SelectContext(ctx, &rows, query); err != nil {
		return nil, err
	}

	result := make([]*domain.LegalEntity, len(rows))
	for i := range rows {
		result[i] = rows[i].toDomain()
	}

	return result, nil
}

// SoftDelete marks a legal entity as deleted.
func (r *LegalEntityRepository) SoftDelete(ctx context.Context, id string) error {
	const query = `
		UPDATE legal_entities
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
		const checkQuery = `SELECT 1 FROM legal_entities WHERE id = $1 AND deleted_at IS NOT NULL`
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

// Undelete restores a soft-deleted legal entity.
func (r *LegalEntityRepository) Undelete(ctx context.Context, id string) error {
	const query = `UPDATE legal_entities SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL`

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
		const checkQuery = `SELECT 1 FROM legal_entities WHERE id = $1 AND deleted_at IS NULL`
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
