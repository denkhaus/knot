package sync

import (
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewSyncDataSet(t *testing.T) {
	dataSet := shared.NewSyncDataSet()

	assert.NotNil(t, dataSet)
	assert.NotNil(t, dataSet.Projects)
	assert.NotNil(t, dataSet.Tasks)
	assert.Equal(t, 0, len(dataSet.Projects))
	assert.Equal(t, 0, len(dataSet.Tasks))
}

func TestSyncOpType_String(t *testing.T) {
	testCases := []struct {
		opType   shared.SyncOpType
		expected string
	}{
		{shared.OpCreate, "create"},
		{shared.OpUpdate, "update"},
		{shared.OpDelete, "delete"},
		{shared.OpConflict, "conflict"},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, string(tc.opType))
	}
}

func TestSyncEntityType_String(t *testing.T) {
	testCases := []struct {
		entityType shared.SyncEntityType
		expected   string
	}{
		{shared.EntityProject, "project"},
		{shared.EntityTask, "task"},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, string(tc.entityType))
	}
}

func TestSyncOpStatus_String(t *testing.T) {
	testCases := []struct {
		status   shared.SyncOpStatus
		expected string
	}{
		{shared.StatusPending, "pending"},
		{shared.StatusInProgress, "in_progress"},
		{shared.StatusCompleted, "completed"},
		{shared.StatusFailed, "failed"},
		{shared.StatusSkipped, "skipped"},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, string(tc.status))
	}
}

func TestSyncDirection_String(t *testing.T) {
	testCases := []struct {
		direction shared.SyncDirection
		expected  string
	}{
		{shared.SyncLocalToMCP, "local_to_mcp"},
		{shared.SyncMcpToLocal, "mcp_to_local"},
		{shared.SyncBidirectional, "bidirectional"},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, string(tc.direction))
	}
}

func TestConflictType_String(t *testing.T) {
	testCases := []struct {
		conflictType shared.ConflictType
		expected     string
	}{
		{shared.ConflictTypeUpdate, "update"},
		{shared.ConflictTypeDelete, "delete"},
		{shared.ConflictTypeState, "state"},
		{shared.ConflictTypeDependency, "dependency"},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, string(tc.conflictType))
	}
}

func TestConflictStrategy_String(t *testing.T) {
	testCases := []struct {
		strategy ConflictStrategy
		expected string
	}{
		{ConflictStrategyLastWriterWins, "last-writer-wins"},
		{ConflictStrategyPreferLocal, "prefer-local"},
		{ConflictStrategyPreferRemote, "prefer-remote"},
		{ConflictStrategyManual, "manual"},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, string(tc.strategy))
	}
}

func TestSyncMetrics_DefaultValues(t *testing.T) {
	metrics := &SyncMetrics{}

	assert.True(t, metrics.StartTime.IsZero())
	assert.True(t, metrics.EndTime.IsZero())
	assert.Equal(t, time.Duration(0), metrics.Duration)
	assert.Equal(t, 0, metrics.TotalOperations)
	assert.Equal(t, 0, metrics.CompletedOps)
	assert.Equal(t, 0, metrics.FailedOps)
	assert.Equal(t, 0, metrics.SkippedOps)
	assert.Equal(t, 0, metrics.ConflictCount)
	assert.Equal(t, 0, metrics.ResolvedConflicts)
	assert.Equal(t, int64(0), metrics.BytesTransferred)
	assert.Equal(t, 0, metrics.NetworkCalls)
	assert.Empty(t, metrics.Errors)
}

func TestSyncError_Creation(t *testing.T) {
	opID := uuid.New()
	timestamp := time.Now()

	syncErr := SyncError{
		OperationID: opID,
		Error:       "test error",
		Timestamp:   timestamp,
		Retryable:   true,
	}

	assert.Equal(t, opID, syncErr.OperationID)
	assert.Equal(t, "test error", syncErr.Error)
	assert.Equal(t, timestamp, syncErr.Timestamp)
	assert.True(t, syncErr.Retryable)
}

func TestSyncOperation_Creation(t *testing.T) {
	opID := uuid.New()
	now := time.Now()

	operation := shared.SyncOperation{
		ID:           opID,
		Type:         shared.OpCreate,
		EntityType:   shared.EntityProject,
		EntityID:     uuid.New(),
		LocalData:    &types.Project{ID: uuid.New()},
		RemoteData:   &types.Project{ID: uuid.New()},
		Direction:    shared.SyncLocalToMCP,
		Reason:       "test reason",
		Priority:     5,
		Status:       shared.StatusPending,
		Error:        "",
		Dependencies: []string{},
		CreatedAt:    now,
	}

	assert.Equal(t, opID, operation.ID)
	assert.Equal(t, shared.OpCreate, operation.Type)
	assert.Equal(t, shared.EntityProject, operation.EntityType)
	assert.NotEmpty(t, operation.Reason)
	assert.Equal(t, 5, operation.Priority)
	assert.Equal(t, shared.StatusPending, operation.Status)
	assert.Empty(t, operation.Error)
	assert.Empty(t, operation.Dependencies)
	assert.False(t, operation.CreatedAt.IsZero())
}
