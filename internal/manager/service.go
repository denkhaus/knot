package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	knoterrors "github.com/denkhaus/knot/internal/errors"
	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
)

// service provides business logic for project task management
type service struct {
	repo   types.Repository
	config *Config
}

// newService creates a new task management service
func newService(repo types.Repository, config *Config) *service {
	if config == nil {
		config = DefaultConfig()
	}
	return &service{
		repo:   repo,
		config: config,
	}
}

// Ensure service implements ProjectManager
var _ ProjectManager = (*service)(nil)

// Project operations

func (s *service) CreateProject(ctx context.Context, title, description, actor string) (*types.Project, error) {
	if err := s.validateProjectInput(title, description); err != nil {
		return nil, err
	}

	project := &types.Project{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		State:       types.ProjectStateActive, // Set initial state to active
		CreatedBy:   actor,
		UpdatedBy:   actor,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateProject(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return s.repo.GetProject(ctx, project.ID)
}

func (s *service) GetProject(ctx context.Context, projectID uuid.UUID) (*types.Project, error) {
	return s.repo.GetProject(ctx, projectID)
}

func (s *service) UpdateProject(ctx context.Context, projectID uuid.UUID, title, description string, actor string) (*types.Project, error) {
	if err := s.validateProjectInput(title, description); err != nil {
		return nil, err
	}

	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	project.Title = title
	project.Description = description
	project.UpdatedBy = actor

	if err := s.repo.UpdateProject(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return s.repo.GetProject(ctx, projectID)
}

func (s *service) UpdateProjectDescription(ctx context.Context, projectID uuid.UUID, description string, actor string) (*types.Project, error) {
	// Validate description length
	if len(description) > s.config.MaxDescriptionLength {
		return nil, fmt.Errorf("description cannot exceed %d characters", s.config.MaxDescriptionLength)
	}

	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	project.Description = description
	project.UpdatedBy = actor

	if err := s.repo.UpdateProject(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project description: %w", err)
	}

	return s.repo.GetProject(ctx, projectID)
}

func (s *service) UpdateProjectState(ctx context.Context, projectID uuid.UUID, state types.ProjectState, actor string) (*types.Project, error) {
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Validate state transition
	if !isValidProjectStateTransition(project.State, state) {
		return nil, fmt.Errorf("invalid project state transition from '%s' to '%s'", project.State, state)
	}

	project.State = state
	project.UpdatedBy = actor
	project.UpdatedAt = time.Now()

	if err := s.repo.UpdateProject(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project state: %w", err)
	}

	return s.repo.GetProject(ctx, projectID)
}

func (s *service) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	return s.repo.DeleteProject(ctx, projectID)
}

func (s *service) ListProjects(ctx context.Context) ([]*types.Project, error) {
	return s.repo.ListProjects(ctx)
}

// Task operations

func (s *service) CreateTask(ctx context.Context, projectID uuid.UUID, parentID *uuid.UUID, title, description string, complexity int, priority types.TaskPriority, actor string) (*types.Task, error) {
	if err := s.validateTaskInput(title, description, complexity); err != nil {
		return nil, err
	}

	// Validate project exists
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Calculate depth and validate constraints
	depth := 0
	if parentID != nil {
		parentTask, err := s.repo.GetTask(ctx, *parentID)
		if err != nil {
			return nil, fmt.Errorf("parent task not found: %w", err)
		}
		if parentTask.ProjectID != projectID {
			return nil, fmt.Errorf("parent task must be in the same project")
		}
		depth = parentTask.Depth + 1
	}

	// Check depth constraints
	if depth > s.config.MaxDepth {
		return nil, fmt.Errorf("maximum depth of %d exceeded", s.config.MaxDepth)
	}

	// Check task count constraints for this depth
	counts, err := s.repo.GetTaskCountByDepth(ctx, projectID, depth)
	if err != nil {
		return nil, fmt.Errorf("failed to check task count constraints: %w", err)
	}
	if counts[depth] >= s.config.MaxTasksPerDepth {
		return nil, knoterrors.TooManyTasksError(counts[depth], s.config.MaxTasksPerDepth, depth)
	}

	task := &types.Task{
		ID:          uuid.New(),
		ProjectID:   projectID,
		ParentID:    parentID,
		Title:       title,
		Description: description,
		State:       types.TaskStatePending,
		Priority:    priority,
		Complexity:  complexity,
		Depth:       depth,
		CreatedBy:   actor,
		UpdatedBy:   actor,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Auto-reduce parent complexity if enabled and this is a subtask
	if s.config.AutoReduceComplexity && parentID != nil {
		if err := s.autoReduceParentComplexity(ctx, *parentID); err != nil {
			// Log error but don't fail the task creation
			// The task was successfully created, complexity reduction is a bonus feature
			fmt.Printf("Warning: Failed to auto-reduce parent complexity: %v\n", err)
		}
	}

	return s.repo.GetTask(ctx, task.ID)
}

func (s *service) GetTask(ctx context.Context, taskID uuid.UUID) (*types.Task, error) {
	return s.repo.GetTask(ctx, taskID)
}

func (s *service) UpdateTaskState(ctx context.Context, taskID uuid.UUID, state types.TaskState, actor string) (*types.Task, error) {
	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Validate state transition
	if !isValidTaskStateTransition(task.State, state) {
		return nil, fmt.Errorf("invalid state transition from '%s' to '%s'", task.State, state)
	}

	task.State = state
	task.UpdatedBy = actor
	task.UpdatedAt = time.Now()

	// Set completion timestamp if transitioning to completed
	if state == types.TaskStateCompleted && task.CompletedAt == nil {
		now := time.Now()
		task.CompletedAt = &now
	} else if state != types.TaskStateCompleted && task.CompletedAt != nil {
		// Clear completion timestamp if moving away from completed
		task.CompletedAt = nil
	}

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task state: %w", err)
	}

	return s.repo.GetTask(ctx, taskID)
}

func (s *service) UpdateTask(ctx context.Context, taskID uuid.UUID, title, description string, complexity int, state types.TaskState, actor string) (*types.Task, error) {
	if err := s.validateTaskInput(title, description, complexity); err != nil {
		return nil, err
	}

	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	task.Title = title
	task.Description = description
	task.Complexity = complexity
	task.State = state
	task.UpdatedBy = actor
	task.UpdatedAt = time.Now()

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return s.repo.GetTask(ctx, taskID)
}

func (s *service) UpdateTaskDescription(ctx context.Context, taskID uuid.UUID, description string, actor string) (*types.Task, error) {
	// Validate description length
	if len(description) > s.config.MaxDescriptionLength {
		return nil, fmt.Errorf("description cannot exceed %d characters", s.config.MaxDescriptionLength)
	}

	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	task.Description = description
	task.UpdatedBy = actor

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task description: %w", err)
	}

	return s.repo.GetTask(ctx, taskID)
}

func (s *service) UpdateTaskTitle(ctx context.Context, taskID uuid.UUID, title string, actor string) (*types.Task, error) {
	// Validate title
	if title == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}
	if len(title) > 200 {
		return nil, fmt.Errorf("title cannot exceed 200 characters")
	}

	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	task.Title = title
	task.UpdatedBy = actor

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task title: %w", err)
	}

	return s.repo.GetTask(ctx, taskID)
}

