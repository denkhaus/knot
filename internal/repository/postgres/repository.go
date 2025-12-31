package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/denkhaus/knot/v2/internal/repository/ent"
	"github.com/denkhaus/knot/v2/internal/repository/ent/session"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
	_ "github.com/lib/pq"
)

// postgresRepository implements the Repository interface using ent ORM with PostgreSQL
type postgresRepository struct {
	client    *ent.Client
	db        *sql.DB // Store db reference for health checks
	logger    *zap.Logger
	dsn       string
	autoMigrate bool
	maxOpenConns int
	maxIdleConns int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
	migrationTimeout time.Duration
}

// NewRepository creates a new PostgreSQL repository using ent ORM
func NewRepository(dsn string, opts ...Option) (types.Repository, error) {
	repo := &postgresRepository{
		dsn:             dsn,
		logger:          zap.NewNop(), // No-op logger by default
		autoMigrate:     true,         // Auto-migrate by default
		maxOpenConns:     25,           // Default PostgreSQL connection pool
		maxIdleConns:     5,
		connMaxLifetime:  5 * time.Minute,
		connMaxIdleTime:  1 * time.Minute,
		migrationTimeout: 30 * time.Second,
	}

	// Apply options
	for _, opt := range opts {
		opt(repo)
	}

	if err := repo.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL repository: %w", err)
	}

	repo.logger.Info("PostgreSQL repository initialized successfully")
	return repo, nil
}

// initialize sets up the ent client and performs migrations
func (r *postgresRepository) initialize() error {
	r.logger.Info("initializing PostgreSQL database", zap.String("dsn", maskDSN(r.dsn)))

	db, err := sql.Open("postgres", r.dsn)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(r.maxOpenConns)
	db.SetMaxIdleConns(r.maxIdleConns)
	db.SetConnMaxLifetime(r.connMaxLifetime)
	db.SetConnMaxIdleTime(r.connMaxIdleTime)

	// Test connection with retry logic
	if err := r.testConnection(db); err != nil {
		return fmt.Errorf("failed to test PostgreSQL connection: %w", err)
	}

	// Create ent client with PostgreSQL driver
	drv := entsql.OpenDB(dialect.Postgres, db)

	// Create ent client with PostgreSQL options
	client := ent.NewClient(
		ent.Driver(drv),
		ent.Log(func(args ...interface{}) { r.logger.Sugar().Info(args...) }),
	)

	// Run auto-migration if enabled
	if r.autoMigrate {
		if err := r.migrateDatabase(context.Background(), client); err != nil {
			return fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	r.client = client
	r.db = db // Store db reference for health checks
	return nil
}

// testConnection tests the database connection with retry logic
func (r *postgresRepository) testConnection(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.migrationTimeout)
	defer cancel()

	var attempts int
	const maxAttempts = 3
	const retryDelay = 2 * time.Second

	for attempts < maxAttempts {
		err := db.PingContext(ctx)
		if err == nil {
			r.logger.Info("PostgreSQL connection test successful", zap.Int("attempt", attempts+1))
			return nil
		}

		attempts++
		if attempts < maxAttempts {
			r.logger.Warn("PostgreSQL connection test failed, retrying",
				zap.Int("attempt", attempts),
				zap.Duration("retry_delay", retryDelay),
				zap.Error(err))
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("failed to connect to PostgreSQL after %d attempts", maxAttempts)
}

// migrateDatabase runs database migrations
func (r *postgresRepository) migrateDatabase(ctx context.Context, client *ent.Client) error {
	r.logger.Info("running PostgreSQL database migrations")

	if err := client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	r.logger.Info("PostgreSQL database migrations completed successfully")
	return nil
}

// maskDSN masks sensitive information in connection string for logging
func maskDSN(dsn string) string {
	// Basic masking - hide password if present
	if len(dsn) > 20 {
		return dsn[:min(20, len(dsn))] + "***"
	}
	return "***"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Close closes the database connection
func (r *postgresRepository) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// Session operations - these methods implement the SessionRepository interface
// Only PostgreSQL repository implements these as sessions are only needed for MCP mode

// CreateSession creates a new session
func (r *postgresRepository) CreateSession(ctx context.Context, clientID string) (*types.Session, error) {
	sessionEnt, err := r.client.Session.Create().
		SetClientID(clientID).
		SetStatus(session.StatusActive).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	r.logger.Debug("Session created",
		zap.String("session_id", sessionEnt.ID.String()),
		zap.String("client_id", clientID))

	return r.convertSessionEntToTypes(sessionEnt), nil
}

// CreateSessionWithID creates a new session with a specific session ID
func (r *postgresRepository) CreateSessionWithID(ctx context.Context, sessionID uuid.UUID, clientID string) (*types.Session, error) {
	sessionEnt, err := r.client.Session.Create().
		SetID(sessionID).
		SetClientID(clientID).
		SetStatus(session.StatusActive).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create session with ID: %w", err)
	}

	r.logger.Debug("Session created with specific ID",
		zap.String("session_id", sessionEnt.ID.String()),
		zap.String("client_id", clientID))

	return r.convertSessionEntToTypes(sessionEnt), nil
}

// GetSession retrieves a session by ID and updates last activity
func (r *postgresRepository) GetSession(ctx context.Context, sessionID uuid.UUID) (*types.Session, error) {
	session, err := r.client.Session.Get(ctx, sessionID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Update last activity
	updatedSession, err := r.client.Session.UpdateOne(session).
		SetLastActivity(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update session activity: %w", err)
	}

	return r.convertSessionEntToTypes(updatedSession), nil
}

// DeleteSession removes a session by ID
func (r *postgresRepository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	err := r.client.Session.DeleteOneID(sessionID).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // Session already deleted
		}
		return fmt.Errorf("failed to delete session: %w", err)
	}

	r.logger.Debug("Session deleted",
		zap.String("session_id", sessionID.String()))

	return nil
}

// ListSessions returns all sessions for a client, or all sessions if clientID is empty
func (r *postgresRepository) ListSessions(ctx context.Context, clientID string) ([]*types.Session, error) {
	query := r.client.Session.Query()
	if clientID != "" {
		query = query.Where(session.ClientID(clientID))
	}

	// Order by last activity descending to get most recent session first
	query = query.Order(ent.Desc(session.FieldLastActivity))

	// Include project data in the query to ensure project is loaded
	query = query.WithProject()

	sessionEnts, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	result := make([]*types.Session, len(sessionEnts))
	for i, s := range sessionEnts {
		result[i] = r.convertSessionEntToTypes(s)
	}

	return result, nil
}

// UpdateSessionActivity updates the last activity timestamp
func (r *postgresRepository) UpdateSessionActivity(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.client.Session.UpdateOneID(sessionID).
		SetLastActivity(time.Now()).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("session not found")
		}
		return fmt.Errorf("failed to update session activity: %w", err)
	}

	return nil
}

