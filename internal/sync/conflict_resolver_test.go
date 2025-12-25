package sync

import (
	"context"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestLogger creates a simple no-op logger for testing
func newTestLogger() logger.Logger {
	return &noopLogger{}
}

// TestConflictResolver_NewConflictResolver tests the constructor
func TestConflictResolver_NewConflictResolver(t *testing.T) {
	mockLogger := newTestLogger()

	resolver := NewConflictResolver(mockLogger, ConflictStrategyLastWriterWins)

	assert.NotNil(t, resolver)
	assert.Equal(t, ConflictStrategyLastWriterWins, resolver.strategy)
}

// TestConflictResolver_IdentifyConflicts_NoConflicts tests operations on different entities
func TestConflictResolver_IdentifyConflicts_NoConflicts(t *testing.T) {
	mockLogger := newTestLogger()
	resolver := NewConflictResolver(mockLogger, ConflictStrategyLastWriterWins)

	now := time.Now()

	operations := []shared.SyncOperation{
		{
			ID:         uuid.New(),
			EntityID:   uuid.New(),
			EntityType: shared.EntityProject,
			Type:       shared.OpCreate,
			Direction:  shared.SyncLocalToMCP,
			CreatedAt:  now,
		},
		{
			ID:         uuid.New(),
			EntityID:   uuid.New(),
			EntityType: shared.EntityProject,
			Type:       shared.OpCreate,
			Direction:  shared.SyncMcpToLocal,
			CreatedAt:  now.Add(time.Minute),
		},
		{
			ID:         uuid.New(),
			EntityID:   uuid.New(),
			EntityType: shared.EntityTask,
			Type:       shared.OpUpdate,
			Direction:  shared.SyncLocalToMCP,
			CreatedAt:  now.Add(2 * time.Minute),
		},
	}

	diffResult := &DiffResult{
		Operations: operations,
	}

	result, err := resolver.ResolveConflicts(context.Background(), diffResult)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Conflicts, "No conflicts expected for operations on different entities")
	assert.Len(t, result.Resolved, 3, "All operations should be resolved")
}

// TestConflictResolver_IdentifyConflicts_SameEntityDifferentTimestamps tests simple timestamp conflicts
func TestConflictResolver_IdentifyConflicts_SameEntityDifferentTimestamps(t *testing.T) {
	mockLogger := newTestLogger()
	resolver := NewConflictResolver(mockLogger, ConflictStrategyLastWriterWins)

	now := time.Now()
	projectID := uuid.New()

	localTime := now
	remoteTime := now.Add(time.Hour)

	localProject := &types.Project{
		ID:        projectID,
		Title:     "Local Project",
		State:     types.ProjectStateActive,
		UpdatedAt: localTime,
	}

	remoteProject := &types.Project{
		ID:        projectID,
		Title:     "Remote Project",
		State:     types.ProjectStateActive,
		UpdatedAt: remoteTime,
	}

	operations := []shared.SyncOperation{
		{
			ID:         uuid.New(),
			EntityID:   projectID,
			EntityType: shared.EntityProject,
			Type:       shared.OpUpdate,
			Direction:  shared.SyncLocalToMCP,
			LocalData:  localProject,
			CreatedAt:  localTime,
		},
		{
			ID:         uuid.New(),
			EntityID:   projectID,
			EntityType: shared.EntityProject,
			Type:       shared.OpUpdate,
			Direction:  shared.SyncMcpToLocal,
			RemoteData: remoteProject,
			CreatedAt:  remoteTime,
		},
	}

	diffResult := &DiffResult{
		Operations: operations,
	}

	result, err := resolver.ResolveConflicts(context.Background(), diffResult)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Conflicts, 1, "One conflict expected")
	assert.Len(t, result.Resolutions, 1, "One resolution expected")
	assert.Equal(t, shared.ConflictTypeUpdate, result.Conflicts[0].ConflictType)

	// LastWriterWins should choose remote (later timestamp)
	assert.NotNil(t, result.Resolutions[0].ChosenData)
	assert.Equal(t, "Remote Project", result.Resolutions[0].ChosenData.(*types.Project).Title)
}

