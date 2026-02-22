package vehicle

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

type IDGenerator func() string

type Service struct {
	fleetRepo ports.FleetRepository
	repo      ports.VehicleRepository
	logger    *zap.Logger
	idGen     IDGenerator
}

func New(fleetRepo ports.FleetRepository, repo ports.VehicleRepository, logger *zap.Logger, idGen IDGenerator) *Service {
	return &Service{fleetRepo: fleetRepo, repo: repo, logger: logger, idGen: idGen}
}

func (s *Service) Create(ctx context.Context, fleetID, make, model, licensePlate string, year int) (*domain.Vehicle, error) {
	make = strings.TrimSpace(make)
	model = strings.TrimSpace(model)
	licensePlate = strings.TrimSpace(licensePlate)
	if fleetID == "" {
		return nil, fmt.Errorf("%w: fleet_id is required", domain.ErrInvalidInput)
	}
	if make == "" {
		return nil, fmt.Errorf("%w: make is required", domain.ErrInvalidInput)
	}
	if model == "" {
		return nil, fmt.Errorf("%w: model is required", domain.ErrInvalidInput)
	}
	if year < 1900 || year > 2100 {
		return nil, fmt.Errorf("%w: year must be between 1900 and 2100", domain.ErrInvalidInput)
	}
	if _, err := s.fleetRepo.FindByID(ctx, fleetID); err != nil {
		return nil, err
	}
	id := s.idGen()
	if id == "" {
		return nil, fmt.Errorf("id generator returned empty ID")
	}
	entity := &domain.Vehicle{ID: id, FleetID: fleetID, Make: make, Model: model, Year: year, LicensePlate: licensePlate}
	if err := s.repo.Save(ctx, entity); err != nil {
		s.logger.Error("Failed to save vehicle", zap.String("id", id), zap.Error(err))
		return nil, err
	}
	s.logger.Info("Created vehicle", zap.String("id", id))
	result := *entity
	return &result, nil
}

func (s *Service) Get(ctx context.Context, id string) (*domain.Vehicle, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListByFleet(ctx context.Context, fleetID string) ([]*domain.Vehicle, error) {
	if fleetID == "" {
		return nil, fmt.Errorf("%w: fleet_id is required", domain.ErrInvalidInput)
	}
	return s.repo.FindByFleetID(ctx, fleetID)
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
