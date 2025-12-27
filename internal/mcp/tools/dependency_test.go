package tools

import (
	"context"
	"testing"

	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestRegisterDependencyTools tests that dependency tools are registered without errors
func TestRegisterDependencyTools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockProjectManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)

	// Create MCP server
	mcpServer := server.NewMCPServer("test", "1.0.0")

	// Register tools - should not panic
	RegisterDependencyTools(mcpServer, mockProjectManager, mockSessionManager)

	// Verify server was created
	assert.NotNil(t, mcpServer)
}

// TestDependencyRequestStructures tests dependency request structures
func TestDependencyRequestStructures(t *testing.T) {
	t.Run("DependencyAddRequest", func(t *testing.T) {
		request := DependencyAddRequest{
			TaskID:          uuid.New().String(),
			DependsOnTaskID: uuid.New().String(),
		}

		// Verify UUIDs are valid
		_, err := uuid.Parse(request.TaskID)
		assert.NoError(t, err, "TaskID should be valid UUID")

		_, err = uuid.Parse(request.DependsOnTaskID)
		assert.NoError(t, err, "DependsOnTaskID should be valid UUID")
	})

	t.Run("DependencyRemoveRequest", func(t *testing.T) {
		request := DependencyRemoveRequest{
			TaskID:          uuid.New().String(),
			DependsOnTaskID: uuid.New().String(),
		}

		// Verify UUIDs are valid
		_, err := uuid.Parse(request.TaskID)
		assert.NoError(t, err, "TaskID should be valid UUID")

		_, err = uuid.Parse(request.DependsOnTaskID)
		assert.NoError(t, err, "DependsOnTaskID should be valid UUID")
	})

	t.Run("DependencyListRequest", func(t *testing.T) {
		taskID := uuid.New()
		request := DependencyListRequest{
			TaskID: taskID.String(),
		}

		assert.Equal(t, taskID.String(), request.TaskID)
	})
}

// TestDependencyResponseStructures tests dependency response structures
func TestDependencyResponseStructures(t *testing.T) {
	taskID := uuid.New()
	depID := uuid.New()

	t.Run("DependencyAddResponse", func(t *testing.T) {
		response := DependencyAddResponse{
			Message:         "Dependency added successfully",
			TaskID:          taskID.String(),
			DependsOnTaskID: depID.String(),
			UpdatedTask: TaskDetails{
				ID:         taskID.String(),
				Title:      "Test Task",
				State:      "pending",
				Priority:   "medium",
				Complexity: 5,
			},
		}

		assert.NotEmpty(t, response.Message)
		assert.Equal(t, taskID.String(), response.TaskID)
		assert.Equal(t, depID.String(), response.DependsOnTaskID)
		assert.Equal(t, "Test Task", response.UpdatedTask.Title)
	})

	t.Run("DependencyRemoveResponse", func(t *testing.T) {
		response := DependencyRemoveResponse{
			Message:         "Dependency removed successfully",
			TaskID:          taskID.String(),
			DependsOnTaskID: depID.String(),
			UpdatedTask: TaskDetails{
				ID:         taskID.String(),
				Title:      "Test Task",
				State:      "pending",
				Priority:   "medium",
				Complexity: 5,
			},
		}

		assert.NotEmpty(t, response.Message)
		assert.Contains(t, response.Message, "removed")
		assert.Equal(t, taskID.String(), response.TaskID)
		assert.Equal(t, depID.String(), response.DependsOnTaskID)
	})

	t.Run("DependencyListResponse", func(t *testing.T) {
		response := DependencyListResponse{
			TaskID: taskID.String(),
			Dependencies: []TaskInfo{
				{
					ID:         depID.String(),
					Title:      "Dependency Task",
					State:      "pending",
					Priority:   "medium",
					Complexity: 5,
				},
			},
			Total:   1,
			Message: "Found 1 dependency",
		}

		assert.Equal(t, taskID.String(), response.TaskID)
		assert.Len(t, response.Dependencies, 1)
		assert.Equal(t, 1, response.Total)
		assert.Equal(t, depID.String(), response.Dependencies[0].ID)
	})
}

