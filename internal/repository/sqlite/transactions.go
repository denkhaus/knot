package sqlite

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/repository/sqlite/ent"
	"github.com/denkhaus/knot/v2/internal/repository/sqlite/ent/task"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
)

// TxFunc represents a function that executes within an ent transaction
type TxFunc func(ctx context.Context, tx *ent.Tx) error

// withTx executes a function within an ent transaction
func (r *sqliteRepository) withTx(ctx context.Context, fn TxFunc) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return NewTransactionError("failed to begin transaction", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic and re-panic
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = fmt.Errorf("rollback failed after panic: %v (original panic: %v)", rollbackErr, p)
			}
			panic(p)
		} else if err != nil {
			// Rollback on error
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = fmt.Errorf("rollback failed: %v (original error: %w)", rollbackErr, err)
			}
		} else {
			// Commit on success
			err = tx.Commit()
			if err != nil {
				err = NewTransactionError("failed to commit transaction", err)
			}
		}
	}()

	err = fn(ctx, tx)
	return err
}

// DuplicateTaskWithDependencies duplicates a task and its dependencies to a new project
func (r *sqliteRepository) DuplicateTaskWithDependencies(ctx context.Context, taskID, newProjectID uuid.UUID) (*types.Task, error) {
	var result *types.Task
	err := r.withTx(ctx, func(ctx context.Context, tx *ent.Tx) error {
		// Get original task
		originalTask, err := tx.Task.Query().
			Where(task.ID(taskID)).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewNotFoundError("task", taskID.String())
			}
			return fmt.Errorf("failed to get original task: %w", err)
		}

		// Create new task with new ID and project
		newTaskID := uuid.New()
		newTask := entTaskToTask(originalTask)
		newTask.ID = newTaskID
		newTask.ProjectID = newProjectID
		newTask.ParentID = nil // Don't duplicate parent relationships across projects

		// Create the new task
		_, err = taskToEntTaskCreate(newTask, tx.Client()).Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create duplicated task: %w", err)
		}

		// Note: Dependencies are not duplicated across projects as they may not exist
		// This is intentional to avoid creating invalid dependencies

		// Update project metrics for the new project
		err = r.updateProjectMetricsInTx(ctx, tx, newProjectID)
		if err != nil {
			return fmt.Errorf("failed to update project metrics: %w", err)
		}

		result = newTask
		return nil
	})
	return result, err
}

// TaskDependencyPair represents a task dependency relationship
type TaskDependencyPair struct {
	TaskID          uuid.UUID
	DependsOnTaskID uuid.UUID
}
