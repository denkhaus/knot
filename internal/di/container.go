// Package di provides dependency injection infrastructure for the KNOT application.
//
// This package implements a dependency injection container using the samber/do library
// to manage service lifecycle and dependencies. It follows the provider pattern
// where each service implements a constructor that receives dependencies from the injector.
//
// Key Components:
//   - Container: Main DI container holding the injector
//   - Service Providers: Constructor functions for each service
//   - Service Registration: Ordered registration of all services
//   - Lifecycle Management: Proper shutdown of all services
//
// Cross-reference: Knot Task knot-gcp (Create DI Container Infrastructure)
package di

import (
	"context"

	configsvc "github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/repository/inmemory"
	"github.com/denkhaus/knot/v2/internal/repository/sqlite"
	"github.com/denkhaus/knot/v2/internal/templates"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
)

// Container holds all dependency injection providers and manages service lifecycle.
type Container struct {
	injector do.Injector
}

// NewContainer creates a new dependency injection container.
// Returns a properly initialized container ready for service registration.
func NewContainer() *Container {
	return &Container{
		injector: do.New(),
	}
}

// GetInjector returns the underlying injector for dependency resolution.
// This allows external code to resolve services when needed.
func (c *Container) GetInjector() do.Injector {
	return c.injector
}

// SetInjector sets the injector (used after service registration)
func (c *Container) SetInjector(injector do.Injector) {
	c.injector = injector
}

// RegisterAllServices registers all application services in the dependency injection container.
// Services are registered in dependency order to ensure proper initialization.
// Returns the injector for immediate use if needed.
func (c *Container) RegisterAllServices(ctx context.Context) do.Injector {
	// Register core services first (no dependencies)
	do.Provide(c.injector, configsvc.NewService)
	do.Provide(c.injector, logger.NewService)

	// Register repository providers
	do.Provide(c.injector, sqlite.NewProvider)
	do.Provide(c.injector, inmemory.NewProvider)

	// Register the main repository (will be resolved with fallback logic)
	do.Provide(c.injector, NewRepositoryProvider)

	// Register template service (depends on repository and logger)
	do.Provide(c.injector, templates.NewService)

	// Register project manager (depends on repository and config)
	do.Provide(c.injector, manager.NewService)

	// TODO: Register remaining services as they are implemented
	// - Session manager
	// - MCP server

	return c.injector
}

// NewRepositoryProvider provides the repository service with fallback logic.
// Attempts to use SQLite repository first, falls back to in-memory if SQLite fails.
func NewRepositoryProvider(injector do.Injector) (types.Repository, error) {
	loggerService := do.MustInvoke[logger.Logger](injector)
	configService := do.MustInvoke[configsvc.Service](injector)

	// Get database path from config
	dbPath := configService.GetDatabasePath()

	// Try SQLite first
	sqliteProvider := do.MustInvoke[*sqlite.Provider](injector)
	repo, err := sqliteProvider.NewRepository(dbPath,
		sqlite.WithLogger(loggerService.ToZap()),
		sqlite.WithAutoMigrate(true),
	)
	if err != nil {
		// Fallback to in-memory repository
		loggerService.Warn("Failed to initialize SQLite repository, falling back to in-memory",
			logger.Error(err))

		inmemoryProvider := do.MustInvoke[*inmemory.Provider](injector)
		return inmemoryProvider.NewRepository(), nil
	}

	loggerService.Info("SQLite repository initialized successfully")
	return repo, nil
}

// Shutdown gracefully shuts down the container and all registered services.
// This should be called during application shutdown to clean up resources.
func (c *Container) Shutdown() {
	do.Shutdown[any](c.injector)
}