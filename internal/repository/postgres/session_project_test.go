package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/repository/ent/project"
	"github.com/denkhaus/knot/v2/internal/repository/ent/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// setupTestPostgresRepository creates a test repository with PostgreSQL
func setupTestPostgresRepository(t *testing.T) (*postgresRepository, func()) {
	// Use test database configuration from environment or default to test instance
	dsn := "postgres://postgres:password@localhost:5432/knot_test?sslmode=disable"

	// Create repository with test configuration
	repo, err := NewRepository(dsn,
		WithAutoMigrate(true),
		WithLogger(zaptest.NewLogger(t)),
		WithConnectionPool(5, 2),
	)
	require.NoError(t, err)

	postgresRepo, ok := repo.(*postgresRepository)
	require.True(t, ok, "Repository should be of type *postgresRepository")

	// Return cleanup function
	cleanup := func() {
		if err := postgresRepo.Close(); err != nil {
			t.Logf("Warning: failed to close repository: %v", err)
		}
	}

	return postgresRepo, cleanup
}

// TestPostgresRepository_ListSessions_ProjectEagerLoading tests that ListSessions properly loads project data
func TestPostgresRepository_ListSessions_ProjectEagerLoading(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepository(t)
	defer cleanup()

	ctx := context.Background()
	clientID := uuid.New().String()

	// Create a test project first
	testProjectID := uuid.New()
	err := repo.client.Project.Create().
		SetID(testProjectID).
		SetTitle("Test Project for Session Loading").
		SetDescription("Project created to test session-project eager loading").
		SetState(project.StateActive).
		SetCreatedAt(time.Now()).
		Exec(ctx)
	require.NoError(t, err)

	// Create multiple sessions, some with projects, some without
	sessionIDs := make([]uuid.UUID, 3)

	// Session 1: No project (oldest)
	sessionIDs[0] = uuid.New()
	_, err = repo.client.Session.Create().
		SetID(sessionIDs[0]).
		SetClientID(clientID).
		SetCreatedAt(time.Now().Add(-3 * time.Hour)).
		SetLastActivity(time.Now().Add(-3 * time.Hour)).
		SetStatus(session.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	// Session 2: With project (most recent)
	sessionIDs[1] = uuid.New()
	_, err = repo.client.Session.Create().
		SetID(sessionIDs[1]).
		SetClientID(clientID).
		SetCreatedAt(time.Now().Add(-2 * time.Hour)).
		SetLastActivity(time.Now().Add(-1 * time.Hour)). // Most recent
		SetStatus(session.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	// Set project for session 2
	err = repo.SetSessionProject(ctx, sessionIDs[1], testProjectID)
	require.NoError(t, err)

	// Session 3: No project (middle)
	sessionIDs[2] = uuid.New()
	_, err = repo.client.Session.Create().
		SetID(sessionIDs[2]).
		SetClientID(clientID).
		SetCreatedAt(time.Now().Add(-1 * time.Hour)).
		SetLastActivity(time.Now().Add(-2 * time.Hour)).
		SetStatus(session.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	// Test ListSessions with client ID filter
	clientSessions, err := repo.ListSessions(ctx, clientID)
	require.NoError(t, err)
	require.Len(t, clientSessions, 3, "Should return exactly 3 sessions for the client")

	// Verify sessions are ordered by LastActivity DESC (most recent first)
	assert.Equal(t, sessionIDs[1], clientSessions[0].ID, "First session should be most recent (with project)")
	assert.Equal(t, sessionIDs[2], clientSessions[1].ID, "Second session should be middle one")
	assert.Equal(t, sessionIDs[0], clientSessions[2].ID, "Third session should be oldest")

	// CRITICAL TEST: Verify project data is eagerly loaded
	assert.NotNil(t, clientSessions[0].ProjectID, "Most recent session should have project loaded")
	assert.Equal(t, testProjectID, *clientSessions[0].ProjectID, "Project ID should match test project")

	// Other sessions should have nil ProjectID
	assert.Nil(t, clientSessions[1].ProjectID, "Middle session should have no project")
	assert.Nil(t, clientSessions[2].ProjectID, "Oldest session should have no project")
}

// TestPostgresRepository_SetSessionProject_GetSessionProject tests the complete project setting workflow
func TestPostgresRepository_SetSessionProject_GetSessionProject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepository(t)
	defer cleanup()

	ctx := context.Background()
	clientID := uuid.New().String()

	// Create a test project
	testProjectID := uuid.New()
	err := repo.client.Project.Create().
		SetID(testProjectID).
		SetTitle("Test Project for Session Operations").
		SetDescription("Project created to test session-project operations").
		SetState(project.StateActive).
		SetCreatedAt(time.Now()).
		Exec(ctx)
	require.NoError(t, err)

	// Create a session without project
	sessionID := uuid.New()
	session, err := repo.CreateSession(ctx, clientID)
	require.NoError(t, err)
	sessionID = session.ID

	// Verify session initially has no project
	retrievedProjectID, err := repo.GetSessionProject(ctx, sessionID)
	require.NoError(t, err)
	assert.Nil(t, retrievedProjectID, "Session should initially have no project")

	// Set project for session
	err = repo.SetSessionProject(ctx, sessionID, testProjectID)
	require.NoError(t, err)

	// Verify project is now set
	retrievedProjectID, err = repo.GetSessionProject(ctx, sessionID)
	require.NoError(t, err)
	require.NotNil(t, retrievedProjectID, "Session should now have a project")
	assert.Equal(t, testProjectID, *retrievedProjectID, "Retrieved project ID should match")

	// Test ListSessions to ensure project is eagerly loaded
	sessions, err := repo.ListSessions(ctx, clientID)
	require.NoError(t, err)
	require.Len(t, sessions, 1, "Should have exactly one session")

	// Verify project data is eagerly loaded
	assert.NotNil(t, sessions[0].ProjectID, "Session should have project loaded via eager loading")
	assert.Equal(t, testProjectID, *sessions[0].ProjectID, "Eager-loaded project ID should match")

	// Test clearing project
	err = repo.ClearSessionProject(ctx, sessionID)
	require.NoError(t, err)

	// Verify project is cleared
	retrievedProjectID, err = repo.GetSessionProject(ctx, sessionID)
	require.NoError(t, err)
	assert.Nil(t, retrievedProjectID, "Session should no longer have a project")
}


// TestPostgresRepository_ListSessions_MultipleSessionsProjectPriority tests GetSessionByClientID behavior
func TestPostgresRepository_ListSessions_MultipleSessionsProjectPriority(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}

	repo, cleanup := setupTestPostgresRepository(t)
	defer cleanup()

	ctx := context.Background()
	clientID := uuid.New().String()

	// Create two projects
	project1ID := uuid.New()
	err := repo.client.Project.Create().
		SetID(project1ID).
		SetTitle("Project 1").
		SetDescription("First test project").
		SetState(project.StateActive).
		SetCreatedAt(time.Now()).
		Exec(ctx)
	require.NoError(t, err)

	project2ID := uuid.New()
	err = repo.client.Project.Create().
		SetID(project2ID).
		SetTitle("Project 2").
		SetDescription("Second test project").
		SetState(project.StateActive).
		SetCreatedAt(time.Now()).
		Exec(ctx)
	require.NoError(t, err)

	// Create sessions with different project associations and timestamps
	sessionNoProject := uuid.New()
	_, err = repo.client.Session.Create().
		SetID(sessionNoProject).
		SetClientID(clientID).
		SetCreatedAt(time.Now().Add(-3 * time.Hour)).
		SetLastActivity(time.Now().Add(-3 * time.Hour)). // Oldest
		SetStatus(session.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	sessionWithProject1 := uuid.New()
	_, err = repo.client.Session.Create().
		SetID(sessionWithProject1).
		SetClientID(clientID).
		SetCreatedAt(time.Now().Add(-2 * time.Hour)).
		SetLastActivity(time.Now().Add(-2 * time.Hour)).
		SetStatus(session.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	err = repo.SetSessionProject(ctx, sessionWithProject1, project1ID)
	require.NoError(t, err)

	// This should be the most recent session with a project
	sessionWithProject2 := uuid.New()
	_, err = repo.client.Session.Create().
		SetID(sessionWithProject2).
		SetClientID(clientID).
		SetCreatedAt(time.Now().Add(-1 * time.Hour)).
		SetLastActivity(time.Now().Add(-1 * time.Hour)). // Most recent
		SetStatus(session.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	err = repo.SetSessionProject(ctx, sessionWithProject2, project2ID)
	require.NoError(t, err)

	// Test ListSessions
	sessions, err := repo.ListSessions(ctx, clientID)
	require.NoError(t, err)
	require.Len(t, sessions, 3, "Should return exactly 3 sessions")

	// Verify ordering (most recent by LastActivity first)
	assert.Equal(t, sessionWithProject2, sessions[0].ID, "First session should be most recent")
	assert.Equal(t, sessionWithProject1, sessions[1].ID, "Second session should be middle")
	assert.Equal(t, sessionNoProject, sessions[2].ID, "Third session should be oldest")

	// CRITICAL: Verify all project data is eagerly loaded correctly
	// First session should have project 2
	assert.NotNil(t, sessions[0].ProjectID, "First session should have project")
	assert.Equal(t, project2ID, *sessions[0].ProjectID, "First session should have project 2")

	// Second session should have project 1
	assert.NotNil(t, sessions[1].ProjectID, "Second session should have project")
	assert.Equal(t, project1ID, *sessions[1].ProjectID, "Second session should have project 1")

	// Third session should have no project
	assert.Nil(t, sessions[2].ProjectID, "Third session should have no project")
}