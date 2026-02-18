package ports

//go:generate mockgen -destination=mocks/mock_services.go -package=mocks . FooEntityService

import (
	"context"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// FooEntityService is the input port for FooEntity operations.
// Used by adapters/in.
type FooEntityService interface {
	CreateEntity(ctx context.Context, name, description string) (*domain.FooEntity, error)
	GetEntity(ctx context.Context, id string) (*domain.FooEntity, error)
	ListEntities(ctx context.Context) ([]*domain.FooEntity, error)
	DeleteEntity(ctx context.Context, id string) error
}
