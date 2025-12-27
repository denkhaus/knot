// Package sync provides HTTP handlers for sync API endpoints
package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

// Constants
const (
	MaxRequestSize = 10 * 1024 * 1024 // 10MB
	RequestTimeout = 30 * time.Second
)

// Errors
var (
	ErrInvalidResponseType    = errors.New("invalid response type")
	ErrRequestTooLarge        = errors.New("request body too large")
	ErrRequestBodyRead        = errors.New("failed to read request body")
	ErrRequestDeserialization = errors.New("failed to deserialize request")
)

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return "validation error: " + e.Message
}

// SyncHandler provides HTTP handlers for sync operations
type SyncHandler interface {
	// writeResponse writes response in appropriate format
	writeResponse(w http.ResponseWriter, r *http.Request, response interface{}) error

	// writeErrorResponse writes error response with proper status code
	writeErrorResponse(w http.ResponseWriter, r *http.Request, requestID uuid.UUID, httpStatus int, err error)

	// validateAndDeserializeRequest validates and deserializes request
	validateAndDeserializeRequest(w http.ResponseWriter, r *http.Request) (*shared.SyncRequest, bool)
}

// syncHandlerImpl provides private implementation of SyncHandler
type syncHandlerImpl struct {
	serializer shared.SyncDataSerializer
}

// Ensure syncHandlerImpl implements SyncHandler
var _ SyncHandler = (*syncHandlerImpl)(nil)

// NewSyncHandler creates a new sync handler following DI pattern
func NewSyncHandler(injector do.Injector) (SyncHandler, error) {
	// Resolve dependencies using do.MustInvoke as per DI pattern
	serializer := do.MustInvoke[shared.SyncDataSerializer](injector)

	return &syncHandlerImpl{
		serializer: serializer,
	}, nil
}

// HTTP request/response helpers


// writeResponse writes response in JSON format
func (h *syncHandlerImpl) writeResponse(w http.ResponseWriter, r *http.Request, response interface{}) error {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Serialize to JSON
	var data []byte
	var err error

	switch resp := response.(type) {
	case *shared.SyncRequest:
		data, err = h.serializer.SerializeRequest(resp)
	case *shared.SyncResponse:
		data, err = h.serializer.SerializeResponse(resp)
	case *shared.ErrorResponse:
		data, err = h.serializer.SerializeErrorResponse(resp)
	default:
		return ErrInvalidResponseType
	}

	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	return err
}

// writeErrorResponse writes error response with proper status code
func (h *syncHandlerImpl) writeErrorResponse(w http.ResponseWriter, r *http.Request, requestID uuid.UUID, httpStatus int, err error) {
	// Create error response
	errorResp := &shared.ErrorResponse{
		Error:     err.Error(),
		RequestID: requestID,
		Timestamp: time.Now(),
		Retryable: httpStatus < 500, // Client errors (4xx) are not retryable
	}

	// Set status code
	w.WriteHeader(httpStatus)

	// Write response
	if writeErr := h.writeResponse(w, r, errorResp); writeErr != nil {
		// If we can't write error response, fallback to plain text
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}
}

// validateAndDeserializeRequest validates and deserializes request (JSON only)
func (h *syncHandlerImpl) validateAndDeserializeRequest(w http.ResponseWriter, r *http.Request) (*shared.SyncRequest, bool) {
	// Check content length
	if r.ContentLength > MaxRequestSize {
		h.writeErrorResponse(w, r, uuid.Nil, http.StatusRequestEntityTooLarge, ErrRequestTooLarge)
		return nil, false
	}

	// Read request body using io.ReadAll to handle large payloads correctly
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeErrorResponse(w, r, uuid.Nil, http.StatusBadRequest, ErrRequestBodyRead)
		return nil, false
	}

	// Deserialize request (JSON only)
	request, err := h.serializer.DeserializeRequest(body)
	if err != nil {
		// Debug: Log the actual deserialization error
		http.Error(w, fmt.Sprintf("DEBUG: Deserialization error: %v", err), http.StatusBadRequest)
		return nil, false
	}

	// Validate request
	validation := h.serializer.ValidateRequest(request)
	if !validation.Valid {
		// Use first validation error for response
		var errStr string
		if len(validation.Errors) > 0 {
			errStr = validation.Errors[0].Message
		} else {
			errStr = "Validation failed"
		}

		h.writeErrorResponse(w, r, request.RequestID, http.StatusBadRequest, &ValidationError{Message: errStr})
		return nil, false
	}

	return request, true
}