// TestDependencyCheckRequest tests the dependency check request structure
func TestDependencyCheckRequest(t *testing.T) {
	taskID := uuid.New()
	depID := uuid.New()

	request := DependencyCheckRequest{
		TaskID:          taskID.String(),
		DependsOnTaskID: depID.String(),
	}

	// Verify UUIDs are valid
	_, err := uuid.Parse(request.TaskID)
	assert.NoError(t, err, "TaskID should be valid UUID")

	_, err = uuid.Parse(request.DependsOnTaskID)
	assert.NoError(t, err, "DependsOnTaskID should be valid UUID")
}

// TestDependencyCheckResponse tests the dependency check response structure
func TestDependencyCheckResponse(t *testing.T) {
	taskID := uuid.New()
	depID := uuid.New()

	t.Run("Circular dependency detected", func(t *testing.T) {
		response := DependencyCheckResponse{
			IsCircular: true,
			Message:    "Adding this dependency would create a circular relationship",
			Path:       []string{taskID.String(), depID.String(), taskID.String()},
		}

		assert.True(t, response.IsCircular)
		assert.NotEmpty(t, response.Message)
		assert.Len(t, response.Path, 3)
		assert.Contains(t, response.Message, "circular")
	})

	t.Run("No circular dependency", func(t *testing.T) {
		response := DependencyCheckResponse{
			IsCircular: false,
			Message:    "No circular dependencies detected",
			Path:       []string{},
		}

		assert.False(t, response.IsCircular)
		assert.NotEmpty(t, response.Message)
		assert.Empty(t, response.Path)
		assert.Contains(t, response.Message, "No circular")
	})
}

// TestCheckCircularDependency tests the circular dependency detection function
func TestCheckCircularDependency(t *testing.T) {
	ctx := context.Background()

	t.Run("detects self-dependency", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProjectManager := mocks.NewMockProjectManager(ctrl)
		taskA := uuid.New()

		isCircular, path := checkCircularDependency(ctx, mockProjectManager, taskA, taskA)

		assert.True(t, isCircular)
		assert.Len(t, path, 2)
		assert.Equal(t, taskA, path[0])
		assert.Equal(t, taskA, path[1])
	})

	t.Run("returns false when GetTaskDependencies fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProjectManager := mocks.NewMockProjectManager(ctrl)
		taskA := uuid.New()
		taskB := uuid.New()

		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskA).
			Return(nil, assert.AnError)

		isCircular, path := checkCircularDependency(ctx, mockProjectManager, taskA, taskB)

		assert.False(t, isCircular)
		assert.Nil(t, path)
	})

	t.Run("returns false for no circular dependency", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProjectManager := mocks.NewMockProjectManager(ctrl)
		taskA := uuid.New()
		taskB := uuid.New()

		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskA).
			Return([]*types.Task{}, nil).AnyTimes()
		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskB).
			Return([]*types.Task{}, nil).AnyTimes()

		isCircular, path := checkCircularDependency(ctx, mockProjectManager, taskA, taskB)

		assert.False(t, isCircular)
		assert.Nil(t, path)
	})

	t.Run("detects direct circular dependency (A -> B -> A)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProjectManager := mocks.NewMockProjectManager(ctrl)
		taskA := uuid.New()
		taskB := uuid.New()
		taskBDependsOnA := []*types.Task{{ID: taskA}}

		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskA).
			Return([]*types.Task{}, nil)
		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskB).
			Return(taskBDependsOnA, nil)

		isCircular, path := checkCircularDependency(ctx, mockProjectManager, taskA, taskB)

		assert.True(t, isCircular)
		assert.Len(t, path, 3)
		assert.Equal(t, taskA, path[0])
		assert.Equal(t, taskB, path[1])
		assert.Equal(t, taskA, path[2])
	})

	t.Run("returns false when no dependencies exist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProjectManager := mocks.NewMockProjectManager(ctrl)
		taskA := uuid.New()
		taskB := uuid.New()
		taskBNoDeps := []*types.Task{}

		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskA).
			Return([]*types.Task{}, nil).AnyTimes()
		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskB).
			Return(taskBNoDeps, nil).AnyTimes()

		isCircular, path := checkCircularDependency(ctx, mockProjectManager, taskA, taskB)

		assert.False(t, isCircular)
		assert.Nil(t, path)
	})
}

