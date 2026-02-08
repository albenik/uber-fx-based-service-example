package http

import (
	"context"
	"net"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides input adapters (driving adapters).
func Module() fx.Option {
	return fx.Module("http",
		fx.Provide(
			NewUserHandler,
			NewServer,
		),
		fx.Invoke(
			httpServerLifecycle, // ensures server starts and stops with the application
		),
	)
}

func httpServerLifecycle(lc fx.Lifecycle, server *http.Server, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", server.Addr)
			if err != nil {
				return err
			}

			logger.Info("HTTP server listening", zap.String("address", server.Addr))
			go server.Serve(ln)

			return nil
		},

		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down HTTP server...")
			return server.Shutdown(ctx)
		},
	})
}
