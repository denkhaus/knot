package tools

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Task management tools provide task CRUD operations

// TaskCreateRequest defines the request for creating a task
type TaskCreateRequest struct {
	Title       string `json:"title" jsonschema_description:"Task title" jsonschema:"required"`
	Description string `json:"description,omitempty" jsonschema_description:"Task description"`
	Complexity  int    `json:"complexity,omitempty" jsonschema_description:"Task complexity (1-10)" jsonschema:"minimum=1,maximum=10,default=5"`
	ParentID    string `json:"parent_id,omitempty" jsonschema_description:"Parent task ID for subtask creation"`
	Priority    string `json:"priority,omitempty" jsonschema_description:"Task priority (low, medium, high)" jsonschema:"enum=low,medium,high,default=medium"`
}

// TaskCreateResponse defines the response for task creation
type TaskCreateResponse struct {
	Message   string `json:"message" jsonschema_description:"Confirmation message"`
	TaskID    string `json:"task_id" jsonschema_description:"Created task ID"`
	ProjectID string `json:"project_id" jsonschema_description:"Project ID"`
	Title     string `json:"title" jsonschema_description:"Task title"`
}

// TaskGetRequest defines the request for getting task details
type TaskGetRequest struct {
	TaskID string `json:"task_id" jsonschema_description:"The ID of the task to get" jsonschema:"required"`
}

// TaskGetResponse defines the response for getting task details
type TaskGetResponse struct {
	Task TaskDetails `json:"task" jsonschema_description:"Task details"`
}

// TaskDetails defines detailed task information
type TaskDetails struct {
	ID           string     `json:"id" jsonschema_description:"Task ID"`
	Title        string     `json:"title" jsonschema_description:"Task title"`
	Description  string     `json:"description,omitempty" jsonschema_description:"Task description"`
	State        string     `json:"state" jsonschema_description:"Task state"`
	Priority     string     `json:"priority" jsonschema_description:"Task priority"`
	Complexity   int        `json:"complexity" jsonschema_description:"Task complexity"`
	ProjectID    string     `json:"project_id" jsonschema_description:"Project ID"`
	ParentID     *string    `json:"parent_id,omitempty" jsonschema_description:"Parent task ID"`
	CreatedAt    string     `json:"created_at" jsonschema_description:"Creation timestamp"`
	UpdatedAt    string     `json:"updated_at" jsonschema_description:"Last update timestamp"`
	Dependencies []TaskInfo `json:"dependencies,omitempty" jsonschema_description:"Task dependencies"`
}

// TaskInfo defines basic task information for listings
type TaskInfo struct {
	ID         string `json:"id" jsonschema_description:"Task ID"`
	Title      string `json:"title" jsonschema_description:"Task title"`
	State      string `json:"state" jsonschema_description:"Task state"`
	Priority   string `json:"priority" jsonschema_description:"Task priority"`
	Complexity int    `json:"complexity" jsonschema_description:"Task complexity"`
}

// TaskUpdateRequest defines the request for updating a task
type TaskUpdateRequest struct {
	TaskID      string `json:"task_id" jsonschema_description:"The ID of the task to update" jsonschema:"required"`
	Title       string `json:"title,omitempty" jsonschema_description:"New task title"`
	Description string `json:"description,omitempty" jsonschema_description:"New task description"`
	Complexity  int    `json:"complexity,omitempty" jsonschema_description:"New task complexity (1-10)" jsonschema:"minimum=1,maximum=10"`
	Priority    string `json:"priority,omitempty" jsonschema_description:"New task priority (low, medium, high)"`
}

// TaskUpdateResponse defines the response for task update
type TaskUpdateResponse struct {
	Message string      `json:"message" jsonschema_description:"Confirmation message"`
	Task    TaskDetails `json:"task" jsonschema_description:"Updated task information"`
}

// TaskUpdateStateRequest defines the request for updating task state
type TaskUpdateStateRequest struct {
	TaskID string `json:"task_id" jsonschema_description:"The ID of the task to update" jsonschema:"required"`
	State  string `json:"state" jsonschema_description:"New task state" jsonschema:"required" jsonschema:"enum=pending,in_progress,completed,cancelled,blocked"`
}

// TaskUpdateStateResponse defines the response for task state update
type TaskUpdateStateResponse struct {
	Message string      `json:"message" jsonschema_description:"Confirmation message"`
	Task    TaskDetails `json:"task" jsonschema_description:"Updated task information"`
}

