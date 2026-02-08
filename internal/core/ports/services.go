package ports

import (
	"context"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// FooEntityService is the input port for user operations.
// Used by adapters/in.
type FooEntityService interface {
	CreateEntity(ctx context.Context, name, email string) (*domain.FooEntity, error)
	GetEntity(ctx context.Context, id string) (*domain.FooEntity, error)
	ListEntities(ctx context.Context) ([]*domain.FooEntity, error)
	DeleteEntity(ctx context.Context, id string) error
}
