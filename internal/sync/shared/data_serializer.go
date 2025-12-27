// Package shared provides common interfaces used across sync subpackages
package shared

import (
	"time"

	"github.com/google/uuid"
)

// SyncDataSerializer provides interface for serializing and validating sync data
type SyncDataSerializer interface {
	// Request serialization
	SerializeRequest(req *SyncRequest) ([]byte, error)
	DeserializeRequest(data []byte) (*SyncRequest, error)

	// Response serialization
	SerializeResponse(resp *SyncResponse) ([]byte, error)
	DeserializeResponse(data []byte) (*SyncResponse, error)

	// Error serialization
	SerializeErrorResponse(errResp *ErrorResponse) ([]byte, error)

	// Validation
	ValidateRequest(req *SyncRequest) ValidationResult
	ValidateResponse(resp *SyncResponse) ValidationResult
}

// SyncRequest represents a sync request payload for HTTP transport
type SyncRequest struct {
	ProjectID    uuid.UUID     `json:"project_id"`
	Direction    SyncDirection `json:"direction"`
	SinceVersion *int64        `json:"since_version,omitempty"`
	LocalData    *SyncDataSet  `json:"local_data,omitempty"`
	RequestID    uuid.UUID     `json:"request_id"`
	Timestamp    time.Time     `json:"timestamp"`
}

// SyncResponse represents a sync response payload for HTTP transport
type SyncResponse struct {
	Success       bool           `json:"success"`
	RequestID     uuid.UUID      `json:"request_id"`
	Processed     int            `json:"processed"`
	Created       int            `json:"created"`
	Updated       int            `json:"updated"`
	Deleted       int            `json:"deleted"`
	RemoteChanges *SyncDataSet   `json:"remote_changes,omitempty"`
	Conflicts     []SyncConflict `json:"conflicts,omitempty"`
	NewVersion    int64          `json:"new_version"`
	Timestamp     time.Time      `json:"timestamp"`
	Duration      time.Duration  `json:"duration"`
	Errors        []string       `json:"errors,omitempty"`
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error      string    `json:"error"`
	RequestID  uuid.UUID `json:"request_id"`
	Timestamp  time.Time `json:"timestamp"`
	Details    string    `json:"details,omitempty"`
	Retryable  bool      `json:"retryable"`
	Suggestion string    `json:"suggestion,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

func (e ValidationError) Error() string {
	return "validation error on field '" + e.Field + "': " + e.Message
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// AddError adds a validation error to the result
func (r *ValidationResult) AddError(field, message string, value interface{}) {
	r.Valid = false
	r.Errors = append(r.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// Validate performs validation on the request
func (r *SyncRequest) Validate() *ValidationResult {
	result := &ValidationResult{Valid: true}

	if r.ProjectID == uuid.Nil {
		result.AddError("project_id", "project_id cannot be empty UUID", r.ProjectID)
	}

	if r.RequestID == uuid.Nil {
		result.AddError("request_id", "request_id cannot be empty UUID", r.RequestID)
	}

	if r.Timestamp.IsZero() {
		result.AddError("timestamp", "timestamp cannot be zero", r.Timestamp)
	}

	// Validate direction
	validDirections := map[SyncDirection]bool{
		SyncLocalToMCP:    true,
		SyncMcpToLocal:    true,
		SyncBidirectional: true,
	}
	if !validDirections[r.Direction] {
		result.AddError("direction", "invalid sync direction", r.Direction)
	}

	return result
}

// Validate performs validation on the response
func (r *SyncResponse) Validate() *ValidationResult {
	result := &ValidationResult{Valid: true}

	if r.RequestID == uuid.Nil {
		result.AddError("request_id", "request_id cannot be empty UUID", r.RequestID)
	}

	if r.Timestamp.IsZero() {
		result.AddError("timestamp", "timestamp cannot be zero", r.Timestamp)
	}

	if r.Processed < 0 {
		result.AddError("processed", "processed count cannot be negative", r.Processed)
	}

	if r.Created < 0 {
		result.AddError("created", "created count cannot be negative", r.Created)
	}

	if r.Updated < 0 {
		result.AddError("updated", "updated count cannot be negative", r.Updated)
	}

	if r.Deleted < 0 {
		result.AddError("deleted", "deleted count cannot be negative", r.Deleted)
	}

	return result
}

// Validate performs validation on the error response
func (e *ErrorResponse) Validate() *ValidationResult {
	result := &ValidationResult{Valid: true}

	if e.Error == "" {
		result.AddError("error", "error message cannot be empty", e.Error)
	}

	if e.RequestID == uuid.Nil {
		result.AddError("request_id", "request_id cannot be empty UUID", e.RequestID)
	}

	if e.Timestamp.IsZero() {
		result.AddError("timestamp", "timestamp cannot be zero", e.Timestamp)
	}

	return result
}
