package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/sync/client"
	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"go.uber.org/zap"
)

// DataExtractor defines the interface for data extraction operations
type DataExtractor interface {
	// ExtractLocalData extracts data from local storage (SQLite)
	ExtractLocalData(ctx context.Context, projectID uuid.UUID) (*shared.SyncDataSet, error)

	// ExtractRemoteData extracts data from sync server via REST API
	ExtractRemoteData(ctx context.Context, projectID uuid.UUID) (*shared.SyncDataSet, error)

	// ExtractProjectData extracts data for a specific project from both sources
	ExtractProjectData(ctx context.Context, projectID uuid.UUID) (*shared.SyncDataSet, error)

	// ValidateDataSet validates a sync data set for consistency
	ValidateDataSet(ctx context.Context, dataSet *shared.SyncDataSet) error

	// FilterByTimeRange filters entities by their update timestamp
	FilterByTimeRange(dataSet *shared.SyncDataSet, since time.Time) *shared.SyncDataSet

	// GetStatistics returns statistics about the data set
	GetStatistics(dataSet *shared.SyncDataSet) *shared.DataStatistics
}

// dataExtractorImpl is the private implementation of DataExtractor
type dataExtractorImpl struct {
	logger         logger.Logger
	client         client.RESTSyncClient
	projectManager manager.ProjectManager
}

// Ensure dataExtractorImpl implements DataExtractor
var _ DataExtractor = (*dataExtractorImpl)(nil)

// NewDataExtractorService creates a new data extractor service for DI
func NewDataExtractorService(injector do.Injector) (DataExtractor, error) {
	logger := do.MustInvoke[logger.Logger](injector)
	projectManager := do.MustInvoke[manager.ProjectManager](injector)
	restClient := do.MustInvoke[client.RESTSyncClient](injector)

	logger.Debug("startup data extractor service")
	return &dataExtractorImpl{
		logger:         logger,
		projectManager: projectManager,
		client:         restClient,
	}, nil
}

// ExtractLocalData extracts data from local storage (SQLite)
func (e *dataExtractorImpl) ExtractLocalData(ctx context.Context, projectID uuid.UUID) (*shared.SyncDataSet, error) {
	e.logger.Info("Extracting local data",
		zap.String("project_id", projectID.String()))

	startTime := time.Now()
	dataSet := shared.NewSyncDataSet()

	// Extract projects from local storage
	projects, err := e.extractLocalProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to extract local projects: %w", err)
	}

	for _, project := range projects {
		dataSet.Projects[project.ID] = project
	}

	// Extract tasks from local storage
	tasks, err := e.extractLocalTasks(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract local tasks: %w", err)
	}

	for _, task := range tasks {
		dataSet.Tasks[task.ID] = task
	}

	e.logger.Info("Local data extraction completed",
		zap.Duration("duration", time.Since(startTime)),
		zap.Int("projects", len(dataSet.Projects)),
		zap.Int("tasks", len(dataSet.Tasks)))

	return dataSet, nil
}

// ExtractRemoteData extracts data from sync server via REST API
func (e *dataExtractorImpl) ExtractRemoteData(ctx context.Context, projectID uuid.UUID) (*shared.SyncDataSet, error) {
	if e.client == nil {
		return nil, fmt.Errorf("REST client is required for remote data extraction")
	}

	e.logger.Info("Extracting remote data from sync server",
		zap.String("project_id", projectID.String()))

	startTime := time.Now()

	// Fetch remote data using REST client
	dataSet, err := e.client.GetRemoteData(ctx, &projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch remote data: %w", err)
	}

	e.logger.Info("Remote data extraction completed",
		zap.Duration("duration", time.Since(startTime)),
		zap.Int("projects", len(dataSet.Projects)),
		zap.Int("tasks", len(dataSet.Tasks)))

	return dataSet, nil
}

// extractLocalProjects extracts projects from local storage
func (e *dataExtractorImpl) extractLocalProjects(ctx context.Context) ([]*types.Project, error) {
	e.logger.Debug("Extracting local projects from SQLite database")

	// For now, we'll return empty slice since we extract projects by ID during sync
	// The project is extracted separately in the main sync function
	return []*types.Project{}, nil
}

// extractLocalTasks extracts tasks from local storage
func (e *dataExtractorImpl) extractLocalTasks(ctx context.Context, projectID uuid.UUID) ([]*types.Task, error) {
	e.logger.Debug("Extracting local tasks from SQLite database",
		zap.String("project_id", projectID.String()))

	// Get all tasks for the project
	tasks, err := e.projectManager.ListTasksForProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks from project manager: %w", err)
	}

	// Collect task IDs and load with dependencies
	taskIDs := make([]uuid.UUID, 0, len(tasks))
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID)
	}

	// Load tasks with dependencies
	tasksWithDeps, err := e.projectManager.GetTasksWithDependencies(ctx, taskIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks with dependencies: %w", err)
	}

	e.logger.Info("Extracted local tasks with dependencies",
		zap.Int("count", len(tasksWithDeps)),
		zap.String("project_id", projectID.String()))
	return tasksWithDeps, nil
}

