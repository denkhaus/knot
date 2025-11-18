package sqlite

import (
	"github.com/denkhaus/knot/internal/repository/sqlite/ent"
	"github.com/denkhaus/knot/internal/repository/sqlite/ent/project"
	"github.com/denkhaus/knot/internal/repository/sqlite/ent/task"
	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
)

// Priority conversion functions

// domainPriorityToEntPriority converts domain TaskPriority to ent task.Priority
func domainPriorityToEntPriority(p types.TaskPriority) task.Priority {
	switch p {
	case types.TaskPriorityHigh:
		return task.PriorityHigh
	case types.TaskPriorityMedium:
		return task.PriorityMedium
	case types.TaskPriorityLow:
		return task.PriorityLow
	default:
		return task.PriorityMedium // Default to medium for invalid values
	}
}

// entPriorityToDomainPriority converts ent task.Priority to domain TaskPriority
func entPriorityToDomainPriority(p task.Priority) types.TaskPriority {
	switch p {
	case task.PriorityHigh:
		return types.TaskPriorityHigh
	case task.PriorityMedium:
		return types.TaskPriorityMedium
	case task.PriorityLow:
		return types.TaskPriorityLow
	default:
		return types.TaskPriorityMedium // Default to medium for invalid values
	}
}

// Project entity mapping functions

// entProjectToProject converts ent Project entity to domain Project model
func entProjectToProject(ep *ent.Project) *types.Project {
	return &types.Project{
		ID:             ep.ID,
		Title:          ep.Title,
		Description:    ep.Description,
		State:          entStateToProjectState(string(ep.State)),
		CreatedAt:      ep.CreatedAt,
		UpdatedAt:      ep.UpdatedAt,
		TotalTasks:     ep.TotalTasks,
		CompletedTasks: ep.CompletedTasks,
		Progress:       ep.Progress,
	}
}

// projectToEntProjectCreate converts domain Project model to ent ProjectCreate
func projectToEntProjectCreate(p *types.Project, client *ent.Client) *ent.ProjectCreate {
	create := client.Project.Create().
		SetTitle(p.Title).
		SetDescription(p.Description).
		SetState(projectStateToEntState(p.State))

	if p.ID != uuid.Nil {
		create.SetID(p.ID)
	}
	if !p.CreatedAt.IsZero() {
		create.SetCreatedAt(p.CreatedAt)
	}
	if !p.UpdatedAt.IsZero() {
		create.SetUpdatedAt(p.UpdatedAt)
	}
	if p.TotalTasks > 0 {
		create.SetTotalTasks(p.TotalTasks)
	}
	if p.CompletedTasks > 0 {
		create.SetCompletedTasks(p.CompletedTasks)
	}
	if p.Progress > 0 {
		create.SetProgress(p.Progress)
	}

	return create
}

// Task entity mapping functions

// entTaskToTask converts ent Task entity to domain Task model
func entTaskToTask(et *ent.Task) *types.Task {
	domainTask := &types.Task{
		ID:          et.ID,
		ProjectID:   et.ProjectID,
		Title:       et.Title,
		Description: et.Description,
		State:       types.TaskState(et.State),
		Priority:    entPriorityToDomainPriority(et.Priority),
		Complexity:  et.Complexity,
		Depth:       et.Depth,
		CreatedAt:   et.CreatedAt,
		UpdatedAt:   et.UpdatedAt,
	}

	// Handle optional/nullable fields
	if et.ParentID != nil {
		domainTask.ParentID = et.ParentID
	}
	if et.Estimate != nil {
		domainTask.Estimate = et.Estimate
	}
	if et.AssignedAgent != nil {
		domainTask.AssignedAgent = et.AssignedAgent
	}
	if et.CompletedAt != nil {
		domainTask.CompletedAt = et.CompletedAt
	}

	// Initialize slices to avoid nil pointer issues
	domainTask.Dependencies = make([]uuid.UUID, 0)
	domainTask.Dependents = make([]uuid.UUID, 0)

	return domainTask
}

// taskToEntTaskCreate converts domain Task model to ent TaskCreate
func taskToEntTaskCreate(t *types.Task, client *ent.Client) *ent.TaskCreate {
	create := client.Task.Create().
		SetProjectID(t.ProjectID).
		SetTitle(t.Title).
		SetDescription(t.Description).
		SetState(task.State(t.State)).
		SetPriority(domainPriorityToEntPriority(t.Priority)).
		SetComplexity(t.Complexity).
		SetDepth(t.Depth)

	if t.ID != uuid.Nil {
		create.SetID(t.ID)
	}
	if t.ParentID != nil {
		create.SetParentID(*t.ParentID)
	}
	if t.Estimate != nil {
		create.SetEstimate(*t.Estimate)
	}
	if t.AssignedAgent != nil {
		create.SetAssignedAgent(*t.AssignedAgent)
	}
	if !t.CreatedAt.IsZero() {
		create.SetCreatedAt(t.CreatedAt)
	}
	if !t.UpdatedAt.IsZero() {
		create.SetUpdatedAt(t.UpdatedAt)
	}
	if t.CompletedAt != nil {
		create.SetCompletedAt(*t.CompletedAt)
	}

	return create
}

