package driverlicense

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
	driverlicensev1 "github.com/albenik/uber-fx-based-service-example/internal/gen/driverlicense/v1"
)

// noopValidator implements ports.DriverLicenseValidator when the gRPC service is not configured.
type noopValidator struct{}

func (noopValidator) ValidateLicense(context.Context, string, string, string) (domain.LicenseValidationResult, error) {
	return "", fmt.Errorf("%w: DRIVER_LICENSE_GRPC_ADDR is empty", domain.ErrValidationServiceUnavailable)
}

// Client implements ports.DriverLicenseValidator using the external gRPC service.
type Client struct {
	grpcClient driverlicensev1.DriverLicenseValidationServiceClient
	logger     *zap.Logger
}

// NewClient creates a new driver license validation gRPC client.
func NewClient(conn *grpc.ClientConn, logger *zap.Logger) *Client {
	return &Client{
		grpcClient: driverlicensev1.NewDriverLicenseValidationServiceClient(conn),
		logger:     logger,
	}
}

// ValidateLicense calls the external gRPC service to validate driver license data.
func (c *Client) ValidateLicense(ctx context.Context, firstName, lastName, licenseNumber string) (domain.LicenseValidationResult, error) {
	resp, err := c.grpcClient.ValidateLicense(ctx, &driverlicensev1.ValidateLicenseRequest{
		FirstName:     firstName,
		LastName:      lastName,
		LicenseNumber: licenseNumber,
	})
	if err != nil {
		c.logger.Error("gRPC license validation failed", zap.Error(err))
		return "", fmt.Errorf("license validation request failed: %w", err)
	}
	return protoResultToDomain(resp.Result), nil
}

// Ensure Client implements ports.DriverLicenseValidator.
var _ ports.DriverLicenseValidator = (*Client)(nil)

func protoResultToDomain(r driverlicensev1.ValidationResult) domain.LicenseValidationResult {
	switch r {
	case driverlicensev1.ValidationResult_VALIDATION_RESULT_OK:
		return domain.LicenseValid
	case driverlicensev1.ValidationResult_VALIDATION_RESULT_NOT_FOUND:
		return domain.LicenseNotFound
	case driverlicensev1.ValidationResult_VALIDATION_RESULT_DATA_MISMATCH:
		return domain.LicenseDataMismatch
	default:
		return domain.LicenseValidationUnknown
	}
}
