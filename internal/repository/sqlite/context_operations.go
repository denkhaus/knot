package sqlite

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/repository/sqlite/ent"
	"github.com/denkhaus/knot/v2/internal/repository/sqlite/ent/project"
	"github.com/denkhaus/knot/v2/internal/repository/sqlite/ent/projectcontext"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetSelectedProject retrieves the currently selected project ID from the database
func (r *sqliteRepository) GetSelectedProject(ctx context.Context) (*uuid.UUID, error) {
	r.logger.Debug("Getting selected project from database")

	// Query the singleton project context record
	pc, err := r.client.ProjectContext.Query().
		Where(projectcontext.IDEQ(1)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			r.logger.Debug("No project context found")
			return nil, nil
		}
		r.logger.Error("Failed to query project context", zap.Error(err))
		return nil, fmt.Errorf("failed to get selected project: %w", err)
	}

	r.logger.Debug("Retrieved selected project", zap.String("projectID", pc.SelectedProjectID.String()))
	return &pc.SelectedProjectID, nil
}

// SetSelectedProject sets the currently selected project ID in the database
func (r *sqliteRepository) SetSelectedProject(ctx context.Context, projectID uuid.UUID, actor string) error {
	r.logger.Debug("Setting selected project in database",
		zap.String("projectID", projectID.String()),
		zap.String("actor", actor))

	// First verify the project exists
	exists, err := r.client.Project.Query().
		Where(project.IDEQ(projectID)).
		Exist(ctx)
	if err != nil {
		r.logger.Error("Failed to check if project exists", zap.Error(err))
		return fmt.Errorf("failed to verify project exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("project with ID %s does not exist", projectID)
	}

	// Use upsert pattern - try to update existing record, create if not exists
	// Try to update existing record first
	updated, err := r.client.ProjectContext.Update().
		Where(projectcontext.IDEQ(1)).
		SetSelectedProjectID(projectID).
		SetUpdatedBy(actor).
		Save(ctx)

	r.logger.Debug("Update attempt result", zap.Int("updated", updated), zap.Error(err))

	if err != nil || updated == 0 {
		// Create new record if update failed or no record exists
		r.logger.Debug("Creating new project context record")
		err = r.client.ProjectContext.Create().
			SetID(1).
			SetSelectedProjectID(projectID).
			SetUpdatedBy(actor).
			Exec(ctx)
		r.logger.Debug("Create attempt result", zap.Error(err))
	}

	if err != nil {
		r.logger.Error("Failed to set selected project", zap.Error(err))
		return fmt.Errorf("failed to set selected project: %w", err)
	}

	r.logger.Info("Selected project updated",
		zap.String("projectID", projectID.String()),
		zap.String("actor", actor))
	return nil
}

// ClearSelectedProject removes the currently selected project from the database
func (r *sqliteRepository) ClearSelectedProject(ctx context.Context) error {
	r.logger.Debug("Clearing selected project from database")

	// Delete the singleton record
	_, err := r.client.ProjectContext.Delete().
		Where(projectcontext.IDEQ(1)).
		Exec(ctx)
	if err != nil {
		r.logger.Error("Failed to clear selected project", zap.Error(err))
		return fmt.Errorf("failed to clear selected project: %w", err)
	}

	r.logger.Info("Selected project cleared")
	return nil
}

// HasSelectedProject checks if there is a currently selected project
func (r *sqliteRepository) HasSelectedProject(ctx context.Context) (bool, error) {
	r.logger.Debug("Checking if project is selected")

	exists, err := r.client.ProjectContext.Query().
		Where(projectcontext.IDEQ(1)).
		Exist(ctx)
	if err != nil {
		r.logger.Error("Failed to check if project is selected", zap.Error(err))
		return false, fmt.Errorf("failed to check selected project: %w", err)
	}

	r.logger.Debug("Project selection status", zap.Bool("hasSelected", exists))
	return exists, nil
}
