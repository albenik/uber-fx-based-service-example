package main_test

import (
	"testing"

	"go.uber.org/fx/fxtest"

	main "github.com/albenik/uber-fx-based-service-example/cmd/server"
)

func TestAppWiring(t *testing.T) {
	app := fxtest.New(t, main.AppModules()...)
	app.RequireStart()
	app.RequireStop()
}
