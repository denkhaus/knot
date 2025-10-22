package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// HealthStatus represents the health status of the database connection
type HealthStatus struct {
	Healthy          bool          `json:"healthy"`
	ConnectionActive bool          `json:"connection_active"`
	PingLatency      time.Duration `json:"ping_latency"`
	OpenConnections  int           `json:"open_connections"`
	IdleConnections  int           `json:"idle_connections"`
	InUseConnections int           `json:"in_use_connections"`
	ErrorMessage     string        `json:"error_message,omitempty"`
	LastChecked      time.Time     `json:"last_checked"`
	DatabasePath     string        `json:"database_path"`
	WALModeEnabled   bool          `json:"wal_mode_enabled"`
	ForeignKeys      bool          `json:"foreign_keys_enabled"`
}

// HealthCheck performs a comprehensive health check of the database connection
func (r *sqliteRepository) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		LastChecked:  time.Now(),
		DatabasePath: r.config.DatabasePath,
	}

	// Get underlying database connection
	db, err := r.getUnderlyingDB()
	if err != nil {
		status.ErrorMessage = fmt.Sprintf("failed to get database connection: %v", err)
		return status, nil
	}

	// Test basic connectivity with ping
	start := time.Now()
	if err := db.PingContext(ctx); err != nil {
		status.ErrorMessage = fmt.Sprintf("ping failed: %v", err)
		return status, nil
	}
	status.PingLatency = time.Since(start)
	status.ConnectionActive = true

	// Get connection pool statistics
	stats := db.Stats()
	status.OpenConnections = stats.OpenConnections
	status.IdleConnections = stats.Idle
	status.InUseConnections = stats.InUse

	// Test basic query execution
	if err := r.testBasicQuery(ctx, db); err != nil {
		status.ErrorMessage = fmt.Sprintf("basic query test failed: %v", err)
		return status, nil
	}

	// Check SQLite-specific settings
	if err := r.checkSQLiteSettings(ctx, db, status); err != nil {
		r.config.Logger.Warn("Failed to check SQLite settings", zap.Error(err))
		// Don't fail health check for settings check failure
	}

	status.Healthy = true
	return status, nil
}

// Ping performs a simple connectivity test
func (r *sqliteRepository) Ping(ctx context.Context) error {
	db, err := r.getUnderlyingDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	return db.PingContext(ctx)
}

// ValidateConnection performs a comprehensive connection validation
func (r *sqliteRepository) ValidateConnection(ctx context.Context) error {
	health, err := r.HealthCheck(ctx)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if !health.Healthy {
		return fmt.Errorf("database connection unhealthy: %s", health.ErrorMessage)
	}

	// Additional validations
	if health.PingLatency > time.Second {
		r.config.Logger.Warn("High database latency detected", 
			zap.Duration("latency", health.PingLatency))
	}

	if health.OpenConnections == 0 {
		return fmt.Errorf("no active database connections")
	}

	r.config.Logger.Info("Database connection validation successful",
		zap.Duration("ping_latency", health.PingLatency),
		zap.Int("open_connections", health.OpenConnections),
		zap.Bool("wal_mode", health.WALModeEnabled))

	return nil
}

// getUnderlyingDB extracts the underlying sql.DB from ent client
func (r *sqliteRepository) getUnderlyingDB() (*sql.DB, error) {
	if r.client == nil {
		return nil, fmt.Errorf("ent client not initialized")
	}

	// For now, we'll use a simpler approach - test through ent client
	// TODO: Find a way to access underlying sql.DB if needed for advanced health checks
	return nil, fmt.Errorf("direct database access not available through ent client")
}

// testBasicQuery tests basic database functionality
func (r *sqliteRepository) testBasicQuery(ctx context.Context, db *sql.DB) error {
	// Test a simple SELECT query
	var result int
	err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("basic query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected query result: got %d, expected 1", result)
	}

	return nil
}

// checkSQLiteSettings verifies SQLite-specific configuration
func (r *sqliteRepository) checkSQLiteSettings(ctx context.Context, db *sql.DB, status *HealthStatus) error {
	settings := []struct {
		name     string
		query    string
		expected string
		field    *bool
	}{
		{"journal_mode", "PRAGMA journal_mode", "wal", &status.WALModeEnabled},
		{"foreign_keys", "PRAGMA foreign_keys", "1", &status.ForeignKeys},
	}

	for _, setting := range settings {
		var value string
		err := db.QueryRowContext(ctx, setting.query).Scan(&value)
		if err != nil {
			r.config.Logger.Warn("Failed to check SQLite setting",
				zap.String("setting", setting.name),
				zap.Error(err))
			continue
		}

		if setting.field != nil {
			*setting.field = (value == setting.expected)
		}

		r.config.Logger.Debug("SQLite setting checked",
			zap.String("setting", setting.name),
			zap.String("value", value),
			zap.String("expected", setting.expected))
	}

	return nil
}

// MonitorConnection starts a background health monitoring routine
func (r *sqliteRepository) MonitorConnection(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.config.Logger.Info("Database connection monitoring stopped")
			return
		case <-ticker.C:
			health, err := r.HealthCheck(ctx)
			if err != nil {
				r.config.Logger.Error("Health check failed", zap.Error(err))
				continue
			}

			if !health.Healthy {
				r.config.Logger.Error("Database connection unhealthy",
					zap.String("error", health.ErrorMessage),
					zap.Duration("ping_latency", health.PingLatency))
			} else {
				r.config.Logger.Debug("Database connection healthy",
					zap.Duration("ping_latency", health.PingLatency),
					zap.Int("open_connections", health.OpenConnections))
			}
		}
	}
}