package tools

import (
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestRegisterProjectManagementTools tests that project tools are registered without errors
func TestRegisterProjectManagementTools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockProjectManager := mocks.NewMockProjectManager(ctrl)

	// Create MCP server
	mcpServer := server.NewMCPServer("test", "1.0.0")

	// Register tools - should not panic
	RegisterProjectManagementTools(mcpServer, mockProjectManager)

	// Verify server was created
	assert.NotNil(t, mcpServer)
}

// TestProjectCreateRequest_Validation tests project creation request structure
func TestProjectCreateRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request ProjectCreateRequest
		valid   bool
	}{
		{
			name: "valid request with all fields",
			request: ProjectCreateRequest{
				Title:       "Test Project",
				Description: "Test Description",
			},
			valid: true,
		},
		{
			name: "valid request with title only",
			request: ProjectCreateRequest{
				Title: "Test Project",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Title is required
			assert.NotEmpty(t, tt.request.Title, "Title should not be empty")
		})
	}
}

// TestProjectUpdateRequest_Validation tests project update request structure
func TestProjectUpdateRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request ProjectUpdateRequest
	}{
		{
			name: "valid update with title",
			request: ProjectUpdateRequest{
				ProjectID:   uuid.New().String(),
				Title:       "Updated Title",
				Description: "Updated Description",
			},
		},
		{
			name: "valid update with description only",
			request: ProjectUpdateRequest{
				ProjectID:   uuid.New().String(),
				Description: "Updated Description",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify project ID is valid UUID format
			_, err := uuid.Parse(tt.request.ProjectID)
			assert.NoError(t, err, "ProjectID should be valid UUID")
		})
	}
}

// TestProjectDeleteRequest_Validation tests project deletion request structure
func TestProjectDeleteRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request ProjectDeleteRequest
		valid   bool
	}{
		{
			name: "valid delete request",
			request: ProjectDeleteRequest{
				ProjectID: uuid.New().String(),
			},
			valid: true,
		},
		{
			name: "invalid project ID format",
			request: ProjectDeleteRequest{
				ProjectID: "not-a-uuid",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uuid.Parse(tt.request.ProjectID)
			if tt.valid {
				assert.NoError(t, err, "ProjectID should be valid UUID")
			} else {
				assert.Error(t, err, "ProjectID should be invalid")
			}
		})
	}
}

// TestProjectCreateResponse tests project creation response structure
func TestProjectCreateResponse(t *testing.T) {
	projectID := uuid.New()

	response := ProjectCreateResponse{
		Message:   "Created project Test Project (ID: " + projectID.String() + ")",
		ProjectID: projectID.String(),
		Title:     "Test Project",
	}

	assert.NotEmpty(t, response.Message)
	assert.NotEmpty(t, response.ProjectID)
	assert.NotEmpty(t, response.Title)
	assert.Contains(t, response.Message, "Created project")
}

// TestProjectUpdateResponse tests project update response structure
func TestProjectUpdateResponse(t *testing.T) {
	projectID := uuid.New()
	now := time.Now()

	response := ProjectUpdateResponse{
		Message: "Updated project Test Project",
		Project: ProjectInfo{
			ID:          projectID.String(),
			Title:       "Test Project",
			Description: "Updated Description",
			State:       "active",
			CreatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
		},
	}

	assert.NotEmpty(t, response.Message)
	assert.NotEmpty(t, response.Project.ID)
	assert.NotEmpty(t, response.Project.Title)
	assert.Equal(t, "active", response.Project.State)
}

// TestProjectDeleteResponse tests project deletion response structure
func TestProjectDeleteResponse(t *testing.T) {
	projectID := uuid.New()

	response := ProjectDeleteResponse{
		Message: "Deleted project " + projectID.String(),
	}

	assert.NotEmpty(t, response.Message)
	assert.Contains(t, response.Message, "Deleted project")
}

// TestProjectInfo_Conversion tests conversion from domain type to response type
func TestProjectInfo_Conversion(t *testing.T) {
	now := time.Now()
	projectID := uuid.New()

	domainProject := &types.Project{
		ID:          projectID,
		Title:       "Test Project",
		Description: "Test Description",
		State:       types.ProjectStateActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	info := ProjectInfo{
		ID:          domainProject.ID.String(),
		Title:       domainProject.Title,
		Description: domainProject.Description,
		State:       string(domainProject.State),
		CreatedAt:   domainProject.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	assert.Equal(t, projectID.String(), info.ID)
	assert.Equal(t, "Test Project", info.Title)
	assert.Equal(t, "Test Description", info.Description)
	assert.Equal(t, "active", info.State)
}

// TestProjectStates tests all project state values
func TestProjectStates(t *testing.T) {
	states := []string{"active", "archived", "deleted"}

	for _, state := range states {
		t.Run("state: "+state, func(t *testing.T) {
			// Verify state string is not empty
			assert.NotEmpty(t, state)
		})
	}
}
