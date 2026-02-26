package contract

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

type IDGenerator func() string

type Clock func() time.Time

type Service struct {
	driverRepo ports.DriverRepository
	legalRepo  ports.LegalEntityRepository
	fleetRepo  ports.FleetRepository
	repo       ports.ContractRepository

	idGen IDGenerator
	clock Clock

	logger *zap.Logger
}

func New(
	driverRepo ports.DriverRepository,
	legalRepo ports.LegalEntityRepository,
	fleetRepo ports.FleetRepository,
	repo ports.ContractRepository,
	idGen IDGenerator,
	clock Clock,
	logger *zap.Logger,
) *Service {
	return &Service{
		driverRepo: driverRepo,
		legalRepo:  legalRepo,
		fleetRepo:  fleetRepo,
		repo:       repo,

		idGen: idGen,
		clock: clock,

		logger: logger,
	}
}

func (s *Service) Create(
	ctx context.Context,
	driverID, legalEntityID, fleetID string,
	startDate, endDate time.Time,
) (*domain.Contract, error) {
	if driverID == "" || legalEntityID == "" || fleetID == "" {
		return nil, fmt.Errorf("%w: driver_id, legal_entity_id, and fleet_id are required", domain.ErrInvalidInput)
	}
	if !endDate.After(startDate) {
		return nil, fmt.Errorf("%w: end_date must be after start_date", domain.ErrInvalidInput)
	}
	if _, err := s.driverRepo.FindByID(ctx, driverID); err != nil {
		return nil, err
	}
	if _, err := s.legalRepo.FindByID(ctx, legalEntityID); err != nil {
		return nil, err
	}
	if _, err := s.fleetRepo.FindByID(ctx, fleetID); err != nil {
		return nil, err
	}
	overlapping, err := s.repo.FindOverlapping(ctx, driverID, legalEntityID, fleetID, startDate, endDate, "")
	if err != nil {
		return nil, err
	}
	if len(overlapping) > 0 {
		return nil, fmt.Errorf("%w: contract dates overlap with existing contract", domain.ErrConflict)
	}
	id := s.idGen()
	if id == "" {
		return nil, fmt.Errorf("id generator returned empty ID")
	}
	entity := &domain.Contract{
		ID:            id,
		DriverID:      driverID,
		LegalEntityID: legalEntityID,
		FleetID:       fleetID,
		StartDate:     startDate,
		EndDate:       endDate,
	}
	if err := s.repo.Save(ctx, entity); err != nil {
		s.logger.Error("Failed to save contract", zap.String("id", id), zap.Error(err))
		return nil, err
	}
	s.logger.Info("Created contract", zap.String("id", id))
	result := *entity
	return &result, nil
}

func (s *Service) Get(ctx context.Context, id string) (*domain.Contract, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListByDriver(ctx context.Context, driverID string) ([]*domain.Contract, error) {
	if driverID == "" {
		return nil, fmt.Errorf("%w: driver_id is required", domain.ErrInvalidInput)
	}
	return s.repo.FindByDriverID(ctx, driverID)
}

func (s *Service) Terminate(ctx context.Context, id, terminatedBy string) (*domain.Contract, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	if terminatedBy == "" {
		return nil, fmt.Errorf("%w: terminated_by is required", domain.ErrInvalidInput)
	}
	entity, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if entity.TerminatedAt != nil {
		return nil, fmt.Errorf("%w: contract is already terminated", domain.ErrConflict)
	}
	now := s.clock()
	entity.TerminatedAt = &now
	entity.TerminatedBy = terminatedBy
	if err := s.repo.Save(ctx, entity); err != nil {
		s.logger.Error("Failed to save terminated contract", zap.String("id", id), zap.Error(err))
		return nil, err
	}
	s.logger.Info("Terminated contract", zap.String("id", id))
	result := *entity
	return &result, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *Service) Undelete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	return s.repo.Undelete(ctx, id)
}
