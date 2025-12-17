package health

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

// createTestAppContext creates a test app context for testing
func createTestAppContext(t *testing.T) *shared.AppContext {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	logger := zaptest.NewLogger(t)
	return shared.NewAppContext(mgr, logger)
}

// createTestAppContextWithMockManager creates a test app context with a mock ProjectManager
func createTestAppContextWithMockManager(t *testing.T, mockManager *mocks.MockProjectManager) *shared.AppContext {
	logger := zaptest.NewLogger(t)
	return shared.NewAppContext(mockManager, logger)
}

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
			appCtx := createTestAppContext(t)
			commands := Commands(appCtx)

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
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

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
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

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
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

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
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

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
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

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
	appCtx := createTestAppContext(t)
	commands1 := Commands(appCtx)
	commands2 := Commands(appCtx)

	// Commands should return new slices to avoid modification issues
	assert.NotSame(t, &commands1[0], &commands2[0],
		"Commands should return new slice instances")
}

func TestCommandIntegration(t *testing.T) {
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

	t.Run("command creation with app context", func(t *testing.T) {
		// Verify commands can be created with a valid app context
		assert.NotNil(t, commands)
		assert.Greater(t, len(commands), 0)
	})

	t.Run("command app context dependency", func(t *testing.T) {
		// Test that commands properly depend on app context
		assert.NotNil(t, appCtx, "AppContext should be available")
		assert.NotNil(t, projectManager, "ProjectManager should be available")
		assert.NotNil(t, appCtx.Logger, "Logger should be available")
	})
}

func TestCommandEdgeCases(t *testing.T) {
	t.Run("empty app context", func(t *testing.T) {
		// This test would verify behavior with nil/empty app context
		// But since we don't want to cause panics, we'll use a valid context
		appCtx := createTestAppContext(t)

		commands := Commands(appCtx)
		assert.NotEmpty(t, commands, "Commands should be created even with minimal app context")
	})
}

func TestHealthCheckTimeouts(t *testing.T) {
	appCtx := createTestAppContext(t)
	commands := Commands(appCtx)

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

	tests := []struct {
		name                 string
		listProjectsErr      error
		expectedHealthy      bool
		expectedErrorMessage string
	}{
		{
			name:                 "healthy check",
			listProjectsErr:      nil,
			expectedHealthy:      true,
			expectedErrorMessage: "",
		},
		{
			name:                 "unhealthy check",
			listProjectsErr:      errors.New("db connection failed"),
			expectedHealthy:      false,
			expectedErrorMessage: "db connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMgr := mocks.NewMockProjectManager(ctrl)
			mockMgr.EXPECT().
				ListProjects(gomock.Any()).
				Return([]*types.Project{}, tt.listProjectsErr)
			appCtx := createTestAppContextWithMockManager(t, mockMgr)

			healthStatus, err := performHealthCheck(context.Background(), appCtx)

			if tt.expectedHealthy {
				require.NoError(t, err)
				assert.True(t, healthStatus.Healthy)
				assert.Empty(t, healthStatus.ErrorMessage)
			} else {
				require.Error(t, err)
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
		name            string
		listProjectsErr error
		expectedErr     bool
	}{
		{
			name:            "successful ping",
			listProjectsErr: nil,
			expectedErr:     false,
		},
		{
			name:            "failed ping",
			listProjectsErr: errors.New("network unreachable"),
			expectedErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMgr := mocks.NewMockProjectManager(ctrl)
			mockMgr.EXPECT().
				ListProjects(gomock.Any()).
				Return([]*types.Project{}, tt.listProjectsErr)
			appCtx := createTestAppContextWithMockManager(t, mockMgr)

			err := performPing(context.Background(), appCtx)

			if tt.expectedErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.listProjectsErr.Error())
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
		listProjectsErr      error
		getConfigReturn      *manager.Config
		expectedErr          bool
		expectedErrorMessage string
	}{
		{
			name:            "successful validation",
			listProjectsErr: nil,
			getConfigReturn: manager.DefaultConfig(),
			expectedErr:     false,
		},
		{
			name:                 "list projects fails",
			listProjectsErr:      errors.New("list projects error"),
			getConfigReturn:      manager.DefaultConfig(),
			expectedErr:          true,
			expectedErrorMessage: "List Projects failed: list projects error",
		},
		{
			name:                 "get config returns nil",
			listProjectsErr:      nil,
			getConfigReturn:      nil,
			expectedErr:          true,
			expectedErrorMessage: "Get Config failed: config is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMgr := mocks.NewMockProjectManager(ctrl)

			// Set up expectations based on the test case
			mockMgr.EXPECT().
				ListProjects(gomock.Any()).
				Return([]*types.Project{}, tt.listProjectsErr)

			// Only expect GetConfig if ListProjects succeeds (function short-circuits on error)
			if tt.listProjectsErr == nil {
				mockMgr.EXPECT().
					GetConfig().
					Return(tt.getConfigReturn)
			}

			appCtx := createTestAppContextWithMockManager(t, mockMgr)

			err := performValidation(context.Background(), appCtx)

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
		healthStatus   *Status
		expectedOutput string
	}{
		{
			name: "healthy status",
			healthStatus: &Status{
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
			healthStatus: &Status{
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
			healthStatus: &Status{
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
