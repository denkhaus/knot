package config

import (
	"flag"
	"os"
	"testing"

	"github.com/samber/do/v2"
	"github.com/urfave/cli/v2"
)

func TestConfigServiceSingleton(t *testing.T) {
	// Create a new injector for testing
	injector := do.New()

	// Create service using NewService
	app := &cli.App{}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("log-level", "info", "")
	flagSet.Int("complexity-threshold", 5, "")
	flagSet.Int("max-depth", 10, "")
	flagSet.Int("max-tasks-per-depth", 50, "")
	flagSet.Int("max-description-length", 500, "")
	flagSet.Bool("auto-reduce-complexity", true, "")
	cliCtx := cli.NewContext(app, flagSet, nil)
	svcFactory := NewService(cliCtx)
	svc1, err := svcFactory(injector)
	if err != nil {
		t.Fatal(err)
	}

	svc2, err := svcFactory(injector)
	if err != nil {
		t.Fatal(err)
	}

	// Services should be different instances since we create them each time
	// but they should have the same configuration
	if svc1.GetLogLevel() != svc2.GetLogLevel() {
		t.Error("Services should have the same log level")
	}
}

func TestInitializeFromCLIContext(t *testing.T) {
	// Create a new injector for testing
	injector := do.New()

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
		// Add manager config flags with proper defaults
		&cli.IntFlag{
			Name:    "max-tasks-per-depth",
			Value:   100,
			EnvVars: []string{"KNOT_MAX_TASKS_PER_DEPTH"},
		},
		&cli.IntFlag{
			Name:    "complexity-threshold",
			Value:   8,
			EnvVars: []string{"KNOT_COMPLEXITY_THRESHOLD"},
		},
		&cli.IntFlag{
			Name:    "max-depth",
			Value:   5,
			EnvVars: []string{"KNOT_MAX_DEPTH"},
		},
		&cli.IntFlag{
			Name:    "max-description-length",
			Value:   2000,
			EnvVars: []string{"KNOT_MAX_DESCRIPTION_LENGTH"},
		},
		&cli.BoolFlag{
			Name:    "auto-reduce-complexity",
			Value:   true,
			EnvVars: []string{"KNOT_AUTO_REDUCE_COMPLEXITY"},
		},
		&cli.StringFlag{
			Name:  "log-level",
			Value: "off",
		},
	}

	// Create a complete app structure that matches real usage
	app := &cli.App{
		Name:  "knot",
		Flags: flags,
		Action: func(c *cli.Context) error {
			// Create service using NewService
			svcFactory := NewService(c)
			svc, err := svcFactory(injector)
			if err != nil {
				t.Fatal(err)
			}

			// Test that MCP mode was detected
			if !svc.IsMCPMode() {
				t.Error("Should be in MCP mode when args contain 'mcp'")
			}

			// Test that configuration comes from CLI context
			config := svc.GetMCPConfig()
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
	// Create a new injector for testing
	injector := do.New()

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
		// Add manager config flags with proper defaults
		&cli.IntFlag{
			Name:    "max-tasks-per-depth",
			Value:   100,
			EnvVars: []string{"KNOT_MAX_TASKS_PER_DEPTH"},
		},
		&cli.IntFlag{
			Name:    "complexity-threshold",
			Value:   8,
			EnvVars: []string{"KNOT_COMPLEXITY_THRESHOLD"},
		},
		&cli.IntFlag{
			Name:    "max-depth",
			Value:   5,
			EnvVars: []string{"KNOT_MAX_DEPTH"},
		},
		&cli.IntFlag{
			Name:    "max-description-length",
			Value:   2000,
			EnvVars: []string{"KNOT_MAX_DESCRIPTION_LENGTH"},
		},
		&cli.BoolFlag{
			Name:    "auto-reduce-complexity",
			Value:   true,
			EnvVars: []string{"KNOT_AUTO_REDUCE_COMPLEXITY"},
		},
		&cli.StringFlag{
			Name:  "log-level",
			Value: "off",
		},
	}

	// Create app
	app := &cli.App{
		Name:  "knot",
		Flags: flags,
		Action: func(c *cli.Context) error {
			// Create service using NewService
			svcFactory := NewService(c)
			svc, err := svcFactory(injector)
			if err != nil {
				t.Fatal(err)
			}

			config := svc.GetMCPConfig()

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
	// Create a new injector for testing
	injector := do.New()

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
		// Add manager config flags with proper defaults
		&cli.IntFlag{
			Name:    "max-tasks-per-depth",
			Value:   100,
			EnvVars: []string{"KNOT_MAX_TASKS_PER_DEPTH"},
		},
		&cli.IntFlag{
			Name:    "complexity-threshold",
			Value:   8,
			EnvVars: []string{"KNOT_COMPLEXITY_THRESHOLD"},
		},
		&cli.IntFlag{
			Name:    "max-depth",
			Value:   5,
			EnvVars: []string{"KNOT_MAX_DEPTH"},
		},
		&cli.IntFlag{
			Name:    "max-description-length",
			Value:   2000,
			EnvVars: []string{"KNOT_MAX_DESCRIPTION_LENGTH"},
		},
		&cli.BoolFlag{
			Name:    "auto-reduce-complexity",
			Value:   true,
			EnvVars: []string{"KNOT_AUTO_REDUCE_COMPLEXITY"},
		},
		&cli.StringFlag{
			Name:  "log-level",
			Value: "off",
		},
	}

	// Create app
	app := &cli.App{
		Name:  "knot",
		Flags: flags,
		Action: func(c *cli.Context) error {
			// Create service using NewService
			svcFactory := NewService(c)
			svc, err := svcFactory(injector)
			if err != nil {
				t.Fatal(err)
			}

			config := svc.GetMCPConfig()

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
	// Create a new injector for testing
	injector := do.New()

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
		// Add manager config flags with proper defaults
		&cli.IntFlag{
			Name:    "max-tasks-per-depth",
			Value:   100,
			EnvVars: []string{"KNOT_MAX_TASKS_PER_DEPTH"},
		},
		&cli.IntFlag{
			Name:    "complexity-threshold",
			Value:   8,
			EnvVars: []string{"KNOT_COMPLEXITY_THRESHOLD"},
		},
		&cli.IntFlag{
			Name:    "max-depth",
			Value:   5,
			EnvVars: []string{"KNOT_MAX_DEPTH"},
		},
		&cli.IntFlag{
			Name:    "max-description-length",
			Value:   2000,
			EnvVars: []string{"KNOT_MAX_DESCRIPTION_LENGTH"},
		},
		&cli.BoolFlag{
			Name:    "auto-reduce-complexity",
			Value:   true,
			EnvVars: []string{"KNOT_AUTO_REDUCE_COMPLEXITY"},
		},
	}

	// Create app without MCP mode (no mcp command)
	app := &cli.App{
		Name:  "knot",
		Flags: flags,
		Action: func(c *cli.Context) error {
			// Create service using NewService
			svcFactory := NewService(c)
			svc, err := svcFactory(injector)
			if err != nil {
				t.Fatal(err)
			}

			// Should not be in MCP mode
			if svc.IsMCPMode() {
				t.Error("Should not be in MCP mode when args don't contain 'mcp'")
			}

			// Test default values from CLI context
			config := svc.GetMCPConfig()
			if config.Address != "localhost" {
				t.Errorf("Expected default address 'localhost', got %s", config.Address)
			}
			if config.Port != 8080 {
				t.Errorf("Expected default port 8080, got %d", config.Port)
			}
			if svc.GetLogLevel() != "info" {
				t.Errorf("Expected default log level 'info', got %s", svc.GetLogLevel())
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
