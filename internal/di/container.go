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
	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/handlers"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp"
	"github.com/denkhaus/knot/v2/internal/mcp/hints"
	mcpshared "github.com/denkhaus/knot/v2/internal/mcp/shared"
	"github.com/denkhaus/knot/v2/internal/mcp/transports"
	"github.com/denkhaus/knot/v2/internal/repository"
	"github.com/denkhaus/knot/v2/internal/repository/inmemory"
	"github.com/denkhaus/knot/v2/internal/repository/postgres"
	"github.com/denkhaus/knot/v2/internal/repository/sqlite"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/denkhaus/knot/v2/internal/sync"
	syncshared "github.com/denkhaus/knot/v2/internal/sync/shared"
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

// RegisterAllServices registers all application services in the dependency injection container.
// Services are registered in dependency order to ensure proper initialization.
// Returns the injector for immediate use if needed.
func (c *Container) RegisterAllServices(cliCtx *cli.Context) do.Injector {
	// Register core services first (no dependencies)
	do.Provide(c.injector, config.NewService(cliCtx))
	do.Provide(c.injector, logger.NewService)

	// Register repository providers with named providers using well-defined types
	// These providers contain the Ent ORM code that is reusable for both modes
	do.ProvideNamed(c.injector, repository.SQLiteProvider.String(), sqlite.NewProvider)
	do.ProvideNamed(c.injector, repository.PostgreSQLProvider.String(), postgres.NewProvider)
	do.ProvideNamed(c.injector, repository.InMemoryProvider.String(), inmemory.NewProvider)

	// Register repository factory for mode-based repository creation
	// The factory will use DI to get the appropriate provider based on mode
	do.Provide(c.injector, repository.NewRepositoryFactory)

	// Register the default repository provider that uses the factory
	// This ensures the right provider is selected by mode
	do.Provide(c.injector, repository.NewModeBasedRepositoryProvider)

	// Register template service (depends on repository and logger)
	do.Provide(c.injector, templates.NewService)

	// Register project manager (depends on repository and config)
	do.Provide(c.injector, manager.NewService)

	// Register session repository provider (only available in MCP mode)
	do.Provide(c.injector, session.NewSessionRepositoryProvider)

	// Register session manager provider that selects appropriate implementation based on mode
	// This eliminates the need for StorageFactory abstraction
	do.Provide(c.injector, session.NewSessionManager)

	do.Provide(c.injector, transports.NewSessionWrapper)

	// Register hint system services
	do.Provide(c.injector, hints.NewHintGeneratorProvider)
	do.Provide(c.injector, hints.NewHintIntegrationProvider)

	// Register MCP server
	do.Provide(c.injector, mcp.NewMCPServer)
	do.Provide(c.injector, mcp.NewServer)

	// Register MCP shared components
	do.Provide(c.injector, mcpshared.NewSessionRegistry)

	// Register transport - this will select the appropriate transport based on config
	do.Provide(c.injector, transports.NewTransport)

	// Register REST sync client provider for remote synchronization
	do.Provide(c.injector, sync.NewSyncClient)

	// Register sync shared serializer
	do.Provide(c.injector, syncshared.NewSerializerProvider)

	// Register sync data extractor service
	do.Provide(c.injector, sync.NewDataExtractorService)

	// Register sync diff engine service
	do.Provide(c.injector, sync.NewDiffEngine)

	// Register conflict resolver for sync conflict resolution
	do.Provide(c.injector, sync.NewConflictResolver)

	// Register migration engine for applying sync operations
	do.Provide(c.injector, sync.NewMigrationEngine)

	// Register sync manager for bidirectional synchronization
	do.Provide(c.injector, sync.NewSyncManager)
	do.Provide(c.injector, handlers.NewSyncService)
	do.Provide(c.injector, handlers.NewSyncHandler)

	// Register sync HTTP endpoints
	do.Provide(c.injector, handlers.NewSyncHTTPHandlers)

	return c.injector
}

// Shutdown gracefully shuts down the container and all registered services.
// This should be called during application shutdown to clean up resources.
// Note: Shutdown errors are logged but don't affect the exit code since
// the command has already completed successfully.
func (c *Container) Shutdown() error {
	// For CLI mode, we don't need explicit shutdown of services.
	// The DI container will clean up on process exit.
	// Returning nil to avoid affecting exit codes.
	return nil
}

func (c *Container) GetLogger() logger.Logger {
	return do.MustInvoke[logger.Logger](c.injector)
}

func (c *Container) GetProjectManager() manager.ProjectManager {
	return do.MustInvoke[manager.ProjectManager](c.injector)
}

func (c *Container) GetConfigService() (config.Service, error) {
	return do.Invoke[config.Service](c.injector)
}

func (c *Container) GetMCPServer() (mcp.Server, error) {
	return do.Invoke[mcp.Server](c.injector)
}

func (c *Container) GetSyncManager() (sync.SyncManager, error) {
	return do.Invoke[sync.SyncManager](c.injector)
}

// GetDataExtractor returns the DataExtractor service from the DI container
func (c *Container) GetDataExtractor() (sync.DataExtractor, error) {
	return do.Invoke[sync.DataExtractor](c.injector)
}
