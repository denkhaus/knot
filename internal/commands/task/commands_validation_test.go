package task

import (
	"flag"
	"strconv"
	"testing"

	"github.com/denkhaus/knot/internal/shared"
	"github.com/denkhaus/knot/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCreateActionValidation(t *testing.T) {
	// Setup test environment
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	tests := []struct {
		name        string
		title       string
		description string
		complexity  int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid task creation",
			title:       "Valid Task",
			description: "Valid description",
			complexity:  5,
			expectError: false,
		},
		{
			name:        "empty title should fail",
			title:       "",
			description: "Valid description",
			complexity:  5,
			expectError: true,
			errorMsg:    "title cannot be empty",
		},
		{
			name:        "title too long should fail",
			title:       string(make([]byte, 201)), // 201 characters
			description: "Valid description",
			complexity:  5,
			expectError: true,
			errorMsg:    "title too long",
		},
		{
			name:        "HTML in title should fail",
			title:       "Task with <script>alert('xss')</script>",
			description: "Valid description",
			complexity:  5,
			expectError: true,
			errorMsg:    "contains HTML tags",
		},
		{
			name:        "invalid complexity should fail",
			title:       "Valid Task",
			description: "Valid description",
			complexity:  15, // Invalid complexity
			expectError: true,
			errorMsg:    "complexity must be between 1 and 10",
		},
		{
			name:        "description too long should fail",
			title:       "Valid Task",
			description: string(make([]byte, 2001)), // 2001 characters
			complexity:  5,
			expectError: true,
			errorMsg:    "description too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create CLI context (no longer needs project-id flag)
			app := &cli.App{}
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			flagSet.String("title", "", "")
			flagSet.String("description", "", "")
			flagSet.String("complexity", "", "")
			flagSet.String("priority", "", "")
			flagSet.String("actor", "", "")

			_ = flagSet.Set("title", tt.title)
			_ = flagSet.Set("description", tt.description)
			_ = flagSet.Set("complexity", strconv.Itoa(tt.complexity))
			_ = flagSet.Set("priority", "medium")
			_ = flagSet.Set("actor", "test-user")

			ctx := cli.NewContext(app, flagSet, nil)

			// Create the action and set project context
			appCtx := &shared.AppContext{
				ProjectManager: mgr,
				Logger:         config.Logger,
			}

			// Set project context for the test
			err := mgr.SetSelectedProject(ctx.Context, project.ID, "test-user")
			require.NoError(t, err)

			action := createAction(appCtx)

			// Execute the action
			err = action(ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputValidationIntegration(t *testing.T) {
	// This test ensures that our input validation is properly integrated
	// into the CLI command handlers

	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Test that validation errors are properly wrapped as EnhancedErrors
	app := &cli.App{}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("title", "", "")
	flagSet.String("description", "", "")
	flagSet.String("complexity", "", "")
	flagSet.String("priority", "", "")
	flagSet.String("actor", "", "")

	_ = flagSet.Set("title", "<script>alert('xss')</script>") // Should trigger validation error
	_ = flagSet.Set("description", "Valid description")
	_ = flagSet.Set("complexity", "5")
	_ = flagSet.Set("priority", "medium")
	_ = flagSet.Set("actor", "test-user")

	ctx := cli.NewContext(app, flagSet, nil)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	// Set project context for the test
	err := mgr.SetSelectedProject(ctx.Context, project.ID, "test-user")
	require.NoError(t, err)

	action := createAction(appCtx)

	err = action(ctx)
	require.Error(t, err)

	// Check that it's wrapped as a validation error
	assert.Contains(t, err.Error(), "title contains HTML tags")
	assert.Contains(t, err.Error(), "not allowed")
}
