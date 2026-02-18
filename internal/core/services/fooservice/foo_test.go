package fooservice_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports/mocks"
	"github.com/albenik/uber-fx-based-service-example/internal/core/services/fooservice"
)

var _ ports.FooEntityService = (*fooservice.Service)(nil)

func stubIDGen() string { return "test-id" }

func TestService_CreateEntity(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFooEntityRepository(ctrl)

	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

	svc := fooservice.New(repo, zaptest.NewLogger(t), stubIDGen)
	entity, err := svc.CreateEntity(t.Context(), "foo", "some description")
	require.NoError(t, err)
	assert.Equal(t, "test-id", entity.ID)
	assert.Equal(t, "foo", entity.Name)
	assert.Equal(t, "some description", entity.Description)
}

func TestService_CreateEntity_EmptyName(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFooEntityRepository(ctrl)

	svc := fooservice.New(repo, zaptest.NewLogger(t), stubIDGen)
	_, err := svc.CreateEntity(t.Context(), "", "some description")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.ErrorContains(t, err, "name is required")
}

func TestService_CreateEntity_EmptyDescription(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFooEntityRepository(ctrl)

	svc := fooservice.New(repo, zaptest.NewLogger(t), stubIDGen)
	_, err := svc.CreateEntity(t.Context(), "foo", "")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.ErrorContains(t, err, "description is required")
}

func TestService_CreateEntity_WhitespaceOnlyName(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFooEntityRepository(ctrl)

	svc := fooservice.New(repo, zaptest.NewLogger(t), stubIDGen)
	_, err := svc.CreateEntity(t.Context(), "   ", "some description")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.ErrorContains(t, err, "name is required")
}

func TestService_CreateEntity_WhitespaceOnlyDescription(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFooEntityRepository(ctrl)

	svc := fooservice.New(repo, zaptest.NewLogger(t), stubIDGen)
	_, err := svc.CreateEntity(t.Context(), "foo", "   ")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.ErrorContains(t, err, "description is required")
}

func TestService_CreateEntity_EmptyIDGenerator(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFooEntityRepository(ctrl)

	emptyIDGen := func() string { return "" }

	svc := fooservice.New(repo, zaptest.NewLogger(t), emptyIDGen)
	_, err := svc.CreateEntity(t.Context(), "foo", "some description")
	require.Error(t, err)
	assert.ErrorContains(t, err, "id generator returned empty ID")
}

func TestService_CreateEntity_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFooEntityRepository(ctrl)

	repoErr := errors.New("save failed")
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(repoErr)

	svc := fooservice.New(repo, zaptest.NewLogger(t), stubIDGen)
	_, err := svc.CreateEntity(t.Context(), "foo", "some description")
	assert.ErrorIs(t, err, repoErr)
}

func TestService_GetEntity(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFooEntityRepository(ctrl)

	expected := &domain.FooEntity{ID: "1", Name: "foo", Description: "bar"}
	repo.EXPECT().FindByID(gomock.Any(), "1").Return(expected, nil)

	svc := fooservice.New(repo, zaptest.NewLogger(t), stubIDGen)
	entity, err := svc.GetEntity(t.Context(), "1")
	require.NoError(t, err)
	assert.Equal(t, expected, entity)
}

func TestService_GetEntity_EmptyID(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFooEntityRepository(ctrl)

	svc := fooservice.New(repo, zaptest.NewLogger(t), stubIDGen)
	_, err := svc.GetEntity(t.Context(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.ErrorContains(t, err, "id is required")
}

func TestService_GetEntity_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFooEntityRepository(ctrl)

	repo.EXPECT().FindByID(gomock.Any(), "1").Return(nil, domain.ErrEntityNotFound)

	svc := fooservice.New(repo, zaptest.NewLogger(t), stubIDGen)
	_, err := svc.GetEntity(t.Context(), "1")
	assert.ErrorIs(t, err, domain.ErrEntityNotFound)
}

func TestService_ListEntities(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFooEntityRepository(ctrl)

	expected := []*domain.FooEntity{
		{ID: "1", Name: "foo", Description: "bar"},
		{ID: "2", Name: "baz", Description: "qux"},
	}
	repo.EXPECT().FindAll(gomock.Any()).Return(expected, nil)

	svc := fooservice.New(repo, zaptest.NewLogger(t), stubIDGen)
	entities, err := svc.ListEntities(t.Context())
	require.NoError(t, err)
	assert.Equal(t, expected, entities)
}

func TestService_ListEntities_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFooEntityRepository(ctrl)

	repo.EXPECT().FindAll(gomock.Any()).Return([]*domain.FooEntity{}, nil)

	svc := fooservice.New(repo, zaptest.NewLogger(t), stubIDGen)
	entities, err := svc.ListEntities(t.Context())
	require.NoError(t, err)
	assert.Empty(t, entities)
}

func TestService_ListEntities_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFooEntityRepository(ctrl)

	repoErr := errors.New("find all failed")
	repo.EXPECT().FindAll(gomock.Any()).Return(nil, repoErr)

	svc := fooservice.New(repo, zaptest.NewLogger(t), stubIDGen)
	_, err := svc.ListEntities(t.Context())
	assert.ErrorIs(t, err, repoErr)
}

func TestService_DeleteEntity(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFooEntityRepository(ctrl)

	repo.EXPECT().Delete(gomock.Any(), "1").Return(nil)

	svc := fooservice.New(repo, zaptest.NewLogger(t), stubIDGen)
	err := svc.DeleteEntity(t.Context(), "1")
	assert.NoError(t, err)
}

func TestService_DeleteEntity_EmptyID(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFooEntityRepository(ctrl)

	svc := fooservice.New(repo, zaptest.NewLogger(t), stubIDGen)
	err := svc.DeleteEntity(t.Context(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.ErrorContains(t, err, "id is required")
}

func TestService_DeleteEntity_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockFooEntityRepository(ctrl)

	repo.EXPECT().Delete(gomock.Any(), "1").Return(domain.ErrEntityNotFound)

	svc := fooservice.New(repo, zaptest.NewLogger(t), stubIDGen)
	err := svc.DeleteEntity(t.Context(), "1")
	assert.ErrorIs(t, err, domain.ErrEntityNotFound)
}
