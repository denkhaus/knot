package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	knotsync "github.com/denkhaus/knot/v2/internal/sync"
	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"go.uber.org/zap"
)

// syncServiceImpl implements SyncService interface and bridges HTTP handlers to database
type syncServiceImpl struct {
	projectManager manager.ProjectManager
	logger         logger.Logger
	diffEngine     knotsync.DiffEngine
}

// Ensure syncService implements SyncService
var _ SyncService = (*syncServiceImpl)(nil)

// NewSyncService creates a new sync service instance following DI pattern
func NewSyncService(injector do.Injector) (SyncService, error) {
	// Resolve dependencies using do.MustInvoke as per DI pattern
	projectManager := do.MustInvoke[manager.ProjectManager](injector)
	logger := do.MustInvoke[logger.Logger](injector)
	diffEngine := do.MustInvoke[knotsync.DiffEngine](injector)

	return &syncServiceImpl{
		projectManager: projectManager,
		logger:         logger,
		diffEngine:     diffEngine,
	}, nil
}

// PerformFullSync performs bidirectional synchronization using DiffEngine
func (s *syncServiceImpl) PerformFullSync(ctx context.Context, request shared.SyncRequest) (*shared.SyncResponse, error) {
	s.logger.Info("Performing full bidirectional sync",
		zap.String("request_id", request.RequestID.String()),
		zap.String("project_id", request.ProjectID.String()),
		zap.String("direction", string(request.Direction)))

	startTime := time.Now()

	// Step 1: Fetch current remote data from database
	remoteData := &shared.SyncDataSet{
		Projects: make(map[uuid.UUID]*types.Project),
		Tasks:    make(map[uuid.UUID]*types.Task),
	}

	// Fetch project data (may not exist for fresh sync)
	project, err := s.projectManager.GetProject(ctx, request.ProjectID)
	if err == nil {
		// Project exists, add to remote data
		remoteData.Projects[project.ID] = project

		// Fetch all tasks for project
		tasks, listErr := s.projectManager.ListTasksForProject(ctx, request.ProjectID)
		if listErr != nil {
			return nil, fmt.Errorf("failed to list tasks: %w", listErr)
		}
		for _, task := range tasks {
			remoteData.Tasks[task.ID] = task
		}
	}
	// If project doesn't exist, remoteData stays empty - this is a fresh sync

	// Step 2: Calculate diff between local and remote data
	diffResult, err := s.diffEngine.CalculateDiff(ctx, request.LocalData, remoteData)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate diff: %w", err)
	}

	// Step 3: Apply sync operations based on diff
	result := &knotsync.SyncResult{
		SyncedAt:  time.Now(),
		Success:   true,
		Processed: 0,
		Created:   0,
		Updated:   0,
		Deleted:   0,
		Conflicts: make([]*shared.SyncConflict, 0),
		Errors:    make([]string, 0),
	}

	// Apply operations
	for _, op := range diffResult.Operations {
		if err := s.applySyncOperation(ctx, op, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to apply operation %s: %v", op.ID, err))
			result.Success = false
		}
	}

	// Step 4: Fetch updated remote data to return to client
	updatedRemoteData := &shared.SyncDataSet{
		Projects: make(map[uuid.UUID]*types.Project),
		Tasks:    make(map[uuid.UUID]*types.Task),
	}

	project, err = s.projectManager.GetProject(ctx, request.ProjectID)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to get updated project: %v", err))
	} else {
		updatedRemoteData.Projects[project.ID] = project
	}

	tasks, listErr := s.projectManager.ListTasksForProject(ctx, request.ProjectID)
	if listErr != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to list updated tasks: %v", listErr))
	} else {
		for _, task := range tasks {
			updatedRemoteData.Tasks[task.ID] = task
		}
	}

	response := &shared.SyncResponse{
		Success:       result.Success,
		RequestID:     request.RequestID,
		Processed:     result.Processed,
		Created:       result.Created,
		Updated:       result.Updated,
		Deleted:       result.Deleted,
		NewVersion:    result.SyncedAt.Unix(),
		Timestamp:     result.SyncedAt,
		Duration:      time.Since(startTime),
		RemoteChanges: updatedRemoteData,
	}

	if len(result.Errors) > 0 {
		response.Errors = result.Errors
	}

	s.logger.Info("Full sync completed",
		zap.String("request_id", request.RequestID.String()),
		zap.Duration("duration", response.Duration),
		zap.Int("processed", response.Processed),
		zap.Int("created", response.Created),
		zap.Int("updated", response.Updated),
		zap.Int("remote_projects", len(updatedRemoteData.Projects)),
		zap.Int("remote_tasks", len(updatedRemoteData.Tasks)))

	return response, nil
}

