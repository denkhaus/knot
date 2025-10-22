package sqlite

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/internal/repository/sqlite/ent"
	"github.com/denkhaus/knot/internal/repository/sqlite/ent/project"
	task "github.com/denkhaus/knot/internal/repository/sqlite/ent/task"
	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
)

// Task Query Operations

// ListTasks retrieves tasks with filtering using ent
func (r *sqliteRepository) ListTasks(ctx context.Context, filter types.TaskFilter) ([]*types.Task, error) {
	query := r.client.Task.Query()

	// Apply filters using ent predicates
	if filter.ProjectID != nil {
		query = query.Where(task.ProjectID(*filter.ProjectID))
	}
	if filter.ParentID != nil {
		if *filter.ParentID == uuid.Nil {
			query = query.Where(task.ParentIDIsNil())
		} else {
			query = query.Where(task.ParentID(*filter.ParentID))
		}
	}
	if filter.State != nil {
		query = query.Where(task.StateEQ(task.State(string(*filter.State))))
	}
	if filter.MinDepth != nil {
		query = query.Where(task.DepthGTE(*filter.MinDepth))
	}
	if filter.MaxDepth != nil {
		query = query.Where(task.DepthLTE(*filter.MaxDepth))
	}
	if filter.MinComplexity != nil {
		query = query.Where(task.ComplexityGTE(*filter.MinComplexity))
	}
	if filter.MaxComplexity != nil {
		query = query.Where(task.ComplexityLTE(*filter.MaxComplexity))
	}

	// Execute query
	entTasks, err := query.
		Order(ent.Asc(task.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, r.mapError("list tasks", err)
	}

	// Convert to domain models
	tasks := make([]*types.Task, len(entTasks))
	for i, entTask := range entTasks {
		tasks[i] = entTaskToTask(entTask)
	}

	return tasks, nil
}

// GetTasksByProject retrieves all tasks for a specific project using ent
func (r *sqliteRepository) GetTasksByProject(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error) {
	entTasks, err := r.client.Task.Query().
		Where(task.ProjectID(projectID)).
		Order(ent.Asc(task.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, r.mapError("get tasks by project", err)
	}

	// Convert to domain models
	tasks := make([]*types.Task, len(entTasks))
	for i, entTask := range entTasks {
		tasks[i] = entTaskToTask(entTask)
	}

	return tasks, nil
}

// GetTasksByParent retrieves all direct children of a parent task using ent
func (r *sqliteRepository) GetTasksByParent(ctx context.Context, parentID uuid.UUID) ([]*types.Task, error) {
	entTasks, err := r.client.Task.Query().
		Where(task.ParentID(parentID)).
		Order(ent.Asc(task.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, r.mapError("get tasks by parent", err)
	}

	tasks := make([]*types.Task, len(entTasks))
	for i, entTask := range entTasks {
		tasks[i] = entTaskToTask(entTask)
	}

	return tasks, nil
}

// GetRootTasks retrieves all root tasks (tasks without parents) for a project using ent
func (r *sqliteRepository) GetRootTasks(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error) {
	entTasks, err := r.client.Task.Query().
		Where(
			task.ProjectID(projectID),
			task.ParentIDIsNil(),
		).
		Order(ent.Asc(task.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, r.mapError("get root tasks", err)
	}

	tasks := make([]*types.Task, len(entTasks))
	for i, entTask := range entTasks {
		tasks[i] = entTaskToTask(entTask)
	}

	return tasks, nil
}

// GetParentTask retrieves the parent task of a given task using ent
func (r *sqliteRepository) GetParentTask(ctx context.Context, taskID uuid.UUID) (*types.Task, error) {
	// Get the task first to get parent ID
	task, err := r.client.Task.Get(ctx, taskID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, NewNotFoundError("task", taskID.String())
		}
		return nil, r.mapError("get task for parent lookup", err)
	}

	if task.ParentID == nil {
		return nil, NewNotFoundError("parent task", "nil")
	}

	// Get the parent taskpred
	return r.GetTask(ctx, *task.ParentID)
}

// GetProjectProgress calculates project progress using ent aggregations
func (r *sqliteRepository) GetProjectProgress(ctx context.Context, projectID uuid.UUID) (*types.ProjectProgress, error) {
	// Verify project exists
	exists, err := r.client.Project.Query().Where(project.ID(projectID)).Exist(ctx)
	if err != nil {
		return nil, r.mapError("check project existence", err)
	}
	if !exists {
		return nil, NewNotFoundError("project", projectID.String())
	}

	// Get task counts by state
	taskCounts := make(map[types.TaskState]int)

	states := []types.TaskState{
		types.TaskStatePending,
		types.TaskStateInProgress,
		types.TaskStateCompleted,
		types.TaskStateBlocked,
		types.TaskStateCancelled,
	}

	totalTasks := 0
	for _, state := range states {
		count, err := r.client.Task.Query().
			Where(
				task.ProjectID(projectID),
				task.StateEQ(task.State(string(state))),
			).
			Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count tasks with state %s: %w", state, err)
		}
		taskCounts[state] = count
		totalTasks += count
	}

	// Get task counts by depth
	tasksByDepth := make(map[int]int)
	if totalTasks > 0 {
		// Get max depth first
		maxDepthResult, err := r.client.Task.Query().
			Where(task.ProjectID(projectID)).
			Aggregate(ent.Max(task.FieldDepth)).
			Int(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get max depth: %w", err)
		}

		// Count tasks for each depth level
		for depth := 0; depth <= maxDepthResult; depth++ {
			count, err := r.client.Task.Query().
				Where(
					task.ProjectID(projectID),
					task.DepthEQ(depth),
				).
				Count(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to count tasks at depth %d: %w", depth, err)
			}
			if count > 0 {
				tasksByDepth[depth] = count
			}
		}
	}

	// Calculate overall progress
	progress := 0.0
	if totalTasks > 0 {
		progress = float64(taskCounts[types.TaskStateCompleted]) / float64(totalTasks) * 100.0
	}

	return &types.ProjectProgress{
		ProjectID:       projectID,
		TotalTasks:      totalTasks,
		CompletedTasks:  taskCounts[types.TaskStateCompleted],
		InProgressTasks: taskCounts[types.TaskStateInProgress],
		PendingTasks:    taskCounts[types.TaskStatePending],
		BlockedTasks:    taskCounts[types.TaskStateBlocked],
		CancelledTasks:  taskCounts[types.TaskStateCancelled],
		OverallProgress: progress,
		TasksByDepth:    tasksByDepth,
	}, nil
}

// GetTaskCountByDepth returns task counts by depth level for a project using ent
func (r *sqliteRepository) GetTaskCountByDepth(ctx context.Context, projectID uuid.UUID, maxDepth int) (map[int]int, error) {
	tasksByDepth := make(map[int]int)

	for depth := 0; depth <= maxDepth; depth++ {
		count, err := r.client.Task.Query().
			Where(
				task.ProjectID(projectID),
				task.DepthEQ(depth),
			).
			Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to count tasks at depth %d: %w", depth, err)
		}
		tasksByDepth[depth] = count
	}

	return tasksByDepth, nil
}
