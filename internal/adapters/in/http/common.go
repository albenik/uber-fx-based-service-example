package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime"
	"net/http"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

func requireJSON(w http.ResponseWriter, r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		http.Error(w, "unsupported media type", http.StatusUnsupportedMediaType)
		return false
	}
	return true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return false
		}
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return false
	}
	if dec.More() {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return false
	}
	return true
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

func mapDomainErrorToStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrEntityNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, domain.ErrContractNotActive):
		return http.StatusUnprocessableEntity
	case errors.Is(err, domain.ErrVehicleAlreadyAssigned):
		return http.StatusConflict
	case errors.Is(err, domain.ErrDriverHasActiveContracts), errors.Is(err, domain.ErrDriverHasActiveAssignments):
		return http.StatusConflict
	case errors.Is(err, domain.ErrAlreadyDeleted):
		return http.StatusConflict
	case errors.Is(err, domain.ErrValidationServiceUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
