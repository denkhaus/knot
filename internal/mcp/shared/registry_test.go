package shared_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/mcp/shared"
	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/server"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewSessionRegistry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("successful creation", func(t *testing.T) {
		mockSessionManager := mocks.NewMockSessionManager(ctrl)
		mockLogger := mocks.NewMockLogger(ctrl)
		mcpServer := server.NewMCPServer("test", "1.0.0")

		injector := do.New()
		do.ProvideValue[session.SessionManager](injector, mockSessionManager)
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue[*server.MCPServer](injector, mcpServer)

		registry, err := shared.NewSessionRegistry(injector)

		require.NoError(t, err)
		require.NotNil(t, registry)
	})

	t.Run("verify initialization", func(t *testing.T) {
		mockSessionManager := mocks.NewMockSessionManager(ctrl)
		mockLogger := mocks.NewMockLogger(ctrl)
		mcpServer := server.NewMCPServer("test", "1.0.0")

		injector := do.New()
		do.ProvideValue[session.SessionManager](injector, mockSessionManager)
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue[*server.MCPServer](injector, mcpServer)
	})
}

func TestSessionRegistry_GetOrCreateSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	mcpServer := server.NewMCPServer("test", "1.0.0")

	// Allow any number of info/debug calls from the registry
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	injector := do.New()
	do.ProvideValue[session.SessionManager](injector, mockSessionManager)
	do.ProvideValue[logger.Logger](injector, mockLogger)
	do.ProvideValue[*server.MCPServer](injector, mcpServer)
	registry, _ := shared.NewSessionRegistry(injector)

	ctx := context.Background()

	t.Run("creates new session", func(t *testing.T) {
		testUUID := uuid.New()
		newSession := &session.SessionContext{
			SessionID: testUUID,
			ClientID:  testUUID.String(),
		}

		mockSessionManager.EXPECT().GetSession(testUUID).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().GetSessionByClientID(testUUID.String()).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().CreateSessionWithID(testUUID, testUUID.String()).Return(newSession, nil)

		mcpSession, err := registry.GetOrCreateSession(ctx, testUUID)

		require.NoError(t, err)
		require.NotNil(t, mcpSession)
		assert.Equal(t, testUUID.String(), mcpSession.SessionID())
	})

	t.Run("retrieves existing session by ID", func(t *testing.T) {
		testUUID := uuid.New()
		existingSession := &session.SessionContext{
			SessionID: testUUID,
			ClientID:  testUUID.String(),
		}

		mockSessionManager.EXPECT().GetSession(testUUID).Return(existingSession, nil)

		mcpSession, err := registry.GetOrCreateSession(ctx, testUUID)

		require.NoError(t, err)
		require.NotNil(t, mcpSession)
	})

	t.Run("retrieves existing session by client ID", func(t *testing.T) {
		testUUID := uuid.New()
		existingSession := &session.SessionContext{
			SessionID: uuid.New(),
			ClientID:  testUUID.String(),
		}

		mockSessionManager.EXPECT().GetSession(testUUID).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().GetSessionByClientID(testUUID.String()).Return(existingSession, nil)

		mcpSession, err := registry.GetOrCreateSession(ctx, testUUID)

		require.NoError(t, err)
		require.NotNil(t, mcpSession)
	})

	t.Run("returns cached session on subsequent calls", func(t *testing.T) {
		testUUID := uuid.New()
		newSession := &session.SessionContext{
			SessionID: testUUID,
			ClientID:  testUUID.String(),
		}

		mockSessionManager.EXPECT().GetSession(testUUID).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().GetSessionByClientID(testUUID.String()).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().CreateSessionWithID(testUUID, testUUID.String()).Return(newSession, nil)

		// First call
		mcpSession1, err := registry.GetOrCreateSession(ctx, testUUID)
		require.NoError(t, err)
		require.NotNil(t, mcpSession1)

		// Second call should use cache
		mcpSession2, err := registry.GetOrCreateSession(ctx, testUUID)
		require.NoError(t, err)
		assert.Same(t, mcpSession1, mcpSession2)
	})

	t.Run("session creation failure", func(t *testing.T) {
		testUUID := uuid.New()
		mockSessionManager.EXPECT().GetSession(testUUID).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().GetSessionByClientID(testUUID.String()).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().CreateSessionWithID(testUUID, testUUID.String()).Return(nil, assert.AnError)

		mcpSession, err := registry.GetOrCreateSession(ctx, testUUID)

		require.Error(t, err)
		assert.Nil(t, mcpSession)
	})

	t.Run("copies project ID from internal session", func(t *testing.T) {
		testUUID := uuid.New()
		projectID := uuid.New()
		existingSession := &session.SessionContext{
			SessionID: testUUID,
			ClientID:  testUUID.String(),
			ProjectID: &projectID,
		}

		mockSessionManager.EXPECT().GetSession(testUUID).Return(existingSession, nil)

		mcpSession, err := registry.GetOrCreateSession(ctx, testUUID)

		require.NoError(t, err)
		assert.Equal(t, &projectID, mcpSession.GetProjectID())
	})
}

