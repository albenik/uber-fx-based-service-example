package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

type ContractHandler struct {
	svc    ports.ContractService
	logger *zap.Logger
}

func NewContractHandler(svc ports.ContractService, logger *zap.Logger) *ContractHandler {
	return &ContractHandler{svc: svc, logger: logger}
}

func (h *ContractHandler) RegisterRoutes(r chi.Router) {
	r.Route("/drivers/{driverId}/contracts", func(r chi.Router) {
		r.Get("/", h.listByDriver)
		r.Post("/", h.create)
	})
	r.Route("/contracts", func(r chi.Router) {
		r.Get("/{id}", h.get)
		r.Post("/{id}/terminate", h.terminate)
		r.Delete("/{id}", h.delete)
		r.Post("/{id}/undelete", h.undelete)
	})
}

type createContractRequest struct {
	LegalEntityID string `json:"legal_entity_id"`
	FleetID       string `json:"fleet_id"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
}

type terminateContractRequest struct {
	TerminatedBy string `json:"terminated_by"`
}

type contractResponse struct {
	ID            string  `json:"id"`
	DriverID      string  `json:"driver_id"`
	LegalEntityID string  `json:"legal_entity_id"`
	FleetID       string  `json:"fleet_id"`
	StartDate     string  `json:"start_date"`
	EndDate       string  `json:"end_date"`
	TerminatedAt  *string `json:"terminated_at,omitempty"`
	TerminatedBy  string  `json:"terminated_by,omitempty"`
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func (h *ContractHandler) create(w http.ResponseWriter, r *http.Request) {
	if !requireJSON(w, r) {
		return
	}
	var req createContractRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	driverID := chi.URLParam(r, "driverId")
	startDate, err := parseDate(req.StartDate)
	if err != nil {
		http.Error(w, "invalid start_date format (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	endDate, err := parseDate(req.EndDate)
	if err != nil {
		http.Error(w, "invalid end_date format (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	entity, err := h.svc.Create(r.Context(), driverID, req.LegalEntityID, req.FleetID, startDate, endDate)
	if err != nil {
		h.handleError(w, r, "create contract", err)
		return
	}
	resp := contractToResponse(entity)
	respondJSON(w, http.StatusCreated, resp)
}

func (h *ContractHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entity, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, r, "get contract", err)
		return
	}
	respondJSON(w, http.StatusOK, contractToResponse(entity))
}

func (h *ContractHandler) listByDriver(w http.ResponseWriter, r *http.Request) {
	driverID := chi.URLParam(r, "driverId")
	entities, err := h.svc.ListByDriver(r.Context(), driverID)
	if err != nil {
		h.handleError(w, r, "list contracts", err)
		return
	}
	resp := make([]contractResponse, 0, len(entities))
	for _, e := range entities {
		resp = append(resp, contractToResponse(e))
	}
	respondJSON(w, http.StatusOK, resp)
}

func (h *ContractHandler) terminate(w http.ResponseWriter, r *http.Request) {
	if !requireJSON(w, r) {
		return
	}
	id := chi.URLParam(r, "id")
	var req terminateContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	entity, err := h.svc.Terminate(r.Context(), id, req.TerminatedBy)
	if err != nil {
		h.handleError(w, r, "terminate contract", err)
		return
	}
	respondJSON(w, http.StatusOK, contractToResponse(entity))
}

func (h *ContractHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleError(w, r, "delete contract", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ContractHandler) undelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Undelete(r.Context(), id); err != nil {
		h.handleError(w, r, "undelete contract", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func contractToResponse(e *domain.Contract) contractResponse {
	var terminatedAt *string
	if e.TerminatedAt != nil {
		s := e.TerminatedAt.Format(time.RFC3339)
		terminatedAt = &s
	}
	return contractResponse{
		ID:            e.ID,
		DriverID:      e.DriverID,
		LegalEntityID: e.LegalEntityID,
		FleetID:       e.FleetID,
		StartDate:     formatDate(e.StartDate),
		EndDate:       formatDate(e.EndDate),
		TerminatedAt:  terminatedAt,
		TerminatedBy:  e.TerminatedBy,
	}
}

func (h *ContractHandler) handleError(w http.ResponseWriter, r *http.Request, op string, err error) {
	if errors.Is(err, domain.ErrInvalidInput) || errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrEntityNotFound) || errors.Is(err, domain.ErrConflict) || errors.Is(err, domain.ErrAlreadyDeleted) {
		http.Error(w, err.Error(), mapDomainErrorToStatus(err))
		return
	}
	h.logger.Error("contract operation failed", zap.String("op", op), zap.Error(err))
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
