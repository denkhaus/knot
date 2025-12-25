package tools

import (
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/mocks"
	knotutils "github.com/denkhaus/knot/v2/internal/utils"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestRegisterStatusTools tests that status tools are registered without errors
func TestRegisterStatusTools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockProjectManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)

	// Create MCP server
	mcpServer := server.NewMCPServer("test", "1.0.0")

	// Register tools - should not panic
	RegisterStatusTools(mcpServer, mockProjectManager, mockSessionManager)

	// Verify server was created
	assert.NotNil(t, mcpServer)
}

// TestStatusRequestStructures tests various request structures
func TestStatusRequestStructures(t *testing.T) {
	t.Run("StatusReadyRequest with limit", func(t *testing.T) {
		limit := 10
		request := StatusReadyRequest{
			Limit: &limit,
		}
		assert.NotNil(t, request.Limit)
		assert.Equal(t, 10, *request.Limit)
	})

	t.Run("StatusReadyRequest without limit", func(t *testing.T) {
		request := StatusReadyRequest{}
		assert.Nil(t, request.Limit)
	})

	t.Run("StatusBlockedRequest with limit", func(t *testing.T) {
		limit := 5
		request := StatusBlockedRequest{
			Limit: &limit,
		}
		assert.NotNil(t, request.Limit)
		assert.Equal(t, 5, *request.Limit)
	})

	t.Run("StatusActionableRequest with limit", func(t *testing.T) {
		limit := 15
		request := StatusActionableRequest{
			Limit: &limit,
		}
		assert.NotNil(t, request.Limit)
		assert.Equal(t, 15, *request.Limit)
	})

	t.Run("StatusProjectRequest", func(t *testing.T) {
		request := StatusProjectRequest{}
		// Empty struct is valid
		assert.NotNil(t, request)
	})

	t.Run("StatusTreeRequest", func(t *testing.T) {
		request := StatusTreeRequest{}
		// Empty struct is valid
		assert.NotNil(t, request)
	})
}

