// Package client provides internal tests for the sync client.
//
// This file uses the same package (client) to access private methods for testing.
// The client_test.go file uses package client_test to test the public interface
// while avoiding import cycles with the mocks package.
package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// testLogger is a minimal logger implementation for testing
type internalTestLogger struct{}

func newInternalTestLogger() *internalTestLogger {
	return &internalTestLogger{}
}

func (l *internalTestLogger) Debug(msg string, fields ...zap.Field) {}
func (l *internalTestLogger) Info(msg string, fields ...zap.Field)  {}
func (l *internalTestLogger) Warn(msg string, fields ...zap.Field)  {}
func (l *internalTestLogger) Error(msg string, fields ...zap.Field) {}
func (l *internalTestLogger) Sync()                                 {}
func (l *internalTestLogger) With(fields ...zap.Field) logger.Logger {
	return l
}

func (l *internalTestLogger) Named(name string) logger.Logger {
	return l
}

func (l *internalTestLogger) ToZap() *zap.Logger {
	return zap.NewNop()
}
func (l *internalTestLogger) SetLevel(level string) {}

// createInternalTestClient creates a test client with the given server URL
func createInternalTestClient(serverURL string) *restSyncClientImpl {
	cfg := &config.SyncConfig{
		ServerURL:        serverURL,
		Timeout:          5 * time.Second,
		RetryAttempts:    2,
		RetryDelay:       100 * time.Millisecond,
		MaxRetryDelay:    500 * time.Millisecond,
		MaxIdleConns:     5,
		IdleConnTimeout:  30 * time.Second,
		ConflictStrategy: "last-writer-wins",
		BatchSize:        50,
		UserAgent:        "knot-test/1.0",
	}

	serializer, _ := shared.NewSerializerProvider(nil)
	return &restSyncClientImpl{
		config:     cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
		serializer: serializer,
		logger:     newInternalTestLogger(),
	}
}

// Test helper to create test project
func createInternalTestProject(id uuid.UUID) *types.Project {
	return &types.Project{
		ID:          id,
		Title:       "Test Project " + id.String()[:8],
		Description: "Test project description",
		State:       types.ProjectStateActive,
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now(),
		CreatedBy:   "test-user",
		UpdatedBy:   "test-user",
	}
}

// Test helper to create test task
func createInternalTestTask(id, projectID uuid.UUID) *types.Task {
	return &types.Task{
		ID:          id,
		ProjectID:   projectID,
		Title:       "Test Task " + id.String()[:8],
		Description: "Test task description",
		State:       types.TaskStatePending,
		Priority:    types.TaskPriority(5),
		Complexity:  7,
		CreatedAt:   time.Now().Add(-12 * time.Hour),
		UpdatedAt:   time.Now(),
		CreatedBy:   "test-user",
		UpdatedBy:   "test-user",
	}
}

// Test helper to create test SyncDataSet
func createInternalTestDataSet() *shared.SyncDataSet {
	projectID := uuid.New()
	taskID := uuid.New()

	project := createInternalTestProject(projectID)
	task := createInternalTestTask(taskID, projectID)

	return &shared.SyncDataSet{
		Projects: map[uuid.UUID]*types.Project{
			project.ID: project,
		},
		Tasks: map[uuid.UUID]*types.Task{
			task.ID: task,
		},
	}
}

