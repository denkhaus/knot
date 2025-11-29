package visualization

import (
	"github.com/denkhaus/knot/v2/internal/types"
)

// TaskIcon represents visual indicators for tasks
type TaskIcon string

const (
	// IconCompleted represents a completed task
	IconCompleted TaskIcon = "[DONE]"
	// IconInProgress represents a task currently being worked on
	IconInProgress TaskIcon = "[WORK]"
	// IconReady represents a task that's ready to start
	IconReady TaskIcon = "[READY]"
	// IconBlocked represents a task that's blocked
	IconBlocked TaskIcon = "[BLOCK]"
	// IconCycle represents a cyclical dependency
	IconCycle TaskIcon = "[CYCLE]"
	// IconUnknown represents an unknown task state
	IconUnknown TaskIcon = "[UNKNOWN]"
	// IconDependency represents a task dependency
	IconDependency TaskIcon = "->"
	// IconBlocks represents a blocking relationship
	IconBlocks TaskIcon = "=>"
	// IconFolder represents a folder/group
	IconFolder TaskIcon = "+"
	// IconFile represents an individual file/task
	IconFile TaskIcon = "-"
	// IconWarning represents a warning indicator
	IconWarning TaskIcon = "[!]"
)

// Mode represents different visualization types (renamed from VisualizationMode to avoid stuttering)
type Mode string

const (
	// ModeTask represents task-focused visualization
	ModeTask Mode = "task"
	// ModeProject represents project-focused visualization
	ModeProject Mode = "project"
	// ModeTree represents tree-style visualization
	ModeTree Mode = "tree"
	// ModeGraph represents graph-style visualization
	ModeGraph Mode = "graph"
	// ModeBlocks represents block-based visualization
	ModeBlocks Mode = "blocks"
)

// TaskRelationship represents the relationship between tasks
type TaskRelationship struct {
	FromTask   *types.Task
	ToTask     *types.Task
	Type       RelationshipType
	IsCircular bool
}

// RelationshipType defines the type of relationship
type RelationshipType string

const (
	// RelationshipDependency represents a dependency relationship
	RelationshipDependency RelationshipType = "dependency"
	// RelationshipBlocks represents a blocking relationship
	RelationshipBlocks RelationshipType = "blocks"
)

// Config holds configuration for visualization (renamed from VisualizationConfig to avoid stuttering)
type Config struct {
	Mode       Mode
	TaskID     string
	MaxDepth   int
	ShowBlocks bool
	JSONOutput bool
	ProjectID  string
}

// TaskAnalysisResult contains analysis results for a task
type TaskAnalysisResult struct {
	Task            *types.Task
	UpstreamTasks   []*types.Task
	DownstreamTasks []*types.Task
	IsBlocked       bool
	BlockingTasks   []*types.Task
	InCycle         bool
	Dependencies    []TaskRelationship
	Dependents      []TaskRelationship
}

// ProjectAnalysisResult contains project-wide analysis results
type ProjectAnalysisResult struct {
	TotalTasks       int
	TasksWithDeps    int
	BlockedTasks     int
	CompletedTasks   int
	InProgressTasks  int
	PendingTasks     int
	Cycles           [][]string // Task titles for display
	RootTasks        []*types.Task
	AllRelationships []TaskRelationship
	AllTasks         []*types.Task // Add this for renderer access
}
