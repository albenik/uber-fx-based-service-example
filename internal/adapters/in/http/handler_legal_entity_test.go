package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func setupLegalEntityHandler(t *testing.T) (*mocks.MockLegalEntityService, chi.Router) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockLegalEntityService(ctrl)
	handler := httpAdapter.NewLegalEntityHandler(mockSvc, zaptest.NewLogger(t))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return mockSvc, r
}

func TestLegalEntityHandler_Create_Success(t *testing.T) {
	mockSvc, router := setupLegalEntityHandler(t)

	entity := &domain.LegalEntity{ID: "1", Name: "Acme", TaxID: "123"}
	mockSvc.EXPECT().Create(gomock.Any(), "Acme", "123").Return(entity, nil)

	body, _ := json.Marshal(map[string]string{"name": "Acme", "tax_id": "123"})
	req := httptest.NewRequest(http.MethodPost, "/legal-entities", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "1", resp["id"])
	assert.Equal(t, "Acme", resp["name"])
	assert.Equal(t, "123", resp["tax_id"])
}

func TestLegalEntityHandler_Create_InvalidInput(t *testing.T) {
	mockSvc, router := setupLegalEntityHandler(t)

	mockSvc.EXPECT().Create(gomock.Any(), "", "123").Return(nil, domain.ErrInvalidInput)

	body, _ := json.Marshal(map[string]string{"tax_id": "123"})
	req := httptest.NewRequest(http.MethodPost, "/legal-entities", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLegalEntityHandler_Get_Success(t *testing.T) {
	mockSvc, router := setupLegalEntityHandler(t)

	entity := &domain.LegalEntity{ID: "1", Name: "Acme", TaxID: "123"}
	mockSvc.EXPECT().Get(gomock.Any(), "1").Return(entity, nil)

	req := httptest.NewRequest(http.MethodGet, "/legal-entities/1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "1", resp["id"])
}

func TestLegalEntityHandler_Get_NotFound(t *testing.T) {
	mockSvc, router := setupLegalEntityHandler(t)

	mockSvc.EXPECT().Get(gomock.Any(), "1").Return(nil, domain.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/legal-entities/1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestLegalEntityHandler_Delete_Success(t *testing.T) {
	mockSvc, router := setupLegalEntityHandler(t)

	mockSvc.EXPECT().Delete(gomock.Any(), "1").Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/legal-entities/1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}
