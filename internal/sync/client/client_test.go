// Package client_test provides tests for the sync client.
//
// IMPORTANT: This test file uses a separate package (client_test) instead of the
// standard client package to avoid import cycle issues. The internal/mocks package
// imports internal/sync, which imports internal/sync/client, creating a circular
// dependency if client tests try to import mocks. Using the _test package suffix
// breaks this cycle and allows proper use of the generated mocks.
//
// NOTE: This file only tests the public RESTSyncClient interface (GetRemoteData).
// Internal/private methods are tested in client_internal_test.go which uses the
// same package and can access private members.
package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/mocks"
	syncclient "github.com/denkhaus/knot/v2/internal/sync/client"
	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// testLogger is a minimal logger implementation for testing
type testLogger struct {
	ctrl *gomock.Controller
}

func newTestLogger(ctrl *gomock.Controller) *testLogger {
	return &testLogger{ctrl: ctrl}
}

func (l *testLogger) Debug(msg string, fields ...zap.Field) {}
func (l *testLogger) Info(msg string, fields ...zap.Field)  {}
func (l *testLogger) Warn(msg string, fields ...zap.Field)  {}
func (l *testLogger) Error(msg string, fields ...zap.Field) {}
func (l *testLogger) Sync()                                 {}
func (l *testLogger) With(fields ...zap.Field) logger.Logger {
	return l
}

func (l *testLogger) Named(name string) logger.Logger {
	return l
}

func (l *testLogger) ToZap() *zap.Logger {
	return zap.NewNop()
}
func (l *testLogger) SetLevel(level string) {}

// createTestClient creates a test client with the given server URL
func createTestClient(t *testing.T, serverURL string) syncclient.RESTSyncClient {
	t.Helper()

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

	ctrl := gomock.NewController(t)
	cfgSvc := mocks.NewMockService(ctrl)
	cfgSvc.EXPECT().GetSyncConfig().Return(cfg).AnyTimes()

	injector := do.New()
	do.ProvideValue[config.Service](injector, cfgSvc)
	do.ProvideValue[logger.Logger](injector, newTestLogger(ctrl))
	do.Provide(injector, shared.NewSerializerProvider)

	client, err := syncclient.NewRESTSyncClient(injector)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	return client
}

