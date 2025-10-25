package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/denkhaus/knot/internal/manager"
	"github.com/denkhaus/knot/internal/repository/inmemory"
	"github.com/denkhaus/knot/internal/repository/sqlite"
	"github.com/denkhaus/knot/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// TestConfig holds configuration for tests
type TestConfig struct {
	UseInMemoryDB bool
	TempDir       string
	Logger        *zap.Logger
}

// NewTestConfig creates a new test configuration
func NewTestConfig(t *testing.T) *TestConfig {
	return &TestConfig{
		UseInMemoryDB: true, // Default to in-memory for speed
		Logger:        zaptest.NewLogger(t),
	}
}

// WithSQLiteDB configures the test to use SQLite database
func (tc *TestConfig) WithSQLiteDB() *TestConfig {
	tc.UseInMemoryDB = false
	return tc
}

// WithTempDir sets a custom temp directory
func (tc *TestConfig) WithTempDir(dir string) *TestConfig {
	tc.TempDir = dir
	return tc
}

// SetupTestRepository creates a test repository based on configuration
func (tc *TestConfig) SetupTestRepository(t *testing.T) types.Repository {
	if tc.UseInMemoryDB {
		return inmemory.NewMemoryRepository()
	}

	// Setup temporary directory for SQLite tests
	tempDir := tc.TempDir
	if tempDir == "" {
		var err error
		tempDir, err = os.MkdirTemp("", "knot_test_*")
		require.NoError(t, err)
		t.Cleanup(func() {
			os.RemoveAll(tempDir)
		})
	}

	// Change to temp directory for SQLite database creation
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	
	err = os.Chdir(tempDir)
	require.NoError(t, err)
	
	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})

	// Create SQLite repository
	repo, err := sqlite.NewRepository(
		sqlite.WithLogger(tc.Logger),
		sqlite.WithAutoMigrate(true),
	)
	require.NoError(t, err)

	// Note: Repository interface doesn't have Close method
	// SQLite repository will be cleaned up when temp directory is removed

	return repo
}

// SetupTestManager creates a test project manager
func (tc *TestConfig) SetupTestManager(t *testing.T) manager.ProjectManager {
	repo := tc.SetupTestRepository(t)
	config := manager.DefaultConfig()
	return manager.NewManagerWithRepository(repo, config)
}

// CreateTestProject creates a test project for testing
func CreateTestProject(t *testing.T, mgr manager.ProjectManager) *types.Project {
	ctx := context.Background()
	
	project, err := mgr.CreateProject(ctx, "Test Project", "Test Description", "test-user")
	require.NoError(t, err)
	require.NotNil(t, project)
	
	return project
}

// CreateTestTask creates a test task for testing
func CreateTestTask(t *testing.T, mgr manager.ProjectManager, projectID uuid.UUID) *types.Task {
	ctx := context.Background()
	
	task, err := mgr.CreateTask(ctx, projectID, nil, "Test Task", "Test Description", 5, types.TaskPriorityMedium, "test-user")
	require.NoError(t, err)
	require.NotNil(t, task)
	
	return task
}

// CreateTestTaskWithParent creates a test task with a parent
func CreateTestTaskWithParent(t *testing.T, mgr manager.ProjectManager, projectID, parentID uuid.UUID) *types.Task {
	ctx := context.Background()
	
	task, err := mgr.CreateTask(ctx, projectID, &parentID, "Test Subtask", "Test Subtask Description", 3, types.TaskPriorityMedium, "test-user")
	require.NoError(t, err)
	require.NotNil(t, task)
	
	return task
}

// AssertTaskState asserts that a task has the expected state
func AssertTaskState(t *testing.T, mgr manager.ProjectManager, taskID uuid.UUID, expectedState types.TaskState) {
	ctx := context.Background()
	
	task, err := mgr.GetTask(ctx, taskID)
	require.NoError(t, err)
	require.Equal(t, expectedState, task.State)
}

// AssertTaskCount asserts the number of tasks in a project
func AssertTaskCount(t *testing.T, mgr manager.ProjectManager, projectID uuid.UUID, expectedCount int) {
	ctx := context.Background()
	
	tasks, err := mgr.ListTasksForProject(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, tasks, expectedCount)
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for condition: %s", message)
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// TempFile creates a temporary file for testing
func TempFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "knot_test_*.tmp")
	require.NoError(t, err)
	
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	
	err = tmpFile.Close()
	require.NoError(t, err)
	
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})
	
	return tmpFile.Name()
}

// TempDir creates a temporary directory for testing
func TempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "knot_test_*")
	require.NoError(t, err)
	
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	
	return dir
}

// MockAppContext provides a mock implementation for testing
type MockAppContext struct {
	Manager manager.ProjectManager
	Logger  *zap.Logger
	Actor   string
}

func (m *MockAppContext) SetActor(actor string) {
	m.Actor = actor
}

func (m *MockAppContext) ProjectManager() manager.ProjectManager {
	return m.Manager
}