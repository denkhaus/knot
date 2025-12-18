package repository

import (
	"context"

	"github.com/denkhaus/knot/v2/internal/types"
)

// Factory defines the interface for creating repositories based on mode
// This enables dependency injection and testing
type Factory interface {
	// CreateRepository creates a repository based on the specified mode
	CreateRepository(ctx context.Context, mode RepositoryMode) (types.Repository, error)
}
