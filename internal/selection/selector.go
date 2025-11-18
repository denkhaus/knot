package selection

import (
	"fmt"
	"sort"
	"time"

	"github.com/denkhaus/knot/internal/types"
)

// TaskSelector is the main coordinator for task selection
type DefaultTaskSelector struct {
	analyzer   DependencyAnalyzer
	filter     TaskFilter
	strategy   ScoringStrategy
	config     *Config
	lastResult *SelectionResult
}

// NewTaskSelector creates a new task selector with the specified strategy
func NewTaskSelector(strategy Strategy, config *Config) (*DefaultTaskSelector, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Validate configuration
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create components
	analyzer := NewDependencyAnalyzer(config)
	filter := NewTaskFilter(analyzer, config)

	// Create scoring strategy
	factory := &StrategyFactory{}
	scoringStrategy, err := factory.NewStrategy(strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to create strategy: %w", err)
	}

	return &DefaultTaskSelector{
		analyzer: analyzer,
		filter:   filter,
		strategy: scoringStrategy,
		config:   config,
	}, nil
}

// SelectNextActionableTask implements the main selection logic
func (ts *DefaultTaskSelector) SelectNextActionableTask(tasks []*types.Task) (*types.Task, error) {
	startTime := time.Now()

	// Validate input
	if len(tasks) == 0 {
		return nil, &SelectionError{
			Type:    ErrorTypeNoTasks,
			Message: "no tasks available",
		}
	}

	// Build dependency graph
	graph, err := ts.analyzer.BuildDependencyGraph(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// Handle cycles if they exist
	if graph.HasCycles {
		return nil, &SelectionError{
			Type:    ErrorTypeCircularDep,
			Message: fmt.Sprintf("circular dependencies detected in tasks: %v", graph.CyclicTasks),
		}
	}

	// Filter actionable tasks
	actionableTasks, err := ts.filter.FilterActionableTasks(tasks)
	if err != nil {
		return nil, err
	}

	if len(actionableTasks) == 0 {
		return nil, &SelectionError{
			Type:    ErrorTypeNoActionable,
			Message: "no actionable tasks found",
		}
	}

	// Score actionable tasks
	scoredTasks, err := ts.scoreActionableTasks(actionableTasks, graph)
	if err != nil {
		return nil, fmt.Errorf("failed to score tasks: %w", err)
	}

	// Select best task
	selectedScore, alternatives := ts.selectBestTask(scoredTasks)
	if selectedScore == nil {
		return nil, &SelectionError{
			Type:    ErrorTypeNoActionable,
			Message: "no suitable task found after scoring",
		}
	}

	// Generate selection reason
	reason := ts.generateSelectionReason(selectedScore, graph)
	selectedScore.SelectionReason = reason

	// Store result
	ts.lastResult = &SelectionResult{
		SelectedTask:  selectedScore.Task,
		Score:         selectedScore,
		Strategy:      ts.config.Strategy,
		Reason:        reason,
		Alternatives:  alternatives,
		SelectedAt:    time.Now(),
		ExecutionTime: time.Since(startTime),
	}

	return selectedScore.Task, nil
}

// scoreActionableTasks calculates selection scores for all actionable tasks
func (ts *DefaultTaskSelector) scoreActionableTasks(actionableTasks []*types.Task, graph *DependencyGraph) ([]*TaskScore, error) {
	scores := make([]*TaskScore, 0, len(actionableTasks))

	for _, task := range actionableTasks {
		score, err := ts.analyzer.CalculateTaskScore(task, graph)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate score for task %s: %w", task.ID, err)
		}

		// Calculate final score using strategy
		score.Score = ts.strategy.CalculateScore(score, ts.config)

		// Apply score threshold if configured
		if ts.config.Advanced.ScoreThreshold > 0 && score.Score < ts.config.Advanced.ScoreThreshold {
			continue // Skip tasks below threshold
		}

		scores = append(scores, score)
	}

	return scores, nil
}

// selectBestTask chooses the highest-scored task with proper tie-breaking
func (ts *DefaultTaskSelector) selectBestTask(scores []*TaskScore) (*TaskScore, []*TaskScore) {
	if len(scores) == 0 {
		return nil, nil
	}

	// Separate in-progress tasks if configured to prefer them
	inProgressScores := make([]*TaskScore, 0)
	pendingScores := make([]*TaskScore, 0)

	for _, score := range scores {
		if score.Task.State == types.TaskStateInProgress {
			inProgressScores = append(inProgressScores, score)
		} else {
			pendingScores = append(pendingScores, score)
		}
	}

	// Choose which group to prioritize
	var candidateScores []*TaskScore
	var alternatives []*TaskScore

	if ts.config.Behavior.PreferInProgress && len(inProgressScores) > 0 {
		candidateScores = inProgressScores
		alternatives = pendingScores
	} else if len(pendingScores) > 0 {
		candidateScores = pendingScores
		alternatives = inProgressScores
	} else {
		candidateScores = inProgressScores
		alternatives = []*TaskScore{}
	}

	// Sort candidates by score (descending)
	ts.sortTaskScores(candidateScores)

	// Sort alternatives for reference
	if len(alternatives) > 0 {
		ts.sortTaskScores(alternatives)
	}

	// Combine alternatives (other candidates + alternatives group)
	allAlternatives := make([]*TaskScore, 0, len(candidateScores)-1+len(alternatives))
	if len(candidateScores) > 1 {
		allAlternatives = append(allAlternatives, candidateScores[1:]...)
	}
	allAlternatives = append(allAlternatives, alternatives...)

	return candidateScores[0], allAlternatives
}

