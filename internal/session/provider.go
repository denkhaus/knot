package session

import (
	"context"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
)

// NewSessionStorageFactoryProvider creates a session storage factory with DI dependencies
func NewSessionStorageFactoryProvider(injector do.Injector) (StorageFactory, error) {
	configService := do.MustInvoke[config.Service](injector)
	loggerService := do.MustInvoke[logger.Logger](injector)

	return NewSessionStorageFactory(configService, loggerService), nil
}

// NewMemorySessionManagerProvider creates a memory-based session manager
func NewMemorySessionManagerProvider(injector do.Injector) (Manager, error) {
	return newManager(), nil // Memory manager doesn't need logger currently
}

// NewDatabaseSessionManagerProvider creates a database-backed session manager with DI dependencies
func NewDatabaseSessionManagerProvider(injector do.Injector) (Manager, error) {
	repo := do.MustInvoke[types.SessionRepository](injector)
	loggerService := do.MustInvoke[logger.Logger](injector)

	return NewDatabaseSessionManager(repo, loggerService), nil
}

// NewSessionManager creates a provider function for the SessionManager
// This follows the dependency injection pattern used throughout the application
func NewSessionManager(injector do.Injector) (Manager, error) {
	// Get the session storage factory
	factory := do.MustInvoke[StorageFactory](injector)

	// Get repository for database session creation
	repo := do.MustInvoke[types.Repository](injector)

	// Create session manager using factory
	return factory.CreateSessionManager(context.Background(), repo)
}
