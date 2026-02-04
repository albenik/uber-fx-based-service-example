package ports

import (
	"context"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// UserRepository is the output port for user persistence.
// Implemented by adapters/out.
type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindAll(ctx context.Context) ([]*domain.User, error)
	Delete(ctx context.Context, id string) error
}
