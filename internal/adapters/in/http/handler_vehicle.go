package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

type VehicleHandler struct {
	svc    ports.VehicleService
	logger *zap.Logger
}

func NewVehicleHandler(svc ports.VehicleService, logger *zap.Logger) *VehicleHandler {
	return &VehicleHandler{svc: svc, logger: logger}
}

func (h *VehicleHandler) RegisterRoutes(r chi.Router) {
	r.Route("/fleets/{fleetId}/vehicles", func(r chi.Router) {
		r.Get("/", h.listByFleet)
		r.Post("/", h.create)
	})
	r.Route("/vehicles", func(r chi.Router) {
		r.Get("/{id}", h.get)
		r.Delete("/{id}", h.delete)
		r.Post("/{id}/undelete", h.undelete)
	})
}

type createVehicleRequest struct {
	Make         string `json:"make"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
	LicensePlate string `json:"license_plate"`
}

type vehicleResponse struct {
	ID           string `json:"id"`
	FleetID      string `json:"fleet_id"`
	Make         string `json:"make"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
	LicensePlate string `json:"license_plate"`
}

func (h *VehicleHandler) create(w http.ResponseWriter, r *http.Request) {
	if !requireJSON(w, r) {
		return
	}
	var req createVehicleRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	fleetID := chi.URLParam(r, "fleetId")
	entity, err := h.svc.Create(r.Context(), fleetID, req.Make, req.Model, req.LicensePlate, req.Year)
	if err != nil {
		h.handleError(w, r, "create vehicle", err)
		return
	}
	respondJSON(w, http.StatusCreated, vehicleResponse{ID: entity.ID, FleetID: entity.FleetID, Make: entity.Make, Model: entity.Model, Year: entity.Year, LicensePlate: entity.LicensePlate})
}

func (h *VehicleHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entity, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, r, "get vehicle", err)
		return
	}
	respondJSON(w, http.StatusOK, vehicleResponse{ID: entity.ID, FleetID: entity.FleetID, Make: entity.Make, Model: entity.Model, Year: entity.Year, LicensePlate: entity.LicensePlate})
}

func (h *VehicleHandler) listByFleet(w http.ResponseWriter, r *http.Request) {
	fleetID := chi.URLParam(r, "fleetId")
	entities, err := h.svc.ListByFleet(r.Context(), fleetID)
	if err != nil {
		h.handleError(w, r, "list vehicles", err)
		return
	}
	resp := make([]vehicleResponse, 0, len(entities))
	for _, e := range entities {
		resp = append(resp, vehicleResponse{ID: e.ID, FleetID: e.FleetID, Make: e.Make, Model: e.Model, Year: e.Year, LicensePlate: e.LicensePlate})
	}
	respondJSON(w, http.StatusOK, resp)
}

func (h *VehicleHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleError(w, r, "delete vehicle", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *VehicleHandler) undelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Undelete(r.Context(), id); err != nil {
		h.handleError(w, r, "undelete vehicle", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *VehicleHandler) handleError(w http.ResponseWriter, r *http.Request, op string, err error) {
	if errors.Is(err, domain.ErrInvalidInput) || errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrEntityNotFound) || errors.Is(err, domain.ErrConflict) || errors.Is(err, domain.ErrAlreadyDeleted) {
		http.Error(w, err.Error(), mapDomainErrorToStatus(err))
		return
	}
	h.logger.Error("vehicle operation failed", zap.String("op", op), zap.Error(err))
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
