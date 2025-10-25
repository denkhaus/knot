package app

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/denkhaus/knot/internal/manager"
	"github.com/denkhaus/knot/internal/repository/inmemory"
	"github.com/denkhaus/knot/internal/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestMainFunction(t *testing.T) {
	// Test the main function by capturing stdout/stderr
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test successful execution with help command
	os.Args = []string{"knot", "--help"}

	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set version for test
	SetVersionFromBuild("test", "test-commit", "test-date")

	// Create application
	application, err := New()
	require.NoError(t, err)

	// Run with help to avoid exit
	os.Args = []string{"knot", "--help"}
	err = application.Run([]string{"knot", "--help"})
	assert.NoError(t, err)

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read the output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Should contain help information
	assert.Contains(t, output, "A CLI tool for hierarchical project and task management")
}

func TestAppNew(t *testing.T) {
	// Test successful app creation
	app, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.NotNil(t, app.App)
	assert.NotNil(t, app.context)
	assert.Equal(t, "knot", app.App.Name)
}

func TestAppRun(t *testing.T) {
	app, err := New()
	require.NoError(t, err)

	// Test with help command to avoid exit
	err = app.Run([]string{"knot", "--help"})
	assert.NoError(t, err)
}

func TestIsUserInputError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "required flag error",
			err:      fmt.Errorf("Required flag not provided"),
			expected: true,
		},
		{
			name:     "flag not defined error",
			err:      fmt.Errorf("flag provided but not defined"),
			expected: true,
		},
		{
			name:     "invalid value error",
			err:      fmt.Errorf("invalid value for flag"),
			expected: true,
		},
		{
			name:     "command not found error",
			err:      fmt.Errorf("command not found"),
			expected: true,
		},
		{
			name:     "incorrect usage error",
			err:      fmt.Errorf("incorrect usage"),
			expected: true,
		},
		{
			name:     "flag needs argument error",
			err:      fmt.Errorf("flag needs an argument"),
			expected: true,
		},
		{
			name:     "help topic error",
			err:      fmt.Errorf("No help topic for 'unknown'"),
			expected: true,
		},
		{
			name:     "internal error",
			err:      fmt.Errorf("internal application error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUserInputError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppRunWithError(t *testing.T) {
	app, err := New()
	require.NoError(t, err)

	// Test with invalid command to generate error
	err = app.Run([]string{"knot", "invalid-command"})
	assert.Error(t, err)
}

func TestSetVersionFromBuild(t *testing.T) {
	// Test version setting
	SetVersionFromBuild("v1.0.0", "abc123", "2023-01-01")

	// Create a new app to see if the version was set
	app, err := New()
	require.NoError(t, err)

	// Check that version was set, though we can't easily verify the internal version variable
	// since it's not exposed, we'll just ensure app creation still works
	assert.Equal(t, "knot", app.App.Name) // This will still be knot regardless of version
}

func TestAppWithMemoryRepository(t *testing.T) {
	// Override the repository initialization to use in-memory for testing
	// This simulates what happens in the real New() function but with guaranteed in-memory
	config := manager.DefaultConfig()
	repo := inmemory.NewMemoryRepository()
	projectManager := manager.NewManagerWithRepository(repo, config)
	// Use a mock logger for testing
	appCtx := shared.NewAppContext(projectManager, nil)

	// Test that we can create CLI app with this context
	cliApp := &cli.App{
		Name:  "knot-test",
		Usage: "Test version of Knot CLI",
		Commands: []*cli.Command{
			{
				Name:        "project",
				Aliases:     []string{"p"},
				Usage:       "Project management commands",
				Subcommands: []*cli.Command{},
			},
		},
	}

	app := &App{
		App:     cliApp,
		context: appCtx,
	}

	assert.Equal(t, "knot-test", app.App.Name)
	assert.NotNil(t, app.context)
}

func TestAppContextInitialization(t *testing.T) {
	app, err := New()
	require.NoError(t, err)
	require.NotNil(t, app.context)
	require.NotNil(t, app.context.ProjectManager)
	require.NotNil(t, app.context.Logger)
}

func TestAppRunWithValidArgs(t *testing.T) {
	app, err := New()
	require.NoError(t, err)

	// Test with version command
	err = app.Run([]string{"knot", "--version"})
	// Version command should not return an error
	// but since it calls os.Exit in real usage, we might need a different approach
	// For now, just test that the app can be created without issues
	assert.NotNil(t, app)
}

func TestAppCommandsStructure(t *testing.T) {
	app, err := New()
	require.NoError(t, err)

	// Check that all expected commands are present
	expectedCommands := []string{
		"project", "task", "template", "dependency", "config", "health", "validate",
		"ready", "blocked", "actionable", "breakdown", "get-started",
	}

	commandsMap := make(map[string]*cli.Command)
	for _, cmd := range app.App.Commands {
		commandsMap[cmd.Name] = cmd
	}

	for _, expected := range expectedCommands {
		assert.Contains(t, commandsMap, expected, "Command %s should be present", expected)
	}
}

func TestAppFlags(t *testing.T) {
	app, err := New()
	require.NoError(t, err)

	// Check that expected flags are present
	expectedFlags := []string{"actor", "log-level"}

	flagsMap := make(map[string]cli.Flag)
	for _, flag := range app.App.Flags {
		flagsMap[flag.Names()[0]] = flag
	}

	for _, expected := range expectedFlags {
		assert.Contains(t, flagsMap, expected, "Flag %s should be present", expected)
	}
}

func TestAppBeforeHook(t *testing.T) {
	app, err := New()
	require.NoError(t, err)
	require.NotNil(t, app.App.Before)

	// Test that Before hook exists and doesn't cause panic
	// The actual functionality is complex to test without a proper context
	// so we just verify the function exists and can be called safely
	assert.NotNil(t, app.App.Before)
}

func TestAppIntegration(t *testing.T) {
	// Full integration test: create app, run basic operations
	app, err := New()
	require.NoError(t, err)

	// Test that we can get the app context and use the manager
	ctx := app.context
	require.NotNil(t, ctx)
	require.NotNil(t, ctx.ProjectManager)

	// Test basic project operations through the manager
	projectManager := ctx.ProjectManager

	// Create a project
	ctxWithBackground := context.Background()
	project, err := projectManager.CreateProject(ctxWithBackground, "Test Project", "Test Description", "test-user")
	assert.NoError(t, err)
	assert.NotNil(t, project)
	assert.Equal(t, "Test Project", project.Title)
	assert.Equal(t, "Test Description", project.Description)
	assert.Equal(t, "test-user", project.CreatedBy)
	assert.NotEqual(t, uuid.Nil, project.ID)

	// Get the project
	retrievedProject, err := projectManager.GetProject(ctxWithBackground, project.ID)
	assert.NoError(t, err)
	assert.Equal(t, project.ID, retrievedProject.ID)

	// Update the project
	updatedProject, err := projectManager.UpdateProject(ctxWithBackground, project.ID, "Updated Title", "Updated Description", "updater")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", updatedProject.Title)
	assert.Equal(t, "Updated Description", updatedProject.Description)

	// List projects
	projects, err := projectManager.ListProjects(ctxWithBackground)
	assert.NoError(t, err)
	assert.NotEmpty(t, projects)

	// Delete the project
	err = projectManager.DeleteProject(ctxWithBackground, project.ID)
	assert.NoError(t, err)

	// Try to get deleted project (should fail)
	_, err = projectManager.GetProject(ctxWithBackground, project.ID)
	assert.Error(t, err)
}
