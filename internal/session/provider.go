package session

import (
	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
)

// NewSessionManager creates a session manager based on the application mode
// This eliminates the need for StorageFactory abstraction
func NewSessionManager(injector do.Injector) (SessionManager, error) {
	configService := do.MustInvoke[config.Service](injector)

	if configService.IsMCPMode() {
		// MCP mode uses database-backed sessions
		repo := do.MustInvoke[types.SessionRepository](injector)
		loggerService := do.MustInvoke[logger.Logger](injector)

		return &databaseSessionManagerImpl{
			repo:   repo,
			logger: loggerService,
		}, nil
	} else {
		// Local mode uses in-memory sessions
		return newManager(), nil
	}
}
