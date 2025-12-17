package mcp

import (
	"fmt"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/flags"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/samber/do/v2"
	"github.com/urfave/cli/v2"
)

// Commands returns the MCP CLI commands
// TODO: No inline Action here. Use the existing pattern like Action: serverAction(appCtx),
func Commands(injector do.Injector) []*cli.Command {
	// Combine all flags
	allFlags := make([]cli.Flag, 0)
	allFlags = append(allFlags, flags.PostgresEndpointFlag())
	allFlags = append(allFlags, flags.MCPServerFlags()...)

	return []*cli.Command{
		{
			Name:  "server",
			Usage: "Start the MCP server",
			Flags: allFlags,
			// TODO: Implement actual server startup logic
			Action: serverAction(injector),
		},
	}
}

// serverAction handles the MCP server startup
// TODO: Implement actual MCP server initialization and startup
func serverAction(injector do.Injector) cli.ActionFunc {
	// Resolve dependencies from DI
	loggerService := do.MustInvoke[logger.Logger](injector)
	configSvc := do.MustInvoke[config.Service](injector)

	return func(c *cli.Context) error {
		// Initialize configuration service (detects MCP mode)
		configSvc.InitializeFromCLIContext(c)

		mcpConfig := configSvc.GetMCPConfig()

		loggerService.Info("Starting MCP server",
			logger.String("address", mcpConfig.Address),
			logger.Int("port", mcpConfig.Port),
			logger.String("log_level", mcpConfig.LogLevel),
			logger.String("database_backend", mcpConfig.Database.Backend),
			logger.String("postgres_endpoint", mcpConfig.Database.Endpoint))

		// Validate required parameters
		if mcpConfig.Database.Endpoint == "" {
			return fmt.Errorf("postgres-endpoint is required. Use --postgres-endpoint flag or KNOT_POSTGRES_ENDPOINT environment variable")
		}

		fmt.Printf("\nTODO: Implement actual server startup with PostgreSQL connection\n")
		fmt.Printf("TODO: Use configuration from config service: %+v\n", mcpConfig)
		return nil
	}
}
