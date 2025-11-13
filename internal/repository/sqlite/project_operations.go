package sqlite

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/internal/repository/sqlite/ent"
	"github.com/denkhaus/knot/internal/repository/sqlite/ent/project"
	"github.com/denkhaus/knot/internal/repository/sqlite/ent/projectcontext"
	"github.com/denkhaus/knot/internal/repository/sqlite/ent/task"
	"github.com/denkhaus/knot/internal/repository/sqlite/ent/taskdependency"
	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Project CRUD Operations

func (r *sqliteRepository) CreateProject(ctx context.Context, project *types.Project) error {
	_, err := projectToEntProjectCreate(project, r.client).Save(ctx)
	if err != nil {
		return r.mapError("create project", err)
	}
	return nil
}

// GetProject retrieves a project by ID using ent
func (r *sqliteRepository) GetProject(ctx context.Context, id uuid.UUID) (*types.Project, error) {
	entProject, err := r.client.Project.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, NewNotFoundError("project", id.String())
		}
		return nil, r.mapError("get project", err)
	}
	return entProjectToProject(entProject), nil
}

// UpdateProject updates an existing project using ent
func (r *sqliteRepository) UpdateProject(ctx context.Context, project *types.Project) error {
	err := r.client.Project.UpdateOneID(project.ID).
		SetTitle(project.Title).
		SetDescription(project.Description).
		SetState(projectStateToEntState(project.State)).
		SetUpdatedAt(project.UpdatedAt).
		SetTotalTasks(project.TotalTasks).
		SetCompletedTasks(project.CompletedTasks).
		SetProgress(project.Progress).
		Exec(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return NewNotFoundError("project", project.ID.String())
		}
		return r.mapError("update project", err)
	}
	return nil
}

// DeleteProject deletes a project and all its tasks using ent transaction
func (r *sqliteRepository) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return r.withTx(ctx, func(ctx context.Context, tx *ent.Tx) error {
		// First, clear any project context that references this project
		_, err := tx.ProjectContext.Delete().
			Where(projectcontext.SelectedProjectIDEQ(id)).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to clear project context: %w", err)
		}

		// Delete all task dependencies for tasks in this project
		taskIDs, err := tx.Task.Query().
			Where(task.ProjectID(id)).
			IDs(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return fmt.Errorf("failed to get task IDs for project: %w", err)
		}

		if len(taskIDs) > 0 {
			// Delete all task dependencies
			_, err = tx.TaskDependency.Delete().
				Where(taskdependency.Or(
					taskdependency.TaskIDIn(taskIDs...),
					taskdependency.DependsOnTaskIDIn(taskIDs...),
				)).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to delete task dependencies: %w", err)
			}

			// Delete all tasks in the project
			_, err = tx.Task.Delete().
				Where(task.ProjectID(id)).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to delete tasks: %w", err)
			}
		}

		// Delete the project
		err = tx.Project.DeleteOneID(id).Exec(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return NewNotFoundError("project", id.String())
			}
			return fmt.Errorf("failed to delete project: %w", err)
		}

		return nil
	})
}

// ListProjects retrieves all projects using ent
func (r *sqliteRepository) ListProjects(ctx context.Context) ([]*types.Project, error) {
	entProjects, err := r.client.Project.Query().
		Order(ent.Asc(project.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, r.mapError("list projects", err)
	}

	// Debug: Log the number of projects found
	r.config.Logger.Info("DEBUG: Found projects in database",
		zap.Int("count", len(entProjects)))
	for i, proj := range entProjects {
		r.config.Logger.Info("DEBUG: Project details",
			zap.Int("index", i),
			zap.String("id", proj.ID.String()),
			zap.String("title", proj.Title),
			zap.String("state", string(proj.State)))
	}

	return entProjectsToProjects(entProjects), nil
}

// updateProjectMetrics updates project metrics (total tasks, completed tasks, progress)
// Currently unused but kept for potential future use
// func (r *sqliteRepository) updateProjectMetrics(ctx context.Context, projectID uuid.UUID) error {
// 	// Get task counts by state using ent aggregation
// 	var totalTasks int
// 	var completedTasks int
//
// 	// Count total tasks
// 	totalTasks, err := r.client.Task.Query().
// 		Where(task.ProjectID(projectID)).
// 		Count(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to count total tasks: %w", err)
// 	}
//
// 	// Count completed tasks
// 	completedTasks, err = r.client.Task.Query().
// 		Where(
// 			task.ProjectID(projectID),
// 			task.StateEQ(task.StateCompleted),
// 		).
// 		Count(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to count completed tasks: %w", err)
// 	}
//
// 	// Calculate progress
// 	progress := 0.0
// 	if totalTasks > 0 {
// 		progress = float64(completedTasks) / float64(totalTasks) * 100.0
// 	}
//
// 	// Update project metrics
// 	err = r.client.Project.UpdateOneID(projectID).
// 		SetTotalTasks(totalTasks).
// 		SetCompletedTasks(completedTasks).
// 		SetProgress(progress).
// 		Exec(ctx)
//
// 	if err != nil {
// 		return fmt.Errorf("failed to update project metrics: %w", err)
// 	}
//
// 	return nil
// }

// updateProjectMetricsInTx updates project metrics within a transaction
func (r *sqliteRepository) updateProjectMetricsInTx(ctx context.Context, tx *ent.Tx, projectID uuid.UUID) error {
	// Get task counts by state using ent aggregation within transaction
	var totalTasks int
	var completedTasks int

	// Count total tasks
	totalTasks, err := tx.Task.Query().
		Where(task.ProjectID(projectID)).
		Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count total tasks: %w", err)
	}

	// Count completed tasks
	completedTasks, err = tx.Task.Query().
		Where(
			task.ProjectID(projectID),
			task.StateEQ(task.StateCompleted),
		).
		Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count completed tasks: %w", err)
	}

	// Calculate progress
	progress := 0.0
	if totalTasks > 0 {
		progress = float64(completedTasks) / float64(totalTasks) * 100.0
	}

	// Update project metrics within transaction
	err = tx.Project.UpdateOneID(projectID).
		SetTotalTasks(totalTasks).
		SetCompletedTasks(completedTasks).
		SetProgress(progress).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update project metrics: %w", err)
	}

	return nil
}
