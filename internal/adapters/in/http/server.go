package http

import (
	"net/http"
)

func NewServer(handler *UserHandler) *http.Server {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Register user routes
	handler.RegisterRoutes(mux)

	return &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
}
