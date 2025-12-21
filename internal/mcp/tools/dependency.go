package tools

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp/shared"
	knotutils "github.com/denkhaus/knot/v2/internal/utils"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Dependency management tools provide task dependency CRUD operations and analysis

// DependencyListRequest defines the request for listing task dependencies
type DependencyListRequest struct {
	TaskID string `json:"task_id" jsonschema_description:"The ID of the task to get dependencies for" jsonschema:"required"`
}

// DependencyListResponse defines the response for task dependencies
type DependencyListResponse struct {
	TaskID       string    `json:"task_id" jsonschema_description:"Task ID"`
	Dependencies []TaskInfo `json:"dependencies" jsonschema_description:"List of task dependencies"`
	Total        int       `json:"total" jsonschema_description:"Number of dependencies"`
	Message      string    `json:"message" jsonschema_description:"Status message"`
}

// DependencyAddRequest defines the request for adding a dependency
type DependencyAddRequest struct {
	TaskID          string `json:"task_id" jsonschema_description:"The ID of the task that will depend on another task" jsonschema:"required"`
	DependsOnTaskID string `json:"depends_on_task_id" jsonschema_description:"The ID of the task to depend on" jsonschema:"required"`
}

// DependencyAddResponse defines the response for adding a dependency
type DependencyAddResponse struct {
	Message          string      `json:"message" jsonschema_description:"Confirmation message"`
	TaskID           string      `json:"task_id" jsonschema_description:"Task ID"`
	DependsOnTaskID  string      `json:"depends_on_task_id" jsonschema_description:"Dependency task ID"`
	UpdatedTask      TaskDetails `json:"updated_task" jsonschema_description:"Updated task information"`
}

// DependencyRemoveRequest defines the request for removing a dependency
type DependencyRemoveRequest struct {
	TaskID          string `json:"task_id" jsonschema_description:"The ID of the task that depends on another task" jsonschema:"required"`
	DependsOnTaskID string `json:"depends_on_task_id" jsonschema_description:"The ID of the task to remove dependency from" jsonschema:"required"`
}

// DependencyRemoveResponse defines the response for removing a dependency
type DependencyRemoveResponse struct {
	Message          string      `json:"message" jsonschema_description:"Confirmation message"`
	TaskID           string      `json:"task_id" jsonschema_description:"Task ID"`
	DependsOnTaskID  string      `json:"depends_on_task_id" jsonschema_description:"Dependency task ID"`
	UpdatedTask      TaskDetails `json:"updated_task" jsonschema_description:"Updated task information"`
}

// DependencyCheckRequest defines the request for checking circular dependencies
type DependencyCheckRequest struct {
	TaskID          string `json:"task_id" jsonschema_description:"The ID of the task that will depend on another task" jsonschema:"required"`
	DependsOnTaskID string `json:"depends_on_task_id" jsonschema_description:"The ID of the task to depend on" jsonschema:"required"`
}

// DependencyCheckResponse defines the response for checking circular dependencies
type DependencyCheckResponse struct {
	IsCircular bool     `json:"is_circular" jsonschema_description:"Whether the dependency would create a cycle"`
	Message    string   `json:"message" jsonschema_description:"Explanation of the check result"`
	Path       []string `json:"path" jsonschema_description:"Path of the circular dependency if found"`
}

