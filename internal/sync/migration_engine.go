package sync

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/sync/client"
	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"go.uber.org/zap"
)

// MigrationEngine handles synchronization between local and remote storage
// Simplified to work with complete datasets instead of individual operations
type MigrationEngine interface {
	// PushSync pushes local changes to remote server
	PushSync(ctx context.Context, dataSet *shared.SyncDataSet) (*SyncResult, error)

	// PullSync pulls remote changes from server and applies locally
	PullSync(ctx context.Context, projectID *uuid.UUID) (*SyncResult, error)

	// FullSync performs bidirectional synchronization
	FullSync(ctx context.Context, localDataSet *shared.SyncDataSet) (*SyncResult, error)
}

// migrationEngineImpl implements the MigrationEngine interface
type migrationEngineImpl struct {
	logger         logger.Logger
	client         client.RESTSyncClient
	projectManager manager.ProjectManager
	metrics        *SyncMetrics
}

// NewMigrationEngine creates a new migration engine using DI injector
func NewMigrationEngine(injector do.Injector) (MigrationEngine, error) {
	logger := do.MustInvoke[logger.Logger](injector)
	restClient := do.MustInvoke[client.RESTSyncClient](injector)
	projectManager := do.MustInvoke[manager.ProjectManager](injector)

	return &migrationEngineImpl{
		logger:         logger,
		client:         restClient,
		projectManager: projectManager,
		metrics: &SyncMetrics{
			StartTime: time.Now(),
		},
	}, nil
}

// PushSync pushes local dataset to remote server
func (e *migrationEngineImpl) PushSync(ctx context.Context, dataSet *shared.SyncDataSet) (*SyncResult, error) {
	e.logger.Info("Starting push sync",
		zap.Int("projects", len(dataSet.Projects)),
		zap.Int("tasks", len(dataSet.Tasks)))

	e.metrics.StartTime = time.Now()

	// Extract project ID from dataset (assuming single project sync for now)
	var projectID uuid.UUID
	for id := range dataSet.Projects {
		projectID = id
		break // Use first project ID
	}

	// Create sync request
	request := &shared.SyncRequest{
		RequestID: uuid.New(),
		ProjectID: projectID,
		Direction: shared.SyncLocalToMCP,
		LocalData: dataSet,
		Timestamp: time.Now(),
	}

	// Send to server
	response, err := e.client.Sync(ctx, request)
	if err != nil {
		e.metrics.FailedOps = len(dataSet.Projects) + len(dataSet.Tasks)
		return &SyncResult{
			Success:  false,
			Errors:   []string{err.Error()},
			SyncedAt: time.Now(),
		}, fmt.Errorf("push sync failed: %w", err)
	}

	e.metrics.EndTime = time.Now()
	e.metrics.Duration = e.metrics.EndTime.Sub(e.metrics.StartTime)

	e.logger.Info("Push sync completed successfully",
		zap.Duration("duration", e.metrics.Duration),
		zap.Int("processed", response.Processed),
		zap.Int("created", response.Created),
		zap.Int("updated", response.Updated),
		zap.Int("deleted", response.Deleted))

	// Convert conflicts from values to pointers
	conflicts := make([]*shared.SyncConflict, len(response.Conflicts))
	for i := range response.Conflicts {
		conflicts[i] = &response.Conflicts[i]
	}

	return &SyncResult{
		Success:   response.Success,
		Processed: response.Processed,
		Created:   response.Created,
		Updated:   response.Updated,
		Deleted:   response.Deleted,
		Conflicts: conflicts,
		Errors:    response.Errors,
		Duration:  e.metrics.Duration,
		SyncedAt:  time.Now(),
	}, nil
}

