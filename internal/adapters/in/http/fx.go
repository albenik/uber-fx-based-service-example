package http

import (
	"context"
	"errors"
	"net"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides input adapters (driving adapters).
func Module() fx.Option {
	return fx.Module("http",
		fx.Provide(
			fx.Annotate(
				NewLegalEntityHandler,
				fx.ResultTags(`group:"routes"`),
			),
			fx.Annotate(
				NewFleetHandler,
				fx.ResultTags(`group:"routes"`),
			),
			fx.Annotate(
				NewVehicleHandler,
				fx.ResultTags(`group:"routes"`),
			),
			fx.Annotate(
				NewDriverHandler,
				fx.ResultTags(`group:"routes"`),
			),
			fx.Annotate(
				NewContractHandler,
				fx.ResultTags(`group:"routes"`),
			),
			fx.Annotate(
				NewAssignmentHandler,
				fx.ResultTags(`group:"routes"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				NewServer,
				fx.ParamTags(``, `group:"routes"`),
			),
		),
		fx.Invoke(httpServerLifecycle),
	)
}

func httpServerLifecycle(lc fx.Lifecycle, server *http.Server, shutdowner fx.Shutdowner, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", server.Addr)
			if err != nil {
				return err
			}

			logger.Info("HTTP server listening", zap.String("address", ln.Addr().String()))
			go func() {
				if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Error("HTTP server error", zap.Error(err))
					if shutdownErr := shutdowner.Shutdown(); shutdownErr != nil {
						logger.Error("failed to trigger shutdown", zap.Error(shutdownErr))
					}
				}
			}()

			return nil
		},

		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down HTTP server")
			return server.Shutdown(ctx)
		},
	})
}
