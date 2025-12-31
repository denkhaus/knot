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
		clientID := "test-user"
		session, err := manager.CreateSession(clientID)
		require.NoError(t, err)
		require.NotNil(t, session)

		assert.NotEqual(t, uuid.Nil, session.SessionID)
		assert.Equal(t, clientID, session.ClientID)
		assert.False(t, session.CreatedAt.IsZero())
		assert.False(t, session.LastActivity.IsZero())
		assert.Nil(t, session.ProjectID)
	})

	t.Run("Create multiple sessions", func(t *testing.T) {
		clientID1 := "user1"
		clientID2 := "user2"

		session1, err := manager.CreateSession(clientID1)
		require.NoError(t, err)

		session2, err := manager.CreateSession(clientID2)
		require.NoError(t, err)

		// Verify sessions are different
		assert.NotEqual(t, session1.SessionID, session2.SessionID)
		assert.Equal(t, clientID1, session1.ClientID)
		assert.Equal(t, clientID2, session2.ClientID)

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
		clientID := "test-user"
		originalSession, err := manager.CreateSession(clientID)
		require.NoError(t, err)

		// Retrieve the session
		retrievedSession, err := manager.GetSession(originalSession.SessionID)
		require.NoError(t, err)
		require.NotNil(t, retrievedSession)

		assert.Equal(t, originalSession.SessionID, retrievedSession.SessionID)
		assert.Equal(t, originalSession.ClientID, retrievedSession.ClientID)
		assert.Equal(t, clientID, retrievedSession.ClientID)
	})

	t.Run("Get non-existent session", func(t *testing.T) {
		nonExistentID := uuid.New()
		session, err := manager.GetSession(nonExistentID)
		assert.Error(t, err) // Now returns error instead of nil, nil
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("Last activity updated on retrieval", func(t *testing.T) {
		clientID := "test-user"
		session, err := manager.CreateSession(clientID)
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
		clientID := "test-user"
		session, err := manager.CreateSession(clientID)
		require.NoError(t, err)

		projectID := uuid.New()
		err = manager.SetProject(session.SessionID, projectID)
		require.NoError(t, err)

		// Verify project was set
		retrievedSession, err := manager.GetSession(session.SessionID)
		require.NoError(t, err)
		require.NotNil(t, retrievedSession.ProjectID)
		assert.Equal(t, projectID, *retrievedSession.ProjectID)

		// Verify last activity was updated (allow for small time differences)
		assert.True(t, retrievedSession.LastActivity.After(session.LastActivity) ||
			retrievedSession.LastActivity.Equal(session.LastActivity))
	})

	t.Run("Set project for non-existent session", func(t *testing.T) {
		nonExistentID := uuid.New()
		projectID := uuid.New()

		err := manager.SetProject(nonExistentID, projectID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("Change project for session", func(t *testing.T) {
		clientID := "test-user"
		session, err := manager.CreateSession(clientID)
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
		clientIDs := []string{"user1", "user2", "user3"}
		sessionIDs := make([]uuid.UUID, 0, len(clientIDs))

		for _, clientID := range clientIDs {
			session, err := manager.CreateSession(clientID)
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
			assert.Error(t, err) // Now returns error since session was deleted
			assert.Nil(t, session)
			assert.Contains(t, err.Error(), "session not found")
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
		clientID := "test-client"
		projectID := uuid.New()
		now := time.Now()

		session := &SessionContext{
			SessionID:    sessionID,
			ClientID:     clientID,
			ProjectID:    &projectID,
			CreatedAt:    now,
			LastActivity: now,
		}

		assert.Equal(t, sessionID, session.SessionID)
		assert.Equal(t, clientID, session.ClientID)
		require.NotNil(t, session.ProjectID)
		assert.Equal(t, projectID, *session.ProjectID)
		assert.Equal(t, now, session.CreatedAt)
		assert.Equal(t, now, session.LastActivity)
	})

	t.Run("SessionContext without project", func(t *testing.T) {
		sessionID := uuid.New()
		clientID := "test-client"
		now := time.Now()

		session := &SessionContext{
			SessionID:    sessionID,
			ClientID:     clientID,
			ProjectID:    nil,
			CreatedAt:    now,
			LastActivity: now,
		}

		assert.Equal(t, sessionID, session.SessionID)
		assert.Equal(t, clientID, session.ClientID)
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
	wg.Add(numGoroutines)

	// Test concurrent session creation
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			clientID := fmt.Sprintf("user-%d", id)
			for j := 0; j < numOperations; j++ {
				session, err := manager.CreateSession(clientID + fmt.Sprintf("-session-%d", j))
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
		assert.NotEmpty(t, session.ClientID)
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
			clientID := fmt.Sprintf("user-%d", i)
			session, err := manager.CreateSession(clientID)
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
		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "client ID cannot be empty")
	})

	t.Run("Special characters in user ID", func(t *testing.T) {
		clientID := "user-with-special-chars_123@#$%"
		session, err := manager.CreateSession(clientID)
		require.NoError(t, err)
		assert.Equal(t, clientID, session.ClientID)

		retrieved, err := manager.GetSession(session.SessionID)
		require.NoError(t, err)
		assert.Equal(t, clientID, retrieved.ClientID)
	})

	t.Run("Very long user ID", func(t *testing.T) {
		clientID := string(make([]byte, 1000))
		for i := range clientID {
			clientID = clientID[:i] + "a" + clientID[i+1:]
		}

		session, err := manager.CreateSession(clientID)
		require.NoError(t, err)
		assert.Equal(t, clientID, session.ClientID)
	})
}

func TestCreateSessionWithID(t *testing.T) {
	manager := newManager()

	t.Run("Create session with specific ID", func(t *testing.T) {
		sessionID := uuid.New()
		clientID := "test-user"

		session, err := manager.CreateSessionWithID(sessionID, clientID)
		require.NoError(t, err)
		require.NotNil(t, session)

		assert.Equal(t, sessionID, session.SessionID)
		assert.Equal(t, clientID, session.ClientID)
		assert.False(t, session.CreatedAt.IsZero())
		assert.False(t, session.LastActivity.IsZero())
	})

	t.Run("Create session with existing ID returns error", func(t *testing.T) {
		sessionID := uuid.New()
		clientID := "test-user"

		// Create first session
		_, err := manager.CreateSessionWithID(sessionID, clientID)
		require.NoError(t, err)

		// Try to create another session with same ID
		_, err = manager.CreateSessionWithID(sessionID, "another-user")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestGetSessionByClientID(t *testing.T) {
	manager := newManager()

	t.Run("Get session by client ID", func(t *testing.T) {
		clientID := "test-user"
		originalSession, err := manager.CreateSession(clientID)
		require.NoError(t, err)

		// Retrieve by client ID
		retrievedSession, err := manager.GetSessionByClientID(clientID)
		require.NoError(t, err)
		require.NotNil(t, retrievedSession)

		assert.Equal(t, originalSession.SessionID, retrievedSession.SessionID)
		assert.Equal(t, clientID, retrievedSession.ClientID)
	})

	t.Run("Get session by non-existent client ID", func(t *testing.T) {
		session, err := manager.GetSessionByClientID("non-existent-user")
		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestDeleteSession(t *testing.T) {
	manager := newManager()

	t.Run("Delete existing session", func(t *testing.T) {
		clientID := "test-user"
		session, err := manager.CreateSession(clientID)
		require.NoError(t, err)

		sessionID := session.SessionID

		// Verify session exists
		_, err = manager.GetSession(sessionID)
		require.NoError(t, err)

		// Delete the session
		err = manager.DeleteSession(sessionID)
		require.NoError(t, err)

		// Verify session is gone
		_, err = manager.GetSession(sessionID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Delete non-existent session", func(t *testing.T) {
		nonExistentID := uuid.New()

		err := manager.DeleteSession(nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestGetProject(t *testing.T) {
	manager := newManager()

	t.Run("Get project for session with project", func(t *testing.T) {
		clientID := "test-user"
		session, err := manager.CreateSession(clientID)
		require.NoError(t, err)

		projectID := uuid.New()
		err = manager.SetProject(session.SessionID, projectID)
		require.NoError(t, err)

		// Get project
		retrievedProjectID, err := manager.GetProject(session.SessionID)
		require.NoError(t, err)
		require.NotNil(t, retrievedProjectID)
		assert.Equal(t, projectID, *retrievedProjectID)
	})

	t.Run("Get project for session without project", func(t *testing.T) {
		clientID := "test-user"
		session, err := manager.CreateSession(clientID)
		require.NoError(t, err)

		// Get project when none is set - returns nil, nil
		projectID, err := manager.GetProject(session.SessionID)
		require.NoError(t, err)
		assert.Nil(t, projectID)
	})

	t.Run("Get project for non-existent session", func(t *testing.T) {
		nonExistentID := uuid.New()

		projectID, err := manager.GetProject(nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, projectID)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestClearProject(t *testing.T) {
	manager := newManager()

	t.Run("Clear project from session", func(t *testing.T) {
		clientID := "test-user"
		session, err := manager.CreateSession(clientID)
		require.NoError(t, err)

		projectID := uuid.New()
		err = manager.SetProject(session.SessionID, projectID)
		require.NoError(t, err)

		// Verify project is set
		retrievedSession, err := manager.GetSession(session.SessionID)
		require.NoError(t, err)
		require.NotNil(t, retrievedSession.ProjectID)

		// Clear the project
		err = manager.ClearProject(session.SessionID)
		require.NoError(t, err)

		// Verify project is cleared
		retrievedSession, err = manager.GetSession(session.SessionID)
		require.NoError(t, err)
		assert.Nil(t, retrievedSession.ProjectID)
	})

	t.Run("Clear project from session without project", func(t *testing.T) {
		clientID := "test-user"
		session, err := manager.CreateSession(clientID)
		require.NoError(t, err)

		// Clear project when none is set (should succeed without error)
		err = manager.ClearProject(session.SessionID)
		require.NoError(t, err)
	})

	t.Run("Clear project for non-existent session", func(t *testing.T) {
		nonExistentID := uuid.New()

		err := manager.ClearProject(nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestSetActor(t *testing.T) {
	manager := newManager()

	t.Run("Set actor for session", func(t *testing.T) {
		clientID := "test-user"
		session, err := manager.CreateSession(clientID)
		require.NoError(t, err)

		actor := "test-actor"
		err = manager.SetActor(session.SessionID, actor)
		require.NoError(t, err)

		// Verify actor was set
		retrievedSession, err := manager.GetSession(session.SessionID)
		require.NoError(t, err)
		assert.Equal(t, actor, retrievedSession.Actor)
	})

	t.Run("Change actor for session", func(t *testing.T) {
		clientID := "test-user"
		session, err := manager.CreateSession(clientID)
		require.NoError(t, err)

		// Set initial actor
		actor1 := "actor1"
		err = manager.SetActor(session.SessionID, actor1)
		require.NoError(t, err)

		// Change to different actor
		actor2 := "actor2"
		err = manager.SetActor(session.SessionID, actor2)
		require.NoError(t, err)

		// Verify actor was changed
		retrievedSession, err := manager.GetSession(session.SessionID)
		require.NoError(t, err)
		assert.Equal(t, actor2, retrievedSession.Actor)
		assert.NotEqual(t, actor1, retrievedSession.Actor)
	})

	t.Run("Set actor for non-existent session", func(t *testing.T) {
		nonExistentID := uuid.New()

		err := manager.SetActor(nonExistentID, "test-actor")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestValidateSession(t *testing.T) {
	manager := newManager()

	t.Run("Validate existing session", func(t *testing.T) {
		clientID := "test-user"
		session, err := manager.CreateSession(clientID)
		require.NoError(t, err)

		// Validate session
		valid := manager.ValidateSession(session.SessionID)
		assert.True(t, valid, "existing session should be valid")
	})

	t.Run("Validate non-existent session", func(t *testing.T) {
		nonExistentID := uuid.New()

		valid := manager.ValidateSession(nonExistentID)
		assert.False(t, valid, "non-existent session should not be valid")
	})
}

func TestCleanupExpiredSessions(t *testing.T) {
	manager := newManager()

	t.Run("Cleanup removes expired sessions", func(t *testing.T) {
		// This test verifies the cleanup function can be called successfully
		// Actual session expiration depends on the implementation
		ctx := context.Background()
		timeout := 1 * time.Hour

		// Create some sessions
		for i := 0; i < 5; i++ {
			_, err := manager.CreateSession(fmt.Sprintf("user-%d", i))
			require.NoError(t, err)
		}

		initialCount := manager.GetSessionCount()
		assert.Equal(t, 5, initialCount)

		// Run cleanup - may not remove any since sessions are recent
		err := manager.CleanupExpiredSessions(ctx, timeout)
		require.NoError(t, err)
	})
}

func TestGetSessionCount(t *testing.T) {
	manager := newManager()

	t.Run("Count increases with sessions", func(t *testing.T) {
		count := manager.GetSessionCount()
		assert.Equal(t, 0, count)

		// Add sessions
		for i := 0; i < 3; i++ {
			_, err := manager.CreateSession(fmt.Sprintf("user-%d", i))
			require.NoError(t, err)
		}

		count = manager.GetSessionCount()
		assert.Equal(t, 3, count)
	})

	t.Run("Count decreases when sessions deleted", func(t *testing.T) {
		// Use a fresh manager to avoid interference from other tests
		freshManager := newManager()

		// Create sessions
		sessionIDs := make([]uuid.UUID, 0, 3)
		for i := 0; i < 3; i++ {
			session, err := freshManager.CreateSession(fmt.Sprintf("user-%d", i))
			require.NoError(t, err)
			sessionIDs = append(sessionIDs, session.SessionID)
		}

		initialCount := freshManager.GetSessionCount()
		assert.Equal(t, 3, initialCount)

		// Delete one session
		err := freshManager.DeleteSession(sessionIDs[0])
		require.NoError(t, err)

		count := freshManager.GetSessionCount()
		assert.Equal(t, 2, count)
	})
}
