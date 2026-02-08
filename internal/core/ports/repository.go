package ports

import (
	"context"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// FooEntityRepository is the output port for user persistence.
// Implemented by adapters/out.
type FooEntityRepository interface {
	Save(ctx context.Context, user *domain.FooEntity) error
	FindByID(ctx context.Context, id string) (*domain.FooEntity, error)
	FindAll(ctx context.Context) ([]*domain.FooEntity, error)
	Delete(ctx context.Context, id string) error
}
