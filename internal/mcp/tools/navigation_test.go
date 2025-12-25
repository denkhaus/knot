package tools

import (
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestRegisterNavigationTools tests that navigation tools are registered without errors
func TestRegisterNavigationTools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockProjectManager := mocks.NewMockProjectManager(ctrl)
	mockSessionManager := mocks.NewMockSessionManager(ctrl)

	// Create MCP server
	mcpServer := server.NewMCPServer("test", "1.0.0")

	// Register tools - should not panic
	RegisterNavigationTools(mcpServer, mockProjectManager, mockSessionManager)

	// Verify server was created
	assert.NotNil(t, mcpServer)
}

// TestNavigationRequestStructures tests various request structures
func TestNavigationRequestStructures(t *testing.T) {
	t.Run("ProjectSelectRequest", func(t *testing.T) {
		projectID := uuid.New()
		request := ProjectSelectRequest{
			ProjectID: projectID.String(),
		}

		// Verify project ID is valid UUID format
		_, err := uuid.Parse(request.ProjectID)
		assert.NoError(t, err, "ProjectID should be valid UUID")
		assert.Equal(t, projectID.String(), request.ProjectID)
	})

	t.Run("ProjectListRequest with limit and offset", func(t *testing.T) {
		limit := 10
		offset := 5
		request := ProjectListRequest{
			Limit:  &limit,
			Offset: &offset,
		}

		assert.NotNil(t, request.Limit)
		assert.NotNil(t, request.Offset)
		assert.Equal(t, 10, *request.Limit)
		assert.Equal(t, 5, *request.Offset)
	})

	t.Run("ProjectListRequest with limit only", func(t *testing.T) {
		limit := 20
		request := ProjectListRequest{
			Limit: &limit,
		}

		assert.NotNil(t, request.Limit)
		assert.Nil(t, request.Offset)
		assert.Equal(t, 20, *request.Limit)
	})

	t.Run("ProjectListRequest with offset only", func(t *testing.T) {
		offset := 15
		request := ProjectListRequest{
			Offset: &offset,
		}

		assert.Nil(t, request.Limit)
		assert.NotNil(t, request.Offset)
		assert.Equal(t, 15, *request.Offset)
	})

	t.Run("ProjectListRequest empty", func(t *testing.T) {
		request := ProjectListRequest{}

		assert.Nil(t, request.Limit)
		assert.Nil(t, request.Offset)
	})

	t.Run("ProjectGetRequest", func(t *testing.T) {
		projectID := uuid.New()
		request := ProjectGetRequest{
			ProjectID: projectID.String(),
		}

		// Verify project ID is valid UUID format
		_, err := uuid.Parse(request.ProjectID)
		assert.NoError(t, err, "ProjectID should be valid UUID")
		assert.Equal(t, projectID.String(), request.ProjectID)
	})
}

// TestNavigationResponseStructures tests various response structures
func TestNavigationResponseStructures(t *testing.T) {
	projectID := uuid.New()

	t.Run("ProjectSelectResponse", func(t *testing.T) {
		response := ProjectSelectResponse{
			Message:   "Selected project " + projectID.String() + " for session",
			ProjectID: projectID.String(),
		}

		assert.NotEmpty(t, response.Message)
		assert.NotEmpty(t, response.ProjectID)
		assert.Contains(t, response.Message, "Selected project")
		assert.Equal(t, projectID.String(), response.ProjectID)
	})

	t.Run("ProjectListResponse", func(t *testing.T) {
		now := time.Now()
		projectID2 := uuid.New()

		response := ProjectListResponse{
			Projects: []ProjectInfo{
				{
					ID:          projectID.String(),
					Title:       "Project 1",
					Description: "Description 1",
					State:       "active",
					CreatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
				},
				{
					ID:          projectID2.String(),
					Title:       "Project 2",
					Description: "Description 2",
					State:       "archived",
					CreatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
				},
			},
			Total: 2,
		}

		assert.Len(t, response.Projects, 2)
		assert.Equal(t, 2, response.Total)
		assert.Equal(t, "Project 1", response.Projects[0].Title)
		assert.Equal(t, "active", response.Projects[0].State)
		assert.Equal(t, "Project 2", response.Projects[1].Title)
		assert.Equal(t, "archived", response.Projects[1].State)
	})

	t.Run("ProjectListResponse empty", func(t *testing.T) {
		response := ProjectListResponse{
			Projects: []ProjectInfo{},
			Total:    0,
		}

		assert.Empty(t, response.Projects)
		assert.Equal(t, 0, response.Total)
	})

	t.Run("ProjectGetResponse", func(t *testing.T) {
		now := time.Now()

		response := ProjectGetResponse{
			Project: ProjectDetails{
				ID:          projectID.String(),
				Title:       "Test Project",
				Description: "Test Description",
				State:       "active",
				CreatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
			},
		}

		assert.Equal(t, projectID.String(), response.Project.ID)
		assert.Equal(t, "Test Project", response.Project.Title)
		assert.Equal(t, "Test Description", response.Project.Description)
		assert.Equal(t, "active", response.Project.State)
	})
}

// TestProjectInfoStructure tests the ProjectInfo structure
func TestProjectInfoStructure(t *testing.T) {
	now := time.Now()
	projectID := uuid.New()

	info := ProjectInfo{
		ID:          projectID.String(),
		Title:       "Test Project",
		Description: "Test Description",
		State:       "active",
		CreatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
	}

	assert.Equal(t, projectID.String(), info.ID)
	assert.Equal(t, "Test Project", info.Title)
	assert.Equal(t, "Test Description", info.Description)
	assert.Equal(t, "active", info.State)
	assert.NotEmpty(t, info.CreatedAt)
}

