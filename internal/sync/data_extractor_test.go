package sync_test

import (
	"context"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mocks"
	knotsync "github.com/denkhaus/knot/v2/internal/sync"
	"github.com/denkhaus/knot/v2/internal/sync/client"
	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// testLogger is a minimal logger implementation for testing
type testLogger struct{}

func newTestLogger() *testLogger {
	return &testLogger{}
}

func (l *testLogger) Debug(msg string, fields ...zap.Field) {}
func (l *testLogger) Info(msg string, fields ...zap.Field)  {}
func (l *testLogger) Warn(msg string, fields ...zap.Field)  {}
func (l *testLogger) Error(msg string, fields ...zap.Field) {}
func (l *testLogger) Sync()                                 {}
func (l *testLogger) With(fields ...zap.Field) logger.Logger {
	return l
}
func (l *testLogger) Named(name string) logger.Logger {
	return l
}
func (l *testLogger) ToZap() *zap.Logger {
	return zap.NewNop()
}
func (l *testLogger) SetLevel(level string) {}

// newTestDataExtractor creates a test DataExtractor using DI with mocked dependencies
func newTestDataExtractor(ctrl *gomock.Controller) knotsync.DataExtractor {
	// Create mocks
	mockPM := mocks.NewMockProjectManager(ctrl)
	mockClient := mocks.NewMockRESTSyncClient(ctrl)

	// Setup default mock expectations
	mockClient.EXPECT().GetRemoteData(gomock.Any(), gomock.Any()).Return(&shared.SyncDataSet{}, nil).AnyTimes()
	mockPM.EXPECT().ListTasksForProject(gomock.Any(), gomock.Any()).Return([]*types.Task{}, nil).AnyTimes()
	mockPM.EXPECT().GetTasksWithDependencies(gomock.Any(), gomock.Any()).Return([]*types.Task{}, nil).AnyTimes()

	// Create test DI injector
	injector := do.New()
	do.ProvideValue[logger.Logger](injector, newTestLogger())
	do.ProvideValue[manager.ProjectManager](injector, mockPM)
	do.ProvideValue[client.RESTSyncClient](injector, mockClient)

	// Use the real provider function to create DataExtractor
	// This uses the actual NewDataExtractorService which will get the injected mocks
	extractor, err := knotsync.NewDataExtractorService(injector)
	if err != nil {
		panic(err)
	}

	return extractor
}

func TestNewDataExtractor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	extractor := newTestDataExtractor(ctrl)

	assert.NotNil(t, extractor)
}

// TestDataExtractor_ExtractLocalData tests extracting local data
func TestDataExtractor_ExtractLocalData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPM := mocks.NewMockProjectManager(ctrl)
	mockClient := mocks.NewMockRESTSyncClient(ctrl)

	// Setup mock expectations
	mockPM.EXPECT().ListTasksForProject(gomock.Any(), gomock.Any()).Return([]*types.Task{}, nil).AnyTimes()
	mockPM.EXPECT().GetTasksWithDependencies(gomock.Any(), gomock.Any()).Return([]*types.Task{}, nil).AnyTimes()
	mockClient.EXPECT().GetRemoteData(gomock.Any(), gomock.Any()).Return(&shared.SyncDataSet{}, nil).AnyTimes()

	injector := do.New()
	do.ProvideValue[logger.Logger](injector, newTestLogger())
	do.ProvideValue[manager.ProjectManager](injector, mockPM)
	do.ProvideValue[client.RESTSyncClient](injector, mockClient)

	extractor, err := knotsync.NewDataExtractorService(injector)
	require.NoError(t, err)

	ctx := context.Background()
	projectID := uuid.New()

	// Test with mock project manager
	data, err := extractor.ExtractLocalData(ctx, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 0, len(data.Projects))
	assert.Equal(t, 0, len(data.Tasks))
}

func TestDataExtractor_ValidateDataSet_EmptyDataSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	extractor := newTestDataExtractor(ctrl)
	ctx := context.Background()
	dataSet := shared.NewSyncDataSet()

	err := extractor.ValidateDataSet(ctx, dataSet)

	assert.NoError(t, err)
}

