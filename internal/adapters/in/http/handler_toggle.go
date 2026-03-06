package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

type ToggleHandler struct {
	svc    ports.FeatureToggleService
	logger *zap.Logger
}

func NewToggleHandler(svc ports.FeatureToggleService, logger *zap.Logger) *ToggleHandler {
	return &ToggleHandler{svc: svc, logger: logger}
}

func (h *ToggleHandler) RegisterRoutes(r chi.Router) {
	r.Route("/toggles", func(r chi.Router) {
		r.Get("/", h.list)
		r.Get("/{name}", h.get)
		r.Get("/{name}/enabled", h.isEnabled)
	})
}

type toggleResponse struct {
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

type isEnabledResponse struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func (h *ToggleHandler) list(w http.ResponseWriter, r *http.Request) {
	toggles, err := h.svc.ListToggles(r.Context())
	if err != nil {
		h.handleError(w, "list toggles", err)
		return
	}
	resp := make([]toggleResponse, 0, len(toggles))
	for _, t := range toggles {
		resp = append(resp, toggleResponse{Name: t.Name, Enabled: t.Enabled, Description: t.Description})
	}
	respondJSON(w, http.StatusOK, resp)
}

func (h *ToggleHandler) get(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	toggle, err := h.svc.GetToggle(r.Context(), name)
	if err != nil {
		h.handleError(w, "get toggle", err)
		return
	}
	respondJSON(w, http.StatusOK, toggleResponse{Name: toggle.Name, Enabled: toggle.Enabled, Description: toggle.Description})
}

func (h *ToggleHandler) isEnabled(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	enabled, err := h.svc.IsEnabled(r.Context(), name)
	if err != nil {
		h.handleError(w, "check toggle", err)
		return
	}
	respondJSON(w, http.StatusOK, isEnabledResponse{Name: name, Enabled: enabled})
}

func (h *ToggleHandler) handleError(w http.ResponseWriter, op string, err error) {
	if domain.IsExposable(err) {
		http.Error(w, err.Error(), mapDomainErrorToStatus(err))
		return
	}
	h.logger.Error("toggle operation failed", zap.String("op", op), zap.Error(err))
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
