// Package shared provides tests for sync data validation
package shared

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestSyncRequest_Validate_Valid tests validation of valid sync request
func TestSyncRequest_Validate_Valid(t *testing.T) {
	req := &SyncRequest{
		ProjectID: uuid.New(),
		Direction: SyncBidirectional,
		RequestID: uuid.New(),
		Timestamp: time.Now(),
	}

	result := req.Validate()

	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

// TestSyncRequest_Validate_EmptyProjectID tests validation with empty project ID
func TestSyncRequest_Validate_EmptyProjectID(t *testing.T) {
	req := &SyncRequest{
		ProjectID: uuid.Nil,
		Direction: SyncBidirectional,
		RequestID: uuid.New(),
		Timestamp: time.Now(),
	}

	result := req.Validate()

	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "project_id", result.Errors[0].Field)
	assert.Contains(t, result.Errors[0].Message, "cannot be empty")
}

// TestSyncRequest_Validate_EmptyRequestID tests validation with empty request ID
func TestSyncRequest_Validate_EmptyRequestID(t *testing.T) {
	req := &SyncRequest{
		ProjectID: uuid.New(),
		Direction: SyncBidirectional,
		RequestID: uuid.Nil,
		Timestamp: time.Now(),
	}

	result := req.Validate()

	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "request_id", result.Errors[0].Field)
}

// TestSyncRequest_Validate_ZeroTimestamp tests validation with zero timestamp
func TestSyncRequest_Validate_ZeroTimestamp(t *testing.T) {
	req := &SyncRequest{
		ProjectID: uuid.New(),
		Direction: SyncBidirectional,
		RequestID: uuid.New(),
		Timestamp: time.Time{},
	}

	result := req.Validate()

	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "timestamp", result.Errors[0].Field)
	assert.Contains(t, result.Errors[0].Message, "cannot be zero")
}

// TestSyncRequest_Validate_InvalidDirection tests validation with invalid direction
func TestSyncRequest_Validate_InvalidDirection(t *testing.T) {
	req := &SyncRequest{
		ProjectID: uuid.New(),
		Direction: SyncDirection("invalid"),
		RequestID: uuid.New(),
		Timestamp: time.Now(),
	}

	result := req.Validate()

	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "direction", result.Errors[0].Field)
	assert.Contains(t, result.Errors[0].Message, "invalid sync direction")
}

// TestSyncRequest_Validate_MultipleErrors tests validation with multiple errors
func TestSyncRequest_Validate_MultipleErrors(t *testing.T) {
	req := &SyncRequest{
		ProjectID: uuid.Nil,
		Direction: SyncDirection("invalid"),
		RequestID: uuid.Nil,
		Timestamp: time.Time{},
	}

	result := req.Validate()

	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 4) // project_id, direction, request_id, timestamp
}

// TestSyncRequest_Validate_AllValidDirections tests all valid sync directions
func TestSyncRequest_Validate_AllValidDirections(t *testing.T) {
	validDirections := []SyncDirection{
		SyncLocalToMCP,
		SyncMcpToLocal,
		SyncBidirectional,
	}

	for _, direction := range validDirections {
		req := &SyncRequest{
			ProjectID: uuid.New(),
			Direction: direction,
			RequestID: uuid.New(),
			Timestamp: time.Now(),
		}

		result := req.Validate()
		assert.True(t, result.Valid, "Direction %s should be valid", direction)
	}
}

// TestSyncResponse_Validate_Valid tests validation of valid sync response
func TestSyncResponse_Validate_Valid(t *testing.T) {
	resp := &SyncResponse{
		Success:   true,
		RequestID: uuid.New(),
		Processed: 10,
		Created:   5,
		Updated:   3,
		Deleted:   2,
		Timestamp: time.Now(),
	}

	result := resp.Validate()

	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

// TestSyncResponse_Validate_EmptyRequestID tests validation with empty request ID
func TestSyncResponse_Validate_EmptyRequestID(t *testing.T) {
	resp := &SyncResponse{
		Success:   true,
		RequestID: uuid.Nil,
		Timestamp: time.Now(),
	}

	result := resp.Validate()

	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "request_id", result.Errors[0].Field)
}

// TestSyncResponse_Validate_ZeroTimestamp tests validation with zero timestamp
func TestSyncResponse_Validate_ZeroTimestamp(t *testing.T) {
	resp := &SyncResponse{
		Success:   true,
		RequestID: uuid.New(),
		Timestamp: time.Time{},
	}

	result := resp.Validate()

	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "timestamp", result.Errors[0].Field)
}

// TestSyncResponse_Validate_NegativeProcessed tests validation with negative processed count
func TestSyncResponse_Validate_NegativeProcessed(t *testing.T) {
	resp := &SyncResponse{
		Success:   true,
		RequestID: uuid.New(),
		Processed: -1,
		Timestamp: time.Now(),
	}

	result := resp.Validate()

	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "processed", result.Errors[0].Field)
	assert.Contains(t, result.Errors[0].Message, "cannot be negative")
}

