package manager_test

import (
	"context"
	"testing"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestCheckCircularDependency tests the circular dependency detection function
func TestCheckCircularDependency(t *testing.T) {
	ctx := context.Background()

	t.Run("detects self-dependency", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProjectManager := mocks.NewMockProjectManager(ctrl)
		taskA := uuid.New()

		isCircular, path := manager.CheckCircularDependency(ctx, mockProjectManager, taskA, taskA)

		assert.True(t, isCircular)
		assert.Len(t, path, 2)
		assert.Equal(t, taskA, path[0])
		assert.Equal(t, taskA, path[1])
	})

	t.Run("returns false when GetTaskDependencies fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProjectManager := mocks.NewMockProjectManager(ctrl)
		taskA := uuid.New()
		taskB := uuid.New()

		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskA).
			Return(nil, assert.AnError)

		isCircular, path := manager.CheckCircularDependency(ctx, mockProjectManager, taskA, taskB)

		assert.False(t, isCircular)
		assert.Nil(t, path)
	})

	t.Run("returns false for no circular dependency", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProjectManager := mocks.NewMockProjectManager(ctrl)
		taskA := uuid.New()
		taskB := uuid.New()

		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskA).
			Return([]*types.Task{}, nil).AnyTimes()
		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskB).
			Return([]*types.Task{}, nil).AnyTimes()

		isCircular, path := manager.CheckCircularDependency(ctx, mockProjectManager, taskA, taskB)

		assert.False(t, isCircular)
		assert.Nil(t, path)
	})

	t.Run("detects direct circular dependency (A -> B -> A)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProjectManager := mocks.NewMockProjectManager(ctrl)
		taskA := uuid.New()
		taskB := uuid.New()
		taskBDependsOnA := []*types.Task{{ID: taskA}}

		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskA).
			Return([]*types.Task{}, nil)
		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskB).
			Return(taskBDependsOnA, nil)

		isCircular, path := manager.CheckCircularDependency(ctx, mockProjectManager, taskA, taskB)

		assert.True(t, isCircular)
		assert.Len(t, path, 3)
		assert.Equal(t, taskA, path[0])
		assert.Equal(t, taskB, path[1])
		assert.Equal(t, taskA, path[2])
	})

	t.Run("returns false when no dependencies exist", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProjectManager := mocks.NewMockProjectManager(ctrl)
		taskA := uuid.New()
		taskB := uuid.New()
		taskBNoDeps := []*types.Task{}

		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskA).
			Return([]*types.Task{}, nil).AnyTimes()
		mockProjectManager.EXPECT().GetTaskDependencies(ctx, taskB).
			Return(taskBNoDeps, nil).AnyTimes()

		isCircular, path := manager.CheckCircularDependency(ctx, mockProjectManager, taskA, taskB)

		assert.False(t, isCircular)
		assert.Nil(t, path)
	})
}

// TestCircularDependencyError tests the error type
func TestCircularDependencyError(t *testing.T) {
	taskA := uuid.New()
	taskB := uuid.New()
	path := []uuid.UUID{taskA, taskB, taskA}

	err := &manager.CircularDependencyError{
		TaskID:    taskA,
		DependsOn: taskB,
		Path:      path,
	}

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency detected")
}

func TestCircularDependencyError_EmptyPath(t *testing.T) {
	taskA := uuid.New()
	taskB := uuid.New()

	err := &manager.CircularDependencyError{
		TaskID:    taskA,
		DependsOn: taskB,
		Path:      nil,
	}

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency detected")
	assert.Contains(t, err.Error(), taskA.String())
	assert.Contains(t, err.Error(), taskB.String())
}
