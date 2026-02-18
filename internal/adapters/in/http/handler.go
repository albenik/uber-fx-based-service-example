package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

const maxRequestBodySize = 1 << 20 // 1 MB

type FooEntityHandler struct {
	svc    ports.FooEntityService
	logger *zap.Logger
}

func NewFooEntityHandler(svc ports.FooEntityService, logger *zap.Logger) *FooEntityHandler {
	return &FooEntityHandler{svc: svc, logger: logger}
}

func (h *FooEntityHandler) RegisterRoutes(mux chi.Router) {
	mux.Get("/foos", h.listFooEntities)
	mux.Post("/foos", h.createFooEntity)
	mux.Get("/foos/{id}", h.getFooEntity)
	mux.Delete("/foos/{id}", h.deleteFooEntity)
}

type createFooEntityRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type fooEntityResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *FooEntityHandler) createFooEntity(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
		return
	}

	var req createFooEntityRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if dec.More() {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	entity, err := h.svc.CreateEntity(r.Context(), req.Name, req.Description)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}
		h.logger.Error("failed to create entity", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, responseFromFooEntity(entity))
}

func (h *FooEntityHandler) getFooEntity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing entity id", http.StatusBadRequest)
		return
	}

	entity, err := h.svc.GetEntity(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}
		if errors.Is(err, domain.ErrEntityNotFound) {
			http.Error(w, "entity not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to get entity", zap.String("id", id), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, responseFromFooEntity(entity))
}

func (h *FooEntityHandler) listFooEntities(w http.ResponseWriter, r *http.Request) {
	entities, err := h.svc.ListEntities(r.Context())
	if err != nil {
		h.logger.Error("failed to list entities", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]fooEntityResponse, 0, len(entities))
	for _, e := range entities {
		resp = append(resp, responseFromFooEntity(e))
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *FooEntityHandler) deleteFooEntity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing entity id", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteEntity(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}
		if errors.Is(err, domain.ErrEntityNotFound) {
			http.Error(w, "entity not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to delete entity", zap.String("id", id), zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func responseFromFooEntity(e *domain.FooEntity) fooEntityResponse {
	return fooEntityResponse{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
	}
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}
