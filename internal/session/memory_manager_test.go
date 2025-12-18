package session

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	manager := newManager()
	require.NotNil(t, manager)
	// Test that we can use the manager interface
	assert.Equal(t, 0, manager.GetSessionCount())
}

func TestCreateSession(t *testing.T) {
	manager := newManager()

	t.Run("Create session successfully", func(t *testing.T) {
		userID := "test-user"
		session, err := manager.CreateSession(userID)
		require.NoError(t, err)
		require.NotNil(t, session)

		assert.NotEqual(t, uuid.Nil, session.SessionID)
		assert.Equal(t, userID, session.UserID)
		assert.False(t, session.CreatedAt.IsZero())
		assert.False(t, session.LastActivity.IsZero())
		assert.Nil(t, session.ProjectID)
	})

	t.Run("Create multiple sessions", func(t *testing.T) {
		userID1 := "user1"
		userID2 := "user2"

		session1, err := manager.CreateSession(userID1)
		require.NoError(t, err)

		session2, err := manager.CreateSession(userID2)
		require.NoError(t, err)

		// Verify sessions are different
		assert.NotEqual(t, session1.SessionID, session2.SessionID)
		assert.Equal(t, userID1, session1.UserID)
		assert.Equal(t, userID2, session2.UserID)

		// Verify both sessions are stored
		retrieved1, err := manager.GetSession(session1.SessionID)
		require.NoError(t, err)
		assert.Equal(t, session1.SessionID, retrieved1.SessionID)

		retrieved2, err := manager.GetSession(session2.SessionID)
		require.NoError(t, err)
		assert.Equal(t, session2.SessionID, retrieved2.SessionID)
	})
}

func TestGetSession(t *testing.T) {
	manager := newManager()

	t.Run("Get existing session", func(t *testing.T) {
		userID := "test-user"
		originalSession, err := manager.CreateSession(userID)
		require.NoError(t, err)

		// Retrieve the session
		retrievedSession, err := manager.GetSession(originalSession.SessionID)
		require.NoError(t, err)
		require.NotNil(t, retrievedSession)

		assert.Equal(t, originalSession.SessionID, retrievedSession.SessionID)
		assert.Equal(t, originalSession.UserID, retrievedSession.UserID)
		assert.Equal(t, userID, retrievedSession.UserID)
	})

	t.Run("Get non-existent session", func(t *testing.T) {
		nonExistentID := uuid.New()
		session, err := manager.GetSession(nonExistentID)
		require.NoError(t, err) // Currently returns nil, nil instead of error
		assert.Nil(t, session)
	})

	t.Run("Last activity updated on retrieval", func(t *testing.T) {
		userID := "test-user"
		session, err := manager.CreateSession(userID)
		require.NoError(t, err)

		originalActivity := session.LastActivity

		// Wait a bit to ensure timestamp difference
		time.Sleep(10 * time.Millisecond)

		// Retrieve session
		retrievedSession, err := manager.GetSession(session.SessionID)
		require.NoError(t, err)

		// Last activity should be updated
		assert.True(t, retrievedSession.LastActivity.After(originalActivity))
	})
}

