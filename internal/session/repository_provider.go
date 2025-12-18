package session

import (
	"fmt"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
)

// NewSessionRepositoryProvider creates a session repository based on the mode
// This provider will only return a SessionRepository for MCP mode (PostgreSQL)
// Local mode (SQLite) does not support sessions
func NewSessionRepositoryProvider(injector do.Injector) (types.SessionRepository, error) {
	repo := do.MustInvoke[types.Repository](injector)
	loggerService := do.MustInvoke[logger.Logger](injector)
	configService := do.MustInvoke[config.Service](injector)

	// Check if we're in MCP mode with PostgreSQL
	if configService.IsMCPMode() && configService.GetMCPConfig().Database.Endpoint != "" {
		// Try to assert to SessionRepository interface
		if sessionRepo, ok := repo.(types.SessionRepository); ok {
			loggerService.Info("Session repository created for PostgreSQL (MCP mode)")
			return sessionRepo, nil
		}

		loggerService.Error("Repository does not implement SessionRepository interface")
		return nil, fmt.Errorf("repository does not support session operations in MCP mode")
	}

	// For non-MCP mode, sessions are not supported
	loggerService.Info("Session repository not needed for local mode")
	return nil, fmt.Errorf("session repository is only available in MCP mode")
}
