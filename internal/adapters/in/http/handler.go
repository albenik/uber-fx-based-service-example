package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

type FooEntityHandler struct {
	svc ports.FooEntityService
}

func NewUserHandler(svc ports.FooEntityService) *FooEntityHandler {
	return &FooEntityHandler{svc: svc}
}

func (h *FooEntityHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /foos", h.listFooEntities)
	mux.HandleFunc("POST /foos", h.createFooEntity)
	mux.HandleFunc("GET /foos/{id}", h.getFooEntity)
	mux.HandleFunc("DELETE /foos/{id}", h.deleteFooEntity)
}

type createFooEntityRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type fooEntityResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *FooEntityHandler) createFooEntity(w http.ResponseWriter, r *http.Request) {
	var req createFooEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.svc.CreateEntity(r.Context(), req.Name, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, responseFromFooEntity(user))
}

func (h *FooEntityHandler) getFooEntity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = strings.TrimPrefix(r.URL.Path, "/users/")
	}

	user, err := h.svc.GetEntity(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrEntityNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, responseFromFooEntity(user))
}

func (h *FooEntityHandler) listFooEntities(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.ListEntities(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]fooEntityResponse, len(users))
	for i, u := range users {
		resp[i] = responseFromFooEntity(u)
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *FooEntityHandler) deleteFooEntity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = strings.TrimPrefix(r.URL.Path, "/users/")
	}

	if err := h.svc.DeleteEntity(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrEntityNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func responseFromFooEntity(u *domain.FooEntity) fooEntityResponse {
	return fooEntityResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Description,
	}
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
