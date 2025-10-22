package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/denkhaus/knot/internal/repository/sqlite/ent"
	"github.com/denkhaus/knot/internal/repository/sqlite/ent/project"
	taskpred "github.com/denkhaus/knot/internal/repository/sqlite/ent/task"
	"github.com/denkhaus/knot/internal/repository/sqlite/ent/taskdependency"
	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
)

// Task CRUD Operations

// CreateTask creates a new task using ent with dependency handling
func (r *sqliteRepository) CreateTask(ctx context.Context, task *types.Task) error {
	return r.withTx(ctx, func(ctx context.Context, tx *ent.Tx) error {
		// Validate project exists
		exists, err := tx.Project.Query().Where(project.ID(task.ProjectID)).Exist(ctx)
		if err != nil {
			return fmt.Errorf("failed to check project existence: %w", err)
		}
		if !exists {
			return NewNotFoundError("project", task.ProjectID.String())
		}

		// Validate parent task if specified
		if task.ParentID != nil {
			parentTask, err := tx.Task.Get(ctx, *task.ParentID)
			if err != nil {
				if ent.IsNotFound(err) {
					return NewNotFoundError("parent task", task.ParentID.String())
				}
				return fmt.Errorf("failed to get parent task: %w", err)
			}
			if parentTask.ProjectID != task.ProjectID {
				return NewConstraintViolationError("parent task must be in the same project", nil)
			}
			// Set correct depth
			task.Depth = parentTask.Depth + 1
		} else {
			task.Depth = 0
		}

		// Set timestamps if not already set
		if task.CreatedAt.IsZero() {
			task.CreatedAt = time.Now()
		}
		if task.UpdatedAt.IsZero() {
			task.UpdatedAt = task.CreatedAt
		}

		// Create the task
		_, err = taskToEntTaskCreate(task, tx.Client()).Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create task: %w", err)
		}

		// Create task dependencies if any
		if len(task.Dependencies) > 0 {
			dependencies := make([]TaskDependencyPair, len(task.Dependencies))
			for i, depID := range task.Dependencies {
				// Validate dependency exists and is in the same project
				depTask, err := tx.Task.Get(ctx, depID)
				if err != nil {
					if ent.IsNotFound(err) {
						return NewNotFoundError("dependency task", depID.String())
					}
					return fmt.Errorf("failed to get dependency task: %w", err)
				}
				if depTask.ProjectID != task.ProjectID {
					return NewConstraintViolationError("dependency task must be in the same project", nil)
				}

				dependencies[i] = TaskDependencyPair{
					TaskID:          task.ID,
					DependsOnTaskID: depID,
				}
			}

			// Check for circular dependencies before creating
			for _, dep := range dependencies {
				if err := r.hasCircularDependencyInTx(ctx, tx, dep.TaskID, dep.DependsOnTaskID, make(map[uuid.UUID]bool)); err != nil {
					return err
				}
			}

			// Create dependencies
			bulk := make([]*ent.TaskDependencyCreate, len(dependencies))
			for i, dep := range dependencies {
				bulk[i] = tx.TaskDependency.Create().
					SetTaskID(dep.TaskID).
					SetDependsOnTaskID(dep.DependsOnTaskID)
			}
			if _, err := tx.TaskDependency.CreateBulk(bulk...).Save(ctx); err != nil {
				return fmt.Errorf("failed to create task dependencies: %w", err)
			}
		}

		// Update project metrics
		return r.updateProjectMetricsInTx(ctx, tx, task.ProjectID)
	})
}

// GetTask retrieves a task by ID with dependencies using ent
func (r *sqliteRepository) GetTask(ctx context.Context, id uuid.UUID) (*types.Task, error) {
	entTask, err := r.client.Task.Query().
		Where(taskpred.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, NewNotFoundError("task", id.String())
		}
		return nil, r.mapError("get task", err)
	}

	domainTask := entTaskToTask(entTask)

	// Load dependencies
	dependencies, err := r.client.TaskDependency.Query().
		Where(taskdependency.TaskID(id)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load dependencies: %w", err)
	}
	domainTask.Dependencies = entTaskDependenciesToTaskIDs(dependencies)

	// Load dependents (tasks that depend on this task)
	dependents, err := r.client.TaskDependency.Query().
		Where(taskdependency.DependsOnTaskID(id)).
		Select(taskdependency.FieldTaskID).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load dependents: %w", err)
	}
	domainTask.Dependents = entTaskDependentsToTaskIDs(dependents)

	return domainTask, nil
}

