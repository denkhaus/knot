package selection

import (
	"fmt"
	"math"

	"github.com/denkhaus/knot/internal/types"
)

// priorityToScore converts priority to scoring value
// Lower priority number = higher score (1=high priority gets high score)
func priorityToScore(priority types.TaskPriority) float64 {
	return float64(4 - priority) // 1->3, 2->2, 3->1
}

// CreationOrderStrategy implements selection by creation time (original behavior)
type CreationOrderStrategy struct{}

// CalculateScore returns a score based on creation time (lower is better for older tasks)
func (s *CreationOrderStrategy) CalculateScore(score *TaskScore, config *Config) float64 {
	// Convert timestamp to score - older tasks get higher scores (inverted)
	timestamp := score.Task.CreatedAt.Unix()
	// Use max timestamp - current timestamp to invert (older = higher score)
	return float64(-timestamp) // Negative so older tasks have higher scores
}

// GetStrategyName returns the strategy name
func (s *CreationOrderStrategy) GetStrategyName() string {
	return "creation-order"
}

// DependencyAwareStrategy implements intelligent scoring based on multiple factors
type DependencyAwareStrategy struct{}

// CalculateScore computes a weighted score considering dependencies, priority, and hierarchy
func (s *DependencyAwareStrategy) CalculateScore(score *TaskScore, config *Config) float64 {
	// Weighted combination of factors
	dependentScore := float64(score.UnblockedTaskCount) * config.Weights.DependentCount
	priorityScore := priorityToScore(score.Priority) * config.Weights.Priority
	depthScore := float64(score.HierarchyDepth+1) * config.Weights.DepthFirst // Prefer deeper tasks for completing branches
	criticalScore := float64(score.CriticalPathLength) * config.Weights.CriticalPath

	totalScore := dependentScore + priorityScore + depthScore + criticalScore

	// Apply bonus for in-progress tasks if configured
	if config.Behavior.PreferInProgress && score.Task.State == types.TaskStateInProgress {
		totalScore *= 1.2 // 20% bonus
	}

	return totalScore
}

// GetStrategyName returns the strategy name
func (s *DependencyAwareStrategy) GetStrategyName() string {
	return "dependency-aware"
}

// PriorityStrategy implements selection primarily based on task priority
type PriorityStrategy struct{}

// CalculateScore returns a score based primarily on priority with dependents as tiebreaker
func (s *PriorityStrategy) CalculateScore(score *TaskScore, config *Config) float64 {
	// Priority is the main factor (lower number = higher priority = higher score)
	priorityScore := priorityToScore(score.Priority) * 100

	// Add dependents as secondary factor
	dependentScore := float64(score.DependentCount)

	return priorityScore + dependentScore
}

// GetStrategyName returns the strategy name
func (s *PriorityStrategy) GetStrategyName() string {
	return "priority"
}

// DepthFirstStrategy implements selection that prioritizes completing branches
type DepthFirstStrategy struct{}

// CalculateScore returns a score that heavily favors deeper tasks (subtasks over parents)
func (s *DepthFirstStrategy) CalculateScore(score *TaskScore, config *Config) float64 {
	// Heavily weight hierarchy depth to complete branches first
	depthScore := float64(score.HierarchyDepth * 1000)

	// Add priority as secondary factor
	priorityScore := priorityToScore(score.Priority)

	// Subtract dependent count to prefer leaf tasks over ones with many dependents
	dependentPenalty := float64(score.DependentCount * 10)

	return depthScore + priorityScore - dependentPenalty
}

// GetStrategyName returns the strategy name
func (s *DepthFirstStrategy) GetStrategyName() string {
	return "depth-first"
}

// CriticalPathStrategy implements selection based on critical path analysis
type CriticalPathStrategy struct{}

// CalculateScore returns a score that prioritizes tasks on the critical path
func (s *CriticalPathStrategy) CalculateScore(score *TaskScore, config *Config) float64 {
	// Primary factor: critical path length
	criticalScore := float64(score.CriticalPathLength * 100)

	// Secondary factor: number of tasks this would unblock
	unblockedScore := float64(score.UnblockedTaskCount * 50)

	// Tertiary factor: priority
	priorityScore := priorityToScore(score.Priority) * 10

	return criticalScore + unblockedScore + priorityScore
}

