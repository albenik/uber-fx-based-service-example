package driver_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports/mocks"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services/driver"
)

func stubIDGen() string { return "test-id" }

func TestService_Delete_RejectsWhenActiveContracts(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockDriverRepository(ctrl)
	contractRepo := mocks.NewMockContractRepository(ctrl)
	assignmentRepo := mocks.NewMockVehicleAssignmentRepository(ctrl)

	future := time.Now().Add(24 * time.Hour)
	contracts := []*domain.Contract{
		{ID: "c1", DriverID: "d1", TerminatedAt: nil, EndDate: future},
	}
	contractRepo.EXPECT().FindByDriverID(gomock.Any(), "d1").Return(contracts, nil)

	validator := mocks.NewMockDriverLicenseValidator(ctrl)
	svc := driver.New(repo, contractRepo, assignmentRepo, validator, zaptest.NewLogger(t), stubIDGen)
	err := svc.Delete(context.Background(), "d1")
	assert.ErrorIs(t, err, domain.ErrDriverHasActiveContracts)
}

func TestService_Delete_RejectsWhenActiveAssignments(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockDriverRepository(ctrl)
	contractRepo := mocks.NewMockContractRepository(ctrl)
	assignmentRepo := mocks.NewMockVehicleAssignmentRepository(ctrl)

	contractRepo.EXPECT().FindByDriverID(gomock.Any(), "d1").Return([]*domain.Contract{}, nil)
	assignmentRepo.EXPECT().FindActiveByDriverID(gomock.Any(), "d1").Return([]*domain.VehicleAssignment{{ID: "a1"}}, nil)

	validator := mocks.NewMockDriverLicenseValidator(ctrl)
	svc := driver.New(repo, contractRepo, assignmentRepo, validator, zaptest.NewLogger(t), stubIDGen)
	err := svc.Delete(context.Background(), "d1")
	assert.ErrorIs(t, err, domain.ErrDriverHasActiveAssignments)
}

func TestService_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockDriverRepository(ctrl)
	contractRepo := mocks.NewMockContractRepository(ctrl)
	assignmentRepo := mocks.NewMockVehicleAssignmentRepository(ctrl)

	contractRepo.EXPECT().FindByDriverID(gomock.Any(), "d1").Return([]*domain.Contract{}, nil)
	assignmentRepo.EXPECT().FindActiveByDriverID(gomock.Any(), "d1").Return(nil, nil)
	repo.EXPECT().SoftDelete(gomock.Any(), "d1").Return(nil)

	validator := mocks.NewMockDriverLicenseValidator(ctrl)
	svc := driver.New(repo, contractRepo, assignmentRepo, validator, zaptest.NewLogger(t), stubIDGen)
	err := svc.Delete(context.Background(), "d1")
	require.NoError(t, err)
}
