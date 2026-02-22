package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

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
	_, err := r.db.Master().Exec(ctx, `
		INSERT INTO legal_entities (id, name, tax_id, deleted_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			tax_id = EXCLUDED.tax_id,
			deleted_at = EXCLUDED.deleted_at
	`, entity.ID, entity.Name, entity.TaxID, entity.DeletedAt)
	return err
}

// FindByID returns a legal entity by ID, excluding soft-deleted.
func (r *LegalEntityRepository) FindByID(ctx context.Context, id string) (*domain.LegalEntity, error) {
	row := r.db.Replica().QueryRow(ctx, `
		SELECT id::text, name, tax_id, deleted_at
		FROM legal_entities
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	var e domain.LegalEntity
	if err := row.Scan(&e.ID, &e.Name, &e.TaxID, &e.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

// FindAll returns all non-deleted legal entities, sorted by ID.
func (r *LegalEntityRepository) FindAll(ctx context.Context) ([]*domain.LegalEntity, error) {
	rows, err := r.db.Replica().Query(ctx, `
		SELECT id::text, name, tax_id, deleted_at
		FROM legal_entities
		WHERE deleted_at IS NULL
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.LegalEntity
	for rows.Next() {
		var e domain.LegalEntity
		if err := rows.Scan(&e.ID, &e.Name, &e.TaxID, &e.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, &e)
	}
	return result, rows.Err()
}

// SoftDelete marks a legal entity as deleted.
func (r *LegalEntityRepository) SoftDelete(ctx context.Context, id string) error {
	res, err := r.db.Master().Exec(ctx, `
		UPDATE legal_entities
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		var n int
		err := r.db.Master().QueryRow(ctx, `SELECT 1 FROM legal_entities WHERE id = $1 AND deleted_at IS NOT NULL`, id).Scan(&n)
		if err == nil {
			return domain.ErrAlreadyDeleted
		}
		return domain.ErrNotFound
	}
	return nil
}

// Undelete restores a soft-deleted legal entity.
func (r *LegalEntityRepository) Undelete(ctx context.Context, id string) error {
	res, err := r.db.Master().Exec(ctx, `
		UPDATE legal_entities SET deleted_at = NULL WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
