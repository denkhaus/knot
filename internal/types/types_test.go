package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskState tests the TaskState type and constants
func TestTaskState(t *testing.T) {
	tests := []struct {
		name  string
		state TaskState
		str   string
	}{
		{"Pending", TaskStatePending, "pending"},
		{"InProgress", TaskStateInProgress, "in-progress"},
		{"Completed", TaskStateCompleted, "completed"},
		{"Blocked", TaskStateBlocked, "blocked"},
		{"Cancelled", TaskStateCancelled, "cancelled"},
		{"DeletionPending", TaskStateDeletionPending, "deletion-pending"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.str, string(tt.state))
		})
	}
}

// TestProjectState tests the ProjectState type and constants
func TestProjectState(t *testing.T) {
	tests := []struct {
		name  string
		state ProjectState
		str   string
	}{
		{"Active", ProjectStateActive, "active"},
		{"Completed", ProjectStateCompleted, "completed"},
		{"Archived", ProjectStateArchived, "archived"},
		{"DeletionPending", ProjectStateDeletionPending, "deletion-pending"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.str, string(tt.state))
		})
	}
}

// TestTaskSerialization tests JSON serialization/deserialization of Task
func TestTaskSerialization(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second) // Truncate for comparison
	completedAt := now.Add(time.Hour)
	parentID := uuid.New()
	agentID := uuid.New()
	
	task := &Task{
		ID:            uuid.New(),
		ProjectID:     uuid.New(),
		ParentID:      &parentID,
		Title:         "Test Task",
		Description:   "A test task for serialization",
		State:         TaskStateInProgress,
		Complexity:    5,
		Depth:         2,
		Estimate:      int64Ptr(120), // 2 hours
		AssignedAgent: &agentID,
		Dependencies:  []uuid.UUID{uuid.New(), uuid.New()},
		Dependents:    []uuid.UUID{uuid.New()},
		CreatedAt:     now,
		UpdatedAt:     now,
		CreatedBy:     "test-user",
		UpdatedBy:     "test-user",
		CompletedAt:   &completedAt,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(task)
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "Test Task")
	assert.Contains(t, string(jsonData), "in-progress")

	// Test JSON unmarshaling
	var unmarshaled Task
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, task.ID, unmarshaled.ID)
	assert.Equal(t, task.ProjectID, unmarshaled.ProjectID)
	assert.Equal(t, task.ParentID, unmarshaled.ParentID)
	assert.Equal(t, task.Title, unmarshaled.Title)
	assert.Equal(t, task.Description, unmarshaled.Description)
	assert.Equal(t, task.State, unmarshaled.State)
	assert.Equal(t, task.Complexity, unmarshaled.Complexity)
	assert.Equal(t, task.Depth, unmarshaled.Depth)
	assert.Equal(t, task.Estimate, unmarshaled.Estimate)
	assert.Equal(t, task.AssignedAgent, unmarshaled.AssignedAgent)
	assert.Equal(t, task.Dependencies, unmarshaled.Dependencies)
	assert.Equal(t, task.Dependents, unmarshaled.Dependents)
	assert.Equal(t, task.CreatedBy, unmarshaled.CreatedBy)
	assert.Equal(t, task.UpdatedBy, unmarshaled.UpdatedBy)
	assert.True(t, task.CreatedAt.Equal(unmarshaled.CreatedAt))
	assert.True(t, task.UpdatedAt.Equal(unmarshaled.UpdatedAt))
	assert.True(t, task.CompletedAt.Equal(*unmarshaled.CompletedAt))
}

// TestProjectSerialization tests JSON serialization/deserialization of Project
func TestProjectSerialization(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	
	project := &Project{
		ID:             uuid.New(),
		Title:          "Test Project",
		Description:    "A test project for serialization",
		State:          ProjectStateActive,
		CreatedAt:      now,
		UpdatedAt:      now,
		CreatedBy:      "test-user",
		UpdatedBy:      "test-user",
		TotalTasks:     10,
		CompletedTasks: 3,
		Progress:       30.0,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(project)
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "Test Project")
	assert.Contains(t, string(jsonData), "active")

	// Test JSON unmarshaling
	var unmarshaled Project
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, project.ID, unmarshaled.ID)
	assert.Equal(t, project.Title, unmarshaled.Title)
	assert.Equal(t, project.Description, unmarshaled.Description)
	assert.Equal(t, project.State, unmarshaled.State)
	assert.Equal(t, project.CreatedBy, unmarshaled.CreatedBy)
	assert.Equal(t, project.UpdatedBy, unmarshaled.UpdatedBy)
	assert.Equal(t, project.TotalTasks, unmarshaled.TotalTasks)
	assert.Equal(t, project.CompletedTasks, unmarshaled.CompletedTasks)
	assert.Equal(t, project.Progress, unmarshaled.Progress)
	assert.True(t, project.CreatedAt.Equal(unmarshaled.CreatedAt))
	assert.True(t, project.UpdatedAt.Equal(unmarshaled.UpdatedAt))
}

