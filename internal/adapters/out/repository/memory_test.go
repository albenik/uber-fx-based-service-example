package repository_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/albenik/uber-fx-based-service-example/internal/adapters/out/repository"
	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

var _ ports.FooEntityRepository = (*repository.MemoryFooEntityRepository)(nil)

func TestMemoryFooEntityRepository_Save(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewMemoryFooEntityRepository()

	entity := &domain.FooEntity{ID: "1", Name: "foo", Description: "bar"}
	err := repo.Save(ctx, entity)
	require.NoError(t, err)

	got, err := repo.FindByID(ctx, "1")
	require.NoError(t, err)
	assert.Equal(t, entity, got)
}

func TestMemoryFooEntityRepository_Save_DefensiveCopy(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewMemoryFooEntityRepository()

	entity := &domain.FooEntity{ID: "1", Name: "original", Description: "desc"}
	require.NoError(t, repo.Save(ctx, entity))

	// Mutate the original after saving
	entity.Name = "mutated"

	got, err := repo.FindByID(ctx, "1")
	require.NoError(t, err)
	assert.Equal(t, "original", got.Name)
}

func TestMemoryFooEntityRepository_Save_Overwrite(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewMemoryFooEntityRepository()

	entity := &domain.FooEntity{ID: "1", Name: "foo", Description: "bar"}
	require.NoError(t, repo.Save(ctx, entity))

	updated := &domain.FooEntity{ID: "1", Name: "updated", Description: "updated"}
	require.NoError(t, repo.Save(ctx, updated))

	got, err := repo.FindByID(ctx, "1")
	require.NoError(t, err)
	assert.Equal(t, "updated", got.Name)
}

func TestMemoryFooEntityRepository_FindByID_NotFound(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewMemoryFooEntityRepository()

	result, err := repo.FindByID(ctx, "nonexistent")
	require.ErrorIs(t, err, domain.ErrEntityNotFound)
	assert.Nil(t, result)
}

func TestMemoryFooEntityRepository_FindAll_Empty(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewMemoryFooEntityRepository()

	entities, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Empty(t, entities)
}

func TestMemoryFooEntityRepository_FindAll(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewMemoryFooEntityRepository()

	e1 := &domain.FooEntity{ID: "1", Name: "foo", Description: "bar"}
	e2 := &domain.FooEntity{ID: "2", Name: "baz", Description: "qux"}
	require.NoError(t, repo.Save(ctx, e1))
	require.NoError(t, repo.Save(ctx, e2))

	entities, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, entities, 2)
}

func TestMemoryFooEntityRepository_Delete(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewMemoryFooEntityRepository()

	entity := &domain.FooEntity{ID: "1", Name: "foo", Description: "bar"}
	require.NoError(t, repo.Save(ctx, entity))

	err := repo.Delete(ctx, "1")
	require.NoError(t, err)

	result, err := repo.FindByID(ctx, "1")
	require.ErrorIs(t, err, domain.ErrEntityNotFound)
	assert.Nil(t, result)
}

func TestMemoryFooEntityRepository_Delete_NotFound(t *testing.T) {
	ctx := t.Context()
	repo := repository.NewMemoryFooEntityRepository()

	err := repo.Delete(ctx, "nonexistent")
	assert.ErrorIs(t, err, domain.ErrEntityNotFound)
}

func TestMemoryFooEntityRepository_ConcurrentAccess(t *testing.T) {
	repo := repository.NewMemoryFooEntityRepository()

	const goroutines = 50

	for i := range goroutines {
		t.Run(fmt.Sprintf("goroutine-%d", i), func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			id := fmt.Sprintf("entity-%d", i)
			entity := &domain.FooEntity{ID: id, Name: "name", Description: "desc"}

			require.NoError(t, repo.Save(ctx, entity))

			got, err := repo.FindByID(ctx, id)
			require.NoError(t, err)
			assert.Equal(t, id, got.ID)

			entities, err := repo.FindAll(ctx)
			assert.NoError(t, err)
			assert.NotNil(t, entities)

			assert.NoError(t, repo.Delete(ctx, id))
		})
	}

	// After all parallel subtests complete, verify the repository is empty.
	t.Run("verify-empty", func(t *testing.T) {
		entities, err := repo.FindAll(t.Context())
		require.NoError(t, err)
		assert.Empty(t, entities)
	})
}
