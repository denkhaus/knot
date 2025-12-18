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

// RepositoryMode defines the different repository modes
type RepositoryMode string

const (
	// LocalMode uses SQLite database in .knot directory (for CLI)
	LocalMode RepositoryMode = "local"
	// MCPMode uses PostgreSQL database (for MCP server)
	MCPMode RepositoryMode = "mcp"
)

// RepositoryFactory creates repositories based on the mode
type RepositoryFactory struct {
	configService config.Service
	logger        logger.Logger
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(injector do.Injector) (*RepositoryFactory, error) {
	configService := do.MustInvoke[config.Service](injector)
	loggerService := do.MustInvoke[logger.Logger](injector)

	return &RepositoryFactory{
		configService: configService,
		logger:        loggerService,
	}, nil
}

// CreateRepository creates a repository based on the specified mode
func (f *RepositoryFactory) CreateRepository(ctx context.Context, mode RepositoryMode) (types.Repository, error) {
	switch mode {
	case LocalMode:
		return f.createLocalRepository(ctx)
	case MCPMode:
		return f.createMCPRepository(ctx)
	default:
		return nil, fmt.Errorf("unsupported repository mode: %s", mode)
	}
}

// createLocalRepository creates a SQLite repository for Local mode
func (f *RepositoryFactory) createLocalRepository(ctx context.Context) (types.Repository, error) {
	f.logger.Info("Creating Local Mode repository (SQLite)")

	// Get database path from config (default: .knot directory)
	dbPath := f.configService.GetDatabasePath()

	// Create SQLite repository with local optimizations
	sqliteProvider := &sqlite.Provider{}
	repo, err := sqliteProvider.NewRepository(dbPath,
		sqlite.WithLogger(f.logger.ToZap()),
		sqlite.WithAutoMigrate(true),
		sqlite.WithConnectionPool(10, 5), // CLI typically has lower concurrency needs
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Local Mode SQLite repository: %w", err)
	}

	f.logger.Info("Local Mode repository created successfully",
		logger.String("database_path", dbPath))

	return repo, nil
}

// createMCPRepository creates a PostgreSQL repository for MCP mode
func (f *RepositoryFactory) createMCPRepository(ctx context.Context) (types.Repository, error) {
	f.logger.Info("Creating MCP Mode repository (PostgreSQL)")

	// Get PostgreSQL DSN from config
	dsn := f.configService.GetMCPConfig().Database.Endpoint
	if dsn == "" {
		return nil, fmt.Errorf("PostgreSQL DSN not configured for MCP Mode")
	}

	// Create PostgreSQL provider
	postgresProvider := &postgres.Provider{}
	repo, err := postgresProvider.NewRepository(dsn,
		postgres.WithLogger(f.logger.ToZap()),
		postgres.WithAutoMigrate(true),
		postgres.WithConnectionPool(25, 5), // Higher concurrency for MCP server
		postgres.WithMigrationTimeout(f.configService.GetMCPConfig().Session.Timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP Mode PostgreSQL repository: %w", err)
	}

	f.logger.Info("MCP Mode repository created successfully",
		logger.String("dsn", "postgresql://***")) // Mask sensitive DSN

	return repo, nil
}

// GetModeFromConfig determines the repository mode from configuration
func (f *RepositoryFactory) GetModeFromConfig() RepositoryMode {
	// If MCP is enabled with PostgreSQL DSN, default to MCP mode
	// Otherwise, use Local mode
	if f.configService.IsMCPMode() && f.configService.GetMCPConfig().Database.Endpoint != "" {
		return MCPMode
	}
	return LocalMode
}

// CreateRepositoryWithDI creates a repository using DI to get the appropriate provider
// This ensures all services go through the DI system and providers are reusable
func (f *RepositoryFactory) CreateRepositoryWithDI(injector do.Injector, mode RepositoryMode) (types.Repository, error) {
	switch mode {
	case LocalMode:
		return f.createLocalRepositoryWithDI(injector)
	case MCPMode:
		return f.createMCPRepositoryWithDI(injector)
	default:
		return nil, fmt.Errorf("unsupported repository mode: %s", mode)
	}
}

// createLocalRepositoryWithDI creates a SQLite repository using DI providers
func (f *RepositoryFactory) createLocalRepositoryWithDI(injector do.Injector) (types.Repository, error) {
	f.logger.Info("Creating Local Mode repository via DI (SQLite)")

	// Get database path from config
	dbPath := f.configService.GetDatabasePath()

	// Get SQLite provider from DI (contains reusable Ent ORM code)
	sqliteProvider := do.MustInvokeNamed[*sqlite.Provider](injector, SQLiteProvider.String())

	// Create repository using the provider
	repo, err := sqliteProvider.NewRepository(dbPath,
		sqlite.WithLogger(f.logger.ToZap()),
		sqlite.WithAutoMigrate(true),
		sqlite.WithConnectionPool(10, 5), // CLI typically has lower concurrency needs
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Local Mode SQLite repository: %w", err)
	}

	f.logger.Info("Local Mode repository created successfully via DI",
		logger.String("database_path", dbPath))

	return repo, nil
}

// createMCPRepositoryWithDI creates a PostgreSQL repository using DI providers
func (f *RepositoryFactory) createMCPRepositoryWithDI(injector do.Injector) (types.Repository, error) {
	f.logger.Info("Creating MCP Mode repository via DI (PostgreSQL)")

	// Get PostgreSQL DSN from config
	dsn := f.configService.GetMCPConfig().Database.Endpoint
	if dsn == "" {
		return nil, fmt.Errorf("PostgreSQL DSN not configured for MCP Mode")
	}

	// Get PostgreSQL provider from DI (contains reusable Ent ORM code)
	postgresProvider := do.MustInvokeNamed[*postgres.Provider](injector, PostgreSQLProvider.String())

	// Create repository using the provider with DI dependencies
	repo, err := postgresProvider.NewRepository(dsn,
		postgres.WithLogger(f.logger.ToZap()),
		postgres.WithAutoMigrate(true),
		postgres.WithConnectionPool(25, 5), // Higher concurrency for MCP server
		postgres.WithMigrationTimeout(f.configService.GetMCPConfig().Session.Timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP Mode PostgreSQL repository: %w", err)
	}

	f.logger.Info("MCP Mode repository created successfully via DI")
	return repo, nil
}

// DetectMode attempts to auto-detect the appropriate mode
func (f *RepositoryFactory) DetectMode() (RepositoryMode, error) {
	// Priority: 1. Explicit config, 2. Environment, 3. Default to Local

	// Check if PostgreSQL is explicitly configured
	if f.configService.IsMCPMode() && f.configService.GetMCPConfig().Database.Endpoint != "" {
		return MCPMode, nil
	}

	// Default to Local mode for CLI usage
	return LocalMode, nil
}

// IsMCPMode checks if the given mode is MCP mode
func IsMCPMode(mode RepositoryMode) bool {
	return mode == MCPMode
}

// IsLocalMode checks if the given mode is Local mode
func IsLocalMode(mode RepositoryMode) bool {
	return mode == LocalMode
}