// TestConflictResolver_IdentifyConflicts_SameTimestamp tests complex conflicts with same timestamps
func TestConflictResolver_IdentifyConflicts_SameTimestamp(t *testing.T) {
	mockLogger := newTestLogger()
	resolver := NewConflictResolver(mockLogger, ConflictStrategyLastWriterWins)

	now := time.Now()
	projectID := uuid.New()

	localProject := &types.Project{
		ID:        projectID,
		Title:     "Local Project",
		State:     types.ProjectStateActive,
		UpdatedAt: now,
	}

	remoteProject := &types.Project{
		ID:        projectID,
		Title:     "Remote Project",
		State:     types.ProjectStateActive,
		UpdatedAt: now, // Same timestamp
	}

	operations := []shared.SyncOperation{
		{
			ID:         uuid.New(),
			EntityID:   projectID,
			EntityType: shared.EntityProject,
			Type:       shared.OpUpdate,
			Direction:  shared.SyncLocalToMCP,
			LocalData:  localProject,
			CreatedAt:  now,
		},
		{
			ID:         uuid.New(),
			EntityID:   projectID,
			EntityType: shared.EntityProject,
			Type:       shared.OpUpdate,
			Direction:  shared.SyncMcpToLocal,
			RemoteData: remoteProject,
			CreatedAt:  now,
		},
	}

	diffResult := &DiffResult{
		Operations: operations,
	}

	result, err := resolver.ResolveConflicts(context.Background(), diffResult)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Conflicts, 1, "One conflict expected")

	// With same timestamp, LastWriterWins prefers remote
	assert.NotNil(t, result.Resolutions[0].ChosenData)
	assert.Equal(t, "Remote Project", result.Resolutions[0].ChosenData.(*types.Project).Title)
}

// TestConflictResolver_MultipleConflicts tests resolving multiple conflicts at once
func TestConflictResolver_MultipleConflicts(t *testing.T) {
	mockLogger := newTestLogger()
	resolver := NewConflictResolver(mockLogger, ConflictStrategyLastWriterWins)

	now := time.Now()
	projectID := uuid.New()
	taskID := uuid.New()

	localProject := &types.Project{
		ID:        projectID,
		Title:     "Local Project",
		State:     types.ProjectStateActive,
		UpdatedAt: now.Add(-2 * time.Hour),
	}

	remoteProject := &types.Project{
		ID:        projectID,
		Title:     "Remote Project",
		State:     types.ProjectStateActive,
		UpdatedAt: now.Add(-1 * time.Hour),
	}

	localTask := &types.Task{
		ID:        taskID,
		ProjectID: projectID,
		Title:     "Local Task",
		State:     types.TaskStatePending,
		UpdatedAt: now.Add(-30 * time.Minute),
	}

	remoteTask := &types.Task{
		ID:        taskID,
		ProjectID: projectID,
		Title:     "Remote Task",
		State:     types.TaskStateInProgress,
		UpdatedAt: now,
	}

	operations := []shared.SyncOperation{
		{
			ID:         uuid.New(),
			EntityID:   projectID,
			EntityType: shared.EntityProject,
			Type:       shared.OpUpdate,
			Direction:  shared.SyncLocalToMCP,
			LocalData:  localProject,
			CreatedAt:  now,
		},
		{
			ID:         uuid.New(),
			EntityID:   projectID,
			EntityType: shared.EntityProject,
			Type:       shared.OpUpdate,
			Direction:  shared.SyncMcpToLocal,
			RemoteData: remoteProject,
			CreatedAt:  now,
		},
		{
			ID:         uuid.New(),
			EntityID:   taskID,
			EntityType: shared.EntityTask,
			Type:       shared.OpUpdate,
			Direction:  shared.SyncLocalToMCP,
			LocalData:  localTask,
			CreatedAt:  now,
		},
		{
			ID:         uuid.New(),
			EntityID:   taskID,
			EntityType: shared.EntityTask,
			Type:       shared.OpUpdate,
			Direction:  shared.SyncMcpToLocal,
			RemoteData: remoteTask,
			CreatedAt:  now,
		},
	}

	diffResult := &DiffResult{
		Operations: operations,
	}

	result, err := resolver.ResolveConflicts(context.Background(), diffResult)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Conflicts, 2, "Two conflicts expected")
	assert.Len(t, result.Resolutions, 2, "Two resolutions expected")

	// Both should prefer remote (later timestamps)
	assert.Equal(t, "Remote Project", result.Resolutions[0].ChosenData.(*types.Project).Title)
	assert.Equal(t, "Remote Task", result.Resolutions[1].ChosenData.(*types.Task).Title)
}

