package domain

// FeatureToggle represents a named boolean switch that controls application behavior.
type FeatureToggle struct {
	Name        string
	Enabled     bool
	Description string
}
