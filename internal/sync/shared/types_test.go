// Package shared provides tests for common types used across sync subpackages
package shared

import (
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestSyncDirection_Constants tests sync direction constants
func TestSyncDirection_Constants(t *testing.T) {
	assert.Equal(t, SyncDirection("local_to_mcp"), SyncLocalToMCP)
	assert.Equal(t, SyncDirection("mcp_to_local"), SyncMcpToLocal)
	assert.Equal(t, SyncDirection("bidirectional"), SyncBidirectional)
}

// TestSyncOpType_Constants tests sync operation type constants
func TestSyncOpType_Constants(t *testing.T) {
	assert.Equal(t, SyncOpType("create"), OpCreate)
	assert.Equal(t, SyncOpType("update"), OpUpdate)
	assert.Equal(t, SyncOpType("delete"), OpDelete)
	assert.Equal(t, SyncOpType("conflict"), OpConflict)
}

// TestSyncEntityType_Constants tests entity type constants
func TestSyncEntityType_Constants(t *testing.T) {
	assert.Equal(t, SyncEntityType("project"), EntityProject)
	assert.Equal(t, SyncEntityType("task"), EntityTask)
}

// TestSyncOpStatus_Constants tests operation status constants
func TestSyncOpStatus_Constants(t *testing.T) {
	assert.Equal(t, SyncOpStatus("pending"), StatusPending)
	assert.Equal(t, SyncOpStatus("in_progress"), StatusInProgress)
	assert.Equal(t, SyncOpStatus("completed"), StatusCompleted)
	assert.Equal(t, SyncOpStatus("failed"), StatusFailed)
	assert.Equal(t, SyncOpStatus("skipped"), StatusSkipped)
}

// TestConflictType_Constants tests conflict type constants
func TestConflictType_Constants(t *testing.T) {
	assert.Equal(t, ConflictType("update"), ConflictTypeUpdate)
	assert.Equal(t, ConflictType("delete"), ConflictTypeDelete)
	assert.Equal(t, ConflictType("state"), ConflictTypeState)
	assert.Equal(t, ConflictType("dependency"), ConflictTypeDependency)
}

// TestSyncDataSet_NewDataSet tests creating a new SyncDataSet
func TestSyncDataSet_NewDataSet(t *testing.T) {
	dataSet := NewSyncDataSet()

	assert.NotNil(t, dataSet)
	assert.NotNil(t, dataSet.Projects)
	assert.NotNil(t, dataSet.Tasks)
	assert.Empty(t, dataSet.Projects)
	assert.Empty(t, dataSet.Tasks)
}

// TestSyncDataSet_WithProjectsAndTasks tests SyncDataSet with data
func TestSyncDataSet_WithProjectsAndTasks(t *testing.T) {
	now := time.Now()
	projectID := uuid.New()
	taskID := uuid.New()

	dataSet := &SyncDataSet{
		Projects: map[uuid.UUID]*types.Project{
			projectID: {
				ID:        projectID,
				Title:     "Test Project",
				State:     types.ProjectStateActive,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		Tasks: map[uuid.UUID]*types.Task{
			taskID: {
				ID:        taskID,
				ProjectID: projectID,
				Title:     "Test Task",
				State:     types.TaskStatePending,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	assert.Len(t, dataSet.Projects, 1)
	assert.Len(t, dataSet.Tasks, 1)
	assert.Equal(t, "Test Project", dataSet.Projects[projectID].Title)
	assert.Equal(t, "Test Task", dataSet.Tasks[taskID].Title)
}

// TestSyncConflict_WithResolution tests SyncConflict with resolution
func TestSyncConflict_WithResolution(t *testing.T) {
	now := time.Now()
	resolvedAt := now.Add(time.Hour)
	project := &types.Project{
		ID:        uuid.New(),
		Title:     "Local Project",
		State:     types.ProjectStateActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	conflict := SyncConflict{
		ID:           uuid.New(),
		EntityID:     uuid.New(),
		EntityType:   EntityProject,
		LocalData:    project,
		RemoteData:   project,
		ConflictType: ConflictTypeUpdate,
		Resolution: &ConflictResolution{
			Strategy:   "prefer-local",
			ResolvedBy: "system",
			ResolvedAt: resolvedAt,
			Reason:     "Local was more recent",
		},
		CreatedAt:  now,
		ResolvedAt: &resolvedAt,
	}

	assert.NotNil(t, conflict.Resolution)
	assert.Equal(t, "prefer-local", conflict.Resolution.Strategy)
	assert.Equal(t, "system", conflict.Resolution.ResolvedBy)
	assert.Equal(t, resolvedAt, conflict.Resolution.ResolvedAt)
	assert.NotNil(t, conflict.ResolvedAt)
}

// TestSyncConflict_WithoutResolution tests SyncConflict without resolution
func TestSyncConflict_WithoutResolution(t *testing.T) {
	now := time.Now()
	conflict := SyncConflict{
		ID:           uuid.New(),
		EntityID:     uuid.New(),
		EntityType:   EntityTask,
		ConflictType: ConflictTypeUpdate,
		CreatedAt:    now,
	}

	assert.Nil(t, conflict.Resolution)
	assert.Nil(t, conflict.ResolvedAt)
}

// TestSyncOperation_FullFields tests SyncOperation with all fields
func TestSyncOperation_FullFields(t *testing.T) {
	now := time.Now()
	startedAt := now.Add(time.Minute)
	completedAt := now.Add(10 * time.Minute)

	task := &types.Task{
		ID:        uuid.New(),
		ProjectID: uuid.New(),
		Title:     "Test Task",
		State:     types.TaskStatePending,
		CreatedAt: now,
		UpdatedAt: now,
	}

	op := SyncOperation{
		ID:         uuid.New(),
		EntityID:   uuid.New(),
		EntityType: EntityTask,
		OpType:     OpCreate,
		Type:       OpCreate,
		Direction:  SyncLocalToMCP,
		Status:     StatusCompleted,
		LocalData:  task,
		ResultData: task,
		Priority:   5,
		Dependencies: []string{"dep-1", "dep-2"},
		CreatedAt:  now,
		StartedAt:  &startedAt,
		CompletedAt: &completedAt,
	}

	assert.Equal(t, OpCreate, op.OpType)
	assert.Equal(t, OpCreate, op.Type)
	assert.NotNil(t, op.StartedAt)
	assert.NotNil(t, op.CompletedAt)
	assert.Len(t, op.Dependencies, 2)
}

// TestSyncOperation_WithoutTimestamps tests SyncOperation without optional timestamps
func TestSyncOperation_WithoutTimestamps(t *testing.T) {
	now := time.Now()
	op := SyncOperation{
		ID:         uuid.New(),
		EntityID:   uuid.New(),
		EntityType: EntityProject,
		OpType:     OpUpdate,
		Status:     StatusPending,
		CreatedAt:  now,
	}

	assert.Nil(t, op.StartedAt)
	assert.Nil(t, op.CompletedAt)
	assert.Empty(t, op.Dependencies)
}

// TestConflictResolution_BackwardCompatibility tests backward compatibility aliases
func TestConflictResolution_BackwardCompatibility(t *testing.T) {
	now := time.Time{}
	resolution := ConflictResolution{
		Strategy:   "prefer-remote",
		ResolvedBy: "user-123",
		Actor:      "user-123", // Alias
		ResolvedAt: now,
		Timestamp:  now, // Alias
		Reason:     "User selected remote version",
	}

	assert.Equal(t, resolution.ResolvedBy, resolution.Actor)
	assert.Equal(t, resolution.ResolvedAt, resolution.Timestamp)
}
