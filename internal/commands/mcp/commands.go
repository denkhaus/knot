package mcp

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/urfave/cli/v2"
)

// Commands returns the MCP CLI commands
func Commands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "server",
			Usage: "Start the MCP server",
			Description: `Start the Model Context Protocol (MCP) server for multi-project task management.

Transport modes and connection examples are displayed on startup.

Examples:
  knot mcp server                                    # Start with default settings
  knot --mcp-transport-mode http mcp server         # Start with HTTP transport
  knot --mcp-address 0.0.0.0 --mcp-port 9090 mcp server  # Custom address/port

PostgreSQL Connection String Format:
  postgres://[user[:password]@][netloc][:port][/dbname][?param1=value1&...]

Common parameters:
  - sslmode: disable, require, verify-ca, verify-full (default: prefer)
  - connect_timeout: connection timeout in seconds (default: 5)
  - statement_timeout: query timeout in seconds

Example connection strings:
  - postgres://user:password@localhost:5432/knot?sslmode=prefer
  - postgres:///knot?host=/var/run/postgresql  (Unix socket)
  - postgres://localhost:5432/knot?sslmode=disable (local development)`,
			Action: serverAction(),
		},
	}
}

// serverAction handles the MCP server startup using dependency injection
func serverAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		container := shared.GetContainerFromContext(c)
		configSvc, err := container.GetConfigService()
		if err != nil {
			return fmt.Errorf("Configuration error: %w\n\nUsage examples:\n  knot --postgres-endpoint=\"postgres://user:pass@localhost:5432/knot?sslmode=prefer\" mcp server\n  KNOT_POSTGRES_ENDPOINT=\"postgres://user:pass@localhost:5432/knot\" knot mcp server", err)
		}

		loggerService := container.GetLogger()
		mcpServer, err := container.GetMCPServer()
		if err != nil {
			return fmt.Errorf("Failed to initialize MCP server: %w\n\nThis usually means the PostgreSQL connection is invalid.\nPlease check:\n1. PostgreSQL server is running\n2. Connection string is correct\n3. Database exists and is accessible", err)
		}
		mcpConfig := configSvc.GetMCPConfig()

		// Log MCP server configuration (excluding sensitive connection string)
		loggerService.Info("Starting MCP server",
			logger.String("transport_mode", string(mcpConfig.Transport.Mode)),
			logger.String("address", mcpConfig.Address),
			logger.Int("port", mcpConfig.Port),
			logger.String("database_backend", mcpConfig.Database.Backend))

		// Display transport information and connection details
		loggerService.Info("MCP Transport Information:")
		switch mcpConfig.Transport.Mode {
		case "http":
			endpoint := fmt.Sprintf("http://%s:%d/mcp", mcpConfig.Address, mcpConfig.Port)
			healthEndpoint := fmt.Sprintf("http://%s:%d/health", mcpConfig.Address, mcpConfig.Port)
			loggerService.Info("HTTP transport enabled",
				logger.String("endpoint", endpoint),
				logger.String("health_check", healthEndpoint))
			loggerService.Info("Connection examples:",
				logger.String("mcp", fmt.Sprintf("curl -X POST %s -d '{\"jsonrpc\":\"2.0\",...}'", endpoint)),
				logger.String("health", fmt.Sprintf("curl %s", healthEndpoint)))
		case "stdio":
			loggerService.Info("STDIO transport enabled",
				logger.String("communication", "stdin/stdout"))
			loggerService.Info("Connection example:",
				logger.String("stdio", "echo '{\"jsonrpc\":\"2.0\",...}' | knot mcp server"))
		case "sse":
			endpoint := fmt.Sprintf("http://%s:%d/sse", mcpConfig.Address, mcpConfig.Port)
			loggerService.Info("SSE transport enabled",
				logger.String("endpoint", endpoint))
			loggerService.Info("Connection example:",
				logger.String("sse", fmt.Sprintf("curl -N %s", endpoint)))
		}

		// Start the MCP server
		loggerService.Info("Initializing MCP server with DI dependencies")

		if err := mcpServer.Start(); err != nil {
			return fmt.Errorf("failed to start MCP server: %w", err)
		}

		loggerService.Info("MCP server started successfully")

		// Setup graceful shutdown
		ctx := context.Background()
		defer func() {
			if err := mcpServer.Stop(ctx); err != nil {
				loggerService.Error("Error stopping MCP server", logger.Error(err))
			}
		}()

		return nil
	}
}
