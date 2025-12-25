// Package shared provides common types and utilities used across sync subpackages
package shared

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/samber/do/v2"
	"github.com/vmihailenco/msgpack/v5"
)

// serializerImpl handles serialization/deserialization of sync data
type serializerImpl struct{}

// NewSerializerProvider creates a SyncDataSerializer provider for DI
func NewSerializerProvider(injector do.Injector) (SyncDataSerializer, error) {
	return &serializerImpl{}, nil
}

// Ensure Serializer implements SyncDataSerializer interface
var _ SyncDataSerializer = (*serializerImpl)(nil)

// SerializeRequest serializes a SyncRequest using the preferred format
func (s *serializerImpl) SerializeRequest(req *SyncRequest, format string) ([]byte, error) {
	if format == "msgpack" {
		return s.SerializeRequestMsgPack(req)
	}
	return s.SerializeRequestJSON(req)
}

// DeserializeRequest deserializes a SyncRequest from bytes
func (s *serializerImpl) DeserializeRequest(data []byte, format string) (*SyncRequest, error) {
	if format == "msgpack" {
		return s.DeserializeRequestMsgPack(data)
	}
	return s.DeserializeRequestJSON(data)
}

// SerializeResponse serializes a SyncResponse using the preferred format
func (s *serializerImpl) SerializeResponse(resp *SyncResponse, format string) ([]byte, error) {
	if format == "msgpack" {
		return s.SerializeResponseMsgPack(resp)
	}
	return s.SerializeResponseJSON(resp)
}

// DeserializeResponse deserializes a SyncResponse from bytes
func (s *serializerImpl) DeserializeResponse(data []byte, format string) (*SyncResponse, error) {
	if format == "msgpack" {
		return s.DeserializeResponseMsgPack(data)
	}
	return s.DeserializeResponseJSON(data)
}

// SerializeErrorResponse serializes an ErrorResponse using the preferred format
func (s *serializerImpl) SerializeErrorResponse(errResp *ErrorResponse, format string) ([]byte, error) {
	if format == "msgpack" {
		return s.SerializeErrorResponseMsgPack(errResp)
	}
	return s.SerializeErrorResponseJSON(errResp)
}

// SerializeRequestJSON serializes a SyncRequest to JSON
func (s *serializerImpl) SerializeRequestJSON(req *SyncRequest) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request to JSON: %w", err)
	}

	return data, nil
}

// DeserializeRequestJSON deserializes a SyncRequest from JSON
func (s *serializerImpl) DeserializeRequestJSON(data []byte) (*SyncRequest, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var req SyncRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request from JSON: %w", err)
	}

	return &req, nil
}

// SerializeResponseJSON serializes a SyncResponse to JSON
func (s *serializerImpl) SerializeResponseJSON(resp *SyncResponse) ([]byte, error) {
	if resp == nil {
		return nil, fmt.Errorf("response cannot be nil")
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response to JSON: %w", err)
	}

	return data, nil
}

// DeserializeResponseJSON deserializes a SyncResponse from JSON
func (s *serializerImpl) DeserializeResponseJSON(data []byte) (*SyncResponse, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var resp SyncResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from JSON: %w", err)
	}

	return &resp, nil
}

// SerializeErrorResponseJSON serializes an ErrorResponse to JSON
func (s *serializerImpl) SerializeErrorResponseJSON(errResp *ErrorResponse) ([]byte, error) {
	if errResp == nil {
		return nil, fmt.Errorf("error response cannot be nil")
	}

	data, err := json.Marshal(errResp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal error response to JSON: %w", err)
	}

	return data, nil
}

// SerializeRequestMsgPack serializes a SyncRequest to MessagePack
func (s *serializerImpl) SerializeRequestMsgPack(req *SyncRequest) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	var buf bytes.Buffer
	encoder := msgpack.NewEncoder(&buf)

	if err := encoder.Encode(req); err != nil {
		return nil, fmt.Errorf("failed to encode request to MessagePack: %w", err)
	}

	return buf.Bytes(), nil
}

