package selection

import (
	"fmt"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
)

// DefaultDependencyAnalyzer implements DependencyAnalyzer interface using modular components
type DefaultDependencyAnalyzer struct {
	config                *Config
	graphBuilder          *DependencyGraphBuilder
	cycleDetector         *CycleDetector
	metricsCalculator     *MetricsCalculator
	actionabilityValidator *ActionabilityValidator
	cache                 *Cache
}

// NewDependencyAnalyzer creates a new dependency analyzer with modular components
func NewDependencyAnalyzer(config *Config) *DefaultDependencyAnalyzer {
	return &DefaultDependencyAnalyzer{
		config:                config,
		graphBuilder:          NewDependencyGraphBuilder(config),
		cycleDetector:         NewCycleDetector(),
		metricsCalculator:     NewMetricsCalculator(config),
		actionabilityValidator: NewActionabilityValidator(config),
		cache:                 NewCache(config),
	}
}

// BuildDependencyGraph creates a comprehensive dependency analysis using modular components
func (da *DefaultDependencyAnalyzer) BuildDependencyGraph(tasks []*types.Task) (*DependencyGraph, error) {
	// Try to get from cache first
	cacheKey := CacheKey{
		TaskHash:   da.computeTaskHash(tasks),
		Strategy:   da.config.Strategy,
		ConfigHash: da.computeConfigHash(),
	}

	if graph, found := da.cache.Get(cacheKey); found {
		return graph, nil
	}

	// Build the dependency graph using the graph builder
	graph, err := da.graphBuilder.Build(tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// Detect cycles using the cycle detector
	da.cycleDetector.Detect(graph)

	// Calculate metrics using the metrics calculator
	if err := da.metricsCalculator.CalculateAll(graph); err != nil {
		return nil, fmt.Errorf("failed to calculate metrics: %w", err)
	}

	// Find critical path
	da.metricsCalculator.FindCriticalPath(graph)

	// Count actionable tasks
	da.actionabilityValidator.CountActionableTasks(graph, tasks)

	// Cache the result
	da.cache.Put(cacheKey, graph)

	return graph, nil
}

// computeTaskHash creates a simple hash for task change detection
func (da *DefaultDependencyAnalyzer) computeTaskHash(tasks []*types.Task) string {
	// Simple implementation - could be enhanced with proper hashing
	if len(tasks) == 0 {
		return "empty"
	}

	// Use first and last task IDs and count as a simple hash
	return fmt.Sprintf("%s-%d-%s", tasks[0].ID.String(), len(tasks), tasks[len(tasks)-1].ID.String())
}

// computeConfigHash creates a hash for configuration change detection
func (da *DefaultDependencyAnalyzer) computeConfigHash() string {
	return fmt.Sprintf("strategy-%d", da.config.Strategy)
}

// CalculateTaskScore computes a detailed score for a task using the metrics calculator
func (da *DefaultDependencyAnalyzer) CalculateTaskScore(task *types.Task, graph *DependencyGraph) (*TaskScore, error) {
	return da.metricsCalculator.CalculateTaskScore(task, graph)
}

// ValidateActionability checks if a task can be worked on right now using the actionability validator
func (da *DefaultDependencyAnalyzer) ValidateActionability(task *types.Task, allTasks []*types.Task) bool {
	return da.actionabilityValidator.ValidateActionability(task, allTasks)
}

// InvalidateCache clears the cache for a specific project
func (da *DefaultDependencyAnalyzer) InvalidateCache(projectID string) {
	// Convert string to UUID if needed, or implement project-based cache invalidation
	da.cache.Invalidate(da.parseProjectID(projectID))
}

// parseProjectID converts project ID string to UUID
func (da *DefaultDependencyAnalyzer) parseProjectID(projectID string) uuid.UUID {
	// Simple implementation - in a real scenario, this would parse the project ID properly
	// For now, return a zero UUID
	return uuid.UUID{}
}