func TestClientGetProjects_Success(t *testing.T) {
	expectedDataSet := createInternalTestDataSet()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/sync/projects" {
			projectsJSON, _ := json.Marshal(map[string]interface{}{
				"projects": expectedDataSet.Projects,
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(projectsJSON)
			return
		}
		t.Fatalf("Unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	client := createInternalTestClient(server.URL)
	ctx := context.Background()

	result, err := client.GetProjects(ctx)
	if err != nil {
		t.Fatalf("getProjects failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if len(result.Projects) != len(expectedDataSet.Projects) {
		t.Errorf("Expected %d projects, got %d", len(expectedDataSet.Projects), len(result.Projects))
	}
}

func TestClientGetTasks_Success(t *testing.T) {
	projectID := uuid.New()
	expectedDataSet := createInternalTestDataSet()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/sync/tasks" {
			// Verify project_id query parameter
			if r.URL.Query().Get("project_id") != projectID.String() {
				t.Errorf("Expected project_id %s, got %s", projectID.String(), r.URL.Query().Get("project_id"))
			}

			tasksJSON, _ := json.Marshal(map[string]interface{}{
				"tasks": expectedDataSet.Tasks,
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(tasksJSON)
			return
		}
		t.Fatalf("Unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	client := createInternalTestClient(server.URL)
	ctx := context.Background()

	result, err := client.GetTasks(ctx, projectID)
	if err != nil {
		t.Fatalf("getTasks failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if len(result.Tasks) != len(expectedDataSet.Tasks) {
		t.Errorf("Expected %d tasks, got %d", len(expectedDataSet.Tasks), len(result.Tasks))
	}
}

func TestClientGetTasks_WithSinceParameter(t *testing.T) {
	projectID := uuid.New()
	sinceTime := time.Now().Add(-24 * time.Hour)
	expectedDataSet := createInternalTestDataSet()

	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/sync/tasks" {
			serverCalled = true

			// Verify project_id and since query parameters
			if r.URL.Query().Get("project_id") != projectID.String() {
				t.Errorf("Expected project_id %s, got %s", projectID.String(), r.URL.Query().Get("project_id"))
			}

			sinceParam := r.URL.Query().Get("since")
			if sinceParam == "" {
				t.Error("Expected 'since' query parameter to be set")
			}

			tasksJSON, _ := json.Marshal(map[string]interface{}{
				"tasks": expectedDataSet.Tasks,
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(tasksJSON)
			return
		}
	}))
	defer server.Close()

	cfg := &config.SyncConfig{
		ServerURL:       server.URL,
		Timeout:         5 * time.Second,
		RetryAttempts:   2,
		RetryDelay:      100 * time.Millisecond,
		MaxRetryDelay:   500 * time.Millisecond,
		MaxIdleConns:    5,
		IdleConnTimeout: 30 * time.Second,
		Since:           &sinceTime,
	}

	serializer, _ := shared.NewSerializerProvider(nil)
	client := &restSyncClientImpl{
		config:     cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
		serializer: serializer,
		logger:     newInternalTestLogger(),
	}

	ctx := context.Background()

	result, err := client.GetTasks(ctx, projectID)
	if err != nil {
		t.Fatalf("getTasks failed: %v", err)
	}

	if !serverCalled {
		t.Error("Server was not called")
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if len(result.Tasks) != len(expectedDataSet.Tasks) {
		t.Errorf("Expected %d tasks, got %d", len(expectedDataSet.Tasks), len(result.Tasks))
	}
}

func TestClientHealthCheck_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/sync/health" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"healthy"}`))
			return
		}
		t.Fatalf("Unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	client := createInternalTestClient(server.URL)
	ctx := context.Background()

	err := client.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}
}

func TestClientHealthCheck_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/sync/health" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	}))
	defer server.Close()

	client := createInternalTestClient(server.URL)
	ctx := context.Background()

	err := client.HealthCheck(ctx)
	if err == nil {
		t.Fatal("Expected error from HealthCheck, got nil")
	}

	if !containsString(err.Error(), "health check failed") {
		t.Errorf("Expected health check error, got: %v", err)
	}
}

func TestClientRetry_SuccessOnSecondAttempt(t *testing.T) {
	attemptCount := 0
	expectedDataSet := createInternalTestDataSet()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		// First attempt fails with 500, second succeeds
		if attemptCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		projectsJSON, _ := json.Marshal(map[string]interface{}{
			"projects": expectedDataSet.Projects,
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(projectsJSON)
	}))
	defer server.Close()

	client := createInternalTestClient(server.URL)
	ctx := context.Background()

	result, err := client.GetProjects(ctx)
	if err != nil {
		t.Fatalf("getProjects failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if attemptCount != 2 {
		t.Errorf("Expected 2 attempts, got %d", attemptCount)
	}
}

func TestClientRetry_MaxAttemptsExceeded(t *testing.T) {
	attemptCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		// Always fail with 500
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := createInternalTestClient(server.URL)
	ctx := context.Background()

	_, err := client.GetProjects(ctx)
	if err == nil {
		t.Fatal("Expected error after retries, got nil")
	}

	// Should have attempted initial + 2 retries = 3 total
	expectedAttempts := 3
	if attemptCount != expectedAttempts {
		t.Errorf("Expected %d attempts, got %d", expectedAttempts, attemptCount)
	}

	if !containsString(err.Error(), "failed after 3 attempts") {
		t.Errorf("Expected 'failed after 3 attempts' error, got: %v", err)
	}
}

func TestClientNonRetryableError(t *testing.T) {
	serverCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		// Return 400 Bad Request - not retryable
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid request"}`))
	}))
	defer server.Close()

	client := createInternalTestClient(server.URL)
	ctx := context.Background()

	_, err := client.GetProjects(ctx)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !serverCalled {
		t.Error("Server should have been called once")
	}

	if !containsString(err.Error(), "sync server error") {
		t.Errorf("Expected sync server error, got: %v", err)
	}
}

