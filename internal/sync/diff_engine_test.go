package sync

import (
	"context"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDiffEngine(t *testing.T) {
	log := &noopLogger{}
	engine := NewDiffEngine(log)

	assert.NotNil(t, engine)
}

func TestDiffEngine_CalculateDiff_EmptyDataSets(t *testing.T) {
	log := &noopLogger{}
	engine := NewDiffEngine(log)
	ctx := context.Background()

	localData := shared.NewSyncDataSet()
	remoteData := shared.NewSyncDataSet()

	result, err := engine.CalculateDiff(ctx, localData, remoteData)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result.Operations))
	assert.Equal(t, 0, result.Summary.Total)
	assert.True(t, result.Duration >= 0)
}

func TestDiffEngine_CalculateDiff_ProjectDifferences(t *testing.T) {
	log := &noopLogger{}
	engine := NewDiffEngine(log)
	ctx := context.Background()

	localData := shared.NewSyncDataSet()
	remoteData := shared.NewSyncDataSet()

	// Add project only in local data
	localProjectID := uuid.New()
	localProject := &types.Project{
		ID:        localProjectID,
		Title:     "Local Only Project",
		State:     types.ProjectStateActive,
		UpdatedAt: testTime,
		CreatedAt: testTime,
	}
	localData.Projects[localProjectID] = localProject

	// Add project only in remote data
	remoteProjectID := uuid.New()
	remoteProject := &types.Project{
		ID:        remoteProjectID,
		Title:     "Remote Only Project",
		State:     types.ProjectStateActive,
		UpdatedAt: testTime.Add(time.Hour),
		CreatedAt: testTime,
	}
	remoteData.Projects[remoteProjectID] = remoteProject

	result, err := engine.CalculateDiff(ctx, localData, remoteData)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result.Operations))

	// Check local-only project creates remote operation
	localOp := findOperationByEntityID(result.Operations, localProjectID)
	assert.NotNil(t, localOp)
	assert.Equal(t, shared.OpCreate, localOp.Type)
	assert.Equal(t, shared.SyncLocalToMCP, localOp.Direction)
	assert.Equal(t, "Local project not found remotely", localOp.Reason)

	// Check remote-only project creates local operation
	remoteOp := findOperationByEntityID(result.Operations, remoteProjectID)
	assert.NotNil(t, remoteOp)
	assert.Equal(t, shared.OpCreate, remoteOp.Type)
	assert.Equal(t, shared.SyncMcpToLocal, remoteOp.Direction)
	assert.Equal(t, "Remote project not found locally", remoteOp.Reason)
}

func TestDiffEngine_CalculateDiff_TaskDifferences(t *testing.T) {
	log := &noopLogger{}
	engine := NewDiffEngine(log)
	ctx := context.Background()

	localData := shared.NewSyncDataSet()
	remoteData := shared.NewSyncDataSet()

	// Add task only in local data
	localTaskID := uuid.New()
	localProjectID := uuid.New()
	localTask := &types.Task{
		ID:        localTaskID,
		ProjectID: localProjectID,
		Title:     "Local Only Task",
		State:     types.TaskStatePending,
		UpdatedAt: testTime,
		CreatedAt: testTime,
	}
	localData.Tasks[localTaskID] = localTask

	// Add task only in remote data
	remoteTaskID := uuid.New()
	remoteProjectID := uuid.New()
	remoteTask := &types.Task{
		ID:        remoteTaskID,
		ProjectID: remoteProjectID,
		Title:     "Remote Only Task",
		State:     types.TaskStatePending,
		UpdatedAt: testTime.Add(time.Hour),
		CreatedAt: testTime,
	}
	remoteData.Tasks[remoteTaskID] = remoteTask

	result, err := engine.CalculateDiff(ctx, localData, remoteData)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result.Operations))

	// Check local-only task creates remote operation
	localOp := findOperationByEntityID(result.Operations, localTaskID)
	assert.NotNil(t, localOp)
	assert.Equal(t, shared.OpCreate, localOp.Type)
	assert.Equal(t, shared.EntityTask, localOp.EntityType)
	assert.Equal(t, shared.SyncLocalToMCP, localOp.Direction)

	// Check remote-only task creates local operation
	remoteOp := findOperationByEntityID(result.Operations, remoteTaskID)
	assert.NotNil(t, remoteOp)
	assert.Equal(t, shared.OpCreate, remoteOp.Type)
	assert.Equal(t, shared.EntityTask, remoteOp.EntityType)
	assert.Equal(t, shared.SyncMcpToLocal, remoteOp.Direction)
}

func TestDiffEngine_ValidateDiff_ValidResult(t *testing.T) {
	log := &noopLogger{}
	engine := NewDiffEngine(log)
	ctx := context.Background()

	projectID := uuid.New()
	localProject := &types.Project{
		ID:        projectID,
		Title:     "Local Project",
		State:     types.ProjectStateActive,
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}

	result := &DiffResult{
		LocalTimestamp:  time.Now(),
		RemoteTimestamp: time.Now(),
		Operations: []shared.SyncOperation{
			{
				ID:         uuid.New(),
				Type:       shared.OpCreate,
				EntityType: shared.EntityProject,
				EntityID:   projectID,
				Direction:  shared.SyncLocalToMCP,
				Reason:     "Project not found remotely",
				LocalData:  localProject,
				Priority:   5,
				CreatedAt:  testTime,
			},
		},
	}

	err := engine.ValidateDiff(ctx, result)
	assert.NoError(t, err)
}

// Helper function to find operation by entity ID
func findOperationByEntityID(operations []shared.SyncOperation, entityID uuid.UUID) *shared.SyncOperation {
	for _, op := range operations {
		if op.EntityID == entityID {
			return &op
		}
	}
	return nil
}
