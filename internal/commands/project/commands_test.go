package project

import (
	"flag"
	"testing"

	"github.com/denkhaus/knot/internal/shared"
	"github.com/denkhaus/knot/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestCreateAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

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
			// Create CLI context
			app := &cli.App{}
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			flagSet.String("title", "", "")
			flagSet.String("description", "", "")
			flagSet.String("actor", "", "")
			
			_ = flagSet.Set("title", tt.title)
			_ = flagSet.Set("description", tt.description)
			_ = flagSet.Set("actor", tt.actor)
			
			ctx := cli.NewContext(app, flagSet, nil)

			// Create app context
			appCtx := &shared.AppContext{
				ProjectManager: mgr,
				Logger:         config.Logger,
			}

			// Execute action
			action := createAction(appCtx)
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

	// Create app context
	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	t.Run("empty project list", func(t *testing.T) {
		// Create CLI context
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		ctx := cli.NewContext(app, flagSet, nil)

		// Execute action
		action := listAction(appCtx)
		err := action(ctx)

		// Should return EmptyResultError
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no results found")
	})

	t.Run("list with projects", func(t *testing.T) {
		// Create test projects
		testutil.CreateTestProject(t, mgr)
		testutil.CreateTestProject(t, mgr)

		// Create CLI context
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		ctx := cli.NewContext(app, flagSet, nil)

		// Execute action
		action := listAction(appCtx)
		err := action(ctx)

		// Should succeed
		assert.NoError(t, err)
	})
}

func TestGetAction(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)
	project := testutil.CreateTestProject(t, mgr)

	// Create app context
	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

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
			// Create CLI context
			app := &cli.App{}
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			flagSet.String("id", "", "")
			_ = flagSet.Set("id", tt.projectID)
			
			ctx := cli.NewContext(app, flagSet, nil)

			// Execute action
			action := getAction(appCtx)
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

// TestValidateProjectIDFunction is deprecated - project validation now happens through context
// This test is kept for backward compatibility but should be removed in future versions
func TestValidateProjectIDFunction_Deprecated(t *testing.T) {
	t.Skip("Skipping deprecated validateProjectID test - project validation now uses context system")
}

func TestProjectCommandsIntegration(t *testing.T) {
	// Integration test to ensure all project commands work together
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	// Test complete workflow: create -> list -> get
	t.Run("complete project workflow", func(t *testing.T) {
		// 1. Create project
		app := &cli.App{}
		flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		flagSet.String("title", "", "")
		flagSet.String("description", "", "")
		flagSet.String("actor", "", "")
		
		_ = flagSet.Set("title", "Integration Test Project")
		_ = flagSet.Set("description", "Created during integration test")
		_ = flagSet.Set("actor", "integration-test")
		
		ctx := cli.NewContext(app, flagSet, nil)

		createActionFunc := createAction(appCtx)
		err := createActionFunc(ctx)
		require.NoError(t, err)

		// 2. List projects (should now have at least one)
		listFlagSet := flag.NewFlagSet("test", flag.ContinueOnError)
		listCtx := cli.NewContext(app, listFlagSet, nil)

		listActionFunc := listAction(appCtx)
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

		getActionFunc := getAction(appCtx)
		err = getActionFunc(getCtx)
		assert.NoError(t, err)
	})
}

func TestListActionIntegration(t *testing.T) {
	config := testutil.NewTestConfig(t)
	mgr := config.SetupTestManager(t)

	// Create app context
	appCtx := &shared.AppContext{
		ProjectManager: mgr,
		Logger:         config.Logger,
	}

	// Create a test app
	app := cli.NewApp()

	t.Run("empty project list", func(t *testing.T) {
		listFlagSet := flag.NewFlagSet("list", flag.ContinueOnError)
		listCtx := cli.NewContext(app, listFlagSet, nil)

		listActionFunc := listAction(appCtx)
		err := listActionFunc(listCtx)

		// Should return an error about no projects found
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no results found")
	})

	t.Run("list after creating project", func(t *testing.T) {
		// First, create a project using the manager
		_ = testutil.CreateTestProject(t, mgr)

		// Now try to list projects
		listFlagSet := flag.NewFlagSet("list", flag.ContinueOnError)
		listCtx := cli.NewContext(app, listFlagSet, nil)

		listActionFunc := listAction(appCtx)
		err := listActionFunc(listCtx)

		// This should succeed and find our project
		assert.NoError(t, err, "listAction should succeed after creating project")
	})

	t.Run("create then list integration test", func(t *testing.T) {
		// Use a fresh manager to test the create/list inconsistency
		config2 := testutil.NewTestConfig(t)
		mgr2 := config2.SetupTestManager(t)

		appCtx2 := &shared.AppContext{
			ProjectManager: mgr2,
			Logger:         config2.Logger,
		}

		// Create a project using createAction
		createFlagSet := flag.NewFlagSet("create", flag.ContinueOnError)
		createFlagSet.String("title", "Integration Test Project", "Project title")
		createFlagSet.String("description", "Testing create/list integration", "Project description")
		createFlagSet.String("actor", "test-user", "Actor")

		createCtx := cli.NewContext(app, createFlagSet, nil)
		createActionFunc := createAction(appCtx2)
		err := createActionFunc(createCtx)
		assert.NoError(t, err, "createAction should succeed")

		// Now try to list projects
		listFlagSet := flag.NewFlagSet("list", flag.ContinueOnError)
		listCtx := cli.NewContext(app, listFlagSet, nil)

		listActionFunc := listAction(appCtx2)
		err = listActionFunc(listCtx)

		// This should succeed and find the project we just created
		assert.NoError(t, err, "listAction should find the project created by createAction")
	})

	t.Run("multiple projects list", func(t *testing.T) {
		// Create multiple projects
		for i := 0; i < 3; i++ {
			_ = testutil.CreateTestProject(t, mgr)
		}

		// List all projects
		listFlagSet := flag.NewFlagSet("list", flag.ContinueOnError)
		listCtx := cli.NewContext(app, listFlagSet, nil)

		listActionFunc := listAction(appCtx)
		err := listActionFunc(listCtx)

		// Should succeed
		assert.NoError(t, err, "listAction should succeed with multiple projects")
	})
}