// TestProjectDetailsStructure tests the ProjectDetails structure
func TestProjectDetailsStructure(t *testing.T) {
	now := time.Now()
	projectID := uuid.New()

	details := ProjectDetails{
		ID:          projectID.String(),
		Title:       "Test Project",
		Description: "Test Description",
		State:       "active",
		CreatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
	}

	assert.Equal(t, projectID.String(), details.ID)
	assert.Equal(t, "Test Project", details.Title)
	assert.Equal(t, "Test Description", details.Description)
	assert.Equal(t, "active", details.State)
	assert.NotEmpty(t, details.CreatedAt)
	assert.NotEmpty(t, details.UpdatedAt)
}

// TestProjectInfo_Conversion_Navigation tests conversion from domain type to response type for navigation
func TestProjectInfo_Conversion_Navigation(t *testing.T) {
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

// TestProjectDetails_Conversion tests conversion from domain type to detailed response type
func TestProjectDetails_Conversion(t *testing.T) {
	now := time.Now()
	projectID := uuid.New()

	domainProject := &types.Project{
		ID:          projectID,
		Title:       "Test Project",
		Description: "Test Description",
		State:       types.ProjectStateArchived,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	details := ProjectDetails{
		ID:          domainProject.ID.String(),
		Title:       domainProject.Title,
		Description: domainProject.Description,
		State:       string(domainProject.State),
		CreatedAt:   domainProject.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   domainProject.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	assert.Equal(t, projectID.String(), details.ID)
	assert.Equal(t, "Test Project", details.Title)
	assert.Equal(t, "Test Description", details.Description)
	assert.Equal(t, "archived", details.State)
	assert.NotEmpty(t, details.CreatedAt)
	assert.NotEmpty(t, details.UpdatedAt)
}

// TestNavigationRequestValidation tests invalid request formats
func TestNavigationRequestValidation(t *testing.T) {
	t.Run("ProjectSelectRequest invalid UUID", func(t *testing.T) {
		request := ProjectSelectRequest{
			ProjectID: "not-a-uuid",
		}

		_, err := uuid.Parse(request.ProjectID)
		assert.Error(t, err, "Should fail to parse invalid UUID")
	})

	t.Run("ProjectGetRequest invalid UUID", func(t *testing.T) {
		request := ProjectGetRequest{
			ProjectID: "invalid-uuid-format",
		}

		_, err := uuid.Parse(request.ProjectID)
		assert.Error(t, err, "Should fail to parse invalid UUID")
	})

	t.Run("ProjectListRequest negative limit", func(t *testing.T) {
		limit := -5
		request := ProjectListRequest{
			Limit: &limit,
		}

		// Request can technically be created with negative values
		// Validation should happen in handler
		assert.NotNil(t, request.Limit)
		assert.Equal(t, -5, *request.Limit)
	})

	t.Run("ProjectListRequest negative offset", func(t *testing.T) {
		offset := -10
		request := ProjectListRequest{
			Offset: &offset,
		}

		// Request can technically be created with negative values
		// Validation should happen in handler
		assert.NotNil(t, request.Offset)
		assert.Equal(t, -10, *request.Offset)
	})
}

// TestNavigationPaginationLogic tests pagination edge cases
func TestNavigationPaginationLogic(t *testing.T) {
	now := time.Now()

	projects := []*types.Project{
		{ID: uuid.New(), Title: "Project 1", State: types.ProjectStateActive, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), Title: "Project 2", State: types.ProjectStateActive, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), Title: "Project 3", State: types.ProjectStateActive, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), Title: "Project 4", State: types.ProjectStateActive, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), Title: "Project 5", State: types.ProjectStateActive, CreatedAt: now, UpdatedAt: now},
	}

	t.Run("First page with limit 2", func(t *testing.T) {
		limit := 2
		offset := 0

		start := offset
		if start > len(projects) {
			start = len(projects)
		}

		end := start + limit
		if end > len(projects) {
			end = len(projects)
		}

		require.Equal(t, 0, start)
		require.Equal(t, 2, end)
		assert.Len(t, projects[start:end], 2)
	})

	t.Run("Second page with limit 2", func(t *testing.T) {
		limit := 2
		offset := 2

		start := offset
		if start > len(projects) {
			start = len(projects)
		}

		end := start + limit
		if end > len(projects) {
			end = len(projects)
		}

		require.Equal(t, 2, start)
		require.Equal(t, 4, end)
		assert.Len(t, projects[start:end], 2)
	})

	t.Run("Last page partial", func(t *testing.T) {
		limit := 2
		offset := 4

		start := offset
		if start > len(projects) {
			start = len(projects)
		}

		end := start + limit
		if end > len(projects) {
			end = len(projects)
		}

		require.Equal(t, 4, start)
		require.Equal(t, 5, end)
		assert.Len(t, projects[start:end], 1)
	})

	t.Run("Offset beyond available projects", func(t *testing.T) {
		limit := 2
		offset := 10

		start := offset
		if start > len(projects) {
			start = len(projects)
		}

		end := start + limit
		if end > len(projects) {
			end = len(projects)
		}

		require.Equal(t, 5, start)
		require.Equal(t, 5, end)
		assert.Len(t, projects[start:end], 0)
	})
}
