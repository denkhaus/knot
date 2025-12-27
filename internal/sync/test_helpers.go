package sync

import (
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// testTime is a helper time for testing
var testTime = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

// noopLogger is a simple logger implementation for testing
type noopLogger struct{}

func (l *noopLogger) Debug(msg string, fields ...zap.Field)  {}
func (l *noopLogger) Info(msg string, fields ...zap.Field)   {}
func (l *noopLogger) Warn(msg string, fields ...zap.Field)   {}
func (l *noopLogger) Error(msg string, fields ...zap.Field)  {}
func (l *noopLogger) Sync()                                  {}
func (l *noopLogger) With(fields ...zap.Field) logger.Logger { return l }
func (l *noopLogger) Named(name string) logger.Logger        { return l }
func (l *noopLogger) ToZap() *zap.Logger                     { return zap.NewNop() }
func (l *noopLogger) SetLevel(logLevel string)               {}

// createTestSyncProject creates a test Project for testing
func createTestSyncProject(t *testing.T) *types.Project {
	id := uuid.New()
	return &types.Project{
		ID:          id,
		Title:       "Test Project",
		Description: "Test Description",
		State:       types.ProjectStateActive,
		CreatedAt:   testTime,
		UpdatedAt:   testTime,
		CreatedBy:   "test-user",
		UpdatedBy:   "test-user",
	}
}

// createTestSyncTask creates a test Task for testing
func createTestSyncTask(t *testing.T) *types.Task {
	id := uuid.New()
	projectID := uuid.New()
	return &types.Task{
		ID:          id,
		ProjectID:   projectID,
		Title:       "Test Task",
		Description: "Test Description",
		State:       types.TaskStatePending,
		Priority:    types.TaskPriority(5),
		Complexity:  3,
		CreatedAt:   testTime,
		UpdatedAt:   testTime,
		CreatedBy:   "test-user",
		UpdatedBy:   "test-user",
	}
}

// createTestSyncConfig creates a test SyncConfig for testing
func createTestSyncConfig() *config.SyncConfig {
	return &config.SyncConfig{
		ServerURL:        "http://localhost:8080",
		AuthToken:        "test-token",
		ConflictStrategy: "last-writer-wins",
		Timeout:          30 * time.Second,
		BatchSize:        100,
		RetryAttempts:    3,
		RetryDelay:       5 * time.Second,
	}
}
