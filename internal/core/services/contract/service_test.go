package contract_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports/mocks"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services/contract"
)

func stubIDGen() string { return "test-id" }

func TestService_Create_RejectsOverlap(t *testing.T) {
	ctrl := gomock.NewController(t)
	driverRepo := mocks.NewMockDriverRepository(ctrl)
	legalRepo := mocks.NewMockLegalEntityRepository(ctrl)
	fleetRepo := mocks.NewMockFleetRepository(ctrl)
	contractRepo := mocks.NewMockContractRepository(ctrl)

	driverRepo.EXPECT().FindByID(gomock.Any(), "d1").Return(&domain.Driver{ID: "d1"}, nil)
	legalRepo.EXPECT().FindByID(gomock.Any(), "le1").Return(&domain.LegalEntity{ID: "le1"}, nil)
	fleetRepo.EXPECT().FindByID(gomock.Any(), "f1").Return(&domain.Fleet{ID: "f1"}, nil)
	contractRepo.EXPECT().FindOverlapping(gomock.Any(), "d1", "le1", "f1",
		gomock.Any(), gomock.Any(), "").Return([]*domain.Contract{{ID: "existing"}}, nil)

	svc := contract.New(driverRepo, legalRepo, fleetRepo, contractRepo, zaptest.NewLogger(t), stubIDGen)
	start := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC)
	_, err := svc.Create(t.Context(), "d1", "le1", "f1", start, end)
	assert.ErrorIs(t, err, domain.ErrConflict)
}

func TestService_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	driverRepo := mocks.NewMockDriverRepository(ctrl)
	legalRepo := mocks.NewMockLegalEntityRepository(ctrl)
	fleetRepo := mocks.NewMockFleetRepository(ctrl)
	contractRepo := mocks.NewMockContractRepository(ctrl)

	driverRepo.EXPECT().FindByID(gomock.Any(), "d1").Return(&domain.Driver{ID: "d1"}, nil)
	legalRepo.EXPECT().FindByID(gomock.Any(), "le1").Return(&domain.LegalEntity{ID: "le1"}, nil)
	fleetRepo.EXPECT().FindByID(gomock.Any(), "f1").Return(&domain.Fleet{ID: "f1"}, nil)
	contractRepo.EXPECT().FindOverlapping(gomock.Any(), "d1", "le1", "f1", gomock.Any(), gomock.Any(), "").Return(nil, nil)
	contractRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

	svc := contract.New(driverRepo, legalRepo, fleetRepo, contractRepo, zaptest.NewLogger(t), stubIDGen)
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	entity, err := svc.Create(t.Context(), "d1", "le1", "f1", start, end)
	require.NoError(t, err)
	assert.Equal(t, "test-id", entity.ID)
	assert.Equal(t, "d1", entity.DriverID)
}