// TestFormatCircularPath tests the circular path formatting function
func TestFormatCircularPath(t *testing.T) {
	t.Run("formats path correctly", func(t *testing.T) {
		id1 := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		id2 := uuid.MustParse("00000000-0000-0000-0000-000000000002")
		id3 := uuid.MustParse("00000000-0000-0000-0000-000000000003")

		path := []uuid.UUID{id1, id2, id3}
		result := formatCircularPath(path)

		assert.Contains(t, result, "00000000")
		assert.Contains(t, result, " -> ")
	})

	t.Run("returns unknown for empty path", func(t *testing.T) {
		result := formatCircularPath([]uuid.UUID{})
		assert.Equal(t, "unknown cycle", result)
	})

	t.Run("handles single element path", func(t *testing.T) {
		id1 := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		path := []uuid.UUID{id1}
		result := formatCircularPath(path)

		assert.Contains(t, result, "00000000")
		assert.NotContains(t, result, " -> ")
	})
}

// TestHasPathTo tests the recursive path detection function
func TestHasPathTo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectManager := mocks.NewMockProjectManager(ctrl)
	ctx := context.Background()

	taskA := uuid.New()
	taskB := uuid.New()
	taskC := uuid.New()

	t.Run("returns false when GetTaskDependencies fails", func(t *testing.T) {
		visited := make(map[uuid.UUID]bool)
		path := []uuid.UUID{}

		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskA).
			Return(nil, assert.AnError)

		result := hasPathTo(ctx, mockProjectManager, taskA, taskB, visited, &path)

		assert.False(t, result)
	})

	t.Run("returns false when no dependencies", func(t *testing.T) {
		visited := make(map[uuid.UUID]bool)
		path := []uuid.UUID{}

		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskA).
			Return([]*types.Task{}, nil)

		result := hasPathTo(ctx, mockProjectManager, taskA, taskB, visited, &path)

		assert.False(t, result)
	})

	t.Run("detects direct path", func(t *testing.T) {
		visited := make(map[uuid.UUID]bool)
		path := []uuid.UUID{}

		taskADeps := []*types.Task{{ID: taskB}}
		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskA).
			Return(taskADeps, nil)

		result := hasPathTo(ctx, mockProjectManager, taskA, taskB, visited, &path)

		assert.True(t, result)
		assert.Contains(t, path, taskB)
	})

	t.Run("prevents infinite recursion with visited map", func(t *testing.T) {
		visited := make(map[uuid.UUID]bool)
		visited[taskA] = true // Mark as already visited
		path := []uuid.UUID{}

		// Should not call GetTaskDependencies since taskA is already visited
		result := hasPathTo(ctx, mockProjectManager, taskA, taskB, visited, &path)

		assert.False(t, result)
	})

	t.Run("detects indirect path through dependencies", func(t *testing.T) {
		visited := make(map[uuid.UUID]bool)
		path := []uuid.UUID{}

		// taskA depends on taskC
		taskADeps := []*types.Task{{ID: taskC}}
		// taskC depends on taskB
		taskCDeps := []*types.Task{{ID: taskB}}

		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskA).
			Return(taskADeps, nil)
		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskC).
			Return(taskCDeps, nil)

		result := hasPathTo(ctx, mockProjectManager, taskA, taskB, visited, &path)

		assert.True(t, result)
		assert.Contains(t, path, taskC)
		assert.Contains(t, path, taskB)
	})
}
