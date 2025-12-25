package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"go.uber.org/zap"
)

// DiffEngine defines the interface for timestamp-based diff calculation for sync operations
type DiffEngine interface {
	// CalculateDiff computes the differences between local and remote data
	CalculateDiff(ctx context.Context, localData, remoteData *shared.SyncDataSet) (*DiffResult, error)

	// ValidateDiff validates diff results for consistency
	ValidateDiff(ctx context.Context, result *DiffResult) error
}

// diffEngineImpl is the private implementation of DiffEngine
type diffEngineImpl struct {
	logger logger.Logger
}

// Ensure diffEngineImpl implements DiffEngine
var _ DiffEngine = (*diffEngineImpl)(nil)

// NewDiffEngineService creates a new diff engine service for DI
func NewDiffEngineService(injector do.Injector) (DiffEngine, error) {
	logger := do.MustInvoke[logger.Logger](injector)

	logger.Debug("startup diff engine service")
	return &diffEngineImpl{
		logger: logger,
	}, nil
}

// NewDiffEngine creates a new diff engine (deprecated - use DI instead)
// Kept for backward compatibility with tests
func NewDiffEngine(logger logger.Logger) DiffEngine {
	return &diffEngineImpl{
		logger: logger,
	}
}

// CalculateDiff computes the differences between local and remote data
func (e *diffEngineImpl) CalculateDiff(ctx context.Context, localData, remoteData *shared.SyncDataSet) (*DiffResult, error) {
	e.logger.Info("Calculating sync diff",
		zap.Int("local_projects", len(localData.Projects)),
		zap.Int("remote_projects", len(remoteData.Projects)),
		zap.Int("local_tasks", len(localData.Tasks)),
		zap.Int("remote_tasks", len(remoteData.Tasks)))

	startTime := time.Now()

	result := &DiffResult{
		LocalTimestamp:  time.Now(),
		RemoteTimestamp: time.Now(),
		Operations:      []shared.SyncOperation{},
	}

	// Calculate project differences
	projectDiff, err := e.calculateProjectDiff(ctx, localData.Projects, remoteData.Projects)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate project diff: %w", err)
	}
	result.Operations = append(result.Operations, projectDiff...)

	// Calculate task differences
	taskDiff, err := e.calculateTaskDiff(ctx, localData.Tasks, remoteData.Tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate task diff: %w", err)
	}
	result.Operations = append(result.Operations, taskDiff...)

	result.Duration = time.Since(startTime)
	result.Summary = e.generateSummary(result.Operations)

	e.logger.Info("Diff calculation completed",
		zap.Duration("duration", result.Duration),
		zap.Int("operations", len(result.Operations)),
		zap.Int("creations", result.Summary.Creations),
		zap.Int("updates", result.Summary.Updates),
		zap.Int("deletes", result.Summary.Deletes),
		zap.Int("conflicts", result.Summary.Conflicts))

	return result, nil
}

// calculateProjectDiff computes differences between local and remote projects
func (e *diffEngineImpl) calculateProjectDiff(ctx context.Context, localProjects, remoteProjects map[uuid.UUID]*types.Project) ([]shared.SyncOperation, error) {
	var operations []shared.SyncOperation

	e.logger.Debug("Calculating project diff",
		zap.Int("local_count", len(localProjects)),
		zap.Int("remote_count", len(remoteProjects)))

	// Find projects that exist locally but not remotely (need to create remotely)
	for id, localProject := range localProjects {
		if remoteProject, exists := remoteProjects[id]; !exists {
			// Local project doesn't exist remotely - create remotely
			operations = append(operations, shared.SyncOperation{
				ID:         uuid.New(),
				Type:       shared.OpCreate,
				EntityType: shared.EntityProject,
				EntityID:   localProject.ID,
				LocalData:  localProject,
				RemoteData: nil,
				Direction:  shared.SyncLocalToMCP,
				Reason:     "Local project not found remotely",
				Priority:   e.calculatePriority(localProject.UpdatedAt, time.Time{}),
				CreatedAt:  time.Now(),
			})
			e.logger.Debug("Project needs remote creation",
				zap.String("project_id", id.String()),
				zap.String("title", localProject.Title))
		} else {
			// Project exists on both sides - check for updates
			op := e.compareProjectTimestamps(localProject, remoteProject)
			if op != nil {
				operations = append(operations, *op)
			}
		}
	}

	// Find projects that exist remotely but not locally (need to create locally)
	for id, remoteProject := range remoteProjects {
		if _, exists := localProjects[id]; !exists {
			// Remote project doesn't exist locally - create locally
			operations = append(operations, shared.SyncOperation{
				ID:         uuid.New(),
				Type:       shared.OpCreate,
				EntityType: shared.EntityProject,
				EntityID:   remoteProject.ID,
				LocalData:  nil,
				RemoteData: remoteProject,
				Direction:  shared.SyncMcpToLocal,
				Reason:     "Remote project not found locally",
				Priority:   e.calculatePriority(time.Time{}, remoteProject.UpdatedAt),
				CreatedAt:  time.Now(),
			})
			e.logger.Debug("Project needs local creation",
				zap.String("project_id", id.String()),
				zap.String("title", remoteProject.Title))
		}
	}

	return operations, nil
}

