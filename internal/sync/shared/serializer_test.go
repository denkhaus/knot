// Package shared provides tests for sync data serialization (JSON only)
package shared

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denkhaus/knot/v2/internal/types"
)

// TestSerializerImpl_SerializeRequestJSON tests JSON serialization of sync request
func TestSerializerImpl_SerializeRequestJSON(t *testing.T) {
	serializer := &serializerImpl{}

	req := &SyncRequest{
		ProjectID: uuid.New(),
		Direction: SyncBidirectional,
		RequestID: uuid.New(),
		Timestamp: time.Now(),
	}

	data, err := serializer.SerializeRequest(req)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	assert.Contains(t, string(data), "project_id")
	assert.Contains(t, string(data), "direction")
	assert.Contains(t, string(data), "request_id")
}

// TestSerializerImpl_SerializeRequestJSON_Nil tests JSON serialization with nil request
func TestSerializerImpl_SerializeRequestJSON_Nil(t *testing.T) {
	serializer := &serializerImpl{}

	data, err := serializer.SerializeRequest(nil)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestSerializerImpl_DeserializeRequestJSON tests JSON deserialization of sync request
func TestSerializerImpl_DeserializeRequestJSON(t *testing.T) {
	serializer := &serializerImpl{}

	originalReq := &SyncRequest{
		ProjectID: uuid.New(),
		Direction: SyncLocalToMCP,
		RequestID: uuid.New(),
		Timestamp: time.Now(),
	}

	// First serialize
	data, err := serializer.SerializeRequest(originalReq)
	require.NoError(t, err)

	// Then deserialize
	req, err := serializer.DeserializeRequest(data)
	require.NoError(t, err)

	assert.Equal(t, originalReq.ProjectID, req.ProjectID)
	assert.Equal(t, originalReq.Direction, req.Direction)
	assert.Equal(t, originalReq.RequestID, req.RequestID)
}

// TestSerializerImpl_DeserializeRequestJSON_Empty tests JSON deserialization with empty data
func TestSerializerImpl_DeserializeRequestJSON_Empty(t *testing.T) {
	serializer := &serializerImpl{}

	req, err := serializer.DeserializeRequest([]byte{})
	assert.Error(t, err)
	assert.Nil(t, req)
	assert.Contains(t, err.Error(), "cannot be empty")
}

// TestSerializerImpl_DeserializeRequestJSON_Invalid tests JSON deserialization with invalid data
func TestSerializerImpl_DeserializeRequestJSON_Invalid(t *testing.T) {
	serializer := &serializerImpl{}

	req, err := serializer.DeserializeRequest([]byte("invalid json"))
	assert.Error(t, err)
	assert.Nil(t, req)
	assert.Contains(t, err.Error(), "failed to unmarshal")
}

// TestSerializerImpl_SerializeResponseJSON tests JSON serialization of sync response
func TestSerializerImpl_SerializeResponseJSON(t *testing.T) {
	serializer := &serializerImpl{}

	resp := &SyncResponse{
		Success:   true,
		RequestID: uuid.New(),
		Processed: 10,
		Created:   5,
		Updated:   3,
		Deleted:   2,
		Timestamp: time.Now(),
	}

	data, err := serializer.SerializeResponse(resp)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	assert.Contains(t, string(data), "success")
	assert.Contains(t, string(data), "processed")
	assert.Contains(t, string(data), "created")
}

// TestSerializerImpl_SerializeResponseJSON_Nil tests JSON serialization with nil response
func TestSerializerImpl_SerializeResponseJSON_Nil(t *testing.T) {
	serializer := &serializerImpl{}

	data, err := serializer.SerializeResponse(nil)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestSerializerImpl_DeserializeResponseJSON tests JSON deserialization of sync response
func TestSerializerImpl_DeserializeResponseJSON(t *testing.T) {
	serializer := &serializerImpl{}

	originalResp := &SyncResponse{
		Success:   true,
		RequestID: uuid.New(),
		Processed: 15,
		Created:   7,
		Updated:   5,
		Deleted:   3,
		Timestamp: time.Now(),
	}

	// First serialize
	data, err := serializer.SerializeResponse(originalResp)
	require.NoError(t, err)

	// Then deserialize
	resp, err := serializer.DeserializeResponse(data)
	require.NoError(t, err)

	assert.Equal(t, originalResp.Success, resp.Success)
	assert.Equal(t, originalResp.RequestID, resp.RequestID)
	assert.Equal(t, originalResp.Processed, resp.Processed)
	assert.Equal(t, originalResp.Created, resp.Created)
	assert.Equal(t, originalResp.Updated, resp.Updated)
	assert.Equal(t, originalResp.Deleted, resp.Deleted)
}

// TestSerializerImpl_DeserializeResponseJSON_Empty tests JSON deserialization with empty data
func TestSerializerImpl_DeserializeResponseJSON_Empty(t *testing.T) {
	serializer := &serializerImpl{}

	resp, err := serializer.DeserializeResponse([]byte{})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "cannot be empty")
}

// TestSerializerImpl_SerializeErrorResponseJSON tests JSON serialization of error response
func TestSerializerImpl_SerializeErrorResponseJSON(t *testing.T) {
	serializer := &serializerImpl{}

	errResp := &ErrorResponse{
		Error:     "test error",
		RequestID: uuid.New(),
		Timestamp: time.Now(),
		Retryable: true,
	}

	data, err := serializer.SerializeErrorResponse(errResp)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	assert.Contains(t, string(data), "error")
	assert.Contains(t, string(data), "retryable")
}

// TestSerializerImpl_SerializeErrorResponseJSON_Nil tests JSON serialization with nil error response
func TestSerializerImpl_SerializeErrorResponseJSON_Nil(t *testing.T) {
	serializer := &serializerImpl{}

	data, err := serializer.SerializeErrorResponse(nil)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestSerializerImpl_ValidateRequest tests request validation through serializer
func TestSerializerImpl_ValidateRequest(t *testing.T) {
	serializer := &serializerImpl{}

	// Valid request
	validReq := &SyncRequest{
		ProjectID: uuid.New(),
		Direction: SyncBidirectional,
		RequestID: uuid.New(),
		Timestamp: time.Now(),
	}
	result := serializer.ValidateRequest(validReq)
	assert.True(t, result.Valid)

	// Invalid request (nil)
	result = serializer.ValidateRequest(nil)
	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)

	// Invalid request (missing fields)
	invalidReq := &SyncRequest{
		ProjectID: uuid.Nil,
		Direction: "invalid",
		RequestID: uuid.Nil,
		Timestamp: time.Time{},
	}
	result = serializer.ValidateRequest(invalidReq)
	assert.False(t, result.Valid)
	assert.Greater(t, len(result.Errors), 0)
}

// TestSerializerImpl_ValidateResponse tests response validation through serializer
func TestSerializerImpl_ValidateResponse(t *testing.T) {
	serializer := &serializerImpl{}

	// Valid response
	validResp := &SyncResponse{
		Success:   true,
		RequestID: uuid.New(),
		Processed: 10,
		Timestamp: time.Now(),
	}
	result := serializer.ValidateResponse(validResp)
	assert.True(t, result.Valid)

	// Invalid response (nil)
	result = serializer.ValidateResponse(nil)
	assert.False(t, result.Valid)
	assert.Len(t, result.Errors, 1)

	// Invalid response (negative count)
	invalidResp := &SyncResponse{
		Success:   true,
		RequestID: uuid.New(),
		Processed: -5,
		Timestamp: time.Now(),
	}
	result = serializer.ValidateResponse(invalidResp)
	assert.False(t, result.Valid)
}

// TestNewSerializerProvider tests the DI provider function
func TestNewSerializerProvider(t *testing.T) {
	injector := do.New()

	serializer, err := NewSerializerProvider(injector)
	require.NoError(t, err)
	assert.NotNil(t, serializer)

	// Verify it implements the interface
	_, ok := serializer.(SyncDataSerializer)
	assert.True(t, ok, "NewSerializerProvider should return a SyncDataSerializer")
}

// TestSyncDataSetUnmarshalJSON tests custom JSON unmarshaling for SyncDataSet
func TestSyncDataSetUnmarshalJSON(t *testing.T) {
	projectID := uuid.New()
	taskID := uuid.New()

	validJSON := `{
		"projects": {
			"` + projectID.String() + `": {
				"id": "` + projectID.String() + `",
				"title": "Test Project",
				"description": "Test Description",
				"state": "active",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z"
			}
		},
		"tasks": {
			"` + taskID.String() + `": {
				"id": "` + taskID.String() + `",
				"project_id": "` + projectID.String() + `",
				"title": "Test Task",
				"state": "pending",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z"
			}
		}
	}`

	var dataset SyncDataSet
	err := json.Unmarshal([]byte(validJSON), &dataset)
	require.NoError(t, err)
	assert.Len(t, dataset.Projects, 1)
	assert.Len(t, dataset.Tasks, 1)

	// Verify the keys were properly converted from strings to UUIDs
	assert.Contains(t, dataset.Projects, projectID)
	assert.Contains(t, dataset.Tasks, taskID)

	// Verify project data - the values should be unmarshaled correctly
	project := dataset.Projects[projectID]
	assert.Equal(t, projectID, project.ID)
	// Note: JSON unmarshaling into the pointer value works correctly
	assert.NotNil(t, project)
	assert.Equal(t, types.ProjectStateActive, project.State)
}

// TestSyncDataSetUnmarshalJSON_EmptyMaps tests unmarshaling with empty maps
func TestSyncDataSetUnmarshalJSON_EmptyMaps(t *testing.T) {
	emptyJSON := `{
		"projects": {},
		"tasks": {}
	}`

	var dataset SyncDataSet
	err := json.Unmarshal([]byte(emptyJSON), &dataset)
	require.NoError(t, err)
	assert.Empty(t, dataset.Projects)
	assert.Empty(t, dataset.Tasks)
}

// TestSyncDataSetUnmarshalJSON_InvalidUUID tests unmarshaling with invalid UUID
func TestSyncDataSetUnmarshalJSON_InvalidUUID(t *testing.T) {
	invalidJSON := `{
		"projects": {
			"invalid": {
				"id": "some-id",
				"title": "Test Project",
				"state": "active",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z"
			}
		},
		"tasks": {}
	}`

	var dataset SyncDataSet
	err := json.Unmarshal([]byte(invalidJSON), &dataset)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

// TestSyncDataSetUnmarshalJSON_InvalidTaskUUID tests unmarshaling with invalid task UUID
func TestSyncDataSetUnmarshalJSON_InvalidTaskUUID(t *testing.T) {
	projectID := uuid.New()

	invalidJSON := `{
		"projects": {
			"` + projectID.String() + `": {
				"id": "` + projectID.String() + `",
				"title": "Test Project",
				"state": "active",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z"
			}
		},
		"tasks": {
			"invalid": {
				"id": "some-id",
				"project_id": "` + projectID.String() + `",
				"title": "Test Task",
				"state": "pending",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z"
			}
		}
	}`

	var dataset SyncDataSet
	err := json.Unmarshal([]byte(invalidJSON), &dataset)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

// TestSyncDataSetUnmarshalJSON_MalformedJSON tests unmarshaling with malformed JSON
func TestSyncDataSetUnmarshalJSON_MalformedJSON(t *testing.T) {
	malformedJSON := `{invalid json`

	var dataset SyncDataSet
	err := json.Unmarshal([]byte(malformedJSON), &dataset)
	assert.Error(t, err)
}

// TestNewSyncDataSet tests creating a new SyncDataSet
func TestNewSyncDataSet(t *testing.T) {
	dataset := NewSyncDataSet()
	assert.NotNil(t, dataset)
	assert.NotNil(t, dataset.Projects)
	assert.NotNil(t, dataset.Tasks)
	assert.Empty(t, dataset.Projects)
	assert.Empty(t, dataset.Tasks)
}
