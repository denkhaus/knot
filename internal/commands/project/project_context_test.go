package project

import (
	"context"
	"flag"
	"testing"

	"github.com/denkhaus/knot/v2/internal/di"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
	"github.com/urfave/cli/v2"
)

func TestProjectSelectCommand(t *testing.T) {
	// Setup test environment with DI
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

	// Get project manager from DI
	projectManager := do.MustInvoke[manager.ProjectManager](diContainer.GetInjector())

	// Create a test project
	project, err := projectManager.CreateProject(context.Background(), "Test Project", "Test Description", "test-actor")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	t.Run("Select existing project", func(t *testing.T) {
		// Create CLI app with select command using the same container
		selectApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
			Commands: []*cli.Command{
				{
					Name:   "select",
					Action: selectAction(),
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
		err := selectApp.Run(args)
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
		// Create CLI app with select command using the same container
		selectApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
			Commands: []*cli.Command{
				{
					Name:   "select",
					Action: selectAction(),
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
		err := selectApp.Run(args)
		if err == nil {
			t.Errorf("Expected error when selecting non-existent project")
		}
	})

	t.Run("Select with invalid UUID", func(t *testing.T) {
		// Create CLI app with select command using the same container
		selectApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
			Commands: []*cli.Command{
				{
					Name:   "select",
					Action: selectAction(),
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
		err := selectApp.Run(args)
		if err == nil {
			t.Errorf("Expected error when selecting with invalid UUID")
		}
	})
}

func TestProjectGetSelectedCommand(t *testing.T) {
	// Setup test environment with DI - reuse container from previous test
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

	// Get project manager from DI
	projectManager := do.MustInvoke[manager.ProjectManager](diContainer.GetInjector())

	// Create a test project
	project, err := projectManager.CreateProject(context.Background(), "Test Project", "Test Description", "test-actor")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	t.Run("Get selected when none selected", func(t *testing.T) {
		// Ensure no project is selected
		_ = projectManager.ClearSelectedProject(context.Background())

		// Create CLI app with get-selected command using the same container
		getApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
			Commands: []*cli.Command{
				{
					Name:   "get-selected",
					Action: getSelectedAction(),
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
		err := getApp.Run(args)
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

		// Create CLI app with get-selected command using the same container
		getApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
			Commands: []*cli.Command{
				{
					Name:   "get-selected",
					Action: getSelectedAction(),
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
		err = getApp.Run(args)
		if err != nil {
			t.Errorf("Unexpected error when getting selected project: %v", err)
		}
	})
}

func TestProjectClearSelectionCommand(t *testing.T) {
	// Setup test environment with DI - reuse container from previous test
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

	// Get project manager from DI
	projectManager := do.MustInvoke[manager.ProjectManager](diContainer.GetInjector())

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

		// Create CLI app with clear-selection command using the same container
		clearApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
			Commands: []*cli.Command{
				{
					Name:   "clear-selection",
					Action: clearSelectionAction(),
				},
			},
		}

		// Test clearing selection
		args := []string{"test", "clear-selection"}
		err = clearApp.Run(args)
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

		// Create CLI app with clear-selection command using the same container
		clearApp := &cli.App{
			Metadata: map[string]interface{}{
				"container": diContainer,
			},
			Commands: []*cli.Command{
				{
					Name:   "clear-selection",
					Action: clearSelectionAction(),
				},
			},
		}

		// Test clearing selection when none is selected
		args := []string{"test", "clear-selection"}
		err := clearApp.Run(args)
		if err != nil {
			t.Errorf("Unexpected error when clearing non-existent selection: %v", err)
		}
	})
}