package utils

import (
	"context"
	"errors"
	"testing"

	"github.com/denkhaus/knot/v2/internal/mcp/shared"
	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetSelectedProject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionManager := mocks.NewMockSessionManager(ctrl)

	t.Run("returns project ID when project selected", func(t *testing.T) {
		projectID := uuid.New()
		sessionUUID := uuid.New()

		sess := &session.SessionContext{
			SessionID: sessionUUID,
			ClientID:  sessionUUID.String(),
			ProjectID: &projectID,
		}

		mockSessionManager.EXPECT().GetSessionByClientID(sessionUUID.String()).Return(sess, nil)

		ctx := shared.ContextWithTestSession(sessionUUID)
		result, err := GetSelectedProject(ctx, mockSessionManager)

		require.NoError(t, err)
		assert.Equal(t, projectID.String(), result)
	})

	t.Run("returns error when no project selected", func(t *testing.T) {
		sessionUUID := uuid.New()

		sess := &session.SessionContext{
			SessionID: sessionUUID,
			ClientID:  sessionUUID.String(),
			ProjectID: nil,
		}

		mockSessionManager.EXPECT().GetSessionByClientID(sessionUUID.String()).Return(sess, nil)

		ctx := shared.ContextWithTestSession(sessionUUID)
		result, err := GetSelectedProject(ctx, mockSessionManager)

		require.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "no project selected")
		assert.Contains(t, err.Error(), "project_select")
	})

	t.Run("returns error when session not found", func(t *testing.T) {
		sessionUUID := uuid.New()

		mockSessionManager.EXPECT().GetSessionByClientID(sessionUUID.String()).Return(nil, errors.New("session not found"))

		ctx := shared.ContextWithTestSession(sessionUUID)
		result, err := GetSelectedProject(ctx, mockSessionManager)

		require.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "failed to get session")
	})

	t.Run("handles invalid session context", func(t *testing.T) {
		// Context without session info
		ctx := context.Background()

		result, err := GetSelectedProject(ctx, mockSessionManager)

		require.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "no session found")
	})
}

func TestGetSelectedProject_ErrorMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionManager := mocks.NewMockSessionManager(ctrl)

	t.Run("no project selected error is user-friendly", func(t *testing.T) {
		sessionUUID := uuid.New()

		sess := &session.SessionContext{
			SessionID: sessionUUID,
			ClientID:  sessionUUID.String(),
			ProjectID: nil,
		}

		mockSessionManager.EXPECT().GetSessionByClientID(sessionUUID.String()).Return(sess, nil)

		ctx := shared.ContextWithTestSession(sessionUUID)
		_, err := GetSelectedProject(ctx, mockSessionManager)

		require.Error(t, err)
		errMsg := err.Error()
		assert.Contains(t, errMsg, "no project selected")
		assert.Contains(t, errMsg, "project_select")
	})

	t.Run("session not found error is descriptive", func(t *testing.T) {
		sessionUUID := uuid.New()

		mockSessionManager.EXPECT().GetSessionByClientID(sessionUUID.String()).Return(nil, errors.New("session not found"))

		ctx := shared.ContextWithTestSession(sessionUUID)
		_, err := GetSelectedProject(ctx, mockSessionManager)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get session")
	})
}

func TestGetSelectedProject_WithRealProjectID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionManager := mocks.NewMockSessionManager(ctrl)

	t.Run("returns valid project ID string", func(t *testing.T) {
		projectID := uuid.New()
		sessionUUID := uuid.New()

		sess := &session.SessionContext{
			SessionID: sessionUUID,
			ClientID:  sessionUUID.String(),
			ProjectID: &projectID,
		}

		mockSessionManager.EXPECT().GetSessionByClientID(sessionUUID.String()).Return(sess, nil)

		ctx := shared.ContextWithTestSession(sessionUUID)
		result, err := GetSelectedProject(ctx, mockSessionManager)

		require.NoError(t, err)
		assert.NotEmpty(t, result)

		// Verify it's a valid UUID
		parsedUUID, err := uuid.Parse(result)
		require.NoError(t, err)
		assert.Equal(t, projectID, parsedUUID)
	})
}

func TestGetSelectedProject_Consistency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionManager := mocks.NewMockSessionManager(ctrl)

	t.Run("consistent results for same session", func(t *testing.T) {
		projectID := uuid.New()
		sessionUUID := uuid.New()

		sess := &session.SessionContext{
			SessionID: sessionUUID,
			ClientID:  sessionUUID.String(),
			ProjectID: &projectID,
		}

		mockSessionManager.EXPECT().GetSessionByClientID(sessionUUID.String()).Return(sess, nil).AnyTimes()

		ctx := shared.ContextWithTestSession(sessionUUID)
		result1, err1 := GetSelectedProject(ctx, mockSessionManager)
		result2, err2 := GetSelectedProject(ctx, mockSessionManager)

		assert.Equal(t, err1, err2)
		assert.Equal(t, result1, result2)
	})
}
