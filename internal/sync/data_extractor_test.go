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

func TestNewDataExtractor(t *testing.T) {
	// Use a simple no-op logger for testing
	log := &noopLogger{}
	extractor := NewDataExtractor(log)

	assert.NotNil(t, extractor)
}

// TestDataExtractor_SetClient tests setting the REST client
func TestDataExtractor_SetClient(t *testing.T) {
	log := &noopLogger{}
	extractor := NewDataExtractor(log)
	projectID := uuid.New()

	// Create mock client
	mockClient := &mockRESTSyncClient{
		projects: map[uuid.UUID]*types.Project{
			projectID: {
				ID:    projectID,
				Title: "Test Project",
				State: types.ProjectStateActive,
			},
		},
		tasks: map[uuid.UUID]*types.Task{},
	}

	// Set client
	extractor.SetClient(mockClient)

	// Client should be set (we can't directly test it as it's not exported)
	assert.NotNil(t, extractor)
}

func TestDataExtractor_ExtractLocalData(t *testing.T) {
	log := &noopLogger{}
	extractor := NewDataExtractor(log)
	ctx := context.Background()
	projectID := uuid.New()

	// Create a mock project manager that implements the required interface
	mockPM := &mockProjectManager{}

	// Test with mock project manager
	data, err := extractor.ExtractLocalData(ctx, mockPM, projectID)

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 0, len(data.Projects))
	assert.Equal(t, 0, len(data.Tasks))
}

func TestDataExtractor_ValidateDataSet_EmptyDataSet(t *testing.T) {
	log := &noopLogger{}
	extractor := NewDataExtractor(log)
	ctx := context.Background()
	dataSet := shared.NewSyncDataSet()

	err := extractor.ValidateDataSet(ctx, dataSet)

	assert.NoError(t, err)
}

func TestDataExtractor_ValidateDataSet_ValidDataSet(t *testing.T) {
	log := &noopLogger{}
	extractor := NewDataExtractor(log)
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

func TestDataExtractor_ValidateDataSet_FutureTimestamp(t *testing.T) {
	log := &noopLogger{}
	extractor := NewDataExtractor(log)
	ctx := context.Background()
	dataSet := shared.NewSyncDataSet()

	// Add project with future timestamp
	projectID := uuid.New()
	futureTime := time.Now().Add(24 * time.Hour)
	project := &types.Project{
		ID:        projectID,
		Title:     "Test Project",
		State:     types.ProjectStateActive,
		CreatedAt: futureTime,
		UpdatedAt: futureTime,
	}
	dataSet.Projects[projectID] = project

	err := extractor.ValidateDataSet(ctx, dataSet)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "future timestamp")
}

func TestDataExtractor_FilterByTimeRange(t *testing.T) {
	log := &noopLogger{}
	extractor := NewDataExtractor(log)
	dataSet := shared.NewSyncDataSet()

	// Add projects with different timestamps
	oldProjectID := uuid.New()
	oldProject := &types.Project{
		ID:        oldProjectID,
		Title:     "Old Project",
		State:     types.ProjectStateActive,
		CreatedAt: testTime.Add(-2 * time.Hour),
		UpdatedAt: testTime.Add(-2 * time.Hour),
	}
	dataSet.Projects[oldProjectID] = oldProject

	newProjectID := uuid.New()
	newProject := &types.Project{
		ID:        newProjectID,
		Title:     "New Project",
		State:     types.ProjectStateActive,
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}
	dataSet.Projects[newProjectID] = newProject

	// Filter by time since 1 hour ago (should only include new project)
	since := testTime.Add(-time.Hour)
	filtered := extractor.FilterByTimeRange(dataSet, since)

	assert.Equal(t, 1, len(filtered.Projects))
	assert.Contains(t, filtered.Projects, newProjectID)
	assert.NotContains(t, filtered.Projects, oldProjectID)
}

