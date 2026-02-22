package legalentity

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
	repo   ports.LegalEntityRepository
	logger *zap.Logger
	idGen  IDGenerator
}

func New(repo ports.LegalEntityRepository, logger *zap.Logger, idGen IDGenerator) *Service {
	return &Service{repo: repo, logger: logger, idGen: idGen}
}

func (s *Service) Create(ctx context.Context, name, taxID string) (*domain.LegalEntity, error) {
	name = strings.TrimSpace(name)
	taxID = strings.TrimSpace(taxID)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
	}
	if taxID == "" {
		return nil, fmt.Errorf("%w: tax_id is required", domain.ErrInvalidInput)
	}
	id := s.idGen()
	if id == "" {
		return nil, fmt.Errorf("id generator returned empty ID")
	}
	entity := &domain.LegalEntity{ID: id, Name: name, TaxID: taxID}
	if err := s.repo.Save(ctx, entity); err != nil {
		s.logger.Error("Failed to save legal entity", zap.String("id", id), zap.Error(err))
		return nil, err
	}
	s.logger.Info("Created legal entity", zap.String("id", id))
	result := *entity
	return &result, nil
}

func (s *Service) Get(ctx context.Context, id string) (*domain.LegalEntity, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*domain.LegalEntity, error) {
	return s.repo.FindAll(ctx)
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
