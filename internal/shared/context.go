// Package shared provides common application context and utilities for KNOT.
//
// This package contains shared application context structures, logging configuration,
// CLI flags, and utility functions used across different components of the KNOT
// project management system. It helps avoid import cycles and provides centralized
// access to common dependencies.
//
// Key Components:
//   - AppContext: Holds application dependencies and configuration
//   - Logger: Provides structured logging throughout the application
//   - CLI Flags: Common command-line flag definitions
//
// Cross-reference: Knot Task 86f3ba2d-3a87-493b-b8fc-96d19f344e89
package shared

import (
	"github.com/denkhaus/knot/v2/internal/manager"
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

// SetActor sets the current actor/user for the context
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
