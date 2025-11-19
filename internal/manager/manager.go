package manager

import (
	"github.com/denkhaus/knot/v2/internal/repository/inmemory"
	"github.com/denkhaus/knot/v2/internal/types"
)

// NewManager creates a new project task manager with default in-memory repository
func NewManager(config *Config) ProjectManager {
	repo := inmemory.NewMemoryRepository()
	svc := newService(repo, config)
	return svc
}

// NewManagerWithRepository creates a new project task manager with custom repository
func NewManagerWithRepository(repo types.Repository, config *Config) ProjectManager {
	svc := newService(repo, config)
	return svc
}
