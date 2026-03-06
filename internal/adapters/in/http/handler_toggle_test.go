package http_test

import (
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

func setupToggleHandler(t *testing.T) (*mocks.MockFeatureToggleService, chi.Router) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockFeatureToggleService(ctrl)
	handler := httpAdapter.NewToggleHandler(mockSvc, zaptest.NewLogger(t))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return mockSvc, r
}

func TestToggleHandler_List_Success(t *testing.T) {
	mockSvc, router := setupToggleHandler(t)

	toggles := []*domain.FeatureToggle{
		{Name: "dark-mode", Enabled: true, Description: "Dark mode"},
		{Name: "beta", Enabled: false, Description: "Beta feature"},
	}
	mockSvc.EXPECT().ListToggles(gomock.Any()).Return(toggles, nil)

	req := httptest.NewRequest(http.MethodGet, "/toggles", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp, 2)
	assert.Equal(t, "dark-mode", resp[0]["name"])
	assert.Equal(t, true, resp[0]["enabled"])
}

func TestToggleHandler_List_ProviderUnavailable(t *testing.T) {
	mockSvc, router := setupToggleHandler(t)

	mockSvc.EXPECT().ListToggles(gomock.Any()).Return(nil, domain.ErrToggleProviderUnavailable)

	req := httptest.NewRequest(http.MethodGet, "/toggles", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestToggleHandler_Get_Success(t *testing.T) {
	mockSvc, router := setupToggleHandler(t)

	toggle := &domain.FeatureToggle{Name: "dark-mode", Enabled: true, Description: "Dark mode"}
	mockSvc.EXPECT().GetToggle(gomock.Any(), "dark-mode").Return(toggle, nil)

	req := httptest.NewRequest(http.MethodGet, "/toggles/dark-mode", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "dark-mode", resp["name"])
	assert.Equal(t, true, resp["enabled"])
}

func TestToggleHandler_Get_NotFound(t *testing.T) {
	mockSvc, router := setupToggleHandler(t)

	mockSvc.EXPECT().GetToggle(gomock.Any(), "nonexistent").Return(nil, domain.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/toggles/nonexistent", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestToggleHandler_IsEnabled_True(t *testing.T) {
	mockSvc, router := setupToggleHandler(t)

	mockSvc.EXPECT().IsEnabled(gomock.Any(), "dark-mode").Return(true, nil)

	req := httptest.NewRequest(http.MethodGet, "/toggles/dark-mode/enabled", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "dark-mode", resp["name"])
	assert.Equal(t, true, resp["enabled"])
}

func TestToggleHandler_IsEnabled_False(t *testing.T) {
	mockSvc, router := setupToggleHandler(t)

	mockSvc.EXPECT().IsEnabled(gomock.Any(), "beta").Return(false, nil)

	req := httptest.NewRequest(http.MethodGet, "/toggles/beta/enabled", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "beta", resp["name"])
	assert.Equal(t, false, resp["enabled"])
}