// PerformPushSync performs push-only synchronization (local to remote)
func (s *syncServiceImpl) PerformPushSync(ctx context.Context, request shared.SyncRequest) (*shared.SyncResponse, error) {
	s.logger.Info("Performing push sync (local to remote)",
		zap.String("request_id", request.RequestID.String()),
		zap.String("project_id", request.ProjectID.String()),
		zap.Int("projects", len(request.LocalData.Projects)),
		zap.Int("tasks", len(request.LocalData.Tasks)))

	startTime := time.Now()

	// Apply local data directly to database
	result := s.applyLocalData(ctx, request.LocalData)

	response := &shared.SyncResponse{
		Success:       result.Success,
		RequestID:     request.RequestID,
		Processed:     result.Processed,
		Created:       result.Created,
		Updated:       result.Updated,
		Deleted:       result.Deleted,
		NewVersion:    result.SyncedAt.Unix(),
		Timestamp:     result.SyncedAt,
		Duration:      time.Since(startTime),
		RemoteChanges: request.LocalData,
	}

	if len(result.Errors) > 0 {
		response.Errors = result.Errors
	}

	s.logger.Info("Push sync completed",
		zap.String("request_id", request.RequestID.String()),
		zap.Duration("duration", response.Duration),
		zap.Int("processed", response.Processed))

	return response, nil
}

// PerformPullSync performs pull-only synchronization (remote to local)
func (s *syncServiceImpl) PerformPullSync(ctx context.Context, request shared.SyncRequest) (*shared.SyncResponse, error) {
	s.logger.Info("Performing pull sync (remote to local)",
		zap.String("request_id", request.RequestID.String()),
		zap.String("project_id", request.ProjectID.String()))

	startTime := time.Now()

	// For pull sync, we need to fetch the current data from the database
	// and return it to the client
	response := &shared.SyncResponse{
		Success:    true,
		RequestID:  request.RequestID,
		Processed:  0,
		Created:    0,
		Updated:    0,
		Deleted:    0,
		NewVersion: time.Now().Unix(),
		Timestamp:  time.Now(),
		Duration:   time.Since(startTime),
		RemoteChanges: &shared.SyncDataSet{
			Projects: make(map[uuid.UUID]*types.Project),
			Tasks:    make(map[uuid.UUID]*types.Task),
		},
	}

	// Fetch project data
	project, err := s.projectManager.GetProject(ctx, request.ProjectID)
	if err != nil {
		response.Success = false
		response.Errors = []string{fmt.Sprintf("failed to get project: %v", err)}
		return response, fmt.Errorf("failed to get project: %w", err)
	}

	response.RemoteChanges.Projects[project.ID] = project

	// Fetch tasks for project
	tasks, err := s.projectManager.ListTasksForProject(ctx, request.ProjectID)
	if err != nil {
		response.Success = false
		response.Errors = []string{fmt.Sprintf("failed to list tasks: %v", err)}
		return response, fmt.Errorf("failed to list tasks: %w", err)
	}

	for _, task := range tasks {
		response.RemoteChanges.Tasks[task.ID] = task
	}

	response.Processed = 1 + len(tasks)

	s.logger.Info("Pull sync completed",
		zap.String("request_id", request.RequestID.String()),
		zap.Duration("duration", response.Duration),
		zap.Int("processed", response.Processed))

	return response, nil
}

// HealthCheck performs health check for sync service
func (s *syncServiceImpl) HealthCheck(ctx context.Context) error {
	// Check if project manager is available
	if s.projectManager == nil {
		return fmt.Errorf("project manager not available")
	}

	s.logger.Debug("Sync service health check passed")
	return nil
}

// applySyncOperation applies a single sync operation based on diff result
func (s *syncServiceImpl) applySyncOperation(ctx context.Context, op shared.SyncOperation, result *knotsync.SyncResult) error {
	switch op.EntityType {
	case shared.EntityProject:
		return s.applyProjectOperation(ctx, op, result)
	case shared.EntityTask:
		return s.applyTaskOperation(ctx, op, result)
	default:
		return fmt.Errorf("unknown entity type: %s", op.EntityType)
	}
}

// applyProjectOperation applies a project sync operation
func (s *syncServiceImpl) applyProjectOperation(ctx context.Context, op shared.SyncOperation, result *knotsync.SyncResult) error {
	var project *types.Project

	// Get the project data based on direction
	if op.Direction == shared.SyncLocalToMCP {
		// Local → Remote: use LocalData
		if op.LocalData == nil {
			return fmt.Errorf("local data is required for local-to-MCP operation")
		}
		project = op.LocalData.(*types.Project)
	} else {
		// Remote → Local: use RemoteData
		if op.RemoteData == nil {
			return fmt.Errorf("remote data is required for MCP-to-local operation")
		}
		project = op.RemoteData.(*types.Project)
	}

	switch op.Type {
	case shared.OpCreate:
		_, err := s.projectManager.SyncCreateProjectWithTimestamps(ctx, project)
		if err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}
		result.Created++
		result.Processed++
		s.logger.Debug("Created project from sync operation",
			zap.String("project_id", project.ID.String()),
			zap.String("title", project.Title))

	case shared.OpUpdate:
		updated, err := s.projectManager.UpdateProject(ctx, project.ID, project.Title, project.Description, "sync-bi")
		if err != nil {
			return fmt.Errorf("failed to update project: %w", err)
		}
		result.Updated++
		result.Processed++
		s.logger.Debug("Updated project from sync operation",
			zap.String("project_id", project.ID.String()),
			zap.String("title", updated.Title))

	default:
		return fmt.Errorf("unsupported operation type: %s", op.Type)
	}

	return nil
}

