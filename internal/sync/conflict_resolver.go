package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ConflictResolver handles synchronization conflicts and provides resolution strategies
type ConflictResolver struct {
	logger   logger.Logger
	strategy ConflictStrategy
}

// NewConflictResolver creates a new conflict resolver
func NewConflictResolver(logger logger.Logger, strategy ConflictStrategy) *ConflictResolver {
	return &ConflictResolver{
		logger:   logger,
		strategy: strategy,
	}
}

// ResolveConflicts processes and resolves conflicts in the diff result
func (r *ConflictResolver) ResolveConflicts(ctx context.Context, diffResult *DiffResult) (*ConflictResolutionResult, error) {
	r.logger.Info("Resolving sync conflicts",
		zap.Int("operations", len(diffResult.Operations)),
		zap.String("strategy", string(r.strategy)))

	startTime := time.Now()

	result := &ConflictResolutionResult{
		Strategy:    r.strategy,
		Resolved:    make([]shared.SyncOperation, 0),
		Conflicts:   make([]*shared.SyncConflict, 0),
		Unresolved:  make([]shared.SyncOperation, 0),
		Resolutions: make([]shared.ConflictResolution, 0),
	}

	// Identify conflicts
	conflicts := r.identifyConflicts(diffResult.Operations)
	result.Conflicts = conflicts

	r.logger.Info("Identified conflicts",
		zap.Int("total_conflicts", len(conflicts)))

	// Resolve conflicts based on strategy
	for _, conflict := range conflicts {
		resolution, err := r.resolveConflict(ctx, conflict)
		if err != nil {
			r.logger.Error("Failed to resolve conflict",
				zap.String("conflict_id", conflict.ID.String()),
				zap.Error(err))
			result.Unresolved = append(result.Unresolved, conflict.Operations...)
			continue
		}

		result.Resolutions = append(result.Resolutions, *resolution)
		conflict.Resolution = resolution
		conflict.ResolvedAt = &resolution.Timestamp

		// Add resolved operations to the result
		result.Resolved = append(result.Resolved, conflict.Operations...)
	}

	// Add non-conflicting operations to result
	for _, op := range diffResult.Operations {
		if !r.isConflictedOperation(op, conflicts) {
			result.Resolved = append(result.Resolved, op)
		}
	}

	result.Duration = time.Since(startTime)

	r.logger.Info("Conflict resolution completed",
		zap.Duration("duration", result.Duration),
		zap.Int("resolved_ops", len(result.Resolved)),
		zap.Int("unresolved_ops", len(result.Unresolved)),
		zap.Int("conflicts_resolved", len(result.Resolutions)))

	return result, nil
}