// PullSync pulls remote dataset and applies it locally
func (e *migrationEngineImpl) PullSync(ctx context.Context, projectID *uuid.UUID) (*SyncResult, error) {
	e.logger.Info("Starting pull sync",
		zap.String("project_id", projectID.String()))

	e.metrics.StartTime = time.Now()

	// Fetch remote data
	remoteData, err := e.client.GetRemoteData(ctx, projectID)
	if err != nil {
		return &SyncResult{
			Success:  false,
			Errors:   []string{err.Error()},
			SyncedAt: time.Now(),
		}, fmt.Errorf("pull sync failed: %w", err)
	}

	// Apply remote changes locally
	result := &SyncResult{
		Success:   true,
		Processed: 0,
		Created:   0,
		Updated:   0,
		Deleted:   0,
		Conflicts: make([]*shared.SyncConflict, 0),
		Errors:    make([]string, 0),
		SyncedAt:  time.Now(),
	}

	// Apply projects
	for _, project := range remoteData.Projects {
		_, err := e.projectManager.SyncCreateProjectWithTimestamps(ctx, project)
		if err != nil {
			// Try update instead
			updated, updateErr := e.projectManager.UpdateProject(ctx, project.ID, project.Title, project.Description, "sync-pull")
			if updateErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to sync project %s: %v", project.ID, err))
				result.Success = false
				e.metrics.FailedOps++
				continue
			}
			result.Updated++
			result.Processed++
			e.logger.Debug("Updated project from remote",
				zap.String("project_id", project.ID.String()),
				zap.String("title", updated.Title))
		} else {
			result.Created++
			result.Processed++
			e.logger.Debug("Created project from remote",
				zap.String("project_id", project.ID.String()),
				zap.String("title", project.Title))
		}
	}

	// Apply tasks in two passes:
	// Pass 1: Create all tasks WITHOUT dependencies (to avoid "dependency not found" errors)
	// Pass 2: Add dependencies after all tasks exist

	// Prepare tasks - save dependencies and clear them for first pass
	taskDeps := make(map[uuid.UUID][]uuid.UUID) // task ID -> dependencies
	taskSlice := make([]*types.Task, 0, len(remoteData.Tasks))
	for _, task := range remoteData.Tasks {
		// Save dependencies
		if len(task.Dependencies) > 0 {
			taskDeps[task.ID] = append([]uuid.UUID{}, task.Dependencies...)
		}
		// Clear dependencies for first pass
		taskCopy := *task
		taskCopy.Dependencies = nil
		taskSlice = append(taskSlice, &taskCopy)
	}

	// Sort by depth (parent tasks before child tasks)
	sort.Slice(taskSlice, func(i, j int) bool {
		return taskSlice[i].Depth < taskSlice[j].Depth
	})

	// Pass 1: Create all tasks without dependencies
	for _, task := range taskSlice {
		_, err := e.projectManager.SyncCreateTaskWithTimestamps(ctx, task)
		if err != nil {
			// Try update instead
			state := types.TaskState(task.State)
			updated, updateErr := e.projectManager.UpdateTask(ctx, task.ID, task.Title, task.Description, task.Complexity, state, "sync-pull")
			if updateErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to sync task %s: %v", task.ID, err))
				result.Success = false
				e.metrics.FailedOps++
				continue
			}
			result.Updated++
			result.Processed++
			e.logger.Debug("Updated task from remote",
				zap.String("task_id", task.ID.String()),
				zap.String("title", updated.Title))
		} else {
			result.Created++
			result.Processed++
			e.logger.Debug("Created task from remote",
				zap.String("task_id", task.ID.String()),
				zap.String("title", task.Title))
		}
	}

	// Pass 2: Add dependencies to tasks
	for taskID, deps := range taskDeps {
		for _, depID := range deps {
			_, err := e.projectManager.AddTaskDependency(ctx, taskID, depID, "sync-pull")
			if err != nil {
				// Log error but don't fail the entire sync
				e.logger.Warn("Failed to add dependency during sync",
					zap.String("task_id", taskID.String()),
					zap.String("depends_on", depID.String()),
					zap.Error(err))
				result.Errors = append(result.Errors, fmt.Sprintf("failed to add dependency %s -> %s: %v", taskID, depID, err))
			} else {
				e.logger.Debug("Added dependency during sync",
					zap.String("task_id", taskID.String()),
					zap.String("depends_on", depID.String()))
			}
		}
	}

	e.metrics.EndTime = time.Now()
	e.metrics.Duration = e.metrics.EndTime.Sub(e.metrics.StartTime)

	e.logger.Info("Pull sync completed successfully",
		zap.Duration("duration", e.metrics.Duration),
		zap.Int("processed", result.Processed),
		zap.Int("created", result.Created),
		zap.Int("updated", result.Updated),
		zap.Int("failed", e.metrics.FailedOps))

	return result, nil
}