// ExtractProjectData extracts data for a specific project from both sources
func (e *dataExtractorImpl) ExtractProjectData(ctx context.Context, projectID uuid.UUID) (*shared.SyncDataSet, error) {
	e.logger.Info("Extracting project data from both sources",
		zap.String("project_id", projectID.String()))

	startTime := time.Now()

	// Extract local data
	localData, err := e.ExtractLocalData(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract local data: %w", err)
	}

	// Extract remote data
	remoteData, err := e.ExtractRemoteData(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract remote data: %w", err)
	}

	e.logger.Info("Project data extraction completed",
		zap.Duration("duration", time.Since(startTime)),
		zap.String("project_id", projectID.String()),
		zap.Int("local_projects", len(localData.Projects)),
		zap.Int("remote_projects", len(remoteData.Projects)),
		zap.Int("local_tasks", len(localData.Tasks)),
		zap.Int("remote_tasks", len(remoteData.Tasks)))

	// Combine data from both sources for comprehensive view
	combinedData := shared.NewSyncDataSet()

	// Add local projects
	for id, project := range localData.Projects {
		combinedData.Projects[id] = project
	}

	// Add remote projects that don't exist locally
	for id, project := range remoteData.Projects {
		if _, exists := combinedData.Projects[id]; !exists {
			combinedData.Projects[id] = project
		}
	}

	// Add local tasks
	for id, task := range localData.Tasks {
		combinedData.Tasks[id] = task
	}

	// Add remote tasks that don't exist locally
	for id, task := range remoteData.Tasks {
		if _, exists := combinedData.Tasks[id]; !exists {
			combinedData.Tasks[id] = task
		}
	}

	return combinedData, nil
}

// ValidateDataSet validates a sync data set for consistency
func (e *dataExtractorImpl) ValidateDataSet(ctx context.Context, dataSet *shared.SyncDataSet) error {
	e.logger.Debug("Validating sync data set",
		zap.Int("projects", len(dataSet.Projects)),
		zap.Int("tasks", len(dataSet.Tasks)))

	// Validate project references in tasks
	for _, task := range dataSet.Tasks {
		if task.ProjectID != uuid.Nil {
			if _, exists := dataSet.Projects[task.ProjectID]; !exists {
				e.logger.Warn("Task references non-existent project",
					zap.Any("task_id", task.ID),
					zap.String("task_title", task.Title),
					zap.Any("project_id", task.ProjectID))
			}
		}
	}

	// Validate timestamps
	now := time.Now()
	for _, project := range dataSet.Projects {
		if project.UpdatedAt.After(now) {
			return fmt.Errorf("project %s has future timestamp: %v", project.ID, project.UpdatedAt)
		}
	}

	for _, task := range dataSet.Tasks {
		if task.UpdatedAt.After(now) {
			return fmt.Errorf("task %s has future timestamp: %v", task.ID, task.UpdatedAt)
		}
	}

	e.logger.Debug("Data set validation completed successfully")
	return nil
}

// FilterByTimeRange filters entities by their update timestamp
func (e *dataExtractorImpl) FilterByTimeRange(dataSet *shared.SyncDataSet, since time.Time) *shared.SyncDataSet {
	e.logger.Debug("Filtering data set by time range",
		zap.Time("since", since))

	filtered := shared.NewSyncDataSet()

	// Filter projects
	for id, project := range dataSet.Projects {
		if project.UpdatedAt.After(since) {
			filtered.Projects[id] = project
		}
	}

	// Filter tasks
	for id, task := range dataSet.Tasks {
		if task.UpdatedAt.After(since) {
			filtered.Tasks[id] = task
		}
	}

	e.logger.Debug("Time range filtering completed",
		zap.Int("original_projects", len(dataSet.Projects)),
		zap.Int("filtered_projects", len(filtered.Projects)),
		zap.Int("original_tasks", len(dataSet.Tasks)),
		zap.Int("filtered_tasks", len(filtered.Tasks)))

	return filtered
}

// GetStatistics returns statistics about the data set
func (e *dataExtractorImpl) GetStatistics(dataSet *shared.SyncDataSet) *shared.DataStatistics {
	stats := &shared.DataStatistics{
		TotalProjects:   len(dataSet.Projects),
		TotalTasks:      len(dataSet.Tasks),
		TasksByState:    make(map[types.TaskState]int),
		ProjectsByState: make(map[types.ProjectState]int),
	}

	// Count tasks by state
	for _, task := range dataSet.Tasks {
		stats.TasksByState[task.State]++
	}

	// Count projects by state
	for _, project := range dataSet.Projects {
		stats.ProjectsByState[project.State]++
	}

	// Calculate age statistics
	var oldestProject, newestProject time.Time
	var oldestTask, newestTask time.Time

	firstProject := true
	firstTask := true

	for _, project := range dataSet.Projects {
		if firstProject || project.CreatedAt.Before(oldestProject) {
			oldestProject = project.CreatedAt
		}
		if firstProject || project.CreatedAt.After(newestProject) {
			newestProject = project.CreatedAt
		}
		firstProject = false
	}

	for _, task := range dataSet.Tasks {
		if firstTask || task.CreatedAt.Before(oldestTask) {
			oldestTask = task.CreatedAt
		}
		if firstTask || task.CreatedAt.After(newestTask) {
			newestTask = task.CreatedAt
		}
		firstTask = false
	}

	if len(dataSet.Projects) > 0 {
		stats.OldestProject = &oldestProject
		stats.NewestProject = &newestProject
		stats.ProjectAgeSpanDays = int(newestProject.Sub(oldestProject).Hours() / 24)
	}

	if len(dataSet.Tasks) > 0 {
		stats.OldestTask = &oldestTask
		stats.NewestTask = &newestTask
		stats.TaskAgeSpanDays = int(newestTask.Sub(oldestTask).Hours() / 24)
	}

	return stats
}