// taskToEntTaskUpdate converts domain Task model to ent TaskUpdateOne
func taskToEntTaskUpdate(t *types.Task, update *ent.TaskUpdateOne) *ent.TaskUpdateOne {
	update = update.
		SetTitle(t.Title).
		SetDescription(t.Description).
		SetState(task.State(t.State)).
		SetPriority(domainPriorityToEntPriority(t.Priority)).
		SetComplexity(t.Complexity).
		SetUpdatedAt(t.UpdatedAt)

	if t.Estimate != nil {
		update.SetEstimate(*t.Estimate)
	} else {
		update.ClearEstimate()
	}

	if t.AssignedAgent != nil {
		update.SetAssignedAgent(*t.AssignedAgent)
	} else {
		update.ClearAssignedAgent()
	}

	if t.CompletedAt != nil {
		update.SetCompletedAt(*t.CompletedAt)
	} else {
		update.ClearCompletedAt()
	}

	return update
}

// TaskDependency entity mapping functions

// entTaskDependenciesToTaskIDs extracts task IDs from ent TaskDependency entities
func entTaskDependenciesToTaskIDs(dependencies []*ent.TaskDependency) []uuid.UUID {
	ids := make([]uuid.UUID, len(dependencies))
	for i, dep := range dependencies {
		ids[i] = dep.DependsOnTaskID
	}
	return ids
}

// entTaskDependentsToTaskIDs extracts dependent task IDs from ent TaskDependency entities
func entTaskDependentsToTaskIDs(dependents []*ent.TaskDependency) []uuid.UUID {
	ids := make([]uuid.UUID, len(dependents))
	for i, dep := range dependents {
		ids[i] = dep.TaskID
	}
	return ids
}

// Helper functions for slice conversions

// entProjectsToProjects converts slice of ent Projects to domain Projects
func entProjectsToProjects(entProjects []*ent.Project) []*types.Project {
	projects := make([]*types.Project, len(entProjects))
	for i, ep := range entProjects {
		projects[i] = entProjectToProject(ep)
	}
	return projects
}

// entTasksToTasks converts slice of ent Tasks to domain Tasks
// Currently unused but kept for potential future use
// func entTasksToTasks(entTasks []*ent.Task) []*types.Task {
// 	tasks := make([]*types.Task, len(entTasks))
// 	for i, et := range entTasks {
// 		tasks[i] = entTaskToTask(et)
// 	}
// 	return tasks
// }

// Helper functions for filtering

// filterMatchesTaskFilter checks if an ent Task matches the given TaskFilter
// Currently unused but kept for potential future use
// func filterMatchesTaskFilter(task *ent.Task, filter types.TaskFilter) bool {
// 	if filter.ProjectID != nil && task.ProjectID != *filter.ProjectID {
// 		return false
// 	}
// 	if filter.ParentID != nil {
// 		if task.ParentID == nil && *filter.ParentID != uuid.Nil {
// 			return false
// 		}
// 		if task.ParentID != nil && *task.ParentID != *filter.ParentID {
// 			return false
// 		}
// 	}
// 	if filter.State != nil && types.TaskState(task.State) != *filter.State {
// 		return false
// 	}
// 	if filter.Priority != nil && types.TaskPriority(task.Priority) != *filter.Priority {
// 		return false
// 	}
// 	if filter.MinDepth != nil && task.Depth < *filter.MinDepth {
// 		return false
// 	}
// 	if filter.MaxDepth != nil && task.Depth > *filter.MaxDepth {
// 		return false
// 	}
// 	if filter.MinComplexity != nil && task.Complexity < *filter.MinComplexity {
// 		return false
// 	}
// 	if filter.MaxComplexity != nil && task.Complexity > *filter.MaxComplexity {
// 		return false
// 	}
// 	return true
// }

// Project state conversion functions

// projectStateToEntState converts domain ProjectState to ent project state
func projectStateToEntState(state types.ProjectState) project.State {
	switch state {
	case types.ProjectStateActive:
		return project.StateActive
	case types.ProjectStateCompleted:
		return project.StateCompleted
	case types.ProjectStateArchived:
		return project.StateArchived
	case types.ProjectStateDeletionPending:
		return project.StateDeletionPending
	default:
		return project.StateActive // Default to active for empty/unknown states
	}
}

// entStateToProjectState converts ent project state to domain ProjectState
func entStateToProjectState(state string) types.ProjectState {
	switch state {
	case "active":
		return types.ProjectStateActive
	case "completed":
		return types.ProjectStateCompleted
	case "archived":
		return types.ProjectStateArchived
	case "deletion-pending":
		return types.ProjectStateDeletionPending
	default:
		return types.ProjectStateActive // Default to active for unknown states
	}
}