// SetSessionProject associates a project with a session
func (r *postgresRepository) SetSessionProject(ctx context.Context, sessionID, projectID uuid.UUID) error {
	// First get the project to ensure it exists
	_, err := r.client.Project.Get(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	// Update the session to associate it with the project
	err = r.client.Session.UpdateOneID(sessionID).
		SetProjectID(projectID).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("session not found")
		}
		return fmt.Errorf("failed to set session project: %w", err)
	}

	r.logger.Debug("Project set for session",
		zap.String("session_id", sessionID.String()),
		zap.String("project_id", projectID.String()))

	return nil
}

// GetSessionProject retrieves the project ID associated with a session
func (r *postgresRepository) GetSessionProject(ctx context.Context, sessionID uuid.UUID) (*uuid.UUID, error) {
	sessionEnt, err := r.client.Session.Query().
		Where(session.IDEQ(sessionID)).
		WithProject().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if sessionEnt.Edges.Project != nil {
		return &sessionEnt.Edges.Project.ID, nil
	}

	return nil, nil // No project associated
}

// ClearSessionProject removes the project association from a session
func (r *postgresRepository) ClearSessionProject(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.client.Session.UpdateOneID(sessionID).
		ClearProject().
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("session not found")
		}
		return fmt.Errorf("failed to clear session project: %w", err)
	}

	r.logger.Debug("Project cleared for session",
		zap.String("session_id", sessionID.String()))

	return nil
}

// SetSessionActor sets the actor for a session
func (r *postgresRepository) SetSessionActor(ctx context.Context, sessionID uuid.UUID, actor string) error {
	_, err := r.client.Session.UpdateOneID(sessionID).
		SetActor(actor).
		SetLastActivity(time.Now()).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("session not found")
		}
		return fmt.Errorf("failed to set session actor: %w", err)
	}

	r.logger.Debug("Actor set for session",
		zap.String("session_id", sessionID.String()),
		zap.String("actor", actor))

	return nil
}

