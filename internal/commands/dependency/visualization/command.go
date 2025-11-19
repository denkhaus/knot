package visualization

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// CommandFactory creates visualization commands
type CommandFactory struct{}

// NewCommandFactory creates a new command factory
func NewCommandFactory() *CommandFactory {
	return &CommandFactory{}
}

// CreateCommand creates the enhanced dependency visualization command
func (f *CommandFactory) CreateCommand(appCtx *shared.AppContext) *cli.Command {
	return &cli.Command{
		Name:  "show",
		Usage: "Show enhanced dependency visualization with arrows and clear relationships",
		Description: `Display dependency relationships with character-based indicators:

  ->  : Direct dependency relationship
  =>  : Blocked by dependency (task waiting for completion)
  [CYCLE] : Circular dependency detected
  [BLOCK] : Task blocked by dependencies
  [READY] : Task ready to work on
  [WORK]  : Task currently in progress
  [DONE]  : Task completed

Examples:
  knot dependency show --task-id <id>          # Show specific task dependencies
  knot dependency show --project             # Show all project dependencies
  knot dependency show --tree                # Show dependency tree
  knot dependency show --graph               # Show dependency graph`,
		Action: f.createAction(appCtx),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "task-id",
				Usage:   "Show dependencies for specific task",
				Aliases: []string{"t"},
			},
			&cli.BoolFlag{
				Name:    "project",
				Usage:   "Show dependency overview for entire project",
				Aliases: []string{"p"},
				Value:   true,
			},
			&cli.BoolFlag{
				Name:    "tree",
				Usage:   "Show dependency tree structure",
				Aliases: []string{"r"},
			},
			&cli.BoolFlag{
				Name:    "graph",
				Usage:   "Show dependency graph with all connections",
				Aliases: []string{"g"},
			},
			&cli.BoolFlag{
				Name:    "blocks",
				Usage:   "Focus on blocking relationships",
				Aliases: []string{"b"},
			},
			shared.NewJSONFlag(),
			&cli.IntFlag{
				Name:    "depth",
				Usage:   "Maximum depth to traverse",
				Aliases: []string{"d"},
				Value:   5,
			},
		},
	}
}

// createAction creates the action function for the command
func (f *CommandFactory) createAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		// Parse configuration
		config, err := f.parseConfig(c, appCtx)
		if err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}

		// Validate configuration
		if err := f.validateConfig(config); err != nil {
			return err
		}

		appCtx.Logger.Info("Enhanced dependency visualization",
			zap.String("mode", string(config.Mode)),
			zap.String("taskID", config.TaskID),
			zap.Bool("json", config.JSONOutput),
			zap.Int("depth", config.MaxDepth))

		// Execute visualization
		return f.executeVisualization(appCtx, config)
	}
}

// parseConfig parses configuration from CLI context
func (f *CommandFactory) parseConfig(c *cli.Context, appCtx *shared.AppContext) (*VisualizationConfig, error) {
	projectID, err := shared.ResolveProjectID(c, appCtx)
	if err != nil {
		return nil, err
	}

	config := &VisualizationConfig{
		TaskID:     c.String("task-id"),
		MaxDepth:   c.Int("depth"),
		ShowBlocks: c.Bool("blocks"),
		JSONOutput: c.Bool("json"),
		ProjectID:  projectID.String(),
	}

	// Determine visualization mode
	if config.TaskID != "" {
		config.Mode = ModeTask
	} else if c.Bool("tree") {
		config.Mode = ModeTree
	} else if c.Bool("graph") {
		config.Mode = ModeGraph
	} else if c.Bool("blocks") {
		config.Mode = ModeBlocks
	} else {
		config.Mode = ModeProject
	}

	return config, nil
}

