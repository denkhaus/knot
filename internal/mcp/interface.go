package mcp

import (
	"context"
)

// Server defines the MCP server interface for knot services.
// This abstracts MCP server operations to enable dependency injection and testing.
type Server interface {
	// Server lifecycle management
	Start() error
	Stop(ctx context.Context) error
	IsRunning() bool

	// Session management
	GetSessionCount() int
	CleanupExpiredSessions(ctx context.Context) error

	// Configuration access
	GetConfig() interface{} // Return interface{} to avoid import cycle
}
