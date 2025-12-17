package repository

import (
	"fmt"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/repository/sqlite"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
)

// NewSQLiteRepository provides the sqlite repository service.
func NewSQLiteRepository(injector do.Injector) (types.Repository, error) {
	loggerService := do.MustInvoke[logger.Logger](injector)
	configService := do.MustInvoke[config.Service](injector)

	// Get database path from config
	dbPath := configService.GetDatabasePath()

	// Try SQLite first
	sqliteProvider := do.MustInvoke[*sqlite.Provider](injector)
	repo, err := sqliteProvider.NewRepository(dbPath,
		sqlite.WithLogger(loggerService.ToZap()),
		sqlite.WithAutoMigrate(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SQLite repository: %w", err)
	}

	loggerService.Info("SQLite repository initialized successfully")
	return repo, nil
}