// TestCalculateTaskStats tests the calculateTaskStats helper function
func TestCalculateTaskStats(t *testing.T) {
	now := time.Now()
	projectID := uuid.New()

	tasks := []*types.Task{
		{
			ID:        uuid.New(),
			ProjectID: projectID,
			Title:     "Task 1",
			State:     types.TaskStatePending,
			Priority:  types.TaskPriorityMedium,
			Complexity: 5,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        uuid.New(),
			ProjectID: projectID,
			Title:     "Task 2",
			State:     types.TaskStateInProgress,
			Priority:  types.TaskPriorityHigh,
			Complexity: 7,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        uuid.New(),
			ProjectID: projectID,
			Title:     "Task 3",
			State:     types.TaskStateCompleted,
			Priority:  types.TaskPriorityLow,
			Complexity: 3,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        uuid.New(),
			ProjectID: projectID,
			Title:     "Task 4",
			State:     types.TaskStateBlocked,
			Priority:  types.TaskPriorityMedium,
			Complexity: 5,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	stats := calculateTaskStats(tasks)

	// Verify state counts
	assert.Equal(t, 1, stats.TotalByState["pending"])
	assert.Equal(t, 1, stats.TotalByState["in-progress"])
	assert.Equal(t, 1, stats.TotalByState["completed"])
	assert.Equal(t, 1, stats.TotalByState["blocked"])

	// Verify priority counts
	assert.Equal(t, 1, stats.TotalByPriority["low"])
	assert.Equal(t, 2, stats.TotalByPriority["medium"])
	assert.Equal(t, 1, stats.TotalByPriority["high"])

	// Verify average complexity
	expectedAvg := float64(5+7+3+5) / 4
	assert.Equal(t, expectedAvg, stats.AverageComplexity)
}

// TestCalculateTaskStats_EmptyTasks tests stats with empty task list
func TestCalculateTaskStats_EmptyTasks(t *testing.T) {
	tasks := []*types.Task{}

	stats := calculateTaskStats(tasks)

	assert.Empty(t, stats.TotalByState)
	assert.Empty(t, stats.TotalByPriority)
	assert.Equal(t, 0.0, stats.AverageComplexity)
}

// TestProgressStructure tests the Progress structure
func TestProgressStructure(t *testing.T) {
	progress := Progress{
		TotalTasks:      100,
		CompletedTasks:  40,
		PendingTasks:    30,
		InProgressTasks: 20,
		BlockedTasks:    10,
		CompletionRate:  40.0,
	}

	assert.Equal(t, 100, progress.TotalTasks)
	assert.Equal(t, 40, progress.CompletedTasks)
	assert.Equal(t, 30, progress.PendingTasks)
	assert.Equal(t, 20, progress.InProgressTasks)
	assert.Equal(t, 10, progress.BlockedTasks)
	assert.Equal(t, 40.0, progress.CompletionRate)
}

// TestTaskStatsStructure tests the TaskStats structure
func TestTaskStatsStructure(t *testing.T) {
	stats := TaskStats{
		TotalByState: map[string]int{
			"pending":     10,
			"in_progress": 5,
			"completed":   15,
		},
		TotalByPriority: map[string]int{
			"low":    5,
			"medium": 15,
			"high":   10,
		},
		AverageComplexity: 5.5,
	}

	assert.Equal(t, 3, len(stats.TotalByState))
	assert.Equal(t, 30, stats.TotalByState["pending"]+stats.TotalByState["in_progress"]+stats.TotalByState["completed"])
	assert.Equal(t, 3, len(stats.TotalByPriority))
	assert.Equal(t, 5.5, stats.AverageComplexity)
}

// TestBlockedTaskInfoStructure tests the BlockedTaskInfo structure
func TestBlockedTaskInfoStructure(t *testing.T) {
	taskID := uuid.New()

	blockedInfo := BlockedTaskInfo{
		TaskInfo: TaskInfo{
			ID:         taskID.String(),
			Title:      "Blocked Task",
			State:      "blocked",
			Priority:   "high",
			Complexity: 7,
		},
		BlockingReasons:   []string{"Waiting for dependency", "Waiting for review"},
		DependenciesCount: 2,
	}

	assert.Equal(t, taskID.String(), blockedInfo.ID)
	assert.Equal(t, "Blocked Task", blockedInfo.Title)
	assert.Equal(t, "blocked", blockedInfo.State)
	assert.Equal(t, 2, len(blockedInfo.BlockingReasons))
	assert.Equal(t, 2, blockedInfo.DependenciesCount)
	assert.Contains(t, blockedInfo.BlockingReasons[0], "dependency")
}

// TestTreeNodeStructure tests the TreeNode structure
func TestTreeNodeStructure(t *testing.T) {
	taskID := uuid.New()
	childID := uuid.New()

	treeNode := TreeNode{
		TaskInfo: TaskInfo{
			ID:         taskID.String(),
			Title:      "Parent Task",
			State:      "in_progress",
			Priority:   "medium",
			Complexity: 5,
		},
		Level: 0,
		Children: []TreeNode{
			{
				TaskInfo: TaskInfo{
					ID:         childID.String(),
					Title:      "Child Task",
					State:      "pending",
					Priority:   "low",
					Complexity: 3,
				},
				Level: 1,
			},
		},
	}

	assert.Equal(t, taskID.String(), treeNode.ID)
	assert.Equal(t, "Parent Task", treeNode.Title)
	assert.Equal(t, 0, treeNode.Level)
	assert.Len(t, treeNode.Children, 1)
	assert.Equal(t, childID.String(), treeNode.Children[0].ID)
	assert.Equal(t, 1, treeNode.Children[0].Level)
}

// TestStatusResponseStructures tests various response structures
func TestStatusResponseStructures(t *testing.T) {
	projectID := uuid.New()
	taskID := uuid.New()

	t.Run("StatusReadyResponse", func(t *testing.T) {
		response := StatusReadyResponse{
			ProjectID: projectID.String(),
			Tasks: []TaskInfo{
				{
					ID:         taskID.String(),
					Title:      "Ready Task",
					State:      "pending",
					Priority:   "medium",
					Complexity: 5,
				},
			},
			Total:   1,
			Message: "Found 1 ready task",
		}

		assert.Equal(t, projectID.String(), response.ProjectID)
		assert.Len(t, response.Tasks, 1)
		assert.Equal(t, 1, response.Total)
		assert.Contains(t, response.Message, "ready task")
	})

	t.Run("StatusActionableResponse", func(t *testing.T) {
		response := StatusActionableResponse{
			ProjectID: projectID.String(),
			Tasks: []TaskInfo{
				{
					ID:         taskID.String(),
					Title:      "Actionable Task",
					State:      "pending",
					Priority:   "high",
					Complexity: 7,
				},
			},
			Total:   1,
			Message: "Found 1 actionable task",
		}

		assert.Equal(t, projectID.String(), response.ProjectID)
		assert.Len(t, response.Tasks, 1)
		assert.Equal(t, 1, response.Total)
		assert.Contains(t, response.Message, "actionable")
	})

	t.Run("StatusProjectResponse", func(t *testing.T) {
		response := StatusProjectResponse{
			ProjectInfo: ProjectInfo{
				ID:          projectID.String(),
				Title:       "Test Project",
				Description: "Test Description",
				State:       "active",
				CreatedAt:   "2024-01-01T00:00:00Z",
			},
			Progress: Progress{
				TotalTasks:      10,
				CompletedTasks:  5,
				CompletionRate:  50.0,
			},
			TaskStats: TaskStats{
				TotalByState: map[string]int{
					"pending": 3,
					"completed": 5,
				},
				AverageComplexity: 5.0,
			},
			Message: "Project status for Test Project",
		}

		assert.Equal(t, projectID.String(), response.ProjectInfo.ID)
		assert.Equal(t, "Test Project", response.ProjectInfo.Title)
		assert.Equal(t, 10, response.Progress.TotalTasks)
		assert.Equal(t, 5, response.Progress.CompletedTasks)
		assert.Equal(t, 50.0, response.Progress.CompletionRate)
		assert.Equal(t, 2, len(response.TaskStats.TotalByState))
	})

	t.Run("StatusBlockedResponse", func(t *testing.T) {
		response := StatusBlockedResponse{
			ProjectID: projectID.String(),
			Tasks: []BlockedTaskInfo{
				{
					TaskInfo: TaskInfo{
						ID:         taskID.String(),
						Title:      "Blocked Task",
						State:      "blocked",
						Priority:   "medium",
						Complexity: 5,
					},
					BlockingReasons:   []string{"Waiting for dependency"},
					DependenciesCount: 1,
				},
			},
			Total:   1,
			Message: "Found 1 blocked task",
		}

		assert.Equal(t, projectID.String(), response.ProjectID)
		assert.Len(t, response.Tasks, 1)
		assert.Equal(t, 1, response.Total)
		assert.Contains(t, response.Message, "blocked")
	})
}

// TestConvertToMCPtreeNodes tests the convertToMCPtreeNodes helper function
func TestConvertToMCPtreeNodes(t *testing.T) {
	now := time.Now()
	projectID := uuid.New()
	parentID := uuid.New()
	childID := uuid.New()

	knotNodes := []knotutils.TaskTreeNode{
		{
			Task: &types.Task{
				ID:        parentID,
				ProjectID: projectID,
				Title:     "Parent Task",
				State:     types.TaskStateCompleted,
				Priority:  types.TaskPriorityMedium,
				Complexity: 5,
				CreatedAt: now,
				UpdatedAt: now,
			},
			Level: 0,
			Children: []knotutils.TaskTreeNode{
				{
					Task: &types.Task{
						ID:        childID,
						ProjectID: projectID,
						Title:     "Child Task",
						State:     types.TaskStatePending,
						Priority:  types.TaskPriorityLow,
						Complexity: 3,
						CreatedAt: now,
						UpdatedAt: now,
					},
					Level: 1,
					Children: []knotutils.TaskTreeNode{},
				},
			},
		},
	}

	result := convertToMCPtreeNodes(knotNodes)

	assert.Len(t, result, 1)
	assert.Equal(t, parentID.String(), result[0].ID)
	assert.Equal(t, "Parent Task", result[0].Title)
	assert.Equal(t, 0, result[0].Level)
	assert.Len(t, result[0].Children, 1)
	assert.Equal(t, childID.String(), result[0].Children[0].ID)
	assert.Equal(t, "Child Task", result[0].Children[0].Title)
	assert.Equal(t, 1, result[0].Children[0].Level)
}
