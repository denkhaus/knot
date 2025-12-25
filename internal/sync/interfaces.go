// Package sync provides bidirectional synchronization capabilities between
// local CLI SQLite workspaces and MCP PostgreSQL database.
//
// This package implements the core synchronization logic that sits between
// the CLI layer and the repository/MCP client layers. It handles:
//   - Bidirectional data synchronization with conflict resolution
//   - Timestamp-based change detection and incremental sync
//   - MCP client integration for server communication
//   - Comprehensive error handling and rollback capabilities
//
// Key Features:
//   - CLI-based synchronization commands
//   - Project-level sync configuration and state tracking
//   - Configurable conflict resolution strategies
//   - Integration with existing project and task management
//   - Offline-first architecture with server sync capability
//
// Example Usage:
//
//	manager := sync.NewSyncManager(injector)
//	err := manager.InitSync(ctx, projectID, config)
//	if err != nil {
//		return fmt.Errorf("failed to init sync: %w", err)
//	}
//
//	result, err := manager.Sync(ctx, projectID, sync.SyncLocalToMCP)
package sync

import (
	"context"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/google/uuid"
)

// SyncManager defines the public interface for bidirectional synchronization
// between local CLI SQLite workspaces and MCP PostgreSQL database.
//
// This interface provides high-level synchronization operations that include:
//   - Project-level synchronization configuration and status tracking
//   - Bidirectional data transfer with conflict resolution
//   - Timestamp-based change detection and incremental sync
//   - MCP client integration for server communication
//   - Comprehensive error handling and rollback capabilities
//
// Implementations should be thread-safe and handle concurrent sync operations.
//

type SyncManager interface {
	// SyncWithLocalData syncs local data to remote (local → MCP)
	SyncWithLocalData(ctx context.Context, localData *shared.SyncDataSet) (*SyncResult, error)

	// SyncPullFromRemote fetches remote data and syncs it to local (MCP → local)
	SyncPullFromRemote(ctx context.Context, projectID uuid.UUID) (*SyncResult, error)

	// SyncBidirectional performs bidirectional sync
	SyncBidirectional(ctx context.Context, localData *shared.SyncDataSet) (*SyncResult, error)
}

// SyncManagerDependencies defines the required dependencies for the sync manager
type SyncManagerDependencies struct {
	ProjectManager manager.ProjectManager
}
