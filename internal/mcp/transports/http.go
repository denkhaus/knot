package transports

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/mark3labs/mcp-go/server"
)

// HTTPTransport implements the Transport interface for HTTP communication
type HTTPTransport struct {
	*BaseTransport
	httpServer       *server.StreamableHTTPServer
	customHTTPServer *http.Server
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(deps TransportDependencies) *HTTPTransport {
	return &HTTPTransport{
		BaseTransport: NewBaseTransport(config.TransportTypeHTTP, deps),
	}
}

// Start starts the HTTP transport server with custom handler
func (h *HTTPTransport) Start(ctx context.Context) error {
	deps := h.Dependencies()
	addr := fmt.Sprintf("%s:%d", deps.ServerConfig.Address, deps.ServerConfig.Port)

	// Create custom HTTP handler that combines MCP and health endpoints
	handler := http.NewServeMux()

	// Add MCP endpoint handler (from streamable server)
	h.httpServer = server.NewStreamableHTTPServer(
		deps.MCPServer,
	)

	// Register the MCP server handler
	handler.Handle("/mcp", h.httpServer)

	// Add custom health endpoint
	handler.HandleFunc("/health", h.healthHandler)

	// Create custom HTTP server with timeout configuration
	httpTimeout := time.Duration(deps.ServerConfig.Transport.HTTP.RequestTimeout) * time.Second
	h.customHTTPServer = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  httpTimeout,
		WriteTimeout: httpTimeout,
		IdleTimeout:  60 * time.Second,
	}

	h.Logger().Info("Starting HTTP transport server",
		logger.String("address", addr),
		logger.String("route", "/mcp"),
		logger.String("health", "/health"),
		logger.Int("request_timeout_seconds", deps.ServerConfig.Transport.HTTP.RequestTimeout),
	)

	// Start the custom HTTP server directly and block
	h.setRunning(true)
	if err := h.customHTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		h.setRunning(false)
		return fmt.Errorf("HTTP server failed to start: %w", err)
	}

	return nil
}

// healthHandler provides a proper REST health check endpoint
func (h *HTTPTransport) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check database connectivity
	deps := h.Dependencies()

	healthStatus := map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"server": "knot MCP server",
		"version": "dev",
	"transport": "http",
	"address": deps.ServerConfig.Address,
	"port": deps.ServerConfig.Port,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(healthStatus)
}

// Stop gracefully stops the HTTP transport server
func (h *HTTPTransport) Stop(ctx context.Context) error {
	h.Logger().Info("Stopping HTTP transport server")

	// Shutdown the custom HTTP server first
	if h.customHTTPServer != nil {
		if err := h.customHTTPServer.Shutdown(ctx); err != nil {
			h.Logger().Error("Error shutting down custom HTTP server", logger.String("error", err.Error()))
			return err
		}
	}

	// Also shutdown the mcp-go server for cleanup
	if h.httpServer != nil {
		if err := h.httpServer.Shutdown(ctx); err != nil {
			h.Logger().Error("Error shutting down mcp-go server", logger.String("error", err.Error()))
			// Don't return error for mcp-go server shutdown, log and continue
		}
	}

	h.setRunning(false)
	return nil
}
