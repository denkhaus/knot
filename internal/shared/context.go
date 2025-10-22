package shared

import (
	"github.com/denkhaus/knot/internal/manager"
	"go.uber.org/zap"
)

// AppContext holds the application dependencies
// This is in a shared package to avoid import cycles
type AppContext struct {
	ProjectManager manager.ProjectManager
	Logger         *zap.Logger
	Actor          string
}

// NewAppContext creates a new application context with all dependencies
func NewAppContext(projectManager manager.ProjectManager, logger *zap.Logger) *AppContext {
	return &AppContext{
		ProjectManager: projectManager,
		Logger:         logger,
	}
}

func (p *AppContext) SetActor(actor string) {
	p.Actor = actor
}

// GetActor returns the current actor
func (p *AppContext) GetActor() string {
	if p.Actor == "" {
		return "unknown"
	}
	return p.Actor
}
