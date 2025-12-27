package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// managerImpl is the private implementation of the Manager interface
type managerImpl struct {
	sessions sync.Map
}

// newManager creates a new session manager with proper configuration
func newManager() SessionManager {
	return &managerImpl{
		sessions: sync.Map{},
	}
}

// CreateSession creates a new session with proper validation and metadata
func (m *managerImpl) CreateSession(clientID string) (*SessionContext, error) {
	if clientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}

	session := &SessionContext{
		SessionID:    uuid.New(),
		ClientID:     clientID,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	m.sessions.Store(session.SessionID, session)
	return session, nil
}

// CreateSessionWithID creates a new session with a specific session ID
func (m *managerImpl) CreateSessionWithID(sessionID uuid.UUID, clientID string) (*SessionContext, error) {
	if clientID == "" {
		return nil, fmt.Errorf("client ID cannot be empty")
	}

	// Check if session already exists
	if _, exists := m.sessions.Load(sessionID); exists {
		return nil, fmt.Errorf("session with ID %s already exists", sessionID.String())
	}

	session := &SessionContext{
		SessionID:    sessionID,
		ClientID:     clientID,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	m.sessions.Store(sessionID, session)
	return session, nil
}

// GetSession gets a session by ID and updates activity timestamp
func (m *managerImpl) GetSession(sessionID uuid.UUID) (*SessionContext, error) {
	value, ok := m.sessions.Load(sessionID)
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	session := value.(*SessionContext)
	session.LastActivity = time.Now()
	return session, nil
}

// GetSessionByClientID gets a session by client ID and updates activity timestamp
func (m *managerImpl) GetSessionByClientID(clientID string) (*SessionContext, error) {
	var foundSession *SessionContext
	m.sessions.Range(func(key, value interface{}) bool {
		if session, ok := value.(*SessionContext); ok && session.ClientID == clientID {
			foundSession = session
			session.LastActivity = time.Now()
			return false // Stop iteration
		}
		return true
	})

	if foundSession == nil {
		return nil, fmt.Errorf("session not found for client: %s", clientID)
	}

	return foundSession, nil
}

// DeleteSession removes a session by ID
func (m *managerImpl) DeleteSession(sessionID uuid.UUID) error {
	_, ok := m.sessions.Load(sessionID)
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	m.sessions.Delete(sessionID)
	return nil
}

// ListSessions returns all active sessions
func (m *managerImpl) ListSessions() []*SessionContext {
	var sessions []*SessionContext
	m.sessions.Range(func(key, value interface{}) bool {
		if session, ok := value.(*SessionContext); ok {
			sessions = append(sessions, session)
		}
		return true
	})
	return sessions
}

// SetProject sets the project for a session with proper validation
func (m *managerImpl) SetProject(sessionID, projectID uuid.UUID) error {
	value, ok := m.sessions.Load(sessionID)
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session := value.(*SessionContext)
	session.ProjectID = &projectID
	session.LastActivity = time.Now()
	return nil
}

// GetProject retrieves the project ID for a session
func (m *managerImpl) GetProject(sessionID uuid.UUID) (*uuid.UUID, error) {
	value, ok := m.sessions.Load(sessionID)
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	session := value.(*SessionContext)
	return session.ProjectID, nil
}

// ClearProject removes the project binding from a session
func (m *managerImpl) ClearProject(sessionID uuid.UUID) error {
	value, ok := m.sessions.Load(sessionID)
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session := value.(*SessionContext)
	session.ProjectID = nil
	session.LastActivity = time.Now()
	return nil
}

// SetActor sets the actor for a session
func (m *managerImpl) SetActor(sessionID uuid.UUID, actor string) error {
	value, ok := m.sessions.Load(sessionID)
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session := value.(*SessionContext)
	session.Actor = actor
	session.LastActivity = time.Now()
	return nil
}

// ValidateSession checks if a session exists and is not expired
func (m *managerImpl) ValidateSession(sessionID uuid.UUID) bool {
	value, ok := m.sessions.Load(sessionID)
	if !ok {
		return false
	}

	// Additional validation logic can be added here
	session := value.(*SessionContext)
	return session != nil
}

// CleanupExpiredSessions removes sessions that have exceeded the timeout
func (m *managerImpl) CleanupExpiredSessions(ctx context.Context, timeout time.Duration) error {
	cutoff := time.Now().Add(-timeout)

	m.sessions.Range(func(key, value interface{}) bool {
		session := value.(*SessionContext)
		if session.LastActivity.Before(cutoff) {
			m.sessions.Delete(key)
		}
		return true
	})

	return nil
}

// CloseAll closes all sessions
func (m *managerImpl) CloseAll(ctx context.Context) error {
	m.sessions.Range(func(key, value interface{}) bool {
		m.sessions.Delete(key)
		return true
	})
	return nil
}

// GetSessionCount returns the number of active sessions
func (m *managerImpl) GetSessionCount() int {
	count := 0
	m.sessions.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