// calculateTaskDiff computes differences between local and remote tasks
func (e *diffEngineImpl) calculateTaskDiff(ctx context.Context, localTasks, remoteTasks map[uuid.UUID]*types.Task) ([]shared.SyncOperation, error) {
	var operations []shared.SyncOperation

	e.logger.Debug("Calculating task diff",
		zap.Int("local_count", len(localTasks)),
		zap.Int("remote_count", len(remoteTasks)))

	// Find tasks that exist locally but not remotely
	for id, localTask := range localTasks {
		if remoteTask, exists := remoteTasks[id]; !exists {
			// Local task doesn't exist remotely - create remotely
			operations = append(operations, shared.SyncOperation{
				ID:         uuid.New(),
				Type:       shared.OpCreate,
				EntityType: shared.EntityTask,
				EntityID:   localTask.ID,
				LocalData:  localTask,
				RemoteData: nil,
				Direction:  shared.SyncLocalToMCP,
				Reason:     "Local task not found remotely",
				Priority:   e.calculatePriority(localTask.UpdatedAt, time.Time{}),
				CreatedAt:  time.Now(),
			})
			e.logger.Debug("Task needs remote creation",
				zap.String("task_id", id.String()),
				zap.String("title", localTask.Title))
		} else {
			// Task exists on both sides - check for updates
			op := e.compareTaskTimestamps(localTask, remoteTask)
			if op != nil {
				operations = append(operations, *op)
			}
		}
	}

	// Find tasks that exist remotely but not locally
	for id, remoteTask := range remoteTasks {
		if _, exists := localTasks[id]; !exists {
			// Remote task doesn't exist locally - create locally
			operations = append(operations, shared.SyncOperation{
				ID:         uuid.New(),
				Type:       shared.OpCreate,
				EntityType: shared.EntityTask,
				EntityID:   remoteTask.ID,
				LocalData:  nil,
				RemoteData: remoteTask,
				Direction:  shared.SyncMcpToLocal,
				Reason:     "Remote task not found locally",
				Priority:   e.calculatePriority(time.Time{}, remoteTask.UpdatedAt),
				CreatedAt:  time.Now(),
			})
			e.logger.Debug("Task needs local creation",
				zap.String("task_id", id.String()),
				zap.String("title", remoteTask.Title))
		}
	}

	return operations, nil
}

// compareProjectTimestamps compares local and remote project timestamps
func (e *diffEngineImpl) compareProjectTimestamps(local, remote *types.Project) *shared.SyncOperation {
	if local.UpdatedAt.Equal(remote.UpdatedAt) {
		return nil // No changes needed
	}

	if local.UpdatedAt.After(remote.UpdatedAt) {
		// Local is newer - update remote
		return &shared.SyncOperation{
			ID:         uuid.New(),
			Type:       shared.OpUpdate,
			EntityType: shared.EntityProject,
			EntityID:   local.ID,
			LocalData:  local,
			RemoteData: remote,
			Direction:  shared.SyncLocalToMCP,
			Reason:     fmt.Sprintf("Local project updated more recently (%v > %v)", local.UpdatedAt, remote.UpdatedAt),
			Priority:   e.calculatePriority(local.UpdatedAt, remote.UpdatedAt),
			CreatedAt:  time.Now(),
		}
	} else {
		// Remote is newer - update local
		return &shared.SyncOperation{
			ID:         uuid.New(),
			Type:       shared.OpUpdate,
			EntityType: shared.EntityProject,
			EntityID:   remote.ID,
			LocalData:  local,
			RemoteData: remote,
			Direction:  shared.SyncMcpToLocal,
			Reason:     fmt.Sprintf("Remote project updated more recently (%v > %v)", remote.UpdatedAt, local.UpdatedAt),
			Priority:   e.calculatePriority(local.UpdatedAt, remote.UpdatedAt),
			CreatedAt:  time.Now(),
		}
	}
}

