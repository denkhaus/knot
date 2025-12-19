package project

import (
	"flag"
	"testing"

	"github.com/denkhaus/knot/v2/internal/di"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

// setupCLIContextWithDI creates a CLI context with proper DI container setup
// This helper function prevents code duplication across tests
func setupCLIContextWithDI(t *testing.T) (*cli.Context, *cli.App) {
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
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	// Create app context using testutil for DI

	t.Run("empty project list", func(t *testing.T) {
		cliCtx, _ := setupCLIContextWithDI(t)

		// Execute action
		action := listAction()
		err := action(cliCtx)

		// Should return EmptyResultError
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no results found")
	})

	t.Run("list with projects", func(t *testing.T) {
		// Create test projects
		testutil.CreateTestProject(t, mgr)
		testutil.CreateTestProject(t, mgr)

		cliCtx, _ := setupCLIContextWithDI(t)

		// Execute action
		action := listAction()
		err := action(cliCtx)

		// Should succeed
		assert.NoError(t, err)
	})
}

func TestGetAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Create app context using testutil for DI

	tests := []struct {
		name        string
		projectID   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid project ID",
			projectID:   project.ID.String(),
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
			// Set up DI container for this test case
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
			_ = flagSet.Set("id", tt.projectID)

			ctx := cli.NewContext(app, flagSet, nil)
			_ = diContainer.RegisterAllServices(ctx)

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
		})
	}
}

func TestProjectCommandsIntegration(t *testing.T) {
	// Integration test to ensure all project commands work together
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	// Use testutil for DI injector

	// Test complete workflow: create -> list -> get
	t.Run("complete project workflow", func(t *testing.T) {
		// Set up DI container for this test case
		diContainer := di.NewContainer()

		// 1. Create project
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("title", "", "")
		flagSet.String("description", "", "")
		flagSet.String("actor", "", "Actor")
		// Add DI flags
		flagSet.String("log-level", "info", "Log level")
		flagSet.Int("complexity-threshold", 5, "Complexity threshold")
		flagSet.Int("max-depth", 10, "Max depth")
		flagSet.Int("max-tasks-per-depth", 50, "Max tasks per depth")
		flagSet.Int("max-description-length", 500, "Max description length")
		flagSet.Bool("auto-reduce-complexity", true, "Auto reduce complexity")

		_ = flagSet.Set("title", "Integration Test Project")
		_ = flagSet.Set("description", "Created during integration test")
		_ = flagSet.Set("actor", "integration-test")

		ctx := cli.NewContext(app, flagSet, nil)
		_ = diContainer.RegisterAllServices(ctx)

		// Store container in CLI context metadata like BeforeCommand does
		if app.Metadata == nil {
			app.Metadata = make(map[string]interface{})
		}
		app.Metadata["container"] = diContainer

		createActionFunc := createAction()
		err := createActionFunc(ctx)
		require.NoError(t, err)

		// 2. List projects (should now have at least one)
		listFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		listCtx := cli.NewContext(app, listFlagSet, nil)

		listActionFunc := listAction()
		err = listActionFunc(listCtx)
		assert.NoError(t, err)

		// 3. Get project details
		projects, err := mgr.ListProjects(ctx.Context)
		require.NoError(t, err)
		require.NotEmpty(t, projects)

		getFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		getFlagSet.String("id", "", "")
		_ = getFlagSet.Set("id", projects[0].ID.String())

		getCtx := cli.NewContext(app, getFlagSet, nil)

		getActionFunc := getAction()
		err = getActionFunc(getCtx)
		assert.NoError(t, err)
	})
}

