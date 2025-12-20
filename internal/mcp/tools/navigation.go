package tools

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp/shared"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Navigation tools provide project browsing and selection capabilities

// ProjectSelectRequest defines the request for selecting a project
type ProjectSelectRequest struct {
	ProjectID string `json:"project_id" jsonschema_description:"The ID of the project to select" jsonschema:"required"`
}

// ProjectSelectResponse defines the response for project selection
type ProjectSelectResponse struct {
	Message   string `json:"message" jsonschema_description:"Confirmation message"`
	ProjectID string `json:"project_id" jsonschema_description:"Selected project ID"`
}

// ProjectListRequest defines the request for listing projects
type ProjectListRequest struct {
	Limit  *int `json:"limit,omitempty" jsonschema_description:"Maximum number of projects to return (default: 50)"`
	Offset *int `json:"offset,omitempty" jsonschema_description:"Number of projects to skip (default: 0)"`
}

// ProjectListResponse defines the response for listing projects
type ProjectListResponse struct {
	Projects []ProjectInfo `json:"projects" jsonschema_description:"List of projects"`
	Total    int           `json:"total" jsonschema_description:"Total number of projects"`
}

// ProjectInfo defines the project information returned in listings
type ProjectInfo struct {
	ID          string `json:"id" jsonschema_description:"Project ID"`
	Title       string `json:"title" jsonschema_description:"Project title"`
	Description string `json:"description,omitempty" jsonschema_description:"Project description"`
	State       string `json:"state" jsonschema_description:"Project state"`
	CreatedAt   string `json:"created_at" jsonschema_description:"Creation timestamp"`
}

// ProjectGetRequest defines the request for getting project details
type ProjectGetRequest struct {
	ProjectID string `json:"project_id" jsonschema_description:"The ID of the project to get" jsonschema:"required"`
}

// ProjectGetResponse defines the response for getting project details
type ProjectGetResponse struct {
	Project ProjectDetails `json:"project" jsonschema_description:"Project details"`
}

// ProjectDetails defines detailed project information
type ProjectDetails struct {
	ID          string `json:"id" jsonschema_description:"Project ID"`
	Title       string `json:"title" jsonschema_description:"Project title"`
	Description string `json:"description,omitempty" jsonschema_description:"Project description"`
	State       string `json:"state" jsonschema_description:"Project state"`
	CreatedAt   string `json:"created_at" jsonschema_description:"Creation timestamp"`
	UpdatedAt   string `json:"updated_at" jsonschema_description:"Last update timestamp"`
}

// RegisterNavigationTools registers all navigation tools with the MCP server
func RegisterNavigationTools(server *server.MCPServer, projectManager manager.ProjectManager, sessionManager session.Manager) {
	// Project selection tool
	projectSelectTool := mcp.NewTool("project_select",
		mcp.WithDescription("Select the active project for this session"),
		mcp.WithInputSchema[ProjectSelectRequest](),
		mcp.WithOutputSchema[ProjectSelectResponse](),
	)
	server.AddTool(projectSelectTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args ProjectSelectRequest) (ProjectSelectResponse, error) {
		projectID, err := uuid.Parse(args.ProjectID)
		if err != nil {
			return ProjectSelectResponse{}, fmt.Errorf("invalid project_id format: %w", err)
		}

		sessionIDUUID, err := shared.GetSessionUUID(ctx)
		if err != nil {
			return ProjectSelectResponse{}, fmt.Errorf("invalid session_id format: %w", err)
		}

		if err := sessionManager.SetProject(sessionIDUUID, projectID); err != nil {
			return ProjectSelectResponse{}, fmt.Errorf("failed to set project: %w", err)
		}

		return ProjectSelectResponse{
			Message:   fmt.Sprintf("Selected project %s for session", projectID),
			ProjectID: projectID.String(),
		}, nil
	}))

	// Project listing tool
	projectListTool := mcp.NewTool("project_list",
		mcp.WithDescription("List all available projects"),
		mcp.WithInputSchema[ProjectListRequest](),
		mcp.WithOutputSchema[ProjectListResponse](),
	)
	server.AddTool(projectListTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args ProjectListRequest) (ProjectListResponse, error) {
		limit := 50
		if args.Limit != nil && *args.Limit > 0 {
			limit = *args.Limit
		}

		projects, err := projectManager.ListProjects(ctx)
		if err != nil {
			return ProjectListResponse{}, fmt.Errorf("failed to list projects: %w", err)
		}

		// Apply pagination
		start := 0
		if args.Offset != nil && *args.Offset > 0 {
			start = *args.Offset
		}

		end := start + limit
		if end > len(projects) {
			end = len(projects)
		}

		if start > len(projects) {
			return ProjectListResponse{
				Projects: []ProjectInfo{},
				Total:    len(projects),
			}, nil
		}

		projectInfos := make([]ProjectInfo, 0, end-start)
		for _, project := range projects[start:end] {
			projectInfos = append(projectInfos, ProjectInfo{
				ID:          project.ID.String(),
				Title:       project.Title,
				Description: project.Description,
				State:       string(project.State),
				CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		return ProjectListResponse{
			Projects: projectInfos,
			Total:    len(projects),
		}, nil
	}))

	// Project details tool
	projectGetTool := mcp.NewTool("project_get",
		mcp.WithDescription("Get detailed information about a specific project"),
		mcp.WithInputSchema[ProjectGetRequest](),
		mcp.WithOutputSchema[ProjectGetResponse](),
	)
	server.AddTool(projectGetTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args ProjectGetRequest) (ProjectGetResponse, error) {
		projectID, err := uuid.Parse(args.ProjectID)
		if err != nil {
			return ProjectGetResponse{}, fmt.Errorf("invalid project_id format: %w", err)
		}

		project, err := projectManager.GetProject(ctx, projectID)
		if err != nil {
			return ProjectGetResponse{}, fmt.Errorf("failed to get project: %w", err)
		}

		return ProjectGetResponse{
			Project: ProjectDetails{
				ID:          project.ID.String(),
				Title:       project.Title,
				Description: project.Description,
				State:       string(project.State),
				CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:   project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			},
		}, nil
	}))
}