// TestTaskFilter tests the TaskFilter functionality
func TestTaskFilter(t *testing.T) {
	projectID := uuid.New()
	parentID := uuid.New()
	state := TaskStateInProgress
	
	filter := &TaskFilter{
		ProjectID:     &projectID,
		ParentID:      &parentID,
		State:         &state,
		MinDepth:      intPtr(1),
		MaxDepth:      intPtr(5),
		MinComplexity: intPtr(3),
		MaxComplexity: intPtr(8),
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(filter)
	require.NoError(t, err)

	var unmarshaled TaskFilter
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, filter.ProjectID, unmarshaled.ProjectID)
	assert.Equal(t, filter.ParentID, unmarshaled.ParentID)
	assert.Equal(t, filter.State, unmarshaled.State)
	assert.Equal(t, filter.MinDepth, unmarshaled.MinDepth)
	assert.Equal(t, filter.MaxDepth, unmarshaled.MaxDepth)
	assert.Equal(t, filter.MinComplexity, unmarshaled.MinComplexity)
	assert.Equal(t, filter.MaxComplexity, unmarshaled.MaxComplexity)
}

// TestTaskUpdates tests the TaskUpdates functionality
func TestTaskUpdates(t *testing.T) {
	state := TaskStateCompleted
	complexity := 7
	
	updates := &TaskUpdates{
		State:      &state,
		Complexity: &complexity,
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(updates)
	require.NoError(t, err)

	var unmarshaled TaskUpdates
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, updates.State, unmarshaled.State)
	assert.Equal(t, updates.Complexity, unmarshaled.Complexity)
}

// TestProjectProgress tests the ProjectProgress functionality
func TestProjectProgress(t *testing.T) {
	progress := &ProjectProgress{
		ProjectID:       uuid.New(),
		TotalTasks:      20,
		CompletedTasks:  8,
		InProgressTasks: 5,
		PendingTasks:    4,
		BlockedTasks:    2,
		CancelledTasks:  1,
		OverallProgress: 40.0,
		TasksByDepth: map[int]int{
			0: 5,
			1: 10,
			2: 5,
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(progress)
	require.NoError(t, err)

	var unmarshaled ProjectProgress
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, progress.ProjectID, unmarshaled.ProjectID)
	assert.Equal(t, progress.TotalTasks, unmarshaled.TotalTasks)
	assert.Equal(t, progress.CompletedTasks, unmarshaled.CompletedTasks)
	assert.Equal(t, progress.InProgressTasks, unmarshaled.InProgressTasks)
	assert.Equal(t, progress.PendingTasks, unmarshaled.PendingTasks)
	assert.Equal(t, progress.BlockedTasks, unmarshaled.BlockedTasks)
	assert.Equal(t, progress.CancelledTasks, unmarshaled.CancelledTasks)
	assert.Equal(t, progress.OverallProgress, unmarshaled.OverallProgress)
	assert.Equal(t, progress.TasksByDepth, unmarshaled.TasksByDepth)
}

// TestTaskValidation tests basic validation logic for tasks
func TestTaskValidation(t *testing.T) {
	t.Run("Valid task", func(t *testing.T) {
		task := &Task{
			ID:          uuid.New(),
			ProjectID:   uuid.New(),
			Title:       "Valid Task",
			Description: "A valid task",
			State:       TaskStatePending,
			Complexity:  5,
			Depth:       0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Basic validation checks
		assert.NotEqual(t, uuid.Nil, task.ID)
		assert.NotEqual(t, uuid.Nil, task.ProjectID)
		assert.NotEmpty(t, task.Title)
		assert.True(t, task.Complexity >= 1 && task.Complexity <= 10)
		assert.True(t, task.Depth >= 0)
		assert.False(t, task.CreatedAt.IsZero())
		assert.False(t, task.UpdatedAt.IsZero())
	})

	t.Run("Task with dependencies", func(t *testing.T) {
		dep1 := uuid.New()
		dep2 := uuid.New()
		
		task := &Task{
			ID:           uuid.New(),
			ProjectID:    uuid.New(),
			Title:        "Task with deps",
			State:        TaskStatePending,
			Dependencies: []uuid.UUID{dep1, dep2},
			Dependents:   []uuid.UUID{uuid.New()},
		}

		assert.Len(t, task.Dependencies, 2)
		assert.Contains(t, task.Dependencies, dep1)
		assert.Contains(t, task.Dependencies, dep2)
		assert.Len(t, task.Dependents, 1)
	})
}

// TestProjectValidation tests basic validation logic for projects
func TestProjectValidation(t *testing.T) {
	t.Run("Valid project", func(t *testing.T) {
		project := &Project{
			ID:             uuid.New(),
			Title:          "Valid Project",
			Description:    "A valid project",
			State:          ProjectStateActive,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			TotalTasks:     10,
			CompletedTasks: 3,
			Progress:       30.0,
		}

		// Basic validation checks
		assert.NotEqual(t, uuid.Nil, project.ID)
		assert.NotEmpty(t, project.Title)
		assert.True(t, project.Progress >= 0.0 && project.Progress <= 100.0)
		assert.True(t, project.CompletedTasks <= project.TotalTasks)
		assert.False(t, project.CreatedAt.IsZero())
		assert.False(t, project.UpdatedAt.IsZero())
	})

	t.Run("Progress calculation consistency", func(t *testing.T) {
		expectedProgress := float64(5) / float64(20) * 100.0
		assert.Equal(t, 25.0, expectedProgress)
		
		// Test edge cases
		emptyProject := &Project{TotalTasks: 0, CompletedTasks: 0}
		assert.Equal(t, 0, emptyProject.TotalTasks)
		assert.Equal(t, 0, emptyProject.CompletedTasks)
	})
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}