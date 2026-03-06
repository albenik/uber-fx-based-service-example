package main_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	main "github.com/albenik/uber-fx-based-service-example/cmd/server"
)

func TestAppWiring(t *testing.T) {
	err := fx.ValidateApp(main.AppModules()...)
	require.NoError(t, err)
}