func TestDataExtractor_GetStatistics(t *testing.T) {
	log := &noopLogger{}
	extractor := NewDataExtractor(log)
	dataSet := shared.NewSyncDataSet()

	// Add test data
	projectID1 := uuid.New()
	projectID2 := uuid.New()
	dataSet.Projects[projectID1] = &types.Project{
		ID:        projectID1,
		Title:     "Project 1",
		State:     types.ProjectStateActive,
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}
	dataSet.Projects[projectID2] = &types.Project{
		ID:        projectID2,
		Title:     "Project 2",
		State:     types.ProjectStateCompleted,
		CreatedAt: testTime.Add(time.Hour),
		UpdatedAt: testTime.Add(time.Hour),
	}

	taskID1 := uuid.New()
	taskID2 := uuid.New()
	taskID3 := uuid.New()
	dataSet.Tasks[taskID1] = &types.Task{
		ID:        taskID1,
		Title:     "Task 1",
		State:     types.TaskStatePending,
		Priority:  5,
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}
	dataSet.Tasks[taskID2] = &types.Task{
		ID:        taskID2,
		Title:     "Task 2",
		State:     types.TaskStateCompleted,
		Priority:  8,
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}
	dataSet.Tasks[taskID3] = &types.Task{
		ID:        taskID3,
		Title:     "Task 3",
		State:     types.TaskStatePending,
		Priority:  3,
		CreatedAt: testTime,
		UpdatedAt: testTime,
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

// mockProjectManager implements the interface expected by ExtractLocalData
type mockProjectManager struct {
	tasks []*types.Task
}

func (m *mockProjectManager) ListTasksForProject(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error) {
	// Return tasks for testing
	return m.tasks, nil
}

func (m *mockProjectManager) GetTasksWithDependencies(ctx context.Context, taskIDs []uuid.UUID) ([]*types.Task, error) {
	// Return tasks for testing (without loading dependencies for simplicity)
	return m.tasks, nil
}

// TestDataExtractor_ExtractRemoteData_Success tests extracting data from remote server
func TestDataExtractor_ExtractRemoteData_Success(t *testing.T) {
	log := &noopLogger{}
	extractor := NewDataExtractor(log)
	ctx := context.Background()
	projectID := uuid.New()

	// Create mock REST client
	mockClient := &mockRESTSyncClient{
		projects: map[uuid.UUID]*types.Project{
			projectID: {
				ID:        projectID,
				Title:     "Remote Project",
				State:     types.ProjectStateActive,
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
		tasks: map[uuid.UUID]*types.Task{
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

	data, err := extractor.ExtractRemoteData(ctx, mockClient, projectID)

	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 1, len(data.Projects))
	assert.Equal(t, 1, len(data.Tasks))
	assert.Equal(t, "Remote Project", data.Projects[projectID].Title)
}

// TestDataExtractor_ExtractRemoteData_NilClient tests error handling for nil client
func TestDataExtractor_ExtractRemoteData_NilClient(t *testing.T) {
	log := &noopLogger{}
	extractor := NewDataExtractor(log)
	ctx := context.Background()
	projectID := uuid.New()

	data, err := extractor.ExtractRemoteData(ctx, nil, projectID)

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "REST client is required")
}

// TestDataExtractor_ExtractRemoteData_ClientError tests error handling from client
func TestDataExtractor_ExtractRemoteData_ClientError(t *testing.T) {
	log := &noopLogger{}
	extractor := NewDataExtractor(log)
	ctx := context.Background()
	projectID := uuid.New()

	// Create mock REST client that returns an error
	mockClient := &mockRESTSyncClient{
		shouldError: true,
		errMsg:      "connection refused",
	}

	data, err := extractor.ExtractRemoteData(ctx, mockClient, projectID)

	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "connection refused")
}

// TestDataExtractor_ExtractProjectData tests combining local and remote data
func TestDataExtractor_ExtractProjectData(t *testing.T) {
	log := &noopLogger{}
	extractor := NewDataExtractor(log)
	ctx := context.Background()
	projectID := uuid.New()

	// Create mock project manager with local data
	localTaskID := uuid.New()
	mockPM := &mockProjectManager{
		tasks: []*types.Task{
			{
				ID:          localTaskID,
				ProjectID:   projectID,
				Title:       "Local Task",
				State:       types.TaskStatePending,
				Priority:    5,
				Complexity:   3,
				CreatedAt:   testTime,
				UpdatedAt:   testTime,
				UpdatedBy:   "local-user",
				Dependencies: []uuid.UUID{},
			},
		},
	}

	// Create mock REST client with remote data
	remoteProjectID := uuid.New()
	remoteTaskID := uuid.New()
	mockClient := &mockRESTSyncClient{
		projects: map[uuid.UUID]*types.Project{
			remoteProjectID: {
				ID:        remoteProjectID,
				Title:     "Remote Project",
				State:     types.ProjectStateActive,
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
		tasks: map[uuid.UUID]*types.Task{
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

	data, err := extractor.ExtractProjectData(ctx, mockPM, mockClient, projectID)

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
	log := &noopLogger{}
	extractor := NewDataExtractor(log)
	ctx := context.Background()
	projectID := uuid.New()

	taskID := uuid.New()
	mockPM := &mockProjectManager{
		tasks: []*types.Task{
			{
				ID:          taskID,
				ProjectID:   projectID,
				Title:       "Test Task",
				Description: "Test Description",
				State:       types.TaskStateInProgress,
				Priority:    7,
				Complexity:   5,
				CreatedAt:   testTime,
				UpdatedAt:   testTime.Add(time.Hour),
				UpdatedBy:   "test-user",
				Dependencies: []uuid.UUID{},
			},
		},
	}

	data, err := extractor.ExtractLocalData(ctx, mockPM, projectID)

	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 1, len(data.Tasks))
	assert.Contains(t, data.Tasks, taskID)
}

// TestDataExtractor_ExtractLocalData_WithDependencies tests extracting tasks with dependencies
func TestDataExtractor_ExtractLocalData_WithDependencies(t *testing.T) {
	log := &noopLogger{}
	extractor := NewDataExtractor(log)
	ctx := context.Background()
	projectID := uuid.New()

	depID := uuid.New()
	taskID := uuid.New()
	mockPM := &mockProjectManager{
		tasks: []*types.Task{
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
		},
	}

	data, err := extractor.ExtractLocalData(ctx, mockPM, projectID)

	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 1, len(data.Tasks))
	assert.Contains(t, data.Tasks, taskID)
	assert.Equal(t, 1, len(data.Tasks[taskID].Dependencies))
	assert.Contains(t, data.Tasks[taskID].Dependencies, depID)
}

// TestDataExtractor_ExtractLocalData_WithParentID tests extracting tasks with parent ID
func TestDataExtractor_ExtractLocalData_WithParentID(t *testing.T) {
	log := &noopLogger{}
	extractor := NewDataExtractor(log)
	ctx := context.Background()
	projectID := uuid.New()

	parentID := uuid.New()
	taskID := uuid.New()
	mockPM := &mockProjectManager{
		tasks: []*types.Task{
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
		},
	}

	data, err := extractor.ExtractLocalData(ctx, mockPM, projectID)

	require.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, 1, len(data.Tasks))
	assert.Equal(t, parentID, *data.Tasks[taskID].ParentID)
}

// mockRESTSyncClient is a mock implementation of RESTClient interface
type mockRESTSyncClient struct {
	projects    map[uuid.UUID]*types.Project
	tasks       map[uuid.UUID]*types.Task
	shouldError bool
	errMsg      string
}

func (m *mockRESTSyncClient) GetRemoteData(ctx context.Context, projectID *uuid.UUID) (*shared.SyncDataSet, error) {
	if m.shouldError {
		return nil, &testDataExtractorError{msg: m.errMsg}
	}

	return &shared.SyncDataSet{
		Projects: m.projects,
		Tasks:    m.tasks,
	}, nil
}

// testDataExtractorError is a simple error implementation for testing
type testDataExtractorError struct {
	msg string
}

func (e *testDataExtractorError) Error() string {
	return e.msg
}