// Test helper to create test project
func createTestProject(id uuid.UUID) *types.Project {
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
func createTestTask(id, projectID uuid.UUID) *types.Task {
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
func createTestDataSet() *shared.SyncDataSet {
	projectID := uuid.New()
	taskID := uuid.New()

	project := createTestProject(projectID)
	task := createTestTask(taskID, projectID)

	return &shared.SyncDataSet{
		Projects: map[uuid.UUID]*types.Project{
			project.ID: project,
		},
		Tasks: map[uuid.UUID]*types.Task{
			task.ID: task,
		},
	}
}

func TestNewRESTSyncClient(t *testing.T) {
	injector := do.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := newTestLogger(ctrl)
	cfgSvc := mocks.NewMockService(ctrl)
	cfg := &config.SyncConfig{
		ServerURL:       "http://localhost:9094",
		Timeout:         30 * time.Second,
		RetryAttempts:   3,
		RetryDelay:      1 * time.Second,
		MaxRetryDelay:   30 * time.Second,
		MaxIdleConns:    10,
		IdleConnTimeout: 90 * time.Second,
	}

	cfgSvc.EXPECT().GetSyncConfig().Return(cfg).AnyTimes()

	do.ProvideValue[config.Service](injector, cfgSvc)
	do.ProvideValue[logger.Logger](injector, log)
	do.Provide(injector, shared.NewSerializerProvider)

	client, err := syncclient.NewRESTSyncClient(injector)
	if err != nil {
		t.Fatalf("Failed to create REST sync client: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	// Verify the client implements the interface
	_, ok := client.(syncclient.RESTSyncClient)
	if !ok {
		t.Error("Client should implement RESTSyncClient interface")
	}
}

func TestClientGetRemoteData_Success(t *testing.T) {
	projectID := uuid.New()
	expectedDataSet := createTestDataSet()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check projects endpoint
		if r.URL.Path == "/api/sync/projects" {
			projectsJSON, _ := json.Marshal(map[string]interface{}{
				"projects": expectedDataSet.Projects,
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(projectsJSON)
			return
		}

		// Check tasks endpoint
		if r.URL.Path == "/api/sync/tasks" {
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

	client := createTestClient(t, server.URL)
	ctx := context.Background()

	result, err := client.GetRemoteData(ctx, &projectID)
	if err != nil {
		t.Fatalf("GetRemoteData failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if len(result.Projects) != len(expectedDataSet.Projects) {
		t.Errorf("Expected %d projects, got %d", len(expectedDataSet.Projects), len(result.Projects))
	}

	if len(result.Tasks) != len(expectedDataSet.Tasks) {
		t.Errorf("Expected %d tasks, got %d", len(expectedDataSet.Tasks), len(result.Tasks))
	}
}

func TestClientGetRemoteData_WithoutProjectID(t *testing.T) {
	expectedDataSet := createTestDataSet()

	// Create test server that only returns projects
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

	client := createTestClient(t, server.URL)
	ctx := context.Background()

	result, err := client.GetRemoteData(ctx, nil)
	if err != nil {
		t.Fatalf("GetRemoteData failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if len(result.Projects) != len(expectedDataSet.Projects) {
		t.Errorf("Expected %d projects, got %d", len(expectedDataSet.Projects), len(result.Projects))
	}

	// Tasks should be empty since no project ID was provided
	if len(result.Tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(result.Tasks))
	}
}

func TestClientGetRemoteData_WithAuthToken(t *testing.T) {
	expectedToken := "test-auth-token-12345"
	expectedDataSet := createTestDataSet()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Authorization header on all requests
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+expectedToken {
			t.Errorf("Expected Authorization header 'Bearer %s', got %s", expectedToken, auth)
		}

		if r.URL.Path == "/api/sync/projects" {
			projectsJSON, _ := json.Marshal(map[string]interface{}{
				"projects": expectedDataSet.Projects,
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(projectsJSON)
			return
		}

		if r.URL.Path == "/api/sync/tasks" {
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

	cfg := &config.SyncConfig{
		ServerURL:       server.URL,
		Timeout:         5 * time.Second,
		RetryAttempts:   2,
		RetryDelay:      100 * time.Millisecond,
		MaxRetryDelay:   500 * time.Millisecond,
		MaxIdleConns:    5,
		IdleConnTimeout: 30 * time.Second,
		AuthToken:       expectedToken,
	}

	ctrl := gomock.NewController(t)
	cfgSvc := mocks.NewMockService(ctrl)
	cfgSvc.EXPECT().GetSyncConfig().Return(cfg).AnyTimes()

	injector := do.New()
	do.ProvideValue[config.Service](injector, cfgSvc)
	do.ProvideValue[logger.Logger](injector, newTestLogger(ctrl))
	do.Provide(injector, shared.NewSerializerProvider)

	client, err := syncclient.NewRESTSyncClient(injector)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	ctx := context.Background()
	projectID := uuid.New()

	result, err := client.GetRemoteData(ctx, &projectID)
	if err != nil {
		t.Fatalf("GetRemoteData failed: %v", err)
	}

	if len(result.Projects) != len(expectedDataSet.Projects) {
		t.Errorf("Expected %d projects, got %d", len(expectedDataSet.Projects), len(result.Projects))
	}
}

func TestClientGetRemoteData_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block indefinitely
		select {}
	}))
	defer server.Close()

	client := createTestClient(t, server.URL)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	_, err := client.GetRemoteData(ctx, nil)
	if err == nil {
		t.Fatal("Expected context cancellation error, got nil")
	}

	if !containsString(err.Error(), "context canceled") {
		t.Errorf("Expected context canceled error, got: %v", err)
	}
}

func TestClientGetRemoteData_NetworkError(t *testing.T) {
	// Use an invalid server URL that will cause a network error
	client := createTestClient(t, "http://localhost:9999") // Unlikely to be running
	ctx := context.Background()

	_, err := client.GetRemoteData(ctx, nil)
	if err == nil {
		t.Fatal("Expected network error, got nil")
	}
}

func TestClientGetRemoteData_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return invalid JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	client := createTestClient(t, server.URL)
	ctx := context.Background()

	_, err := client.GetRemoteData(ctx, nil)
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}

	if !containsString(err.Error(), "failed to deserialize") {
		t.Errorf("Expected deserialization error, got: %v", err)
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
