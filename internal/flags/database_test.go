package flags

import (
	"testing"

	"github.com/urfave/cli/v2"
)

func TestPostgresEndpointFlag(t *testing.T) {
	flag := PostgresEndpointFlag()

	// Type assert to access StringFlag properties
	stringFlag, ok := flag.(*cli.StringFlag)
	if !ok {
		t.Fatal("Flag should be a StringFlag")
	}

	// Test flag properties
	if stringFlag.Name != "postgres-endpoint" {
		t.Errorf("Expected flag name to be 'postgres-endpoint', got %s", stringFlag.Name)
	}

	if len(stringFlag.Aliases) != 1 || stringFlag.Aliases[0] != "pg-endpoint" {
		t.Errorf("Expected alias to be ['pg-endpoint'], got %v", stringFlag.Aliases)
	}

	if len(stringFlag.EnvVars) != 1 || stringFlag.EnvVars[0] != "KNOT_POSTGRES_ENDPOINT" {
		t.Errorf("Expected EnvVars to be ['KNOT_POSTGRES_ENDPOINT'], got %v", stringFlag.EnvVars)
	}

	// Check that usage description contains example
	usage := stringFlag.Usage
	if usage == "" {
		t.Error("Usage description should not be empty")
	}
}

func TestMCPServerFlags(t *testing.T) {
	flags := MCPServerFlags()

	if len(flags) != 3 {
		t.Errorf("Expected 3 MCP server flags, got %d", len(flags))
	}

	// Test address flag
	addressFlag, ok := flags[0].(*cli.StringFlag)
	if !ok || addressFlag.Name != "address" {
		t.Errorf("Expected first flag to be 'address', got %v", flags[0])
	}

	// Test port flag
	portFlag, ok := flags[1].(*cli.IntFlag)
	if !ok || portFlag.Name != "port" {
		t.Errorf("Expected second flag to be 'port', got %v", flags[1])
	}

	// Test log-level flag
	logLevelFlag, ok := flags[2].(*cli.StringFlag)
	if !ok || logLevelFlag.Name != "log-level" {
		t.Errorf("Expected third flag to be 'log-level', got %v", flags[2])
	}
}

func TestAllDatabaseFlags(t *testing.T) {
	flags := AllDatabaseFlags()

	if len(flags) != 1 {
		t.Errorf("Expected 1 database flag, got %d", len(flags))
	}

	// Should only contain postgres-endpoint flag
	dbFlag, ok := flags[0].(*cli.StringFlag)
	if !ok || dbFlag.Name != "postgres-endpoint" {
		t.Errorf("Expected database flag to be 'postgres-endpoint', got %v", flags[0])
	}
}

