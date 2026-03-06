package grpc

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
	driverlicensev1 "github.com/albenik/uber-fx-based-service-example/internal/gen/driverlicense/v1"
)

// noopLicenseValidator implements ports.DriverLicenseValidator when the gRPC service is not configured.
type noopLicenseValidator struct{}

func (noopLicenseValidator) ValidateLicense(context.Context, string, string, string) (domain.LicenseValidationResult, error) {
	return "", fmt.Errorf("%w: DRIVER_LICENSE_GRPC_ADDR is empty", domain.ErrValidationServiceUnavailable)
}

// DriverLicenseClient implements ports.DriverLicenseValidator using the external gRPC service.
type DriverLicenseClient struct {
	grpcClient driverlicensev1.DriverLicenseValidationServiceClient
	logger     *zap.Logger
}

// NewDriverLicenseClient creates a new driver license validation gRPC client.
func NewDriverLicenseClient(conn grpc.ClientConnInterface, logger *zap.Logger) *DriverLicenseClient {
	return &DriverLicenseClient{
		grpcClient: driverlicensev1.NewDriverLicenseValidationServiceClient(conn),
		logger:     logger,
	}
}

// ValidateLicense calls the external gRPC service to validate driver license data.
func (c *DriverLicenseClient) ValidateLicense(ctx context.Context, firstName, lastName, licenseNumber string) (domain.LicenseValidationResult, error) {
	resp, err := c.grpcClient.ValidateLicense(ctx, &driverlicensev1.ValidateLicenseRequest{
		FirstName:     firstName,
		LastName:      lastName,
		LicenseNumber: licenseNumber,
	})
	if err != nil {
		c.logger.Error("gRPC license validation failed", zap.Error(err))
		return "", domain.ErrValidationServiceUnavailable
	}
	return driverLicenseProtoResultToDomain(resp.Result), nil
}

var _ ports.DriverLicenseValidator = (*DriverLicenseClient)(nil)

func driverLicenseProtoResultToDomain(r driverlicensev1.ValidationResult) domain.LicenseValidationResult {
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
