package health

import (
	"context"

	"time"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
)

// MockProjectManager is a mock implementation of the ProjectManager interface for testing.
type MockProjectManager struct {
	// Expectations for ListProjects
	ListProjectsFunc func(ctx context.Context) ([]*types.Project, error)
	// Expectations for GetConfig
	GetConfigFunc func() *manager.Config
}

// Ensure MockProjectManager implements the ProjectManager interface.
var _ manager.ProjectManager = (*MockProjectManager)(nil)

// NewMockProjectManager creates a new instance of MockProjectManager.
func NewMockProjectManager() *MockProjectManager {
	return &MockProjectManager{
		ListProjectsFunc: func(ctx context.Context) ([]*types.Project, error) {
			return []*types.Project{}, nil // Default: return empty list, no error
		},
		GetConfigFunc: func() *manager.Config {
			return manager.DefaultConfig() // Default: return default config
		},
	}
}

// Implement ProjectManager interface methods.
// Only implement the methods used by the health commands.

func (m *MockProjectManager) CreateProject(ctx context.Context, title, description, actor string) (*types.Project, error) {
	panic("not implemented")
}
func (m *MockProjectManager) GetProject(ctx context.Context, projectID uuid.UUID) (*types.Project, error) {
	panic("not implemented")
}
func (m *MockProjectManager) UpdateProject(ctx context.Context, projectID uuid.UUID, title, description string, actor string) (*types.Project, error) {
	panic("not implemented")
}
func (m *MockProjectManager) UpdateProjectDescription(ctx context.Context, projectID uuid.UUID, description string, actor string) (*types.Project, error) {
	panic("not implemented")
}
func (m *MockProjectManager) UpdateProjectState(ctx context.Context, projectID uuid.UUID, state types.ProjectState, actor string) (*types.Project, error) {
	panic("not implemented")
}
func (m *MockProjectManager) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	panic("not implemented")
}
func (m *MockProjectManager) ListProjects(ctx context.Context) ([]*types.Project, error) {
	return m.ListProjectsFunc(ctx)
}
func (m *MockProjectManager) CreateTask(ctx context.Context, projectID uuid.UUID, parentID *uuid.UUID, title, description string, complexity int, priority types.TaskPriority, actor string) (*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) GetTask(ctx context.Context, taskID uuid.UUID) (*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) GetTasksWithDependencies(ctx context.Context, taskIDs []uuid.UUID) ([]*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) UpdateTask(ctx context.Context, taskID uuid.UUID, title, description string, complexity int, state types.TaskState, actor string) (*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) UpdateTaskDescription(ctx context.Context, taskID uuid.UUID, description string, actor string) (*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) UpdateTaskTitle(ctx context.Context, taskID uuid.UUID, title string, actor string) (*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) UpdateTaskPriority(ctx context.Context, taskID uuid.UUID, priority types.TaskPriority, actor string) (*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) UpdateTaskState(ctx context.Context, taskID uuid.UUID, state types.TaskState, actor string) (*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) DeleteTask(ctx context.Context, taskID uuid.UUID, actor string) error {
	panic("not implemented")
}
func (m *MockProjectManager) DeleteTaskSubtree(ctx context.Context, taskID uuid.UUID, actor string) error {
	panic("not implemented")
}
func (m *MockProjectManager) GetParentTask(ctx context.Context, taskID uuid.UUID) (*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) GetChildTasks(ctx context.Context, taskID uuid.UUID) ([]*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) GetRootTasks(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) ListTasksForProject(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) FindNextActionableTask(ctx context.Context, projectID uuid.UUID) (*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) FindTasksNeedingBreakdown(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) GetProjectProgress(ctx context.Context, projectID uuid.UUID) (*types.ProjectProgress, error) {
	panic("not implemented")
}
func (m *MockProjectManager) ListTasksByState(ctx context.Context, projectID uuid.UUID, state types.TaskState) ([]*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) BulkUpdateTasks(ctx context.Context, taskIDs []uuid.UUID, updates types.TaskUpdates, actor string) error {
	panic("not implemented")
}
func (m *MockProjectManager) DuplicateTask(ctx context.Context, taskID uuid.UUID, newProjectID uuid.UUID) (*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) SetTaskEstimate(ctx context.Context, taskID uuid.UUID, estimate int64) (*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) AssignTaskToAgent(ctx context.Context, taskID uuid.UUID, agentID uuid.UUID) (*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) UnassignTaskFromAgent(ctx context.Context, taskID uuid.UUID) (*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) ListTasksByAgent(ctx context.Context, projectID uuid.UUID, agentID uuid.UUID) ([]*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) ListUnassignedTasks(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) AddTaskDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID, actor string) (*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) RemoveTaskDependency(ctx context.Context, taskID uuid.UUID, dependsOnTaskID uuid.UUID, actor string) (*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) GetTaskDependencies(ctx context.Context, taskID uuid.UUID) ([]*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) GetDependentTasks(ctx context.Context, taskID uuid.UUID) ([]*types.Task, error) {
	panic("not implemented")
}
func (m *MockProjectManager) GetConfig() *manager.Config {
	return m.GetConfigFunc()
}
func (m *MockProjectManager) UpdateConfig(config *manager.Config) {
	panic("not implemented")
}
func (m *MockProjectManager) LoadConfigFromFile() error {
	panic("not implemented")
}
func (m *MockProjectManager) SaveConfigToFile() error {
	panic("not implemented")
}
func (m *MockProjectManager) GetSelectedProject(ctx context.Context) (*uuid.UUID, error) {
	panic("not implemented")
}
func (m *MockProjectManager) SetSelectedProject(ctx context.Context, projectID uuid.UUID, actor string) error {
	panic("not implemented")
}
func (m *MockProjectManager) ClearSelectedProject(ctx context.Context) error {
	panic("not implemented")
}
func (m *MockProjectManager) HasSelectedProject(ctx context.Context) (bool, error) {
	panic("not implemented")
}
func (m *MockProjectManager) GetCurrentTime() time.Time {
	return time.Now()
}