// TestConflictResolver_NoOperations tests empty operations list
func TestConflictResolver_NoOperations(t *testing.T) {
	mockLogger := newTestLogger()
	resolver := NewConflictResolver(mockLogger, ConflictStrategyLastWriterWins)

	diffResult := &DiffResult{
		Operations: []shared.SyncOperation{},
	}

	result, err := resolver.ResolveConflicts(context.Background(), diffResult)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Conflicts)
	assert.Empty(t, result.Resolved)
	assert.Empty(t, result.Resolutions)
}

// TestConflictResolver_StrategyPreferLocal tests PreferLocal strategy
func TestConflictResolver_StrategyPreferLocal(t *testing.T) {
	mockLogger := newTestLogger()
	resolver := NewConflictResolver(mockLogger, ConflictStrategyPreferLocal)

	now := time.Now()
	projectID := uuid.New()

	localProject := &types.Project{
		ID:        projectID,
		Title:     "Local Project",
		State:     types.ProjectStateActive,
		UpdatedAt: now.Add(-2 * time.Hour), // Older!
	}

	remoteProject := &types.Project{
		ID:        projectID,
		Title:     "Remote Project",
		State:     types.ProjectStateActive,
		UpdatedAt: now, // Newer
	}

	operations := []shared.SyncOperation{
		{
			ID:         uuid.New(),
			EntityID:   projectID,
			EntityType: shared.EntityProject,
			Type:       shared.OpUpdate,
			Direction:  shared.SyncLocalToMCP,
			LocalData:  localProject,
			CreatedAt:  now,
		},
		{
			ID:         uuid.New(),
			EntityID:   projectID,
			EntityType: shared.EntityProject,
			Type:       shared.OpUpdate,
			Direction:  shared.SyncMcpToLocal,
			RemoteData: remoteProject,
			CreatedAt:  now,
		},
	}

	diffResult := &DiffResult{
		Operations: operations,
	}

	result, err := resolver.ResolveConflicts(context.Background(), diffResult)

	require.NoError(t, err)
	assert.NotNil(t, result)

	// PreferLocal should choose local despite older timestamp
	assert.Equal(t, "Local Project", result.Resolutions[0].ChosenData.(*types.Project).Title)
}

// TestConflictResolver_StrategyPreferRemote tests PreferRemote strategy
func TestConflictResolver_StrategyPreferRemote(t *testing.T) {
	mockLogger := newTestLogger()
	resolver := NewConflictResolver(mockLogger, ConflictStrategyPreferRemote)

	now := time.Now()
	projectID := uuid.New()

	localProject := &types.Project{
		ID:        projectID,
		Title:     "Local Project",
		State:     types.ProjectStateActive,
		UpdatedAt: now, // Newer
	}

	remoteProject := &types.Project{
		ID:        projectID,
		Title:     "Remote Project",
		State:     types.ProjectStateActive,
		UpdatedAt: now.Add(-2 * time.Hour), // Older
	}

	operations := []shared.SyncOperation{
		{
			ID:         uuid.New(),
			EntityID:   projectID,
			EntityType: shared.EntityProject,
			Type:       shared.OpUpdate,
			Direction:  shared.SyncLocalToMCP,
			LocalData:  localProject,
			CreatedAt:  now,
		},
		{
			ID:         uuid.New(),
			EntityID:   projectID,
			EntityType: shared.EntityProject,
			Type:       shared.OpUpdate,
			Direction:  shared.SyncMcpToLocal,
			RemoteData: remoteProject,
			CreatedAt:  now,
		},
	}

	diffResult := &DiffResult{
		Operations: operations,
	}

	result, err := resolver.ResolveConflicts(context.Background(), diffResult)

	require.NoError(t, err)
	assert.NotNil(t, result)

	// PreferRemote should choose remote despite older timestamp
	assert.Equal(t, "Remote Project", result.Resolutions[0].ChosenData.(*types.Project).Title)
}

