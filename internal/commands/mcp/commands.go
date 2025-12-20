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

Examples:
  knot --postgres-endpoint="postgres://user:pass@localhost:5432/knot?sslmode=prefer" mcp server
  KNOT_POSTGRES_ENDPOINT="postgres://user:pass@localhost:5432/knot" knot mcp server
  knot --mcp-endpoint 0.0.0.0 --mcp-port 9090 --postgres-endpoint="postgres://user:pass@host:5432/db?sslmode=disable" mcp server

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

		loggerService.Info("Starting MCP server",
			logger.String("address", mcpConfig.Address),
			logger.Int("port", mcpConfig.Port),
			logger.String("database_backend", mcpConfig.Database.Backend),
			logger.String("postgres_endpoint", mcpConfig.Database.Endpoint))

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
