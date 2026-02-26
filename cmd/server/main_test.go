package main_test

import (
	"testing"

	"go.uber.org/fx/fxtest"

	main "github.com/albenik/uber-fx-based-service-example/cmd/server"
)

func TestAppWiring(t *testing.T) {
	t.Setenv("DATABASE_MASTER_URL", "postgres://localhost/test")
	app := fxtest.New(t, main.AppModules()...)
	app.RequireStart()
	app.RequireStop()
}
