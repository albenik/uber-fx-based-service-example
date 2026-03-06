package config

// FeatureToggleConfig holds configuration for the feature toggle subsystem.
// Backend selects the adapter: "grpc", "redis", or "" (no-op).
type FeatureToggleConfig struct {
	Backend        string // "grpc", "redis", or "" for no-op
	GRPCAddr       string
	GRPCTLSEnabled bool
	RedisAddr      string // Redis URL, e.g. "redis://localhost:6379/0"
}
