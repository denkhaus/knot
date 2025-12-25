package sync

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// setupTestDI creates an injector with mocks for testing
func setupTestDI(mockSerializer *mocks.MockSyncDataSerializer, mockService *mocks.MockSyncService) do.Injector {
	injector := do.New()
	do.ProvideValue[shared.SyncDataSerializer](injector, mockSerializer)
	do.ProvideValue[SyncService](injector, mockService)
	do.Provide(injector, NewSyncHandler)
	return injector
}

func TestSyncHandler_Middleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockSerializer := mocks.NewMockSyncDataSerializer(ctrl)

	mockService := mocks.NewMockSyncService(ctrl)

	// Setup DI injector
	injector := setupTestDI(mockSerializer, mockService)

	// Create handlers
	handlers := &syncEndpoints{
		handler:    do.MustInvoke[SyncHandler](injector),
		service:    mockService,
		serializer: mockSerializer,
	}

	// Test middleware
	middleware := handlers.Middleware()
	assert.NotNil(t, middleware)

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with middleware
	wrapped := middleware(testHandler)
	assert.NotNil(t, wrapped)

	// Test OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/api/sync/test", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "POST, GET, OPTIONS, PUT, DELETE, PATCH", w.Header().Get("Access-Control-Allow-Methods"))
}

func TestSyncHandler_ValidateAndDeserializeRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockSerializer := mocks.NewMockSyncDataSerializer(ctrl)

	mockService := mocks.NewMockSyncService(ctrl)

	// Setup DI injector
	injector := setupTestDI(mockSerializer, mockService)

	// Create test data
	request := shared.SyncRequest{
		ProjectID: uuid.New(),
		Direction: shared.SyncBidirectional,
		RequestID: uuid.New(),
		Timestamp: time.Now(),
	}

	// Expectations
	mockSerializer.EXPECT().DetectFormat("application/json").Return("json").Times(1)
	mockSerializer.EXPECT().DeserializeRequest(gomock.Any(), "json").Return(&request, nil).Times(1)
	mockSerializer.EXPECT().ValidateRequest(&request).Return(shared.ValidationResult{Valid: true}).Times(1)

	// Create handlers
	handlers := &syncEndpoints{
		handler:    do.MustInvoke[SyncHandler](injector),
		service:    mockService,
		serializer: mockSerializer,
	}

	// Create test request
	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/sync/test", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call validation method
	result, ok := handlers.handler.validateAndDeserializeRequest(w, req)

	// Verify result
	assert.True(t, ok)
	assert.NotNil(t, result)
	assert.Equal(t, request.ProjectID, result.ProjectID)
}

func TestSyncHandler_WriteResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockSerializer := mocks.NewMockSyncDataSerializer(ctrl)

	mockService := mocks.NewMockSyncService(ctrl)

	// Setup DI injector
	injector := setupTestDI(mockSerializer, mockService)

	// Create test data
	response := &shared.SyncResponse{
		Success:    true,
		RequestID:  uuid.New(),
		Processed:  10,
		Created:    5,
		Updated:    3,
		Deleted:    2,
		NewVersion: 124,
		Timestamp:  time.Now(),
		Duration:   100 * time.Millisecond,
	}

	// Expectations
	mockSerializer.EXPECT().GetContentType("json").Return("application/json").Times(1)
	mockSerializer.EXPECT().SerializeResponse(response, "json").Return([]byte(`{"success":true}`), nil).Times(1)

	// Create handlers
	handlers := &syncEndpoints{
		handler:    do.MustInvoke[SyncHandler](injector),
		service:    mockService,
		serializer: mockSerializer,
	}

	// Create test request
	req := httptest.NewRequest("POST", "/api/sync/test", nil)
	w := httptest.NewRecorder()

	// Call write response
	err := handlers.handler.writeResponse(w, req, response)

	// Verify result
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, `{"success":true}`, w.Body.String())
}

func TestSyncHandler_WriteErrorResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockSerializer := mocks.NewMockSyncDataSerializer(ctrl)

	mockService := mocks.NewMockSyncService(ctrl)

	// Setup DI injector
	injector := setupTestDI(mockSerializer, mockService)

	// Create test data
	requestID := uuid.New()
	testError := &ValidationError{Message: "test error"}

	// Expectations
	mockSerializer.EXPECT().GetContentType("json").Return("application/json").Times(1)
	mockSerializer.EXPECT().SerializeErrorResponse(gomock.Any(), "json").Return([]byte(`{"error":"test error"}`), nil).Times(1)

	// Create handlers
	handlers := &syncEndpoints{
		handler:    do.MustInvoke[SyncHandler](injector),
		service:    mockService,
		serializer: mockSerializer,
	}

	// Create test request
	req := httptest.NewRequest("POST", "/api/sync/test", nil)
	w := httptest.NewRecorder()

	// Call write error response
	handlers.handler.writeErrorResponse(w, req, requestID, http.StatusBadRequest, testError)

	// Verify result
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "test error")
}