// UpdateTask updates an existing task using ent
func (r *sqliteRepository) UpdateTask(ctx context.Context, task *types.Task) error {
	return r.withTx(ctx, func(ctx context.Context, tx *ent.Tx) error {
		// Get existing task to preserve certain fields
		existingTask, err := tx.Task.Get(ctx, task.ID)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewNotFoundError("task", task.ID.String())
			}
			return fmt.Errorf("failed to get existing task: %w", err)
		}

		// Preserve immutable fields
		task.CreatedAt = existingTask.CreatedAt
		task.ProjectID = existingTask.ProjectID
		task.ParentID = existingTask.ParentID
		task.Depth = existingTask.Depth
		task.UpdatedAt = time.Now()

		// Handle completion timestamp
		if task.State == types.TaskStateCompleted && string(existingTask.State) != string(types.TaskStateCompleted) {
			now := time.Now()
			task.CompletedAt = &now
		} else if task.State != types.TaskStateCompleted {
			task.CompletedAt = nil
		}

		// Update the task
		update := tx.Task.UpdateOneID(task.ID)
		taskToEntTaskUpdate(task, update)
		err = update.Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}

		// Update project metrics if state changed
		if string(existingTask.State) != string(task.State) {
			return r.updateProjectMetricsInTx(ctx, tx, task.ProjectID)
		}

		return nil
	})
}

// DeleteTask deletes a task if it has no children using ent
func (r *sqliteRepository) DeleteTask(ctx context.Context, id uuid.UUID) error {
	return r.withTx(ctx, func(ctx context.Context, tx *ent.Tx) error {
		// Get task info
		task, err := tx.Task.Get(ctx, id)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewNotFoundError("task", id.String())
			}
			return fmt.Errorf("failed to get task: %w", err)
		}

		// Check if task has children
		childrenCount, err := tx.Task.Query().
			Where(taskpred.ParentIDEQ(id)).
			Count(ctx)
		if err != nil {
			return fmt.Errorf("failed to count children: %w", err)
		}
		if childrenCount > 0 {
			return NewConstraintViolationError("cannot delete task with children", nil)
		}

		// Delete all task dependencies (both incoming and outgoing)
		_, err = tx.TaskDependency.Delete().
			Where(taskdependency.Or(
				taskdependency.TaskID(id),
				taskdependency.DependsOnTaskID(id),
			)).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete task dependencies: %w", err)
		}

		// Delete the task
		err = tx.Task.DeleteOneID(id).Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete task: %w", err)
		}

		// Update project metrics
		return r.updateProjectMetricsInTx(ctx, tx, task.ProjectID)
	})
}

// DeleteTaskSubtree deletes a task and all its descendants using ent with recursive CTE
func (r *sqliteRepository) DeleteTaskSubtree(ctx context.Context, taskID uuid.UUID) error {
	return r.withTx(ctx, func(ctx context.Context, tx *ent.Tx) error {
		// Get the task to ensure it exists and get project ID
		task, err := tx.Task.Get(ctx, taskID)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewNotFoundError("task", taskID.String())
			}
			return fmt.Errorf("failed to get task: %w", err)
		}

		// Get all descendant task IDs using recursive query
		descendantIDs, err := r.getDescendantTaskIDsInTx(ctx, tx, taskID)
		if err != nil {
			return fmt.Errorf("failed to get descendant task IDs: %w", err)
		}

		// Add the root task to the list
		allTaskIDs := append(descendantIDs, taskID)

		// Delete all task dependencies for these tasks
		_, err = tx.TaskDependency.Delete().
			Where(taskdependency.Or(
				taskdependency.TaskIDIn(allTaskIDs...),
				taskdependency.DependsOnTaskIDIn(allTaskIDs...),
			)).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete task dependencies: %w", err)
		}

		// Delete all tasks in the subtree (children first due to foreign key constraints)
		_, err = tx.Task.Delete().
			Where(taskpred.IDIn(allTaskIDs...)).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete tasks: %w", err)
		}

		// Update project metrics
		return r.updateProjectMetricsInTx(ctx, tx, task.ProjectID)
	})
}

// getDescendantTaskIDsInTx gets all descendant task IDs using recursive approach
func (r *sqliteRepository) getDescendantTaskIDsInTx(ctx context.Context, tx *ent.Tx, taskID uuid.UUID) ([]uuid.UUID, error) {
	var allDescendants []uuid.UUID
	queue := []uuid.UUID{taskID}
	visited := make(map[uuid.UUID]bool)

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if visited[currentID] {
			continue
		}
		visited[currentID] = true

		// Get direct children
		children, err := tx.Task.Query().
			Where(taskpred.ParentIDEQ(currentID)).
			IDs(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get children for task %s: %w", currentID, err)
		}

		for _, childID := range children {
			if !visited[childID] {
				allDescendants = append(allDescendants, childID)
				queue = append(queue, childID)
			}
		}
	}

	return allDescendants, nil
}
