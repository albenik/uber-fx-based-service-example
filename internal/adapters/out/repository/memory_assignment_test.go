package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/albenik/uber-fx-based-service-example/internal/adapters/out/repository"
	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

var _ ports.VehicleAssignmentRepository = (*repository.MemoryVehicleAssignmentRepository)(nil)

func TestMemoryVehicleAssignmentRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	vehicleRepo := repository.NewMemoryVehicleRepository()
	vehicleRepo.Save(ctx, &domain.Vehicle{ID: "v1", FleetID: "f1", Make: "Toyota", Model: "Camry", Year: 2020})
	assignmentRepo := repository.NewMemoryVehicleAssignmentRepository(vehicleRepo)

	entity := &domain.VehicleAssignment{
		ID: "a1", DriverID: "d1", VehicleID: "v1", ContractID: "c1",
		StartTime: time.Now(),
	}
	require.NoError(t, assignmentRepo.Save(ctx, entity))

	got, err := assignmentRepo.FindByID(ctx, "a1")
	require.NoError(t, err)
	assert.Equal(t, "d1", got.DriverID)
	assert.Equal(t, "v1", got.VehicleID)

	active, err := assignmentRepo.FindActiveByDriverIDAndFleetID(ctx, "d1", "f1")
	require.NoError(t, err)
	require.NotNil(t, active)
	assert.Equal(t, "a1", active.ID)
}
