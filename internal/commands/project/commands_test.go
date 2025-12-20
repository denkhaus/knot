package project

import (
	"flag"
	"testing"

	"github.com/denkhaus/knot/v2/internal/di"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

// setupCLIContextWithDI creates a CLI context with proper DI container setup
// This helper function prevents code duplication across tests
func setupCLIContextWithDI(t *testing.T) (*cli.Context, *cli.App) {
	// Use testutil to create isolated container with in-memory database
	config := testutil.NewTestConfig(t)
	diContainer := di.NewContainer()

	// Create CLI context with proper flags for DI
	app := &cli.App{}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flagSet.String("log-level", "info", "Log level")
	flagSet.Int("complexity-threshold", 5, "Complexity threshold")
	flagSet.Int("max-depth", 10, "Max depth")
	flagSet.Int("max-tasks-per-depth", 50, "Max tasks per depth")
	flagSet.Int("max-description-length", 500, "Max description length")
	flagSet.Bool("auto-reduce-complexity", true, "Auto reduce complexity")

	cliCtx := cli.NewContext(app, flagSet, nil)
	_ = diContainer.RegisterAllServices(cliCtx)

	// Override repository with in-memory database for test isolation
	injector := diContainer.GetInjector()
	do.Override(injector, func(do.Injector) (types.Repository, error) {
		return config.SetupTestRepository(t), nil
	})

	// Store container in CLI context metadata like BeforeCommand does
	if app.Metadata == nil {
		app.Metadata = make(map[string]interface{})
	}
	app.Metadata["container"] = diContainer

	return cliCtx, app
}

func TestCreateAction(t *testing.T) {

	tests := []struct {
		name        string
		title       string
		description string
		actor       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid project creation",
			title:       "Test Project",
			description: "Test Description",
			actor:       "test-user",
			expectError: false,
		},
		{
			name:        "empty title should fail",
			title:       "",
			description: "Test Description",
			actor:       "test-user",
			expectError: true,
			errorMsg:    "project title cannot be empty",
		},
		{
			name:        "title too long should fail",
			title:       string(make([]rune, 201)),
			description: "Test Description",
			actor:       "test-user",
			expectError: true,
			errorMsg:    "project title too long",
		},
		{
			name:        "HTML in title should fail",
			title:       "Project <script>alert('xss')</script>",
			description: "Test Description",
			actor:       "test-user",
			expectError: true,
			errorMsg:    "contains HTML tags",
		},
		{
			name:        "description too long should fail",
			title:       "Test Project",
			description: string(make([]rune, 2001)),
			actor:       "test-user",
			expectError: true,
			errorMsg:    "project description too long",
		},
		{
			name:        "default actor when empty",
			title:       "Test Project",
			description: "Test Description",
			actor:       "", // Should default to $USER
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up DI container for this test case
			diContainer := di.NewContainer()

			// Create CLI context
			app := &cli.App{}
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			flagSet.String("title", "", "")
			flagSet.String("description", "", "")
			flagSet.String("actor", "", "")
			// Add DI flags
			flagSet.String("log-level", "info", "Log level")
			flagSet.Int("complexity-threshold", 5, "Complexity threshold")
			flagSet.Int("max-depth", 10, "Max depth")
			flagSet.Int("max-tasks-per-depth", 50, "Max tasks per depth")
			flagSet.Int("max-description-length", 500, "Max description length")
			flagSet.Bool("auto-reduce-complexity", true, "Auto reduce complexity")

			_ = flagSet.Set("title", tt.title)
			_ = flagSet.Set("description", tt.description)
			_ = flagSet.Set("actor", tt.actor)

			ctx := cli.NewContext(app, flagSet, nil)
			_ = diContainer.RegisterAllServices(ctx)

			// Store container in CLI context metadata like BeforeCommand does
			if app.Metadata == nil {
				app.Metadata = make(map[string]interface{})
			}
			app.Metadata["container"] = diContainer

			// Execute action
			action := createAction()
			err := action(ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListAction(t *testing.T) {
	t.Run("empty project list", func(t *testing.T) {
		// Use isolated test config with in-memory database
		cliCtx, _ := setupCLIContextWithDI(t)

		// Execute action
		action := listAction()
		err := action(cliCtx)

		// Should return EmptyResultError since we're using isolated in-memory database
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no results found")
	})

	t.Run("list with projects", func(t *testing.T) {
		// Use isolated test config with in-memory database, but create projects in the same container
		cliCtx, _ := setupCLIContextWithDI(t)
		diContainer := shared.GetContainerFromContext(cliCtx)
		projectManager := do.MustInvoke[manager.ProjectManager](diContainer.GetInjector())

		// Create test projects in this isolated manager
		testutil.CreateTestProject(t, projectManager)
		testutil.CreateTestProject(t, projectManager)

		// Execute action using the same context
		action := listAction()
		err := action(cliCtx)

		// Should succeed
		assert.NoError(t, err)
	})
}

func TestGetAction(t *testing.T) {
	tests := []struct {
		name        string
		projectID   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid project ID",
			projectID:   "", // Will be set in test
			expectError: false,
		},
		{
			name:        "invalid project ID format",
			projectID:   "invalid-uuid",
			expectError: true,
			errorMsg:    "invalid UUID length",
		},
		{
			name:        "non-existent project ID",
			projectID:   "123e4567-e89b-12d3-a456-426614174000",
			expectError: true,
			errorMsg:    "failed to get project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var projectID string

			// For valid project ID test, create a project first using setupCLIContextWithDI
			if tt.name == "valid project ID" {
				// Create a project using the same DI container that will be used for getAction
				cliCtx, app := setupCLIContextWithDI(t)
				diContainer := shared.GetContainerFromContext(cliCtx)
				projectManager := do.MustInvoke[manager.ProjectManager](diContainer.GetInjector())
				project := testutil.CreateTestProject(t, projectManager)
				projectID = project.ID.String()

				// Now set up the get action test with the same container
				flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
				flagSet.String("id", "", "")
				// Add DI flags
				flagSet.String("log-level", "info", "Log level")
				flagSet.Int("complexity-threshold", 5, "Complexity threshold")
				flagSet.Int("max-depth", 10, "Max depth")
				flagSet.Int("max-tasks-per-depth", 50, "Max tasks per depth")
				flagSet.Int("max-description-length", 500, "Max description length")
				flagSet.Bool("auto-reduce-complexity", true, "Auto reduce complexity")

				_ = flagSet.Set("id", projectID)

				getCtx := cli.NewContext(app, flagSet, nil)

				// Execute action
				action := getAction()
				err := action(getCtx)

				if tt.expectError {
					require.Error(t, err)
					assert.Contains(t, err.Error(), tt.errorMsg)
				} else {
					assert.NoError(t, err)
				}
			} else {
				// For error cases (invalid ID format, non-existent), use separate DI container
				projectID = tt.projectID

				diContainer := di.NewContainer()

				// Create CLI context
				app := &cli.App{}
				flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
				flagSet.String("id", "", "")
				// Add DI flags
				flagSet.String("log-level", "info", "Log level")
				flagSet.Int("complexity-threshold", 5, "Complexity threshold")
				flagSet.Int("max-depth", 10, "Max depth")
				flagSet.Int("max-tasks-per-depth", 50, "Max tasks per depth")
				flagSet.Int("max-description-length", 500, "Max description length")
				flagSet.Bool("auto-reduce-complexity", true, "Auto reduce complexity")

				_ = flagSet.Set("id", projectID)

				ctx := cli.NewContext(app, flagSet, nil)
				_ = diContainer.RegisterAllServices(ctx)

				// Override repository with in-memory database for test isolation
				injector := diContainer.GetInjector()
				config := testutil.NewTestConfig(t)
				do.Override(injector, func(do.Injector) (types.Repository, error) {
					return config.SetupTestRepository(t), nil
				})

				// Store container in CLI context metadata like BeforeCommand does
				if app.Metadata == nil {
					app.Metadata = make(map[string]interface{})
				}
				app.Metadata["container"] = diContainer

				// Execute action
				action := getAction()
				err := action(ctx)

				if tt.expectError {
					require.Error(t, err)
					assert.Contains(t, err.Error(), tt.errorMsg)
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestProjectCommandsIntegration(t *testing.T) {
	// Integration test to ensure all project commands work together
	// Test complete workflow: create -> list -> get
	t.Run("complete project workflow", func(t *testing.T) {
		// Use setupCLIContextWithDI to create isolated container with in-memory database
		cliCtx, app := setupCLIContextWithDI(t)
		diContainer := shared.GetContainerFromContext(cliCtx)
		projectManager := do.MustInvoke[manager.ProjectManager](diContainer.GetInjector())

		// 1. Create project
		createFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		createFlagSet.String("title", "", "")
		createFlagSet.String("description", "", "")
		createFlagSet.String("actor", "", "Actor")
		// Add DI flags
		createFlagSet.String("log-level", "info", "Log level")
		createFlagSet.Int("complexity-threshold", 5, "Complexity threshold")
		createFlagSet.Int("max-depth", 10, "Max depth")
		createFlagSet.Int("max-tasks-per-depth", 50, "Max tasks per depth")
		createFlagSet.Int("max-description-length", 500, "Max description length")
		createFlagSet.Bool("auto-reduce-complexity", true, "Auto reduce complexity")

		_ = createFlagSet.Set("title", "Integration Test Project")
		_ = createFlagSet.Set("description", "Created during integration test")
		_ = createFlagSet.Set("actor", "integration-test")

		createCtx := cli.NewContext(app, createFlagSet, nil)
		createActionFunc := createAction()
		err := createActionFunc(createCtx)
		require.NoError(t, err)

		// Get the created project ID for the get test
		projects, err := projectManager.ListProjects(createCtx.Context)
		require.NoError(t, err)
		require.NotEmpty(t, projects)
		createdProjectID := projects[0].ID.String()

		// 2. List projects (should now have at least one)
		listFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		listCtx := cli.NewContext(app, listFlagSet, nil)

		listActionFunc := listAction()
		err = listActionFunc(listCtx)
		assert.NoError(t, err)

		// 3. Get project details
		getFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		getFlagSet.String("id", "", "")
		_ = getFlagSet.Set("id", createdProjectID)

		getCtx := cli.NewContext(app, getFlagSet, nil)

		getActionFunc := getAction()
		err = getActionFunc(getCtx)
		assert.NoError(t, err)
	})
}

// TestListActionIntegration has been integrated into TestListAction and TestProjectCommandsIntegration
// to avoid test duplication and ensure proper isolation