// compareTaskTimestamps compares local and remote task timestamps
func (e *diffEngineImpl) compareTaskTimestamps(local, remote *types.Task) *shared.SyncOperation {
	if local.UpdatedAt.Equal(remote.UpdatedAt) {
		return nil // No changes needed
	}

	if local.UpdatedAt.After(remote.UpdatedAt) {
		// Local is newer - update remote
		return &shared.SyncOperation{
			ID:         uuid.New(),
			Type:       shared.OpUpdate,
			EntityType: shared.EntityTask,
			EntityID:   local.ID,
			LocalData:  local,
			RemoteData: remote,
			Direction:  shared.SyncLocalToMCP,
			Reason:     fmt.Sprintf("Local task updated more recently (%v > %v)", local.UpdatedAt, remote.UpdatedAt),
			Priority:   e.calculatePriority(local.UpdatedAt, remote.UpdatedAt),
			CreatedAt:  time.Now(),
		}
	} else {
		// Remote is newer - update local
		return &shared.SyncOperation{
			ID:         uuid.New(),
			Type:       shared.OpUpdate,
			EntityType: shared.EntityTask,
			EntityID:   remote.ID,
			LocalData:  local,
			RemoteData: remote,
			Direction:  shared.SyncMcpToLocal,
			Reason:     fmt.Sprintf("Remote task updated more recently (%v > %v)", remote.UpdatedAt, local.UpdatedAt),
			Priority:   e.calculatePriority(local.UpdatedAt, remote.UpdatedAt),
			CreatedAt:  time.Now(),
		}
	}
}

// calculatePriority calculates operation priority based on timestamp differences
func (e *diffEngineImpl) calculatePriority(localTime, remoteTime time.Time) int {
	if localTime.IsZero() || remoteTime.IsZero() {
		return 10 // High priority for new entities
	}

	diff := localTime.Sub(remoteTime)
	if diff < 0 {
		diff = -diff
	}

	// Higher priority for more recent changes
	switch {
	case diff < time.Hour:
		return 10
	case diff < 24*time.Hour:
		return 8
	case diff < 7*24*time.Hour:
		return 6
	case diff < 30*24*time.Hour:
		return 4
	default:
		return 2
	}
}

// generateSummary creates a summary of sync operations
func (e *diffEngineImpl) generateSummary(operations []shared.SyncOperation) SyncSummary {
	summary := SyncSummary{
		Total:        len(operations),
		Creations:    0,
		Updates:      0,
		Deletes:      0,
		Conflicts:    0,
		HighPriority: 0,
	}

	for _, op := range operations {
		switch op.Type {
		case shared.OpCreate:
			summary.Creations++
		case shared.OpUpdate:
			summary.Updates++
		case shared.OpDelete:
			summary.Deletes++
		case shared.OpConflict:
			summary.Conflicts++
		}

		if op.Priority >= 8 {
			summary.HighPriority++
		}
	}

	return summary
}

// ValidateDiff validates diff results for consistency
func (e *diffEngineImpl) ValidateDiff(ctx context.Context, result *DiffResult) error {
	e.logger.Debug("Validating diff result",
		zap.Int("operations", len(result.Operations)))

	// Check for circular dependencies
	visited := make(map[uuid.UUID]bool)
	for _, op := range result.Operations {
		if visited[op.ID] {
			return fmt.Errorf("duplicate operation ID: %s", op.ID)
		}
		visited[op.ID] = true
	}

	// Validate operation data
	for _, op := range result.Operations {
		if err := e.validateOperation(&op); err != nil {
			return fmt.Errorf("invalid operation %s: %w", op.ID, err)
		}
	}

	e.logger.Debug("Diff validation completed successfully")
	return nil
}

// validateOperation validates a single sync operation
func (e *diffEngineImpl) validateOperation(op *shared.SyncOperation) error {
	if op.ID == uuid.Nil {
		return fmt.Errorf("operation ID is required")
	}

	if op.EntityType == "" {
		return fmt.Errorf("entity type is required")
	}

	if op.Direction == "" {
		return fmt.Errorf("direction is required")
	}

	if op.Reason == "" {
		return fmt.Errorf("reason is required")
	}

	// Validate that we have appropriate data for the operation type
	switch op.Type {
	case shared.OpCreate, shared.OpUpdate:
		if op.Direction == shared.SyncLocalToMCP && op.LocalData == nil {
			return fmt.Errorf("local data is required for local-to-MCP operations")
		}
		if op.Direction == shared.SyncMcpToLocal && op.RemoteData == nil {
			return fmt.Errorf("remote data is required for MCP-to-local operations")
		}
	case shared.OpDelete:
		// Delete operations should have at least one side's data for reference
		if op.LocalData == nil && op.RemoteData == nil {
			return fmt.Errorf("at least one side's data is required for delete operations")
		}
	}

	return nil
}
