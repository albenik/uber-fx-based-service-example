package assignment

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
	contractRepo ports.ContractRepository
	vehicleRepo  ports.VehicleRepository
	repo         ports.VehicleAssignmentRepository
	logger       *zap.Logger
	idGen        IDGenerator
	clock        Clock
}

func New(contractRepo ports.ContractRepository, vehicleRepo ports.VehicleRepository, repo ports.VehicleAssignmentRepository, logger *zap.Logger, idGen IDGenerator, clock Clock) *Service {
	return &Service{
		contractRepo: contractRepo,
		vehicleRepo:  vehicleRepo,
		repo:         repo,
		logger:       logger,
		idGen:        idGen,
		clock:        clock,
	}
}

func (s *Service) Assign(ctx context.Context, contractID, vehicleID string) (*domain.VehicleAssignment, error) {
	if contractID == "" || vehicleID == "" {
		return nil, fmt.Errorf("%w: contract_id and vehicle_id are required", domain.ErrInvalidInput)
	}
	contract, err := s.contractRepo.FindByID(ctx, contractID)
	if err != nil {
		return nil, err
	}
	vehicle, err := s.vehicleRepo.FindByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if vehicle.FleetID != contract.FleetID {
		return nil, fmt.Errorf("%w: vehicle must belong to the contract's fleet", domain.ErrInvalidInput)
	}
	now := s.clock()
	// contract is active through the entire EndDate day (inclusive)
	if now.Before(contract.StartDate) || !now.Before(contract.EndDate.AddDate(0, 0, 1)) {
		return nil, domain.ErrContractNotActive
	}
	if contract.TerminatedAt != nil && now.After(*contract.TerminatedAt) {
		return nil, domain.ErrContractNotActive
	}
	existing, err := s.repo.FindActiveByDriverIDAndFleetID(ctx, contract.DriverID, contract.FleetID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrDriverAlreadyAssignedInFleet
	}
	id := s.idGen()
	if id == "" {
		return nil, fmt.Errorf("id generator returned empty ID")
	}
	entity := &domain.VehicleAssignment{
		ID:         id,
		DriverID:   contract.DriverID,
		VehicleID:  vehicleID,
		ContractID: contractID,
		StartTime:  now,
	}
	if err := s.repo.Save(ctx, entity); err != nil {
		s.logger.Error("Failed to save vehicle assignment", zap.String("id", id), zap.Error(err))
		return nil, err
	}
	s.logger.Info("Created vehicle assignment", zap.String("id", id))
	result := *entity
	return &result, nil
}

func (s *Service) Get(ctx context.Context, id string) (*domain.VehicleAssignment, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListByContract(ctx context.Context, contractID string) ([]*domain.VehicleAssignment, error) {
	if contractID == "" {
		return nil, fmt.Errorf("%w: contract_id is required", domain.ErrInvalidInput)
	}
	return s.repo.FindByContractID(ctx, contractID)
}

func (s *Service) Return(ctx context.Context, id string) (*domain.VehicleAssignment, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	entity, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if entity.EndTime != nil {
		return nil, fmt.Errorf("%w: vehicle already returned", domain.ErrConflict)
	}
	now := s.clock()
	entity.EndTime = &now
	if err := s.repo.Save(ctx, entity); err != nil {
		s.logger.Error("Failed to save returned assignment", zap.String("id", id), zap.Error(err))
		return nil, err
	}
	s.logger.Info("Returned vehicle assignment", zap.String("id", id))
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
