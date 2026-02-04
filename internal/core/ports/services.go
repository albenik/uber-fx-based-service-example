package ports

import (
	"context"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// UserService is the input port for user operations.
// Used by adapters/in.
type UserService interface {
	CreateUser(ctx context.Context, name, email string) (*domain.User, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	ListUsers(ctx context.Context) ([]*domain.User, error)
	DeleteUser(ctx context.Context, id string) error
}
