package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/albenik/uber-fx-based-service-example/internal/adapters/out/repository"
	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

var _ ports.LegalEntityRepository = (*repository.MemoryLegalEntityRepository)(nil)

func TestMemoryLegalEntityRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryLegalEntityRepository()

	entity := &domain.LegalEntity{ID: "1", Name: "Acme", TaxID: "123"}
	require.NoError(t, repo.Save(ctx, entity))

	got, err := repo.FindByID(ctx, "1")
	require.NoError(t, err)
	assert.Equal(t, "1", got.ID)
	assert.Equal(t, "Acme", got.Name)
	assert.Equal(t, "123", got.TaxID)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)
}

func TestMemoryLegalEntityRepository_SoftDelete(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryLegalEntityRepository()

	entity := &domain.LegalEntity{ID: "1", Name: "Acme", TaxID: "123"}
	require.NoError(t, repo.Save(ctx, entity))

	require.NoError(t, repo.SoftDelete(ctx, "1"))

	_, err := repo.FindByID(ctx, "1")
	assert.ErrorIs(t, err, domain.ErrNotFound)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Empty(t, all)
}

func TestMemoryLegalEntityRepository_Undelete(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryLegalEntityRepository()

	entity := &domain.LegalEntity{ID: "1", Name: "Acme", TaxID: "123"}
	require.NoError(t, repo.Save(ctx, entity))
	require.NoError(t, repo.SoftDelete(ctx, "1"))

	require.NoError(t, repo.Undelete(ctx, "1"))

	got, err := repo.FindByID(ctx, "1")
	require.NoError(t, err)
	assert.Equal(t, "Acme", got.Name)
}

func TestMemoryLegalEntityRepository_SoftDelete_AlreadyDeleted(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryLegalEntityRepository()

	entity := &domain.LegalEntity{ID: "1", Name: "Acme", TaxID: "123"}
	require.NoError(t, repo.Save(ctx, entity))
	require.NoError(t, repo.SoftDelete(ctx, "1"))

	err := repo.SoftDelete(ctx, "1")
	assert.ErrorIs(t, err, domain.ErrAlreadyDeleted)
}
