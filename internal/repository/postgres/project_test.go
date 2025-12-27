package postgres

import (
	"context"
	"testing"
	"time"

	enttask "github.com/denkhaus/knot/v2/internal/repository/ent/task"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostgresRepository_DeleteProject_NoTasks tests deleting a project without tasks
func TestPostgresRepository_DeleteProject_NoTasks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepository(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test project
	projectID := uuid.New()
	testProject := &types.Project{
		ID:          projectID,
		Title:       "Test Project Without Tasks",
		Description: "Project to test deletion without tasks",
		State:       types.ProjectStateActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.CreateProject(ctx, testProject)
	require.NoError(t, err)

	// Verify project exists
	_, err = repo.GetProject(ctx, projectID)
	require.NoError(t, err)

	// Delete the project
	err = repo.DeleteProject(ctx, projectID)
	require.NoError(t, err)

	// Verify project is deleted
	_, err = repo.GetProject(ctx, projectID)
	assert.Error(t, err, "Project should not exist after deletion")
}

// TestPostgresRepository_DeleteProject_WithTasks tests deleting a project with tasks
func TestPostgresRepository_DeleteProject_WithTasks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepository(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test project
	projectID := uuid.New()
	testProject := &types.Project{
		ID:          projectID,
		Title:       "Test Project With Tasks",
		Description: "Project to test deletion with tasks",
		State:       types.ProjectStateActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.CreateProject(ctx, testProject)
	require.NoError(t, err)

	// Create multiple tasks
	task1ID := uuid.New()
	task1 := &types.Task{
		ID:          task1ID,
		ProjectID:   projectID,
		Title:       "Task 1",
		Description: "First task",
		State:       types.TaskStatePending,
		Priority:    types.TaskPriorityMedium,
		Complexity:  5,
		Depth:       0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	task2ID := uuid.New()
	task2 := &types.Task{
		ID:          task2ID,
		ProjectID:   projectID,
		Title:       "Task 2",
		Description: "Second task",
		State:       types.TaskStatePending,
		Priority:    types.TaskPriorityHigh,
		Complexity:  7,
		Depth:       0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.CreateTask(ctx, task1)
	require.NoError(t, err)

	err = repo.CreateTask(ctx, task2)
	require.NoError(t, err)

	// Create a dependency between tasks
	_, err = repo.AddTaskDependency(ctx, task2ID, task1ID)
	require.NoError(t, err)

	// Verify tasks exist
	_, err = repo.GetTask(ctx, task1ID)
	require.NoError(t, err)
	_, err = repo.GetTask(ctx, task2ID)
	require.NoError(t, err)

	// Delete the project (should cascade delete tasks and dependencies)
	err = repo.DeleteProject(ctx, projectID)
	require.NoError(t, err)

	// Verify project is deleted
	_, err = repo.GetProject(ctx, projectID)
	assert.Error(t, err, "Project should not exist after deletion")

	// Verify tasks are deleted
	_, err = repo.GetTask(ctx, task1ID)
	assert.Error(t, err, "Task 1 should not exist after project deletion")

	_, err = repo.GetTask(ctx, task2ID)
	assert.Error(t, err, "Task 2 should not exist after project deletion")
}

// TestPostgresRepository_DeleteProject_WithDependencies tests deleting a project with complex dependencies
func TestPostgresRepository_DeleteProject_WithDependencies(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepository(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test project
	projectID := uuid.New()
	testProject := &types.Project{
		ID:          projectID,
		Title:       "Test Project With Complex Dependencies",
		Description: "Project to test deletion with complex dependencies",
		State:       types.ProjectStateActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.CreateProject(ctx, testProject)
	require.NoError(t, err)

	// Create tasks with hierarchical relationships
	parentTaskID := uuid.New()
	parentTask := &types.Task{
		ID:          parentTaskID,
		ProjectID:   projectID,
		Title:       "Parent Task",
		Description: "Parent task with children",
		State:       types.TaskStatePending,
		Priority:    types.TaskPriorityHigh,
		Complexity:  8,
		Depth:       0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	childTask1ID := uuid.New()
	childTask1 := &types.Task{
		ID:          childTask1ID,
		ProjectID:   projectID,
		Title:       "Child Task 1",
		Description: "Child task 1",
		State:       types.TaskStatePending,
		Priority:    types.TaskPriorityMedium,
		Complexity:  5,
		Depth:       1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ParentID:    &parentTaskID,
	}

	childTask2ID := uuid.New()
	childTask2 := &types.Task{
		ID:          childTask2ID,
		ProjectID:   projectID,
		Title:       "Child Task 2",
		Description: "Child task 2",
		State:       types.TaskStatePending,
		Priority:    types.TaskPriorityLow,
		Complexity:  3,
		Depth:       1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ParentID:    &parentTaskID,
	}

	err = repo.CreateTask(ctx, parentTask)
	require.NoError(t, err)

	err = repo.CreateTask(ctx, childTask1)
	require.NoError(t, err)

	err = repo.CreateTask(ctx, childTask2)
	require.NoError(t, err)

	// Create cross-task dependencies
	_, err = repo.AddTaskDependency(ctx, childTask1ID, parentTaskID)
	require.NoError(t, err)

	_, err = repo.AddTaskDependency(ctx, childTask2ID, parentTaskID)
	require.NoError(t, err)

	// Verify all tasks and dependencies exist
	tasks, err := repo.GetTasksByProject(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, tasks, 3, "Should have 3 tasks before deletion")

	// Delete the project (should cascade delete all tasks and dependencies)
	err = repo.DeleteProject(ctx, projectID)
	require.NoError(t, err)

	// Verify project is deleted
	_, err = repo.GetProject(ctx, projectID)
	assert.Error(t, err, "Project should not exist after deletion")

	// Verify all tasks are deleted
	_, err = repo.GetTask(ctx, parentTaskID)
	assert.Error(t, err, "Parent task should not exist after project deletion")

	_, err = repo.GetTask(ctx, childTask1ID)
	assert.Error(t, err, "Child task 1 should not exist after project deletion")

	_, err = repo.GetTask(ctx, childTask2ID)
	assert.Error(t, err, "Child task 2 should not exist after project deletion")

	// Verify no tasks remain for this project
	tasks, err = repo.GetTasksByProject(ctx, projectID)
	require.NoError(t, err)
	assert.Empty(t, tasks, "Should have no tasks after project deletion")
}

// TestPostgresRepository_DeleteProject_NonExistent tests deleting a non-existent project
func TestPostgresRepository_DeleteProject_NonExistent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepository(t)
	defer cleanup()

	ctx := context.Background()

	// Try to delete a non-existent project
	fakeProjectID := uuid.New()
	err := repo.DeleteProject(ctx, fakeProjectID)

	assert.Error(t, err, "Deleting non-existent project should return an error")
}

// TestPostgresRepository_DeleteProject_EmptyTasksList tests deleting a project when task query returns empty list
func TestPostgresRepository_DeleteProject_EmptyTasksList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepository(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test project
	projectID := uuid.New()
	testProject := &types.Project{
		ID:          projectID,
		Title:       "Test Empty Project",
		Description: "Project with no tasks",
		State:       types.ProjectStateActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.CreateProject(ctx, testProject)
	require.NoError(t, err)

	// Directly verify no tasks exist using ent client
	count, err := repo.client.Task.Query().
		Where(enttask.ProjectID(projectID)).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Project should have no tasks")

	// Delete the project (should work even with no tasks)
	err = repo.DeleteProject(ctx, projectID)
	require.NoError(t, err)

	// Verify project is deleted
	_, err = repo.GetProject(ctx, projectID)
	assert.Error(t, err, "Project should not exist after deletion")
}

// TestPostgresRepository_DeleteProject_MultipleProjects tests deleting one project doesn't affect others
func TestPostgresRepository_DeleteProject_MultipleProjects(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepository(t)
	defer cleanup()

	ctx := context.Background()

	// Create two projects
	project1ID := uuid.New()
	project1 := &types.Project{
		ID:          project1ID,
		Title:       "Project 1",
		Description: "First project",
		State:       types.ProjectStateActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	project2ID := uuid.New()
	project2 := &types.Project{
		ID:          project2ID,
		Title:       "Project 2",
		Description: "Second project",
		State:       types.ProjectStateActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.CreateProject(ctx, project1)
	require.NoError(t, err)

	err = repo.CreateProject(ctx, project2)
	require.NoError(t, err)

	// Add tasks to both projects
	task1ID := uuid.New()
	task1 := &types.Task{
		ID:          task1ID,
		ProjectID:   project1ID,
		Title:       "Task in Project 1",
		Description: "Task belonging to project 1",
		State:       types.TaskStatePending,
		Priority:    types.TaskPriorityMedium,
		Complexity:  5,
		Depth:       0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	task2ID := uuid.New()
	task2 := &types.Task{
		ID:          task2ID,
		ProjectID:   project2ID,
		Title:       "Task in Project 2",
		Description: "Task belonging to project 2",
		State:       types.TaskStatePending,
		Priority:    types.TaskPriorityHigh,
		Complexity:  7,
		Depth:       0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.CreateTask(ctx, task1)
	require.NoError(t, err)

	err = repo.CreateTask(ctx, task2)
	require.NoError(t, err)

	// Delete project 1
	err = repo.DeleteProject(ctx, project1ID)
	require.NoError(t, err)

	// Verify project 1 is deleted
	_, err = repo.GetProject(ctx, project1ID)
	assert.Error(t, err, "Project 1 should not exist after deletion")

	// Verify project 1's task is deleted
	_, err = repo.GetTask(ctx, task1ID)
	assert.Error(t, err, "Task 1 should not exist after project 1 deletion")

	// Verify project 2 still exists
	project2After, err := repo.GetProject(ctx, project2ID)
	require.NoError(t, err)
	assert.Equal(t, project2ID, project2After.ID, "Project 2 should still exist")

	// Verify project 2's task still exists
	task2After, err := repo.GetTask(ctx, task2ID)
	require.NoError(t, err)
	assert.Equal(t, task2ID, task2After.ID, "Task 2 should still exist")
}