func (s *service) UpdateTaskPriority(ctx context.Context, taskID uuid.UUID, priority types.TaskPriority, actor string) (*types.Task, error) {
	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	task.Priority = priority
	task.UpdatedBy = actor
	task.UpdatedAt = time.Now()

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task priority: %w", err)
	}

	return s.repo.GetTask(ctx, taskID)
}

func (s *service) DeleteTask(ctx context.Context, taskID uuid.UUID, actor string) error {
	return s.repo.DeleteTask(ctx, taskID)
}

func (s *service) DeleteTaskSubtree(ctx context.Context, taskID uuid.UUID, actor string) error {
	return s.repo.DeleteTaskSubtree(ctx, taskID)
}

// Task queries and analysis

func (s *service) GetParentTask(ctx context.Context, taskID uuid.UUID) (*types.Task, error) {
	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.ParentID == nil {
		return nil, nil // No parent
	}
	return s.repo.GetTask(ctx, *task.ParentID)
}

func (s *service) GetChildTasks(ctx context.Context, taskID uuid.UUID) ([]*types.Task, error) {
	// Validate task exists
	if _, err := s.repo.GetTask(ctx, taskID); err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}
	return s.repo.GetTasksByParent(ctx, taskID)
}

func (s *service) GetRootTasks(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error) {
	return s.repo.GetRootTasks(ctx, projectID)
}