func TestSetProject(t *testing.T) {
	manager := newManager()

	t.Run("Set project for existing session", func(t *testing.T) {
		userID := "test-user"
		session, err := manager.CreateSession(userID)
		require.NoError(t, err)

		projectID := uuid.New()
		err = manager.SetProject(session.SessionID, projectID)
		require.NoError(t, err)

		// Verify project was set
		retrievedSession, err := manager.GetSession(session.SessionID)
		require.NoError(t, err)
		require.NotNil(t, retrievedSession.ProjectID)
		assert.Equal(t, projectID, *retrievedSession.ProjectID)

		// Verify last activity was updated
		assert.True(t, retrievedSession.LastActivity.After(session.LastActivity))
	})

	t.Run("Set project for non-existent session", func(t *testing.T) {
		nonExistentID := uuid.New()
		projectID := uuid.New()

		err := manager.SetProject(nonExistentID, projectID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("Change project for session", func(t *testing.T) {
		userID := "test-user"
		session, err := manager.CreateSession(userID)
		require.NoError(t, err)

		// Set initial project
		projectID1 := uuid.New()
		err = manager.SetProject(session.SessionID, projectID1)
		require.NoError(t, err)

		// Change to different project
		projectID2 := uuid.New()
		err = manager.SetProject(session.SessionID, projectID2)
		require.NoError(t, err)

		// Verify project was changed
		retrievedSession, err := manager.GetSession(session.SessionID)
		require.NoError(t, err)
		require.NotNil(t, retrievedSession.ProjectID)
		assert.Equal(t, projectID2, *retrievedSession.ProjectID)
		assert.NotEqual(t, projectID1, *retrievedSession.ProjectID)
	})
}

func TestCloseAll(t *testing.T) {
	manager := newManager()

	t.Run("Close sessions", func(t *testing.T) {
		// Create multiple sessions
		userIDs := []string{"user1", "user2", "user3"}
		sessionIDs := make([]uuid.UUID, 0, len(userIDs))

		for _, userID := range userIDs {
			session, err := manager.CreateSession(userID)
			require.NoError(t, err)
			sessionIDs = append(sessionIDs, session.SessionID)

			// Set some projects
			if len(sessionIDs) > 1 {
				projectID := uuid.New()
				err = manager.SetProject(session.SessionID, projectID)
				require.NoError(t, err)
			}
		}

		// Verify sessions exist
		for _, sessionID := range sessionIDs {
			session, err := manager.GetSession(sessionID)
			require.NoError(t, err)
			assert.NotNil(t, session)
		}

		// Close all sessions
		ctx := context.Background()
		err := manager.CloseAll(ctx)
		require.NoError(t, err)

		// Verify sessions are gone
		for _, sessionID := range sessionIDs {
			session, err := manager.GetSession(sessionID)
			require.NoError(t, err) // Still returns nil, nil instead of error
			assert.Nil(t, session)
		}
	})

	t.Run("Close empty manager", func(t *testing.T) {
		emptyManager := newManager()
		ctx := context.Background()
		err := emptyManager.CloseAll(ctx)
		require.NoError(t, err)
	})
}

func TestSessionContext(t *testing.T) {
	t.Run("SessionContext fields", func(t *testing.T) {
		sessionID := uuid.New()
		userID := "test-user"
		projectID := uuid.New()
		now := time.Now()

		session := &SessionContext{
			SessionID:    sessionID,
			UserID:       userID,
			ProjectID:    &projectID,
			CreatedAt:    now,
			LastActivity: now,
		}

		assert.Equal(t, sessionID, session.SessionID)
		assert.Equal(t, userID, session.UserID)
		require.NotNil(t, session.ProjectID)
		assert.Equal(t, projectID, *session.ProjectID)
		assert.Equal(t, now, session.CreatedAt)
		assert.Equal(t, now, session.LastActivity)
	})

	t.Run("SessionContext without project", func(t *testing.T) {
		sessionID := uuid.New()
		userID := "test-user"
		now := time.Now()

		session := &SessionContext{
			SessionID:    sessionID,
			UserID:       userID,
			ProjectID:    nil,
			CreatedAt:    now,
			LastActivity: now,
		}

		assert.Equal(t, sessionID, session.SessionID)
		assert.Equal(t, userID, session.UserID)
		assert.Nil(t, session.ProjectID)
		assert.Equal(t, now, session.CreatedAt)
		assert.Equal(t, now, session.LastActivity)
	})
}

func TestConcurrentAccess(t *testing.T) {
	manager := newManager()
	numGoroutines := 100
	numOperations := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3) // Create, Get, Set operations

	// Test concurrent session creation
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			userID := fmt.Sprintf("user-%d", id)
			for j := 0; j < numOperations; j++ {
				session, err := manager.CreateSession(userID + fmt.Sprintf("-session-%d", j))
				require.NoError(t, err)
				require.NotNil(t, session)

				// Test concurrent retrieval
				retrieved, err := manager.GetSession(session.SessionID)
				require.NoError(t, err)
				assert.Equal(t, session.SessionID, retrieved.SessionID)

				// Test concurrent project setting
				if j%2 == 0 {
					projectID := uuid.New()
					err = manager.SetProject(session.SessionID, projectID)
					require.NoError(t, err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify no data races occurred by checking some sessions
	sessions := manager.ListSessions()
	for _, session := range sessions {
		assert.NotEqual(t, uuid.Nil, session.SessionID)
		assert.NotEmpty(t, session.UserID)
	}
}

// Test that sessions can handle garbage collection scenarios
func TestMemoryManagement(t *testing.T) {
	manager := newManager()

	t.Run("Large number of sessions", func(t *testing.T) {
		numSessions := 1000
		sessionIDs := make([]uuid.UUID, 0, numSessions)

		// Create many sessions
		for i := 0; i < numSessions; i++ {
			userID := fmt.Sprintf("user-%d", i)
			session, err := manager.CreateSession(userID)
			require.NoError(t, err)
			sessionIDs = append(sessionIDs, session.SessionID)

			// Set project for some sessions
			if i%3 == 0 {
				projectID := uuid.New()
				err = manager.SetProject(session.SessionID, projectID)
				require.NoError(t, err)
			}
		}

		// Verify a sample of sessions
		for i := 0; i < 10; i++ {
			idx := i * 100
			if idx < len(sessionIDs) {
				session, err := manager.GetSession(sessionIDs[idx])
				require.NoError(t, err)
				assert.NotNil(t, session)
			}
		}

		// Clean up
		ctx := context.Background()
		err := manager.CloseAll(ctx)
		require.NoError(t, err)
	})
}

func TestEdgeCases(t *testing.T) {
	manager := newManager()

	t.Run("Empty user ID", func(t *testing.T) {
		session, err := manager.CreateSession("")
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, "", session.UserID)
	})

	t.Run("Special characters in user ID", func(t *testing.T) {
		userID := "user-with-special-chars_123@#$%"
		session, err := manager.CreateSession(userID)
		require.NoError(t, err)
		assert.Equal(t, userID, session.UserID)

		retrieved, err := manager.GetSession(session.SessionID)
		require.NoError(t, err)
		assert.Equal(t, userID, retrieved.UserID)
	})

	t.Run("Very long user ID", func(t *testing.T) {
		userID := string(make([]byte, 1000))
		for i := range userID {
			userID = userID[:i] + "a" + userID[i+1:]
		}

		session, err := manager.CreateSession(userID)
		require.NoError(t, err)
		assert.Equal(t, userID, session.UserID)
	})
}