// sortTaskScores sorts task scores by score and applies tie-breaking rules
func (ts *DefaultTaskSelector) sortTaskScores(scores []*TaskScore) {
	sort.Slice(scores, func(i, j int) bool {
		scoreI := scores[i].Score
		scoreJ := scores[j].Score

		// Primary sort by score (descending)
		if scoreI != scoreJ {
			return scoreI > scoreJ
		}

		// Tie-breaking
		if ts.config.Behavior.BreakTiesByCreation {
			// Secondary sort by creation time (ascending - older first)
			return scores[i].Task.CreatedAt.Before(scores[j].Task.CreatedAt)
		}

		// Final tie-breaker by task ID for consistency
		return scores[i].Task.ID.String() < scores[j].Task.ID.String()
	})
}

// generateSelectionReason creates a human-readable explanation of why a task was selected
func (ts *DefaultTaskSelector) generateSelectionReason(score *TaskScore, graph *DependencyGraph) string {
	reasons := make([]string, 0)

	// Strategy-specific reasons
	strategyName := ts.strategy.GetStrategyName()
	reasons = append(reasons, fmt.Sprintf("selected using %s strategy", strategyName))

	// Specific factors
	if score.UnblockedTaskCount > 0 {
		reasons = append(reasons, fmt.Sprintf("will unblock %d task(s)", score.UnblockedTaskCount))
	}

	if score.Priority <= 2 {
		reasons = append(reasons, "high priority")
	}

	if score.DependentCount > 0 {
		reasons = append(reasons, fmt.Sprintf("%d task(s) depend on this", score.DependentCount))
	}

	if score.Task.State == types.TaskStateInProgress {
		reasons = append(reasons, "already in progress")
	}

	if score.HierarchyDepth > 0 {
		reasons = append(reasons, "subtask (completing branch)")
	}

	// Critical path information
	node := graph.Nodes[score.Task.ID]
	if node != nil && score.CriticalPathLength > 1 {
		reasons = append(reasons, fmt.Sprintf("on critical path (length %d)", score.CriticalPathLength))
	}

	if len(reasons) == 1 {
		// Only strategy reason, add score information
		reasons = append(reasons, fmt.Sprintf("score: %.2f", score.Score))
	}

	return reasons[0]
}

// GetSelectionReason returns the reason for the last selection
func (ts *DefaultTaskSelector) GetSelectionReason() string {
	if ts.lastResult == nil {
		return "no selection has been made"
	}
	return ts.lastResult.Reason
}

// GetLastResult returns the complete result of the last selection
func (ts *DefaultTaskSelector) GetLastResult() *SelectionResult {
	return ts.lastResult
}

// UpdateConfig updates the selector's configuration
func (ts *DefaultTaskSelector) UpdateConfig(config *Config) error {
	if err := ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	ts.config = config

	// Update strategy if it changed
	if ts.config.Strategy.String() != ts.strategy.GetStrategyName() {
		factory := &StrategyFactory{}
		newStrategy, err := factory.NewStrategy(ts.config.Strategy)
		if err != nil {
			return fmt.Errorf("failed to update strategy: %w", err)
		}
		ts.strategy = newStrategy
	}

	// Update analyzer and filter with new config
	ts.analyzer = NewDependencyAnalyzer(config)
	ts.filter = NewTaskFilter(ts.analyzer, config)

	return nil
}

// ValidateConfig validates a configuration
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate weights for strategies that use them
	factory := &StrategyFactory{}
	if err := factory.ValidateWeights(config.Strategy, config.Weights); err != nil {
		return fmt.Errorf("invalid weights: %w", err)
	}

	// Validate advanced settings
	if config.Advanced.MaxDependencyDepth < 0 {
		return fmt.Errorf("max dependency depth cannot be negative")
	}

	if config.Advanced.ScoreThreshold < 0 {
		return fmt.Errorf("score threshold cannot be negative")
	}

	if config.Advanced.CacheDuration < 0 {
		return fmt.Errorf("cache duration cannot be negative")
	}

	return nil
}
