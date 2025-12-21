package tools

import (
	"context"
	"fmt"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mcp/utils"
	knotutils "github.com/denkhaus/knot/v2/internal/utils"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Status and query tools provide project status, task state queries, and actionable work discovery

// StatusBlockedRequest defines the request for getting blocked tasks
type StatusBlockedRequest struct {
	Limit *int `json:"limit,omitempty" jsonschema_description:"Maximum number of tasks to return (default: 20)"`
}

// StatusBlockedResponse defines the response for blocked tasks
type StatusBlockedResponse struct {
	ProjectID string          `json:"project_id" jsonschema_description:"Project ID"`
	Tasks     []BlockedTaskInfo `json:"tasks" jsonschema_description:"List of blocked tasks with reasons"`
	Total     int             `json:"total" jsonschema_description:"Total number of blocked tasks"`
	Message   string          `json:"message" jsonschema_description:"Status message"`
}

// BlockedTaskInfo extends TaskInfo with blocking reasons
type BlockedTaskInfo struct {
	TaskInfo
	BlockingReasons   []string `json:"blocking_reasons" jsonschema_description:"Reasons why this task is blocked"`
	DependenciesCount int      `json:"dependencies_count" jsonschema_description:"Number of dependencies"`
}

// StatusTreeRequest defines the request for task hierarchy
type StatusTreeRequest struct {
}

// StatusTreeResponse defines the response for task hierarchy
type StatusTreeResponse struct {
	ProjectID string      `json:"project_id" jsonschema_description:"Project ID"`
	Tree      []TreeNode  `json:"tree" jsonschema_description:"Task hierarchy tree"`
	Total     int         `json:"total" jsonschema_description:"Total number of tasks"`
	Message   string      `json:"message" jsonschema_description:"Status message"`
}

// TreeNode represents a node in the task hierarchy tree for MCP responses
type TreeNode struct {
	TaskInfo
	Children []TreeNode `json:"children,omitempty" jsonschema_description:"Child tasks"`
	Level    int        `json:"level" jsonschema_description:"Depth level in the tree"`
}


// StatusReadyRequest defines the request for getting ready tasks
type StatusReadyRequest struct {
	Limit *int `json:"limit,omitempty" jsonschema_description:"Maximum number of tasks to return (default: 20)"`
}

// StatusReadyResponse defines the response for ready tasks
type StatusReadyResponse struct {
	ProjectID string     `json:"project_id" jsonschema_description:"Project ID"`
	Tasks     []TaskInfo `json:"tasks" jsonschema_description:"List of ready tasks"`
	Total     int        `json:"total" jsonschema_description:"Total number of ready tasks"`
	Message   string     `json:"message" jsonschema_description:"Status message"`
}

// StatusActionableRequest defines the request for getting actionable tasks
type StatusActionableRequest struct {
	Limit *int `json:"limit,omitempty" jsonschema_description:"Maximum number of tasks to return (default: 20)"`
}

// StatusActionableResponse defines the response for actionable tasks
type StatusActionableResponse struct {
	ProjectID string     `json:"project_id" jsonschema_description:"Project ID"`
	Tasks     []TaskInfo `json:"tasks" jsonschema_description:"List of actionable tasks"`
	Total     int        `json:"total" jsonschema_description:"Total number of actionable tasks"`
	Message   string     `json:"message" jsonschema_description:"Status message"`
}

// StatusProjectRequest defines the request for getting project status
type StatusProjectRequest struct {
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
	TotalTasks      int     `json:"total_tasks" jsonschema_description:"Total number of tasks"`
	CompletedTasks  int     `json:"completed_tasks" jsonschema_description:"Number of completed tasks"`
	PendingTasks    int     `json:"pending_tasks" jsonschema_description:"Number of pending tasks"`
	InProgressTasks int     `json:"in_progress_tasks" jsonschema_description:"Number of in-progress tasks"`
	BlockedTasks    int     `json:"blocked_tasks" jsonschema_description:"Number of blocked tasks"`
	CompletionRate  float64 `json:"completion_rate" jsonschema_description:"Completion percentage (0-100)"`
}

// TaskStats defines task statistics
type TaskStats struct {
	TotalByState      map[string]int `json:"total_by_state" jsonschema_description:"Task count by state"`
	TotalByPriority   map[string]int `json:"total_by_priority" jsonschema_description:"Task count by priority"`
	AverageComplexity float64        `json:"average_complexity" jsonschema_description:"Average task complexity"`
}

// RegisterStatusTools registers all status and query tools with the MCP server
func RegisterStatusTools(mcpServer *server.MCPServer, projectManager manager.ProjectManager, sessionManager session.SessionManager) {
	// status_ready - Get ready tasks (pending state)
	statusReadyTool := mcp.NewTool("status_ready",
		mcp.WithDescription("Get tasks that are ready to work on (pending state) in the selected project"),
		mcp.WithInputSchema[StatusReadyRequest](),
		mcp.WithOutputSchema[StatusReadyResponse](),
	)
	mcpServer.AddTool(statusReadyTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args StatusReadyRequest) (StatusReadyResponse, error) {
		// Get project ID from session
		projectID, err := utils.GetSelectedProject(ctx, sessionManager)
		if err != nil {
			return StatusReadyResponse{}, err
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
				ID:         task.ID.String(),
				Title:      task.Title,
				State:      string(task.State),
				Priority:   task.Priority.ToExternalString(),
				Complexity: task.Complexity,
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
		// Get project ID from session
		projectID, err := utils.GetSelectedProject(ctx, sessionManager)
		if err != nil {
			return StatusActionableResponse{}, err
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
				ID:         task.ID.String(),
				Title:      task.Title,
				State:      string(task.State),
				Priority:   task.Priority.ToExternalString(),
				Complexity: task.Complexity,
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
		// Get project ID from session
		projectID, err := utils.GetSelectedProject(ctx, sessionManager)
		if err != nil {
			return StatusProjectResponse{}, err
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

	// status_blocked - Get blocked tasks with blocking reasons
	statusBlockedTool := mcp.NewTool("status_blocked",
		mcp.WithDescription("Get tasks that are blocked and explain why they are blocked"),
		mcp.WithInputSchema[StatusBlockedRequest](),
		mcp.WithOutputSchema[StatusBlockedResponse](),
	)
	mcpServer.AddTool(statusBlockedTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args StatusBlockedRequest) (StatusBlockedResponse, error) {
		// Get project ID from session
		projectID, err := utils.GetSelectedProject(ctx, sessionManager)
		if err != nil {
			return StatusBlockedResponse{}, err
		}

		// Parse project ID
		projectUUID, err := uuid.Parse(projectID)
		if err != nil {
			return StatusBlockedResponse{}, fmt.Errorf("invalid project ID: %w", err)
		}

		// Get blocked tasks
		blockedTasks, err := projectManager.ListTasksByState(ctx, projectUUID, types.TaskStateBlocked)
		if err != nil {
			return StatusBlockedResponse{}, fmt.Errorf("failed to get blocked tasks: %w", err)
		}

		// Convert to blocked task info with reasons
		limit := 20
		if args.Limit != nil {
			limit = *args.Limit
		}

		blockedTaskInfos := make([]BlockedTaskInfo, 0, len(blockedTasks))
		for i, task := range blockedTasks {
			if i >= limit {
				break
			}

			// Get dependencies to determine blocking reasons
			dependencies, err := projectManager.GetTaskDependencies(ctx, task.ID)
			if err != nil {
				dependencies = []*types.Task{} // Continue without dependencies if error
			}

			// Create blocking reasons
			blockingReasons := make([]string, 0, len(dependencies))
			for _, dep := range dependencies {
				if dep.State != types.TaskStateCompleted {
					blockingReasons = append(blockingReasons, fmt.Sprintf("Waiting for dependency: %s", dep.Title))
				}
			}

			// If no blocking dependencies found, check other potential reasons
			if len(blockingReasons) == 0 {
				blockingReasons = append(blockingReasons, "Task marked as blocked - check task notes for details")
			}

			blockedTaskInfos = append(blockedTaskInfos, BlockedTaskInfo{
				TaskInfo: TaskInfo{
					ID:         task.ID.String(),
					Title:      task.Title,
					State:      string(task.State),
					Priority:   task.Priority.ToExternalString(),
					Complexity: task.Complexity,
				},
				BlockingReasons: blockingReasons,
				DependenciesCount: len(dependencies),
			})
		}

		return StatusBlockedResponse{
			ProjectID: projectID,
			Tasks:     blockedTaskInfos,
			Total:     len(blockedTasks),
			Message:   fmt.Sprintf("Found %d blocked tasks in project", len(blockedTasks)),
		}, nil
	}))

	// status_tree - Show task hierarchy with indentation
	statusTreeTool := mcp.NewTool("status_tree",
		mcp.WithDescription("Show task hierarchy with indentation based on parent-child relationships"),
		mcp.WithInputSchema[StatusTreeRequest](),
		mcp.WithOutputSchema[StatusTreeResponse](),
	)
	mcpServer.AddTool(statusTreeTool, mcp.NewStructuredToolHandler(func(ctx context.Context, request mcp.CallToolRequest, args StatusTreeRequest) (StatusTreeResponse, error) {
		// Get project ID from session
		projectID, err := utils.GetSelectedProject(ctx, sessionManager)
		if err != nil {
			return StatusTreeResponse{}, err
		}

		// Parse project ID
		projectUUID, err := uuid.Parse(projectID)
		if err != nil {
			return StatusTreeResponse{}, fmt.Errorf("invalid project ID: %w", err)
		}

		// Get all tasks for the project
		allTasks, err := projectManager.ListTasksForProject(ctx, projectUUID)
		if err != nil {
			return StatusTreeResponse{}, fmt.Errorf("failed to get project tasks: %w", err)
		}

		// Build tree structure using utility function
		knotTreeNodes := knotutils.BuildTaskTreeFromTasks(allTasks)

		// Convert utility tree nodes to MCP tree nodes
		treeNodes := convertToMCPtreeNodes(knotTreeNodes)

		return StatusTreeResponse{
			ProjectID: projectID,
			Tree:      treeNodes,
			Total:     len(allTasks),
			Message:   fmt.Sprintf("Task hierarchy for project with %d tasks", len(allTasks)),
		}, nil
	}))

	}

// calculateTaskStats calculates task statistics
func calculateTaskStats(tasks []*types.Task) TaskStats {
	stats := TaskStats{
		TotalByState:    make(map[string]int),
		TotalByPriority: make(map[string]int),
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

// convertToMCPtreeNodes converts utility TaskTreeNode to MCP TreeNode
func convertToMCPtreeNodes(knotNodes []knotutils.TaskTreeNode) []TreeNode {
	var treeNodes []TreeNode
	for _, knotNode := range knotNodes {
		node := TreeNode{
			TaskInfo: TaskInfo{
				ID:         knotNode.Task.ID.String(),
				Title:      knotNode.Task.Title,
				State:      string(knotNode.Task.State),
				Priority:   knotNode.Task.Priority.ToExternalString(),
				Complexity: knotNode.Task.Complexity,
			},
			Level:    knotNode.Level,
			Children: convertToMCPtreeNodes(knotNode.Children),
		}
		treeNodes = append(treeNodes, node)
	}
	return treeNodes
}