// ListTasksForProject returns all tasks in a project regardless of hierarchy level
func (s *service) ListTasksForProject(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error) {
	// Validate project exists
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return s.repo.GetTasksByProject(ctx, projectID)
}

// BulkUpdateTasks updates multiple tasks with the same updates
func (s *service) BulkUpdateTasks(ctx context.Context, taskIDs []uuid.UUID, updates types.TaskUpdates) error {
	if len(taskIDs) == 0 {
		return nil // Nothing to update
	}

	// Validate updates
	if updates.State == nil && updates.Complexity == nil {
		return fmt.Errorf("at least one field must be specified for update")
	}

	// Validate complexity if provided
	if updates.Complexity != nil && (*updates.Complexity < 1 || *updates.Complexity > 10) {
		return fmt.Errorf("complexity is %d but must be between 1 and 10", *updates.Complexity)
	}

	// Update each task
	for _, taskID := range taskIDs {
		task, err := s.repo.GetTask(ctx, taskID)
		if err != nil {
			return fmt.Errorf("failed to get task %s: %w", taskID, err)
		}

		// Apply updates
		if updates.State != nil {
			task.State = *updates.State
			if task.State == types.TaskStateCompleted && task.CompletedAt == nil {
				now := time.Now()
				task.CompletedAt = &now
			} else if task.State != types.TaskStateCompleted && task.CompletedAt != nil {
				task.CompletedAt = nil
			}
		}
		if updates.Complexity != nil {
			task.Complexity = *updates.Complexity
		}

		task.UpdatedAt = time.Now()

		if err := s.repo.UpdateTask(ctx, task); err != nil {
			return fmt.Errorf("failed to update task %s: %w", taskID, err)
		}
	}

	return nil
}

// DuplicateTask creates a copy of a task in a new project
func (s *service) DuplicateTask(ctx context.Context, taskID uuid.UUID, newProjectID uuid.UUID) (*types.Task, error) {
	// Validate source task exists
	originalTask, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source task: %w", err)
	}

	// Validate target project exists
	if _, err := s.repo.GetProject(ctx, newProjectID); err != nil {
		return nil, fmt.Errorf("target project not found: %w", err)
	}

	// Create a copy of the task
	newTask := &types.Task{
		ID:          uuid.New(),
		ProjectID:   newProjectID,
		ParentID:    originalTask.ParentID, // This will be nil for the duplicated task
		Title:       originalTask.Title,
		Description: originalTask.Description,
		State:       types.TaskStatePending, // Reset state to pending
		Complexity:  originalTask.Complexity,
		Depth:       0, // Reset depth to 0 as it's now a root task
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CompletedAt: nil, // Reset completion status
	}

	// Save the new task
	if err := s.repo.CreateTask(ctx, newTask); err != nil {
		return nil, fmt.Errorf("failed to create duplicated task: %w", err)
	}

	return s.repo.GetTask(ctx, newTask.ID)
}

// SetTaskEstimate sets the time estimate for a task
func (s *service) SetTaskEstimate(ctx context.Context, taskID uuid.UUID, estimate int64) (*types.Task, error) {
	// Validate estimate
	if estimate < 0 {
		return nil, errors.New("estimate must be non-negative")
	}

	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	task.Estimate = &estimate
	task.UpdatedAt = time.Now()

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task estimate: %w", err)
	}

	return s.repo.GetTask(ctx, taskID)
}

// Agent assignment methods

// AssignTaskToAgent assigns a task to a specific agent
func (s *service) AssignTaskToAgent(ctx context.Context, taskID uuid.UUID, agentID uuid.UUID) (*types.Task, error) {
	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	task.AssignedAgent = &agentID
	task.UpdatedAt = time.Now()

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to assign task to agent: %w", err)
	}

	return s.repo.GetTask(ctx, taskID)
}

// UnassignTaskFromAgent removes agent assignment from a task
func (s *service) UnassignTaskFromAgent(ctx context.Context, taskID uuid.UUID) (*types.Task, error) {
	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	task.AssignedAgent = nil
	task.UpdatedAt = time.Now()

	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to unassign task from agent: %w", err)
	}

	return s.repo.GetTask(ctx, taskID)
}

