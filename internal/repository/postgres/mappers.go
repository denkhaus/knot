package postgres

import (
	"github.com/denkhaus/knot/v2/internal/repository/ent"
	entproject "github.com/denkhaus/knot/v2/internal/repository/ent/project"
	enttask "github.com/denkhaus/knot/v2/internal/repository/ent/task"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
)

// Project state conversion functions

// projectStateToEntState converts domain ProjectState to ent project state
func projectStateToEntState(state types.ProjectState) entproject.State {
	switch state {
	case types.ProjectStateActive:
		return entproject.StateActive
	case types.ProjectStateCompleted:
		return entproject.StateCompleted
	case types.ProjectStateArchived:
		return entproject.StateArchived
	case types.ProjectStateDeletionPending:
		return entproject.StateDeletionPending
	default:
		return entproject.StateActive // Default to active for empty/unknown states
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
		CreatedBy:      ep.CreatedBy,
		UpdatedBy:      ep.UpdatedBy,
	}
}

// projectToEntProjectCreate converts domain Project model to ent ProjectCreate
func projectToEntProjectCreate(p *types.Project, client *ent.Client) *ent.ProjectCreate {
	create := client.Project.Create().
		SetTitle(p.Title).
		SetDescription(p.Description).
		SetState(projectStateToEntState(p.State)).
		SetTotalTasks(p.TotalTasks).
		SetCompletedTasks(p.CompletedTasks).
		SetProgress(p.Progress)

	if p.ID != uuid.Nil {
		create.SetID(p.ID)
	}
	if !p.CreatedAt.IsZero() {
		create.SetCreatedAt(p.CreatedAt)
	}
	if !p.UpdatedAt.IsZero() {
		create.SetUpdatedAt(p.UpdatedAt)
	}
	if p.CreatedBy != "" {
		create.SetCreatedBy(p.CreatedBy)
	}
	if p.UpdatedBy != "" {
		create.SetUpdatedBy(p.UpdatedBy)
	}

	return create
}

// Task entity mapping functions

// Priority conversion functions

// taskPriorityToEntPriority converts domain TaskPriority to ent task.Priority
func taskPriorityToEntPriority(priority types.TaskPriority) enttask.Priority {
	switch priority {
	case types.TaskPriorityHigh:
		return enttask.PriorityHigh
	case types.TaskPriorityMedium:
		return enttask.PriorityMedium
	case types.TaskPriorityLow:
		return enttask.PriorityLow
	default:
		return enttask.PriorityMedium // Default to medium for invalid values
	}
}

// entPriorityToTaskPriority converts ent task.Priority to domain TaskPriority
func entPriorityToTaskPriority(p enttask.Priority) types.TaskPriority {
	switch p {
	case enttask.PriorityHigh:
		return types.TaskPriorityHigh
	case enttask.PriorityMedium:
		return types.TaskPriorityMedium
	case enttask.PriorityLow:
		return types.TaskPriorityLow
	default:
		return types.TaskPriorityMedium // Default to medium for invalid values
	}
}

// entTaskToTask converts ent Task entity to domain Task model
func entTaskToTask(et *ent.Task) *types.Task {
	domainTask := &types.Task{
		ID:          et.ID,
		ProjectID:   et.ProjectID,
		Title:       et.Title,
		Description: et.Description,
		State:       types.TaskState(et.State),
		Priority:    entPriorityToTaskPriority(et.Priority),
		Complexity:  et.Complexity,
		Depth:       et.Depth,
		CreatedAt:   et.CreatedAt,
		UpdatedAt:   et.UpdatedAt,
		CreatedBy:   et.CreatedBy,
		UpdatedBy:   et.UpdatedBy,
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

// Helper functions for slice conversions

// entProjectsToProjects converts slice of ent Projects to domain Projects
func entProjectsToProjects(entProjects []*ent.Project) []*types.Project {
	projects := make([]*types.Project, len(entProjects))
	for i, ep := range entProjects {
		projects[i] = entProjectToProject(ep)
	}
	return projects
}