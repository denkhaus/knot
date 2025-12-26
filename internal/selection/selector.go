package selection

import (
	"fmt"
	"sort"
	"time"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
)

// TaskSelector is the main coordinator for task selection
type DefaultTaskSelector struct {
	analyzer   DependencyAnalyzer
	filter     TaskFilter
	config     *Config
	lastResult *SelectionResult
}

// NewTaskSelector creates a new task selector with dependency-aware strategy
func NewTaskSelector(config *Config) (*DefaultTaskSelector, error) {
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

	return &DefaultTaskSelector{
		analyzer: analyzer,
		filter:   filter,
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

	// Handle cycles if they exist with enhanced error
	if graph.HasCycles {
		taskTitles := make(map[uuid.UUID]string)
		for _, taskID := range graph.CyclicTasks {
			if node, exists := graph.Nodes[taskID]; exists {
				taskTitles[taskID] = node.Task.Title
			}
		}

		enhancedErr := NewCircularDependencyError(graph.CyclicTasks, taskTitles)
		return nil, &SelectionError{
			Type:    ErrorTypeCircularDep,
			Message: enhancedErr.Error(),
		}
	}

	// Filter actionable tasks
	actionableTasks, err := ts.filter.FilterActionableTasks(tasks)
	if err != nil {
		return nil, err
	}

	if len(actionableTasks) == 0 {
		// Use enhanced error with suggestions
		enhancedErr := NewNoActionableTasksError(uuid.UUID{}, len(tasks))
		return nil, &SelectionError{
			Type:    ErrorTypeNoActionable,
			Message: enhancedErr.Error(),
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

		// Calculate final score using dependency-aware logic
		score.Score = ts.calculateScore(score)

		// Apply score threshold if configured
		if ts.config.ScoreThreshold > 0 && score.Score < ts.config.ScoreThreshold {
			continue // Skip tasks below threshold
		}

		scores = append(scores, score)
	}

	return scores, nil
}

// calculateScore implements dependency-aware scoring logic
func (ts *DefaultTaskSelector) calculateScore(score *TaskScore) float64 {
	// Weighted combination of factors
	dependentScore := float64(score.UnblockedTaskCount) * ts.config.DependentCountWeight
	priorityScore := ts.priorityToScore(score.Priority) * ts.config.PriorityWeight
	depthScore := float64(score.HierarchyDepth+1) * ts.config.DepthFirstWeight // Prefer deeper tasks for completing branches
	criticalScore := float64(score.CriticalPathLength) * ts.config.CriticalPathWeight

	totalScore := dependentScore + priorityScore + depthScore + criticalScore

	// Apply bonus for in-progress tasks
	if ts.config.PreferInProgress && score.Task.State == types.TaskStateInProgress {
		totalScore *= 1.2 // 20% bonus
	}

	return totalScore
}

// priorityToScore converts priority to scoring value
// Lower priority number = higher score (1=high priority gets high score)
func (ts *DefaultTaskSelector) priorityToScore(priority types.TaskPriority) float64 {
	return float64(4 - priority) // 1->3, 2->2, 3->1
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

	switch {
	case ts.config.PreferInProgress && len(inProgressScores) > 0:
		candidateScores = inProgressScores
		alternatives = pendingScores
	case len(pendingScores) > 0:
		candidateScores = pendingScores
		alternatives = inProgressScores
	default:
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

		// Tie-breaking by creation time (ascending - older first)
		return scores[i].Task.CreatedAt.Before(scores[j].Task.CreatedAt)
	})
}

// generateSelectionReason creates a human-readable explanation of why a task was selected
func (ts *DefaultTaskSelector) generateSelectionReason(score *TaskScore, graph *DependencyGraph) string {
	reasons := make([]string, 0)

	// Strategy-specific reasons
	reasons = append(reasons, "selected using dependency-aware strategy")

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

// ValidateConfig validates a configuration
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate weights sum to approximately 1.0
	total := config.DependentCountWeight + config.PriorityWeight + config.DepthFirstWeight + config.CriticalPathWeight
	if total < 0.9 || total > 1.1 { // Allow 10% tolerance
		return fmt.Errorf("weights should sum to approximately 1.0, got %.2f", total)
	}

	// Ensure all weights are non-negative
	if config.DependentCountWeight < 0 || config.PriorityWeight < 0 || config.DepthFirstWeight < 0 || config.CriticalPathWeight < 0 {
		return fmt.Errorf("all weights must be non-negative")
	}

	// Validate advanced settings
	if config.MaxDependencyDepth < 0 {
		return fmt.Errorf("max dependency depth cannot be negative")
	}

	if config.ScoreThreshold < 0 {
		return fmt.Errorf("score threshold cannot be negative")
	}

	if config.CacheDuration < 0 {
		return fmt.Errorf("cache duration cannot be negative")
	}

	return nil
}
