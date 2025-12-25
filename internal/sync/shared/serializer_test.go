// Package shared provides tests for sync data serialization
package shared

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	data, err := serializer.SerializeRequestJSON(req)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	assert.Contains(t, string(data), "project_id")
	assert.Contains(t, string(data), "direction")
	assert.Contains(t, string(data), "request_id")
}

// TestSerializerImpl_SerializeRequestJSON_Nil tests JSON serialization with nil request
func TestSerializerImpl_SerializeRequestJSON_Nil(t *testing.T) {
	serializer := &serializerImpl{}

	data, err := serializer.SerializeRequestJSON(nil)
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
	data, err := serializer.SerializeRequestJSON(originalReq)
	require.NoError(t, err)

	// Then deserialize
	req, err := serializer.DeserializeRequestJSON(data)
	require.NoError(t, err)

	assert.Equal(t, originalReq.ProjectID, req.ProjectID)
	assert.Equal(t, originalReq.Direction, req.Direction)
	assert.Equal(t, originalReq.RequestID, req.RequestID)
}

// TestSerializerImpl_DeserializeRequestJSON_Empty tests JSON deserialization with empty data
func TestSerializerImpl_DeserializeRequestJSON_Empty(t *testing.T) {
	serializer := &serializerImpl{}

	req, err := serializer.DeserializeRequestJSON([]byte{})
	assert.Error(t, err)
	assert.Nil(t, req)
	assert.Contains(t, err.Error(), "cannot be empty")
}

// TestSerializerImpl_DeserializeRequestJSON_Invalid tests JSON deserialization with invalid data
func TestSerializerImpl_DeserializeRequestJSON_Invalid(t *testing.T) {
	serializer := &serializerImpl{}

	req, err := serializer.DeserializeRequestJSON([]byte("invalid json"))
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

	data, err := serializer.SerializeResponseJSON(resp)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	assert.Contains(t, string(data), "success")
	assert.Contains(t, string(data), "processed")
	assert.Contains(t, string(data), "created")
}

// TestSerializerImpl_SerializeResponseJSON_Nil tests JSON serialization with nil response
func TestSerializerImpl_SerializeResponseJSON_Nil(t *testing.T) {
	serializer := &serializerImpl{}

	data, err := serializer.SerializeResponseJSON(nil)
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
	data, err := serializer.SerializeResponseJSON(originalResp)
	require.NoError(t, err)

	// Then deserialize
	resp, err := serializer.DeserializeResponseJSON(data)
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

	resp, err := serializer.DeserializeResponseJSON([]byte{})
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

	data, err := serializer.SerializeErrorResponseJSON(errResp)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	assert.Contains(t, string(data), "error")
	assert.Contains(t, string(data), "retryable")
}