// TaskDeleteRequest defines the request for deleting a task
type TaskDeleteRequest struct {
	TaskID string `json:"task_id" jsonschema_description:"The ID of the task to delete" jsonschema:"required"`
}

// TaskDeleteResponse defines the response for task deletion
type TaskDeleteResponse struct {
	Message string `json:"message" jsonschema_description:"Confirmation message"`
}

// RegisterTaskManagementTools registers all task management tools with the MCP server
func RegisterTaskManagementTools(server *server.MCPServer, projectManager manager.ProjectManager, sessionManager session.Manager) {
	// Task creation tool
	taskCreateTool := mcp.NewTool("task_create",
		mcp.WithDescription("Create a new task in the selected project"),
		mcp.WithInputSchema[TaskCreateRequest](),
		mcp.WithOutputSchema[TaskCreateResponse](),
	)
	server.AddTool(taskCreateTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args TaskCreateRequest) (TaskCreateResponse, error) {
		complexity := args.Complexity
		if complexity == 0 {
			// TODO(knot-dmd): make this configurable
			complexity = 5 // Default complexity
		}

		priority := parsePriority(args.Priority)
		actor := getSessionActor(ctx)

		// Get session context
		sessionID := getSessionID(ctx)
		sessionIDUUID, err := uuid.Parse(sessionID)
		if err != nil {
			return TaskCreateResponse{}, fmt.Errorf("invalid session_id format: %w", err)
		}

		session, err := sessionManager.GetSession(sessionIDUUID)
		if err != nil {
			return TaskCreateResponse{}, fmt.Errorf("failed to get session: %w", err)
		}

		if session.ProjectID == nil {
			return TaskCreateResponse{}, fmt.Errorf("no project selected for this session")
		}

		var parentID *uuid.UUID
		if args.ParentID != "" {
			parsedParentID, err := uuid.Parse(args.ParentID)
			if err != nil {
				return TaskCreateResponse{}, fmt.Errorf("invalid parent_id format: %w", err)
			}
			parentID = &parsedParentID
		}

		task, err := projectManager.CreateTask(ctx, *session.ProjectID, parentID, args.Title, args.Description, complexity, priority, actor)
		if err != nil {
			return TaskCreateResponse{}, fmt.Errorf("failed to create task: %w", err)
		}

		return TaskCreateResponse{
			Message:   fmt.Sprintf("Created task %s (ID: %s)", task.Title, task.ID),
			TaskID:    task.ID.String(),
			ProjectID: session.ProjectID.String(),
			Title:     task.Title,
		}, nil
	}))

	// Task details tool
	taskGetTool := mcp.NewTool("task_get",
		mcp.WithDescription("Get detailed information about a specific task"),
		mcp.WithInputSchema[TaskGetRequest](),
		mcp.WithOutputSchema[TaskGetResponse](),
	)
	server.AddTool(taskGetTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args TaskGetRequest) (TaskGetResponse, error) {
		taskID, err := uuid.Parse(args.TaskID)
		if err != nil {
			return TaskGetResponse{}, fmt.Errorf("invalid task_id format: %w", err)
		}

		task, err := projectManager.GetTask(ctx, taskID)
		if err != nil {
			return TaskGetResponse{}, fmt.Errorf("failed to get task: %w", err)
		}

		dependencies, err := projectManager.GetTaskDependencies(ctx, taskID)
		if err != nil {
			return TaskGetResponse{}, fmt.Errorf("failed to get task dependencies: %w", err)
		}

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

		var parentID *string
		if task.ParentID != nil {
			parentStr := task.ParentID.String()
			parentID = &parentStr
		}

		return TaskGetResponse{
			Task: TaskDetails{
				ID:           task.ID.String(),
				Title:        task.Title,
				Description:  task.Description,
				State:        string(task.State),
				Priority:     task.Priority.ToExternalString(),
				Complexity:   task.Complexity,
				ProjectID:    task.ProjectID.String(),
				ParentID:     parentID,
				CreatedAt:    task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:    task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
				Dependencies: dependencyInfos,
			},
		}, nil
	}))

	// Task update tool
	taskUpdateTool := mcp.NewTool("task_update",
		mcp.WithDescription("Update an existing task"),
		mcp.WithInputSchema[TaskUpdateRequest](),
		mcp.WithOutputSchema[TaskUpdateResponse](),
	)
	server.AddTool(taskUpdateTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args TaskUpdateRequest) (TaskUpdateResponse, error) {
		taskID, err := uuid.Parse(args.TaskID)
		if err != nil {
			return TaskUpdateResponse{}, fmt.Errorf("invalid task_id format: %w", err)
		}

		actor := getSessionActor(ctx)
		priority := parsePriority(args.Priority)

		task, err := projectManager.UpdateTask(ctx, taskID, args.Title, args.Description, args.Complexity, types.TaskState(""), actor)
		if err != nil {
			return TaskUpdateResponse{}, fmt.Errorf("failed to update task: %w", err)
		}

		// Update priority if specified
		if args.Priority != "" {
			task, err = projectManager.UpdateTaskPriority(ctx, taskID, priority, actor)
			if err != nil {
				return TaskUpdateResponse{}, fmt.Errorf("failed to update task priority: %w", err)
			}
		}

		var parentID *string
		if task.ParentID != nil {
			parentStr := task.ParentID.String()
			parentID = &parentStr
		}

		return TaskUpdateResponse{
			Message: fmt.Sprintf("Updated task %s", task.Title),
			Task: TaskDetails{
				ID:          task.ID.String(),
				Title:       task.Title,
				Description: task.Description,
				State:       string(task.State),
				Priority:    task.Priority.ToExternalString(),
				Complexity:  task.Complexity,
				ProjectID:   task.ProjectID.String(),
				ParentID:    parentID,
				CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			},
		}, nil
	}))

	// Task state update tool
	taskUpdateStateTool := mcp.NewTool("task_update_state",
		mcp.WithDescription("Update the state of a task"),
		mcp.WithInputSchema[TaskUpdateStateRequest](),
		mcp.WithOutputSchema[TaskUpdateStateResponse](),
	)
	server.AddTool(taskUpdateStateTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args TaskUpdateStateRequest) (TaskUpdateStateResponse, error) {
		taskID, err := uuid.Parse(args.TaskID)
		if err != nil {
			return TaskUpdateStateResponse{}, fmt.Errorf("invalid task_id format: %w", err)
		}

		state := parseTaskState(args.State)
		actor := getSessionActor(ctx)

		task, err := projectManager.UpdateTaskState(ctx, taskID, state, actor)
		if err != nil {
			return TaskUpdateStateResponse{}, fmt.Errorf("failed to update task state: %w", err)
		}

		var parentID *string
		if task.ParentID != nil {
			parentStr := task.ParentID.String()
			parentID = &parentStr
		}

		return TaskUpdateStateResponse{
			Message: fmt.Sprintf("Updated task %s state to %s", task.Title, state),
			Task: TaskDetails{
				ID:          task.ID.String(),
				Title:       task.Title,
				Description: task.Description,
				State:       string(task.State),
				Priority:    task.Priority.ToExternalString(),
				Complexity:  task.Complexity,
				ProjectID:   task.ProjectID.String(),
				ParentID:    parentID,
				CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			},
		}, nil
	}))

	// Task deletion tool
	taskDeleteTool := mcp.NewTool("task_delete",
		mcp.WithDescription("Delete a task"),
		mcp.WithInputSchema[TaskDeleteRequest](),
		mcp.WithOutputSchema[TaskDeleteResponse](),
	)
	server.AddTool(taskDeleteTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args TaskDeleteRequest) (TaskDeleteResponse, error) {
		taskID, err := uuid.Parse(args.TaskID)
		if err != nil {
			return TaskDeleteResponse{}, fmt.Errorf("invalid task_id format: %w", err)
		}

		actor := getSessionActor(ctx)
		err = projectManager.DeleteTask(ctx, taskID, actor)
		if err != nil {
			return TaskDeleteResponse{}, fmt.Errorf("failed to delete task: %w", err)
		}

		return TaskDeleteResponse{
			Message: fmt.Sprintf("Deleted task %s", taskID),
		}, nil
	}))
}

// parsePriority converts string priority to TaskPriority
func parsePriority(priority string) types.TaskPriority {
	switch priority {
	case "low":
		return types.TaskPriorityLow
	case "high":
		return types.TaskPriorityHigh
	default:
		return types.TaskPriorityMedium
	}
}

// parseTaskState converts string state to TaskState
func parseTaskState(state string) types.TaskState {
	switch state {
	case "pending":
		return types.TaskStatePending
	case "in_progress":
		return types.TaskStateInProgress
	case "completed":
		return types.TaskStateCompleted
	case "cancelled":
		return types.TaskStateCancelled
	case "blocked":
		return types.TaskStateBlocked
	default:
		return types.TaskStatePending
	}
}