// ListTasksByAgent returns all tasks assigned to a specific agent in a project
func (s *service) ListTasksByAgent(ctx context.Context, projectID uuid.UUID, agentID uuid.UUID) ([]*types.Task, error) {
	// Validate project exists
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Get all tasks in the project
	allTasks, err := s.repo.GetTasksByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project tasks: %w", err)
	}

	// Filter tasks assigned to the specific agent
	var assignedTasks []*types.Task
	for _, task := range allTasks {
		if task.AssignedAgent != nil && *task.AssignedAgent == agentID {
			assignedTasks = append(assignedTasks, task)
		}
	}

	return assignedTasks, nil
}

// ListUnassignedTasks returns all tasks that have no agent assigned in a project
func (s *service) ListUnassignedTasks(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error) {
	// Validate project exists
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Get all tasks in the project
	allTasks, err := s.repo.GetTasksByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project tasks: %w", err)
	}

	// Filter tasks with no agent assignment
	var unassignedTasks []*types.Task
	for _, task := range allTasks {
		if task.AssignedAgent == nil {
			unassignedTasks = append(unassignedTasks, task)
		}
	}

	return unassignedTasks, nil
}

func (s *service) FindNextActionableTask(ctx context.Context, projectID uuid.UUID) (*types.Task, error) {
	// Validate project exists
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Get all tasks in the project
	allTasks, err := s.repo.GetTasksByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project tasks: %w", err)
	}

	// Create a map of task IDs to tasks for quick lookup
	taskMap := make(map[uuid.UUID]*types.Task)
	for _, task := range allTasks {
		taskMap[task.ID] = task
	}

	// Separate tasks by state
	var pendingTasks, inProgressTasks []*types.Task
	for _, task := range allTasks {
		switch task.State {
		case types.TaskStatePending:
			pendingTasks = append(pendingTasks, task)
		case types.TaskStateInProgress:
			inProgressTasks = append(inProgressTasks, task)
		}
	}

	// Prioritize in-progress tasks first
	if len(inProgressTasks) > 0 {
		// For in-progress tasks, find one that has all its dependencies met
		for _, task := range inProgressTasks {
			if s.areDependenciesMet(task, taskMap) {
				return task, nil
			}
		}
		// If no in-progress task has its dependencies met, this indicates an inconsistency
		// Since we prevent circular dependencies and should maintain data integrity,
		// this should not happen. We'll return an error to highlight the issue.
		return nil, fmt.Errorf("in-progress tasks exist but none have all dependencies met - possible data inconsistency")
	}

	// For pending tasks, find one that has all its dependencies met
	for _, task := range pendingTasks {
		if s.areDependenciesMet(task, taskMap) {
			return task, nil
		}
	}

	// If we reach here, it means either:
	// 1. There are no pending or in-progress tasks
	// 2. All pending tasks have unmet dependencies (potential deadlock scenario)
	// Since we prevent circular dependencies, case 2 suggests a logical error in task setup
	if len(pendingTasks) > 0 {
		return nil, fmt.Errorf("pending tasks exist but none have all dependencies met - possible deadlock scenario")
	}

	// No actionable tasks found
	return nil, fmt.Errorf("no actionable tasks found")
}

// areDependenciesMet checks if all dependencies of a task are completed
func (s *service) areDependenciesMet(task *types.Task, taskMap map[uuid.UUID]*types.Task) bool {
	for _, depID := range task.Dependencies {
		depTask, exists := taskMap[depID]
		if !exists || depTask.State != types.TaskStateCompleted {
			// If a dependency doesn't exist or isn't completed, the dependencies aren't met
			return false
		}
	}
	return true
}

func (s *service) FindTasksNeedingBreakdown(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error) {
	// Validate project exists
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Find tasks with complexity above threshold that have no children
	tasks, err := s.repo.GetTasksByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project tasks: %w", err)
	}

	var needsBreakdown []*types.Task
	for _, task := range tasks {
		if task.Complexity >= s.config.ComplexityThreshold {
			// Check if task has children
			children, err := s.repo.GetTasksByParent(ctx, task.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to check task children: %w", err)
			}
			if len(children) == 0 {
				needsBreakdown = append(needsBreakdown, task)
			}
		}
	}

	return needsBreakdown, nil
}

func (s *service) GetProjectProgress(ctx context.Context, projectID uuid.UUID) (*types.ProjectProgress, error) {
	// Validate project exists
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return s.repo.GetProjectProgress(ctx, projectID)
}

