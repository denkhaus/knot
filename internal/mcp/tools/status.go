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
	"github.com/denkhaus/knot/v2/internal/selection"
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
	Children []TreeNode `json:"children,omitempty" jsonschema_description:"Child tasks" jsonschema:"{\"type\":\"array\",\"items\":{\"$ref\":\"#/definitions/TreeNode\"}}"`
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

		// Get all tasks to check dependencies
		allTasks, err := projectManager.ListTasksForProject(ctx, projectUUID)
		if err != nil {
			return StatusReadyResponse{}, fmt.Errorf("failed to get project tasks: %w", err)
		}

		// Extract task IDs for dependency loading
		taskIDs := make([]uuid.UUID, len(allTasks))
		for i, task := range allTasks {
			taskIDs[i] = task.ID
		}

		// Get tasks with dependencies populated
		tasksWithDeps, err := projectManager.GetTasksWithDependencies(ctx, taskIDs)
		if err != nil {
			return StatusReadyResponse{}, fmt.Errorf("failed to get tasks with dependencies: %w", err)
		}

		// Filter for pending tasks whose dependencies are completed
		readyTasks := make([]*types.Task, 0)
		for _, task := range tasksWithDeps {
			if task.State == types.TaskStatePending {
				// Check if all dependencies are completed
				if areTaskDependenciesCompleted(task, tasksWithDeps) {
					readyTasks = append(readyTasks, task)
				}
			}
		}

		// Convert to task info
		limit := 20
		if args.Limit != nil {
			limit = *args.Limit
		}

		tasks := make([]TaskInfo, 0, len(readyTasks))
		for i, task := range readyTasks {
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
			Total:     len(readyTasks),
			Message:   fmt.Sprintf("Found %d ready tasks in project", len(readyTasks)),
		}, nil
	}))

	// status_actionable - Get actionable tasks using dependency-aware selection
	statusActionableTool := mcp.NewTool("status_actionable",
		mcp.WithDescription("Get the next actionable task using dependency-aware selection"),
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

		// Get all tasks in the project
		allTasks, err := projectManager.ListTasksForProject(ctx, projectUUID)
		if err != nil {
			return StatusActionableResponse{}, fmt.Errorf("failed to get project tasks: %w", err)
		}

		// Use the same selection logic as the CLI actionable command
		config := selection.DefaultConfig()
		selector, err := selection.NewTaskSelector(config)
		if err != nil {
			return StatusActionableResponse{}, fmt.Errorf("failed to create task selector: %w", err)
		}

		// Select next actionable task
		selectedTask, err := selector.SelectNextActionableTask(allTasks)
		if err != nil {
			// Handle specific error types
			if selErr, ok := err.(*selection.SelectionError); ok {
				switch selErr.Type {
				case selection.ErrorTypeNoTasks:
					return StatusActionableResponse{
						ProjectID: projectID,
						Tasks:     []TaskInfo{},
						Total:     0,
						Message:   "No tasks found in project",
					}, nil
				case selection.ErrorTypeNoActionable:
					return StatusActionableResponse{
						ProjectID: projectID,
						Tasks:     []TaskInfo{},
						Total:     0,
						Message:   "No actionable tasks available",
					}, nil
				case selection.ErrorTypeDeadlock:
					return StatusActionableResponse{
						ProjectID: projectID,
						Tasks:     []TaskInfo{},
						Total:     0,
						Message:   fmt.Sprintf("No actionable tasks found: %s", selErr.Message),
					}, nil
				case selection.ErrorTypeCircularDep:
					return StatusActionableResponse{
						ProjectID: projectID,
						Tasks:     []TaskInfo{},
						Total:     0,
						Message:   fmt.Sprintf("Circular dependencies detected: %s", selErr.Message),
					}, nil
				default:
					return StatusActionableResponse{}, fmt.Errorf("task selection failed: %w", err)
				}
			}
			return StatusActionableResponse{}, fmt.Errorf("failed to select actionable task: %w", err)
		}

		// Get selection result for additional context
		result := selector.GetLastResult()

		// Build response with selected task
		selectedTaskInfo := TaskInfo{
			ID:         selectedTask.ID.String(),
			Title:      selectedTask.Title,
			State:      string(selectedTask.State),
			Priority:   selectedTask.Priority.ToExternalString(),
			Complexity: selectedTask.Complexity,
		}

		// Build message with selection reasoning
		message := result.Reason
		if result.Score.UnblockedTaskCount > 0 {
			message += fmt.Sprintf(" | Will unblock: %d task(s)", result.Score.UnblockedTaskCount)
		}
		if result.Score.DependentCount > 0 {
			message += fmt.Sprintf(" | Dependent tasks: %d", result.Score.DependentCount)
		}

		return StatusActionableResponse{
			ProjectID: projectID,
			Tasks:     []TaskInfo{selectedTaskInfo},
			Total:     1,
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

		// Get all tasks to check for blocking dependencies
		allTasks, err := projectManager.ListTasksForProject(ctx, projectUUID)
		if err != nil {
			return StatusBlockedResponse{}, fmt.Errorf("failed to get project tasks: %w", err)
		}

		// Extract task IDs for dependency loading
		taskIDs := make([]uuid.UUID, len(allTasks))
		for i, task := range allTasks {
			taskIDs[i] = task.ID
		}

		// Get tasks with dependencies populated
		tasksWithDeps, err := projectManager.GetTasksWithDependencies(ctx, taskIDs)
		if err != nil {
			return StatusBlockedResponse{}, fmt.Errorf("failed to get tasks with dependencies: %w", err)
		}

		// Find tasks that are pending or in-progress but have incomplete dependencies
		blockedTasks := make([]*types.Task, 0)
		for _, task := range tasksWithDeps {
			if (task.State == types.TaskStatePending || task.State == types.TaskStateInProgress) &&
				!areTaskDependenciesCompleted(task, tasksWithDeps) {
				blockedTasks = append(blockedTasks, task)
			}
		}

		// Create a task map for dependency lookup
		taskMap := make(map[uuid.UUID]*types.Task)
		for _, t := range tasksWithDeps {
			taskMap[t.ID] = t
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

			// Create blocking reasons from incomplete dependencies
			blockingReasons := make([]string, 0)
			for _, depID := range task.Dependencies {
				if depTask, exists := taskMap[depID]; exists {
					if depTask.State != types.TaskStateCompleted {
						blockingReasons = append(blockingReasons, fmt.Sprintf("Waiting for dependency: %s", depTask.Title))
					}
				}
			}

			blockedTaskInfos = append(blockedTaskInfos, BlockedTaskInfo{
				TaskInfo: TaskInfo{
					ID:         task.ID.String(),
					Title:      task.Title,
					State:      string(task.State),
					Priority:   task.Priority.ToExternalString(),
					Complexity: task.Complexity,
				},
				BlockingReasons:   blockingReasons,
				DependenciesCount: len(task.Dependencies),
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
	// Temporarily disabled due to recursive schema generation issues
	// TODO: Fix TreeNode recursive schema with proper $ref definitions
	/*
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
	*/

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

// areTaskDependenciesCompleted checks if all of a task's dependencies are in completed state
func areTaskDependenciesCompleted(task *types.Task, allTasks []*types.Task) bool {
	// Create a task map for quick lookup
	taskMap := make(map[uuid.UUID]*types.Task)
	for _, t := range allTasks {
		taskMap[t.ID] = t
	}

	// Check if all dependencies are completed
	for _, depID := range task.Dependencies {
		depTask, exists := taskMap[depID]
		if !exists {
			// Missing dependency means not ready
			return false
		}
		if depTask.State != types.TaskStateCompleted {
			return false
		}
	}

	return true
}
