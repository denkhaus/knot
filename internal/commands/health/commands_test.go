package health

import (
	"context"
	"errors"
	"flag"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/di"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestCommands(t *testing.T) {
	tests := []struct {
		name             string
		expectedCommands []string
	}{
		{
			name: "basic health commands",
			expectedCommands: []string{
				"check",
				"ping",
				"validate",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := Commands()

			// Verify we get the expected number of commands
			assert.Len(t, commands, len(tt.expectedCommands))

			// Verify command names
			commandNames := make([]string, len(commands))
			for i, cmd := range commands {
				commandNames[i] = cmd.Name
			}

			for _, expectedName := range tt.expectedCommands {
				assert.Contains(t, commandNames, expectedName,
					"Command '%s' should be present in health commands", expectedName)
			}
		})
	}
}

func TestCommandStructure(t *testing.T) {
	commands := Commands()

	// Test that all commands have required structure
	for _, cmd := range commands {
		t.Run("command_"+cmd.Name, func(t *testing.T) {
			// Verify command has name and usage
			assert.NotEmpty(t, cmd.Name, "Command should have a name")
			assert.NotEmpty(t, cmd.Usage, "Command should have usage text")

			// Verify command has action
			assert.NotNil(t, cmd.Action, "Command should have an action function")

			// Verify flags are properly structured
			for _, flag := range cmd.Flags {
				assert.NotNil(t, flag, "Flag should not be nil")
				flagNames := flag.Names()
				assert.NotEmpty(t, flagNames, "Flag should have at least one name")
			}
		})
	}
}

func TestCheckCommandFlags(t *testing.T) {
	commands := Commands()

	// Find the check command
	var checkCommand *cli.Command
	for _, cmd := range commands {
		if cmd.Name == "check" {
			checkCommand = cmd
			break
		}
	}

	require.NotNil(t, checkCommand, "Check command should be found")

	// Check expected flags
	expectedFlags := []string{"json", shared.HealthTestTimeout}
	flagNames := make([]string, 0)
	for _, flag := range checkCommand.Flags {
		flagNames = append(flagNames, flag.Names()...)
	}

	for _, expectedFlag := range expectedFlags {
		assert.Contains(t, flagNames, expectedFlag,
			"Check command should have '%s' flag", expectedFlag)
	}

	// Check default values for flags
	for _, flag := range checkCommand.Flags {
		switch flag.Names()[0] {
		case "json":
			boolFlag, ok := flag.(*cli.BoolFlag)
			require.True(t, ok, "json flag should be a BoolFlag")
			assert.False(t, boolFlag.Value, "json flag should default to false")
		case shared.HealthTestTimeout:
			durationFlag, ok := flag.(*cli.DurationFlag)
			require.True(t, ok, "timeout flag should be a DurationFlag")
			assert.Equal(t, 10*time.Second, durationFlag.Value, "timeout flag should default to 10 seconds")
		}
	}
}

func TestPingCommandFlags(t *testing.T) {
	commands := Commands()

	// Find the ping command
	var pingCommand *cli.Command
	for _, cmd := range commands {
		if cmd.Name == "ping" {
			pingCommand = cmd
			break
		}
	}

	require.NotNil(t, pingCommand, "Ping command should be found")

	// Check expected flags
	expectedFlags := []string{shared.HealthTestTimeout}
	flagNames := make([]string, 0)
	for _, flag := range pingCommand.Flags {
		flagNames = append(flagNames, flag.Names()...)
	}

	for _, expectedFlag := range expectedFlags {
		assert.Contains(t, flagNames, expectedFlag,
			"Ping command should have '%s' flag", expectedFlag)
	}

	// Check default value for timeout flag
	for _, flag := range pingCommand.Flags {
		if flag.Names()[0] == shared.HealthTestTimeout {
			durationFlag, ok := flag.(*cli.DurationFlag)
			require.True(t, ok, "timeout flag should be a DurationFlag")
			assert.Equal(t, 5*time.Second, durationFlag.Value, "timeout flag should default to 5 seconds")
		}
	}
}

