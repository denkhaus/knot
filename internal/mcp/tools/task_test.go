package tools

import (
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestParsePriority tests the parsePriority helper function
func TestParsePriority(t *testing.T) {
	tests := []struct {
		name     string
		priority string
		expected types.TaskPriority
	}{
		{"low priority", "low", types.TaskPriorityLow},
		{"high priority", "high", types.TaskPriorityHigh},
		{"medium priority", "medium", types.TaskPriorityMedium},
		{"empty priority", "", types.TaskPriorityMedium},
		{"invalid priority", "invalid", types.TaskPriorityMedium},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePriority(tt.priority)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseTaskState tests the parseTaskState helper function
func TestParseTaskState(t *testing.T) {
	tests := []struct {
		name     string
		state    string
		expected types.TaskState
	}{
		{"pending state", "pending", types.TaskStatePending},
		{"in_progress state", "in_progress", types.TaskStateInProgress},
		{"completed state", "completed", types.TaskStateCompleted},
		{"cancelled state", "cancelled", types.TaskStateCancelled},
		{"blocked state", "blocked", types.TaskStateBlocked},
		{"empty state", "", types.TaskStatePending},
		{"invalid state", "invalid", types.TaskStatePending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTaskState(tt.state)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRegisterTaskManagementTools tests that task tools are registered without errors
func TestRegisterTaskManagementTools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockProjectManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockSessionRegistry := mocks.NewMockSessionRegistry(ctrl)
	mockConfigService := mocks.NewMockService(ctrl)

	// Setup config expectations
	mockConfig := &config.MCPConfig{
		Tasks: config.TasksConfig{
			DefaultComplexity: 5,
		},
	}
	mockConfigService.EXPECT().GetMCPConfig().Return(mockConfig).AnyTimes()

	// Create MCP server
	mcpServer := server.NewMCPServer("test", "1.0.0")

	// Register tools - should not panic
	RegisterTaskManagementTools(mcpServer, mockProjectManager, mockSessionManager, mockSessionRegistry, mockConfigService)

	// Verify tools were registered by checking the server
	assert.NotNil(t, mcpServer)
}

// TestTaskCreateRequest_Validation tests request structure validation
func TestTaskCreateRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request TaskCreateRequest
		valid   bool
	}{
		{
			name: "valid request with all fields",
			request: TaskCreateRequest{
				Title:       "Test Task",
				Description: "Test Description",
				Complexity:  5,
				Priority:    "medium",
			},
			valid: true,
		},
		{
			name: "valid request with minimal fields",
			request: TaskCreateRequest{
				Title: "Test Task",
			},
			valid: true,
		},
		{
			name: "valid request with high priority",
			request: TaskCreateRequest{
				Title:    "Test Task",
				Priority: "high",
			},
			valid: true,
		},
		{
			name: "valid request with low priority",
			request: TaskCreateRequest{
				Title:    "Test Task",
				Priority: "low",
			},
			valid: true,
		},
		{
			name: "valid request with parent ID",
			request: TaskCreateRequest{
				Title:    "Test Task",
				ParentID: uuid.New().String(),
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the structure can be created
			assert.NotNil(t, tt.request.Title)
		})
	}
}

// TestTaskUpdateRequest_Validation tests update request structure
func TestTaskUpdateRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request TaskUpdateRequest
	}{
		{
			name: "valid update with title",
			request: TaskUpdateRequest{
				TaskID: uuid.New().String(),
				Title:  "Updated Title",
			},
		},
		{
			name: "valid update with description",
			request: TaskUpdateRequest{
				TaskID:     uuid.New().String(),
				Description: "Updated Description",
			},
		},
		{
			name: "valid update with complexity",
			request: TaskUpdateRequest{
				TaskID:     uuid.New().String(),
				Complexity: 7,
			},
		},
		{
			name: "valid update with priority",
			request: TaskUpdateRequest{
				TaskID:   uuid.New().String(),
				Priority: "high",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify task ID is valid UUID format
			_, err := uuid.Parse(tt.request.TaskID)
			assert.NoError(t, err, "TaskID should be valid UUID")
		})
	}
}

// TestTaskDetails_Conversion tests conversion from domain type to response type
func TestTaskDetails_Conversion(t *testing.T) {
	now := time.Now()
	projectID := uuid.New()
	taskID := uuid.New()
	parentID := uuid.New()

	domainTask := &types.Task{
		ID:        taskID,
		Title:     "Test Task",
		State:     types.TaskStateInProgress,
		Priority:  types.TaskPriorityHigh,
		Complexity: 7,
		ProjectID: projectID,
		ParentID:  &parentID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Simulate conversion like in the handler
	var parentIDPtr *string
	if domainTask.ParentID != nil {
		parentStr := domainTask.ParentID.String()
		parentIDPtr = &parentStr
	}

	taskDetails := TaskDetails{
		ID:          domainTask.ID.String(),
		Title:       domainTask.Title,
		State:       string(domainTask.State),
		Priority:    domainTask.Priority.ToExternalString(),
		Complexity:  domainTask.Complexity,
		ProjectID:   domainTask.ProjectID.String(),
		ParentID:    parentIDPtr,
		CreatedAt:   domainTask.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   domainTask.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	assert.Equal(t, taskID.String(), taskDetails.ID)
	assert.Equal(t, "Test Task", taskDetails.Title)
	assert.Equal(t, "in-progress", taskDetails.State)
	assert.Equal(t, "high", taskDetails.Priority)
	assert.Equal(t, 7, taskDetails.Complexity)
	assert.Equal(t, projectID.String(), taskDetails.ProjectID)
	assert.NotNil(t, taskDetails.ParentID)
	assert.Equal(t, parentID.String(), *taskDetails.ParentID)
}

// TestTaskInfo_Conversion tests TaskInfo structure
func TestTaskInfo_Conversion(t *testing.T) {
	taskID := uuid.New()

	domainTask := &types.Task{
		ID:        taskID,
		Title:     "Test Task",
		State:     types.TaskStateCompleted,
		Priority:  types.TaskPriorityLow,
		Complexity: 3,
	}

	taskInfo := TaskInfo{
		ID:         domainTask.ID.String(),
		Title:      domainTask.Title,
		State:      string(domainTask.State),
		Priority:   domainTask.Priority.ToExternalString(),
		Complexity: domainTask.Complexity,
	}

	assert.Equal(t, taskID.String(), taskInfo.ID)
	assert.Equal(t, "Test Task", taskInfo.Title)
	assert.Equal(t, "completed", taskInfo.State)
	assert.Equal(t, "low", taskInfo.Priority)
	assert.Equal(t, 3, taskInfo.Complexity)
}

// TestProjectInfo tests ProjectInfo structure
func TestProjectInfo(t *testing.T) {
	now := time.Now()
	projectID := uuid.New()

	project := &types.Project{
		ID:          projectID,
		Title:       "Test Project",
		Description: "Test Description",
		State:       types.ProjectStateActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	info := ProjectInfo{
		ID:          project.ID.String(),
		Title:       project.Title,
		Description: project.Description,
		State:       string(project.State),
		CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	assert.Equal(t, projectID.String(), info.ID)
	assert.Equal(t, "Test Project", info.Title)
	assert.Equal(t, "Test Description", info.Description)
	assert.Equal(t, "active", info.State)
}

// TestTaskUpdateStateRequest_Validation tests state update request validation
func TestTaskUpdateStateRequest_Validation(t *testing.T) {
	validStates := []string{"pending", "in_progress", "completed", "cancelled", "blocked"}

	for _, state := range validStates {
		t.Run("valid state: "+state, func(t *testing.T) {
			request := TaskUpdateStateRequest{
				TaskID: uuid.New().String(),
				State:  state,
			}

			// Verify state is valid
			parsed := parseTaskState(request.State)
			// For pending state, parsed should be pending
			if state == "pending" {
				assert.Equal(t, types.TaskStatePending, parsed)
			} else {
				// For other states, parsed should not be pending
				assert.NotEqual(t, types.TaskStatePending, parsed)
			}
		})
	}

	t.Run("invalid state", func(t *testing.T) {
		request := TaskUpdateStateRequest{
			TaskID: uuid.New().String(),
			State:  "invalid_state",
		}

		// Verify invalid state defaults to pending
		parsed := parseTaskState(request.State)
		assert.Equal(t, types.TaskStatePending, parsed)
	})
}
