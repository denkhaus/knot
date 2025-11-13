// Package types defines the core domain models and interfaces for the KNOT project management system.
//
// This package contains the fundamental data structures used throughout the application,
// including projects, tasks, dependencies, and repository interfaces. All types are
// designed to be serializable to JSON and use UUIDs for primary keys.
//
// Key Concepts:
//   - Projects: Top-level containers for related tasks
//   - Tasks: Individual work units with hierarchical relationships
//   - Dependencies: Task execution order constraints
//   - States: Task and project lifecycle management
//
// Example Usage:
//
//	project := &types.Project{
//		Title:       "My Project",
//		Description: "A sample project",
//		State:       types.ProjectStateActive,
//	}
//
//	task := &types.Task{
//		ProjectID:  project.ID,
//		Title:      "Sample Task",
//		Complexity: 5,
//		State:      types.TaskStatePending,
//	}
package types

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TaskState represents the current state of a task in its lifecycle.
// Tasks progress through these states as work is completed.
//
// State Flow:
//
//	pending → in-progress → completed
//	pending → blocked → in-progress → completed
//	any state → cancelled
//	any state → deletion-pending → deleted
type TaskState string

const (
	// TaskStatePending indicates the task is ready to be started.
	TaskStatePending TaskState = "pending"

	// TaskStateInProgress indicates the task is currently being worked on.
	TaskStateInProgress TaskState = "in-progress"

	// TaskStateCompleted indicates the task has been finished successfully.
	TaskStateCompleted TaskState = "completed"

	// TaskStateBlocked indicates the task cannot proceed due to dependencies.
	TaskStateBlocked TaskState = "blocked"

	// TaskStateCancelled indicates the task was cancelled before completion.
	TaskStateCancelled TaskState = "cancelled"

	// TaskStateDeletionPending indicates the task is marked for deletion.
	TaskStateDeletionPending TaskState = "deletion-pending"
)

// TaskPriority represents the priority level of a task.
// Priority helps determine which tasks should be worked on first when multiple tasks are available.
type TaskPriority string

const (
	// TaskPriorityLow indicates the task has low priority and can be deferred.
	TaskPriorityLow TaskPriority = "low"

	// TaskPriorityMedium indicates the task has normal priority.
	TaskPriorityMedium TaskPriority = "medium"

	// TaskPriorityHigh indicates the task has high priority and should be completed soon.
	TaskPriorityHigh TaskPriority = "high"
)

// Task represents a single task in the project hierarchy.
//
// Tasks are the fundamental work units in KNOT and support:
//   - Hierarchical relationships via ParentID
//   - Dependency management via Dependencies and Dependents
//   - Complexity-based breakdown decisions
//   - Priority-based ordering
//   - Agent assignment for automated workflows
//   - Time tracking and estimation
//
// Example:
//
//	task := &types.Task{
//		ProjectID:   projectID,
//		Title:       "Implement user authentication",
//		Description: "Add login and registration functionality",
//		Complexity:  8, // High complexity - should be broken down
//		Priority:    types.TaskPriorityHigh,
//		State:       types.TaskStatePending,
//		Dependencies: []uuid.UUID{designTaskID},
//	}
type Task struct {
	ID            uuid.UUID    `json:"id"`
	ProjectID     uuid.UUID    `json:"project_id"`
	ParentID      *uuid.UUID   `json:"parent_id,omitempty"` // nil for root tasks
	Title         string       `json:"title"`
	Description   string       `json:"description"`
	State         TaskState    `json:"state"`
	Priority      TaskPriority `json:"priority"`                 // Task priority level
	Complexity    int          `json:"complexity"`               // Used for breakdown decisions
	Depth         int          `json:"depth"`                    // 0 for root tasks
	Estimate      *int64       `json:"estimate,omitempty"`       // Time estimate in minutes
	AssignedAgent *uuid.UUID   `json:"assigned_agent,omitempty"` // Agent assigned to this task
	Dependencies  []uuid.UUID  `json:"dependencies,omitempty"`   // Tasks this task depends on
	Dependents    []uuid.UUID  `json:"dependents,omitempty"`     // Tasks that depend on this task
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	CreatedBy     string       `json:"created_by,omitempty"` // Actor who created the task
	UpdatedBy     string       `json:"updated_by,omitempty"` // Actor who last updated the task
	CompletedAt   *time.Time   `json:"completed_at,omitempty"`
}

// ProjectState represents the current state of a project
type ProjectState string

const (
	ProjectStateActive          ProjectState = "active"
	ProjectStateCompleted       ProjectState = "completed"
	ProjectStateArchived        ProjectState = "archived"
	ProjectStateDeletionPending ProjectState = "deletion-pending"
)

// Project represents a project containing hierarchical tasks
type Project struct {
	ID          uuid.UUID    `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	State       ProjectState `json:"state"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	CreatedBy   string       `json:"created_by,omitempty"` // Actor who created the project
	UpdatedBy   string       `json:"updated_by,omitempty"` // Actor who last updated the project
	// Progress metrics
	TotalTasks     int     `json:"total_tasks"`
	CompletedTasks int     `json:"completed_tasks"`
	Progress       float64 `json:"progress"` // Percentage (0-100)
}

