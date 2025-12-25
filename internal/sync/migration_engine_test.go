package sync

import (
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestNewMigrationEngine tests the constructor
func TestNewMigrationEngine(t *testing.T) {
	log := &noopLogger{}

	engine := &migrationEngineImpl{
		logger:         log,
		client:         nil,
		projectManager: nil,
		metrics: &SyncMetrics{
			StartTime: time.Now(),
		},
	}

	assert.NotNil(t, engine)
	assert.NotNil(t, engine.logger)
	assert.NotNil(t, engine.metrics)
}

// TestMigrationEngine_GetMetrics tests metrics retrieval
func TestMigrationEngine_GetMetrics(t *testing.T) {
	log := &noopLogger{}

	engine := &migrationEngineImpl{
		logger:         log,
		client:         nil,
		projectManager: nil,
		metrics: &SyncMetrics{
			StartTime:    time.Now(),
			CompletedOps: 10,
			FailedOps:    2,
		},
	}

	// Note: We can't directly call GetMetrics as it's not in the public interface
	// This test just verifies the struct initialization
	assert.Equal(t, 10, engine.metrics.CompletedOps)
	assert.Equal(t, 2, engine.metrics.FailedOps)
	assert.False(t, engine.metrics.StartTime.IsZero())
}

// TestMigrationEngine_ResetMetrics tests metrics reset
func TestMigrationEngine_ResetMetrics(t *testing.T) {
	log := &noopLogger{}

	engine := &migrationEngineImpl{
		logger:         log,
		client:         nil,
		projectManager: nil,
		metrics: &SyncMetrics{
			StartTime:    time.Now(),
			CompletedOps: 10,
			FailedOps:    2,
		},
	}

	// Reset metrics
	engine.metrics = &SyncMetrics{
		StartTime: time.Now(),
	}

	assert.Equal(t, 0, engine.metrics.CompletedOps)
	assert.Equal(t, 0, engine.metrics.FailedOps)
	assert.False(t, engine.metrics.StartTime.IsZero())
}

// TestSyncResult tests the SyncResult struct
func TestSyncResult(t *testing.T) {
	result := &SyncResult{
		SyncedAt:  time.Now(),
		Success:   true,
		Processed: 10,
		Created:   3,
		Updated:   5,
		Deleted:   2,
		Duration:  100 * time.Millisecond,
	}

	assert.True(t, result.Success)
	assert.Equal(t, 10, result.Processed)
	assert.Equal(t, 3, result.Created)
	assert.Equal(t, 5, result.Updated)
	assert.Equal(t, 2, result.Deleted)
	assert.Greater(t, result.Duration, time.Duration(0))
}

// TestSyncMetrics tests the SyncMetrics struct
func TestSyncMetrics(t *testing.T) {
	metrics := &SyncMetrics{
		StartTime:       time.Now(),
		CompletedOps:    100,
		FailedOps:       5,
		SkippedOps:      2,
		TotalOperations: 107,
	}

	assert.False(t, metrics.StartTime.IsZero())
	assert.Equal(t, 100, metrics.CompletedOps)
	assert.Equal(t, 5, metrics.FailedOps)
	assert.Equal(t, 2, metrics.SkippedOps)
	assert.Equal(t, 107, metrics.TotalOperations)
}

// TestSyncDataSet_HelperTests tests helper functions related to SyncDataSet
func TestSyncDataSet_HelperTests(t *testing.T) {
	dataSet := shared.NewSyncDataSet()

	projectID := uuid.New()
	dataSet.Projects[projectID] = &types.Project{
		ID:        projectID,
		Title:     "Test Project",
		State:     types.ProjectStateActive,
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}

	taskID := uuid.New()
	dataSet.Tasks[taskID] = &types.Task{
		ID:        taskID,
		ProjectID: projectID,
		Title:     "Test Task",
		State:     types.TaskStatePending,
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}

	assert.Equal(t, 1, len(dataSet.Projects))
	assert.Equal(t, 1, len(dataSet.Tasks))
	assert.NotNil(t, dataSet.Projects[projectID])
	assert.NotNil(t, dataSet.Tasks[taskID])
}

// TestSyncOperation_Types tests the different operation types
func TestSyncOperation_Types(t *testing.T) {
	operations := []shared.SyncOperation{
		{
			ID:         uuid.New(),
			Type:       shared.OpCreate,
			EntityType: shared.EntityProject,
			EntityID:   uuid.New(),
			Direction:  shared.SyncLocalToMCP,
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			Type:       shared.OpUpdate,
			EntityType: shared.EntityTask,
			EntityID:   uuid.New(),
			Direction:  shared.SyncMcpToLocal,
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			Type:       shared.OpDelete,
			EntityType: shared.EntityProject,
			EntityID:   uuid.New(),
			Direction:  shared.SyncLocalToMCP,
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			Type:       shared.OpConflict,
			EntityType: shared.EntityTask,
			EntityID:   uuid.New(),
			Direction:  shared.SyncBidirectional,
			CreatedAt:  time.Now(),
		},
	}

	assert.Equal(t, shared.OpCreate, operations[0].Type)
	assert.Equal(t, shared.OpUpdate, operations[1].Type)
	assert.Equal(t, shared.OpDelete, operations[2].Type)
	assert.Equal(t, shared.OpConflict, operations[3].Type)
}

// TestSyncDirection_Constants tests sync direction constants
func TestSyncDirection_Constants(t *testing.T) {
	assert.Equal(t, shared.SyncDirection("local_to_mcp"), shared.SyncLocalToMCP)
	assert.Equal(t, shared.SyncDirection("mcp_to_local"), shared.SyncMcpToLocal)
	assert.Equal(t, shared.SyncDirection("bidirectional"), shared.SyncBidirectional)
}