func (s *service) ListTasksByState(ctx context.Context, projectID uuid.UUID, state types.TaskState) ([]*types.Task, error) {
	// Validate project exists
	if _, err := s.repo.GetProject(ctx, projectID); err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	return s.repo.ListTasks(ctx, types.TaskFilter{
		ProjectID: &projectID,
		State:     &state,
	})
}

// Validation helpers

func (s *service) validateProjectInput(title, description string) error {
	if title == "" {
		return errors.New("title cannot be empty")
	}
	if len(title) > 200 {
		return errors.New("title cannot exceed 200 characters")
	}
	if len(description) > s.config.MaxDescriptionLength {
		return fmt.Errorf("description cannot exceed %d characters", s.config.MaxDescriptionLength)
	}
	return nil
}

func (s *service) validateTaskInput(title, description string, complexity int) error {
	if title == "" {
		return errors.New("title cannot be empty")
	}
	if len(title) > 200 {
		return errors.New("title cannot exceed 200 characters")
	}
	if len(description) > s.config.MaxDescriptionLength {
		return fmt.Errorf("description cannot exceed %d characters", s.config.MaxDescriptionLength)
	}
	if complexity < 1 || complexity > 10 {
		return errors.New("complexity must be between 1 and 10")
	}
	return nil
}

// Config management

func (s *service) GetConfig() *Config {
	return s.config
}

func (s *service) UpdateConfig(config *Config) {
	if config != nil {
		s.config = config
	}
}

// LoadConfigFromFile loads configuration from .knot/config.json
func (s *service) LoadConfigFromFile() error {
	// Import here to avoid circular dependency
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// If config file doesn't exist, keep current config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // No error, just use current config
	}

	// Read and parse config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate and update config
	if err := validateConfig(&config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	s.config = &config
	return nil
}

