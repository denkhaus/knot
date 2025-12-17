package manager

import (
	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
)

// NewService creates a new manager service instance following the DI pattern.
// This replaces the direct instantiation in favor of dependency injection.
func NewService(injector do.Injector) (ProjectManager, error) {
	// Resolve dependencies from DI container
	repo := do.MustInvoke[types.Repository](injector)
	config := config.DefaultConfig() // Use default config for now

	// Create the service with injected dependencies
	svc := newService(repo, config)
	return svc, nil
}

// ProvideService is a convenience wrapper for DI registration
func ProvideService(injector do.Injector) (ProjectManager, error) {
	return NewService(injector)
}