package tools

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp/shared"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Project management tools provide project CRUD operations

// ProjectCreateRequest defines the request for creating a project
type ProjectCreateRequest struct {
	Title       string `json:"title" jsonschema_description:"Project title" jsonschema:"required"`
	Description string `json:"description,omitempty" jsonschema_description:"Project description"`
}

// ProjectCreateResponse defines the response for project creation
type ProjectCreateResponse struct {
	Message   string `json:"message" jsonschema_description:"Confirmation message"`
	ProjectID string `json:"project_id" jsonschema_description:"Created project ID"`
	Title     string `json:"title" jsonschema_description:"Project title"`
}

// ProjectUpdateRequest defines the request for updating a project
type ProjectUpdateRequest struct {
	ProjectID   string `json:"project_id" jsonschema_description:"The ID of the project to update" jsonschema:"required"`
	Title       string `json:"title,omitempty" jsonschema_description:"New project title"`
	Description string `json:"description,omitempty" jsonschema_description:"New project description"`
}

// ProjectUpdateResponse defines the response for project update
type ProjectUpdateResponse struct {
	Message string `json:"message" jsonschema_description:"Confirmation message"`
	Project ProjectInfo `json:"project" jsonschema_description:"Updated project information"`
}

// ProjectDeleteRequest defines the request for deleting a project
type ProjectDeleteRequest struct {
	ProjectID string `json:"project_id" jsonschema_description:"The ID of the project to delete" jsonschema:"required"`
}

// ProjectDeleteResponse defines the response for project deletion
type ProjectDeleteResponse struct {
	Message string `json:"message" jsonschema_description:"Confirmation message"`
}

// RegisterProjectManagementTools registers all project management tools with the MCP server
func RegisterProjectManagementTools(server *server.MCPServer, projectManager manager.ProjectManager) {
	// Project creation tool
	projectCreateTool := mcp.NewTool("project_create",
		mcp.WithDescription("Create a new project"),
		mcp.WithInputSchema[ProjectCreateRequest](),
		mcp.WithOutputSchema[ProjectCreateResponse](),
	)
	server.AddTool(projectCreateTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args ProjectCreateRequest) (ProjectCreateResponse, error) {
	project, err := projectManager.CreateProject(ctx, args.Title, args.Description, shared.GetSessionActor(ctx))
		if err != nil {
			return ProjectCreateResponse{}, fmt.Errorf("failed to create project: %w", err)
		}

		return ProjectCreateResponse{
			Message:   fmt.Sprintf("Created project %s (ID: %s)", project.Title, project.ID),
			ProjectID: project.ID.String(),
			Title:     project.Title,
		}, nil
	}))

	// Project update tool
	projectUpdateTool := mcp.NewTool("project_update",
		mcp.WithDescription("Update an existing project"),
		mcp.WithInputSchema[ProjectUpdateRequest](),
		mcp.WithOutputSchema[ProjectUpdateResponse](),
	)
	server.AddTool(projectUpdateTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args ProjectUpdateRequest) (ProjectUpdateResponse, error) {
		projectID, err := uuid.Parse(args.ProjectID)
		if err != nil {
			return ProjectUpdateResponse{}, fmt.Errorf("invalid project_id format: %w", err)
		}

		project, err := projectManager.UpdateProject(ctx, projectID, args.Title, args.Description, shared.GetSessionActor(ctx))
		if err != nil {
			return ProjectUpdateResponse{}, fmt.Errorf("failed to update project: %w", err)
		}

		return ProjectUpdateResponse{
			Message: fmt.Sprintf("Updated project %s", project.Title),
			Project: ProjectInfo{
				ID:          project.ID.String(),
				Title:       project.Title,
				Description: project.Description,
				State:       string(project.State),
				CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			},
		}, nil
	}))

	// Project deletion tool
	projectDeleteTool := mcp.NewTool("project_delete",
		mcp.WithDescription("Delete a project"),
		mcp.WithInputSchema[ProjectDeleteRequest](),
		mcp.WithOutputSchema[ProjectDeleteResponse](),
	)
	server.AddTool(projectDeleteTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args ProjectDeleteRequest) (ProjectDeleteResponse, error) {
		projectID, err := uuid.Parse(args.ProjectID)
		if err != nil {
			return ProjectDeleteResponse{}, fmt.Errorf("invalid project_id format: %w", err)
		}

		err = projectManager.DeleteProject(ctx, projectID)
		if err != nil {
			return ProjectDeleteResponse{}, fmt.Errorf("failed to delete project: %w", err)
		}

		return ProjectDeleteResponse{
			Message: fmt.Sprintf("Deleted project %s", projectID),
		}, nil
	}))
}

