package telemetry

import "go.uber.org/fx"

func Module() fx.Option {
	return fx.Module("telemetry",
		fx.Provide(NewLogger),
	)
}
