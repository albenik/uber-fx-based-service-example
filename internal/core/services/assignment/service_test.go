package assignment_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports/mocks"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services/assignment"
)

func stubIDGen() string { return "test-id" }

func TestService_Assign_RejectsWhenContractInactive(t *testing.T) {
	ctrl := gomock.NewController(t)
	contractRepo := mocks.NewMockContractRepository(ctrl)
	vehicleRepo := mocks.NewMockVehicleRepository(ctrl)
	assignmentRepo := mocks.NewMockVehicleAssignmentRepository(ctrl)

	terminated := time.Now().Add(-1 * time.Hour)
	contract := &domain.Contract{
		ID: "c1", DriverID: "d1", FleetID: "f1",
		StartDate:    time.Now().Add(-24 * time.Hour),
		EndDate:      time.Now().Add(24 * time.Hour),
		TerminatedAt: &terminated,
	}
	contractRepo.EXPECT().FindByID(gomock.Any(), "c1").Return(contract, nil)
	vehicleRepo.EXPECT().FindByID(gomock.Any(), "v1").Return(&domain.Vehicle{ID: "v1", FleetID: "f1"}, nil)

	svc := assignment.New(contractRepo, vehicleRepo, assignmentRepo, zaptest.NewLogger(t), stubIDGen, time.Now)
	_, err := svc.Assign(t.Context(), "c1", "v1")
	assert.ErrorIs(t, err, domain.ErrContractNotActive)
}

func TestService_Assign_RejectsWhenAlreadyAssigned(t *testing.T) {
	ctrl := gomock.NewController(t)
	contractRepo := mocks.NewMockContractRepository(ctrl)
	vehicleRepo := mocks.NewMockVehicleRepository(ctrl)
	assignmentRepo := mocks.NewMockVehicleAssignmentRepository(ctrl)

	contract := &domain.Contract{
		ID: "c1", DriverID: "d1", FleetID: "f1",
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now().Add(24 * time.Hour),
	}
	contractRepo.EXPECT().FindByID(gomock.Any(), "c1").Return(contract, nil)
	vehicleRepo.EXPECT().FindByID(gomock.Any(), "v1").Return(&domain.Vehicle{ID: "v1", FleetID: "f1"}, nil)
	assignmentRepo.EXPECT().FindActiveByDriverIDAndFleetID(gomock.Any(), "d1", "f1").Return(&domain.VehicleAssignment{ID: "a1"}, nil)

	svc := assignment.New(contractRepo, vehicleRepo, assignmentRepo, zaptest.NewLogger(t), stubIDGen, time.Now)
	_, err := svc.Assign(t.Context(), "c1", "v1")
	assert.ErrorIs(t, err, domain.ErrDriverAlreadyAssignedInFleet)
}

func TestService_Assign_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	contractRepo := mocks.NewMockContractRepository(ctrl)
	vehicleRepo := mocks.NewMockVehicleRepository(ctrl)
	assignmentRepo := mocks.NewMockVehicleAssignmentRepository(ctrl)

	contract := &domain.Contract{
		ID: "c1", DriverID: "d1", FleetID: "f1",
		StartDate: time.Now().Add(-24 * time.Hour),
		EndDate:   time.Now().Add(24 * time.Hour),
	}
	contractRepo.EXPECT().FindByID(gomock.Any(), "c1").Return(contract, nil)
	vehicleRepo.EXPECT().FindByID(gomock.Any(), "v1").Return(&domain.Vehicle{ID: "v1", FleetID: "f1"}, nil)
	assignmentRepo.EXPECT().FindActiveByDriverIDAndFleetID(gomock.Any(), "d1", "f1").Return(nil, nil)
	assignmentRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

	svc := assignment.New(contractRepo, vehicleRepo, assignmentRepo, zaptest.NewLogger(t), stubIDGen, time.Now)
	entity, err := svc.Assign(t.Context(), "c1", "v1")
	require.NoError(t, err)
	assert.Equal(t, "test-id", entity.ID)
	assert.Equal(t, "v1", entity.VehicleID)
}
