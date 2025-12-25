// Package sync provides CLI commands for bidirectional synchronization operations in KNOT.
//
// This package implements command-line interface commands for synchronizing
// local CLI SQLite workspaces with MCP PostgreSQL database.
//
// Key Commands:
//   - Sync: Performs bidirectional synchronization with configurable direction
//
// Cross-reference: Knot Task knot-4jb (Implement bidirectional sync between CLI SQLite workspaces and MCP PostgreSQL)
package sync

import (
	"fmt"
	"os"

	knotshared "github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/sync"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// syncAction handles intelligent bidirectional synchronization
func syncAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		fmt.Fprintf(os.Stderr, "DEBUG: syncAction called\n")

		// Get services from DI container
		container := knotshared.GetContainerFromContext(c)
		syncManager, err := container.GetSyncManager()
		if err != nil {
			fmt.Fprintf(os.Stderr, "DEBUG: Error getting sync manager: %v\n", err)
			return fmt.Errorf("failed to get sync manager: %w", err)
		}
		loggerService := container.GetLogger()

		// Parse and validate project ID
		projectIDStr := c.String("project-id")
		fmt.Fprintf(os.Stderr, "DEBUG: projectIDStr = %s\n", projectIDStr)

		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "DEBUG: Error parsing UUID: %v\n", err)
			return fmt.Errorf("invalid project ID: %w", err)
		}

		// Determine sync direction
		direction := c.String("direction")
		fmt.Fprintf(os.Stderr, "DEBUG: direction = %s\n", direction)

		loggerService.Info("Starting intelligent sync",
			zap.String("project_id", projectID.String()),
			zap.String("direction", direction))

		var result *sync.SyncResult

		fmt.Fprintf(os.Stderr, "DEBUG: About to switch on direction\n")

		switch direction {
		case "push-only", "push":
			fmt.Fprintf(os.Stderr, "DEBUG: Push sync\n")
			// Extract local data and push to remote
			dataExtractor, err := container.GetDataExtractor()
			if err != nil {
				return fmt.Errorf("failed to get data extractor: %w", err)
			}
			projectManager := container.GetProjectManager()

			// Get project and extract local data
			project, err := projectManager.GetProject(c.Context, projectID)
			if err != nil {
				return fmt.Errorf("failed to get project: %w", err)
			}

			localData, err := dataExtractor.ExtractLocalData(c.Context, projectManager, projectID)
			if err != nil {
				return fmt.Errorf("failed to extract local data: %w", err)
			}
			// Add project to the dataset
			localData.Projects[project.ID] = project

			result, err = syncManager.SyncWithLocalData(c.Context, localData)
			if err != nil {
				loggerService.Error("Push sync failed", zap.Error(err))
				return fmt.Errorf("push sync failed: %w", err)
			}

		case "pull-only", "pull":
			// Pull from remote and apply locally
			result, err = syncManager.SyncPullFromRemote(c.Context, projectID)
			if err != nil {
				loggerService.Error("Pull sync failed", zap.Error(err))
				return fmt.Errorf("pull sync failed: %w", err)
			}

		case "bidirectional", "bi":
			fmt.Fprintf(os.Stderr, "DEBUG: Bidirectional sync\n")
			// Extract local data and perform bidirectional sync
			dataExtractor, err := container.GetDataExtractor()
			if err != nil {
				return fmt.Errorf("failed to get data extractor: %w", err)
			}
			projectManager := container.GetProjectManager()

			// Get project and extract local data
			project, err := projectManager.GetProject(c.Context, projectID)
			if err != nil {
				return fmt.Errorf("failed to get project: %w", err)
			}

			localData, err := dataExtractor.ExtractLocalData(c.Context, projectManager, projectID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "DEBUG: Error extracting local data: %v\n", err)
				return fmt.Errorf("failed to extract local data: %w", err)
			}
			// Add project to the dataset
			localData.Projects[project.ID] = project

			// Debug: Show local data being sent
			fmt.Fprintf(os.Stderr, "DEBUG: Sending %d local tasks\n", len(localData.Tasks))
			count := 0
			for id, task := range localData.Tasks {
				if count < 3 {
					fmt.Fprintf(os.Stderr, "DEBUG: Local task: %s - %s\n", id, task.Title)
					count++
				}
			}

			fmt.Fprintf(os.Stderr, "DEBUG: About to call SyncBidirectional\n")
			result, err = syncManager.SyncBidirectional(c.Context, localData)
			if err != nil {
				fmt.Fprintf(os.Stderr, "DEBUG: SyncBidirectional error: %v\n", err)
				loggerService.Error("Bidirectional sync failed", zap.Error(err))
				return fmt.Errorf("bidirectional sync failed: %w", err)
			}
			fmt.Fprintf(os.Stderr, "DEBUG: SyncBidirectional completed - Created: %d, Updated: %d, Errors: %d\n",
				result.Created, result.Updated, len(result.Errors))

		default:
			return fmt.Errorf("invalid direction '%s'. Valid options: push, push-only, pull, pull-only, bi, bidirectional", direction)
		}

		// Display results
		printSyncResult(result)
		return nil
	}
}

// Commands returns all sync-related CLI commands
func Commands() []*cli.Command {
	return []*cli.Command{
		{
			Name:   "sync",
			Usage:  "Intelligent bidirectional synchronization with MCP server",
			Action: syncAction(),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "project-id",
					Aliases:  []string{"p"},
					Usage:    "Project ID to sync",
					Required: true,
				},
				&cli.StringFlag{
					Name:  "direction",
					Usage: "Sync direction: push, pull, or bi (default: bi)",
					Value: "bi",
				},
			},
		},
	}
}

// printSyncResult displays sync operation results
func printSyncResult(result *sync.SyncResult) {
	fmt.Printf("\nSync Result:\n")
	if result.Success {
		fmt.Printf("  ✓ Success: %t\n", result.Success)
	} else {
		fmt.Printf("  ✗ Success: %t\n", result.Success)
	}
	fmt.Printf("  Duration: %v\n", result.Duration)
	fmt.Printf("  Processed: %d\n", result.Processed)
	fmt.Printf("  Created: %d\n", result.Created)
	fmt.Printf("  Updated: %d\n", result.Updated)
	fmt.Printf("  Deleted: %d\n", result.Deleted)
	fmt.Printf("  Conflicts Resolved: %d\n", len(result.Conflicts))

	if len(result.Errors) > 0 {
		fmt.Printf("  Errors:\n")
		for _, err := range result.Errors {
			fmt.Printf("    - %s\n", err)
		}
	}
}
