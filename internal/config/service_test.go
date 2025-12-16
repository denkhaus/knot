package config

import (
	"os"
	"sync"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestConfigServiceSingleton(t *testing.T) {
	// Reset singleton for testing
	instance = nil
	once = sync.Once{}

	svc1 := GetConfigService()
	svc2 := GetConfigService()

	if svc1 != svc2 {
		t.Error("GetConfigService should return the same singleton instance")
	}
}

func TestInitializeFromCLIContext(t *testing.T) {
	// Reset singleton for testing
	instance = nil
	once = sync.Once{}

	// Save original os.Args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set os.Args to simulate the full command path (flags before subcommands)
	os.Args = []string{"knot", "--address", "0.0.0.0", "--port", "9090", "mcp", "server"}

	// Create flags with defaults
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "address",
			EnvVars: []string{"KNOT_MCP_ADDRESS"},
			Value:   "localhost",
		},
		&cli.IntFlag{
			Name:    "port",
			EnvVars: []string{"KNOT_MCP_PORT"},
			Value:   8080,
		},
		&cli.StringFlag{
			Name:    "postgres-endpoint",
			EnvVars: []string{"KNOT_POSTGRES_ENDPOINT"},
		},
	}

	// Create a complete app structure that matches real usage
	app := &cli.App{
		Name:  "knot",
		Flags: flags,
		Action: func(c *cli.Context) error {
			// Initialize configuration (detects MCP mode)
			InitializeFromCLIContext(c)

			// Test that MCP mode was detected
			svc := GetConfigService()
			if !svc.IsMCPMode() {
				t.Error("Should be in MCP mode when args contain 'mcp'")
			}

			// Test that configuration comes from CLI context
			config := svc.GetMCPConfig(c)
			if config.Address != "0.0.0.0" {
				t.Errorf("Expected address '0.0.0.0', got %s", config.Address)
			}
			if config.Port != 9090 {
				t.Errorf("Expected port 9090, got %d", config.Port)
			}
			return nil
		},
	}

	// Run the app
	err := app.Run(os.Args)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEnvironmentVariableHandling(t *testing.T) {
	// Reset singleton for testing
	instance = nil
	once = sync.Once{}

	// Save original os.Args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set os.Args to simulate the command (flags before subcommands)
	os.Args = []string{"knot", "mcp", "server"}

	// Set environment variables
	os.Setenv("KNOT_MCP_ADDRESS", "env.example.com")
	os.Setenv("KNOT_MCP_PORT", "9999")
	os.Setenv("KNOT_POSTGRES_ENDPOINT", "postgres://env:pass@host/db")
	defer func() {
		os.Unsetenv("KNOT_MCP_ADDRESS")
		os.Unsetenv("KNOT_MCP_PORT")
		os.Unsetenv("KNOT_POSTGRES_ENDPOINT")
	}()

	// Create flags with defaults and env var support
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "address",
			EnvVars: []string{"KNOT_MCP_ADDRESS"},
			Value:   "localhost",
		},
		&cli.IntFlag{
			Name:    "port",
			EnvVars: []string{"KNOT_MCP_PORT"},
			Value:   8080,
		},
		&cli.StringFlag{
			Name:    "postgres-endpoint",
			EnvVars: []string{"KNOT_POSTGRES_ENDPOINT"},
		},
	}

	// Create app
	app := &cli.App{
		Name:  "knot",
		Flags: flags,
		Action: func(c *cli.Context) error {
			// Initialize configuration
			InitializeFromCLIContext(c)

			svc := GetConfigService()
			config := svc.GetMCPConfig(c)

			// urfave/cli/v2 should prioritize environment vars when CLI flags not set
			if config.Address != "env.example.com" {
				t.Errorf("Expected address from env var 'env.example.com', got %s", config.Address)
			}
			if config.Port != 9999 {
				t.Errorf("Expected port from env var 9999, got %d", config.Port)
			}
			if config.Database.Endpoint != "postgres://env:pass@host/db" {
				t.Errorf("Expected endpoint from env var 'postgres://env:pass@host/db', got %s", config.Database.Endpoint)
			}
			return nil
		},
	}

	// Run the app
	err := app.Run(os.Args)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPriorityCLIOverEnvVars(t *testing.T) {
	// Reset singleton for testing
	instance = nil
	once = sync.Once{}

	// Save original os.Args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Set os.Args to simulate the command with CLI flags (flags before subcommands)
	os.Args = []string{"knot", "--address", "cli.example.com", "--port", "7777", "mcp", "server"}

	// Set environment variables (should be overridden by CLI flags)
	os.Setenv("KNOT_MCP_ADDRESS", "env.example.com")
	os.Setenv("KNOT_MCP_PORT", "9999")
	defer func() {
		os.Unsetenv("KNOT_MCP_ADDRESS")
		os.Unsetenv("KNOT_MCP_PORT")
	}()

	// Create flags with defaults and env var support
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "address",
			EnvVars: []string{"KNOT_MCP_ADDRESS"},
			Value:   "localhost",
		},
		&cli.IntFlag{
			Name:    "port",
			EnvVars: []string{"KNOT_MCP_PORT"},
			Value:   8080,
		},
	}

	// Create app
	app := &cli.App{
		Name:  "knot",
		Flags: flags,
		Action: func(c *cli.Context) error {
			// Initialize configuration
			InitializeFromCLIContext(c)

			svc := GetConfigService()
			config := svc.GetMCPConfig(c)

			// CLI flags should have higher priority than environment variables
			if config.Address != "cli.example.com" {
				t.Errorf("Expected address from CLI flag 'cli.example.com', got %s", config.Address)
			}
			if config.Port != 7777 {
				t.Errorf("Expected port from CLI flag 7777, got %d", config.Port)
			}
			return nil
		},
	}

	// Run the app
	err := app.Run(os.Args)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDefaultValues(t *testing.T) {
	// Reset singleton for testing
	instance = nil
	once = sync.Once{}

	// Create flags with defaults
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "address",
			EnvVars: []string{"KNOT_MCP_ADDRESS"},
			Value:   "localhost",
		},
		&cli.IntFlag{
			Name:    "port",
			EnvVars: []string{"KNOT_MCP_PORT"},
			Value:   8080,
		},
		&cli.StringFlag{
			Name:    "log-level",
			EnvVars: []string{"KNOT_LOG_LEVEL"},
			Value:   "info",
		},
	}

	// Create app without MCP mode (no mcp command)
	app := &cli.App{
		Name:  "knot",
		Flags: flags,
		Action: func(c *cli.Context) error {
			// Initialize configuration
			InitializeFromCLIContext(c)

			svc := GetConfigService()

			// Should not be in MCP mode
			if svc.IsMCPMode() {
				t.Error("Should not be in MCP mode when args don't contain 'mcp'")
			}

			// Test default values from CLI context
			config := svc.GetMCPConfig(c)
			if config.Address != "localhost" {
				t.Errorf("Expected default address 'localhost', got %s", config.Address)
			}
			if config.Port != 8080 {
				t.Errorf("Expected default port 8080, got %d", config.Port)
			}
			if config.LogLevel != "info" {
				t.Errorf("Expected default log level 'info', got %s", config.LogLevel)
			}
			return nil
		},
	}

	// Run the app with no mcp command
	err := app.Run([]string{"knot"})
	if err != nil {
		t.Fatal(err)
	}
}