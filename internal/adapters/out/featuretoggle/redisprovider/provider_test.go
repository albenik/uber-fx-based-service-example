package redisprovider_test

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/albenik/uber-fx-based-service-example/internal/adapters/out/featuretoggle/redisprovider"
	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

func newTestClient(t *testing.T) *redis.Client {
	t.Helper()
	client := redis.NewClient(&redis.Options{
		Addr:       "localhost:6379",
		DB:         15,
		MaxRetries: 0,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		t.Skipf("Redis not available: %v", err)
	}
	t.Cleanup(func() { client.FlushDB(context.Background()); _ = client.Close() })
	return client
}

func seedToggle(t *testing.T, client *redis.Client, name, value string) {
	t.Helper()
	require.NoError(t, client.HSet(context.Background(), "feature_toggles", name, value).Err())
}

func TestProvider_IsEnabled_True(t *testing.T) {
	client := newTestClient(t)
	seedToggle(t, client, "dark-mode", `{"enabled":true,"description":"Dark mode"}`)

	p := redisprovider.NewProvider(client, zaptest.NewLogger(t))
	enabled, err := p.IsEnabled(context.Background(), "dark-mode")
	require.NoError(t, err)
	assert.True(t, enabled)
}

func TestProvider_IsEnabled_False(t *testing.T) {
	client := newTestClient(t)
	seedToggle(t, client, "beta", `{"enabled":false,"description":"Beta"}`)

	p := redisprovider.NewProvider(client, zaptest.NewLogger(t))
	enabled, err := p.IsEnabled(context.Background(), "beta")
	require.NoError(t, err)
	assert.False(t, enabled)
}

func TestProvider_IsEnabled_Missing(t *testing.T) {
	client := newTestClient(t)

	p := redisprovider.NewProvider(client, zaptest.NewLogger(t))
	enabled, err := p.IsEnabled(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.False(t, enabled)
}

func TestProvider_GetToggle(t *testing.T) {
	client := newTestClient(t)
	seedToggle(t, client, "dark-mode", `{"enabled":true,"description":"Enable dark mode"}`)

	p := redisprovider.NewProvider(client, zaptest.NewLogger(t))
	toggle, err := p.GetToggle(context.Background(), "dark-mode")
	require.NoError(t, err)
	assert.Equal(t, "dark-mode", toggle.Name)
	assert.True(t, toggle.Enabled)
	assert.Equal(t, "Enable dark mode", toggle.Description)
}

func TestProvider_GetToggle_NotFound(t *testing.T) {
	client := newTestClient(t)

	p := redisprovider.NewProvider(client, zaptest.NewLogger(t))
	_, err := p.GetToggle(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestProvider_ListToggles(t *testing.T) {
	client := newTestClient(t)
	seedToggle(t, client, "dark-mode", `{"enabled":true,"description":"Dark mode"}`)
	seedToggle(t, client, "beta", `{"enabled":false,"description":"Beta feature"}`)

	p := redisprovider.NewProvider(client, zaptest.NewLogger(t))
	toggles, err := p.ListToggles(context.Background())
	require.NoError(t, err)
	assert.Len(t, toggles, 2)

	byName := make(map[string]*domain.FeatureToggle, len(toggles))
	for _, tgl := range toggles {
		byName[tgl.Name] = tgl
	}
	assert.True(t, byName["dark-mode"].Enabled)
	assert.False(t, byName["beta"].Enabled)
}

func TestProvider_ListToggles_Empty(t *testing.T) {
	client := newTestClient(t)

	p := redisprovider.NewProvider(client, zaptest.NewLogger(t))
	toggles, err := p.ListToggles(context.Background())
	require.NoError(t, err)
	assert.Empty(t, toggles)
}

func TestProvider_IsEnabled_CorruptData(t *testing.T) {
	client := newTestClient(t)
	seedToggle(t, client, "broken", `not-json`)

	p := redisprovider.NewProvider(client, zaptest.NewLogger(t))
	_, err := p.IsEnabled(context.Background(), "broken")
	assert.ErrorIs(t, err, domain.ErrToggleProviderUnavailable)
}

func TestProvider_ListToggles_SkipsCorruptEntries(t *testing.T) {
	client := newTestClient(t)
	seedToggle(t, client, "good", `{"enabled":true,"description":"OK"}`)
	seedToggle(t, client, "broken", `not-json`)

	p := redisprovider.NewProvider(client, zaptest.NewLogger(t))
	toggles, err := p.ListToggles(context.Background())
	require.NoError(t, err)
	assert.Len(t, toggles, 1)
	assert.Equal(t, "good", toggles[0].Name)
}
