package fooservice

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

// IDGenerator generates unique identifiers for entities.
type IDGenerator func() string

type Service struct {
	repo   ports.FooEntityRepository
	logger *zap.Logger
	idGen  IDGenerator
}

func New(repo ports.FooEntityRepository, logger *zap.Logger, idGen IDGenerator) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
		idGen:  idGen,
	}
}

func (s *Service) CreateEntity(ctx context.Context, name, description string) (*domain.FooEntity, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
	}
	if description == "" {
		return nil, fmt.Errorf("%w: description is required", domain.ErrInvalidInput)
	}

	id := s.idGen()
	if id == "" {
		// Not wrapped with domain.ErrInvalidInput: this is an infrastructure failure, not user error (maps to 500).
		return nil, errors.New("id generator returned empty ID")
	}

	entity := &domain.FooEntity{
		ID:          id,
		Name:        name,
		Description: description,
	}

	if err := s.repo.Save(ctx, entity); err != nil {
		s.logger.Error("Failed to save entity", zap.String("entityID", entity.ID), zap.Error(err))
		return nil, err
	}

	s.logger.Info("Created entity", zap.String("entityID", entity.ID))

	result := *entity
	return &result, nil
}

func (s *Service) GetEntity(ctx context.Context, id string) (*domain.FooEntity, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListEntities(ctx context.Context) ([]*domain.FooEntity, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) DeleteEntity(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: id is required", domain.ErrInvalidInput)
	}
	return s.repo.Delete(ctx, id)
}
