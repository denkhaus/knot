package task

import (
	"flag"
	"strconv"
	"testing"

	"github.com/denkhaus/knot/v2/internal/di"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

// setupCLIContextForValidation creates a CLI context with proper DI container setup for validation tests
func setupCLIContextForValidation(t *testing.T) (*cli.Context, *cli.App, *di.Container) {
	diContainer := di.NewContainer()

	// Create CLI context with proper flags for DI
	app := &cli.App{}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("log-level", "info", "")
	flagSet.Int("complexity-threshold", 5, "")
	flagSet.Int("max-depth", 10, "")
	flagSet.Int("max-tasks-per-depth", 50, "")
	flagSet.Int("max-description-length", 500, "")
	flagSet.Bool("auto-reduce-complexity", true, "")

	cliCtx := cli.NewContext(app, flagSet, nil)
	_ = diContainer.RegisterAllServices(cliCtx)

	// Store container in CLI context metadata like BeforeCommand does
	if app.Metadata == nil {
		app.Metadata = make(map[string]interface{})
	}
	app.Metadata["container"] = diContainer

	return cliCtx, app, diContainer
}

func TestCreateActionValidation(t *testing.T) {
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
			cliCtx, app, _ := setupCLIContextForValidation(t)

			// Get project manager from DI container
			diContainer := shared.GetContainerFromContext(cliCtx)
			injector := diContainer.GetInjector()
			projectManager := do.MustInvoke[manager.ProjectManager](injector)

			// Create a test project and set it as selected
			project := testutil.CreateTestProject(t, projectManager)
			err := projectManager.SetSelectedProject(cliCtx.Context, project.ID, "test-user")
			require.NoError(t, err)

			// Set up flags for task creation
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

			// Update context with new flagSet
			cliCtx = cli.NewContext(app, flagSet, nil)
			// Preserve the container metadata
			cliCtx.App.Metadata["container"] = diContainer

			action := createAction()

			// Execute the action
			err = action(cliCtx)

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
	cliCtx, app, _ := setupCLIContextForValidation(t)

	// Get project manager from DI container
	diContainer := shared.GetContainerFromContext(cliCtx)
	injector := diContainer.GetInjector()
	projectManager := do.MustInvoke[manager.ProjectManager](injector)

	// Create a test project and set it as selected
	project := testutil.CreateTestProject(t, projectManager)
	err := projectManager.SetSelectedProject(cliCtx.Context, project.ID, "test-user")
	require.NoError(t, err)

	// Set up flags for task creation
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

	// Update context with new flagSet
	cliCtx = cli.NewContext(app, flagSet, nil)
	// Preserve the container metadata
	cliCtx.App.Metadata["container"] = diContainer

	action := createAction()

	err = action(cliCtx)
	require.Error(t, err)

	// Check that it's wrapped as a validation error
	assert.Contains(t, err.Error(), "title contains HTML tags")
	assert.Contains(t, err.Error(), "not allowed")
}