// ProjectProgress represents detailed progress information
type ProjectProgress struct {
	ProjectID       uuid.UUID   `json:"project_id"`
	TotalTasks      int         `json:"total_tasks"`
	CompletedTasks  int         `json:"completed_tasks"`
	InProgressTasks int         `json:"in_progress_tasks"`
	PendingTasks    int         `json:"pending_tasks"`
	BlockedTasks    int         `json:"blocked_tasks"`
	CancelledTasks  int         `json:"cancelled_tasks"`
	OverallProgress float64     `json:"overall_progress"`
	TasksByDepth    map[int]int `json:"tasks_by_depth"`
}

// TaskFilter represents filtering options for task queries
type TaskFilter struct {
	ProjectID     *uuid.UUID    `json:"project_id,omitempty"`
	ParentID      *uuid.UUID    `json:"parent_id,omitempty"`
	State         *TaskState    `json:"state,omitempty"`
	Priority      *TaskPriority `json:"priority,omitempty"`
	MinDepth      *int          `json:"min_depth,omitempty"`
	MaxDepth      *int          `json:"max_depth,omitempty"`
	MinComplexity *int          `json:"min_complexity,omitempty"`
	MaxComplexity *int          `json:"max_complexity,omitempty"`
}

// TaskUpdates represents the fields that can be updated in bulk
type TaskUpdates struct {
	State      *TaskState    `json:"state,omitempty"`
	Priority   *TaskPriority `json:"priority,omitempty"`
	Complexity *int          `json:"complexity,omitempty"`
}

// Repository defines the interface for task and project persistence and retrieval.
//
// This interface provides a complete abstraction layer for data storage operations,
// allowing different storage backends (SQLite, in-memory, PostgreSQL, etc.) to be
// used interchangeably. All operations accept context.Context for cancellation
// and timeout control.
//
// Implementations should handle:
//   - Transaction management for data consistency
//   - Concurrent access safely
//   - Proper error handling and wrapping
//   - Performance optimization for large datasets
//
// Example Usage:
//
//	repo := sqlite.NewRepository("data.db", sqlite.WithAutoMigrate(true))
//	defer repo.Close()
//
//	project := &types.Project{Title: "My Project", State: types.ProjectStateActive}
//	err := repo.CreateProject(ctx, project)
//	if err != nil {
//		return fmt.Errorf("failed to create project: %w", err)
//	}
type Repository interface {
	// Project operations
	// CreateProject persists a new project to the storage backend.
	CreateProject(ctx context.Context, project *Project) error

	// GetProject retrieves a project by its ID. Returns an error if not found.
	GetProject(ctx context.Context, id uuid.UUID) (*Project, error)

	// UpdateProject updates an existing project with new values.
	UpdateProject(ctx context.Context, project *Project) error

	// DeleteProject removes a project and all its associated tasks.
	DeleteProject(ctx context.Context, id uuid.UUID) error

	// ListProjects returns all projects sorted by creation date.
	ListProjects(ctx context.Context) ([]*Project, error)

	// Task operations
	CreateTask(ctx context.Context, task *Task) error
	GetTask(ctx context.Context, id uuid.UUID) (*Task, error)
	UpdateTask(ctx context.Context, task *Task) error
	DeleteTask(ctx context.Context, id uuid.UUID) error

	// Task queries
	ListTasks(ctx context.Context, filter TaskFilter) ([]*Task, error)
	GetTasksByProject(ctx context.Context, projectID uuid.UUID) ([]*Task, error)
	GetTasksByParent(ctx context.Context, parentID uuid.UUID) ([]*Task, error)
	GetRootTasks(ctx context.Context, projectID uuid.UUID) ([]*Task, error)
	GetParentTask(ctx context.Context, taskID uuid.UUID) (*Task, error)

	// Hierarchy operations
	DeleteTaskSubtree(ctx context.Context, taskID uuid.UUID) error

	// Dependency management
	AddTaskDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID) (*Task, error)
	RemoveTaskDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID) (*Task, error)
	GetTaskDependencies(ctx context.Context, taskID uuid.UUID) ([]*Task, error)
	GetDependentTasks(ctx context.Context, taskID uuid.UUID) ([]*Task, error)

	// Metrics and analysis
	GetProjectProgress(ctx context.Context, projectID uuid.UUID) (*ProjectProgress, error)
	GetTaskCountByDepth(ctx context.Context, projectID uuid.UUID, maxDepth int) (map[int]int, error)

	// Project context management
	GetSelectedProject(ctx context.Context) (*uuid.UUID, error)
	SetSelectedProject(ctx context.Context, projectID uuid.UUID, actor string) error
	ClearSelectedProject(ctx context.Context) error
	HasSelectedProject(ctx context.Context) (bool, error)
}