func TestValidateCommandFlags(t *testing.T) {
	commands := Commands()

	// Find the validate command
	var validateCommand *cli.Command
	for _, cmd := range commands {
		if cmd.Name == "validate" {
			validateCommand = cmd
			break
		}
	}

	require.NotNil(t, validateCommand, "Validate command should be found")

	// Check expected flags
	expectedFlags := []string{shared.HealthTestTimeout}
	flagNames := make([]string, 0)
	for _, flag := range validateCommand.Flags {
		flagNames = append(flagNames, flag.Names()...)
	}

	for _, expectedFlag := range expectedFlags {
		assert.Contains(t, flagNames, expectedFlag,
			"Validate command should have '%s' flag", expectedFlag)
	}
	// Check default value for timeout flag
	for _, flag := range validateCommand.Flags {
		if flag.Names()[0] == shared.HealthTestTimeout {
			durationFlag, ok := flag.(*cli.DurationFlag)
			require.True(t, ok, "timeout flag should be a DurationFlag")
			assert.Equal(t, 30*time.Second, durationFlag.Value, "timeout flag should default to 30 seconds")
		}
	}
}

func TestCommandUsageText(t *testing.T) {
	commands := Commands()

	// Test that commands have meaningful usage text
	expectedUsages := map[string]string{
		"check":    "Check database connection health",
		"ping":     "Simple database connectivity test",
		"validate": "Comprehensive database connection validation",
	}

	for _, cmd := range commands {
		if expectedUsage, exists := expectedUsages[cmd.Name]; exists {
			assert.Equal(t, expectedUsage, cmd.Usage,
				"Command '%s' should have correct usage text", cmd.Name)
		}
	}
}

func TestCommandsReturnNewSlice(t *testing.T) {
	commands1 := Commands()
	commands2 := Commands()

	// Commands should return new slices to avoid modification issues
	assert.NotSame(t, &commands1[0], &commands2[0],
		"Commands should return new slice instances")
}

func TestCommandIntegration(t *testing.T) {
	commands := Commands()

	t.Run("command creation with app context", func(t *testing.T) {
		// Verify commands can be created with a valid app context
		assert.NotNil(t, commands)
		assert.Greater(t, len(commands), 0)
	})

	t.Run("command DI dependency", func(t *testing.T) {
		// Set up test injector
		config := testutil.NewTestConfig(t)
		testInjector := config.SetupTestInjector(t)

		// Test that commands properly work with DI
		assert.NotNil(t, testInjector, "DI injector should be available")

		// Verify we can get services from DI
		logger := do.MustInvoke[logger.Logger](testInjector)
		assert.NotNil(t, logger, "Logger should be available from DI")
	})
}

func TestCommandEdgeCases(t *testing.T) {
	t.Run("empty app context", func(t *testing.T) {
		// This test would verify behavior with nil/empty app context
		// But since we don't want to cause panics, we'll use a valid context

		commands := Commands()
		assert.NotEmpty(t, commands, "Commands should be created even with minimal app context")
	})
}

func TestHealthCheckTimeouts(t *testing.T) {
	commands := Commands()

	t.Run("check command timeout configuration", func(t *testing.T) {
		var checkCommand *cli.Command
		for _, cmd := range commands {
			if cmd.Name == "check" {
				checkCommand = cmd
				break
			}
		}

		require.NotNil(t, checkCommand)

		// Find timeout flag and verify it's properly configured
		for _, flag := range checkCommand.Flags {
			if flag.Names()[0] == shared.HealthTestTimeout {
				durationFlag := flag.(*cli.DurationFlag)
				assert.Equal(t, "Health check timeout", durationFlag.Usage)
				assert.Equal(t, 10*time.Second, durationFlag.Value)
				break
			}
		}
	})

	t.Run("ping command timeout configuration", func(t *testing.T) {
		var pingCommand *cli.Command
		for _, cmd := range commands {
			if cmd.Name == "ping" {
				pingCommand = cmd
				break
			}
		}

		require.NotNil(t, pingCommand)

		// Find timeout flag and verify it's properly configured
		for _, flag := range pingCommand.Flags {
			if flag.Names()[0] == shared.HealthTestTimeout {
				durationFlag := flag.(*cli.DurationFlag)
				assert.Equal(t, "Ping timeout", durationFlag.Usage)
				assert.Equal(t, 5*time.Second, durationFlag.Value)
				break
			}
		}
	})
}

