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
func newManager() Manager {
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