// RegisterDependencyTools registers all dependency management tools with the MCP server
func RegisterDependencyTools(mcpServer *server.MCPServer, projectManager manager.ProjectManager, sessionManager session.SessionManager) {
	// dependency_list - Show task dependencies
	dependencyListTool := mcp.NewTool("dependency_list",
		mcp.WithDescription("List all dependencies for a specific task"),
		mcp.WithInputSchema[DependencyListRequest](),
		mcp.WithOutputSchema[DependencyListResponse](),
	)
	mcpServer.AddTool(dependencyListTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args DependencyListRequest) (DependencyListResponse, error) {
		// Parse task ID
		taskUUID, err := uuid.Parse(args.TaskID)
		if err != nil {
			return DependencyListResponse{}, fmt.Errorf("invalid task ID: %w", err)
		}

		// Get task dependencies
		dependencies, err := projectManager.GetTaskDependencies(ctx, taskUUID)
		if err != nil {
			return DependencyListResponse{}, fmt.Errorf("failed to get task dependencies: %w", err)
		}

		// Convert to task info
		dependencyInfos := make([]TaskInfo, 0, len(dependencies))
		for _, dep := range dependencies {
			dependencyInfos = append(dependencyInfos, TaskInfo{
				ID:         dep.ID.String(),
				Title:      dep.Title,
				State:      string(dep.State),
				Priority:   dep.Priority.ToExternalString(),
				Complexity: dep.Complexity,
			})
		}

		return DependencyListResponse{
			TaskID:       args.TaskID,
			Dependencies: dependencyInfos,
			Total:        len(dependencies),
			Message:      fmt.Sprintf("Found %d dependencies for task", len(dependencies)),
		}, nil
	}))

	// dependency_add - Add dependency between tasks
	dependencyAddTool := mcp.NewTool("dependency_add",
		mcp.WithDescription("Add a dependency relationship between two tasks"),
		mcp.WithInputSchema[DependencyAddRequest](),
		mcp.WithOutputSchema[DependencyAddResponse](),
	)
	mcpServer.AddTool(dependencyAddTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args DependencyAddRequest) (DependencyAddResponse, error) {
		// Parse task IDs
		taskUUID, err := uuid.Parse(args.TaskID)
		if err != nil {
			return DependencyAddResponse{}, fmt.Errorf("invalid task ID: %w", err)
		}

		dependsOnUUID, err := uuid.Parse(args.DependsOnTaskID)
		if err != nil {
			return DependencyAddResponse{}, fmt.Errorf("invalid dependency task ID: %w", err)
		}

		// Get actor from session
		actor := shared.GetSessionActor(ctx)

		// Add dependency
		updatedTask, err := projectManager.AddTaskDependency(ctx, taskUUID, dependsOnUUID, actor)
		if err != nil {
			return DependencyAddResponse{}, fmt.Errorf("failed to add dependency: %w", err)
		}

		return DependencyAddResponse{
			Message: fmt.Sprintf("Successfully added dependency: task %s now depends on task %s", args.TaskID, args.DependsOnTaskID),
			TaskID:  args.TaskID,
			DependsOnTaskID: args.DependsOnTaskID,
			UpdatedTask: TaskDetails{
				ID:           updatedTask.ID.String(),
				Title:        updatedTask.Title,
				Description:  updatedTask.Description,
				State:        string(updatedTask.State),
				Priority:     updatedTask.Priority.ToExternalString(),
				Complexity:   updatedTask.Complexity,
				ProjectID:    updatedTask.ProjectID.String(),
				ParentID:     knotutils.GetParentIDAsString(updatedTask.ParentID),
				CreatedAt:    updatedTask.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:    updatedTask.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			},
		}, nil
	}))

	// dependency_remove - Remove dependency between tasks
	dependencyRemoveTool := mcp.NewTool("dependency_remove",
		mcp.WithDescription("Remove a dependency relationship between two tasks"),
		mcp.WithInputSchema[DependencyRemoveRequest](),
		mcp.WithOutputSchema[DependencyRemoveResponse](),
	)
	mcpServer.AddTool(dependencyRemoveTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args DependencyRemoveRequest) (DependencyRemoveResponse, error) {
		// Parse task IDs
		taskUUID, err := uuid.Parse(args.TaskID)
		if err != nil {
			return DependencyRemoveResponse{}, fmt.Errorf("invalid task ID: %w", err)
		}

		dependsOnUUID, err := uuid.Parse(args.DependsOnTaskID)
		if err != nil {
			return DependencyRemoveResponse{}, fmt.Errorf("invalid dependency task ID: %w", err)
		}

		// Get actor from session
		actor := shared.GetSessionActor(ctx)

		// Remove dependency
		updatedTask, err := projectManager.RemoveTaskDependency(ctx, taskUUID, dependsOnUUID, actor)
		if err != nil {
			return DependencyRemoveResponse{}, fmt.Errorf("failed to remove dependency: %w", err)
		}

		return DependencyRemoveResponse{
			Message: fmt.Sprintf("Successfully removed dependency: task %s no longer depends on task %s", args.TaskID, args.DependsOnTaskID),
			TaskID:  args.TaskID,
			DependsOnTaskID: args.DependsOnTaskID,
			UpdatedTask: TaskDetails{
				ID:           updatedTask.ID.String(),
				Title:        updatedTask.Title,
				Description:  updatedTask.Description,
				State:        string(updatedTask.State),
				Priority:     updatedTask.Priority.ToExternalString(),
				Complexity:   updatedTask.Complexity,
				ProjectID:    updatedTask.ProjectID.String(),
				ParentID:     knotutils.GetParentIDAsString(updatedTask.ParentID),
				CreatedAt:    updatedTask.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:    updatedTask.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			},
		}, nil
	}))

	// dependency_check - Check for circular dependencies
	dependencyCheckTool := mcp.NewTool("dependency_check",
		mcp.WithDescription("Check if adding a dependency would create circular dependencies"),
		mcp.WithInputSchema[DependencyCheckRequest](),
		mcp.WithOutputSchema[DependencyCheckResponse](),
	)
	mcpServer.AddTool(dependencyCheckTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args DependencyCheckRequest) (DependencyCheckResponse, error) {
		// Parse task IDs
		taskUUID, err := uuid.Parse(args.TaskID)
		if err != nil {
			return DependencyCheckResponse{}, fmt.Errorf("invalid task ID: %w", err)
		}

		dependsOnUUID, err := uuid.Parse(args.DependsOnTaskID)
		if err != nil {
			return DependencyCheckResponse{}, fmt.Errorf("invalid dependency task ID: %w", err)
		}

		// Check if tasks are the same
		if taskUUID == dependsOnUUID {
			return DependencyCheckResponse{
				IsCircular: true,
				Message:    "Task cannot depend on itself",
				Path:       []string{taskUUID.String(), taskUUID.String()},
			}, nil
		}

		// For circular dependency detection, we would need a recursive function
		// For now, we'll provide a basic implementation
		// TODO: Implement full circular dependency detection in the repository layer

		// Get current dependencies of the task that would become the dependency
		taskDependencies, err := projectManager.GetTaskDependencies(ctx, taskUUID)
		if err != nil {
			return DependencyCheckResponse{}, fmt.Errorf("failed to get task dependencies for check: %w", err)
		}

		// Check if dependsOnUUID is already in the task's dependencies
		for _, dep := range taskDependencies {
			if dep.ID == dependsOnUUID {
				return DependencyCheckResponse{
					IsCircular: false,
					Message:    "Dependency already exists",
					Path:       []string{},
				}, nil
			}
		}

		// Basic check: see if dependsOnTask depends on current task
		dependsOnDependencies, err := projectManager.GetTaskDependencies(ctx, dependsOnUUID)
		if err != nil {
			return DependencyCheckResponse{}, fmt.Errorf("failed to get dependency task dependencies for check: %w", err)
		}

		for _, dep := range dependsOnDependencies {
			if dep.ID == taskUUID {
				return DependencyCheckResponse{
					IsCircular: true,
					Message:    "Adding this dependency would create a circular relationship",
					Path:       []string{taskUUID.String(), dependsOnUUID.String(), taskUUID.String()},
				}, nil
			}
		}

		return DependencyCheckResponse{
			IsCircular: false,
			Message:    "No circular dependencies detected",
			Path:       []string{},
		}, nil
	}))
}