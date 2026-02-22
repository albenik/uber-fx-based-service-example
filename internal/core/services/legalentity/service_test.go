package legalentity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports/mocks"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services/legalentity"
)

func stubIDGen() string { return "test-id" }

func TestService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockLegalEntityRepository(ctrl)

	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

	svc := legalentity.New(repo, zaptest.NewLogger(t), stubIDGen)
	entity, err := svc.Create(t.Context(), "Acme", "123")
	require.NoError(t, err)
	assert.Equal(t, "test-id", entity.ID)
	assert.Equal(t, "Acme", entity.Name)
	assert.Equal(t, "123", entity.TaxID)
}

func TestService_Create_EmptyName(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockLegalEntityRepository(ctrl)

	svc := legalentity.New(repo, zaptest.NewLogger(t), stubIDGen)
	_, err := svc.Create(t.Context(), "", "123")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestService_Create_EmptyTaxID(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockLegalEntityRepository(ctrl)

	svc := legalentity.New(repo, zaptest.NewLogger(t), stubIDGen)
	_, err := svc.Create(t.Context(), "Acme", "")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}
