// Package shared provides common types and utilities used across sync subpackages
package shared

import (
	"encoding/json"
	"fmt"

	"github.com/samber/do/v2"
)

// serializerImpl handles serialization/deserialization of sync data (JSON only)
type serializerImpl struct{}

// NewSerializerProvider creates a SyncDataSerializer provider for DI
func NewSerializerProvider(injector do.Injector) (SyncDataSerializer, error) {
	return &serializerImpl{}, nil
}

// Ensure Serializer implements SyncDataSerializer interface
var _ SyncDataSerializer = (*serializerImpl)(nil)

// SerializeRequest serializes a SyncRequest to JSON
func (s *serializerImpl) SerializeRequest(req *SyncRequest) ([]byte, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request to JSON: %w", err)
	}

	return data, nil
}

// DeserializeRequest deserializes a SyncRequest from JSON
func (s *serializerImpl) DeserializeRequest(data []byte) (*SyncRequest, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var req SyncRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request from JSON: %w", err)
	}

	return &req, nil
}

// SerializeResponse serializes a SyncResponse to JSON
func (s *serializerImpl) SerializeResponse(resp *SyncResponse) ([]byte, error) {
	if resp == nil {
		return nil, fmt.Errorf("response cannot be nil")
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response to JSON: %w", err)
	}

	return data, nil
}

// DeserializeResponse deserializes a SyncResponse from JSON
func (s *serializerImpl) DeserializeResponse(data []byte) (*SyncResponse, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var resp SyncResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from JSON: %w", err)
	}

	return &resp, nil
}

// SerializeErrorResponse serializes an ErrorResponse to JSON
func (s *serializerImpl) SerializeErrorResponse(errResp *ErrorResponse) ([]byte, error) {
	if errResp == nil {
		return nil, fmt.Errorf("error response cannot be nil")
	}

	data, err := json.Marshal(errResp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal error response to JSON: %w", err)
	}

	return data, nil
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
