package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/albenik/uber-fx-based-service-example/internal/config"
)

const maxRequestBodySize = 1 << 20 // 1 MB

// RouteRegistrar registers HTTP routes on a chi router.
type RouteRegistrar interface {
	RegisterRoutes(chi.Router)
}

func NewServer(cfg *config.HTTPServerConfig, handler RouteRegistrar) *http.Server {
	mux := chi.NewRouter()

	mux.Use(maxBytesMiddleware(maxRequestBodySize))

	// Health check
	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Register entity routes
	handler.RegisterRoutes(mux)

	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

func maxBytesMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}
