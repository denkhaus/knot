package project

import (
	"context"
	"fmt"
	"os"

	"github.com/denkhaus/knot/internal/errors"
	"github.com/denkhaus/knot/internal/shared"
	"github.com/denkhaus/knot/internal/types"
	"github.com/denkhaus/knot/internal/validation"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// validateProjectID validates and returns the project ID from the CLI context
func validateProjectID(c *cli.Context) (uuid.UUID, error) {
	projectIDStr := c.String("project-id")
	if projectIDStr == "" {
		return uuid.Nil, errors.MissingRequiredFlagError("project-id", c.Command.FullName())
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return uuid.Nil, errors.InvalidUUIDError("project-id", projectIDStr)
	}
	return projectID, nil
}

// Commands returns all project-related CLI commands
func Commands(appCtx *shared.AppContext) []*cli.Command {
	return []*cli.Command{
		{
			Name:   "create",
			Usage:  "Create a new project",
			Action: createAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "title",
					Aliases:  []string{"t"},
					Usage:    "Project title",
					Required: true,
				},
				&cli.StringFlag{
					Name:    "description",
					Aliases: []string{"d"},
					Usage:   "Project description",
				},
			},
		},
		{
			Name:   "list",
			Usage:  "List all projects",
			Action: listAction(appCtx),
		},
		{
			Name:   "get",
			Usage:  "Get project details",
			Action: getAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "id",
					Usage:    "Project ID",
					Required: true,
				},
			},
		},
		{
			Name:   "delete",
			Usage:  "Delete a project with two-step confirmation",
			Action: deleteAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "id",
					Usage:    "Project ID to delete",
					Required: true,
				},
				&cli.BoolFlag{
					Name:  "dry-run",
					Usage: "Show what would be deleted without actually deleting",
					Value: false,
				},
			},
		},
	}
}

func createAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		title := c.String("title")
		description := c.String("description")
		actor := c.String("actor")

		// Create input validator
		validator := validation.NewInputValidator()

		// Validate inputs
		if err := validator.ValidateProjectTitle(title); err != nil {
			return errors.NewValidationError("invalid project title", err)
		}

		if err := validator.ValidateProjectDescription(description); err != nil {
			return errors.NewValidationError("invalid project description", err)
		}

		// Default to $USER if actor is not provided
		if actor == "" {
			actor = os.Getenv("USER")
			if actor == "" {
				actor = "unknown"
			}
		}

		appCtx.Logger.Info("Creating project", zap.String("title", title), zap.String("description", description), zap.String("actor", actor))

		project, err := appCtx.ProjectManager.CreateProject(context.Background(), title, description, actor)
		if err != nil {
			appCtx.Logger.Error("Failed to create project", zap.Error(err))
			return errors.WrapWithSuggestion(err, "creating project")
		}

		appCtx.Logger.Info("Project created successfully", zap.String("projectID", project.ID.String()), zap.String("title", project.Title), zap.String("actor", actor))
		fmt.Printf("Created project: %s (ID: %s)\n", project.Title, project.ID)
		fmt.Printf("  Created by: %s\n", actor)
		if project.Description != "" {
			fmt.Printf("  Description: %s\n", project.Description)
		}
		return nil
	}
}

func listAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		appCtx.Logger.Info("Listing projects")

		projects, err := appCtx.ProjectManager.ListProjects(context.Background())
		if err != nil {
			appCtx.Logger.Error("Failed to list projects", zap.Error(err))
			return errors.WrapWithSuggestion(err, "listing projects")
		}

		appCtx.Logger.Info("Projects retrieved", zap.Int("count", len(projects)))

		if len(projects) == 0 {
			return errors.EmptyResultError("list projects", "current workspace")
		}

		fmt.Printf("Found %d project(s):\n\n", len(projects))
		for _, project := range projects {
			fmt.Printf("• %s (ID: %s)\n", project.Title, project.ID)
			if project.Description != "" {
				fmt.Printf("  %s\n", project.Description)
			}
			fmt.Printf("  Progress: %.1f%% (%d/%d tasks completed)\n",
				project.Progress, project.CompletedTasks, project.TotalTasks)
			fmt.Println()
		}
		return nil
	}
}

func getAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		idStr := c.String("id")
		projectID, err := uuid.Parse(idStr)
		if err != nil {
			return fmt.Errorf("invalid project ID: %w", err)
		}

		appCtx.Logger.Info("Getting project", zap.String("projectID", projectID.String()))

		project, err := appCtx.ProjectManager.GetProject(context.Background(), projectID)
		if err != nil {
			appCtx.Logger.Error("Failed to get project", zap.Error(err))
			return fmt.Errorf("failed to get project: %w", err)
		}

		fmt.Printf("Project: %s\n", project.Title)
		fmt.Printf("ID: %s\n", project.ID)
		if project.Description != "" {
			fmt.Printf("Description: %s\n", project.Description)
		}
		fmt.Printf("Progress: %.1f%% (%d/%d tasks completed)\n",
			project.Progress, project.CompletedTasks, project.TotalTasks)
		fmt.Printf("Created: %s\n", project.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated: %s\n", project.UpdatedAt.Format("2006-01-02 15:04:05"))

		return nil
	}
}

func deleteAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		projectIDStr := c.String("id")
		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			return &errors.EnhancedError{
				Operation:   "parsing project ID",
				Cause:       err,
				Suggestion:  "Provide a valid UUID for the project ID",
				Example:     "knot project delete --id 550e8400-e29b-41d4-a716-446655440000",
				HelpCommand: "knot project delete --help",
			}
		}

		dryRun := c.Bool("dry-run")

		// Get project details
		project, err := appCtx.ProjectManager.GetProject(context.Background(), projectID)
		if err != nil {
			return &errors.EnhancedError{
				Operation:   "retrieving project",
				Cause:       err,
				Suggestion:  "Verify the project ID exists",
				Example:     "knot project list",
				HelpCommand: "knot project get --help",
			}
		}

		// Check if project has tasks
		tasks, err := appCtx.ProjectManager.ListTasksForProject(context.Background(), projectID)
		if err != nil {
			return &errors.EnhancedError{
				Operation:   "checking project tasks",
				Cause:       err,
				Suggestion:  "Unable to verify if project has tasks",
				HelpCommand: "knot task list --help",
			}
		}

		// Two-step deletion process
		if project.State == types.ProjectStateDeletionPending {
			// Second call - actually delete the project
			if dryRun {
				fmt.Printf("🔍 DRY RUN: Project would be permanently deleted (no actual changes made)\n")
				return nil
			}

			// Show what will be deleted
			fmt.Printf("🗑️  Final deletion of project:\n")
			fmt.Printf("  • %s (ID: %s)\n", project.Title, project.ID)
			if project.Description != "" {
				fmt.Printf("    %s\n", project.Description)
			}
			if len(tasks) > 0 {
				fmt.Printf("    ⚠️  This will also delete %d task(s)\n", len(tasks))
			}

			// Perform deletion
			err = appCtx.ProjectManager.DeleteProject(context.Background(), projectID)
			if err != nil {
				return &errors.EnhancedError{
					Operation:   "deleting project",
					Cause:       err,
					Suggestion:  "Check if the project still exists or if there are constraint violations",
					HelpCommand: "knot project get --help",
				}
			}

			fmt.Printf("✅ Project permanently deleted: %s\n", project.Title)
			return nil
		} else {
			// First call - mark for deletion
			if dryRun {
				fmt.Printf("🔍 DRY RUN: Project would be marked for deletion (no actual changes made)\n")
				return nil
			}

			// Show what will be marked for deletion
			fmt.Printf("📋 Project to be marked for deletion:\n")
			fmt.Printf("  • %s (ID: %s)\n", project.Title, project.ID)
			if project.Description != "" {
				fmt.Printf("    %s\n", project.Description)
			}
			fmt.Printf("    Current State: %s\n", project.State)
			fmt.Printf("    Progress: %.1f%% (%d/%d tasks)\n", project.Progress, project.CompletedTasks, project.TotalTasks)

			if len(tasks) > 0 {
				fmt.Printf("\n  ⚠️  This project contains %d task(s):\n", len(tasks))
				for i, task := range tasks {
					if i < 5 { // Show first 5 tasks
						fmt.Printf("    • %s (%s)\n", task.Title, task.State)
					} else if i == 5 {
						fmt.Printf("    • ... and %d more task(s)\n", len(tasks)-5)
						break
					}
				}
				fmt.Printf("    All tasks will be deleted with the project.\n")
			}

			// Mark project for deletion
			_, err = appCtx.ProjectManager.UpdateProjectState(context.Background(), projectID, types.ProjectStateDeletionPending, appCtx.Actor)
			if err != nil {
				return &errors.EnhancedError{
					Operation:   "marking project for deletion",
					Cause:       err,
					Suggestion:  "Check if the project state transition is valid",
					HelpCommand: "knot project get --help",
				}
			}

			fmt.Printf("\n⚠️  Project marked for deletion. To confirm deletion, run the same command again:\n")
			fmt.Printf("    knot project delete --id %s\n", projectID)
			fmt.Printf("\n💡 To cancel deletion, change the project state:\n")
			fmt.Printf("    knot project update-state --id %s --state active\n", projectID)

			return nil
		}
	}
}
