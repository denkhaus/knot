package inmemory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
)

// simpleMemoryRepository implements Repository interface with simple in-memory storage
type simpleMemoryRepository struct {
	mu                sync.RWMutex
	projects          map[uuid.UUID]*types.Project
	tasks             map[uuid.UUID]*types.Task
	tasksByProject    map[uuid.UUID][]uuid.UUID
	tasksByParent     map[uuid.UUID][]uuid.UUID
	taskDependencies  map[uuid.UUID][]uuid.UUID // taskID -> list of dependency taskIDs
}

// NewMemoryRepository creates a new in-memory repository
func NewMemoryRepository() types.Repository {
	return &simpleMemoryRepository{
		projects:         make(map[uuid.UUID]*types.Project),
		tasks:            make(map[uuid.UUID]*types.Task),
		tasksByProject:   make(map[uuid.UUID][]uuid.UUID),
		tasksByParent:    make(map[uuid.UUID][]uuid.UUID),
		taskDependencies: make(map[uuid.UUID][]uuid.UUID),
	}
}

// Project operations
func (r *simpleMemoryRepository) CreateProject(ctx context.Context, project *types.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if project.ID == uuid.Nil {
		project.ID = uuid.New()
	}
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	r.projects[project.ID] = project
	return nil
}

func (r *simpleMemoryRepository) GetProject(ctx context.Context, id uuid.UUID) (*types.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	project, exists := r.projects[id]
	if !exists {
		return nil, fmt.Errorf("project not found")
	}
	return project, nil
}

func (r *simpleMemoryRepository) UpdateProject(ctx context.Context, project *types.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.projects[project.ID]; !exists {
		return fmt.Errorf("project not found")
	}

	project.UpdatedAt = time.Now()
	r.projects[project.ID] = project
	return nil
}

func (r *simpleMemoryRepository) DeleteProject(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.projects[id]; !exists {
		return fmt.Errorf("project not found")
	}

	delete(r.projects, id)
	delete(r.tasksByProject, id)
	return nil
}

func (r *simpleMemoryRepository) ListProjects(ctx context.Context) ([]*types.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	projects := make([]*types.Project, 0, len(r.projects))
	for _, project := range r.projects {
		projects = append(projects, project)
	}
	return projects, nil
}

// Task operations
func (r *simpleMemoryRepository) CreateTask(ctx context.Context, task *types.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	r.tasks[task.ID] = task
	
	// Add to project tasks
	r.tasksByProject[task.ProjectID] = append(r.tasksByProject[task.ProjectID], task.ID)
	
	// Add to parent tasks if applicable
	if task.ParentID != nil {
		r.tasksByParent[*task.ParentID] = append(r.tasksByParent[*task.ParentID], task.ID)
	}

	return nil
}

func (r *simpleMemoryRepository) GetTask(ctx context.Context, id uuid.UUID) (*types.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, exists := r.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task not found")
	}
	return task, nil
}

func (r *simpleMemoryRepository) UpdateTask(ctx context.Context, task *types.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[task.ID]; !exists {
		return fmt.Errorf("task not found")
	}

	task.UpdatedAt = time.Now()
	r.tasks[task.ID] = task
	return nil
}

func (r *simpleMemoryRepository) DeleteTask(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, exists := r.tasks[id]
	if !exists {
		return fmt.Errorf("task not found")
	}

	delete(r.tasks, id)
	
	// Remove from project tasks
	if projectTasks, exists := r.tasksByProject[task.ProjectID]; exists {
		for i, taskID := range projectTasks {
			if taskID == id {
				r.tasksByProject[task.ProjectID] = append(projectTasks[:i], projectTasks[i+1:]...)
				break
			}
		}
	}

	// Remove from parent tasks if applicable
	if task.ParentID != nil {
		if parentTasks, exists := r.tasksByParent[*task.ParentID]; exists {
			for i, taskID := range parentTasks {
				if taskID == id {
					r.tasksByParent[*task.ParentID] = append(parentTasks[:i], parentTasks[i+1:]...)
					break
				}
			}
		}
	}

	return nil
}

