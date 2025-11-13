package project

import (
	"context"
	"testing"

	"github.com/denkhaus/knot/internal/manager"
	"github.com/denkhaus/knot/internal/repository/inmemory"
	"github.com/denkhaus/knot/internal/shared"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func TestProjectSelectCommand(t *testing.T) {
	// Setup test environment
	repo := inmemory.NewMemoryRepository()
	config := manager.DefaultConfig()
	projectManager := manager.NewManagerWithRepository(repo, config)
	logger := zap.NewNop()
	appCtx := shared.NewAppContext(projectManager, logger)

	// Create a test project
	project, err := projectManager.CreateProject(context.Background(), "Test Project", "Test Description", "test-actor")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	t.Run("Select existing project", func(t *testing.T) {
		// Create CLI app with select command
		app := &cli.App{
			Commands: []*cli.Command{
				{
					Name:   "select",
					Action: selectAction(appCtx),
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     "id",
							Required: true,
						},
					},
				},
			},
		}

		// Test selecting the project
		args := []string{"test", "select", "--id", project.ID.String()}
		err := app.Run(args)
		if err != nil {
			t.Errorf("Failed to select project: %v", err)
		}

		// Verify project is selected
		selectedID, err := projectManager.GetSelectedProject(context.Background())
		if err != nil {
			t.Errorf("Failed to get selected project: %v", err)
		}
		if selectedID == nil || *selectedID != project.ID {
			t.Errorf("Expected selected project %v, got %v", project.ID, selectedID)
		}
	})

	t.Run("Select non-existent project", func(t *testing.T) {
		// Create CLI app with select command
		app := &cli.App{
			Commands: []*cli.Command{
				{
					Name:   "select",
					Action: selectAction(appCtx),
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     "id",
							Required: true,
						},
					},
				},
			},
		}

		// Test selecting non-existent project
		args := []string{"test", "select", "--id", "550e8400-e29b-41d4-a716-446655440000"}
		err := app.Run(args)
		if err == nil {
			t.Errorf("Expected error when selecting non-existent project")
		}
	})

	t.Run("Select with invalid UUID", func(t *testing.T) {
		// Create CLI app with select command
		app := &cli.App{
			Commands: []*cli.Command{
				{
					Name:   "select",
					Action: selectAction(appCtx),
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     "id",
							Required: true,
						},
					},
				},
			},
		}

		// Test selecting with invalid UUID
		args := []string{"test", "select", "--id", "invalid-uuid"}
		err := app.Run(args)
		if err == nil {
			t.Errorf("Expected error when selecting with invalid UUID")
		}
	})
}

func TestProjectGetSelectedCommand(t *testing.T) {
	// Setup test environment
	repo := inmemory.NewMemoryRepository()
	config := manager.DefaultConfig()
	projectManager := manager.NewManagerWithRepository(repo, config)
	logger := zap.NewNop()
	appCtx := shared.NewAppContext(projectManager, logger)

	// Create a test project
	project, err := projectManager.CreateProject(context.Background(), "Test Project", "Test Description", "test-actor")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	t.Run("Get selected when none selected", func(t *testing.T) {
		// Ensure no project is selected
		_ = projectManager.ClearSelectedProject(context.Background())

		// Create CLI app with get-selected command
		app := &cli.App{
			Commands: []*cli.Command{
				{
					Name:   "get-selected",
					Action: getSelectedAction(appCtx),
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name: "json",
						},
					},
				},
			},
		}

		// Test getting selected project when none is selected
		args := []string{"test", "get-selected"}
		err := app.Run(args)
		if err != nil {
			t.Errorf("Unexpected error when no project selected: %v", err)
		}
	})

	t.Run("Get selected when project is selected", func(t *testing.T) {
		// Select a project
		err := projectManager.SetSelectedProject(context.Background(), project.ID, "test-actor")
		if err != nil {
			t.Fatalf("Failed to select project: %v", err)
		}

		// Create CLI app with get-selected command
		app := &cli.App{
			Commands: []*cli.Command{
				{
					Name:   "get-selected",
					Action: getSelectedAction(appCtx),
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name: "json",
						},
					},
				},
			},
		}

		// Test getting selected project
		args := []string{"test", "get-selected"}
		err = app.Run(args)
		if err != nil {
			t.Errorf("Unexpected error when getting selected project: %v", err)
		}
	})
}

func TestProjectClearSelectionCommand(t *testing.T) {
	// Setup test environment
	repo := inmemory.NewMemoryRepository()
	config := manager.DefaultConfig()
	projectManager := manager.NewManagerWithRepository(repo, config)
	logger := zap.NewNop()
	appCtx := shared.NewAppContext(projectManager, logger)

	// Create a test project
	project, err := projectManager.CreateProject(context.Background(), "Test Project", "Test Description", "test-actor")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	t.Run("Clear selection when project is selected", func(t *testing.T) {
		// Select a project first
		err := projectManager.SetSelectedProject(context.Background(), project.ID, "test-actor")
		if err != nil {
			t.Fatalf("Failed to select project: %v", err)
		}

		// Create CLI app with clear-selection command
		app := &cli.App{
			Commands: []*cli.Command{
				{
					Name:   "clear-selection",
					Action: clearSelectionAction(appCtx),
				},
			},
		}

		// Test clearing selection
		args := []string{"test", "clear-selection"}
		err = app.Run(args)
		if err != nil {
			t.Errorf("Unexpected error when clearing selection: %v", err)
		}

		// Verify selection is cleared
		hasSelected, err := projectManager.HasSelectedProject(context.Background())
		if err != nil {
			t.Errorf("Failed to check if project is selected: %v", err)
		}
		if hasSelected {
			t.Errorf("Expected no project selected after clearing")
		}
	})

	t.Run("Clear selection when no project is selected", func(t *testing.T) {
		// Ensure no project is selected
		_ = projectManager.ClearSelectedProject(context.Background())

		// Create CLI app with clear-selection command
		app := &cli.App{
			Commands: []*cli.Command{
				{
					Name:   "clear-selection",
					Action: clearSelectionAction(appCtx),
				},
			},
		}

		// Test clearing selection when none is selected
		args := []string{"test", "clear-selection"}
		err := app.Run(args)
		if err != nil {
			t.Errorf("Unexpected error when clearing non-existent selection: %v", err)
		}
	})
}
