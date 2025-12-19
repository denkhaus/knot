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
			Name:   "server",
			Usage:  "Start the MCP server",
			Action: serverAction(),
		},
	}
}

// serverAction handles the MCP server startup using dependency injection
func serverAction() cli.ActionFunc {
	// Resolve dependencies from DI

	return func(c *cli.Context) error {
		container := shared.GetContainerFromCLIContext(c)
		configSvc := container.GetConfigService()
		loggerService := container.GetLogger()
		mcpServer := container.GetMCPServer()
		mcpConfig := configSvc.GetMCPConfig()

		loggerService.Info("Starting MCP server",
			logger.String("address", mcpConfig.Address),
			logger.Int("port", mcpConfig.Port),
			logger.String("database_backend", mcpConfig.Database.Backend),
			logger.String("postgres_endpoint", mcpConfig.Database.Endpoint))

		// Validate required parameters
		if !mcpConfig.Enabled {
			return fmt.Errorf("MCP server is disabled. Use --mcp-enabled flag or KNOT_MCP_ENABLED environment variable")
		}

		if mcpConfig.Database.Endpoint == "" {
			return fmt.Errorf("postgres-endpoint is required. Use --postgres-endpoint flag or KNOT_POSTGRES_ENDPOINT environment variable")
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
