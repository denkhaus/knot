package mocks_test

import (
	"context"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

// ExampleTest demonstrates how to use the central MockProjectManager
// This is an example for reference in other test files
func ExampleTest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a new mock instance
	mockMgr := mocks.NewMockProjectManager(ctrl)

	// Set up expectations
	mockMgr.EXPECT().
		ListProjects(gomock.Any()).
		Return([]*types.Project{
			{
				ID:          uuid.New(),
				Title:       "Test Project",
				Description: "A test project",
				State:       types.ProjectStateActive,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}, nil)

	// Use the mock in your test
	appCtx := &shared.AppContext{
		ProjectManager: mockMgr,
		Logger:         zaptest.NewLogger(t),
	}

	// Test code that uses the appCtx
	projects, err := appCtx.ProjectManager.ListProjects(context.Background())
	require.NoError(t, err)
	assert.Len(t, projects, 1)
	assert.Equal(t, "Test Project", projects[0].Title)
}
