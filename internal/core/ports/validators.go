package ports

//go:generate go tool mockgen -destination=mocks/mock_validators.go -package=mocks . DriverLicenseValidator

import (
	"context"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// DriverLicenseValidator is the output port for external driver license validation.
type DriverLicenseValidator interface {
	ValidateLicense(ctx context.Context, firstName, lastName, licenseNumber string) (domain.LicenseValidationResult, error)
}
