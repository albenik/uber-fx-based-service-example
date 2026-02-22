package driver_test

import (
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
	svc := driver.New(repo, contractRepo, assignmentRepo, validator, stubIDGen, time.Now, zaptest.NewLogger(t))
	err := svc.Delete(t.Context(), "d1")
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
	svc := driver.New(repo, contractRepo, assignmentRepo, validator, stubIDGen, time.Now, zaptest.NewLogger(t))
	err := svc.Delete(t.Context(), "d1")
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
	svc := driver.New(repo, contractRepo, assignmentRepo, validator, stubIDGen, time.Now, zaptest.NewLogger(t))
	err := svc.Delete(t.Context(), "d1")
	require.NoError(t, err)
}

func TestService_Create_SuccessWhenValidationOk(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockDriverRepository(ctrl)
	contractRepo := mocks.NewMockContractRepository(ctrl)
	assignmentRepo := mocks.NewMockVehicleAssignmentRepository(ctrl)
	validator := mocks.NewMockDriverLicenseValidator(ctrl)

	validator.EXPECT().ValidateLicense(gomock.Any(), "John", "Doe", "DL123").Return(domain.LicenseValid, nil)
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

	svc := driver.New(repo, contractRepo, assignmentRepo, validator, stubIDGen, time.Now, zaptest.NewLogger(t))
	entity, err := svc.Create(t.Context(), "John", "Doe", "DL123")
	require.NoError(t, err)
	assert.Equal(t, "test-id", entity.ID)
	assert.Equal(t, "John", entity.FirstName)
	assert.Equal(t, "Doe", entity.LastName)
	assert.Equal(t, "DL123", entity.LicenseNumber)
}

func TestService_Create_FailsWhenValidationNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockDriverRepository(ctrl)
	contractRepo := mocks.NewMockContractRepository(ctrl)
	assignmentRepo := mocks.NewMockVehicleAssignmentRepository(ctrl)
	validator := mocks.NewMockDriverLicenseValidator(ctrl)

	validator.EXPECT().ValidateLicense(gomock.Any(), "John", "Doe", "DL999").Return(domain.LicenseNotFound, nil)

	svc := driver.New(repo, contractRepo, assignmentRepo, validator, stubIDGen, time.Now, zaptest.NewLogger(t))
	_, err := svc.Create(t.Context(), "John", "Doe", "DL999")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrLicenseValidationFailed)
	assert.Contains(t, err.Error(), "not_found")
}

func TestService_Create_FailsWhenValidatorReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockDriverRepository(ctrl)
	contractRepo := mocks.NewMockContractRepository(ctrl)
	assignmentRepo := mocks.NewMockVehicleAssignmentRepository(ctrl)
	validator := mocks.NewMockDriverLicenseValidator(ctrl)

	validator.EXPECT().ValidateLicense(gomock.Any(), "John", "Doe", "DL123").Return(domain.LicenseValidationResult(""), domain.ErrValidationServiceUnavailable)

	svc := driver.New(repo, contractRepo, assignmentRepo, validator, stubIDGen, time.Now, zaptest.NewLogger(t))
	_, err := svc.Create(t.Context(), "John", "Doe", "DL123")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrValidationServiceUnavailable)
}
