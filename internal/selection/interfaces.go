package selection

import "github.com/denkhaus/knot/v2/internal/types"

// TaskSelector defines the interface for task selection strategies
type TaskSelector interface {
	SelectNextActionableTask(tasks []*types.Task) (*types.Task, error)
	GetSelectionReason() string
}

// DependencyAnalyzer provides dependency graph analysis capabilities
type DependencyAnalyzer interface {
	BuildDependencyGraph(tasks []*types.Task) (*DependencyGraph, error)
	CalculateTaskScore(task *types.Task, graph *DependencyGraph) (*TaskScore, error)
	ValidateActionability(task *types.Task, allTasks []*types.Task) bool
}

// ScoringStrategy defines how tasks should be scored for selection
type ScoringStrategy interface {
	CalculateScore(score *TaskScore, config *Config) float64
	GetStrategyName() string
}

// TaskFilter provides task filtering capabilities
type TaskFilter interface {
	FilterActionableTasks(tasks []*types.Task) ([]*types.Task, error)
	IsTaskActionable(task *types.Task, allTasks []*types.Task) bool
}

// ConfigProvider provides configuration for task selection
type ConfigProvider interface {
	GetConfig() *Config
	SetConfig(config *Config)
	ValidateConfig() error
}
