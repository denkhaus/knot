package tools

import (
	"testing"

	"github.com/denkhaus/knot/v2/internal/mocks"
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

