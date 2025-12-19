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

// Status and query tools provide project status, task state queries, and actionable work discovery

// StatusReadyRequest defines the request for getting ready tasks
type StatusReadyRequest struct {
	ProjectID string `json:"project_id,omitempty" jsonschema_description:"The ID of the project to check (optional, uses selected project if not provided)"`
	Limit     *int   `json:"limit,omitempty" jsonschema_description:"Maximum number of tasks to return (default: 20)"`
}

// StatusReadyResponse defines the response for ready tasks
type StatusReadyResponse struct {
	ProjectID string      `json:"project_id" jsonschema_description:"Project ID"`
	Tasks     []TaskInfo  `json:"tasks" jsonschema_description:"List of ready tasks"`
	Total     int         `json:"total" jsonschema_description:"Total number of ready tasks"`
	Message   string      `json:"message" jsonschema_description:"Status message"`
}

// StatusActionableRequest defines the request for getting actionable tasks
type StatusActionableRequest struct {
	ProjectID string `json:"project_id,omitempty" jsonschema_description:"The ID of the project to check (optional, uses selected project if not provided)"`
	Limit     *int   `json:"limit,omitempty" jsonschema_description:"Maximum number of tasks to return (default: 20)"`
}

// StatusActionableResponse defines the response for actionable tasks
type StatusActionableResponse struct {
	ProjectID string      `json:"project_id" jsonschema_description:"Project ID"`
	Tasks     []TaskInfo  `json:"tasks" jsonschema_description:"List of actionable tasks"`
	Total     int         `json:"total" jsonschema_description:"Total number of actionable tasks"`
	Message   string      `json:"message" jsonschema_description:"Status message"`
}

// StatusProjectRequest defines the request for getting project status
type StatusProjectRequest struct {
	ProjectID string `json:"project_id,omitempty" jsonschema_description:"The ID of the project to check (optional, uses selected project if not provided)"`
}

// StatusProjectResponse defines the response for project status
type StatusProjectResponse struct {
	ProjectInfo ProjectInfo `json:"project_info" jsonschema_description:"Project information"`
	Progress    Progress    `json:"progress" jsonschema_description:"Project progress"`
	TaskStats   TaskStats   `json:"task_stats" jsonschema_description:"Task statistics"`
	Message     string      `json:"message" jsonschema_description:"Status message"`
}

// Progress defines project progress information
type Progress struct {
	TotalTasks       int     `json:"total_tasks" jsonschema_description:"Total number of tasks"`
	CompletedTasks   int     `json:"completed_tasks" jsonschema_description:"Number of completed tasks"`
	PendingTasks     int     `json:"pending_tasks" jsonschema_description:"Number of pending tasks"`
	InProgressTasks  int     `json:"in_progress_tasks" jsonschema_description:"Number of in-progress tasks"`
	BlockedTasks     int     `json:"blocked_tasks" jsonschema_description:"Number of blocked tasks"`
	CompletionRate   float64 `json:"completion_rate" jsonschema_description:"Completion percentage (0-100)"`
}

// TaskStats defines task statistics
type TaskStats struct {
	TotalByState      map[string]int `json:"total_by_state" jsonschema_description:"Task count by state"`
	TotalByPriority   map[string]int `json:"total_by_priority" jsonschema_description:"Task count by priority"`
	AverageComplexity float64       `json:"average_complexity" jsonschema_description:"Average task complexity"`
}

