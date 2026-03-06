package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

// FeatureToggleExampleHandler demonstrates internal usage of the
// FeatureToggleProvider port. Application code injects the provider wherever
// behaviour must change at runtime based on a toggle value.
type FeatureToggleExampleHandler struct {
	toggles ports.FeatureToggleProvider
	logger  *zap.Logger
}

func NewFeatureToggleExampleHandler(toggles ports.FeatureToggleProvider, logger *zap.Logger) *FeatureToggleExampleHandler {
	return &FeatureToggleExampleHandler{toggles: toggles, logger: logger}
}

func (h *FeatureToggleExampleHandler) RegisterRoutes(r chi.Router) {
	r.Get("/feature-toggle-example", h.example)
}

func (h *FeatureToggleExampleHandler) example(w http.ResponseWriter, r *http.Request) {
	enabled, err := h.toggles.IsEnabled(r.Context(), "foo")
	if err != nil {
		h.logger.Error("feature toggle check failed", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if enabled {
		_, _ = w.Write([]byte("FOO"))
	} else {
		_, _ = w.Write([]byte("bar"))
	}
}
