package visualization

import (
	"github.com/denkhaus/knot/internal/types"
)

// TaskIcon represents visual indicators for tasks
type TaskIcon string

const (
	IconCompleted    TaskIcon = "[DONE]"
	IconInProgress   TaskIcon = "[WORK]"
	IconReady        TaskIcon = "[READY]"
	IconBlocked      TaskIcon = "[BLOCK]"
	IconCycle        TaskIcon = "[CYCLE]"
	IconUnknown      TaskIcon = "[UNKNOWN]"
	IconDependency   TaskIcon = "->"
	IconBlocks       TaskIcon = "=>"
	IconFolder       TaskIcon = "+"
	IconFile         TaskIcon = "-"
	IconWarning      TaskIcon = "[!]"
)

// VisualizationMode represents different visualization types
type VisualizationMode string

const (
	ModeTask      VisualizationMode = "task"
	ModeProject   VisualizationMode = "project"
	ModeTree      VisualizationMode = "tree"
	ModeGraph     VisualizationMode = "graph"
	ModeBlocks    VisualizationMode = "blocks"
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
	RelationshipDependency RelationshipType = "dependency"
	RelationshipBlocks    RelationshipType = "blocks"
)

// VisualizationConfig holds configuration for visualization
type VisualizationConfig struct {
	Mode          VisualizationMode
	TaskID        string
	MaxDepth      int
	ShowBlocks    bool
	JSONOutput    bool
	ProjectID     string
}

// TaskAnalysisResult contains analysis results for a task
type TaskAnalysisResult struct {
	Task               *types.Task
	UpstreamTasks      []*types.Task
	DownstreamTasks    []*types.Task
	IsBlocked          bool
	BlockingTasks      []*types.Task
	InCycle            bool
	Dependencies       []TaskRelationship
	Dependents         []TaskRelationship
}

// ProjectAnalysisResult contains project-wide analysis results
type ProjectAnalysisResult struct {
	TotalTasks         int
	TasksWithDeps       int
	BlockedTasks       int
	CompletedTasks      int
	InProgressTasks    int
	PendingTasks       int
	Cycles             [][]string // Task titles for display
	RootTasks          []*types.Task
	AllRelationships   []TaskRelationship
	AllTasks           []*types.Task // Add this for renderer access
}