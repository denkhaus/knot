package manager

import (
	"context"
	"time"

	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
)

// ProjectManager defines the public interface for project task management
type ProjectManager interface {
	// Project operations
	CreateProject(ctx context.Context, title, description, actor string) (*types.Project, error)
	GetProject(ctx context.Context, projectID uuid.UUID) (*types.Project, error)
	UpdateProject(ctx context.Context, projectID uuid.UUID, title, description string, actor string) (*types.Project, error)
	UpdateProjectDescription(ctx context.Context, projectID uuid.UUID, description string, actor string) (*types.Project, error)
	UpdateProjectState(ctx context.Context, projectID uuid.UUID, state types.ProjectState, actor string) (*types.Project, error)
	DeleteProject(ctx context.Context, projectID uuid.UUID) error
	ListProjects(ctx context.Context) ([]*types.Project, error)

	// Task operations
	CreateTask(ctx context.Context, projectID uuid.UUID, parentID *uuid.UUID, title, description string, complexity int, priority types.TaskPriority, actor string) (*types.Task, error)
	GetTask(ctx context.Context, taskID uuid.UUID) (*types.Task, error)
	UpdateTask(ctx context.Context, taskID uuid.UUID, title, description string, complexity int, state types.TaskState, actor string) (*types.Task, error)
	UpdateTaskDescription(ctx context.Context, taskID uuid.UUID, description string, actor string) (*types.Task, error)
	UpdateTaskTitle(ctx context.Context, taskID uuid.UUID, title string, actor string) (*types.Task, error)
	UpdateTaskPriority(ctx context.Context, taskID uuid.UUID, priority types.TaskPriority, actor string) (*types.Task, error)
	UpdateTaskState(ctx context.Context, taskID uuid.UUID, state types.TaskState, actor string) (*types.Task, error)
	DeleteTask(ctx context.Context, taskID uuid.UUID, actor string) error
	DeleteTaskSubtree(ctx context.Context, taskID uuid.UUID, actor string) error

	// Task queries and analysis
	GetParentTask(ctx context.Context, taskID uuid.UUID) (*types.Task, error)
	GetChildTasks(ctx context.Context, taskID uuid.UUID) ([]*types.Task, error)
	GetRootTasks(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error)
	ListTasksForProject(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error)
	FindNextActionableTask(ctx context.Context, projectID uuid.UUID) (*types.Task, error)
	FindTasksNeedingBreakdown(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error)
	GetProjectProgress(ctx context.Context, projectID uuid.UUID) (*types.ProjectProgress, error)
	ListTasksByState(ctx context.Context, projectID uuid.UUID, state types.TaskState) ([]*types.Task, error)
	BulkUpdateTasks(ctx context.Context, taskIDs []uuid.UUID, updates types.TaskUpdates) error
	DuplicateTask(ctx context.Context, taskID uuid.UUID, newProjectID uuid.UUID) (*types.Task, error)
	SetTaskEstimate(ctx context.Context, taskID uuid.UUID, estimate int64) (*types.Task, error)

	// Agent assignment management
	AssignTaskToAgent(ctx context.Context, taskID uuid.UUID, agentID uuid.UUID) (*types.Task, error)
	UnassignTaskFromAgent(ctx context.Context, taskID uuid.UUID) (*types.Task, error)
	ListTasksByAgent(ctx context.Context, projectID uuid.UUID, agentID uuid.UUID) ([]*types.Task, error)
	ListUnassignedTasks(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error)

	// Dependency management
	AddTaskDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID, actor string) (*types.Task, error)
	RemoveTaskDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID, actor string) (*types.Task, error)
	GetTaskDependencies(ctx context.Context, taskID uuid.UUID) ([]*types.Task, error)
	GetDependentTasks(ctx context.Context, taskID uuid.UUID) ([]*types.Task, error)

	// Configuration
	GetConfig() *Config
	UpdateConfig(config *Config)
	LoadConfigFromFile() error
	SaveConfigToFile() error

	// Utility methods
	GetCurrentTime() time.Time
}

// ToolSetProvider defines the interface for creating project task tool sets (not used in CLI)
// type ToolSetProvider interface {
//	CreateToolSet(opts ...Option) (tool.ToolSet, error)
// }

// Config holds configuration for the task management system
type Config struct {
	MaxTasksPerDepth     int  // Maximum tasks allowed per depth level (applies to all depths)
	ComplexityThreshold  int  // Threshold for task breakdown suggestions
	MaxDepth             int  // Maximum allowed depth
	MaxDescriptionLength int  // Maximum length for descriptions
	AutoReduceComplexity bool // Automatically reduce parent task complexity when subtasks are added
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *Config {
	return &Config{
		MaxTasksPerDepth:     100,  // Increased from 20 to 100 for better scalability
		ComplexityThreshold:  8,
		MaxDepth:             5,
		MaxDescriptionLength: 2000, // Default maximum description length
		AutoReduceComplexity: true, // Enable auto-reduce by default
	}
}