// GetStrategyName returns the strategy name
func (s *CriticalPathStrategy) GetStrategyName() string {
	return "critical-path"
}

// StrategyFactory creates scoring strategies
type StrategyFactory struct{}

// NewStrategy creates a new scoring strategy based on the given strategy type
func (f *StrategyFactory) NewStrategy(strategy Strategy) (ScoringStrategy, error) {
	switch strategy {
	case StrategyCreationOrder:
		return &CreationOrderStrategy{}, nil
	case StrategyDependencyAware:
		return &DependencyAwareStrategy{}, nil
	case StrategyPriority:
		return &PriorityStrategy{}, nil
	case StrategyDepthFirst:
		return &DepthFirstStrategy{}, nil
	case StrategyCriticalPath:
		return &CriticalPathStrategy{}, nil
	default:
		return nil, fmt.Errorf("unknown strategy: %v", strategy)
	}
}

// GetAvailableStrategies returns all available strategies
func (f *StrategyFactory) GetAvailableStrategies() []Strategy {
	return []Strategy{
		StrategyCreationOrder,
		StrategyDependencyAware,
		StrategyPriority,
		StrategyDepthFirst,
		StrategyCriticalPath,
	}
}

// ValidateWeights ensures that weights are valid for the given strategy
func (f *StrategyFactory) ValidateWeights(strategy Strategy, weights Weights) error {
	switch strategy {
	case StrategyDependencyAware:
		// For dependency-aware strategy, weights should sum to approximately 1.0
		total := weights.DependentCount + weights.Priority + weights.DepthFirst + weights.CriticalPath
		if math.Abs(total-1.0) > 0.1 { // Allow 10% tolerance
			return fmt.Errorf("weights should sum to 1.0, got %.2f", total)
		}

		// Ensure all weights are non-negative
		if weights.DependentCount < 0 || weights.Priority < 0 || weights.DepthFirst < 0 || weights.CriticalPath < 0 {
			return fmt.Errorf("all weights must be non-negative")
		}

	case StrategyCreationOrder, StrategyPriority, StrategyDepthFirst, StrategyCriticalPath:
		// Other strategies don't use weights, but we don't need to error
		// Just ignore the weights for these strategies

	default:
		return fmt.Errorf("unknown strategy for weight validation: %v", strategy)
	}

	return nil
}

// GetDefaultWeightsForStrategy returns sensible default weights for each strategy
func (f *StrategyFactory) GetDefaultWeightsForStrategy(strategy Strategy) Weights {
	switch strategy {
	case StrategyDependencyAware:
		return Weights{
			DependentCount: 0.4,
			Priority:       0.3,
			DepthFirst:     0.2,
			CriticalPath:   0.1,
		}
	case StrategyPriority:
		return Weights{
			Priority:       1.0,
			DependentCount: 0.0,
			DepthFirst:     0.0,
			CriticalPath:   0.0,
		}
	case StrategyDepthFirst:
		return Weights{
			DepthFirst:     1.0,
			Priority:       0.0,
			DependentCount: 0.0,
			CriticalPath:   0.0,
		}
	case StrategyCriticalPath:
		return Weights{
			CriticalPath:   0.6,
			DependentCount: 0.3,
			Priority:       0.1,
			DepthFirst:     0.0,
		}
	default: // Including StrategyCreationOrder
		return Weights{
			DependentCount: 0.0,
			Priority:       0.0,
			DepthFirst:     0.0,
			CriticalPath:   0.0,
		}
	}
}

// RecommendStrategy suggests the best strategy based on project characteristics
func (f *StrategyFactory) RecommendStrategy(taskCount, avgDependencies int, hasHierarchy bool) Strategy {
	// Simple heuristics for strategy recommendation
	if taskCount < 10 {
		return StrategyCreationOrder // Simple projects can use creation order
	}

	if avgDependencies > 2 {
		return StrategyDependencyAware // Complex dependencies benefit from dependency-aware
	}

	if hasHierarchy {
		return StrategyDepthFirst // Projects with subtasks benefit from depth-first
	}

	// Default to dependency-aware for most cases
	return StrategyDependencyAware
}
