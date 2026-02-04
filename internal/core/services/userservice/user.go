package userservice

import (
	"context"

	"go.uber.org/zap"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

type Service struct {
	repo   ports.UserRepository
	logger *zap.Logger
}

func New(repo ports.UserRepository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) CreateUser(ctx context.Context, name, email string) (*domain.User, error) {
	user := &domain.User{
		ID:    "generated-id", // In a real implementation, generate a unique ID
		Name:  name,
		Email: email,
	}

	if err := s.repo.Save(ctx, user); err != nil {
		return nil, err
	}

	s.logger.Info("Created user", zap.String("userID", user.ID))

	return user, nil
}

func (s *Service) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) DeleteUser(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
