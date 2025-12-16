package mcp

import (
	"fmt"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/flags"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/urfave/cli/v2"
)

// Commands returns the MCP CLI commands
// TODO: No inline Action here. Use the existing pattern like Action: serverAction(appCtx),
func Commands(appCtx *shared.AppContext) []*cli.Command {
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
			Action: serverAction(appCtx),
		},
	}
}

// serverAction handles the MCP server startup
// TODO: Implement actual MCP server initialization and startup
func serverAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		// Initialize configuration service (detects MCP mode)
		config.InitializeFromCLIContext(c)

		// Get configuration from service using CLI context
		configSvc := config.GetConfigService()
		mcpConfig := configSvc.GetMCPConfig(c)

		fmt.Printf("MCP Server Configuration:\n")
		fmt.Printf("  Address: %s\n", mcpConfig.Address)
		fmt.Printf("  Port: %d\n", mcpConfig.Port)
		fmt.Printf("  Log Level: %s\n", mcpConfig.LogLevel)
		fmt.Printf("  Database Backend: %s\n", mcpConfig.Database.Backend)
		fmt.Printf("  PostgreSQL Endpoint: %s\n", mcpConfig.Database.Endpoint)

		// Validate required parameters
		if mcpConfig.Database.Endpoint == "" {
			return fmt.Errorf("postgres-endpoint is required. Use --postgres-endpoint flag or KNOT_POSTGRES_ENDPOINT environment variable")
		}

		fmt.Printf("\nTODO: Implement actual server startup with PostgreSQL connection\n")
		fmt.Printf("TODO: Use configuration from config service: %+v\n", mcpConfig)
		return nil
	}
}
