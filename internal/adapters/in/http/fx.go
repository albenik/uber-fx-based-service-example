package http

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"go.uber.org/fx"
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

func httpServerLifecycle(lc fx.Lifecycle, server *http.Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", server.Addr)
			if err != nil {
				return err
			}
			fmt.Printf("HTTP server listening on %s\n", server.Addr)
			go server.Serve(ln)
			return nil
		},

		OnStop: func(ctx context.Context) error {
			fmt.Println("Shutting down HTTP server...")
			return server.Shutdown(ctx)
		},
	})
}
