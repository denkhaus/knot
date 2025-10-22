package dependency

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/internal/manager"
	"github.com/denkhaus/knot/internal/shared"
	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// chainAction shows the dependency chain for a task
func chainAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		taskIDStr := c.String("task-id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return fmt.Errorf("invalid task ID: %w", err)
		}

		upstream := c.Bool("upstream")
		downstream := c.Bool("downstream")

		// If neither specified, show upstream by default
		if !upstream && !downstream {
			upstream = true
		}

		appCtx.Logger.Info("Showing dependency chain",
			zap.String("taskID", taskID.String()),
			zap.Bool("upstream", upstream),
			zap.Bool("downstream", downstream))

		// Get the original task
		task, err := appCtx.ProjectManager.GetTask(context.Background(), taskID)
		if err != nil {
			appCtx.Logger.Error("Failed to get task", zap.Error(err))
			return fmt.Errorf("failed to get task: %w", err)
		}

		fmt.Printf("Dependency chain for '%s' (ID: %s):\n\n", task.Title, taskID)

		if upstream {
			fmt.Println("📈 UPSTREAM DEPENDENCIES (what this task depends on):")
			if err := showUpstreamChain(appCtx.ProjectManager, taskID, 0); err != nil {
				return fmt.Errorf("failed to show upstream chain: %w", err)
			}
			fmt.Println()
		}

		if downstream {
			fmt.Println("📉 DOWNSTREAM DEPENDENCIES (what depends on this task):")
			if err := showDownstreamChain(appCtx.ProjectManager, taskID, 0); err != nil {
				return fmt.Errorf("failed to show downstream chain: %w", err)
			}
			fmt.Println()
		}

		return nil
	}
}

// showUpstreamChain recursively shows what a task depends on
func showUpstreamChain(projectManager manager.ProjectManager, taskID uuid.UUID, depth int) error {
	dependencies, err := projectManager.GetTaskDependencies(context.Background(), taskID)
	if err != nil {
		return err
	}

	if len(dependencies) == 0 {
		if depth == 0 {
			fmt.Println("  No upstream dependencies")
		}
		return nil
	}

	for _, dep := range dependencies {
		indent := ""
		for i := 0; i < depth; i++ {
			indent += "  "
		}
		fmt.Printf("%s  ├─ %s (ID: %s) - %s\n", indent, dep.Title, dep.ID, dep.State)

		// Recursively show dependencies of this dependency
		if err := showUpstreamChain(projectManager, dep.ID, depth+1); err != nil {
			return err
		}
	}

	return nil
}

// showDownstreamChain recursively shows what depends on a task
func showDownstreamChain(projectManager manager.ProjectManager, taskID uuid.UUID, depth int) error {
	dependents, err := projectManager.GetDependentTasks(context.Background(), taskID)
	if err != nil {
		return err
	}

	if len(dependents) == 0 {
		if depth == 0 {
			fmt.Println("  No downstream dependencies")
		}
		return nil
	}

	for _, dep := range dependents {
		indent := ""
		for i := 0; i < depth; i++ {
			indent += "  "
		}
		fmt.Printf("%s  ├─ %s (ID: %s) - %s\n", indent, dep.Title, dep.ID, dep.State)

		// Recursively show dependents of this dependent
		if err := showDownstreamChain(projectManager, dep.ID, depth+1); err != nil {
			return err
		}
	}

	return nil
}

// cyclesAction detects circular dependencies in a project
func cyclesAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		projectIDStr := c.String("project-id")
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			return fmt.Errorf("invalid project ID: %w", err)
		}

		appCtx.Logger.Info("Detecting dependency cycles", zap.String("projectID", projectID.String()))

		// Get all tasks in the project
		tasks, err := appCtx.ProjectManager.ListTasksForProject(context.Background(), projectID)
		if err != nil {
			appCtx.Logger.Error("Failed to get project tasks", zap.Error(err))
			return fmt.Errorf("failed to get project tasks: %w", err)
		}

		cycles := detectCycles(tasks)

		fmt.Printf("Circular dependency analysis for project %s:\n\n", projectID)

		if len(cycles) == 0 {
			fmt.Println("✅ No circular dependencies detected!")
			return nil
		}

		fmt.Printf("⚠️  Found %d circular dependency cycle(s):\n\n", len(cycles))

		for i, cycle := range cycles {
			fmt.Printf("Cycle %d:\n", i+1)
			for j, taskID := range cycle {
				// Find task details
				var task *types.Task
				for _, t := range tasks {
					if t.ID == taskID {
						task = t
						break
					}
				}

				if task != nil {
					fmt.Printf("  %d. %s (ID: %s)\n", j+1, task.Title, taskID)
				} else {
					fmt.Printf("  %d. Unknown task (ID: %s)\n", j+1, taskID)
				}
			}
			fmt.Printf("  └─ Back to: %s\n\n", cycle[0])
		}

		return nil
	}
}