func TestClientClose(t *testing.T) {
	client := createInternalTestClient("http://localhost:9094")

	err := client.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify close was logged
	// Note: Can't easily test idle connections are closed without accessing internals
}

func TestClientCalculateRetryDelay(t *testing.T) {
	client := createInternalTestClient("http://localhost:9094")

	tests := []struct {
		name          string
		attempt       int
		expectedDelay time.Duration
	}{
		{"First retry", 1, 100 * time.Millisecond},
		{"Second retry", 2, 200 * time.Millisecond},
		{"Third retry", 3, 400 * time.Millisecond},
		{"Fourth retry (capped)", 4, 500 * time.Millisecond}, // Capped at MaxRetryDelay
		{"Fifth retry (capped)", 5, 500 * time.Millisecond},  // Capped at MaxRetryDelay
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := client.calculateRetryDelay(tt.attempt)
			if delay != tt.expectedDelay {
				t.Errorf("Expected delay %v, got %v", tt.expectedDelay, delay)
			}
		})
	}
}

func TestClientIsRetryableStatusCode(t *testing.T) {
	client := createInternalTestClient("http://localhost:9094")

	retryableCodes := []int{
		http.StatusTooManyRequests,     // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout,      // 504
	}

	for _, code := range retryableCodes {
		if !client.isRetryableStatusCode(code) {
			t.Errorf("Expected status code %d to be retryable", code)
		}
	}

	nonRetryableCodes := []int{
		http.StatusBadRequest,       // 400
		http.StatusUnauthorized,     // 401
		http.StatusForbidden,        // 403
		http.StatusNotFound,         // 404
		http.StatusMethodNotAllowed, // 405
		http.StatusConflict,         // 409
		http.StatusOK,               // 200
	}

	for _, code := range nonRetryableCodes {
		if client.isRetryableStatusCode(code) {
			t.Errorf("Expected status code %d to NOT be retryable", code)
		}
	}
}

func TestClientFormatErrorResponse_WithErrorJSON(t *testing.T) {
	client := createInternalTestClient("http://localhost:9094")

	statusCode := http.StatusBadRequest
	body := []byte(`{"error":"invalid project ID","code":"INVALID_INPUT"}`)

	err := client.formatErrorResponse(statusCode, body)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !containsString(err.Error(), "invalid project ID") {
		t.Errorf("Expected error message to contain 'invalid project ID', got: %v", err)
	}

	if !containsString(err.Error(), "400") {
		t.Errorf("Expected error message to contain status code, got: %v", err)
	}
}

func TestClientFormatErrorResponse_WithRawBody(t *testing.T) {
	client := createInternalTestClient("http://localhost:9094")

	statusCode := http.StatusInternalServerError
	body := []byte("Internal server error occurred")

	err := client.formatErrorResponse(statusCode, body)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !containsString(err.Error(), "Internal server error occurred") {
		t.Errorf("Expected error message to contain body, got: %v", err)
	}
}

func TestClientBuildEndpointURL(t *testing.T) {
	tests := []struct {
		name         string
		serverURL    string
		endpointPath string
		expectedURL  string
		expectError  bool
	}{
		{
			name:         "Valid URL and path",
			serverURL:    "http://localhost:9094",
			endpointPath: "/api/sync/full",
			expectedURL:  "http://localhost:9094/api/sync/full",
			expectError:  false,
		},
		{
			name:         "URL with existing path",
			serverURL:    "http://localhost:9094/api",
			endpointPath: "sync/full",
			expectedURL:  "http://localhost:9094/api/sync/full",
			expectError:  false,
		},
		{
			name:         "Invalid server URL",
			serverURL:    ":invalid-url",
			endpointPath: "/api/sync/full",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createInternalTestClient(tt.serverURL)

			resultURL, err := client.buildEndpointURL(tt.endpointPath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if resultURL != tt.expectedURL {
				t.Errorf("Expected URL %s, got %s", tt.expectedURL, resultURL)
			}
		})
	}
}

