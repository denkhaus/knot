// Package sync provides interfaces for sync HTTP handlers
package handlers

import (
	"context"
	"net/http"

	"github.com/denkhaus/knot/v2/internal/sync/shared"
)

// SyncHTTPHandlers provides interface for sync HTTP handlers
type SyncHTTPHandlers interface {
	// Sync operation handlers
	FullSyncHandler() http.HandlerFunc
	PushSyncHandler() http.HandlerFunc
	PullSyncHandler() http.HandlerFunc
	HealthHandler() http.HandlerFunc

	// Read-only list handlers (for debugging/inspection)
	ListProjectsHandler() http.HandlerFunc
	ListTasksHandler() http.HandlerFunc

	// Middleware
	Middleware() func(http.Handler) http.Handler
}

// SyncService provides interface for sync business logic
// This separates HTTP handling from sync logic
type SyncService interface {
	// Sync operations
	PerformFullSync(ctx context.Context, request shared.SyncRequest) (*shared.SyncResponse, error)
	PerformPushSync(ctx context.Context, request shared.SyncRequest) (*shared.SyncResponse, error)
	PerformPullSync(ctx context.Context, request shared.SyncRequest) (*shared.SyncResponse, error)

	// Health operations
	HealthCheck(ctx context.Context) error
}
