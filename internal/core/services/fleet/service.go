package fleet

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
	legalEntityRepo ports.LegalEntityRepository
	repo            ports.FleetRepository
	logger          *zap.Logger
	idGen           IDGenerator
}

func New(legalEntityRepo ports.LegalEntityRepository, repo ports.FleetRepository, logger *zap.Logger, idGen IDGenerator) *Service {
	return &Service{legalEntityRepo: legalEntityRepo, repo: repo, logger: logger, idGen: idGen}
}

func (s *Service) Create(ctx context.Context, legalEntityID, name string) (*domain.Fleet, error) {
	name = strings.TrimSpace(name)
	if legalEntityID == "" {
		return nil, fmt.Errorf("%w: legal_entity_id is required", domain.ErrInvalidInput)
	}
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
	}
	if _, err := s.legalEntityRepo.FindByID(ctx, legalEntityID); err != nil {
		return nil, err
	}
	id := s.idGen()
	if id == "" {
		return nil, fmt.Errorf("id generator returned empty ID")
	}
	entity := &domain.Fleet{ID: id, LegalEntityID: legalEntityID, Name: name}
	if err := s.repo.Save(ctx, entity); err != nil {
		s.logger.Error("Failed to save fleet", zap.String("id", id), zap.Error(err))
		return nil, err
	}
	s.logger.Info("Created fleet", zap.String("id", id))
	result := *entity
	return &result, nil
}

func (s *Service) Get(ctx context.Context, id string) (*domain.Fleet, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListByLegalEntity(ctx context.Context, legalEntityID string) ([]*domain.Fleet, error) {
	if legalEntityID == "" {
		return nil, fmt.Errorf("%w: legal_entity_id is required", domain.ErrInvalidInput)
	}
	return s.repo.FindByLegalEntityID(ctx, legalEntityID)
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
