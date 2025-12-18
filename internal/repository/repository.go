package repository

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/repository/postgres"
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

	// Create SQLite repository for Local mode
	sqliteProvider := &sqlite.Provider{}
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

// NewPostgreSQLRepository provides the PostgreSQL repository service for MCP mode.
func NewPostgreSQLRepository(injector do.Injector) (types.Repository, error) {
	loggerService := do.MustInvoke[logger.Logger](injector)
	configService := do.MustInvoke[config.Service](injector)

	// Get PostgreSQL DSN from MCP config
	dsn := configService.GetMCPConfig().Database.Endpoint
	if dsn == "" {
		return nil, fmt.Errorf("PostgreSQL DSN not configured for MCP Mode")
	}

	// Create PostgreSQL repository for MCP mode
	postgresProvider := &postgres.Provider{}
	repo, err := postgresProvider.NewRepository(dsn,
		postgres.WithLogger(loggerService.ToZap()),
		postgres.WithAutoMigrate(true),
		postgres.WithConnectionPool(25, 5),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL repository: %w", err)
	}

	loggerService.Info("PostgreSQL MCP Mode repository initialized successfully")
	return repo, nil
}

// NewModeBasedRepository creates a repository based on the detected mode (Local vs MCP).
func NewModeBasedRepository(injector do.Injector) (types.Repository, error) {
	configService := do.MustInvoke[config.Service](injector)
	loggerService := do.MustInvoke[logger.Logger](injector)

	// Determine mode from config
	var mode RepositoryMode
	if configService.IsMCPMode() && configService.GetMCPConfig().Database.Endpoint != "" {
		mode = MCPMode
	} else {
		mode = LocalMode
	}

	loggerService.Info("Creating mode-based repository",
		logger.String("mode", string(mode)))

	// Create repository factory
	factory := &RepositoryFactory{
		configService: configService,
		logger:        loggerService,
	}

	// Create repository based on detected mode
	return factory.CreateRepository(context.Background(), mode)
}