// TestSyncResponse_Validate_NegativeCreated tests validation with negative created count
func TestSyncResponse_Validate_NegativeCreated(t *testing.T) {
	resp := &SyncResponse{
		Success:   true,
		RequestID: uuid.New(),
		Created:   -1,
		Timestamp: time.Now(),
	}

	result := resp.Validate()

	assert.False(t, result.Valid)
	assert.Equal(t, "created", result.Errors[0].Field)
}

// TestSyncResponse_Validate_NegativeUpdated tests validation with negative updated count
func TestSyncResponse_Validate_NegativeUpdated(t *testing.T) {
	resp := &SyncResponse{
		Success:   true,
		RequestID: uuid.New(),
		Updated:   -1,
		Timestamp: time.Now(),
	}

	result := resp.Validate()

	assert.False(t, result.Valid)
	assert.Equal(t, "updated", result.Errors[0].Field)
}

// TestSyncResponse_Validate_NegativeDeleted tests validation with negative deleted count
func TestSyncResponse_Validate_NegativeDeleted(t *testing.T) {
	resp := &SyncResponse{
		Success:   true,
		RequestID: uuid.New(),
		Deleted:   -1,
		Timestamp: time.Now(),
	}

	result := resp.Validate()

	assert.False(t, result.Valid)
	assert.Equal(t, "deleted", result.Errors[0].Field)
}

// TestSyncResponse_Validate_AllZeroCounts tests validation with all zero counts (valid)
func TestSyncResponse_Validate_AllZeroCounts(t *testing.T) {
	resp := &SyncResponse{
		Success:   true,
		RequestID: uuid.New(),
		Processed: 0,
		Created:   0,
		Updated:   0,
		Deleted:   0,
		Timestamp: time.Now(),
	}

	result := resp.Validate()

	assert.True(t, result.Valid)
}

// TestErrorResponse_Validate_Valid tests validation of valid error response
func TestErrorResponse_Validate_Valid(t *testing.T) {
	errResp := &ErrorResponse{
		Error:     "something went wrong",
		RequestID: uuid.New(),
		Timestamp: time.Now(),
		Retryable: false,
	}

	result := errResp.Validate()

	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

// TestErrorResponse_Validate_EmptyError tests validation with empty error message
func TestErrorResponse_Validate_EmptyError(t *testing.T) {
	errResp := &ErrorResponse{
		Error:     "",
		RequestID: uuid.New(),
		Timestamp: time.Now(),
	}

	result := errResp.Validate()

	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "error", result.Errors[0].Field)
	assert.Contains(t, result.Errors[0].Message, "cannot be empty")
}

// TestErrorResponse_Validate_EmptyRequestID tests validation with empty request ID
func TestErrorResponse_Validate_EmptyRequestID(t *testing.T) {
	errResp := &ErrorResponse{
		Error:     "test error",
		RequestID: uuid.Nil,
		Timestamp: time.Now(),
	}

	result := errResp.Validate()

	assert.False(t, result.Valid)
	assert.Equal(t, "request_id", result.Errors[0].Field)
}

// TestErrorResponse_Validate_ZeroTimestamp tests validation with zero timestamp
func TestErrorResponse_Validate_ZeroTimestamp(t *testing.T) {
	errResp := &ErrorResponse{
		Error:     "test error",
		RequestID: uuid.New(),
		Timestamp: time.Time{},
	}

	result := errResp.Validate()

	assert.False(t, result.Valid)
	assert.Equal(t, "timestamp", result.Errors[0].Field)
}

// TestValidationResult_AddError tests adding errors to validation result
func TestValidationResult_AddError(t *testing.T) {
	result := &ValidationResult{Valid: true}

	result.AddError("field1", "error message 1", "value1")
	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "field1", result.Errors[0].Field)
	assert.Equal(t, "error message 1", result.Errors[0].Message)
	assert.Equal(t, "value1", result.Errors[0].Value)

	result.AddError("field2", "error message 2", nil)
	assert.Len(t, result.Errors, 2)
	assert.Equal(t, "field2", result.Errors[1].Field)
}

// TestValidationError_Error tests the Error method of ValidationError
func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Field:   "test_field",
		Message: "is invalid",
		Value:   "bad_value",
	}

	expected := "validation error on field 'test_field': is invalid"
	assert.Equal(t, expected, err.Error())
}

// TestValidationError_Error_NoValue tests Error method without value
func TestValidationError_Error_NoValue(t *testing.T) {
	err := ValidationError{
		Field:   "test_field",
		Message: "is required",
	}

	expected := "validation error on field 'test_field': is required"
	assert.Equal(t, expected, err.Error())
}