// TestConflictResolver_TaskDependencyConflict tests dependency-related conflicts
func TestConflictResolver_TaskDependencyConflict(t *testing.T) {
	mockLogger := newTestLogger()
	resolver := NewConflictResolver(mockLogger, ConflictStrategyLastWriterWins)

	now := time.Now()
	taskID := uuid.New()
	depID := uuid.New()

	localTask := &types.Task{
		ID:          taskID,
		ProjectID:   uuid.New(),
		Title:       "Local Task",
		State:       types.TaskStatePending,
		Dependencies: []uuid.UUID{depID},
		UpdatedAt:   now,
	}

	remoteTask := &types.Task{
		ID:          taskID,
		ProjectID:   uuid.New(),
		Title:       "Remote Task",
		State:       types.TaskStatePending,
		Dependencies: []uuid.UUID{}, // No dependencies
		UpdatedAt:   now.Add(time.Hour),
	}

	operations := []shared.SyncOperation{
		{
			ID:         uuid.New(),
			EntityID:   taskID,
			EntityType: shared.EntityTask,
			Type:       shared.OpUpdate,
			Direction:  shared.SyncLocalToMCP,
			LocalData:  localTask,
			CreatedAt:  now,
		},
		{
			ID:         uuid.New(),
			EntityID:   taskID,
			EntityType: shared.EntityTask,
			Type:       shared.OpUpdate,
			Direction:  shared.SyncMcpToLocal,
			RemoteData: remoteTask,
			CreatedAt:  now.Add(time.Hour),
		},
	}

	diffResult := &DiffResult{
		Operations: operations,
	}

	result, err := resolver.ResolveConflicts(context.Background(), diffResult)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Conflicts, 1)

	// Remote wins due to later timestamp
	chosenTask := result.Resolutions[0].ChosenData.(*types.Task)
	assert.Equal(t, "Remote Task", chosenTask.Title)
	assert.Empty(t, chosenTask.Dependencies, "Remote has no dependencies")
}

// TestConflictResolver_CreateOperationsDontConflict tests that create operations don't conflict
func TestConflictResolver_CreateOperationsDontConflict(t *testing.T) {
	mockLogger := newTestLogger()
	resolver := NewConflictResolver(mockLogger, ConflictStrategyLastWriterWins)

	now := time.Now()
	projectID := uuid.New()

	project := &types.Project{
		ID:        projectID,
		Title:     "Test Project",
		State:     types.ProjectStateActive,
		UpdatedAt: now,
	}

	operations := []shared.SyncOperation{
		{
			ID:         uuid.New(),
			EntityID:   projectID,
			EntityType: shared.EntityProject,
			Type:       shared.OpCreate,
			Direction:  shared.SyncLocalToMCP,
			LocalData:  project,
			CreatedAt:  now,
		},
		{
			ID:         uuid.New(),
			EntityID:   projectID,
			EntityType: shared.EntityProject,
			Type:       shared.OpCreate,
			Direction:  shared.SyncMcpToLocal,
			RemoteData: project,
			CreatedAt:  now.Add(time.Minute),
		},
	}

	diffResult := &DiffResult{
		Operations: operations,
	}

	result, err := resolver.ResolveConflicts(context.Background(), diffResult)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Conflicts, "Create operations shouldn't conflict")
	assert.Len(t, result.Resolved, 2, "Both create operations should be resolved")
}

// TestConflictResolver_DeleteVsUpdateConflict tests delete vs update conflicts
func TestConflictResolver_DeleteVsUpdateConflict(t *testing.T) {
	mockLogger := newTestLogger()
	resolver := NewConflictResolver(mockLogger, ConflictStrategyPreferLocal)

	now := time.Now()
	taskID := uuid.New()

	task := &types.Task{
		ID:        taskID,
		ProjectID: uuid.New(),
		Title:     "Test Task",
		State:     types.TaskStatePending,
		UpdatedAt: now,
	}

	operations := []shared.SyncOperation{
		{
			ID:         uuid.New(),
			EntityID:   taskID,
			EntityType: shared.EntityTask,
			Type:       shared.OpUpdate,
			Direction:  shared.SyncLocalToMCP,
			LocalData:  task,
			CreatedAt:  now,
		},
		{
			ID:         uuid.New(),
			EntityID:   taskID,
			EntityType: shared.EntityTask,
			Type:       shared.OpDelete,
			Direction:  shared.SyncMcpToLocal,
			RemoteData: task,
			CreatedAt:  now,
		},
	}

	diffResult := &DiffResult{
		Operations: operations,
	}

	result, err := resolver.ResolveConflicts(context.Background(), diffResult)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Conflicts, 1, "Delete vs update should conflict")
	assert.Equal(t, shared.ConflictTypeDelete, result.Conflicts[0].ConflictType)

	// PreferLocal should choose update over delete
	assert.Equal(t, "Test Task", result.Resolutions[0].ChosenData.(*types.Task).Title)
}
