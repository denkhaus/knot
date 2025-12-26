package selection

import (
	"time"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
)

// Config holds configuration for task selection (dependency-aware only)
type Config struct {
	// Weight factors for scoring (should sum to approximately 1.0)
	DependentCountWeight float64 `json:"dependent_count_weight"` // How much to weight tasks that unblock others
	PriorityWeight       float64 `json:"priority_weight"`        // How much to weight explicit priority
	DepthFirstWeight     float64 `json:"depth_first_weight"`     // How much to prefer completing subtasks first
	CriticalPathWeight   float64 `json:"critical_path_weight"`   // How much to weight critical path position

	// Behavioral flags
	PreferInProgress bool `json:"prefer_in_progress"` // Whether to prioritize in-progress tasks

	// Advanced options
	MaxDependencyDepth int           `json:"max_dependency_depth"` // Maximum depth to analyze in dependency chains
	ScoreThreshold     float64       `json:"score_threshold"`      // Minimum score threshold for task selection
	CacheGraphs        bool          `json:"cache_graphs"`         // Whether to cache dependency graphs
	CacheDuration      time.Duration `json:"cache_duration"`       // How long to cache graphs
}

// DefaultConfig returns a balanced default configuration
func DefaultConfig() *Config {
	return &Config{
		DependentCountWeight: 0.4,
		PriorityWeight:       0.3,
		DepthFirstWeight:     0.2,
		CriticalPathWeight:   0.1,
		PreferInProgress:     true,
		MaxDependencyDepth:   10,
		ScoreThreshold:       0.0,
		CacheGraphs:          true,
		CacheDuration:        5 * time.Minute,
	}
}

// TaskScore represents a scored task with selection metrics
type TaskScore struct {
	Task               *types.Task        `json:"task"`
	DependentCount     int                `json:"dependent_count"`      // Number of tasks that depend on this task
	UnblockedTaskCount int                `json:"unblocked_task_count"` // Total tasks that would become actionable
	DependencyDepth    int                `json:"dependency_depth"`     // Depth in the dependency chain (0 = no deps)
	CriticalPathLength int                `json:"critical_path_length"` // Length of longest dependency chain through this task
	HierarchyDepth     int                `json:"hierarchy_depth"`      // Depth in parent-child hierarchy
	Priority           types.TaskPriority `json:"priority"`             // Explicit task priority
	Score              float64            `json:"score"`                // Calculated selection score
	SelectionReason    string             `json:"selection_reason"`     // Why this task was selected
	CalculatedAt       time.Time          `json:"calculated_at"`        // When the score was calculated
}

// DependencyNode represents a task in the dependency graph
type DependencyNode struct {
	TaskID             uuid.UUID   `json:"task_id"`
	Task               *types.Task `json:"task"`
	Dependencies       []uuid.UUID `json:"dependencies"` // Tasks this one depends on
	Dependents         []uuid.UUID `json:"dependents"`   // Tasks that depend on this one
	Children           []uuid.UUID `json:"children"`     // Subtasks of this task
	Parent             *uuid.UUID  `json:"parent"`       // Parent task ID
	DependentCount     int         `json:"dependent_count"`
	ChildCount         int         `json:"child_count"`
	DependencyDepth    int         `json:"dependency_depth"`     // How deep in dependency chain
	CriticalPathLength int         `json:"critical_path_length"` // Longest path through this node
	UnblockedCount     int         `json:"unblocked_count"`      // Total tasks that would become actionable
	IsActionable       bool        `json:"is_actionable"`        // Whether the task can be worked on now
	BlockingReasons    []string    `json:"blocking_reasons"`     // Why the task is blocked (if it is)
}

// DependencyGraph represents the complete task dependency graph
type DependencyGraph struct {
	Nodes           map[uuid.UUID]*DependencyNode `json:"nodes"`
	RootTasks       []uuid.UUID                   `json:"root_tasks"`       // Tasks with no dependencies
	LeafTasks       []uuid.UUID                   `json:"leaf_tasks"`       // Tasks with no dependents
	CriticalPath    []uuid.UUID                   `json:"critical_path"`    // Longest dependency chain
	HasCycles       bool                          `json:"has_cycles"`       // Whether the graph contains cycles
	CyclicTasks     []uuid.UUID                   `json:"cyclic_tasks"`     // Tasks involved in cycles
	AnalyzedAt      time.Time                     `json:"analyzed_at"`      // When the graph was built
	TaskCount       int                           `json:"task_count"`       // Total number of tasks
	ActionableCount int                           `json:"actionable_count"` // Number of currently actionable tasks
}

// SelectionResult contains the result of task selection
type SelectionResult struct {
	SelectedTask  *types.Task  `json:"selected_task"`
	Score         *TaskScore   `json:"score"`
	Reason        string       `json:"reason"`
	Alternatives  []*TaskScore `json:"alternatives"` // Other tasks that could be selected
	SelectedAt    time.Time    `json:"selected_at"`
	ExecutionTime time.Duration `json:"execution_time"` // How long selection took
}

// ValidationError represents a validation error in the selection process
type ValidationError struct {
	TaskID  uuid.UUID `json:"task_id"`
	Message string    `json:"message"`
	Type    string    `json:"type"`
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return e.Message
}

// SelectionError represents errors that occur during task selection
type SelectionError struct {
	Type           string            `json:"type"`
	Message        string            `json:"message"`
	TaskID         *uuid.UUID        `json:"task_id,omitempty"`
	ValidationErrs []ValidationError `json:"validation_errors,omitempty"`
}

// Error implements the error interface
func (e SelectionError) Error() string {
	return e.Message
}

// Common error types
const (
	ErrorTypeNoTasks       = "no_tasks"
	ErrorTypeNoActionable  = "no_actionable"
	ErrorTypeDeadlock      = "deadlock"
	ErrorTypeInvalidConfig = "invalid_config"
	ErrorTypeCircularDep   = "circular_dependency"
	ErrorTypeValidation    = "validation"
)
