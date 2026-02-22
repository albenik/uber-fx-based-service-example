package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

type FleetHandler struct {
	svc    ports.FleetService
	logger *zap.Logger
}

func NewFleetHandler(svc ports.FleetService, logger *zap.Logger) *FleetHandler {
	return &FleetHandler{svc: svc, logger: logger}
}

func (h *FleetHandler) RegisterRoutes(r chi.Router) {
	r.Route("/legal-entities/{legalEntityId}/fleets", func(r chi.Router) {
		r.Get("/", h.listByLegalEntity)
		r.Post("/", h.create)
	})
	r.Route("/fleets", func(r chi.Router) {
		r.Get("/{id}", h.get)
		r.Delete("/{id}", h.delete)
		r.Post("/{id}/undelete", h.undelete)
	})
}

type createFleetRequest struct {
	Name string `json:"name"`
}

type fleetResponse struct {
	ID             string `json:"id"`
	LegalEntityID  string `json:"legal_entity_id"`
	Name           string `json:"name"`
}

func (h *FleetHandler) create(w http.ResponseWriter, r *http.Request) {
	if !requireJSON(w, r) {
		return
	}
	var req createFleetRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	legalEntityID := chi.URLParam(r, "legalEntityId")
	entity, err := h.svc.Create(r.Context(), legalEntityID, req.Name)
	if err != nil {
		h.handleError(w, "create fleet", err)
		return
	}
	respondJSON(w, http.StatusCreated, fleetResponse{ID: entity.ID, LegalEntityID: entity.LegalEntityID, Name: entity.Name})
}

func (h *FleetHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entity, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, "get fleet", err)
		return
	}
	respondJSON(w, http.StatusOK, fleetResponse{ID: entity.ID, LegalEntityID: entity.LegalEntityID, Name: entity.Name})
}

func (h *FleetHandler) listByLegalEntity(w http.ResponseWriter, r *http.Request) {
	legalEntityID := chi.URLParam(r, "legalEntityId")
	entities, err := h.svc.ListByLegalEntity(r.Context(), legalEntityID)
	if err != nil {
		h.handleError(w, "list fleets", err)
		return
	}
	resp := make([]fleetResponse, 0, len(entities))
	for _, e := range entities {
		resp = append(resp, fleetResponse{ID: e.ID, LegalEntityID: e.LegalEntityID, Name: e.Name})
	}
	respondJSON(w, http.StatusOK, resp)
}

func (h *FleetHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleError(w, "delete fleet", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FleetHandler) undelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Undelete(r.Context(), id); err != nil {
		h.handleError(w, "undelete fleet", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FleetHandler) handleError(w http.ResponseWriter, op string, err error) {
	if domain.IsExposable(err) {
		http.Error(w, err.Error(), mapDomainErrorToStatus(err))
		return
	}
	h.logger.Error("fleet operation failed", zap.String("op", op), zap.Error(err))
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
