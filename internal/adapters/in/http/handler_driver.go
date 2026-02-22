package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

type DriverHandler struct {
	svc    ports.DriverService
	logger *zap.Logger
}

func NewDriverHandler(svc ports.DriverService, logger *zap.Logger) *DriverHandler {
	return &DriverHandler{svc: svc, logger: logger}
}

func (h *DriverHandler) RegisterRoutes(r chi.Router) {
	r.Route("/drivers", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{id}", h.get)
		r.Delete("/{id}", h.delete)
		r.Post("/{id}/undelete", h.undelete)
		r.Post("/{id}/validate", h.validateLicense)
	})
}

type createDriverRequest struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	LicenseNumber string `json:"license_number"`
}

type driverResponse struct {
	ID            string `json:"id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	LicenseNumber string `json:"license_number"`
}

func (h *DriverHandler) create(w http.ResponseWriter, r *http.Request) {
	if !requireJSON(w, r) {
		return
	}
	var req createDriverRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	entity, err := h.svc.Create(r.Context(), req.FirstName, req.LastName, req.LicenseNumber)
	if err != nil {
		h.handleError(w, r, "create driver", err)
		return
	}
	respondJSON(w, http.StatusCreated, driverResponse{ID: entity.ID, FirstName: entity.FirstName, LastName: entity.LastName, LicenseNumber: entity.LicenseNumber})
}

func (h *DriverHandler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entity, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, r, "get driver", err)
		return
	}
	respondJSON(w, http.StatusOK, driverResponse{ID: entity.ID, FirstName: entity.FirstName, LastName: entity.LastName, LicenseNumber: entity.LicenseNumber})
}

func (h *DriverHandler) list(w http.ResponseWriter, r *http.Request) {
	entities, err := h.svc.List(r.Context())
	if err != nil {
		h.handleError(w, r, "list drivers", err)
		return
	}
	resp := make([]driverResponse, 0, len(entities))
	for _, e := range entities {
		resp = append(resp, driverResponse{ID: e.ID, FirstName: e.FirstName, LastName: e.LastName, LicenseNumber: e.LicenseNumber})
	}
	respondJSON(w, http.StatusOK, resp)
}

func (h *DriverHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleError(w, r, "delete driver", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *DriverHandler) undelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Undelete(r.Context(), id); err != nil {
		h.handleError(w, r, "undelete driver", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type validateLicenseResponse struct {
	DriverID string `json:"driver_id"`
	Result   string `json:"result"`
}

func (h *DriverHandler) validateLicense(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, err := h.svc.ValidateLicense(r.Context(), id)
	if err != nil {
		h.handleError(w, r, "validate driver license", err)
		return
	}
	respondJSON(w, http.StatusOK, validateLicenseResponse{DriverID: id, Result: string(result)})
}

func (h *DriverHandler) handleError(w http.ResponseWriter, r *http.Request, op string, err error) {
	if errors.Is(err, domain.ErrInvalidInput) || errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrEntityNotFound) || errors.Is(err, domain.ErrConflict) || errors.Is(err, domain.ErrAlreadyDeleted) || errors.Is(err, domain.ErrDriverHasActiveContracts) || errors.Is(err, domain.ErrDriverHasActiveAssignments) || errors.Is(err, domain.ErrValidationServiceUnavailable) {
		http.Error(w, err.Error(), mapDomainErrorToStatus(err))
		return
	}
	h.logger.Error("driver operation failed", zap.String("op", op), zap.Error(err))
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
