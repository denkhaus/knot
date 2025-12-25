package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/sync/client"
	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"go.uber.org/zap"
)

// service provides bidirectional synchronization between local SQLite and sync server
type service struct {
	projectManager  manager.ProjectManager
	logger          logger.Logger
	syncConfig      *config.SyncConfig
	restClient      client.RESTSyncClient
	migrationEngine MigrationEngine
	dataExtractor   DataExtractor
	diffEngine      DiffEngine
}

// Ensure service implements SyncManager
var _ SyncManager = (*service)(nil)

// NewSyncManager creates a new sync manager service instance following the DI pattern.
// This constructor resolves dependencies from the DI container and returns the public interface.
func NewSyncManager(injector do.Injector) (SyncManager, error) {
	// Resolve required dependencies from DI container
	projectManager := do.MustInvoke[manager.ProjectManager](injector)
	log := do.MustInvoke[logger.Logger](injector)
	cnf := do.MustInvoke[config.Service](injector)

	// Resolve sync-specific services
	dataExtractor := do.MustInvoke[DataExtractor](injector)
	diffEngine := do.MustInvoke[DiffEngine](injector)

	// Try to resolve REST sync client (optional - may not be available in all modes)
	restClient, err := do.Invoke[client.RESTSyncClient](injector)
	if err != nil {
		// REST client is optional - log and continue
		log.Debug("REST sync client not available in DI container", zap.Error(err))
	}

	// Try to resolve MigrationEngine (optional - may not be available in all modes)
	var migrationEngine MigrationEngine
	if restClient != nil {
		migrationEngine, err = do.Invoke[MigrationEngine](injector)
		if err != nil {
			log.Debug("MigrationEngine not available in DI container", zap.Error(err))
		}
	}

	return &service{
		syncConfig:      cnf.GetSyncConfig(),
		projectManager:  projectManager,
		logger:          log,
		restClient:      restClient,
		migrationEngine: migrationEngine,
		dataExtractor:   dataExtractor,
		diffEngine:      diffEngine,
	}, nil
}

// SyncWithLocalData syncs local data to remote (local → MCP)
func (s *service) SyncWithLocalData(ctx context.Context, localData *shared.SyncDataSet) (*SyncResult, error) {
	s.logger.Info("Starting local to MCP sync",
		zap.Int("projects", len(localData.Projects)),
		zap.Int("tasks", len(localData.Tasks)))

	startTime := time.Now()
	result := &SyncResult{
		SyncedAt: time.Now(),
		Success:  true,
	}

	// Use migration engine to apply local data
	if s.migrationEngine != nil {
		migrationResult, err := s.migrationEngine.PushSync(ctx, localData)
		if err != nil {
			result.Success = false
			result.Errors = append(result.Errors, fmt.Sprintf("failed to apply local data: %v", err))
			return result, fmt.Errorf("failed to apply local data: %w", err)
		}

		result.Processed = migrationResult.Processed
		result.Created = migrationResult.Created
		result.Updated = migrationResult.Updated
		result.Deleted = migrationResult.Deleted
	} else {
		// No migration engine - just count what we received
		result.Processed = len(localData.Projects) + len(localData.Tasks)
		result.Created = len(localData.Projects) + len(localData.Tasks)
	}

	result.Duration = time.Since(startTime)

	s.logger.Info("Local to MCP sync completed",
		zap.Duration("duration", result.Duration),
		zap.Int("processed", result.Processed),
		zap.Int("created", result.Created),
		zap.Int("updated", result.Updated))

	return result, nil
}

// SyncPullFromRemote fetches remote data and syncs it to local (MCP → local)
func (s *service) SyncPullFromRemote(ctx context.Context, projectID uuid.UUID) (*SyncResult, error) {
	s.logger.Info("Starting MCP to local sync",
		zap.String("project_id", projectID.String()))

	startTime := time.Now()
	result := &SyncResult{
		SyncedAt: time.Now(),
		Success:  true,
	}

	// Use migration engine to fetch and apply remote data
	if s.migrationEngine != nil {
		migrationResult, err := s.migrationEngine.PullSync(ctx, &projectID)
		if err != nil {
			result.Success = false
			result.Errors = append(result.Errors, fmt.Sprintf("failed to pull remote data: %v", err))
			return result, fmt.Errorf("failed to pull remote data: %w", err)
		}

		result.Processed = migrationResult.Processed
		result.Created = migrationResult.Created
		result.Updated = migrationResult.Updated
		result.Deleted = migrationResult.Deleted
	} else {
		// No migration engine - error
		result.Success = false
		result.Errors = append(result.Errors, "migration engine not available")
		return result, fmt.Errorf("migration engine not available")
	}

	result.Duration = time.Since(startTime)

	s.logger.Info("MCP to local sync completed",
		zap.Duration("duration", result.Duration),
		zap.Int("processed", result.Processed),
		zap.Int("created", result.Created),
		zap.Int("updated", result.Updated))

	return result, nil
}