func TestSyncHandler_ParseFormatFromRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockSerializer := mocks.NewMockSyncDataSerializer(ctrl)

	mockService := mocks.NewMockSyncService(ctrl)

	// Setup DI injector
	injector := setupTestDI(mockSerializer, mockService)

	// Create handlers
	handlers := &syncEndpoints{
		handler:    do.MustInvoke[SyncHandler](injector),
		service:    mockService,
		serializer: mockSerializer,
	}

	// Test JSON format
	req1 := httptest.NewRequest("POST", "/api/sync/test", nil)
	req1.Header.Set("Content-Type", "application/json")
	assert.Equal(t, "json", handlers.handler.parseFormatFromRequest(req1))

	// Test MessagePack format
	req2 := httptest.NewRequest("POST", "/api/sync/test", nil)
	req2.Header.Set("Content-Type", "application/msgpack")
	assert.Equal(t, "msgpack", handlers.handler.parseFormatFromRequest(req2))

	// Test Accept header
	req3 := httptest.NewRequest("POST", "/api/sync/test", nil)
	req3.Header.Set("Accept", "application/msgpack")
	assert.Equal(t, "msgpack", handlers.handler.parseFormatFromRequest(req3))

	// Test query parameter
	req4 := httptest.NewRequest("GET", "/api/sync/test?format=msgpack", nil)
	assert.Equal(t, "msgpack", handlers.handler.parseFormatFromRequest(req4))
}

func TestSyncHandler_HealthHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockSerializer := mocks.NewMockSyncDataSerializer(ctrl)

	mockService := mocks.NewMockSyncService(ctrl)

	// Setup DI injector
	injector := setupTestDI(mockSerializer, mockService)

	// Expectations
	mockService.EXPECT().HealthCheck(gomock.Any()).Return(nil).Times(1)

	// Create handlers
	handlers := &syncEndpoints{
		handler:    do.MustInvoke[SyncHandler](injector),
		service:    mockService,
		serializer: mockSerializer,
	}

	// Create test request
	req := httptest.NewRequest("GET", "/api/sync/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler := handlers.HealthHandler()
	handler(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestSyncHandler_HealthHandler_Unhealthy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockSerializer := mocks.NewMockSyncDataSerializer(ctrl)

	mockService := mocks.NewMockSyncService(ctrl)

	// Setup DI injector
	injector := setupTestDI(mockSerializer, mockService)

	// Expectations
	mockService.EXPECT().HealthCheck(gomock.Any()).Return(assert.AnError).Times(1)

	// Create handlers
	handlers := &syncEndpoints{
		handler:    do.MustInvoke[SyncHandler](injector),
		service:    mockService,
		serializer: mockSerializer,
	}

	// Create test request
	req := httptest.NewRequest("GET", "/api/sync/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler := handlers.HealthHandler()
	handler(w, req)

	// Verify response
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "unhealthy")
}

func TestSyncHandler_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockSerializer := mocks.NewMockSyncDataSerializer(ctrl)

	mockService := mocks.NewMockSyncService(ctrl)

	// Setup DI injector
	injector := setupTestDI(mockSerializer, mockService)

	// Create handlers
	handlers := &syncEndpoints{
		handler:    do.MustInvoke[SyncHandler](injector),
		service:    mockService,
		serializer: mockSerializer,
	}

	// Test GET request on POST endpoint
	req := httptest.NewRequest("GET", "/api/sync/full", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler := handlers.FullSyncHandler()
	handler(w, req)

	// Verify response
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestSyncHandler_RequestTooLarge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mocks
	mockSerializer := mocks.NewMockSyncDataSerializer(ctrl)

	mockService := mocks.NewMockSyncService(ctrl)

	// Setup DI injector
	injector := setupTestDI(mockSerializer, mockService)

	// Add expectation for GetContentType (called before size check)
	mockSerializer.EXPECT().GetContentType(gomock.Any()).Return("application/json").AnyTimes()
	// Add expectation for SerializeErrorResponse (called when size check fails)
	mockSerializer.EXPECT().SerializeErrorResponse(gomock.Any(), "json").Return([]byte(`{"error":"request body too large"}`), nil).AnyTimes()

	// Create handlers
	handlers := &syncEndpoints{
		handler:    do.MustInvoke[SyncHandler](injector),
		service:    mockService,
		serializer: mockSerializer,
	}

	// Create request that's too large
	largeBody := make([]byte, MaxRequestSize+1)
	req := httptest.NewRequest("POST", "/api/sync/full", bytes.NewReader(largeBody))
	req.ContentLength = MaxRequestSize + 1
	w := httptest.NewRecorder()

	// Call validation method
	result, ok := handlers.handler.validateAndDeserializeRequest(w, req)

	// Verify error response
	assert.False(t, ok)
	assert.Nil(t, result)
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}