func TestClientSetRequestHeaders(t *testing.T) {
	tests := []struct {
		name       string
		format     string
		authToken  string
		userAgent  string
		expectedCT string
		expectAuth bool
		expectUA   bool
	}{
		{
			name:       "JSON format with auth",
			format:     "json",
			authToken:  "test-token",
			userAgent:  "test-agent/1.0",
			expectedCT: "application/json",
			expectAuth: true,
			expectUA:   true,
		},
		{
			name:       "JSON format without auth",
			format:     "json",
			authToken:  "",
			userAgent:  "",
			expectedCT: "application/json",
			expectAuth: false,
			expectUA:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.SyncConfig{
				ServerURL:       "http://localhost:9094",
				AuthToken:       tt.authToken,
				UserAgent:       tt.userAgent,
				Timeout:         5 * time.Second,
				RetryAttempts:   2,
				RetryDelay:      100 * time.Millisecond,
				MaxRetryDelay:   500 * time.Millisecond,
				MaxIdleConns:    5,
				IdleConnTimeout: 30 * time.Second,
			}

			serializer, _ := shared.NewSerializerProvider(nil)
			client := &restSyncClientImpl{
				config:     cfg,
				httpClient: &http.Client{Timeout: cfg.Timeout},
				serializer: serializer,
				logger:     newInternalTestLogger(),
			}

			req, _ := http.NewRequest("POST", "http://example.com", nil)
			client.setRequestHeaders(req, tt.format)

			// Check Content-Type
			ct := req.Header.Get("Content-Type")
			if ct != tt.expectedCT {
				t.Errorf("Expected Content-Type %s, got %s", tt.expectedCT, ct)
			}

			// Check Accept
			accept := req.Header.Get("Accept")
			if accept != tt.expectedCT {
				t.Errorf("Expected Accept %s, got %s", tt.expectedCT, accept)
			}

			// Check Authorization
			auth := req.Header.Get("Authorization")
			if tt.expectAuth {
				if auth != "Bearer "+tt.authToken {
					t.Errorf("Expected Authorization header, got %s", auth)
				}
			} else {
				if auth != "" {
					t.Errorf("Expected no Authorization header, got %s", auth)
				}
			}

			// Check User-Agent
			ua := req.Header.Get("User-Agent")
			if tt.expectUA {
				if ua != tt.userAgent {
					t.Errorf("Expected User-Agent %s, got %s", tt.userAgent, ua)
				}
			} else {
				if ua != "" {
					t.Errorf("Expected no User-Agent header, got %s", ua)
				}
			}
		})
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && indexOfString(s, substr) >= 0)
}