// SyncBidirectional performs bidirectional sync
func (s *service) SyncBidirectional(ctx context.Context, localData *shared.SyncDataSet) (*SyncResult, error) {
	s.logger.Info("Starting bidirectional sync",
		zap.Int("local_projects", len(localData.Projects)),
		zap.Int("local_tasks", len(localData.Tasks)))

	startTime := time.Now()
	result := &SyncResult{
		SyncedAt: time.Now(),
		Success:  true,
	}

	// Use migration engine for full sync
	if s.migrationEngine != nil {
		migrationResult, err := s.migrationEngine.FullSync(ctx, localData)
		if err != nil {
			result.Success = false
			result.Errors = append(result.Errors, fmt.Sprintf("failed to apply sync: %v", err))
			return result, fmt.Errorf("failed to apply sync: %w", err)
		}

		result.Processed = migrationResult.Processed
		result.Created = migrationResult.Created
		result.Updated = migrationResult.Updated
		result.Deleted = migrationResult.Deleted
	} else {
		// No migration engine
		result.Processed = len(localData.Projects) + len(localData.Tasks)
		result.Created = len(localData.Projects) + len(localData.Tasks)
	}

	result.Duration = time.Since(startTime)

	s.logger.Info("Bidirectional sync completed",
		zap.Duration("duration", result.Duration),
		zap.Int("processed", result.Processed),
		zap.Int("created", result.Created))

	return result, nil
}

// DEPRECATED: Use SyncWithLocalData, SyncPullFromRemote, or SyncBidirectional instead
// Sync performs bidirectional synchronization using available infrastructure
// This method is kept for backward compatibility and may be removed in future versions
func (s *service) Sync(ctx context.Context, projectID uuid.UUID, direction shared.SyncDirection) (*SyncResult, error) {
	s.logger.Info("Starting complete synchronization (DEPRECATED method)",
		zap.String("project_id", projectID.String()),
		zap.String("direction", string(direction)))

	// Delegate to new methods based on direction
	switch direction {
	case shared.SyncLocalToMCP:
		// Extract local data and push to remote
		project, err := s.projectManager.GetProject(ctx, projectID)
		if err != nil {
			return &SyncResult{Success: false, Errors: []string{err.Error()}, SyncedAt: time.Now()}, err
		}

		localData, err := s.dataExtractor.ExtractLocalData(ctx, s.projectManager, projectID)
		if err != nil {
			return &SyncResult{Success: false, Errors: []string{err.Error()}, SyncedAt: time.Now()}, err
		}
		localData.Projects[project.ID] = project

		return s.SyncWithLocalData(ctx, localData)

	case shared.SyncMcpToLocal:
		// Pull from remote and apply locally
		return s.SyncPullFromRemote(ctx, projectID)

	case shared.SyncBidirectional:
		// Extract local data and perform bidirectional sync
		project, err := s.projectManager.GetProject(ctx, projectID)
		if err != nil {
			return &SyncResult{Success: false, Errors: []string{err.Error()}, SyncedAt: time.Now()}, err
		}

		localData, err := s.dataExtractor.ExtractLocalData(ctx, s.projectManager, projectID)
		if err != nil {
			return &SyncResult{Success: false, Errors: []string{err.Error()}, SyncedAt: time.Now()}, err
		}
		localData.Projects[project.ID] = project

		return s.SyncBidirectional(ctx, localData)

	default:
		return &SyncResult{Success: false, Errors: []string{"unsupported direction"}, SyncedAt: time.Now()},
			fmt.Errorf("unsupported sync direction: %s", direction)
	}
}