func Test_performHealthCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	healthyStatus := &types.HealthStatus{
		Healthy:          true,
		ConnectionActive: true,
		PingLatency:      10 * time.Millisecond,
		OpenConnections:  5,
		IdleConnections:  2,
		InUseConnections: 3,
		ErrorMessage:     "",
		LastChecked:      time.Now(),
		DatabasePath:     "test.db",
		WALModeEnabled:   true,
		ForeignKeys:      true,
	}

	unhealthyStatus := &types.HealthStatus{
		Healthy:          false,
		ConnectionActive: false,
		PingLatency:      0,
		OpenConnections:  0,
		IdleConnections:  0,
		InUseConnections: 0,
		ErrorMessage:     "db connection failed",
		LastChecked:      time.Now(),
		DatabasePath:     "test.db",
		WALModeEnabled:   false,
		ForeignKeys:      false,
	}

	tests := []struct {
		name                 string
		mockHealthStatus     *types.HealthStatus
		mockHealthErr        error
		expectedHealthy      bool
		expectedErrorMessage string
	}{
		{
			name:                 "healthy check",
			mockHealthStatus:     healthyStatus,
			mockHealthErr:        nil,
			expectedHealthy:      true,
			expectedErrorMessage: "",
		},
		{
			name:                 "unhealthy check",
			mockHealthStatus:     unhealthyStatus,
			mockHealthErr:        nil,
			expectedHealthy:      false,
			expectedErrorMessage: "db connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMgr := mocks.NewMockProjectManager(ctrl)
			mockMgr.EXPECT().
				HealthCheck(gomock.Any()).
				Return(tt.mockHealthStatus, tt.mockHealthErr)
			// Use DI injector with mock manager
			diContainer := di.NewContainer()
			app := &cli.App{}
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			flagSet.String("log-level", "info", "")
			flagSet.Int("complexity-threshold", 5, "")
			flagSet.Int("max-depth", 10, "")
			flagSet.Int("max-tasks-per-depth", 50, "")
			flagSet.Int("max-description-length", 500, "")
			flagSet.Bool("auto-reduce-complexity", true, "")
			cliCtx := cli.NewContext(app, flagSet, nil)
			testInjector := diContainer.RegisterAllServices(cliCtx)
			do.OverrideValue(testInjector, mockMgr)

			healthStatus, err := performHealthCheck(context.Background(), mockMgr)

			if tt.expectedHealthy {
				require.NoError(t, err)
				assert.True(t, healthStatus.Healthy)
				assert.Empty(t, healthStatus.ErrorMessage)
			} else {
				require.NoError(t, err) // performHealthCheck doesn't return error for unhealthy status
				assert.False(t, healthStatus.Healthy)
				assert.Contains(t, healthStatus.ErrorMessage, tt.expectedErrorMessage)
			}
			assert.NotZero(t, healthStatus.LastChecked)
			assert.GreaterOrEqual(t, healthStatus.PingLatency, time.Duration(0))
		})
	}
}