// validateConfig validates the configuration
func (f *CommandFactory) validateConfig(config *VisualizationConfig) error {
	// Validate task ID if provided
	if config.TaskID != "" {
		if _, err := uuid.Parse(config.TaskID); err != nil {
			return fmt.Errorf("invalid task ID: %w", err)
		}
	}

	// Validate depth
	if config.MaxDepth < 1 || config.MaxDepth > 10 {
		return fmt.Errorf("depth must be between 1 and 10")
	}

	return nil
}

// executeVisualization executes the visualization based on configuration
func (f *CommandFactory) executeVisualization(appCtx *shared.AppContext, config *VisualizationConfig) error {
	// Get project tasks
	projectID, err := uuid.Parse(config.ProjectID)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	tasks, err := appCtx.ProjectManager.ListTasksForProject(context.Background(), projectID)
	if err != nil {
		return fmt.Errorf("failed to get project tasks: %w", err)
	}

	// Create components
	analyzer := NewAnalyzer(appCtx.ProjectManager, tasks)
	renderer := NewRenderer(config)

	// Execute based on mode
	switch config.Mode {
	case ModeTask:
		return f.executeTaskVisualization(analyzer, renderer, config)
	case ModeTree:
		return f.executeTreeVisualization(analyzer, renderer, config)
	case ModeGraph:
		return f.executeGraphVisualization(analyzer, renderer, config)
	case ModeBlocks:
		return f.executeBlocksVisualization(analyzer, renderer, config)
	default:
		return f.executeProjectVisualization(analyzer, renderer, config)
	}
}

// executeTaskVisualization handles task-specific visualization
func (f *CommandFactory) executeTaskVisualization(analyzer *Analyzer, renderer *Renderer, config *VisualizationConfig) error {
	taskID, err := uuid.Parse(config.TaskID)
	if err != nil {
		return fmt.Errorf("invalid task ID: %w", err)
	}

	// Analyze task
	result, err := analyzer.AnalyzeTask(taskID)
	if err != nil {
		return err
	}

	// Render
	if config.JSONOutput {
		return renderer.RenderJSON(result, nil)
	}

	if err := renderer.RenderTaskAnalysis(result); err != nil {
		return err
	}

	return renderer.Render()
}

// executeTreeVisualization handles tree visualization
func (f *CommandFactory) executeTreeVisualization(analyzer *Analyzer, renderer *Renderer, config *VisualizationConfig) error {
	result, err := analyzer.AnalyzeProject()
	if err != nil {
		return err
	}

	if config.JSONOutput {
		return renderer.RenderJSON(nil, result)
	}

	if err := renderer.RenderTree(result); err != nil {
		return err
	}

	return renderer.Render()
}

// executeGraphVisualization handles graph visualization
func (f *CommandFactory) executeGraphVisualization(analyzer *Analyzer, renderer *Renderer, config *VisualizationConfig) error {
	result, err := analyzer.AnalyzeProject()
	if err != nil {
		return err
	}

	if config.JSONOutput {
		return renderer.RenderJSON(nil, result)
	}

	if err := renderer.RenderGraph(result); err != nil {
		return err
	}

	return renderer.Render()
}

// executeBlocksVisualization handles blocking visualization
func (f *CommandFactory) executeBlocksVisualization(analyzer *Analyzer, renderer *Renderer, config *VisualizationConfig) error {
	result, err := analyzer.AnalyzeProject()
	if err != nil {
		return err
	}

	if config.JSONOutput {
		return renderer.RenderJSON(nil, result)
	}

	// Focus on blocked tasks
	if err := renderer.renderBlockedTasks(result); err != nil {
		return err
	}

	return renderer.Render()
}

// executeProjectVisualization handles project overview visualization
func (f *CommandFactory) executeProjectVisualization(analyzer *Analyzer, renderer *Renderer, config *VisualizationConfig) error {
	result, err := analyzer.AnalyzeProject()
	if err != nil {
		return err
	}

	if config.JSONOutput {
		return renderer.RenderJSON(nil, result)
	}

	if err := renderer.RenderProjectOverview(result); err != nil {
		return err
	}

	return renderer.Render()
}
