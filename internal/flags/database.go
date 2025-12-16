package flags

import (
	"github.com/urfave/cli/v2"
)

// PostgresEndpointFlag returns the PostgreSQL connection string flag
// This flag provides a single endpoint for all PostgreSQL configuration
// using the standard lib/pq connection string format
func PostgresEndpointFlag() cli.Flag {
	return &cli.StringFlag{
		Name:    "postgres-endpoint",
		Aliases: []string{"pg-endpoint"},
		EnvVars: []string{"KNOT_POSTGRES_ENDPOINT"},
		Usage:   "PostgreSQL connection string (e.g., postgres://user:pass@host:port/dbname?sslmode=prefer)",
	}
}

// MCPServerFlags returns flags specific to the MCP server
func MCPServerFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "address",
			Aliases: []string{"addr"},
			EnvVars: []string{"KNOT_MCP_ADDRESS"},
			Value:   "localhost",
			Usage:   "MCP server address",
		},
		&cli.IntFlag{
			Name:    "port",
			Aliases: []string{"p"},
			EnvVars: []string{"KNOT_MCP_PORT"},
			Value:   8080,
			Usage:   "MCP server port",
		},
		&cli.StringFlag{
			Name:    "log-level",
			Aliases: []string{"ll"},
			EnvVars: []string{"KNOT_MCP_LOG_LEVEL"},
			Value:   "info",
			Usage:   "Log level (debug, info, warn, error)",
		},
	}
}

// AllDatabaseFlags returns all database-related flags
// This can be used by commands that need database configuration
func AllDatabaseFlags() []cli.Flag {
	return []cli.Flag{
		PostgresEndpointFlag(),
	}
}