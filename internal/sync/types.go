// Package sync provides bidirectional synchronization capabilities between
// local CLI SQLite workspaces and MCP PostgreSQL database.
package sync

import (
	"time"

	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/google/uuid"
)

// SyncEntity represents a synchronizable entity with metadata
type SyncEntity struct {
	ID           uuid.UUID   `json:"id"`
	Type         string      `json:"type"` // "project" or "task"
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	LocalExists  bool        `json:"local_exists"`
	RemoteExists bool        `json:"remote_exists"`
	Data         interface{} `json:"data,omitempty"`
}

// SyncPlan represents a plan of synchronization operations
type SyncPlan struct {
	Direction  shared.SyncDirection   `json:"direction"`
	ProjectID  uuid.UUID              `json:"project_id"`
	Operations []shared.SyncOperation `json:"operations"`
	Entities   []*SyncEntity          `json:"entities"`
	Conflicts  []*shared.SyncConflict `json:"conflicts"`
	CreatedAt  time.Time              `json:"created_at"`
	LastSyncAt *time.Time             `json:"last_sync_at,omitempty"`
}

// SyncResult represents the result of a synchronization operation
type SyncResult struct {
	Plan      *SyncPlan              `json:"plan"`
	Success   bool                   `json:"success"`
	Processed int                    `json:"processed"`
	Created   int                    `json:"created"`
	Updated   int                    `json:"updated"`
	Deleted   int                    `json:"deleted"`
	Conflicts []*shared.SyncConflict `json:"conflicts_resolved"`
	Errors    []string               `json:"errors,omitempty"`
	Duration  time.Duration          `json:"duration"`
	SyncedAt  time.Time              `json:"synced_at"`
}

// ConflictStrategy represents the strategy for resolving conflicts
type ConflictStrategy string

const (
	ConflictStrategyLastWriterWins ConflictStrategy = "last-writer-wins"
	ConflictStrategyPreferLocal    ConflictStrategy = "prefer-local"
	ConflictStrategyPreferRemote   ConflictStrategy = "prefer-remote"
	ConflictStrategyManual         ConflictStrategy = "manual"
)

// SyncMetrics represents metrics collected during sync operations
type SyncMetrics struct {
	StartTime         time.Time     `json:"start_time"`
	EndTime           time.Time     `json:"end_time,omitempty"`
	Duration          time.Duration `json:"duration"`
	TotalOperations   int           `json:"total_operations"`
	ProcessedCount    int           `json:"processed_count"`
	CompletedOps      int           `json:"completed_ops"`
	SuccessCount      int           `json:"success_count"`
	FailureCount      int           `json:"failure_count"`
	FailedOps         int           `json:"failed_ops"`
	SkippedOps        int           `json:"skipped_ops"`
	ConflictCount     int           `json:"conflict_count"`
	ResolvedConflicts int           `json:"resolved_conflicts"`
	RetryCount        int           `json:"retry_count"`
	BytesTransferred  int64         `json:"bytes_transferred"`
	NetworkCalls      int           `json:"network_calls"`
	Errors            []SyncError   `json:"errors,omitempty"`
}

// DiffResult represents the result of comparing local and remote data
type DiffResult struct {
	LocalVersion    int                    `json:"local_version"`
	RemoteVersion   int                    `json:"remote_version"`
	LocalTimestamp  time.Time              `json:"local_timestamp"`
	RemoteTimestamp time.Time              `json:"remote_timestamp"`
	Duration        time.Duration          `json:"duration"`
	Operations      []shared.SyncOperation `json:"operations"`
	Summary         SyncSummary            `json:"summary"`
}

// SyncSummary provides a summary of sync operations
type SyncSummary struct {
	LocalToRemote int `json:"local_to_remote"`
	RemoteToLocal int `json:"remote_to_local"`
	Conflicts     int `json:"conflicts"`
	Unchanged     int `json:"unchanged"`
	Total         int `json:"total"`
	Creations     int `json:"creations"`
	Updates       int `json:"updates"`
	Deletes       int `json:"deletes"`
	HighPriority  int `json:"high_priority"`
}

// SyncError represents an error that occurred during a sync operation
type SyncError struct {
	OperationID uuid.UUID `json:"operation_id"`
	Error       string    `json:"error"`
	Timestamp   time.Time `json:"timestamp"`
	Retryable   bool      `json:"retryable"`
}
