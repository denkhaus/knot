package sqlite

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/internal/repository/sqlite/ent"
	"github.com/denkhaus/knot/internal/repository/sqlite/ent/task"
	"github.com/denkhaus/knot/internal/repository/sqlite/ent/taskdependency"
	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
)

// Task Dependency Management Operations

// AddTaskDependency adds a dependency relationship between tasks using ent
func (r *sqliteRepository) AddTaskDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID) (*types.Task, error) {
	var result *types.Task
	err := r.withTx(ctx, func(ctx context.Context, tx *ent.Tx) error {
		// Validate both tasks exist and are in the same project
		task, err := tx.Task.Get(ctx, taskID)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewNotFoundError("task", taskID.String())
			}
			return fmt.Errorf("failed to get task: %w", err)
		}

		dependsOnTask, err := tx.Task.Get(ctx, dependsOnTaskID)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewNotFoundError("depends on task", dependsOnTaskID.String())
			}
			return fmt.Errorf("failed to get depends on task: %w", err)
		}

		if task.ProjectID != dependsOnTask.ProjectID {
			return NewConstraintViolationError("tasks must be in the same project", nil)
		}

		// Check if dependency already exists
		exists, err := tx.TaskDependency.Query().
			Where(
				taskdependency.TaskID(taskID),
				taskdependency.DependsOnTaskID(dependsOnTaskID),
			).
			Exist(ctx)
		if err != nil {
			return fmt.Errorf("failed to check existing dependency: %w", err)
		}
		if exists {
			return NewConstraintViolationError("dependency already exists", nil)
		}

		// Check for circular dependencies
		if err := r.hasCircularDependencyInTx(ctx, tx, taskID, dependsOnTaskID, make(map[uuid.UUID]bool)); err != nil {
			return err
		}

		// Create the dependency
		_, err = tx.TaskDependency.Create().
			SetTaskID(taskID).
			SetDependsOnTaskID(dependsOnTaskID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create task dependency: %w", err)
		}

		// Return updated task
		result, err = r.getTaskInTx(ctx, tx, taskID)
		return err
	})
	return result, err
}

// RemoveTaskDependency removes a dependency relationship between tasks using ent
func (r *sqliteRepository) RemoveTaskDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID) (*types.Task, error) {
	var result *types.Task
	err := r.withTx(ctx, func(ctx context.Context, tx *ent.Tx) error {
		// Validate task exists
		_, err := tx.Task.Get(ctx, taskID)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewNotFoundError("task", taskID.String())
			}
			return fmt.Errorf("failed to get task: %w", err)
		}

		// Delete the dependency
		deletedCount, err := tx.TaskDependency.Delete().
			Where(
				taskdependency.TaskID(taskID),
				taskdependency.DependsOnTaskID(dependsOnTaskID),
			).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete task dependency: %w", err)
		}

		if deletedCount == 0 {
			return NewNotFoundError("task dependency", fmt.Sprintf("%s -> %s", taskID, dependsOnTaskID))
		}

		// Return updated task
		result, err = r.getTaskInTx(ctx, tx, taskID)
		return err
	})
	return result, err
}