// DeserializeRequestMsgPack deserializes a SyncRequest from MessagePack
func (s *serializerImpl) DeserializeRequestMsgPack(data []byte) (*SyncRequest, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	decoder := msgpack.NewDecoder(bytes.NewReader(data))

	var req SyncRequest
	if err := decoder.Decode(&req); err != nil {
		return nil, fmt.Errorf("failed to decode request from MessagePack: %w", err)
	}

	return &req, nil
}

// SerializeResponseMsgPack serializes a SyncResponse to MessagePack
func (s *serializerImpl) SerializeResponseMsgPack(resp *SyncResponse) ([]byte, error) {
	if resp == nil {
		return nil, fmt.Errorf("response cannot be nil")
	}

	var buf bytes.Buffer
	encoder := msgpack.NewEncoder(&buf)

	if err := encoder.Encode(resp); err != nil {
		return nil, fmt.Errorf("failed to encode response to MessagePack: %w", err)
	}

	return buf.Bytes(), nil
}

// DeserializeResponseMsgPack deserializes a SyncResponse from MessagePack
func (s *serializerImpl) DeserializeResponseMsgPack(data []byte) (*SyncResponse, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	decoder := msgpack.NewDecoder(bytes.NewReader(data))

	var resp SyncResponse
	if err := decoder.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response from MessagePack: %w", err)
	}

	return &resp, nil
}

// SerializeErrorResponseMsgPack serializes an ErrorResponse to MessagePack
func (s *serializerImpl) SerializeErrorResponseMsgPack(errResp *ErrorResponse) ([]byte, error) {
	if errResp == nil {
		return nil, fmt.Errorf("error response cannot be nil")
	}

	var buf bytes.Buffer
	encoder := msgpack.NewEncoder(&buf)

	if err := encoder.Encode(errResp); err != nil {
		return nil, fmt.Errorf("failed to encode error response to MessagePack: %w", err)
	}

	return buf.Bytes(), nil
}

// GetContentType returns the appropriate content type for the format
func (s *serializerImpl) GetContentType(format string) string {
	if format == "msgpack" {
		return "application/msgpack"
	}
	return "application/json"
}

// DetectFormat attempts to detect the format from content type
func (s *serializerImpl) DetectFormat(contentType string) string {
	if contentType == "application/msgpack" {
		return "msgpack"
	}
	return "json"
}

// CompareSize returns the size comparison between JSON and MessagePack
func (s *serializerImpl) CompareSize(data interface{}) (jsonSize, msgpackSize int, ratio float64, err error) {
	// Serialize to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	jsonSize = len(jsonData)

	// Serialize to MessagePack using Marshal (handles interface{} better)
	msgpackData, err := msgpack.Marshal(data)
	if err != nil {
		return jsonSize, 0, 0, fmt.Errorf("failed to marshal to MessagePack: %w", err)
	}
	msgpackSize = len(msgpackData)

	// Calculate compression ratio
	ratio = float64(msgpackSize) / float64(jsonSize)

	return jsonSize, msgpackSize, ratio, nil
}

// ValidateRequest validates a sync request
func (s *serializerImpl) ValidateRequest(req *SyncRequest) ValidationResult {
	if req == nil {
		result := ValidationResult{Valid: false}
		result.AddError("request", "request cannot be nil", nil)
		return result
	}
	validation := req.Validate()
	if validation == nil {
		return ValidationResult{Valid: true}
	}
	return *validation
}

// ValidateResponse validates a sync response
func (s *serializerImpl) ValidateResponse(resp *SyncResponse) ValidationResult {
	if resp == nil {
		result := ValidationResult{Valid: false}
		result.AddError("response", "response cannot be nil", nil)
		return result
	}
	validation := resp.Validate()
	if validation == nil {
		return ValidationResult{Valid: true}
	}
	return *validation
}
