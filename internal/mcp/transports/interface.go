package transports

import (
	"context"

	"github.com/denkhaus/knot/v2/internal/config"
)

// Transport defines the interface for MCP server transport implementations
// Different transports (stdio, HTTP, SSE) implement this interface
type Transport interface {
	// Start starts the transport server
	Start(ctx context.Context) error

	// Stop gracefully stops the transport server
	Stop(ctx context.Context) error

	// IsRunning returns true if the transport is currently running
	IsRunning() bool

	// GetType returns the transport type (stdio, http, sse)
	GetType() config.TransportType
}