// CleanupExpiredSessions removes sessions that have expired
func (r *postgresRepository) CleanupExpiredSessions(ctx context.Context, before time.Time) error {
	_, err := r.client.Session.Delete().
		Where(
			session.And(
				session.ExpiresAtNotNil(),
				session.ExpiresAtLT(before),
			),
		).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	r.logger.Info("Expired sessions cleaned up",
		zap.Time("before", before))

	return nil
}

// GetSessionCount returns the number of active sessions
func (r *postgresRepository) GetSessionCount(ctx context.Context) (int, error) {
	count, err := r.client.Session.Query().
		Where(session.StatusEQ(session.StatusActive)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get session count: %w", err)
	}

	return count, nil
}

// ValidateSession checks if a session exists and is active
func (r *postgresRepository) ValidateSession(ctx context.Context, sessionID uuid.UUID) bool {
	exists, err := r.client.Session.Query().
		Where(
			session.IDEQ(sessionID),
			session.StatusEQ(session.StatusActive),
		).Exist(ctx)
	if err != nil {
		r.logger.Error("Failed to validate session",
			zap.String("session_id", sessionID.String()),
			zap.Error(err))
		return false
	}

	return exists
}

// convertSessionEntToTypes converts an ent Session entity to a types.Session
func (r *postgresRepository) convertSessionEntToTypes(sessionEnt *ent.Session) *types.Session {
	var projectID *uuid.UUID
	if sessionEnt.Edges.Project != nil {
		projectID = &sessionEnt.Edges.Project.ID
	}

	// Handle ExpiresAt conversion
	var expiresAt *time.Time
	if !sessionEnt.ExpiresAt.IsZero() {
		expiresAt = &sessionEnt.ExpiresAt
	}

	return &types.Session{
		ID:           sessionEnt.ID,
		ClientID:     sessionEnt.ClientID,
		CreatedAt:    sessionEnt.CreatedAt,
		LastActivity: sessionEnt.LastActivity,
		ExpiresAt:    expiresAt,
		Metadata:     sessionEnt.Metadata,
		Actor:        sessionEnt.Actor,
		Status:       types.SessionStatus(sessionEnt.Status),
		ProjectID:    projectID,
	}
}

// Health check operations

// HealthCheck performs a comprehensive health check of the PostgreSQL connection
func (r *postgresRepository) HealthCheck(ctx context.Context) (*types.HealthStatus, error) {
	status := &types.HealthStatus{
		LastChecked:  time.Now(),
		// Don't expose full DSN for security - it may contain credentials
		DatabasePath: "postgresql",
	}

	// Check db reference
	if r.db == nil {
		status.ErrorMessage = "database not initialized"
		return status, nil
	}

	// Test basic connectivity with ping
	start := time.Now()
	if err := r.db.PingContext(ctx); err != nil {
		status.ErrorMessage = fmt.Sprintf("ping failed: %v", err)
		return status, nil
	}
	status.PingLatency = time.Since(start)
	status.ConnectionActive = true

	// Get connection pool statistics
	stats := r.db.Stats()
	status.OpenConnections = stats.OpenConnections
	status.IdleConnections = stats.Idle
	status.InUseConnections = stats.InUse

	// Test basic query execution
	if err := r.testBasicQuery(ctx); err != nil {
		status.ErrorMessage = fmt.Sprintf("basic query test failed: %v", err)
		return status, nil
	}

	status.Healthy = true
	return status, nil
}

// Ping performs a simple connectivity test
func (r *postgresRepository) Ping(ctx context.Context) error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return r.db.PingContext(ctx)
}

// ValidateConnection performs a comprehensive connection validation
func (r *postgresRepository) ValidateConnection(ctx context.Context) error {
	health, err := r.HealthCheck(ctx)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if !health.Healthy {
		return fmt.Errorf("database connection unhealthy: %s", health.ErrorMessage)
	}

	// Additional validations
	if health.PingLatency > time.Second {
		r.logger.Warn("High database latency detected",
			zap.Duration("latency", health.PingLatency))
	}

	if health.OpenConnections == 0 {
		return fmt.Errorf("no active database connections")
	}

	r.logger.Info("Database connection validation successful",
		zap.Duration("ping_latency", health.PingLatency),
		zap.Int("open_connections", health.OpenConnections))

	return nil
}

// testBasicQuery tests basic database functionality
func (r *postgresRepository) testBasicQuery(ctx context.Context) error {
	if r.db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Test a simple SELECT query
	var result int
	err := r.db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("basic query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected query result: got %d, expected 1", result)
	}

	return nil
}