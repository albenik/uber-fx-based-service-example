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

var _ ports.VehicleRepository = (*repository.MemoryVehicleRepository)(nil)

func TestMemoryVehicleRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryVehicleRepository()

	entity := &domain.Vehicle{ID: "v1", FleetID: "f1", Make: "Toyota", Model: "Camry", Year: 2020, LicensePlate: "ABC-123"}
	require.NoError(t, repo.Save(ctx, entity))

	got, err := repo.FindByID(ctx, "v1")
	require.NoError(t, err)
	assert.Equal(t, "Toyota", got.Make)
	assert.Equal(t, 2020, got.Year)

	list, err := repo.FindByFleetID(ctx, "f1")
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestMemoryVehicleRepository_SoftDelete(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryVehicleRepository()

	entity := &domain.Vehicle{ID: "v1", FleetID: "f1", Make: "Toyota", Model: "Camry", Year: 2020}
	require.NoError(t, repo.Save(ctx, entity))
	require.NoError(t, repo.SoftDelete(ctx, "v1"))

	_, err := repo.FindByID(ctx, "v1")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