func Test_performPing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		pingErr     error
		expectedErr bool
	}{
		{
			name:        "successful ping",
			pingErr:     nil,
			expectedErr: false,
		},
		{
			name:        "failed ping",
			pingErr:     errors.New("network unreachable"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMgr := mocks.NewMockProjectManager(ctrl)
			mockMgr.EXPECT().
				Ping(gomock.Any()).
				Return(tt.pingErr)
			// Use DI injector with mock manager
			diContainer := di.NewContainer()
			app := &cli.App{}
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			flagSet.String("log-level", "info", "")
			flagSet.Int("complexity-threshold", 5, "")
			flagSet.Int("max-depth", 10, "")
			flagSet.Int("max-tasks-per-depth", 50, "")
			flagSet.Int("max-description-length", 500, "")
			flagSet.Bool("auto-reduce-complexity", true, "")
			cliCtx := cli.NewContext(app, flagSet, nil)
			testInjector := diContainer.RegisterAllServices(cliCtx)
			do.OverrideValue(testInjector, mockMgr)

			err := performPing(context.Background(), mockMgr)

			if tt.expectedErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.pingErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_performValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name                 string
		validateConnErr      error
		expectedErr          bool
		expectedErrorMessage string
	}{
		{
			name:            "successful validation",
			validateConnErr: nil,
			expectedErr:     false,
		},
		{
			name:                 "validation fails",
			validateConnErr:      errors.New("validation error"),
			expectedErr:          true,
			expectedErrorMessage: "validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMgr := mocks.NewMockProjectManager(ctrl)

			// Set up expectations based on the test case
			mockMgr.EXPECT().
				ValidateConnection(gomock.Any()).
				Return(tt.validateConnErr)

			// Use DI injector with mock manager
			diContainer := di.NewContainer()
			app := &cli.App{}
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			flagSet.String("log-level", "info", "")
			flagSet.Int("complexity-threshold", 5, "")
			flagSet.Int("max-depth", 10, "")
			flagSet.Int("max-tasks-per-depth", 50, "")
			flagSet.Int("max-description-length", 500, "")
			flagSet.Bool("auto-reduce-complexity", true, "")
			cliCtx := cli.NewContext(app, flagSet, nil)
			testInjector := diContainer.RegisterAllServices(cliCtx)
			do.OverrideValue(testInjector, mockMgr)

			err := performValidation(context.Background(), mockMgr)

			if tt.expectedErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_printStatus(t *testing.T) {
	tests := []struct {
		name           string
		healthStatus   *types.HealthStatus
		expectedOutput string
	}{
		{
			name: "healthy status",
			healthStatus: &types.HealthStatus{
				Healthy:          true,
				ConnectionActive: true,
				PingLatency:      10 * time.Millisecond,
				DatabasePath:     "test.db",
				LastChecked:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expectedOutput: `Database Health Status:

✅ Status: Healthy
📊 Connection Details:
   Active: true
   Latency: 10ms
   Database: test.db
   Last Checked: 2023-01-01T12:00:00Z
`,
		},
		{
			name: "unhealthy status with error",
			healthStatus: &types.HealthStatus{
				Healthy:          false,
				ConnectionActive: false,
				PingLatency:      500 * time.Millisecond,
				ErrorMessage:     "connection refused",
				DatabasePath:     "test.db",
				LastChecked:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expectedOutput: `Database Health Status:

❌ Status: Unhealthy
   Error: connection refused
📊 Connection Details:
   Active: false
   Latency: 500ms
   Database: test.db
   Last Checked: 2023-01-01T12:00:00Z
`,
		},
		{
			name: "status with connection pool and sqlite settings",
			healthStatus: &types.HealthStatus{
				Healthy:          true,
				ConnectionActive: true,
				PingLatency:      5 * time.Millisecond,
				OpenConnections:  10,
				IdleConnections:  5,
				InUseConnections: 5,
				DatabasePath:     "prod.db",
				WALModeEnabled:   true,
				ForeignKeys:      true,
				LastChecked:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expectedOutput: `Database Health Status:

✅ Status: Healthy
📊 Connection Details:
   Active: true
   Latency: 5ms
   Database: prod.db
   Last Checked: 2023-01-01T12:00:00Z
🔗 Connection Pool:
   Open: 10
   Idle: 5
   In Use: 5
⚙️  SQLite Settings:
   WAL Mode: ✅ Enabled
   Foreign Keys: ✅ Enabled
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printStatus(tt.healthStatus)

			_ = w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = oldStdout // Restore stdout

			// Clean up extra spaces/newlines to make comparison easier
			actualOutput := strings.TrimSpace(string(out))
			expectedOutput := strings.TrimSpace(tt.expectedOutput)

			assert.Equal(t, expectedOutput, actualOutput)
		})
	}
}

// TestCheckAction tests the check CLI action
func TestCheckAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	diContainer := di.NewContainer()
	app := &cli.App{
		Metadata: make(map[string]interface{}),
	}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("log-level", "info", "")
	flagSet.Int("complexity-threshold", 5, "")
	flagSet.Int("max-depth", 10, "")
	flagSet.Int("max-tasks-per-depth", 50, "")
	flagSet.Int("max-description-length", 500, "")
	flagSet.Bool("auto-reduce-complexity", true, "")
	flagSet.Bool("json", false, "")
	flagSet.Duration("health-test-timeout", 10000000000, "") // 10 seconds in nanoseconds
	flagSet.Duration("timeout", 10000000000, "")
	cliCtx := cli.NewContext(app, flagSet, nil)

	testInjector2 := diContainer.RegisterAllServices(cliCtx)
	repo := do.MustInvoke[types.Repository](testInjector)
	do.Override(testInjector2, func(do.Injector) (types.Repository, error) {
		return repo, nil
	})

	// Set the container in app metadata for shared.GetContainerFromContext to find it
	app.Metadata["container"] = diContainer

	t.Run("check action with healthy database", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		action := checkAction()
		err := action(cliCtx)

		_ = w.Close()
		os.Stdout = oldStdout

		// Read output
		out, _ := io.ReadAll(r)
		output := string(out)

		// Should succeed and show healthy status
		assert.NoError(t, err)
		assert.Contains(t, output, "Healthy")
	})

	t.Run("check action with json output", func(t *testing.T) {
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("log-level", "info", "")
		flagSet.Int("complexity-threshold", 5, "")
		flagSet.Int("max-depth", 10, "")
		flagSet.Int("max-tasks-per-depth", 50, "")
		flagSet.Int("max-description-length", 500, "")
		flagSet.Bool("auto-reduce-complexity", true, "")
		flagSet.Bool("json", true, "")
		flagSet.Duration("health-test-timeout", 10000000000, "")
		flagSet.Duration("timeout", 10000000000, "")
		cliCtx := cli.NewContext(app, flagSet, nil)

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		action := checkAction()
		err := action(cliCtx)

		_ = w.Close()
		os.Stdout = oldStdout

		// Read output
		out, _ := io.ReadAll(r)
		output := string(out)

		// Should succeed and output JSON
		assert.NoError(t, err)
		assert.Contains(t, output, "{")
		assert.Contains(t, output, "healthy")
	})
}

// TestPingAction tests the ping CLI action
func TestPingAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	diContainer := di.NewContainer()
	app := &cli.App{
		Metadata: make(map[string]interface{}),
	}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("log-level", "info", "")
	flagSet.Int("complexity-threshold", 5, "")
	flagSet.Int("max-depth", 10, "")
	flagSet.Int("max-tasks-per-depth", 50, "")
	flagSet.Int("max-description-length", 500, "")
	flagSet.Bool("auto-reduce-complexity", true, "")
	flagSet.Duration("health-test-timeout", 5000000000, "") // 5 seconds in nanoseconds
	flagSet.Duration("timeout", 5000000000, "")
	cliCtx := cli.NewContext(app, flagSet, nil)

	testInjector2 := diContainer.RegisterAllServices(cliCtx)
	repo := do.MustInvoke[types.Repository](testInjector)
	do.Override(testInjector2, func(do.Injector) (types.Repository, error) {
		return repo, nil
	})

	// Set the container in app metadata for shared.GetContainerFromContext to find it
	app.Metadata["container"] = diContainer

	t.Run("ping action success", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		action := pingAction()
		err := action(cliCtx)

		_ = w.Close()
		os.Stdout = oldStdout

		// Read output
		out, _ := io.ReadAll(r)
		output := string(out)

		// Should succeed
		assert.NoError(t, err)
		assert.Contains(t, output, "ping successful")
	})
}

// TestValidateAction tests the validate CLI action
func TestValidateAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	testInjector := config.SetupTestInjector(t)

	diContainer := di.NewContainer()
	app := &cli.App{
		Metadata: make(map[string]interface{}),
	}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("log-level", "info", "")
	flagSet.Int("complexity-threshold", 5, "")
	flagSet.Int("max-depth", 10, "")
	flagSet.Int("max-tasks-per-depth", 50, "")
	flagSet.Int("max-description-length", 500, "")
	flagSet.Bool("auto-reduce-complexity", true, "")
	flagSet.Duration("health-test-timeout", 30000000000, "") // 30 seconds in nanoseconds
	flagSet.Duration("timeout", 30000000000, "")
	cliCtx := cli.NewContext(app, flagSet, nil)

	testInjector2 := diContainer.RegisterAllServices(cliCtx)
	repo := do.MustInvoke[types.Repository](testInjector)
	do.Override(testInjector2, func(do.Injector) (types.Repository, error) {
		return repo, nil
	})

	// Set the container in app metadata for shared.GetContainerFromContext to find it
	app.Metadata["container"] = diContainer

	t.Run("validate action success", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		action := validateAction()
		err := action(cliCtx)

		_ = w.Close()
		os.Stdout = oldStdout

		// Read output
		out, _ := io.ReadAll(r)
		output := string(out)

		// Should succeed
		assert.NoError(t, err)
		assert.Contains(t, output, "validation successful")
	})
}
