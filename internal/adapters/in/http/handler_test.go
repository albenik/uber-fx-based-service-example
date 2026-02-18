package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	httpAdapter "github.com/albenik/uber-fx-based-service-example/internal/adapters/in/http"
	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports/mocks"
)

func setupHandler(t *testing.T) (*mocks.MockFooEntityService, chi.Router) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockFooEntityService(ctrl)
	handler := httpAdapter.NewFooEntityHandler(mockSvc, zaptest.NewLogger(t))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return mockSvc, r
}

func TestCreateFooEntity_Success(t *testing.T) {
	mockSvc, router := setupHandler(t)

	entity := &domain.FooEntity{ID: "1", Name: "foo", Description: "bar"}
	mockSvc.EXPECT().CreateEntity(gomock.Any(), "foo", "bar").Return(entity, nil)

	body, _ := json.Marshal(map[string]string{"name": "foo", "description": "bar"})
	req := httptest.NewRequest(http.MethodPost, "/foos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "1", resp["id"])
	assert.Equal(t, "foo", resp["name"])
	assert.Equal(t, "bar", resp["description"])
}

func TestCreateFooEntity_BadJSON(t *testing.T) {
	_, router := setupHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/foos", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateFooEntity_MissingName(t *testing.T) {
	mockSvc, router := setupHandler(t)

	mockSvc.EXPECT().CreateEntity(gomock.Any(), "", "bar").
		Return(nil, domain.ErrInvalidInput)

	body, _ := json.Marshal(map[string]string{"description": "bar"})
	req := httptest.NewRequest(http.MethodPost, "/foos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateFooEntity_MissingDescription(t *testing.T) {
	mockSvc, router := setupHandler(t)

	mockSvc.EXPECT().CreateEntity(gomock.Any(), "foo", "").
		Return(nil, domain.ErrInvalidInput)

	body, _ := json.Marshal(map[string]string{"name": "foo"})
	req := httptest.NewRequest(http.MethodPost, "/foos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateFooEntity_WhitespaceOnlyFields(t *testing.T) {
	mockSvc, router := setupHandler(t)

	mockSvc.EXPECT().CreateEntity(gomock.Any(), "   ", "   ").
		Return(nil, domain.ErrInvalidInput)

	body, _ := json.Marshal(map[string]string{"name": "   ", "description": "   "})
	req := httptest.NewRequest(http.MethodPost, "/foos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateFooEntity_UnknownFields(t *testing.T) {
	_, router := setupHandler(t)

	body, _ := json.Marshal(map[string]string{"name": "foo", "description": "bar", "extra": "field"})
	req := httptest.NewRequest(http.MethodPost, "/foos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateFooEntity_UnsupportedMediaType(t *testing.T) {
	_, router := setupHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/foos", strings.NewReader(`{"name":"foo"}`))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
}

func TestCreateFooEntity_BodyTooLarge(t *testing.T) {
	_, router := setupHandler(t)

	body := strings.NewReader(`{"name":"foo","description":"bar"}`)
	req := httptest.NewRequest(http.MethodPost, "/foos", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	req.Body = http.MaxBytesReader(rec, req.Body, 1) // 1-byte limit to trigger MaxBytesError
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
}

func TestCreateFooEntity_TrailingData(t *testing.T) {
	_, router := setupHandler(t)

	body := strings.NewReader(`{"name":"foo","description":"bar"}{"extra":true}`)
	req := httptest.NewRequest(http.MethodPost, "/foos", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateFooEntity_MissingContentType(t *testing.T) {
	_, router := setupHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/foos", strings.NewReader(`{"name":"foo"}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
}

func TestCreateFooEntity_ServiceError(t *testing.T) {
	mockSvc, router := setupHandler(t)

	mockSvc.EXPECT().CreateEntity(gomock.Any(), "foo", "bar").Return(nil, errors.New("service error"))

	body, _ := json.Marshal(map[string]string{"name": "foo", "description": "bar"})
	req := httptest.NewRequest(http.MethodPost, "/foos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetFooEntity_Success(t *testing.T) {
	mockSvc, router := setupHandler(t)

	entity := &domain.FooEntity{ID: "1", Name: "foo", Description: "bar"}
	mockSvc.EXPECT().GetEntity(gomock.Any(), "1").Return(entity, nil)

	req := httptest.NewRequest(http.MethodGet, "/foos/1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "1", resp["id"])
	assert.Equal(t, "foo", resp["name"])
	assert.Equal(t, "bar", resp["description"])
}

func TestGetFooEntity_NotFound(t *testing.T) {
	mockSvc, router := setupHandler(t)

	mockSvc.EXPECT().GetEntity(gomock.Any(), "1").Return(nil, domain.ErrEntityNotFound)

	req := httptest.NewRequest(http.MethodGet, "/foos/1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetFooEntity_Error(t *testing.T) {
	mockSvc, router := setupHandler(t)

	mockSvc.EXPECT().GetEntity(gomock.Any(), "1").Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/foos/1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestListFooEntities_Success(t *testing.T) {
	mockSvc, router := setupHandler(t)

	entities := []*domain.FooEntity{
		{ID: "1", Name: "foo", Description: "bar"},
		{ID: "2", Name: "baz", Description: "qux"},
	}
	mockSvc.EXPECT().ListEntities(gomock.Any()).Return(entities, nil)

	req := httptest.NewRequest(http.MethodGet, "/foos", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp []map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
}

func TestListFooEntities_Empty(t *testing.T) {
	mockSvc, router := setupHandler(t)

	mockSvc.EXPECT().ListEntities(gomock.Any()).Return([]*domain.FooEntity{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/foos", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "[]\n", rec.Body.String())
}

func TestListFooEntities_Error(t *testing.T) {
	mockSvc, router := setupHandler(t)

	mockSvc.EXPECT().ListEntities(gomock.Any()).Return(nil, errors.New("list error"))

	req := httptest.NewRequest(http.MethodGet, "/foos", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDeleteFooEntity_Success(t *testing.T) {
	mockSvc, router := setupHandler(t)

	mockSvc.EXPECT().DeleteEntity(gomock.Any(), "1").Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/foos/1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteFooEntity_NotFound(t *testing.T) {
	mockSvc, router := setupHandler(t)

	mockSvc.EXPECT().DeleteEntity(gomock.Any(), "1").Return(domain.ErrEntityNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/foos/1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDeleteFooEntity_Error(t *testing.T) {
	mockSvc, router := setupHandler(t)

	mockSvc.EXPECT().DeleteEntity(gomock.Any(), "1").Return(errors.New("delete error"))

	req := httptest.NewRequest(http.MethodDelete, "/foos/1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