// detectCycles uses DFS to detect circular dependencies
func detectCycles(tasks []*types.Task) [][]uuid.UUID {
	// Build adjacency list
	graph := make(map[uuid.UUID][]uuid.UUID)
	for _, task := range tasks {
		graph[task.ID] = task.Dependencies
	}

	var cycles [][]uuid.UUID
	visited := make(map[uuid.UUID]bool)
	recStack := make(map[uuid.UUID]bool)
	path := make([]uuid.UUID, 0)

	var dfs func(uuid.UUID) bool
	dfs = func(taskID uuid.UUID) bool {
		visited[taskID] = true
		recStack[taskID] = true
		path = append(path, taskID)

		for _, depID := range graph[taskID] {
			if !visited[depID] {
				if dfs(depID) {
					return true
				}
			} else if recStack[depID] {
				// Found a cycle - extract it from path
				cycleStart := -1
				for i, id := range path {
					if id == depID {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := make([]uuid.UUID, len(path)-cycleStart)
					copy(cycle, path[cycleStart:])
					cycles = append(cycles, cycle)
				}
				return true
			}
		}

		recStack[taskID] = false
		path = path[:len(path)-1]
		return false
	}

	// Check all nodes
	for _, task := range tasks {
		if !visited[task.ID] {
			dfs(task.ID)
		}
	}

	return cycles
}

// validateAction validates all dependencies in a project
func validateAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		projectIDStr := c.String("project-id")
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			return fmt.Errorf("invalid project ID: %w", err)
		}

		appCtx.Logger.Info("Validating dependencies", zap.String("projectID", projectID.String()))

		// Get all tasks in the project
		tasks, err := appCtx.ProjectManager.ListTasksForProject(context.Background(), projectID)
		if err != nil {
			appCtx.Logger.Error("Failed to get project tasks", zap.Error(err))
			return fmt.Errorf("failed to get project tasks: %w", err)
		}

		fmt.Printf("Dependency validation for project %s:\n\n", projectID)

		// Create task map for quick lookup
		taskMap := make(map[uuid.UUID]*types.Task)
		for _, task := range tasks {
			taskMap[task.ID] = task
		}

		var issues []string
		orphanedDeps := 0
		totalDeps := 0

		// Validate each task's dependencies
		for _, task := range tasks {
			for _, depID := range task.Dependencies {
				totalDeps++
				if _, exists := taskMap[depID]; !exists {
					issues = append(issues, fmt.Sprintf("Task '%s' (ID: %s) depends on non-existent task %s",
						task.Title, task.ID, depID))
					orphanedDeps++
				}
			}
		}

		// Check for cycles
		cycles := detectCycles(tasks)

		// Report results
		fmt.Printf("📊 VALIDATION SUMMARY:\n")
		fmt.Printf("  Total tasks: %d\n", len(tasks))
		fmt.Printf("  Total dependencies: %d\n", totalDeps)
		fmt.Printf("  Orphaned dependencies: %d\n", orphanedDeps)
		fmt.Printf("  Circular dependencies: %d\n", len(cycles))
		fmt.Println()

		if len(issues) == 0 && len(cycles) == 0 {
			fmt.Println("✅ All dependencies are valid!")
			return nil
		}

		if len(issues) > 0 {
			fmt.Printf("⚠️  ORPHANED DEPENDENCIES (%d):\n", len(issues))
			for i, issue := range issues {
				fmt.Printf("  %d. %s\n", i+1, issue)
			}
			fmt.Println()
		}

		if len(cycles) > 0 {
			fmt.Printf("⚠️  CIRCULAR DEPENDENCIES (%d cycles detected)\n", len(cycles))
			fmt.Println("  Run 'knot dependency cycles' for detailed cycle information")
			fmt.Println()
		}

		return nil
	}
}