func TestSessionRegistry_GetSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	mcpServer := server.NewMCPServer("test", "1.0.0")

	// Allow any number of info/debug calls from the registry
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	injector := do.New()
	do.ProvideValue[session.SessionManager](injector, mockSessionManager)
	do.ProvideValue[logger.Logger](injector, mockLogger)
	do.ProvideValue[*server.MCPServer](injector, mcpServer)
	registry, _ := shared.NewSessionRegistry(injector)

	t.Run("retrieves cached session", func(t *testing.T) {
		ctx := context.Background()
		testUUID := uuid.New()

		// First create a session
		newSession := &session.SessionContext{
			SessionID: testUUID,
			ClientID:  testUUID.String(),
		}

		mockSessionManager.EXPECT().GetSession(testUUID).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().GetSessionByClientID(testUUID.String()).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().CreateSessionWithID(testUUID, testUUID.String()).Return(newSession, nil)

		mcpSession1, _ := registry.GetOrCreateSession(ctx, testUUID)

		// Then retrieve it
		mcpSession2, err := registry.GetSession(testUUID)

		require.NoError(t, err)
		assert.Same(t, mcpSession1, mcpSession2)
	})

	t.Run("creates session if not in cache", func(t *testing.T) {
		testUUID := uuid.New()
		existingSession := &session.SessionContext{
			SessionID: testUUID,
			ClientID:  testUUID.String(),
		}

		mockSessionManager.EXPECT().GetSession(testUUID).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().GetSessionByClientID(testUUID.String()).Return(existingSession, nil)

		mcpSession, err := registry.GetSession(testUUID)

		require.NoError(t, err)
		require.NotNil(t, mcpSession)
	})

	t.Run("creates new session if none exists", func(t *testing.T) {
		testUUID := uuid.New()
		newSession := &session.SessionContext{
			SessionID: testUUID,
			ClientID:  testUUID.String(),
		}

		mockSessionManager.EXPECT().GetSession(testUUID).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().GetSessionByClientID(testUUID.String()).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().CreateSessionWithID(testUUID, testUUID.String()).Return(newSession, nil)

		mcpSession, err := registry.GetSession(testUUID)

		require.NoError(t, err)
		require.NotNil(t, mcpSession)
	})
}

func TestSessionRegistry_RemoveSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	mcpServer := server.NewMCPServer("test", "1.0.0")

	// Allow any number of info/debug calls from the registry
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	injector := do.New()
	do.ProvideValue[session.SessionManager](injector, mockSessionManager)
	do.ProvideValue[logger.Logger](injector, mockLogger)
	do.ProvideValue[*server.MCPServer](injector, mcpServer)
	registry, _ := shared.NewSessionRegistry(injector)

	ctx := context.Background()
	testUUID := uuid.New()

	t.Run("removes existing session", func(t *testing.T) {
		// Create a session first
		newSession := &session.SessionContext{
			SessionID: testUUID,
			ClientID:  testUUID.String(),
		}

		mockSessionManager.EXPECT().GetSession(testUUID).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().GetSessionByClientID(testUUID.String()).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().CreateSessionWithID(testUUID, testUUID.String()).Return(newSession, nil)
		registry.GetOrCreateSession(ctx, testUUID)

		// Then remove it
		mockSessionManager.EXPECT().DeleteSession(testUUID).Return(nil)

		err := registry.RemoveSession(ctx, testUUID)

		require.NoError(t, err)
	})

	t.Run("removes non-existent session", func(t *testing.T) {
		mockSessionManager.EXPECT().DeleteSession(testUUID).Return(assert.AnError)
		mockLogger.EXPECT().Warn("Failed to delete internal session",
			logger.String("session_id", testUUID.String()),
			logger.Error(assert.AnError),
		)

		err := registry.RemoveSession(ctx, testUUID)

		require.NoError(t, err) // Should not error even if deletion fails
	})
}

func TestSessionRegistry_CleanupExpiredSessions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	mcpServer := server.NewMCPServer("test", "1.0.0")

	injector := do.New()
	do.ProvideValue[session.SessionManager](injector, mockSessionManager)
	do.ProvideValue[logger.Logger](injector, mockLogger)
	do.ProvideValue[*server.MCPServer](injector, mcpServer)
	registry, _ := shared.NewSessionRegistry(injector)

	ctx := context.Background()
	timeout := 30 * time.Minute

	t.Run("delegates to session manager", func(t *testing.T) {
		mockSessionManager.EXPECT().CleanupExpiredSessions(ctx, timeout).Return(nil)

		err := registry.CleanupExpiredSessions(ctx, timeout)

		require.NoError(t, err)
	})

	t.Run("returns session manager error", func(t *testing.T) {
		mockSessionManager.EXPECT().CleanupExpiredSessions(ctx, timeout).Return(assert.AnError)

		err := registry.CleanupExpiredSessions(ctx, timeout)

		require.Error(t, err)
		assert.Same(t, assert.AnError, err)
	})
}