func TestDataExtractor_ValidateDataSet_ValidDataSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	extractor := newTestDataExtractor(ctrl)
	ctx := context.Background()
	dataSet := shared.NewSyncDataSet()

	// Add valid project and task
	projectID := uuid.New()
	project := &types.Project{
		ID:        projectID,
		Title:     "Test Project",
		State:     types.ProjectStateActive,
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}
	dataSet.Projects[projectID] = project

	taskID := uuid.New()
	task := &types.Task{
		ID:        taskID,
		ProjectID: projectID,
		Title:     "Test Task",
		State:     types.TaskStatePending,
		Priority:  5,
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}
	dataSet.Tasks[taskID] = task

	err := extractor.ValidateDataSet(ctx, dataSet)

	assert.NoError(t, err)
}

func TestDataExtractor_ValidateDataSet_InvalidProjectReference(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	extractor := newTestDataExtractor(ctrl)
	ctx := context.Background()
	dataSet := shared.NewSyncDataSet()

	// Add task with invalid project reference
	taskID := uuid.New()
	projectID := uuid.New()
	task := &types.Task{
		ID:        taskID,
		ProjectID: projectID,
		Title:     "Test Task",
		State:     types.TaskStatePending,
		Priority:  5,
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}
	dataSet.Tasks[taskID] = task

	// This should not return an error, just log a warning
	err := extractor.ValidateDataSet(ctx, dataSet)
	assert.NoError(t, err)
}

func TestDataExtractor_ValidateDataSet_FutureTimestamp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	extractor := newTestDataExtractor(ctrl)
	ctx := context.Background()
	dataSet := shared.NewSyncDataSet()

	// Add project with future timestamp
	projectID := uuid.New()
	project := &types.Project{
		ID:        projectID,
		Title:     "Test Project",
		State:     types.ProjectStateActive,
		CreatedAt: testTime,
		UpdatedAt: time.Now().Add(24 * time.Hour),
	}
	dataSet.Projects[projectID] = project

	err := extractor.ValidateDataSet(ctx, dataSet)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "future timestamp")
}

