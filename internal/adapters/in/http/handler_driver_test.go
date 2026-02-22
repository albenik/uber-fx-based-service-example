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

func setupDriverHandler(t *testing.T) (*mocks.MockDriverService, chi.Router) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockDriverService(ctrl)
	handler := httpAdapter.NewDriverHandler(mockSvc, zaptest.NewLogger(t))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return mockSvc, r
}

func TestDriverHandler_Delete_RejectsActiveContracts(t *testing.T) {
	mockSvc, router := setupDriverHandler(t)

	mockSvc.EXPECT().Delete(gomock.Any(), "d1").Return(domain.ErrDriverHasActiveContracts)

	req := httptest.NewRequest(http.MethodDelete, "/drivers/d1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "active contracts")
}

func TestDriverHandler_Delete_RejectsActiveAssignments(t *testing.T) {
	mockSvc, router := setupDriverHandler(t)

	mockSvc.EXPECT().Delete(gomock.Any(), "d1").Return(domain.ErrDriverHasActiveAssignments)

	req := httptest.NewRequest(http.MethodDelete, "/drivers/d1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestDriverHandler_Create_Success(t *testing.T) {
	mockSvc, router := setupDriverHandler(t)

	entity := &domain.Driver{ID: "d1", FirstName: "John", LastName: "Doe", LicenseNumber: "DL-123"}
	mockSvc.EXPECT().Create(gomock.Any(), "John", "Doe", "DL-123").Return(entity, nil)

	body, _ := json.Marshal(map[string]string{"first_name": "John", "last_name": "Doe", "license_number": "DL-123"})
	req := httptest.NewRequest(http.MethodPost, "/drivers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "d1", resp["id"])
	assert.Equal(t, "John", resp["first_name"])
}
