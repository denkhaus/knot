package selection

import (
	"time"

	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
)

// Strategy represents different task selection strategies
type Strategy int

const (
	StrategyCreationOrder Strategy = iota
	StrategyDependencyAware
	StrategyDepthFirst
	StrategyPriority
	StrategyCriticalPath
)

// String returns the string representation of a strategy
func (s Strategy) String() string {
	switch s {
	case StrategyCreationOrder:
		return "creation-order"
	case StrategyDependencyAware:
		return "dependency-aware"
	case StrategyDepthFirst:
		return "depth-first"
	case StrategyPriority:
		return "priority"
	case StrategyCriticalPath:
		return "critical-path"
	default:
		return "unknown"
	}
}

// ParseStrategy parses a string into a Strategy
func ParseStrategy(s string) Strategy {
	switch s {
	case "creation-order":
		return StrategyCreationOrder
	case "dependency-aware":
		return StrategyDependencyAware
	case "depth-first":
		return StrategyDepthFirst
	case "priority":
		return StrategyPriority
	case "critical-path":
		return StrategyCriticalPath
	default:
		return StrategyDependencyAware // default
	}
}

// Config holds configuration for task selection
type Config struct {
	// Selection strategy
	Strategy Strategy `json:"strategy"`

	// Weight factors for scoring (should sum to 1.0)
	Weights Weights `json:"weights"`

	// Behavioral flags
	Behavior BehaviorConfig `json:"behavior"`

	// Advanced options
	Advanced AdvancedConfig `json:"advanced"`
}

// Weights defines scoring weight factors
type Weights struct {
	DependentCount float64 `json:"dependent_count"` // How much to weight tasks that unblock others
	Priority       float64 `json:"priority"`        // How much to weight explicit priority
	DepthFirst     float64 `json:"depth_first"`     // How much to prefer completing subtasks first
	CriticalPath   float64 `json:"critical_path"`   // How much to weight critical path position
}

// BehaviorConfig defines behavioral options
type BehaviorConfig struct {
	AllowParentWithSubtasks bool `json:"allow_parent_with_subtasks"` // Whether to allow parent tasks when subtasks exist
	PreferInProgress        bool `json:"prefer_in_progress"`         // Whether to prioritize in-progress tasks
	BreakTiesByCreation     bool `json:"break_ties_by_creation"`     // Use creation time as final tiebreaker
	StrictDependencies      bool `json:"strict_dependencies"`        // Whether to strictly enforce dependency order
}

// AdvancedConfig defines advanced configuration options
type AdvancedConfig struct {
	MaxDependencyDepth int           `json:"max_dependency_depth"` // Maximum depth to analyze in dependency chains
	ScoreThreshold     float64       `json:"score_threshold"`      // Minimum score threshold for task selection
	CacheGraphs        bool          `json:"cache_graphs"`         // Whether to cache dependency graphs
	CacheDuration      time.Duration `json:"cache_duration"`       // How long to cache graphs
}

// DefaultConfig returns a balanced default configuration
func DefaultConfig() *Config {
	return &Config{
		Strategy: StrategyDependencyAware,
		Weights: Weights{
			DependentCount: 0.4,
			Priority:       0.3,
			DepthFirst:     0.2,
			CriticalPath:   0.1,
		},
		Behavior: BehaviorConfig{
			AllowParentWithSubtasks: false,
			PreferInProgress:        true,
			BreakTiesByCreation:     true,
			StrictDependencies:      true,
		},
		Advanced: AdvancedConfig{
			MaxDependencyDepth: 10,
			ScoreThreshold:     0.0,
			CacheGraphs:        true,
			CacheDuration:      5 * time.Minute,
		},
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
	SelectedTask  *types.Task   `json:"selected_task"`
	Score         *TaskScore    `json:"score"`
	Strategy      Strategy      `json:"strategy"`
	Reason        string        `json:"reason"`
	Alternatives  []*TaskScore  `json:"alternatives"` // Other tasks that could be selected
	SelectedAt    time.Time     `json:"selected_at"`
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
