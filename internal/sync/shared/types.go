// Package shared provides common types used across sync subpackages
package shared

import (
	"time"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
)

// SyncDataSet represents a collection of synchronized data
type SyncDataSet struct {
	Projects map[uuid.UUID]*types.Project `json:"projects"`
	Tasks    map[uuid.UUID]*types.Task    `json:"tasks"`
}

// NewSyncDataSet creates a new empty SyncDataSet
func NewSyncDataSet() *SyncDataSet {
	return &SyncDataSet{
		Projects: make(map[uuid.UUID]*types.Project),
		Tasks:    make(map[uuid.UUID]*types.Task),
	}
}

// SyncConflict represents a synchronization conflict
type SyncConflict struct {
	ID           uuid.UUID           `json:"id"`
	EntityID     uuid.UUID           `json:"entity_id"`
	EntityType   SyncEntityType      `json:"entity_type"`
	LocalData    interface{}         `json:"local_data"`
	RemoteData   interface{}         `json:"remote_data"`
	ConflictType ConflictType        `json:"conflict_type"`
	Resolution   *ConflictResolution `json:"resolution,omitempty"`
	Operations   []SyncOperation     `json:"operations"`
	CreatedAt    time.Time           `json:"created_at"`
	ResolvedAt   *time.Time          `json:"resolved_at,omitempty"`
}

// SyncDirection represents the direction of synchronization
type SyncDirection string

const (
	SyncLocalToMCP    SyncDirection = "local_to_mcp"
	SyncMcpToLocal    SyncDirection = "mcp_to_local"
	SyncBidirectional SyncDirection = "bidirectional"
)

// SyncOpType represents the type of sync operation
type SyncOpType string

const (
	OpCreate   SyncOpType = "create"
	OpUpdate   SyncOpType = "update"
	OpDelete   SyncOpType = "delete"
	OpConflict SyncOpType = "conflict"
)

// SyncEntityType represents the type of entity being synchronized
type SyncEntityType string

const (
	EntityProject SyncEntityType = "project"
	EntityTask    SyncEntityType = "task"
)

// SyncOpStatus represents the status of a sync operation
type SyncOpStatus string

const (
	StatusPending    SyncOpStatus = "pending"
	StatusInProgress SyncOpStatus = "in_progress"
	StatusCompleted  SyncOpStatus = "completed"
	StatusFailed     SyncOpStatus = "failed"
	StatusSkipped    SyncOpStatus = "skipped"
)

// ConflictType represents the type of conflict
type ConflictType string

const (
	ConflictTypeUpdate     ConflictType = "update"
	ConflictTypeDelete     ConflictType = "delete"
	ConflictTypeState      ConflictType = "state"
	ConflictTypeDependency ConflictType = "dependency"
)

// ConflictResolution represents how a conflict was resolved
type ConflictResolution struct {
	Strategy   string      `json:"strategy"`        // "prefer-local", "prefer-remote", "manual", etc.
	ResolvedBy string      `json:"resolved_by"`     // Who or what resolved the conflict
	Actor      string      `json:"actor,omitempty"` // Alias for ResolvedBy for backward compatibility
	ResolvedAt time.Time   `json:"resolved_at"`
	Timestamp  time.Time   `json:"timestamp"`             // Alias for ResolvedAt for backward compatibility
	Reason     string      `json:"reason,omitempty"`      // Why this resolution was chosen
	ChosenData interface{} `json:"chosen_data,omitempty"` // The data that was selected
	MergedData interface{} `json:"merged_data,omitempty"` // If data was merged
}

// DataStatistics holds strongly-typed statistics about a data set
type DataStatistics struct {
	TotalProjects     int                          `json:"total_projects"`
	TotalTasks        int                          `json:"total_tasks"`
	TasksByState      map[types.TaskState]int      `json:"tasks_by_state"`
	ProjectsByState   map[types.ProjectState]int   `json:"projects_by_state"`
	OldestProject     *time.Time                   `json:"oldest_project,omitempty"`
	NewestProject     *time.Time                   `json:"newest_project,omitempty"`
	ProjectAgeSpanDays int                         `json:"project_age_span_days,omitempty"`
	OldestTask        *time.Time                   `json:"oldest_task,omitempty"`
	NewestTask        *time.Time                   `json:"newest_task,omitempty"`
	TaskAgeSpanDays   int                          `json:"task_age_span_days,omitempty"`
}

// SyncOperation represents a single synchronization operation
type SyncOperation struct {
	ID           uuid.UUID      `json:"id"`
	EntityID     uuid.UUID      `json:"entity_id"`
	EntityType   SyncEntityType `json:"entity_type"`
	OpType       SyncOpType     `json:"op_type"`
	Type         SyncOpType     `json:"type"` // Alias for OpType for backward compatibility
	Direction    SyncDirection  `json:"direction"`
	Status       SyncOpStatus   `json:"status"`
	LocalData    interface{}    `json:"local_data,omitempty"`
	RemoteData   interface{}    `json:"remote_data,omitempty"`
	ResultData   interface{}    `json:"result_data,omitempty"`
	Error        string         `json:"error,omitempty"`
	Reason       string         `json:"reason,omitempty"`
	Priority     int            `json:"priority,omitempty"`
	Dependencies []string       `json:"dependencies,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	StartedAt    *time.Time     `json:"started_at,omitempty"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
}
