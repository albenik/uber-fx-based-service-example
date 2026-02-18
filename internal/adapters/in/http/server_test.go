package http_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"

	httpAdapter "github.com/albenik/uber-fx-based-service-example/internal/adapters/in/http"
	"github.com/albenik/uber-fx-based-service-example/internal/config"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports/mocks"
)

func TestNewServer_UsesConfigAddr(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockFooEntityService(ctrl)
	handler := httpAdapter.NewFooEntityHandler(mockSvc, zaptest.NewLogger(t))

	srv := httpAdapter.NewServer(&config.HTTPServerConfig{Addr: ":9090"}, handler)
	assert.Equal(t, ":9090", srv.Addr)
}

func TestNewServer_HealthCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockFooEntityService(ctrl)
	handler := httpAdapter.NewFooEntityHandler(mockSvc, zaptest.NewLogger(t))

	srv := httpAdapter.NewServer(&config.HTTPServerConfig{Addr: ":8080"}, handler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/plain; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Equal(t, "ok", rec.Body.String())
}

func TestNewServer_BodyTooLarge(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := mocks.NewMockFooEntityService(ctrl)
	handler := httpAdapter.NewFooEntityHandler(mockSvc, zaptest.NewLogger(t))

	srv := httpAdapter.NewServer(&config.HTTPServerConfig{Addr: ":8080"}, handler)

	largeBody := `{"name":"` + strings.Repeat("x", 2<<20) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/foos", strings.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
}
