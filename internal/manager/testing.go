package manager

import (
	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/types"
)

// NewManagerWithRepository creates a manager instance with the given repository for testing purposes.
// This bypasses the dependency injection system and should only be used in tests.
func NewManagerWithRepository(repo types.Repository, cfg *config.ManagerConfig) ProjectManager {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	return &service{
		repo:   repo,
		config: cfg,
	}
}

