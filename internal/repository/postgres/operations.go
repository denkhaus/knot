package postgres

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/repository/ent"
	entproject "github.com/denkhaus/knot/v2/internal/repository/ent/project"
	enttask "github.com/denkhaus/knot/v2/internal/repository/ent/task"
	"github.com/denkhaus/knot/v2/internal/repository/ent/taskdependency"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateProject creates a new project
func (r *postgresRepository) CreateProject(ctx context.Context, project *types.Project) error {
	r.logger.Debug("Creating project", zap.String("title", project.Title))

	result, err := projectToEntProjectCreate(project, r.client).Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create project", zap.Error(err))
		return fmt.Errorf("failed to create project: %w", err)
	}

	// Update the project with the generated ID and timestamps
	project.ID = result.ID
	project.CreatedAt = result.CreatedAt
	project.UpdatedAt = result.UpdatedAt

	r.logger.Info("Project created successfully",
		zap.String("project_id", project.ID.String()),
		zap.String("title", project.Title))

	return nil
}

// GetProject retrieves a project by ID
func (r *postgresRepository) GetProject(ctx context.Context, id uuid.UUID) (*types.Project, error) {
	entProject, err := r.client.Project.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("project not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return entProjectToProject(entProject), nil
}

// UpdateProject updates an existing project
func (r *postgresRepository) UpdateProject(ctx context.Context, project *types.Project) error {
	r.logger.Debug("Updating project", zap.String("project_id", project.ID.String()))

	_, err := r.client.Project.UpdateOneID(project.ID).
		SetTitle(project.Title).
		SetDescription(project.Description).
		SetState(entproject.State(project.State)).
		SetProgress(project.Progress).
		Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update project", zap.Error(err))
		return fmt.Errorf("failed to update project: %w", err)
	}

	r.logger.Info("Project updated successfully",
		zap.String("project_id", project.ID.String()),
		zap.String("title", project.Title))

	return nil
}

// DeleteProject removes a project and all its tasks
func (r *postgresRepository) DeleteProject(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug("Deleting project", zap.String("project_id", id.String()))

	err := r.client.Project.DeleteOneID(id).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete project", zap.Error(err))
		return fmt.Errorf("failed to delete project: %w", err)
	}

	r.logger.Info("Project deleted successfully", zap.String("project_id", id.String()))
	return nil
}