// identifyConflicts identifies conflicts in the sync operations
func (r *ConflictResolver) identifyConflicts(operations []shared.SyncOperation) []*shared.SyncConflict {
	conflicts := make([]*shared.SyncConflict, 0)
	operationMap := make(map[string][]*shared.SyncOperation)

	// Group operations by entity
	for i := range operations {
		op := &operations[i]
		key := fmt.Sprintf("%s:%s", op.EntityType, op.EntityID)
		operationMap[key] = append(operationMap[key], op)
	}

	// Find entities with operations on both sides
	for _, ops := range operationMap {
		if len(ops) > 1 {
			conflict := r.createConflictFromOperations(ops)
			if conflict != nil {
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return conflicts
}

// createConflictFromOperations creates a conflict from multiple operations on the same entity
func (r *ConflictResolver) createConflictFromOperations(operations []*shared.SyncOperation) *shared.SyncConflict {
	if len(operations) < 2 {
		return nil
	}

	// Find the most recent local and remote operations
	var localOp, remoteOp *shared.SyncOperation
	var latestLocalTime, latestRemoteTime time.Time

	for _, op := range operations {
		if op.Direction == shared.SyncLocalToMCP {
			if op.CreatedAt.After(latestLocalTime) {
				localOp = op
				latestLocalTime = op.CreatedAt
			}
		} else if op.Direction == shared.SyncMcpToLocal {
			if op.CreatedAt.After(latestRemoteTime) {
				remoteOp = op
				latestRemoteTime = op.CreatedAt
			}
		}
	}

	// Determine conflict type
	conflictType := r.determineConflictType(localOp, remoteOp)
	if conflictType == "" {
		return nil // No actual conflict
	}

	entityType := shared.EntityProject
	if localOp != nil {
		entityType = localOp.EntityType
	} else if remoteOp != nil {
		entityType = remoteOp.EntityType
	}

	operationSlice := make([]shared.SyncOperation, len(operations))
	for i, op := range operations {
		operationSlice[i] = *op
	}

	return &shared.SyncConflict{
		ID:           uuid.New(),
		EntityID:     operations[0].EntityID,
		EntityType:   entityType,
		LocalData:    r.extractDataFromOperation(localOp),
		RemoteData:   r.extractDataFromOperation(remoteOp),
		ConflictType: conflictType,
		Operations:   operationSlice,
		CreatedAt:    time.Now(),
	}
}

// determineConflictType determines the type of conflict
func (r *ConflictResolver) determineConflictType(localOp, remoteOp *shared.SyncOperation) shared.ConflictType {
	if localOp == nil || remoteOp == nil {
		return "" // No conflict if only one side has operation
	}

	// Both sides have updates
	if localOp.Type == shared.OpUpdate && remoteOp.Type == shared.OpUpdate {
		return shared.ConflictTypeUpdate
	}

	// One side wants to delete, other wants to update
	if (localOp.Type == shared.OpDelete && remoteOp.Type == shared.OpUpdate) ||
		(localOp.Type == shared.OpUpdate && remoteOp.Type == shared.OpDelete) {
		return shared.ConflictTypeDelete
	}

	// Different states
	if localOp.Type != remoteOp.Type {
		return shared.ConflictTypeState
	}

	return "" // No conflict
}

// extractDataFromOperation extracts entity data from an operation
func (r *ConflictResolver) extractDataFromOperation(op *shared.SyncOperation) interface{} {
	if op == nil {
		return nil
	}

	// Prefer the data that's being synced TO the other side
	switch op.Direction {
	case shared.SyncLocalToMCP:
		return op.LocalData
	case shared.SyncMcpToLocal:
		return op.RemoteData
	default:
		// Fall back to local data
		return op.LocalData
	}
}

// resolveConflict resolves a single conflict based on the configured strategy
func (r *ConflictResolver) resolveConflict(ctx context.Context, conflict *shared.SyncConflict) (*shared.ConflictResolution, error) {
	r.logger.Debug("Resolving conflict",
		zap.String("conflict_id", conflict.ID.String()),
		zap.Any("entity_id", conflict.EntityID),
		zap.String("conflict_type", string(conflict.ConflictType)),
		zap.String("strategy", string(r.strategy)))

	switch r.strategy {
	case ConflictStrategyLastWriterWins:
		return r.resolveLastWriterWins(ctx, conflict)
	case ConflictStrategyPreferLocal:
		return r.resolvePreferLocal(ctx, conflict)
	case ConflictStrategyPreferRemote:
		return r.resolvePreferRemote(ctx, conflict)
	case ConflictStrategyManual:
		return r.resolveManual(ctx, conflict)
	default:
		return nil, fmt.Errorf("unknown conflict resolution strategy: %s", r.strategy)
	}
}

// resolveLastWriterWins resolves conflict by choosing the most recently updated entity
func (r *ConflictResolver) resolveLastWriterWins(ctx context.Context, conflict *shared.SyncConflict) (*shared.ConflictResolution, error) {
	localTime := r.getEntityTimestamp(conflict.LocalData)
	remoteTime := r.getEntityTimestamp(conflict.RemoteData)

	var chosenData interface{}

	if localTime.After(remoteTime) {
		chosenData = conflict.LocalData
	} else {
		chosenData = conflict.RemoteData
	}

	return &shared.ConflictResolution{
		Strategy:   string(ConflictStrategyLastWriterWins),
		ResolvedBy: "system",
		ResolvedAt: time.Now(),
		Timestamp:  time.Now(),
		Actor:      "system",
		ChosenData: chosenData,
	}, nil
}

// resolvePreferLocal resolves conflict by always choosing the local version
func (r *ConflictResolver) resolvePreferLocal(ctx context.Context, conflict *shared.SyncConflict) (*shared.ConflictResolution, error) {
	return &shared.ConflictResolution{
		Strategy:   string(ConflictStrategyPreferLocal),
		ResolvedBy: "system",
		ResolvedAt: time.Now(),
		Timestamp:  time.Now(),
		Actor:      "system",
		ChosenData: conflict.LocalData,
	}, nil
}

// resolvePreferRemote resolves conflict by always choosing the remote version
func (r *ConflictResolver) resolvePreferRemote(ctx context.Context, conflict *shared.SyncConflict) (*shared.ConflictResolution, error) {
	return &shared.ConflictResolution{
		Strategy:   string(ConflictStrategyPreferRemote),
		ResolvedBy: "system",
		ResolvedAt: time.Now(),
		Timestamp:  time.Now(),
		Actor:      "system",
		ChosenData: conflict.RemoteData,
	}, nil
}

// resolveManual marks conflict for manual resolution
func (r *ConflictResolver) resolveManual(ctx context.Context, conflict *shared.SyncConflict) (*shared.ConflictResolution, error) {
	return &shared.ConflictResolution{
		Strategy:   string(ConflictStrategyManual),
		ResolvedBy: "system",
		ResolvedAt: time.Now(),
		Timestamp:  time.Now(),
		Actor:      "system",
	}, nil
}

// resolveMerge attempts to merge conflicting entities
func (r *ConflictResolver) resolveMerge(ctx context.Context, conflict *shared.SyncConflict) (*shared.ConflictResolution, error) {
	mergedData, err := r.mergeEntities(conflict.LocalData, conflict.RemoteData)
	if err != nil {
		r.logger.Warn("Failed to merge entities, falling back to last writer wins",
			zap.String("conflict_id", conflict.ID.String()),
			zap.Error(err))
		return r.resolveLastWriterWins(ctx, conflict)
	}

	return &shared.ConflictResolution{
		Strategy:   "merge",
		ResolvedBy: "system",
		ResolvedAt: time.Now(),
		Timestamp:  time.Now(),
		Actor:      "system",
		MergedData: mergedData,
	}, nil
}

// mergeEntities attempts to merge two entities
func (r *ConflictResolver) mergeEntities(localData, remoteData interface{}) (interface{}, error) {
	// This is a simplified merge implementation
	// In a real implementation, you might want more sophisticated merging logic

	localProject, localIsProject := localData.(*types.Project)
	remoteProject, remoteIsProject := remoteData.(*types.Project)

	if localIsProject && remoteIsProject {
		return r.mergeProjects(localProject, remoteProject), nil
	}

	localTask, localIsTask := localData.(*types.Task)
	remoteTask, remoteIsTask := remoteData.(*types.Task)

	if localIsTask && remoteIsTask {
		return r.mergeTasks(localTask, remoteTask), nil
	}

	// Fallback: return the more recently updated entity
	localTime := r.getEntityTimestamp(localData)
	remoteTime := r.getEntityTimestamp(remoteData)

	if localTime.After(remoteTime) {
		return localData, nil
	}
	return remoteData, nil
}

// mergeProjects merges two projects
func (r *ConflictResolver) mergeProjects(local, remote *types.Project) *types.Project {
	if local == nil {
		return remote
	}
	if remote == nil {
		return local
	}

	// Choose the most recently updated version as base
	var base, other *types.Project
	if local.UpdatedAt.After(remote.UpdatedAt) {
		base, other = local, remote
	} else {
		base, other = remote, local
	}

	// Create merged project
	merged := *base

	// Merge descriptions (prefer non-empty)
	if other.Description != "" && (merged.Description == "" || len(other.Description) > len(merged.Description)) {
		merged.Description = other.Description
	}

	// Use most recent state
	if other.UpdatedAt.After(merged.UpdatedAt) {
		merged.State = other.State
	}

	merged.UpdatedAt = time.Now()

	return &merged
}

// mergeTasks merges two tasks
func (r *ConflictResolver) mergeTasks(local, remote *types.Task) *types.Task {
	if local == nil {
		return remote
	}
	if remote == nil {
		return local
	}

	// Choose the most recently updated version as base
	var base, other *types.Task
	if local.UpdatedAt.After(remote.UpdatedAt) {
		base, other = local, remote
	} else {
		base, other = remote, local
	}

	// Create merged task
	merged := *base

	// Merge descriptions (prefer non-empty)
	if other.Description != "" && (merged.Description == "" || len(other.Description) > len(merged.Description)) {
		merged.Description = other.Description
	}

	// Use higher priority
	if other.Priority > merged.Priority {
		merged.Priority = other.Priority
	}

	// Use most recent state
	if other.UpdatedAt.After(merged.UpdatedAt) {
		merged.State = other.State
	}

	merged.UpdatedAt = time.Now()

	return &merged
}

// mergeStringSlices merges two string slices without duplicates
func (r *ConflictResolver) mergeStringSlices(slice1, slice2 []string) []string {
	merged := make([]string, 0)
	seen := make(map[string]bool)

	for _, s := range slice1 {
		if !seen[s] {
			merged = append(merged, s)
			seen[s] = true
		}
	}

	for _, s := range slice2 {
		if !seen[s] {
			merged = append(merged, s)
			seen[s] = true
		}
	}

	return merged
}

// mergeStringMaps merges two string maps
func (r *ConflictResolver) mergeStringMaps(map1, map2 map[string]string) map[string]string {
	merged := make(map[string]string)

	for k, v := range map1 {
		merged[k] = v
	}

	for k, v := range map2 {
		merged[k] = v
	}

	return merged
}

// getEntityTimestamp extracts the updated timestamp from an entity
func (r *ConflictResolver) getEntityTimestamp(data interface{}) time.Time {
	if data == nil {
		return time.Time{}
	}

	switch entity := data.(type) {
	case *types.Project:
		return entity.UpdatedAt
	case *types.Task:
		return entity.UpdatedAt
	default:
		return time.Time{}
	}
}

// isConflictedOperation checks if an operation is part of a conflict
func (r *ConflictResolver) isConflictedOperation(op shared.SyncOperation, conflicts []*shared.SyncConflict) bool {
	for _, conflict := range conflicts {
		if conflict.EntityID == op.EntityID && conflict.EntityType == op.EntityType {
			return true
		}
	}
	return false
}

// ConflictResolutionResult represents the result of conflict resolution
type ConflictResolutionResult struct {
	Strategy    ConflictStrategy            `json:"strategy"`
	Resolved    []shared.SyncOperation      `json:"resolved_operations"`
	Conflicts   []*shared.SyncConflict      `json:"conflicts"`
	Unresolved  []shared.SyncOperation      `json:"unresolved_operations"`
	Resolutions []shared.ConflictResolution `json:"resolutions"`
	Duration    time.Duration               `json:"duration"`
}

// ValidateResolutions validates conflict resolutions for consistency
func (r *ConflictResolver) ValidateResolutions(ctx context.Context, result *ConflictResolutionResult) error {
	r.logger.Debug("Validating conflict resolutions",
		zap.Int("conflicts", len(result.Conflicts)),
		zap.Int("resolutions", len(result.Resolutions)))

	// Check that all conflicts have resolutions
	for _, conflict := range result.Conflicts {
		if conflict.Resolution == nil {
			return fmt.Errorf("conflict %s has no resolution", conflict.ID)
		}
	}

	// Validate resolution data
	for _, resolution := range result.Resolutions {
		if err := r.validateResolution(&resolution); err != nil {
			return fmt.Errorf("invalid resolution for conflict: %w", err)
		}
	}

	r.logger.Debug("Conflict resolution validation completed successfully")
	return nil
}

// validateResolution validates a single conflict resolution
func (r *ConflictResolver) validateResolution(resolution *shared.ConflictResolution) error {
	if resolution.Timestamp.IsZero() {
		return fmt.Errorf("resolution timestamp is required")
	}

	if resolution.Actor == "" {
		return fmt.Errorf("resolution actor is required")
	}

	if resolution.Strategy == string(ConflictStrategyManual) && resolution.MergedData != nil {
		return fmt.Errorf("manual resolution should not have merged data")
	}

	if resolution.Strategy == "merge" && resolution.MergedData == nil {
		return fmt.Errorf("merge resolution must have merged data")
	}

	return nil
}
