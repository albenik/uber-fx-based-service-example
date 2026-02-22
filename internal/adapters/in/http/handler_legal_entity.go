package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

type LegalEntityHandler struct {
	svc    ports.LegalEntityService
	logger *zap.Logger
}

func NewLegalEntityHandler(svc ports.LegalEntityService, logger *zap.Logger) *LegalEntityHandler {
	return &LegalEntityHandler{svc: svc, logger: logger}
}

func (h *LegalEntityHandler) RegisterRoutes(r chi.Router) {
	r.Route("/legal-entities", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{id}", h.get)
		r.Delete("/{id}", h.delete)
		r.Post("/{id}/undelete", h.undelete)
	})
}

type createLegalEntityRequest struct {
	Name  string `json:"name"`
	TaxID string `json:"tax_id"`
}

type legalEntityResponse struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	TaxID string  `json:"tax_id"`
}

func (h *LegalEntityHandler) create(w http.ResponseWriter, r *http.Request) {
	if !requireJSON(w, r) {
		return
	}
	var req createLegalEntityRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	entity, err := h.svc.Create(r.Context(), req.Name, req.TaxID)
	if err != nil {
		h.handleError(w, "create legal entity", err)
		return
	}
	respondJSON(w, http.StatusCreated, legalEntityResponse{ID: entity.ID, Name: entity.Name, TaxID: entity.TaxID})
}

func (h *LegalEntityHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entity, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, "get legal entity", err)
		return
	}
	respondJSON(w, http.StatusOK, legalEntityResponse{ID: entity.ID, Name: entity.Name, TaxID: entity.TaxID})
}

func (h *LegalEntityHandler) list(w http.ResponseWriter, r *http.Request) {
	entities, err := h.svc.List(r.Context())
	if err != nil {
		h.handleError(w, "list legal entities", err)
		return
	}
	resp := make([]legalEntityResponse, 0, len(entities))
	for _, e := range entities {
		resp = append(resp, legalEntityResponse{ID: e.ID, Name: e.Name, TaxID: e.TaxID})
	}
	respondJSON(w, http.StatusOK, resp)
}

func (h *LegalEntityHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleError(w, "delete legal entity", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *LegalEntityHandler) undelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Undelete(r.Context(), id); err != nil {
		h.handleError(w, "undelete legal entity", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *LegalEntityHandler) handleError(w http.ResponseWriter, op string, err error) {
	if domain.IsExposable(err) {
		status := mapDomainErrorToStatus(err)
		http.Error(w, err.Error(), status)
		return
	}
	h.logger.Error("legal entity operation failed", zap.String("op", op), zap.Error(err))
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