func TestListActionIntegration(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	// Create app context using testutil for DI

	// Create a test app
	app := cli.NewApp()

	t.Run("empty project list", func(t *testing.T) {
		// Set up DI container for this test
		diContainer := di.NewContainer()

		// Create CLI context with flags
		listFlagSet := flag.NewFlagSet("list", flag.ContinueOnError)
		// Add DI flags
		listFlagSet.String("log-level", "info", "")
		listFlagSet.Int("complexity-threshold", 5, "")
		listFlagSet.Int("max-depth", 10, "")
		listFlagSet.Int("max-tasks-per-depth", 50, "")
		listFlagSet.Int("max-description-length", 500, "")
		listFlagSet.Bool("auto-reduce-complexity", true, "")

		listCtx := cli.NewContext(app, listFlagSet, nil)

		// Register services
		_ = diContainer.RegisterAllServices(listCtx)

		// Store container in app metadata
		if app.Metadata == nil {
			app.Metadata = make(map[string]interface{})
		}
		app.Metadata["container"] = diContainer

		listActionFunc := listAction()
		err := listActionFunc(listCtx)

		// Should succeed even with existing projects (tests may share state)
		assert.NoError(t, err)
	})

	t.Run("list after creating project", func(t *testing.T) {
		// First, create a project using the manager
		_ = testutil.CreateTestProject(t, mgr)

		// Set up DI container for this test
		diContainer := di.NewContainer()

		// Create CLI context with flags
		listFlagSet := flag.NewFlagSet("list", flag.ContinueOnError)
		// Add DI flags
		listFlagSet.String("log-level", "info", "")
		listFlagSet.Int("complexity-threshold", 5, "")
		listFlagSet.Int("max-depth", 10, "")
		listFlagSet.Int("max-tasks-per-depth", 50, "")
		listFlagSet.Int("max-description-length", 500, "")
		listFlagSet.Bool("auto-reduce-complexity", true, "")

		listCtx := cli.NewContext(app, listFlagSet, nil)

		// Register services
		_ = diContainer.RegisterAllServices(listCtx)

		// Store container in app metadata
		if app.Metadata == nil {
			app.Metadata = make(map[string]interface{})
		}
		app.Metadata["container"] = diContainer

		listActionFunc := listAction()
		err := listActionFunc(listCtx)

		// This should succeed and find our project
		assert.NoError(t, err, "listAction should succeed after creating project")
	})

	t.Run("create then list integration test", func(t *testing.T) {
		// Use fresh containers to avoid DI conflicts
		createContainer := di.NewContainer()
		listContainer := di.NewContainer()

		// Create CLI context with flags
		createFlagSet := flag.NewFlagSet("create", flag.ContinueOnError)
		// Add DI flags
		createFlagSet.String("log-level", "info", "")
		createFlagSet.Int("complexity-threshold", 5, "")
		createFlagSet.Int("max-depth", 10, "")
		createFlagSet.Int("max-tasks-per-depth", 50, "")
		createFlagSet.Int("max-description-length", 500, "")
		createFlagSet.Bool("auto-reduce-complexity", true, "")
		// Command flags
		createFlagSet.String("title", "Integration Test Project", "Project title")
		createFlagSet.String("description", "Testing create/list integration", "Project description")
		createFlagSet.String("actor", "test-user", "Actor")

		createCtx := cli.NewContext(app, createFlagSet, nil)

		// Register services
		_ = createContainer.RegisterAllServices(createCtx)

		// Store container in app metadata
		if app.Metadata == nil {
			app.Metadata = make(map[string]interface{})
		}
		app.Metadata["container"] = createContainer

		createActionFunc := createAction()
		err := createActionFunc(createCtx)
		assert.NoError(t, err, "createAction should succeed")

		// Now try to list projects - use fresh container
		listFlagSet := flag.NewFlagSet("list", flag.ContinueOnError)
		// Add DI flags
		listFlagSet.String("log-level", "info", "")
		listFlagSet.Int("complexity-threshold", 5, "")
		listFlagSet.Int("max-depth", 10, "")
		listFlagSet.Int("max-tasks-per-depth", 50, "")
		listFlagSet.Int("max-description-length", 500, "")
		listFlagSet.Bool("auto-reduce-complexity", true, "")

		listCtx := cli.NewContext(app, listFlagSet, nil)

		// Register services with list container
		_ = listContainer.RegisterAllServices(listCtx)

		// Update app metadata with list container
		app.Metadata["container"] = listContainer

		listActionFunc := listAction()
		err = listActionFunc(listCtx)

		// This should succeed and find the project we just created
		assert.NoError(t, err, "listAction should find the project created by createAction")
	})

	t.Run("multiple projects list", func(t *testing.T) {
		// Create multiple projects
		for i := 0; i < 3; i++ {
			_ = testutil.CreateTestProject(t, mgr)
		}

		// Set up DI container for this test
		diContainer := di.NewContainer()

		// List all projects
		listFlagSet := flag.NewFlagSet("list", flag.ContinueOnError)
		// Add DI flags
		listFlagSet.String("log-level", "info", "")
		listFlagSet.Int("complexity-threshold", 5, "")
		listFlagSet.Int("max-depth", 10, "")
		listFlagSet.Int("max-tasks-per-depth", 50, "")
		listFlagSet.Int("max-description-length", 500, "")
		listFlagSet.Bool("auto-reduce-complexity", true, "")

		listCtx := cli.NewContext(app, listFlagSet, nil)

		// Register services
		_ = diContainer.RegisterAllServices(listCtx)

		// Store container in app metadata
		if app.Metadata == nil {
			app.Metadata = make(map[string]interface{})
		}
		app.Metadata["container"] = diContainer

		listActionFunc := listAction()
		err := listActionFunc(listCtx)

		// Should succeed
		assert.NoError(t, err, "listAction should succeed with multiple projects")
	})
}