// RegisterStatusTools registers all status and query tools with the MCP server
func RegisterStatusTools(mcpServer *server.MCPServer, projectManager manager.ProjectManager, sessionManager session.Manager) {
	// status_ready - Get ready tasks (pending state)
	statusReadyTool := mcp.NewTool("status_ready",
		mcp.WithDescription("Get tasks that are ready to work on (pending state) in the selected project"),
		mcp.WithInputSchema[StatusReadyRequest](),
		mcp.WithOutputSchema[StatusReadyResponse](),
	)
	mcpServer.AddTool(statusReadyTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args StatusReadyRequest) (StatusReadyResponse, error) {
		// Get project ID from request or session
		projectID := args.ProjectID
		if projectID == "" {
			sessionID := getSessionID(ctx)
			if sessionID == "" {
				return StatusReadyResponse{}, fmt.Errorf("no session found and no project_id provided")
			}

			sessionUUID, err := uuid.Parse(sessionID)
			if err != nil {
				return StatusReadyResponse{}, fmt.Errorf("invalid session ID: %w", err)
			}

			selectedProjectID, err := sessionManager.GetProject(sessionUUID)
			if err != nil {
				return StatusReadyResponse{}, fmt.Errorf("failed to get selected project: %w", err)
			}

			if selectedProjectID == nil {
				return StatusReadyResponse{}, fmt.Errorf("no project selected")
			}

			projectID = selectedProjectID.String()
		}

		// Parse project ID
		projectUUID, err := uuid.Parse(projectID)
		if err != nil {
			return StatusReadyResponse{}, fmt.Errorf("invalid project ID: %w", err)
		}

		// Get pending tasks (ready to work on)
		pendingTasks, err := projectManager.ListTasksByState(ctx, projectUUID, types.TaskStatePending)
		if err != nil {
			return StatusReadyResponse{}, fmt.Errorf("failed to get pending tasks: %w", err)
		}

		// Convert to task info
		limit := 20
		if args.Limit != nil {
			limit = *args.Limit
		}

		tasks := make([]TaskInfo, 0, len(pendingTasks))
		for i, task := range pendingTasks {
			if i >= limit {
				break
			}
			tasks = append(tasks, TaskInfo{
				ID:          task.ID.String(),
				Title:       task.Title,
				State:       string(task.State),
				Priority:    task.Priority.ToExternalString(),
				Complexity:  task.Complexity,
			})
		}

		return StatusReadyResponse{
			ProjectID: projectID,
			Tasks:     tasks,
			Total:     len(pendingTasks),
			Message:   fmt.Sprintf("Found %d ready tasks in project", len(pendingTasks)),
		}, nil
	}))

	// status_actionable - Get actionable tasks
	statusActionableTool := mcp.NewTool("status_actionable",
		mcp.WithDescription("Get tasks that are actionable (pending or in-progress) and recommend the next task to work on"),
		mcp.WithInputSchema[StatusActionableRequest](),
		mcp.WithOutputSchema[StatusActionableResponse](),
	)
	mcpServer.AddTool(statusActionableTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args StatusActionableRequest) (StatusActionableResponse, error) {
		// Get project ID from request or session
		projectID := args.ProjectID
		if projectID == "" {
			sessionID := getSessionID(ctx)
			if sessionID == "" {
				return StatusActionableResponse{}, fmt.Errorf("no session found and no project_id provided")
			}

			sessionUUID, err := uuid.Parse(sessionID)
			if err != nil {
				return StatusActionableResponse{}, fmt.Errorf("invalid session ID: %w", err)
			}

			selectedProjectID, err := sessionManager.GetProject(sessionUUID)
			if err != nil {
				return StatusActionableResponse{}, fmt.Errorf("failed to get selected project: %w", err)
			}

			if selectedProjectID == nil {
				return StatusActionableResponse{}, fmt.Errorf("no project selected")
			}

			projectID = selectedProjectID.String()
		}

		// Parse project ID
		projectUUID, err := uuid.Parse(projectID)
		if err != nil {
			return StatusActionableResponse{}, fmt.Errorf("invalid project ID: %w", err)
		}

		// Find next actionable task
		nextTask, err := projectManager.FindNextActionableTask(ctx, projectUUID)
		if err != nil {
			return StatusActionableResponse{}, fmt.Errorf("failed to find next actionable task: %w", err)
		}

		// Get all pending and in-progress tasks (actionable tasks)
		pendingTasks, err := projectManager.ListTasksByState(ctx, projectUUID, types.TaskStatePending)
		if err != nil {
			return StatusActionableResponse{}, fmt.Errorf("failed to get pending tasks: %w", err)
		}

		inProgressTasks, err := projectManager.ListTasksByState(ctx, projectUUID, types.TaskStateInProgress)
		if err != nil {
			return StatusActionableResponse{}, fmt.Errorf("failed to get in-progress tasks: %w", err)
		}

		// Combine tasks
		allActionableTasks := append(pendingTasks, inProgressTasks...)

		// Convert to task info
		limit := 20
		if args.Limit != nil {
			limit = *args.Limit
		}

		tasks := make([]TaskInfo, 0, len(allActionableTasks))
		for i, task := range allActionableTasks {
			if i >= limit {
				break
			}
			tasks = append(tasks, TaskInfo{
				ID:          task.ID.String(),
				Title:       task.Title,
				State:       string(task.State),
				Priority:    task.Priority.ToExternalString(),
				Complexity:  task.Complexity,
			})
		}

		message := fmt.Sprintf("Found %d actionable tasks in project", len(allActionableTasks))
		if nextTask != nil {
			message += fmt.Sprintf(". Next recommended task: %s", nextTask.Title)
		}

		return StatusActionableResponse{
			ProjectID: projectID,
			Tasks:     tasks,
			Total:     len(allActionableTasks),
			Message:   message,
		}, nil
	}))

	// status_project - Get project status
	statusProjectTool := mcp.NewTool("status_project",
		mcp.WithDescription("Get comprehensive status and statistics for a project"),
		mcp.WithInputSchema[StatusProjectRequest](),
		mcp.WithOutputSchema[StatusProjectResponse](),
	)
	mcpServer.AddTool(statusProjectTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args StatusProjectRequest) (StatusProjectResponse, error) {
		// Get project ID from request or session
		projectID := args.ProjectID
		if projectID == "" {
			sessionID := getSessionID(ctx)
			if sessionID == "" {
				return StatusProjectResponse{}, fmt.Errorf("no session found and no project_id provided")
			}

			sessionUUID, err := uuid.Parse(sessionID)
			if err != nil {
				return StatusProjectResponse{}, fmt.Errorf("invalid session ID: %w", err)
			}

			selectedProjectID, err := sessionManager.GetProject(sessionUUID)
			if err != nil {
				return StatusProjectResponse{}, fmt.Errorf("failed to get selected project: %w", err)
			}

			if selectedProjectID == nil {
				return StatusProjectResponse{}, fmt.Errorf("no project selected")
			}

			projectID = selectedProjectID.String()
		}

		// Parse project ID
		projectUUID, err := uuid.Parse(projectID)
		if err != nil {
			return StatusProjectResponse{}, fmt.Errorf("invalid project ID: %w", err)
		}

		// Get project details
		project, err := projectManager.GetProject(ctx, projectUUID)
		if err != nil {
			return StatusProjectResponse{}, fmt.Errorf("failed to get project: %w", err)
		}

		// Get project progress
		progress, err := projectManager.GetProjectProgress(ctx, projectUUID)
		if err != nil {
			return StatusProjectResponse{}, fmt.Errorf("failed to get project progress: %w", err)
		}

		// Get all tasks for statistics
		allTasks, err := projectManager.ListTasksForProject(ctx, projectUUID)
		if err != nil {
			return StatusProjectResponse{}, fmt.Errorf("failed to get project tasks: %w", err)
		}

		// Calculate task statistics
		taskStats := calculateTaskStats(allTasks)

		// Create project info
		projectInfo := ProjectInfo{
			ID:          project.ID.String(),
			Title:       project.Title,
			Description: project.Description,
			State:       string(project.State),
			CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}

		// Create progress info
		progressInfo := Progress{
			TotalTasks:      len(allTasks),
			CompletedTasks:  progress.CompletedTasks,
			PendingTasks:    progress.PendingTasks,
			InProgressTasks: progress.InProgressTasks,
			BlockedTasks:    progress.BlockedTasks,
			CompletionRate:  progress.OverallProgress,
		}

		return StatusProjectResponse{
			ProjectInfo: projectInfo,
			Progress:    progressInfo,
			TaskStats:   taskStats,
			Message:     fmt.Sprintf("Project status for %s", project.Title),
		}, nil
	}))
}

// calculateTaskStats calculates task statistics
func calculateTaskStats(tasks []*types.Task) TaskStats {
	stats := TaskStats{
		TotalByState:     make(map[string]int),
		TotalByPriority:  make(map[string]int),
	}

	totalComplexity := 0
	for _, task := range tasks {
		// Count by state
		state := string(task.State)
		stats.TotalByState[state]++

		// Count by priority
		priority := task.Priority.ToExternalString()
		stats.TotalByPriority[priority]++

		// Sum complexity for average
		totalComplexity += task.Complexity
	}

	// Calculate average complexity
	if len(tasks) > 0 {
		stats.AverageComplexity = float64(totalComplexity) / float64(len(tasks))
	}

	return stats
}