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

var _ ports.FleetRepository = (*repository.MemoryFleetRepository)(nil)

func TestMemoryFleetRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryFleetRepository()

	entity := &domain.Fleet{ID: "f1", LegalEntityID: "le1", Name: "Fleet A"}
	require.NoError(t, repo.Save(ctx, entity))

	got, err := repo.FindByID(ctx, "f1")
	require.NoError(t, err)
	assert.Equal(t, "f1", got.ID)
	assert.Equal(t, "le1", got.LegalEntityID)
	assert.Equal(t, "Fleet A", got.Name)

	list, err := repo.FindByLegalEntityID(ctx, "le1")
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestMemoryFleetRepository_SoftDelete(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryFleetRepository()

	entity := &domain.Fleet{ID: "f1", LegalEntityID: "le1", Name: "Fleet A"}
	require.NoError(t, repo.Save(ctx, entity))
	require.NoError(t, repo.SoftDelete(ctx, "f1"))

	_, err := repo.FindByID(ctx, "f1")
	assert.ErrorIs(t, err, domain.ErrNotFound)

	require.NoError(t, repo.Undelete(ctx, "f1"))
	got, err := repo.FindByID(ctx, "f1")
	require.NoError(t, err)
	assert.Equal(t, "Fleet A", got.Name)
}