// SaveConfigToFile saves current configuration to .knot/config.json
func (s *service) SaveConfigToFile() error {
	if err := validateConfig(s.config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Ensure .knot directory exists
	knotDir := filepath.Dir(configPath)
	if err := os.MkdirAll(knotDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddTaskDependency adds a dependency between tasks
func (s *service) AddTaskDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID, actor string) (*types.Task, error) {
	return s.repo.AddTaskDependency(ctx, taskID, dependsOnTaskID)
}

// RemoveTaskDependency removes a dependency between tasks
func (s *service) RemoveTaskDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID, actor string) (*types.Task, error) {
	return s.repo.RemoveTaskDependency(ctx, taskID, dependsOnTaskID)
}

// GetTaskDependencies gets all tasks that the given task depends on
func (s *service) GetTaskDependencies(ctx context.Context, taskID uuid.UUID) ([]*types.Task, error) {
	return s.repo.GetTaskDependencies(ctx, taskID)
}

// GetDependentTasks gets all tasks that depend on the given task
func (s *service) GetDependentTasks(ctx context.Context, taskID uuid.UUID) ([]*types.Task, error) {
	return s.repo.GetDependentTasks(ctx, taskID)
}

// Helper functions for config file management

// getConfigPath returns the path to the knot configuration file
func getConfigPath() (string, error) {
	// Use .knot directory for configuration (same as database)
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	knotDir := filepath.Join(cwd, ".knot")
	configPath := filepath.Join(knotDir, "config.json")
	return configPath, nil
}

// validateConfig checks if the configuration values are valid
func validateConfig(c *Config) error {
	if c.MaxTasksPerDepth < 1 {
		return fmt.Errorf("max_tasks_per_depth must be at least 1, got %d", c.MaxTasksPerDepth)
	}
	if c.ComplexityThreshold < 1 || c.ComplexityThreshold > 10 {
		return fmt.Errorf("complexity_threshold must be between 1 and 10, got %d", c.ComplexityThreshold)
	}
	if c.MaxDepth < 1 {
		return fmt.Errorf("max_depth must be at least 1, got %d", c.MaxDepth)
	}
	if c.MaxDescriptionLength < 1 {
		return fmt.Errorf("max_description_length must be at least 1, got %d", c.MaxDescriptionLength)
	}
	return nil
}

// autoReduceParentComplexity reduces parent task complexity when subtasks are added
func (s *service) autoReduceParentComplexity(ctx context.Context, parentID uuid.UUID) error {
	// Get parent task
	parentTask, err := s.repo.GetTask(ctx, parentID)
	if err != nil {
		return fmt.Errorf("failed to get parent task: %w", err)
	}

	// Only reduce if parent complexity is above threshold
	if parentTask.Complexity < s.config.ComplexityThreshold {
		return nil // No need to reduce
	}

	// Get all children to determine new complexity
	children, err := s.repo.GetTasksByParent(ctx, parentID)
	if err != nil {
		return fmt.Errorf("failed to get child tasks: %w", err)
	}

	// Calculate new complexity based on number of children
	// Logic: High complexity tasks become coordination tasks when broken down
	var newComplexity int
	childCount := len(children)

	switch {
	case childCount == 1:
		// First subtask: reduce by 2 (e.g., 9 -> 7)
		newComplexity = parentTask.Complexity - 2
	case childCount <= 3:
		// 2-3 subtasks: reduce to coordination level (complexity 4-5)
		newComplexity = 4
	case childCount <= 5:
		// 4-5 subtasks: well-broken down, reduce to oversight level (complexity 3)
		newComplexity = 3
	default:
		// Many subtasks: very well broken down, minimal coordination (complexity 2)
		newComplexity = 2
	}

	// Ensure complexity doesn't go below 1
	if newComplexity < 1 {
		newComplexity = 1
	}

	// Only update if complexity actually changed
	if newComplexity != parentTask.Complexity {
		parentTask.Complexity = newComplexity
		if err := s.repo.UpdateTask(ctx, parentTask); err != nil {
			return fmt.Errorf("failed to update parent task complexity: %w", err)
		}

		fmt.Printf("Auto-reduced parent task complexity: %s (ID: %s) %d -> %d (based on %d subtasks)\n",
			parentTask.Title, parentTask.ID, parentTask.Complexity+2, newComplexity, childCount)
	}

	return nil
}

// State validation functions

// isValidTaskStateTransition checks if a task state transition is valid
func isValidTaskStateTransition(from, to types.TaskState) bool {
	// Define valid transitions
	validTransitions := map[types.TaskState][]types.TaskState{
		types.TaskStatePending: {
			types.TaskStateInProgress,
			types.TaskStateBlocked,
			types.TaskStateCancelled,
			types.TaskStateDeletionPending,
		},
		types.TaskStateInProgress: {
			types.TaskStateCompleted,
			types.TaskStateBlocked,
			types.TaskStateCancelled,
			types.TaskStateDeletionPending,
		},
		types.TaskStateBlocked: {
			types.TaskStatePending,
			types.TaskStateInProgress,
			types.TaskStateCancelled,
			types.TaskStateDeletionPending,
		},
		types.TaskStateCompleted: {
			types.TaskStateDeletionPending,
			// Generally, completed tasks shouldn't transition back, but allow for corrections
		},
		types.TaskStateCancelled: {
			types.TaskStatePending,
			types.TaskStateDeletionPending,
		},
		types.TaskStateDeletionPending: {
			// No transitions allowed from deletion pending
		},
	}

	// Allow staying in the same state
	if from == to {
		return true
	}

	// Check if transition is in the valid list
	allowedStates, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowedState := range allowedStates {
		if to == allowedState {
			return true
		}
	}

	return false
}

// isValidProjectStateTransition checks if a project state transition is valid
func isValidProjectStateTransition(from, to types.ProjectState) bool {
	// Define valid transitions
	validTransitions := map[types.ProjectState][]types.ProjectState{
		types.ProjectStateActive: {
			types.ProjectStateCompleted,
			types.ProjectStateArchived,
			types.ProjectStateDeletionPending,
		},
		types.ProjectStateCompleted: {
			types.ProjectStateArchived,
			types.ProjectStateDeletionPending,
			types.ProjectStateActive, // Allow reopening completed projects
		},
		types.ProjectStateArchived: {
			types.ProjectStateActive, // Allow unarchiving
			types.ProjectStateDeletionPending,
		},
		types.ProjectStateDeletionPending: {
			// No transitions allowed from deletion pending
		},
	}

	// Allow staying in the same state
	if from == to {
		return true
	}

	// Handle empty/initial state - allow transition to active or completed
	if from == "" && (to == types.ProjectStateActive || to == types.ProjectStateCompleted) {
		return true
	}

	// Check if transition is in the valid list
	allowedStates, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowedState := range allowedStates {
		if to == allowedState {
			return true
		}
	}

	return false
}

// GetCurrentTime returns the current time
func (s *service) GetCurrentTime() time.Time {
	return time.Now()
}