// ListProjects returns all projects
func (r *postgresRepository) ListProjects(ctx context.Context) ([]*types.Project, error) {
	projects, err := r.client.Project.Query().
		Order(ent.Desc(entproject.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	result := make([]*types.Project, len(projects))
	for i, p := range projects {
		result[i] = entProjectToProject(p)
	}

	return result, nil
}

// CreateTask creates a new task
func (r *postgresRepository) CreateTask(ctx context.Context, task *types.Task) error {
	r.logger.Debug("Creating task", zap.String("title", task.Title))

	creator := r.client.Task.Create().
		SetProjectID(task.ProjectID).
		SetTitle(task.Title).
		SetDescription(task.Description).
		SetState(enttask.State(task.State)).
		SetPriority(taskPriorityToEntPriority(task.Priority)).
		SetComplexity(task.Complexity).
		SetDepth(task.Depth)

	if task.ParentID != nil {
		creator = creator.SetParentID(*task.ParentID)
	}

	result, err := creator.Save(ctx)
	if err != nil {
		r.logger.Error("Failed to create task", zap.Error(err))
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Update the task with the generated ID
	task.ID = result.ID
	task.CreatedAt = result.CreatedAt
	task.UpdatedAt = result.UpdatedAt

	r.logger.Info("Task created successfully",
		zap.String("task_id", task.ID.String()),
		zap.String("title", task.Title))

	return nil
}

// GetTask retrieves a task by ID
func (r *postgresRepository) GetTask(ctx context.Context, id uuid.UUID) (*types.Task, error) {
	entTask, err := r.client.Task.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("task not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return entTaskToTask(entTask), nil
}

// GetTasksWithDependencies retrieves tasks with their dependencies
func (r *postgresRepository) GetTasksWithDependencies(ctx context.Context, taskIDs []uuid.UUID) ([]*types.Task, error) {
	if len(taskIDs) == 0 {
		return []*types.Task{}, nil
	}

	var tasks []*types.Task

	// Get all tasks first
	entTasks, err := r.client.Task.Query().
		Where(enttask.IDIn(taskIDs...)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	// Get dependencies for each task
	for _, t := range entTasks {
		task := entTaskToTask(t)

		// Get dependencies for this task
		dependencies, err := r.client.TaskDependency.Query().
			Where(taskdependency.TaskID(t.ID)).
			All(ctx)
		if err != nil {
			r.logger.Warn("Failed to get dependencies for task",
				zap.String("task_id", t.ID.String()),
				zap.Error(err))
			// Continue without dependencies if there's an error
			tasks = append(tasks, task)
			continue
		}

		// Convert dependencies
		if len(dependencies) > 0 {
			task.Dependencies = make([]uuid.UUID, len(dependencies))
			for j, dep := range dependencies {
				task.Dependencies[j] = dep.DependsOnTaskID
			}
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTask updates an existing task
func (r *postgresRepository) UpdateTask(ctx context.Context, task *types.Task) error {
	r.logger.Debug("Updating task", zap.String("task_id", task.ID.String()))

	_, err := r.client.Task.UpdateOneID(task.ID).
		SetTitle(task.Title).
		SetDescription(task.Description).
		SetState(enttask.State(task.State)).
		SetPriority(taskPriorityToEntPriority(task.Priority)).
		SetComplexity(task.Complexity).
		SetDepth(task.Depth).
		Save(ctx)
	if err != nil {
		r.logger.Error("Failed to update task", zap.Error(err))
		return fmt.Errorf("failed to update task: %w", err)
	}

	r.logger.Info("Task updated successfully",
		zap.String("task_id", task.ID.String()),
		zap.String("title", task.Title))

	return nil
}

// DeleteTask removes a task
func (r *postgresRepository) DeleteTask(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug("Deleting task", zap.String("task_id", id.String()))

	err := r.client.Task.DeleteOneID(id).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete task", zap.Error(err))
		return fmt.Errorf("failed to delete task: %w", err)
	}

	r.logger.Info("Task deleted successfully", zap.String("task_id", id.String()))
	return nil
}

// ListTasks returns tasks filtered by criteria
func (r *postgresRepository) ListTasks(ctx context.Context, filter types.TaskFilter) ([]*types.Task, error) {
	query := r.client.Task.Query()

	if filter.ProjectID != nil {
		query = query.Where(enttask.ProjectID(*filter.ProjectID))
	}

	if filter.ParentID != nil {
		if *filter.ParentID == uuid.Nil {
			query = query.Where(enttask.ParentIDIsNil())
		} else {
			query = query.Where(enttask.ParentID(*filter.ParentID))
		}
	}

	if filter.State != nil {
		query = query.Where(enttask.StateEQ(enttask.State(*filter.State)))
	}

	if filter.Priority != nil {
		query = query.Where(enttask.PriorityEQ(taskPriorityToEntPriority(*filter.Priority)))
	}

	if filter.MinDepth != nil {
		query = query.Where(enttask.DepthGTE(*filter.MinDepth))
	}

	if filter.MaxDepth != nil {
		query = query.Where(enttask.DepthLTE(*filter.MaxDepth))
	}

	tasks, err := query.Order(ent.Desc(enttask.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	result := make([]*types.Task, len(tasks))
	for i, t := range tasks {
		result[i] = entTaskToTask(t)
	}

	return result, nil
}

// GetTasksByProject returns all tasks for a project
func (r *postgresRepository) GetTasksByProject(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error) {
	tasks, err := r.client.Task.Query().
		Where(enttask.ProjectID(projectID)).
		Order(ent.Asc(enttask.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by project: %w", err)
	}

	result := make([]*types.Task, len(tasks))
	for i, t := range tasks {
		result[i] = entTaskToTask(t)
	}

	return result, nil
}

// GetTasksByParent returns all child tasks of a parent task
func (r *postgresRepository) GetTasksByParent(ctx context.Context, parentID uuid.UUID) ([]*types.Task, error) {
	tasks, err := r.client.Task.Query().
		Where(enttask.ParentID(parentID)).
		Order(ent.Asc(enttask.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by parent: %w", err)
	}

	result := make([]*types.Task, len(tasks))
	for i, t := range tasks {
		result[i] = entTaskToTask(t)
	}

	return result, nil
}

// GetRootTasks returns all root tasks (tasks with no parent) for a project
func (r *postgresRepository) GetRootTasks(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error) {
	tasks, err := r.client.Task.Query().
		Where(
			enttask.And(
				enttask.ProjectID(projectID),
				enttask.ParentIDIsNil(),
			),
		).
		Order(ent.Asc(enttask.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get root tasks: %w", err)
	}

	result := make([]*types.Task, len(tasks))
	for i, t := range tasks {
		result[i] = entTaskToTask(t)
	}

	return result, nil
}

// GetParentTask returns the parent task of a given task
func (r *postgresRepository) GetParentTask(ctx context.Context, taskID uuid.UUID) (*types.Task, error) {
	task, err := r.client.Task.Query().
		Where(enttask.ID(taskID)).
		WithParent().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent task: %w", err)
	}

	if task.Edges.Parent == nil {
		return nil, fmt.Errorf("task has no parent: %s", taskID)
	}

	return entTaskToTask(task.Edges.Parent), nil
}

// DeleteTaskSubtree deletes a task and all its descendants
func (r *postgresRepository) DeleteTaskSubtree(ctx context.Context, taskID uuid.UUID) error {
	r.logger.Debug("Deleting task subtree", zap.String("root_task_id", taskID.String()))

	// This is a complex operation that would need to be implemented recursively
	// For now, we'll delete the task and let the database handle cascading deletes if configured
	err := r.client.Task.DeleteOneID(taskID).Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to delete task subtree", zap.Error(err))
		return fmt.Errorf("failed to delete task subtree: %w", err)
	}

	r.logger.Info("Task subtree deleted successfully", zap.String("root_task_id", taskID.String()))
	return nil
}

// AddTaskDependency adds a dependency relationship between tasks
func (r *postgresRepository) AddTaskDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID) (*types.Task, error) {
	r.logger.Debug("Adding task dependency",
		zap.String("task_id", taskID.String()),
		zap.String("depends_on", dependsOnTaskID.String()))

	_, err := r.client.TaskDependency.Create().
		SetTaskID(taskID).
		SetDependsOnTaskID(dependsOnTaskID).
		Save(ctx)
	if err != nil {
		r.logger.Error("Failed to add task dependency", zap.Error(err))
		return nil, fmt.Errorf("failed to add task dependency: %w", err)
	}

	// Return the updated task
	return r.GetTask(ctx, taskID)
}

// RemoveTaskDependency removes a dependency relationship between tasks
func (r *postgresRepository) RemoveTaskDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID) (*types.Task, error) {
	r.logger.Debug("Removing task dependency",
		zap.String("task_id", taskID.String()),
		zap.String("depends_on", dependsOnTaskID.String()))

	_, err := r.client.TaskDependency.Delete().
		Where(
			taskdependency.And(
				taskdependency.TaskID(taskID),
				taskdependency.DependsOnTaskID(dependsOnTaskID),
			),
		).
		Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to remove task dependency", zap.Error(err))
		return nil, fmt.Errorf("failed to remove task dependency: %w", err)
	}

	// Return the updated task
	return r.GetTask(ctx, taskID)
}

// GetTaskDependencies returns all tasks that the given task depends on
func (r *postgresRepository) GetTaskDependencies(ctx context.Context, taskID uuid.UUID) ([]*types.Task, error) {
	dependencies, err := r.client.TaskDependency.Query().
		Where(taskdependency.TaskID(taskID)).
		WithDependsOnTask().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task dependencies: %w", err)
	}

	result := make([]*types.Task, len(dependencies))
	for i, dep := range dependencies {
		if dep.Edges.DependsOnTask != nil {
			result[i] = entTaskToTask(dep.Edges.DependsOnTask)
		}
	}

	return result, nil
}

// GetDependentTasks returns all tasks that depend on the given task
func (r *postgresRepository) GetDependentTasks(ctx context.Context, taskID uuid.UUID) ([]*types.Task, error) {
	dependents, err := r.client.TaskDependency.Query().
		Where(taskdependency.DependsOnTaskID(taskID)).
		WithTask().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependent tasks: %w", err)
	}

	result := make([]*types.Task, len(dependents))
	for i, dep := range dependents {
		if dep.Edges.Task != nil {
			result[i] = entTaskToTask(dep.Edges.Task)
		}
	}

	return result, nil
}

// GetProjectProgress calculates the progress of a project
func (r *postgresRepository) GetProjectProgress(ctx context.Context, projectID uuid.UUID) (*types.ProjectProgress, error) {
	// Get total and completed tasks counts
	total, err := r.client.Task.Query().
		Where(enttask.ProjectID(projectID)).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total task count: %w", err)
	}

	completed, err := r.client.Task.Query().
		Where(
			enttask.And(
				enttask.ProjectID(projectID),
				enttask.StateEQ(enttask.State(types.TaskStateCompleted)),
			),
		).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed task count: %w", err)
	}

	progress := &types.ProjectProgress{
		ProjectID:       projectID,
		TotalTasks:      int(total),
		CompletedTasks:  int(completed),
		OverallProgress: float64(completed) / float64(total) * 100,
	}

	return progress, nil
}

// GetTaskCountByDepth returns the count of tasks at each depth level
func (r *postgresRepository) GetTaskCountByDepth(ctx context.Context, projectID uuid.UUID, maxDepth int) (map[int]int, error) {
	tasks, err := r.client.Task.Query().
		Where(
			enttask.And(
				enttask.ProjectID(projectID),
				enttask.DepthLTE(maxDepth),
			),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by depth: %w", err)
	}

	result := make(map[int]int)
	for _, task := range tasks {
		result[task.Depth]++
	}

	return result, nil
}