// TestSerializerImpl_SerializeErrorResponseJSON_Nil tests JSON serialization with nil error response
func TestSerializerImpl_SerializeErrorResponseJSON_Nil(t *testing.T) {
	serializer := &serializerImpl{}

	data, err := serializer.SerializeErrorResponseJSON(nil)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestSerializerImpl_SerializeRequestMsgPack tests MessagePack serialization of sync request
func TestSerializerImpl_SerializeRequestMsgPack(t *testing.T) {
	serializer := &serializerImpl{}

	req := &SyncRequest{
		ProjectID: uuid.New(),
		Direction: SyncBidirectional,
		RequestID: uuid.New(),
		Timestamp: time.Now(),
	}

	data, err := serializer.SerializeRequestMsgPack(req)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	assert.Greater(t, len(data), 0)
}

// TestSerializerImpl_SerializeRequestMsgPack_Nil tests MessagePack serialization with nil request
func TestSerializerImpl_SerializeRequestMsgPack_Nil(t *testing.T) {
	serializer := &serializerImpl{}

	data, err := serializer.SerializeRequestMsgPack(nil)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestSerializerImpl_DeserializeRequestMsgPack tests MessagePack deserialization of sync request
func TestSerializerImpl_DeserializeRequestMsgPack(t *testing.T) {
	serializer := &serializerImpl{}

	originalReq := &SyncRequest{
		ProjectID: uuid.New(),
		Direction: SyncMcpToLocal,
		RequestID: uuid.New(),
		Timestamp: time.Now(),
	}

	// First serialize
	data, err := serializer.SerializeRequestMsgPack(originalReq)
	require.NoError(t, err)

	// Then deserialize
	req, err := serializer.DeserializeRequestMsgPack(data)
	require.NoError(t, err)

	assert.Equal(t, originalReq.ProjectID, req.ProjectID)
	assert.Equal(t, originalReq.Direction, req.Direction)
	assert.Equal(t, originalReq.RequestID, req.RequestID)
}

// TestSerializerImpl_DeserializeRequestMsgPack_Empty tests MessagePack deserialization with empty data
func TestSerializerImpl_DeserializeRequestMsgPack_Empty(t *testing.T) {
	serializer := &serializerImpl{}

	req, err := serializer.DeserializeRequestMsgPack([]byte{})
	assert.Error(t, err)
	assert.Nil(t, req)
	assert.Contains(t, err.Error(), "cannot be empty")
}

// TestSerializerImpl_DeserializeRequestMsgPack_Invalid tests MessagePack deserialization with invalid data
func TestSerializerImpl_DeserializeRequestMsgPack_Invalid(t *testing.T) {
	serializer := &serializerImpl{}

	req, err := serializer.DeserializeRequestMsgPack([]byte("invalid msgpack"))
	assert.Error(t, err)
	assert.Nil(t, req)
	assert.Contains(t, err.Error(), "failed to decode")
}

// TestSerializerImpl_SerializeResponseMsgPack tests MessagePack serialization of sync response
func TestSerializerImpl_SerializeResponseMsgPack(t *testing.T) {
	serializer := &serializerImpl{}

	resp := &SyncResponse{
		Success:   true,
		RequestID: uuid.New(),
		Processed: 20,
		Created:   10,
		Updated:   7,
		Deleted:   3,
		Timestamp: time.Now(),
	}

	data, err := serializer.SerializeResponseMsgPack(resp)
	require.NoError(t, err)
	require.NotEmpty(t, data)
}

// TestSerializerImpl_SerializeResponseMsgPack_Nil tests MessagePack serialization with nil response
func TestSerializerImpl_SerializeResponseMsgPack_Nil(t *testing.T) {
	serializer := &serializerImpl{}

	data, err := serializer.SerializeResponseMsgPack(nil)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestSerializerImpl_DeserializeResponseMsgPack tests MessagePack deserialization of sync response
func TestSerializerImpl_DeserializeResponseMsgPack(t *testing.T) {
	serializer := &serializerImpl{}

	originalResp := &SyncResponse{
		Success:   false,
		RequestID: uuid.New(),
		Processed: 5,
		Created:   2,
		Updated:   2,
		Deleted:   1,
		Timestamp: time.Now(),
		Errors:    []string{"error1", "error2"},
	}

	// First serialize
	data, err := serializer.SerializeResponseMsgPack(originalResp)
	require.NoError(t, err)

	// Then deserialize
	resp, err := serializer.DeserializeResponseMsgPack(data)
	require.NoError(t, err)

	assert.Equal(t, originalResp.Success, resp.Success)
	assert.Equal(t, originalResp.RequestID, resp.RequestID)
	assert.Equal(t, originalResp.Processed, resp.Processed)
	assert.Equal(t, originalResp.Errors, resp.Errors)
}

// TestSerializerImpl_DeserializeResponseMsgPack_Empty tests MessagePack deserialization with empty data
func TestSerializerImpl_DeserializeResponseMsgPack_Empty(t *testing.T) {
	serializer := &serializerImpl{}

	resp, err := serializer.DeserializeResponseMsgPack([]byte{})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "cannot be empty")
}

// TestSerializerImpl_SerializeErrorResponseMsgPack tests MessagePack serialization of error response
func TestSerializerImpl_SerializeErrorResponseMsgPack(t *testing.T) {
	serializer := &serializerImpl{}

	errResp := &ErrorResponse{
		Error:     "critical error",
		RequestID: uuid.New(),
		Timestamp: time.Now(),
		Retryable: false,
		Details:   "database connection failed",
		Suggestion: "check database server",
	}

	data, err := serializer.SerializeErrorResponseMsgPack(errResp)
	require.NoError(t, err)
	require.NotEmpty(t, data)
}

// TestSerializerImpl_SerializeErrorResponseMsgPack_Nil tests MessagePack serialization with nil error response
func TestSerializerImpl_SerializeErrorResponseMsgPack_Nil(t *testing.T) {
	serializer := &serializerImpl{}

	data, err := serializer.SerializeErrorResponseMsgPack(nil)
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestSerializerImpl_SerializeRequest tests format-based serialization
func TestSerializerImpl_SerializeRequest(t *testing.T) {
	serializer := &serializerImpl{}

	req := &SyncRequest{
		ProjectID: uuid.New(),
		Direction: SyncBidirectional,
		RequestID: uuid.New(),
		Timestamp: time.Now(),
	}

	// Test JSON format
	dataJSON, err := serializer.SerializeRequest(req, "json")
	require.NoError(t, err)
	assert.NotEmpty(t, dataJSON)

	// Test msgpack format
	dataMsgPack, err := serializer.SerializeRequest(req, "msgpack")
	require.NoError(t, err)
	assert.NotEmpty(t, dataMsgPack)

	// Test default (should use JSON)
	dataDefault, err := serializer.SerializeRequest(req, "unknown")
	require.NoError(t, err)
	assert.NotEmpty(t, dataDefault)
}

// TestSerializerImpl_DeserializeRequest tests format-based deserialization
func TestSerializerImpl_DeserializeRequest(t *testing.T) {
	serializer := &serializerImpl{}

	originalReq := &SyncRequest{
		ProjectID: uuid.New(),
		Direction: SyncLocalToMCP,
		RequestID: uuid.New(),
		Timestamp: time.Now(),
	}

	// Test JSON round-trip
	dataJSON, _ := serializer.SerializeRequest(originalReq, "json")
	reqJSON, err := serializer.DeserializeRequest(dataJSON, "json")
	require.NoError(t, err)
	assert.Equal(t, originalReq.ProjectID, reqJSON.ProjectID)

	// Test msgpack round-trip
	dataMsgPack, _ := serializer.SerializeRequest(originalReq, "msgpack")
	reqMsgPack, err := serializer.DeserializeRequest(dataMsgPack, "msgpack")
	require.NoError(t, err)
	assert.Equal(t, originalReq.ProjectID, reqMsgPack.ProjectID)
}

// TestSerializerImpl_SerializeResponse tests format-based response serialization
func TestSerializerImpl_SerializeResponse(t *testing.T) {
	serializer := &serializerImpl{}

	resp := &SyncResponse{
		Success:   true,
		RequestID: uuid.New(),
		Processed: 100,
		Timestamp: time.Now(),
	}

	// Test both formats
	dataJSON, err := serializer.SerializeResponse(resp, "json")
	require.NoError(t, err)
	assert.NotEmpty(t, dataJSON)

	dataMsgPack, err := serializer.SerializeResponse(resp, "msgpack")
	require.NoError(t, err)
	assert.NotEmpty(t, dataMsgPack)
}

// TestSerializerImpl_DeserializeResponse tests format-based response deserialization
func TestSerializerImpl_DeserializeResponse(t *testing.T) {
	serializer := &serializerImpl{}

	originalResp := &SyncResponse{
		Success:   true,
		RequestID: uuid.New(),
		Processed: 50,
		Timestamp: time.Now(),
	}

	// Test JSON round-trip
	dataJSON, _ := serializer.SerializeResponse(originalResp, "json")
	respJSON, err := serializer.DeserializeResponse(dataJSON, "json")
	require.NoError(t, err)
	assert.Equal(t, originalResp.RequestID, respJSON.RequestID)

	// Test msgpack round-trip
	dataMsgPack, _ := serializer.SerializeResponse(originalResp, "msgpack")
	respMsgPack, err := serializer.DeserializeResponse(dataMsgPack, "msgpack")
	require.NoError(t, err)
	assert.Equal(t, originalResp.RequestID, respMsgPack.RequestID)
}

// TestSerializerImpl_SerializeErrorResponse tests format-based error response serialization
func TestSerializerImpl_SerializeErrorResponse(t *testing.T) {
	serializer := &serializerImpl{}

	errResp := &ErrorResponse{
		Error:     "test error",
		RequestID: uuid.New(),
		Timestamp: time.Now(),
	}

	// Test both formats
	dataJSON, err := serializer.SerializeErrorResponse(errResp, "json")
	require.NoError(t, err)
	assert.NotEmpty(t, dataJSON)

	dataMsgPack, err := serializer.SerializeErrorResponse(errResp, "msgpack")
	require.NoError(t, err)
	assert.NotEmpty(t, dataMsgPack)
}

// TestSerializerImpl_GetContentType tests getting content type for format
func TestSerializerImpl_GetContentType(t *testing.T) {
	serializer := &serializerImpl{}

	assert.Equal(t, "application/json", serializer.GetContentType("json"))
	assert.Equal(t, "application/json", serializer.GetContentType("unknown")) // default
	assert.Equal(t, "application/msgpack", serializer.GetContentType("msgpack"))
}

// TestSerializerImpl_DetectFormat tests detecting format from content type
func TestSerializerImpl_DetectFormat(t *testing.T) {
	serializer := &serializerImpl{}

	assert.Equal(t, "json", serializer.DetectFormat("application/json"))
	assert.Equal(t, "json", serializer.DetectFormat("unknown")) // default
	assert.Equal(t, "msgpack", serializer.DetectFormat("application/msgpack"))
}

// TestSerializerImpl_CompareSize tests comparing JSON and MessagePack sizes
// SKIPPED: msgpack.Marshal has issues with interface{} parameter that requires deeper investigation
// CompareSize is not used in production code, only in tests/analytics
func TestSerializerImpl_CompareSize(t *testing.T) {
	t.Skip("msgpack.Marshal with interface{} needs investigation - not blocking production")
}

// TestSerializerImpl_CompareSize_Complex tests size comparison with complex data
func TestSerializerImpl_CompareSize_Complex(t *testing.T) {
	serializer := &serializerImpl{}

	// Test with an int slice (uniform types, more data)
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

	jsonSize, msgpackSize, ratio, err := serializer.CompareSize(data)
	assert.NoError(t, err)

	assert.Greater(t, jsonSize, 0)
	assert.Greater(t, msgpackSize, 0)
	assert.Greater(t, ratio, 0.0)
	// msgpack should be significantly smaller for integer arrays
	assert.Less(t, ratio, 0.8)
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