func TestSessionRegistry_GetSessionCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	mcpServer := server.NewMCPServer("test", "1.0.0")

	// Allow any number of info/debug calls from the registry
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	injector := do.New()
	do.ProvideValue[session.SessionManager](injector, mockSessionManager)
	do.ProvideValue[logger.Logger](injector, mockLogger)
	do.ProvideValue[*server.MCPServer](injector, mcpServer)
	registry, _ := shared.NewSessionRegistry(injector)

	ctx := context.Background()

	t.Run("count increases with sessions", func(t *testing.T) {
		count := registry.GetSessionCount()
		assert.Equal(t, 0, count)

		// Add a session
		testUUID := uuid.New()
		newSession := &session.SessionContext{
			SessionID: testUUID,
			ClientID:  testUUID.String(),
		}

		mockSessionManager.EXPECT().GetSession(testUUID).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().GetSessionByClientID(testUUID.String()).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().CreateSessionWithID(testUUID, testUUID.String()).Return(newSession, nil)

		registry.GetOrCreateSession(ctx, testUUID)

		count = registry.GetSessionCount()
		assert.Equal(t, 1, count)
	})

	t.Run("count decreases when sessions removed", func(t *testing.T) {
		// Get the starting count (should be 1 from previous test)
		startCount := registry.GetSessionCount()

		// Add a new session
		testUUID := uuid.New()
		newSession := &session.SessionContext{
			SessionID: testUUID,
			ClientID:  testUUID.String(),
		}

		mockSessionManager.EXPECT().GetSession(testUUID).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().GetSessionByClientID(testUUID.String()).Return(nil, assert.AnError)
		mockSessionManager.EXPECT().CreateSessionWithID(testUUID, testUUID.String()).Return(newSession, nil)

		registry.GetOrCreateSession(ctx, testUUID)

		// Count should have increased by 1
		countAfterAdd := registry.GetSessionCount()
		assert.Equal(t, startCount+1, countAfterAdd)

		// Remove it
		mockSessionManager.EXPECT().DeleteSession(testUUID).Return(nil)
		registry.RemoveSession(ctx, testUUID)

		// Count should be back to starting count
		count := registry.GetSessionCount()
		assert.Equal(t, startCount, count)
	})
}

func TestSessionRegistry_SyncExistingSessions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	mcpServer := server.NewMCPServer("test", "1.0.0")

	// Allow any number of info/debug calls from the registry
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	injector := do.New()
	do.ProvideValue[session.SessionManager](injector, mockSessionManager)
	do.ProvideValue[logger.Logger](injector, mockLogger)
	do.ProvideValue[*server.MCPServer](injector, mcpServer)
	registry, _ := shared.NewSessionRegistry(injector)

	ctx := context.Background()
	testUUID := uuid.New()

	t.Run("syncs existing sessions", func(t *testing.T) {
		existingSession := &session.SessionContext{
			SessionID: testUUID,
			ClientID:  testUUID.String(),
		}

		mockSessionManager.EXPECT().ListSessions().Return([]*session.SessionContext{existingSession})

		err := registry.SyncExistingSessions(ctx)

		require.NoError(t, err)
	})

	t.Run("handles session already registered error", func(t *testing.T) {
		// Use the same UUID as the first test - the session was already synced
		// So this sync should trigger the "already registered" Debug call
		existingSession := &session.SessionContext{
			SessionID: testUUID,
			ClientID:  testUUID.String(),
		}

		mockSessionManager.EXPECT().ListSessions().Return([]*session.SessionContext{existingSession})
		// Note: The Debug expectation is set up, but since we have Debug().AnyTimes() above,
		// this specific expectation might not be strictly required

		err := registry.SyncExistingSessions(ctx)

		// Should succeed even if some sessions are already registered
		require.NoError(t, err)
	})
}

func TestSessionRegistry_ConcurrentAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionManager := mocks.NewMockSessionManager(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	mcpServer := server.NewMCPServer("test", "1.0.0")

	// Allow any number of info/debug calls from the registry
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	injector := do.New()
	do.ProvideValue[session.SessionManager](injector, mockSessionManager)
	do.ProvideValue[logger.Logger](injector, mockLogger)
	do.ProvideValue(injector, mcpServer)
	registry, _ := shared.NewSessionRegistry(injector)

	ctx := context.Background()

	t.Run("concurrent session creation", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				testUUID := uuid.New()
				newSession := &session.SessionContext{
					SessionID: testUUID,
					ClientID:  testUUID.String(),
				}

				mockSessionManager.EXPECT().GetSession(testUUID).Return(nil, assert.AnError).AnyTimes()
				mockSessionManager.EXPECT().GetSessionByClientID(testUUID.String()).Return(nil, assert.AnError).AnyTimes()
				mockSessionManager.EXPECT().CreateSessionWithID(testUUID, testUUID.String()).Return(newSession, nil).AnyTimes()

				_, err := registry.GetOrCreateSession(ctx, testUUID)
				assert.NoError(t, err)
			}(i)
		}

		wg.Wait()

		// Should have all sessions
		count := registry.GetSessionCount()
		assert.Equal(t, numGoroutines, count)
	})
}