// FullSync performs bidirectional synchronization
func (e *migrationEngineImpl) FullSync(ctx context.Context, localDataSet *shared.SyncDataSet) (*SyncResult, error) {
	e.logger.Info("Starting full bidirectional sync")

	e.metrics.StartTime = time.Now()

	// Extract project ID from dataset (assuming single project sync for now)
	var projectID uuid.UUID
	for id := range localDataSet.Projects {
		projectID = id
		break // Use first project ID
	}

	// Create sync request
	request := &shared.SyncRequest{
		RequestID: uuid.New(),
		ProjectID: projectID,
		Direction: shared.SyncBidirectional,
		LocalData: localDataSet,
		Timestamp: time.Now(),
	}

	// Step 1: Send local data to server to apply
	response, err := e.client.Sync(ctx, request)
	if err != nil {
		e.metrics.FailedOps = len(localDataSet.Projects) + len(localDataSet.Tasks)
		return &SyncResult{
			Success:  false,
			Errors:   []string{err.Error()},
			SyncedAt: time.Now(),
		}, fmt.Errorf("full sync failed: %w", err)
	}

	// Step 2: Apply remote changes returned from server to local storage
	result := &SyncResult{
		Success:   response.Success,
		Processed: 0, // Will count local applications
		Created:   0, // Will count local applications
		Updated:   0, // Will count local applications
		Deleted:   response.Deleted,
		Conflicts: make([]*shared.SyncConflict, 0),
		Errors:    response.Errors,
		SyncedAt:  time.Now(),
	}

	if response.RemoteChanges != nil {
		e.logger.Info("Applying remote changes locally",
			zap.Int("remote_projects", len(response.RemoteChanges.Projects)),
			zap.Int("remote_tasks", len(response.RemoteChanges.Tasks)))

		// Apply remote projects
		for _, project := range response.RemoteChanges.Projects {
			// Check if project already exists to correctly count create vs update
			_, err := e.projectManager.GetProject(ctx, project.ID)
			exists := err == nil

			_, err = e.projectManager.SyncCreateProjectWithTimestamps(ctx, project)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to sync remote project %s: %v", project.ID, err))
				result.Success = false
				e.metrics.FailedOps++
				continue
			}

			if exists {
				result.Updated++
				e.logger.Debug("Updated project from remote",
					zap.String("project_id", project.ID.String()),
					zap.String("title", project.Title))
			} else {
				result.Created++
				e.logger.Debug("Created project from remote",
					zap.String("project_id", project.ID.String()),
					zap.String("title", project.Title))
			}
			result.Processed++
		}

		// Apply remote tasks
		for _, task := range response.RemoteChanges.Tasks {
			// Check if task already exists to correctly count create vs update
			_, err := e.projectManager.GetTask(ctx, task.ID)
			exists := err == nil

			_, err = e.projectManager.SyncCreateTaskWithTimestamps(ctx, task)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to sync remote task %s: %v", task.ID, err))
				result.Success = false
				e.metrics.FailedOps++
				continue
			}

			if exists {
				result.Updated++
				e.logger.Debug("Updated task from remote",
					zap.String("task_id", task.ID.String()),
					zap.String("title", task.Title))
			} else {
				result.Created++
				e.logger.Debug("Created task from remote",
					zap.String("task_id", task.ID.String()),
					zap.String("title", task.Title))
			}
			result.Processed++
		}
	}

	// Convert conflicts from values to pointers
	conflicts := make([]*shared.SyncConflict, len(response.Conflicts))
	for i := range response.Conflicts {
		conflicts[i] = &response.Conflicts[i]
	}
	result.Conflicts = conflicts

	e.metrics.EndTime = time.Now()
	e.metrics.Duration = e.metrics.EndTime.Sub(e.metrics.StartTime)

	e.logger.Info("Full sync completed",
		zap.Duration("duration", e.metrics.Duration),
		zap.Int("processed", result.Processed),
		zap.Int("created", result.Created),
		zap.Int("updated", result.Updated),
		zap.Int("conflicts", len(result.Conflicts)),
		zap.Int("failed", e.metrics.FailedOps))

	return result, nil
}

// GetMetrics returns the migration metrics
func (e *migrationEngineImpl) GetMetrics() *SyncMetrics {
	return e.metrics
}

// ResetMetrics resets the migration metrics
func (e *migrationEngineImpl) ResetMetrics() {
	e.metrics = &SyncMetrics{
		StartTime: time.Now(),
	}
}
