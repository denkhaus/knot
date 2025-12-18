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
	"github.com/denkhaus/knot/v2/internal/mcp"
	"github.com/denkhaus/knot/v2/internal/mcp/session"
	"github.com/denkhaus/knot/v2/internal/repository"
	"github.com/denkhaus/knot/v2/internal/repository/inmemory"
	"github.com/denkhaus/knot/v2/internal/repository/postgres"
	"github.com/denkhaus/knot/v2/internal/repository/sqlite"
	"github.com/denkhaus/knot/v2/internal/templates"
	"github.com/samber/do/v2"
	"github.com/urfave/cli/v2"
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
func (c *Container) RegisterAllServices(ctx context.Context, cliCtx *cli.Context) do.Injector {
	// Register core services first (no dependencies)
	do.Provide(c.injector, configsvc.NewService(cliCtx))
	do.Provide(c.injector, logger.NewService)

	// Register repository providers with named providers using well-defined types
	do.ProvideNamedValue(c.injector, repository.SQLiteProvider.String(), sqlite.NewProvider)
	do.ProvideNamedValue(c.injector, repository.PostgreSQLProvider.String(), postgres.NewProvider)
	do.ProvideNamedValue(c.injector, repository.InMemoryProvider.String(), inmemory.NewProvider)

	// Register repository factory for mode-based repository creation
	do.Provide(c.injector, repository.NewRepositoryFactory)

	// Register repositories with named providers using well-defined types
	do.ProvideNamedValue(c.injector, repository.LocalRepository.String(), repository.NewSQLiteRepository)
	do.ProvideNamedValue(c.injector, repository.MCPRepository.String(), repository.NewPostgreSQLRepository)

	// Register the default repository based on detected mode
	do.Provide(c.injector, repository.NewModeBasedRepository)

	// Register template service (depends on repository and logger)
	do.Provide(c.injector, templates.NewService)

	// Register project manager (depends on repository and config)
	do.Provide(c.injector, manager.NewService)

	// Register session manager for MCP services
	do.Provide(c.injector, session.NewSessionManager)

	// Register MCP server
	do.Provide(c.injector, mcp.NewServer)

	return c.injector
}

// Shutdown gracefully shuts down the container and all registered services.
// This should be called during application shutdown to clean up resources.
func (c *Container) Shutdown() {
	do.Shutdown[any](c.injector)
}
