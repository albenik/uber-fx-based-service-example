package http_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	httpAdapter "github.com/albenik/uber-fx-based-service-example/internal/adapters/in/http"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports/mocks"
)

func setupFeatureToggleExampleHandler(t *testing.T) (*mocks.MockFeatureToggleProvider, chi.Router) {
	ctrl := gomock.NewController(t)
	mockProvider := mocks.NewMockFeatureToggleProvider(ctrl)
	handler := httpAdapter.NewFeatureToggleExampleHandler(mockProvider, zaptest.NewLogger(t))
	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return mockProvider, r
}

func TestFeatureToggleExample_FooEnabled(t *testing.T) {
	mockProvider, router := setupFeatureToggleExampleHandler(t)
	mockProvider.EXPECT().IsEnabled(gomock.Any(), "foo").Return(true, nil)

	req := httptest.NewRequest(http.MethodGet, "/feature-toggle-example", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "FOO", rec.Body.String())
}

func TestFeatureToggleExample_FooDisabled(t *testing.T) {
	mockProvider, router := setupFeatureToggleExampleHandler(t)
	mockProvider.EXPECT().IsEnabled(gomock.Any(), "foo").Return(false, nil)

	req := httptest.NewRequest(http.MethodGet, "/feature-toggle-example", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "bar", rec.Body.String())
}

func TestFeatureToggleExample_ProviderError(t *testing.T) {
	mockProvider, router := setupFeatureToggleExampleHandler(t)
	mockProvider.EXPECT().IsEnabled(gomock.Any(), "foo").Return(false, errors.New("redis down"))

	req := httptest.NewRequest(http.MethodGet, "/feature-toggle-example", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
