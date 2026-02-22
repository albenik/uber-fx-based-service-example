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

type AssignmentHandler struct {
	svc    ports.VehicleAssignmentService
	logger *zap.Logger
}

func NewAssignmentHandler(svc ports.VehicleAssignmentService, logger *zap.Logger) *AssignmentHandler {
	return &AssignmentHandler{svc: svc, logger: logger}
}

func (h *AssignmentHandler) RegisterRoutes(r chi.Router) {
	r.Route("/contracts/{contractId}/assignments", func(r chi.Router) {
		r.Get("/", h.listByContract)
		r.Post("/", h.assign)
	})
	r.Route("/assignments", func(r chi.Router) {
		r.Get("/{id}", h.get)
		r.Post("/{id}/return", h.returnVehicle)
		r.Delete("/{id}", h.delete)
		r.Post("/{id}/undelete", h.undelete)
	})
}

type createAssignmentRequest struct {
	VehicleID string `json:"vehicle_id"`
}

type assignmentResponse struct {
	ID         string  `json:"id"`
	DriverID   string  `json:"driver_id"`
	VehicleID  string  `json:"vehicle_id"`
	ContractID string  `json:"contract_id"`
	StartTime  string  `json:"start_time"`
	EndTime    *string `json:"end_time,omitempty"`
}

func (h *AssignmentHandler) assign(w http.ResponseWriter, r *http.Request) {
	if !requireJSON(w, r) {
		return
	}
	var req createAssignmentRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	contractID := chi.URLParam(r, "contractId")
	entity, err := h.svc.Assign(r.Context(), contractID, req.VehicleID)
	if err != nil {
		h.handleError(w, r, "assign vehicle", err)
		return
	}
	respondJSON(w, http.StatusCreated, assignmentToResponse(entity))
}

func (h *AssignmentHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entity, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, r, "get assignment", err)
		return
	}
	respondJSON(w, http.StatusOK, assignmentToResponse(entity))
}

func (h *AssignmentHandler) listByContract(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractId")
	entities, err := h.svc.ListByContract(r.Context(), contractID)
	if err != nil {
		h.handleError(w, r, "list assignments", err)
		return
	}
	resp := make([]assignmentResponse, 0, len(entities))
	for _, e := range entities {
		resp = append(resp, assignmentToResponse(e))
	}
	respondJSON(w, http.StatusOK, resp)
}

func (h *AssignmentHandler) returnVehicle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&struct{}{})
	}
	entity, err := h.svc.Return(r.Context(), id)
	if err != nil {
		h.handleError(w, r, "return vehicle", err)
		return
	}
	respondJSON(w, http.StatusOK, assignmentToResponse(entity))
}

func (h *AssignmentHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleError(w, r, "delete assignment", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AssignmentHandler) undelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Undelete(r.Context(), id); err != nil {
		h.handleError(w, r, "undelete assignment", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func assignmentToResponse(e *domain.VehicleAssignment) assignmentResponse {
	var endTime *string
	if e.EndTime != nil {
		s := e.EndTime.Format(time.RFC3339)
		endTime = &s
	}
	return assignmentResponse{
		ID:         e.ID,
		DriverID:   e.DriverID,
		VehicleID:  e.VehicleID,
		ContractID: e.ContractID,
		StartTime:  e.StartTime.Format(time.RFC3339),
		EndTime:    endTime,
	}
}

func (h *AssignmentHandler) handleError(w http.ResponseWriter, r *http.Request, op string, err error) {
	if errors.Is(err, domain.ErrInvalidInput) || errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrEntityNotFound) || errors.Is(err, domain.ErrConflict) || errors.Is(err, domain.ErrAlreadyDeleted) || errors.Is(err, domain.ErrContractNotActive) || errors.Is(err, domain.ErrVehicleAlreadyAssigned) {
		http.Error(w, err.Error(), mapDomainErrorToStatus(err))
		return
	}
	h.logger.Error("assignment operation failed", zap.String("op", op), zap.Error(err))
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
