package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/denkhaus/knot/v2/internal/sync"
	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// TestSyncService_PerformPullSync_FirstTimeSync tests that PerformPullSync handles
// missing remote projects gracefully for first-time sync scenarios
func TestSyncService_PerformPullSync_FirstTimeSync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockLogger := &mockNoopLogger{}
	mockProjectManager := mocks.NewMockProjectManager(ctrl)

	// Create sync service (diffEngine not used in PerformPullSync)
	service := &syncServiceImpl{
		projectManager: mockProjectManager,
		logger:         mockLogger,
		diffEngine:     &noopDiffEngine{},
	}

	// Create test request
	projectID := uuid.New()
	request := shared.SyncRequest{
		RequestID: uuid.New(),
		ProjectID: projectID,
		Direction: shared.SyncMcpToLocal,
		Timestamp: time.Now(),
	}

	// Setup expectation: GetProject returns error (project not found)
	mockProjectManager.EXPECT().
		GetProject(gomock.Any(), projectID).
		Return(nil, errors.New("project not found")).
		Times(1)

	// Perform pull sync
	response, err := service.PerformPullSync(context.Background(), request)

	// Verify response
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Success, "Response should be successful even when project doesn't exist")
	assert.Equal(t, request.RequestID, response.RequestID)
	assert.Equal(t, 0, response.Processed, "No items should be processed for missing project")
	assert.Equal(t, 0, response.Created)
	assert.Equal(t, 0, response.Updated)
	assert.Equal(t, 0, response.Deleted)
	assert.Empty(t, response.Errors, "No errors should be returned for first-time sync")

	// Verify RemoteChanges is empty
	assert.NotNil(t, response.RemoteChanges)
	assert.Empty(t, response.RemoteChanges.Projects, "Projects should be empty for missing project")
	assert.Empty(t, response.RemoteChanges.Tasks, "Tasks should be empty for missing project")
}

// TestSyncService_PerformPullSync_ExistingProject tests that PerformPullSync
// returns project data when it exists on remote
func TestSyncService_PerformPullSync_ExistingProject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockLogger := &mockNoopLogger{}
	mockProjectManager := mocks.NewMockProjectManager(ctrl)

	// Create sync service (diffEngine not used in PerformPullSync)
	service := &syncServiceImpl{
		projectManager: mockProjectManager,
		logger:         mockLogger,
		diffEngine:     &noopDiffEngine{},
	}

	// Create test data
	projectID := uuid.New()
	taskID1 := uuid.New()
	taskID2 := uuid.New()

	project := &types.Project{
		ID:        projectID,
		Title:     "Test Project",
		State:     types.ProjectStateActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tasks := []*types.Task{
		{
			ID:        taskID1,
			ProjectID: projectID,
			Title:     "Task 1",
			State:     types.TaskStatePending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        taskID2,
			ProjectID: projectID,
			Title:     "Task 2",
			State:     types.TaskStateInProgress,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	request := shared.SyncRequest{
		RequestID: uuid.New(),
		ProjectID: projectID,
		Direction: shared.SyncMcpToLocal,
		Timestamp: time.Now(),
	}

	// Setup expectations
	mockProjectManager.EXPECT().
		GetProject(gomock.Any(), projectID).
		Return(project, nil).
		Times(1)

	mockProjectManager.EXPECT().
		ListTasksForProject(gomock.Any(), projectID).
		Return(tasks, nil).
		Times(1)

	// Perform pull sync
	response, err := service.PerformPullSync(context.Background(), request)

	// Verify response
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Success)
	assert.Equal(t, request.RequestID, response.RequestID)
	assert.Equal(t, 3, response.Processed, "Should process 1 project + 2 tasks")
	assert.Equal(t, 0, response.Created)
	assert.Equal(t, 0, response.Updated)
	assert.Equal(t, 0, response.Deleted)
	assert.Empty(t, response.Errors)

	// Verify RemoteChanges contains project and tasks
	assert.NotNil(t, response.RemoteChanges)
	assert.Len(t, response.RemoteChanges.Projects, 1)
	assert.Len(t, response.RemoteChanges.Tasks, 2)
	assert.Equal(t, projectID, response.RemoteChanges.Projects[projectID].ID)
	assert.Equal(t, taskID1, response.RemoteChanges.Tasks[taskID1].ID)
	assert.Equal(t, taskID2, response.RemoteChanges.Tasks[taskID2].ID)
}

// TestSyncService_PerformPullSync_TasksListError tests error handling
// when ListTasksForProject fails after successful GetProject
func TestSyncService_PerformPullSync_TasksListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockLogger := &mockNoopLogger{}
	mockProjectManager := mocks.NewMockProjectManager(ctrl)

	// Create sync service (diffEngine not used in PerformPullSync)
	service := &syncServiceImpl{
		projectManager: mockProjectManager,
		logger:         mockLogger,
		diffEngine:     &noopDiffEngine{},
	}

	// Create test data
	projectID := uuid.New()
	project := &types.Project{
		ID:        projectID,
		Title:     "Test Project",
		State:     types.ProjectStateActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	request := shared.SyncRequest{
		RequestID: uuid.New(),
		ProjectID: projectID,
		Direction: shared.SyncMcpToLocal,
		Timestamp: time.Now(),
	}

	// Setup expectations
	mockProjectManager.EXPECT().
		GetProject(gomock.Any(), projectID).
		Return(project, nil).
		Times(1)

	mockProjectManager.EXPECT().
		ListTasksForProject(gomock.Any(), projectID).
		Return(nil, errors.New("database error")).
		Times(1)

	// Perform pull sync
	response, err := service.PerformPullSync(context.Background(), request)

	// Verify error is returned
	require.Error(t, err)
	assert.NotNil(t, response)
	assert.False(t, response.Success)
	assert.NotEmpty(t, response.Errors)
	assert.Contains(t, response.Errors[0], "failed to list tasks")
}

// mockNoopLogger is a simple no-op logger for testing
type mockNoopLogger struct{}

func (l *mockNoopLogger) Debug(msg string, fields ...zap.Field) {}
func (l *mockNoopLogger) Info(msg string, fields ...zap.Field)  {}
func (l *mockNoopLogger) Warn(msg string, fields ...zap.Field)  {}
func (l *mockNoopLogger) Error(msg string, fields ...zap.Field) {}
func (l *mockNoopLogger) Fatal(msg string, fields ...zap.Field) {}
func (l *mockNoopLogger) Sync()                                 {}
func (l *mockNoopLogger) With(fields ...zap.Field) logger.Logger {
	return l
}

func (l *mockNoopLogger) Named(name string) logger.Logger {
	return l
}

func (l *mockNoopLogger) ToZap() *zap.Logger {
	return zap.NewNop()
}
func (l *mockNoopLogger) SetLevel(logLevel string) {}

// Ensure mockNoopLogger implements logger.Logger
var _ logger.Logger = (*mockNoopLogger)(nil)

// noopDiffEngine is a no-op diff engine for testing
type noopDiffEngine struct{}

func (e *noopDiffEngine) CalculateDiff(ctx context.Context, localData, remoteData *shared.SyncDataSet) (*sync.DiffResult, error) {
	return &sync.DiffResult{}, nil
}

func (e *noopDiffEngine) ValidateDiff(ctx context.Context, result *sync.DiffResult) error {
	return nil
}
