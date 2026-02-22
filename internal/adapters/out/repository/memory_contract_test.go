package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/albenik/uber-fx-based-service-example/internal/adapters/out/repository"
	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
	"github.com/albenik/uber-fx-based-service-example/internal/core/ports"
)

var _ ports.ContractRepository = (*repository.MemoryContractRepository)(nil)

func TestMemoryContractRepository_FindOverlapping(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryContractRepository()

	// Existing: Jan 1 - Jan 31
	c1 := &domain.Contract{
		ID: "c1", DriverID: "d1", LegalEntityID: "le1", FleetID: "f1",
		StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
	}
	require.NoError(t, repo.Save(ctx, c1))

	// Overlapping: Jan 15 - Feb 15
	overlapping, err := repo.FindOverlapping(ctx, "d1", "le1", "f1",
		time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 2, 15, 0, 0, 0, 0, time.UTC), "")
	require.NoError(t, err)
	assert.Len(t, overlapping, 1)

	// Non-overlapping: Feb 1 - Feb 28
	none, err := repo.FindOverlapping(ctx, "d1", "le1", "f1",
		time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC), "")
	require.NoError(t, err)
	assert.Empty(t, none)
}
