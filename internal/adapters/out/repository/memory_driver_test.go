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

var _ ports.DriverRepository = (*repository.MemoryDriverRepository)(nil)

func TestMemoryDriverRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryDriverRepository()

	entity := &domain.Driver{ID: "d1", FirstName: "John", LastName: "Doe", LicenseNumber: "DL-123"}
	require.NoError(t, repo.Save(ctx, entity))

	got, err := repo.FindByID(ctx, "d1")
	require.NoError(t, err)
	assert.Equal(t, "John", got.FirstName)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)
}

func TestMemoryDriverRepository_SoftDelete(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryDriverRepository()

	entity := &domain.Driver{ID: "d1", FirstName: "John", LastName: "Doe", LicenseNumber: "DL-123"}
	require.NoError(t, repo.Save(ctx, entity))
	require.NoError(t, repo.SoftDelete(ctx, "d1"))

	_, err := repo.FindByID(ctx, "d1")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
