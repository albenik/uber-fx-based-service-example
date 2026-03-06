package featuretoggle_test

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/albenik/uber-fx-based-service-example/internal/adapters/out/grpc/featuretoggle"
	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	featuretogglev1 "github.com/albenik/uber-fx-based-service-example/internal/gen/featuretoggle/v1"
)

type fakeServer struct {
	featuretogglev1.UnimplementedFeatureToggleServiceServer
	toggles map[string]*featuretogglev1.FeatureToggle
}

func (s *fakeServer) IsEnabled(_ context.Context, req *featuretogglev1.IsEnabledRequest) (*featuretogglev1.IsEnabledResponse, error) {
	t, ok := s.toggles[req.Name]
	if !ok {
		return &featuretogglev1.IsEnabledResponse{Enabled: false}, nil
	}
	return &featuretogglev1.IsEnabledResponse{Enabled: t.Enabled}, nil
}

func (s *fakeServer) GetToggle(_ context.Context, req *featuretogglev1.GetToggleRequest) (*featuretogglev1.GetToggleResponse, error) {
	t, ok := s.toggles[req.Name]
	if !ok {
		return &featuretogglev1.GetToggleResponse{Toggle: nil}, nil
	}
	return &featuretogglev1.GetToggleResponse{Toggle: t}, nil
}

func (s *fakeServer) ListToggles(context.Context, *featuretogglev1.ListTogglesRequest) (*featuretogglev1.ListTogglesResponse, error) {
	var out []*featuretogglev1.FeatureToggle
	for _, t := range s.toggles {
		out = append(out, t)
	}
	return &featuretogglev1.ListTogglesResponse{Toggles: out}, nil
}

func setupGRPC(t *testing.T, toggles map[string]*featuretogglev1.FeatureToggle) *featuretoggle.Client {
	t.Helper()
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	srv := grpc.NewServer()
	featuretogglev1.RegisterFeatureToggleServiceServer(srv, &fakeServer{toggles: toggles})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.GracefulStop)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return featuretoggle.NewClient(conn, zaptest.NewLogger(t))
}

func TestClient_IsEnabled(t *testing.T) {
	client := setupGRPC(t, map[string]*featuretogglev1.FeatureToggle{
		"dark-mode": {Name: "dark-mode", Enabled: true, Description: "Dark mode"},
	})

	enabled, err := client.IsEnabled(context.Background(), "dark-mode")
	require.NoError(t, err)
	assert.True(t, enabled)
}

func TestClient_IsEnabled_Missing(t *testing.T) {
	client := setupGRPC(t, map[string]*featuretogglev1.FeatureToggle{})

	enabled, err := client.IsEnabled(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.False(t, enabled)
}

func TestClient_GetToggle(t *testing.T) {
	client := setupGRPC(t, map[string]*featuretogglev1.FeatureToggle{
		"dark-mode": {Name: "dark-mode", Enabled: true, Description: "Dark mode UI"},
	})

	toggle, err := client.GetToggle(context.Background(), "dark-mode")
	require.NoError(t, err)
	assert.Equal(t, "dark-mode", toggle.Name)
	assert.True(t, toggle.Enabled)
	assert.Equal(t, "Dark mode UI", toggle.Description)
}

func TestClient_GetToggle_NotFound(t *testing.T) {
	client := setupGRPC(t, map[string]*featuretogglev1.FeatureToggle{})

	_, err := client.GetToggle(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestClient_ListToggles(t *testing.T) {
	client := setupGRPC(t, map[string]*featuretogglev1.FeatureToggle{
		"a": {Name: "a", Enabled: true, Description: "Toggle A"},
		"b": {Name: "b", Enabled: false, Description: "Toggle B"},
	})

	toggles, err := client.ListToggles(context.Background())
	require.NoError(t, err)
	assert.Len(t, toggles, 2)
}