// GetTaskDependencies retrieves all tasks that the given task depends on using ent
func (r *sqliteRepository) GetTaskDependencies(ctx context.Context, taskID uuid.UUID) ([]*types.Task, error) {
	// Get dependency task IDs
	dependencyTaskIDs, err := r.client.TaskDependency.Query().
		Where(taskdependency.TaskID(taskID)).
		Select(taskdependency.FieldDependsOnTaskID).
		All(ctx)
	if err != nil {
		return nil, r.mapError("get task dependencies", err)
	}

	if len(dependencyTaskIDs) == 0 {
		return []*types.Task{}, nil
	}

	// Extract IDs
	ids := make([]uuid.UUID, len(dependencyTaskIDs))
	for i, dep := range dependencyTaskIDs {
		ids[i] = dep.DependsOnTaskID
	}

	// Get the actual tasks
	entTasks, err := r.client.Task.Query().
		Where(task.IDIn(ids...)).
		Order(ent.Asc(task.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, r.mapError("get dependency tasks", err)
	}

	tasks := make([]*types.Task, len(entTasks))
	for i, entTask := range entTasks {
		tasks[i] = entTaskToTask(entTask)
	}

	return tasks, nil
}

// GetDependentTasks retrieves all tasks that depend on the given task using ent
func (r *sqliteRepository) GetDependentTasks(ctx context.Context, taskID uuid.UUID) ([]*types.Task, error) {
	// Get dependent task IDs
	dependentTaskIDs, err := r.client.TaskDependency.Query().
		Where(taskdependency.DependsOnTaskID(taskID)).
		Select(taskdependency.FieldTaskID).
		All(ctx)
	if err != nil {
		return nil, r.mapError("get dependent task IDs", err)
	}

	if len(dependentTaskIDs) == 0 {
		return []*types.Task{}, nil
	}

	// Extract IDs
	ids := make([]uuid.UUID, len(dependentTaskIDs))
	for i, dep := range dependentTaskIDs {
		ids[i] = dep.TaskID
	}

	// Get the actual tasks
	entTasks, err := r.client.Task.Query().
		Where(task.IDIn(ids...)).
		Order(ent.Asc(task.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, r.mapError("get dependent tasks", err)
	}

	tasks := make([]*types.Task, len(entTasks))
	for i, entTask := range entTasks {
		tasks[i] = entTaskToTask(entTask)
	}

	return tasks, nil
}

// hasCircularDependencyInTx checks if adding a dependency would create a circular dependency
func (r *sqliteRepository) hasCircularDependencyInTx(ctx context.Context, tx *ent.Tx, fromTaskID, toTaskID uuid.UUID, visited map[uuid.UUID]bool) error {
	// If we've reached the original task, we have a cycle
	if toTaskID == fromTaskID {
		return NewCircularDependencyError(fmt.Sprintf("circular dependency detected: task %s", fromTaskID))
	}

	// If we've already visited this task, skip it to avoid infinite loops
	if visited[toTaskID] {
		return nil
	}

	visited[toTaskID] = true

	// Get all tasks that toTaskID depends on
	dependencies, err := tx.TaskDependency.Query().
		Where(taskdependency.TaskID(toTaskID)).
		Select(taskdependency.FieldDependsOnTaskID).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to get dependencies for circular check: %w", err)
	}

	// Recursively check each dependency
	for _, dep := range dependencies {
		if err := r.hasCircularDependencyInTx(ctx, tx, fromTaskID, dep.DependsOnTaskID, visited); err != nil {
			return err
		}
	}

	return nil
}

// getTaskInTx retrieves a task within a transaction with all its dependencies
func (r *sqliteRepository) getTaskInTx(ctx context.Context, tx *ent.Tx, taskID uuid.UUID) (*types.Task, error) {
	entTask, err := tx.Task.Query().
		Where(task.ID(taskID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, NewNotFoundError("task", taskID.String())
		}
		return nil, fmt.Errorf("failed to get task in transaction: %w", err)
	}

	task := entTaskToTask(entTask)

	// Load dependencies
	dependencies, err := tx.TaskDependency.Query().
		Where(taskdependency.TaskID(taskID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load dependencies in transaction: %w", err)
	}
	task.Dependencies = entTaskDependenciesToTaskIDs(dependencies)

	// Load dependents
	dependents, err := tx.TaskDependency.Query().
		Where(taskdependency.DependsOnTaskID(taskID)).
		Select(taskdependency.FieldTaskID).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load dependents in transaction: %w", err)
	}
	task.Dependents = entTaskDependentsToTaskIDs(dependents)

	return task, nil
}

// NewCircularDependencyError creates a new circular dependency error
// This function is already defined in errors.go, so this is removed to avoid duplication