// applyTaskOperation applies a task sync operation
func (s *syncServiceImpl) applyTaskOperation(ctx context.Context, op shared.SyncOperation, result *knotsync.SyncResult) error {
	var task *types.Task

	// Get the task data based on direction
	if op.Direction == shared.SyncLocalToMCP {
		// Local → Remote: use LocalData
		if op.LocalData == nil {
			return fmt.Errorf("local data is required for local-to-MCP operation")
		}
		task = op.LocalData.(*types.Task)
	} else {
		// Remote → Local: use RemoteData
		if op.RemoteData == nil {
			return fmt.Errorf("remote data is required for MCP-to-local operation")
		}
		task = op.RemoteData.(*types.Task)
	}

	switch op.Type {
	case shared.OpCreate:
		_, err := s.projectManager.SyncCreateTaskWithTimestamps(ctx, task)
		if err != nil {
			return fmt.Errorf("failed to create task: %w", err)
		}
		result.Created++
		result.Processed++
		s.logger.Debug("Created task from sync operation",
			zap.String("task_id", task.ID.String()),
			zap.String("title", task.Title))

	case shared.OpUpdate:
		state := types.TaskState(task.State)
		updated, err := s.projectManager.UpdateTask(ctx, task.ID, task.Title, task.Description, task.Complexity, state, "sync-bi")
		if err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}
		result.Updated++
		result.Processed++
		s.logger.Debug("Updated task from sync operation",
			zap.String("task_id", task.ID.String()),
			zap.String("title", updated.Title))

	default:
		return fmt.Errorf("unsupported operation type: %s", op.Type)
	}

	return nil
}

// applyLocalData applies local data directly to the database (deprecated - use diff-based approach)
func (s *syncServiceImpl) applyLocalData(ctx context.Context, localData *shared.SyncDataSet) *knotsync.SyncResult {
	result := &knotsync.SyncResult{
		SyncedAt:  time.Now(),
		Success:   true,
		Processed: 0,
		Created:   0,
		Updated:   0,
		Deleted:   0,
		Conflicts: make([]*shared.SyncConflict, 0),
		Errors:    make([]string, 0),
	}

	// Apply projects
	for _, project := range localData.Projects {
		_, err := s.projectManager.SyncCreateProjectWithTimestamps(ctx, project)
		if err != nil {
			// Try update instead
			updated, updateErr := s.projectManager.UpdateProject(ctx, project.ID, project.Title, project.Description, "sync-push")
			if updateErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to sync project %s: %v", project.ID, err))
				result.Success = false
				continue
			}
			result.Updated++
			result.Processed++
			s.logger.Debug("Updated project from sync",
				zap.String("project_id", project.ID.String()),
				zap.String("title", updated.Title))
		} else {
			result.Created++
			result.Processed++
			s.logger.Debug("Created project from sync",
				zap.String("project_id", project.ID.String()),
				zap.String("title", project.Title))
		}
	}

	// Apply tasks
	for _, task := range localData.Tasks {
		_, err := s.projectManager.SyncCreateTaskWithTimestamps(ctx, task)
		if err != nil {
			// Try update instead
			state := types.TaskState(task.State)
			updated, updateErr := s.projectManager.UpdateTask(ctx, task.ID, task.Title, task.Description, task.Complexity, state, "sync-push")
			if updateErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to sync task %s: %v", task.ID, err))
				result.Success = false
				continue
			}
			result.Updated++
			result.Processed++
			s.logger.Debug("Updated task from sync",
				zap.String("task_id", task.ID.String()),
				zap.String("title", updated.Title))
		} else {
			result.Created++
			result.Processed++
			s.logger.Debug("Created task from sync",
				zap.String("task_id", task.ID.String()),
				zap.String("title", task.Title))
		}
	}

	return result
}

// Helper methods for conversion between types

// convertResultToResponse converts sync.SyncResult to shared.SyncResponse
func (s *syncServiceImpl) convertResultToResponse(requestID uuid.UUID, result *knotsync.SyncResult) *shared.SyncResponse {
	response := &shared.SyncResponse{
		Success:    result.Success,
		RequestID:  requestID,
		Processed:  result.Processed,
		Created:    result.Created,
		Updated:    result.Updated,
		Deleted:    result.Deleted,
		NewVersion: result.SyncedAt.Unix(), // Use timestamp as version
		Timestamp:  result.SyncedAt,
		Duration:   result.Duration,
	}

	// Include errors if any
	if len(result.Errors) > 0 {
		// TODO: Convert errors to appropriate format
	}

	return response
}
