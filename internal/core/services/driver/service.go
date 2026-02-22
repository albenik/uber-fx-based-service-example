package driver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

type IDGenerator func() string

type Clock func() time.Time

type Service struct {
	repo           ports.DriverRepository
	contractRepo   ports.ContractRepository
	assignmentRepo ports.VehicleAssignmentRepository
	validator      ports.DriverLicenseValidator
	idGen          IDGenerator
	clock          Clock
	logger         *zap.Logger
}

func New(
	repo ports.DriverRepository,
	contractRepo ports.ContractRepository,
	assignmentRepo ports.VehicleAssignmentRepository,
	validator ports.DriverLicenseValidator,
	idGen IDGenerator,
	clock Clock,
	logger *zap.Logger,
) *Service {
	return &Service{
		repo:           repo,
		contractRepo:   contractRepo,
		assignmentRepo: assignmentRepo,
		validator:      validator,
		idGen:          idGen,
		clock:          clock,
		logger:         logger,
	}
}

func (s *Service) Create(ctx context.Context, firstName, lastName, licenseNumber string) (*domain.Driver, error) {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	licenseNumber = strings.TrimSpace(licenseNumber)
	if firstName == "" {
		return nil, fmt.Errorf("%w: first_name is required", domain.ErrInvalidInput)
	}
	if lastName == "" {
		return nil, fmt.Errorf("%w: last_name is required", domain.ErrInvalidInput)
	}
	if licenseNumber == "" {
		return nil, fmt.Errorf("%w: license_number is required", domain.ErrInvalidInput)
	}
	result, err := s.validator.ValidateLicense(ctx, firstName, lastName, licenseNumber)
	if err != nil {
		return nil, err
	}
	if result != domain.LicenseValid {
		return nil, fmt.Errorf("%w: %s", domain.ErrLicenseValidationFailed, result)
	}
	id := s.idGen()
	if id == "" {
		return nil, fmt.Errorf("id generator returned empty ID")
	}
	entity := &domain.Driver{ID: id, FirstName: firstName, LastName: lastName, LicenseNumber: licenseNumber}
	if err := s.repo.Save(ctx, entity); err != nil {
		s.logger.Error("Failed to save driver", zap.String("id", id), zap.Error(err))
		return nil, err
	}
	s.logger.Info("Created driver", zap.String("id", id))
	out := *entity
	return &out, nil
}

func (s *Service) Get(ctx context.Context, id string) (*domain.Driver, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*domain.Driver, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	contracts, err := s.contractRepo.FindByDriverID(ctx, id)
	if err != nil {
		return err
	}
	now := s.clock()
	for _, c := range contracts {
		if c.TerminatedAt == nil && now.Before(c.EndDate.AddDate(0, 0, 1)) {
			return domain.ErrDriverHasActiveContracts
		}
	}
	activeAssignments, err := s.assignmentRepo.FindActiveByDriverID(ctx, id)
	if err != nil {
		return err
	}
	if len(activeAssignments) > 0 {
		return domain.ErrDriverHasActiveAssignments
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *Service) Undelete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	return s.repo.Undelete(ctx, id)
}

func (s *Service) ValidateLicense(ctx context.Context, id string) (domain.LicenseValidationResult, error) {
	if id == "" {
		return "", fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	driver, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	return s.validator.ValidateLicense(ctx, driver.FirstName, driver.LastName, driver.LicenseNumber)
}