func indexOfString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Sync method tests
func TestClientSync_PushDirection(t *testing.T) {
	projectID := uuid.New()
	requestID := uuid.New()

	testData := createInternalTestDataSet()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/sync/push" {
			// Verify method
			if r.Method != "POST" {
				t.Errorf("Expected POST, got %s", r.Method)
			}

			// Verify headers
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
			}

			// Verify request body contains sync data
			var reqBody map[string]interface{}
			json.NewDecoder(r.Body).Decode(&reqBody)

			if reqBody["request_id"] == nil {
				t.Error("Expected request_id in body")
			}

			// Return success response
			response := map[string]interface{}{
				"success":   true,
				"processed": 1,
				"message":   "Sync completed",
				"data": map[string]interface{}{
					"projects": testData.Projects,
					"tasks":    testData.Tasks,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}
		t.Fatalf("Unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	client := createInternalTestClient(server.URL)

	req := &shared.SyncRequest{
		RequestID: requestID,
		ProjectID: projectID,
		Direction: shared.SyncLocalToMCP,
		LocalData: testData,
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	resp, err := client.Sync(ctx, req)

	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}

	if resp.Processed != 1 {
		t.Errorf("Expected processed=1, got %d", resp.Processed)
	}
}

func TestClientSync_PullDirection(t *testing.T) {
	projectID := uuid.New()
	requestID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/sync/pull" {
			// Return success response
			response := map[string]interface{}{
				"success":   true,
				"processed": 5,
				"message":   "Pull completed",
				"data":      createInternalTestDataSet(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}
		t.Fatalf("Unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	client := createInternalTestClient(server.URL)

	req := &shared.SyncRequest{
		RequestID: requestID,
		ProjectID: projectID,
		Direction: shared.SyncMcpToLocal,
		LocalData: createInternalTestDataSet(),
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	resp, err := client.Sync(ctx, req)

	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}

	if resp.Processed != 5 {
		t.Errorf("Expected processed=5, got %d", resp.Processed)
	}
}

func TestClientSync_BidirectionalDirection(t *testing.T) {
	projectID := uuid.New()
	requestID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/sync/full" {
			// Return success response
			response := map[string]interface{}{
				"success":   true,
				"processed": 10,
				"message":   "Full sync completed",
				"data":      createInternalTestDataSet(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}
		t.Fatalf("Unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	client := createInternalTestClient(server.URL)

	req := &shared.SyncRequest{
		RequestID: requestID,
		ProjectID: projectID,
		Direction: shared.SyncBidirectional,
		LocalData: createInternalTestDataSet(),
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	resp, err := client.Sync(ctx, req)

	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}

	if resp.Processed != 10 {
		t.Errorf("Expected processed=10, got %d", resp.Processed)
	}
}

func TestClientSync_UnsupportedDirection(t *testing.T) {
	projectID := uuid.New()
	requestID := uuid.New()

	client := createInternalTestClient("http://localhost:9094")

	req := &shared.SyncRequest{
		RequestID: requestID,
		ProjectID: projectID,
		Direction: shared.SyncDirection("invalid"),
		LocalData: createInternalTestDataSet(),
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	_, err := client.Sync(ctx, req)

	if err == nil {
		t.Fatal("Expected error for unsupported direction, got nil")
	}

	if !containsString(err.Error(), "unsupported sync direction") {
		t.Errorf("Expected 'unsupported sync direction' error, got: %v", err)
	}
}

func TestClientSync_InvalidServerURL(t *testing.T) {
	projectID := uuid.New()
	requestID := uuid.New()

	// Create client with invalid URL
	client := createInternalTestClient(":invalid-url")

	req := &shared.SyncRequest{
		RequestID: requestID,
		ProjectID: projectID,
		Direction: shared.SyncLocalToMCP,
		LocalData: createInternalTestDataSet(),
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	_, err := client.Sync(ctx, req)

	if err == nil {
		t.Fatal("Expected error for invalid URL, got nil")
	}

	if !containsString(err.Error(), "failed to build endpoint URL") {
		t.Errorf("Expected 'failed to build endpoint URL' error, got: %v", err)
	}
}

func TestClientSync_ServerError(t *testing.T) {
	projectID := uuid.New()
	requestID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	client := createInternalTestClient(server.URL)

	req := &shared.SyncRequest{
		RequestID: requestID,
		ProjectID: projectID,
		Direction: shared.SyncLocalToMCP,
		LocalData: createInternalTestDataSet(),
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	_, err := client.Sync(ctx, req)

	if err == nil {
		t.Fatal("Expected error from server, got nil")
	}
}

func TestClientSync_InvalidResponseJSON(t *testing.T) {
	projectID := uuid.New()
	requestID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := createInternalTestClient(server.URL)

	req := &shared.SyncRequest{
		RequestID: requestID,
		ProjectID: projectID,
		Direction: shared.SyncLocalToMCP,
		LocalData: createInternalTestDataSet(),
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	_, err := client.Sync(ctx, req)

	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}

	if !containsString(err.Error(), "failed to deserialize response") {
		t.Errorf("Expected 'failed to deserialize response' error, got: %v", err)
	}
}

func TestClientSync_ContextCancellation(t *testing.T) {
	projectID := uuid.New()
	requestID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't respond, wait for context cancellation
		<-r.Context().Done()
	}))
	defer server.Close()

	client := createInternalTestClient(server.URL)

	req := &shared.SyncRequest{
		RequestID: requestID,
		ProjectID: projectID,
		Direction: shared.SyncLocalToMCP,
		LocalData: createInternalTestDataSet(),
		Timestamp: time.Now(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately
	cancel()

	_, err := client.Sync(ctx, req)

	if err == nil {
		t.Fatal("Expected error due to context cancellation, got nil")
	}
}