// Task queries
func (r *simpleMemoryRepository) ListTasks(ctx context.Context, filter types.TaskFilter) ([]*types.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tasks []*types.Task
	for _, task := range r.tasks {
		if r.matchesFilter(task, filter) {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (r *simpleMemoryRepository) GetTasksByProject(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	taskIDs, exists := r.tasksByProject[projectID]
	if !exists {
		return []*types.Task{}, nil
	}

	tasks := make([]*types.Task, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		if task, exists := r.tasks[taskID]; exists {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (r *simpleMemoryRepository) GetTasksByParent(ctx context.Context, parentID uuid.UUID) ([]*types.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	taskIDs, exists := r.tasksByParent[parentID]
	if !exists {
		return []*types.Task{}, nil
	}

	tasks := make([]*types.Task, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		if task, exists := r.tasks[taskID]; exists {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (r *simpleMemoryRepository) GetRootTasks(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error) {
	tasks, err := r.GetTasksByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var rootTasks []*types.Task
	for _, task := range tasks {
		if task.ParentID == nil {
			rootTasks = append(rootTasks, task)
		}
	}
	return rootTasks, nil
}

func (r *simpleMemoryRepository) GetParentTask(ctx context.Context, taskID uuid.UUID) (*types.Task, error) {
	task, err := r.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if task.ParentID == nil {
		return nil, fmt.Errorf("task has no parent")
	}

	return r.GetTask(ctx, *task.ParentID)
}

// Hierarchy operations
func (r *simpleMemoryRepository) DeleteTaskSubtree(ctx context.Context, taskID uuid.UUID) error {
	// Get all child tasks recursively
	childTasks, err := r.GetTasksByParent(ctx, taskID)
	if err != nil {
		return err
	}

	// Delete all children first
	for _, child := range childTasks {
		if err := r.DeleteTaskSubtree(ctx, child.ID); err != nil {
			return err
		}
	}

	// Delete the task itself
	return r.DeleteTask(ctx, taskID)
}

// Dependency management
func (r *simpleMemoryRepository) AddTaskDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID) (*types.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Verify both tasks exist
	task, exists := r.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found")
	}
	
	if _, exists := r.tasks[dependsOnTaskID]; !exists {
		return nil, fmt.Errorf("dependency task not found")
	}

	// Add dependency
	deps := r.taskDependencies[taskID]
	for _, dep := range deps {
		if dep == dependsOnTaskID {
			return task, nil // Already exists
		}
	}
	
	r.taskDependencies[taskID] = append(deps, dependsOnTaskID)
	
	// Update task dependencies slice
	task.Dependencies = r.taskDependencies[taskID]
	r.tasks[taskID] = task
	
	return task, nil
}

func (r *simpleMemoryRepository) RemoveTaskDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID) (*types.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, exists := r.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found")
	}

	deps := r.taskDependencies[taskID]
	for i, dep := range deps {
		if dep == dependsOnTaskID {
			r.taskDependencies[taskID] = append(deps[:i], deps[i+1:]...)
			break
		}
	}
	
	// Update task dependencies slice
	task.Dependencies = r.taskDependencies[taskID]
	r.tasks[taskID] = task
	
	return task, nil
}

func (r *simpleMemoryRepository) GetTaskDependencies(ctx context.Context, taskID uuid.UUID) ([]*types.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	deps := r.taskDependencies[taskID]
	tasks := make([]*types.Task, 0, len(deps))
	
	for _, depID := range deps {
		if task, exists := r.tasks[depID]; exists {
			tasks = append(tasks, task)
		}
	}
	
	return tasks, nil
}

func (r *simpleMemoryRepository) GetDependentTasks(ctx context.Context, taskID uuid.UUID) ([]*types.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var dependents []*types.Task
	
	for otherTaskID, deps := range r.taskDependencies {
		for _, depID := range deps {
			if depID == taskID {
				if task, exists := r.tasks[otherTaskID]; exists {
					dependents = append(dependents, task)
				}
				break
			}
		}
	}
	
	return dependents, nil
}

// Metrics and analysis
func (r *simpleMemoryRepository) GetProjectProgress(ctx context.Context, projectID uuid.UUID) (*types.ProjectProgress, error) {
	tasks, err := r.GetTasksByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	progress := &types.ProjectProgress{
		ProjectID:    projectID,
		TotalTasks:   len(tasks),
		TasksByDepth: make(map[int]int),
	}

	for _, task := range tasks {
		progress.TasksByDepth[task.Depth]++
		
		switch task.State {
		case types.TaskStateCompleted:
			progress.CompletedTasks++
		case types.TaskStateInProgress:
			progress.InProgressTasks++
		case types.TaskStatePending:
			progress.PendingTasks++
		case types.TaskStateBlocked:
			progress.BlockedTasks++
		case types.TaskStateCancelled:
			progress.CancelledTasks++
		}
	}

	if progress.TotalTasks > 0 {
		progress.OverallProgress = float64(progress.CompletedTasks) / float64(progress.TotalTasks) * 100
	}

	return progress, nil
}

func (r *simpleMemoryRepository) GetTaskCountByDepth(ctx context.Context, projectID uuid.UUID, maxDepth int) (map[int]int, error) {
	tasks, err := r.GetTasksByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	counts := make(map[int]int)
	for _, task := range tasks {
		if task.Depth <= maxDepth {
			counts[task.Depth]++
		}
	}

	return counts, nil
}

// Helper function to match tasks against filter
func (r *simpleMemoryRepository) matchesFilter(task *types.Task, filter types.TaskFilter) bool {
	if filter.ProjectID != nil && task.ProjectID != *filter.ProjectID {
		return false
	}
	if filter.ParentID != nil {
		if task.ParentID == nil || *task.ParentID != *filter.ParentID {
			return false
		}
	}
	if filter.State != nil && task.State != *filter.State {
		return false
	}
	if filter.MinDepth != nil && task.Depth < *filter.MinDepth {
		return false
	}
	if filter.MaxDepth != nil && task.Depth > *filter.MaxDepth {
		return false
	}
	if filter.MinComplexity != nil && task.Complexity < *filter.MinComplexity {
		return false
	}
	if filter.MaxComplexity != nil && task.Complexity > *filter.MaxComplexity {
		return false
	}
	return true
}