func TestDataExtractor_FilterByTimeRange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	extractor := newTestDataExtractor(ctrl)

	projectID1 := uuid.New()
	projectID2 := uuid.New()
	taskID1 := uuid.New()
	taskID2 := uuid.New()

	dataSet := &shared.SyncDataSet{
		Projects: map[uuid.UUID]*types.Project{
			projectID1: {
				ID:        projectID1,
				Title:     "Old Project",
				State:     types.ProjectStateActive,
				CreatedAt: testTime.Add(-48 * time.Hour),
				UpdatedAt: testTime.Add(-48 * time.Hour),
			},
			projectID2: {
				ID:        projectID2,
				Title:     "New Project",
				State:     types.ProjectStateActive,
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
		Tasks: map[uuid.UUID]*types.Task{
			taskID1: {
				ID:        taskID1,
				ProjectID: projectID2,
				Title:     "Old Task",
				State:     types.TaskStatePending,
				CreatedAt: testTime.Add(-24 * time.Hour),
				UpdatedAt: testTime.Add(-24 * time.Hour),
			},
			taskID2: {
				ID:        taskID2,
				ProjectID: projectID2,
				Title:     "New Task",
				State:     types.TaskStatePending,
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
	}

	// Filter to only include items updated in the last 1 hour
	filtered := extractor.FilterByTimeRange(dataSet, testTime.Add(-1*time.Hour))

	// Only the new project and task should remain
	assert.Equal(t, 1, len(filtered.Projects))
	assert.Equal(t, 1, len(filtered.Tasks))
	assert.Contains(t, filtered.Projects, projectID2)
	assert.Contains(t, filtered.Tasks, taskID2)
}

func TestDataExtractor_GetStatistics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	extractor := newTestDataExtractor(ctrl)

	projectID1 := uuid.New()
	projectID2 := uuid.New()
	taskID1 := uuid.New()
	taskID2 := uuid.New()
	taskID3 := uuid.New()

	dataSet := &shared.SyncDataSet{
		Projects: map[uuid.UUID]*types.Project{
			projectID1: {
				ID:        projectID1,
				Title:     "Active Project",
				State:     types.ProjectStateActive,
				CreatedAt: testTime.Add(-48 * time.Hour),
				UpdatedAt: testTime.Add(-48 * time.Hour),
			},
			projectID2: {
				ID:        projectID2,
				Title:     "Completed Project",
				State:     types.ProjectStateCompleted,
				CreatedAt: testTime.Add(-24 * time.Hour),
				UpdatedAt: testTime.Add(-24 * time.Hour),
			},
		},
		Tasks: map[uuid.UUID]*types.Task{
			taskID1: {
				ID:        taskID1,
				Title:     "Task 1",
				State:     types.TaskStatePending,
				Priority:  5,
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
			taskID2: {
				ID:        taskID2,
				Title:     "Task 2",
				State:     types.TaskStateCompleted,
				Priority:  8,
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
			taskID3: {
				ID:        taskID3,
				Title:     "Task 3",
				State:     types.TaskStatePending,
				Priority:  3,
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
	}

	stats := extractor.GetStatistics(dataSet)

	assert.Equal(t, 2, stats.TotalProjects)
	assert.Equal(t, 3, stats.TotalTasks)
	assert.Equal(t, 2, stats.TasksByState[types.TaskStatePending])
	assert.Equal(t, 1, stats.TasksByState[types.TaskStateCompleted])
	assert.Equal(t, 1, stats.ProjectsByState[types.ProjectStateActive])
	assert.Equal(t, 1, stats.ProjectsByState[types.ProjectStateCompleted])

	// Check that the age span is calculated
	assert.NotNil(t, stats.OldestProject)
	assert.NotNil(t, stats.NewestProject)
	assert.NotNil(t, stats.OldestTask)
	assert.NotNil(t, stats.NewestTask)
}

// TestDataExtractor_ExtractRemoteData_Success tests extracting data from remote server
func TestDataExtractor_ExtractRemoteData_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPM := mocks.NewMockProjectManager(ctrl)
	mockClient := mocks.NewMockRESTSyncClient(ctrl)

	ctx := context.Background()
	projectID := uuid.New()

	// Setup mock expectations
	expectedDataSet := &shared.SyncDataSet{
		Projects: map[uuid.UUID]*types.Project{
			projectID: {
				ID:        projectID,
				Title:     "Remote Project",
				State:     types.ProjectStateActive,
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
		Tasks: map[uuid.UUID]*types.Task{
			uuid.New(): {
				ID:        uuid.New(),
				ProjectID: projectID,
				Title:     "Remote Task",
				State:     types.TaskStatePending,
				Priority:  5,
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
	}

	mockPM.EXPECT().ListTasksForProject(gomock.Any(), gomock.Any()).Return([]*types.Task{}, nil).AnyTimes()
	mockPM.EXPECT().GetTasksWithDependencies(gomock.Any(), gomock.Any()).Return([]*types.Task{}, nil).AnyTimes()
	mockClient.EXPECT().GetRemoteData(ctx, &projectID).Return(expectedDataSet, nil)

	injector := do.New()
	do.ProvideValue[logger.Logger](injector, newTestLogger())
	do.ProvideValue[manager.ProjectManager](injector, mockPM)
	do.ProvideValue[client.RESTSyncClient](injector, mockClient)

	extractor, err := knotsync.NewDataExtractorService(injector)
	require.NoError(t, err)

	data, err := extractor.ExtractRemoteData(ctx, projectID)

	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 1, len(data.Projects))
	assert.Equal(t, 1, len(data.Tasks))
	assert.Equal(t, "Remote Project", data.Projects[projectID].Title)
}

// TestDataExtractor_ExtractRemoteData_ClientError tests error handling from client
func TestDataExtractor_ExtractRemoteData_ClientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPM := mocks.NewMockProjectManager(ctrl)
	mockClient := mocks.NewMockRESTSyncClient(ctrl)

	ctx := context.Background()
	projectID := uuid.New()

	// Setup mock to return error
	mockPM.EXPECT().ListTasksForProject(gomock.Any(), gomock.Any()).Return([]*types.Task{}, nil).AnyTimes()
	mockPM.EXPECT().GetTasksWithDependencies(gomock.Any(), gomock.Any()).Return([]*types.Task{}, nil).AnyTimes()
	mockClient.EXPECT().GetRemoteData(ctx, &projectID).Return(nil, assert.AnError)

	injector := do.New()
	do.ProvideValue[logger.Logger](injector, newTestLogger())
	do.ProvideValue[manager.ProjectManager](injector, mockPM)
	do.ProvideValue[client.RESTSyncClient](injector, mockClient)

	extractor, err := knotsync.NewDataExtractorService(injector)
	require.NoError(t, err)

	data, err := extractor.ExtractRemoteData(ctx, projectID)

	assert.Error(t, err)
	assert.Nil(t, data)
}

// TestDataExtractor_ExtractProjectData tests combining local and remote data
func TestDataExtractor_ExtractProjectData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPM := mocks.NewMockProjectManager(ctrl)
	mockClient := mocks.NewMockRESTSyncClient(ctrl)

	ctx := context.Background()
	projectID := uuid.New()

	// Setup mock expectations
	localTaskID := uuid.New()
	localTasks := []*types.Task{
		{
			ID:          localTaskID,
			ProjectID:   projectID,
			Title:       "Local Task",
			State:       types.TaskStatePending,
			Priority:    5,
			Complexity:  3,
			CreatedAt:   testTime,
			UpdatedAt:   testTime,
			UpdatedBy:   "local-user",
			Dependencies: []uuid.UUID{},
		},
	}

	mockPM.EXPECT().ListTasksForProject(ctx, projectID).Return(localTasks, nil)
	mockPM.EXPECT().GetTasksWithDependencies(ctx, gomock.Any()).Return(localTasks, nil)

	remoteProjectID := uuid.New()
	remoteTaskID := uuid.New()
	remoteDataSet := &shared.SyncDataSet{
		Projects: map[uuid.UUID]*types.Project{
			remoteProjectID: {
				ID:        remoteProjectID,
				Title:     "Remote Project",
				State:     types.ProjectStateActive,
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
		Tasks: map[uuid.UUID]*types.Task{
			remoteTaskID: {
				ID:        remoteTaskID,
				ProjectID: projectID,
				Title:     "Remote Task",
				State:     types.TaskStatePending,
				Priority:  8,
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
	}
	mockClient.EXPECT().GetRemoteData(ctx, &projectID).Return(remoteDataSet, nil)

	injector := do.New()
	do.ProvideValue[logger.Logger](injector, newTestLogger())
	do.ProvideValue[manager.ProjectManager](injector, mockPM)
	do.ProvideValue[client.RESTSyncClient](injector, mockClient)

	extractor, err := knotsync.NewDataExtractorService(injector)
	require.NoError(t, err)

	data, err := extractor.ExtractProjectData(ctx, projectID)

	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 1, len(data.Projects))
	assert.Equal(t, 2, len(data.Tasks))
	assert.Contains(t, data.Projects, remoteProjectID)
	assert.Contains(t, data.Tasks, localTaskID)
	assert.Contains(t, data.Tasks, remoteTaskID)
}

// TestDataExtractor_ExtractLocalData_WithTasks tests extracting local data with tasks
func TestDataExtractor_ExtractLocalData_WithTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPM := mocks.NewMockProjectManager(ctrl)
	mockClient := mocks.NewMockRESTSyncClient(ctrl)

	ctx := context.Background()
	projectID := uuid.New()

	taskID := uuid.New()
	tasks := []*types.Task{
		{
			ID:          taskID,
			ProjectID:   projectID,
			Title:       "Test Task",
			Description: "Test Description",
			State:       types.TaskStateInProgress,
			Priority:    7,
			Complexity:  5,
			CreatedAt:   testTime,
			UpdatedAt:   testTime.Add(time.Hour),
			UpdatedBy:   "test-user",
			Dependencies: []uuid.UUID{},
		},
	}

	mockPM.EXPECT().ListTasksForProject(ctx, projectID).Return(tasks, nil)
	mockPM.EXPECT().GetTasksWithDependencies(ctx, gomock.Any()).Return(tasks, nil)
	mockClient.EXPECT().GetRemoteData(gomock.Any(), gomock.Any()).Return(&shared.SyncDataSet{}, nil).AnyTimes()

	injector := do.New()
	do.ProvideValue[logger.Logger](injector, newTestLogger())
	do.ProvideValue[manager.ProjectManager](injector, mockPM)
	do.ProvideValue[client.RESTSyncClient](injector, mockClient)

	extractor, err := knotsync.NewDataExtractorService(injector)
	require.NoError(t, err)

	data, err := extractor.ExtractLocalData(ctx, projectID)

	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 1, len(data.Tasks))
	assert.Contains(t, data.Tasks, taskID)
}

// TestDataExtractor_ExtractLocalData_WithDependencies tests extracting tasks with dependencies
func TestDataExtractor_ExtractLocalData_WithDependencies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPM := mocks.NewMockProjectManager(ctrl)
	mockClient := mocks.NewMockRESTSyncClient(ctrl)

	ctx := context.Background()
	projectID := uuid.New()

	depID := uuid.New()
	taskID := uuid.New()
	tasks := []*types.Task{
		{
			ID:          taskID,
			ProjectID:   projectID,
			Title:       "Task with Dependencies",
			State:       types.TaskStatePending,
			Priority:    5,
			CreatedAt:   testTime,
			UpdatedAt:   testTime,
			Dependencies: []uuid.UUID{depID},
		},
	}

	mockPM.EXPECT().ListTasksForProject(ctx, projectID).Return(tasks, nil)
	mockPM.EXPECT().GetTasksWithDependencies(ctx, gomock.Any()).Return(tasks, nil)
	mockClient.EXPECT().GetRemoteData(gomock.Any(), gomock.Any()).Return(&shared.SyncDataSet{}, nil).AnyTimes()

	injector := do.New()
	do.ProvideValue[logger.Logger](injector, newTestLogger())
	do.ProvideValue[manager.ProjectManager](injector, mockPM)
	do.ProvideValue[client.RESTSyncClient](injector, mockClient)

	extractor, err := knotsync.NewDataExtractorService(injector)
	require.NoError(t, err)

	data, err := extractor.ExtractLocalData(ctx, projectID)

	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 1, len(data.Tasks))
	assert.Contains(t, data.Tasks, taskID)
	assert.Equal(t, 1, len(data.Tasks[taskID].Dependencies))
	assert.Contains(t, data.Tasks[taskID].Dependencies, depID)
}

// TestDataExtractor_ExtractLocalData_WithParentID tests extracting tasks with parent ID
func TestDataExtractor_ExtractLocalData_WithParentID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPM := mocks.NewMockProjectManager(ctrl)
	mockClient := mocks.NewMockRESTSyncClient(ctrl)

	ctx := context.Background()
	projectID := uuid.New()

	parentID := uuid.New()
	taskID := uuid.New()
	tasks := []*types.Task{
		{
			ID:        taskID,
			ProjectID: projectID,
			Title:     "Child Task",
			State:     types.TaskStatePending,
			Priority:  5,
			CreatedAt: testTime,
			UpdatedAt: testTime,
			ParentID:  &parentID,
		},
	}

	mockPM.EXPECT().ListTasksForProject(ctx, projectID).Return(tasks, nil)
	mockPM.EXPECT().GetTasksWithDependencies(ctx, gomock.Any()).Return(tasks, nil)
	mockClient.EXPECT().GetRemoteData(gomock.Any(), gomock.Any()).Return(&shared.SyncDataSet{}, nil).AnyTimes()

	injector := do.New()
	do.ProvideValue[logger.Logger](injector, newTestLogger())
	do.ProvideValue[manager.ProjectManager](injector, mockPM)
	do.ProvideValue[client.RESTSyncClient](injector, mockClient)

	extractor, err := knotsync.NewDataExtractorService(injector)
	require.NoError(t, err)

	data, err := extractor.ExtractLocalData(ctx, projectID)

	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 1, len(data.Tasks))
	assert.Equal(t, parentID, *data.Tasks[taskID].ParentID)
}

// testTime is a fixed time for testing
var testTime = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
