package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

type UserHandler struct {
	svc ports.UserService
}

func NewUserHandler(svc ports.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users", h.listUsers)
	mux.HandleFunc("POST /users", h.createUser)
	mux.HandleFunc("GET /users/{id}", h.getUser)
	mux.HandleFunc("DELETE /users/{id}", h.deleteUser)
}

type createUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type userResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.svc.CreateUser(r.Context(), req.Name, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, toUserResponse(user))
}

func (h *UserHandler) getUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = strings.TrimPrefix(r.URL.Path, "/users/")
	}

	user, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, toUserResponse(user))
}

func (h *UserHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.ListUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]userResponse, len(users))
	for i, u := range users {
		resp[i] = toUserResponse(u)
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) deleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = strings.TrimPrefix(r.URL.Path, "/users/")
	}

	if err := h.svc.DeleteUser(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
	}
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
