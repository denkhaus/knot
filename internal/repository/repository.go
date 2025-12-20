package repository

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"

	// TODO: PostgreSQL provider will be added in knot-ec0
	// "github.com/denkhaus/knot/v2/internal/repository/postgres"
	"github.com/denkhaus/knot/v2/internal/repository/sqlite"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
)

// NewSQLiteRepository provides the sqlite repository service for Local mode.
func NewSQLiteRepository(injector do.Injector) (types.Repository, error) {
	loggerService := do.MustInvoke[logger.Logger](injector)
	configService := do.MustInvoke[config.Service](injector)

	// Get database path from config
	dbPath := configService.GetDatabasePath()

	// In MCP mode, we don't use SQLite - return error to indicate this
	if dbPath == "" {
		return nil, fmt.Errorf("SQLite repository not available in MCP mode - use PostgreSQL repository instead")
	}

	// Create SQLite repository for Local mode
	sqliteProvider, _ := sqlite.NewProvider(nil)
	repo, err := sqliteProvider.NewRepository(dbPath,
		sqlite.WithLogger(loggerService.ToZap()),
		sqlite.WithAutoMigrate(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SQLite repository: %w", err)
	}

	loggerService.Info("SQLite Local Mode repository initialized successfully")
	return repo, nil
}

// NewModeBasedRepositoryProvider creates a repository using DI and the factory
// This ensures the right provider is selected by mode and all services go through DI
func NewModeBasedRepositoryProvider(injector do.Injector) (types.Repository, error) {
	configService := do.MustInvoke[config.Service](injector)
	loggerService := do.MustInvoke[logger.Logger](injector)
	factory := do.MustInvoke[Factory](injector)

	// Determine mode from config
	var mode RepositoryMode
	if configService.IsMCPMode() && configService.GetMCPConfig().Database.Endpoint != "" {
		mode = MCPMode
	} else {
		mode = LocalMode
	}

	loggerService.Info("Creating mode-based repository via DI",
		logger.String("mode", string(mode)))

	// Use the factory to create repository based on detected mode
	// The factory will internally use DI to get the appropriate provider
	return factory.CreateRepository(context.Background(), mode)
}
