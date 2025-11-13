package sqlite

import (
	"time"

	"go.uber.org/zap"
)

// Config holds configuration for the SQLite repository
type Config struct {
	// Database connection settings
	DatabasePath    string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	// Migration settings
	AutoMigrate      bool
	MigrationTimeout time.Duration
	Logger           *zap.Logger
}

// DefaultConfig returns a default configuration optimized for SQLite
func DefaultConfig() *Config {
	return &Config{
		DatabasePath: "",
		// SQLite optimized connection pool settings:
		// SQLite works best with limited concurrent connections due to file locking
		MaxOpenConns:     1,                // SQLite is single-writer, multiple readers - limit to 1 for writes
		MaxIdleConns:     1,                // Keep 1 idle connection to avoid reconnection overhead
		ConnMaxLifetime:  0,                // No connection lifetime limit for SQLite (file-based)
		ConnMaxIdleTime:  time.Minute * 30, // Longer idle time for file-based DB
		AutoMigrate:      true,
		MigrationTimeout: time.Minute * 5,
		Logger:           zap.NewNop(),
	}
}

// Option is a function that configures a SQLite repository
type Option func(*sqliteRepository)

// WithConfig sets the entire configuration
func WithConfig(config *Config) Option {
	return func(r *sqliteRepository) {
		r.config = config
	}
}

// WithDatabasePath sets the database file path
func WithDatabasePath(path string) Option {
	return func(r *sqliteRepository) {
		r.config.DatabasePath = path
	}
}

// WithLogger sets a logger for debugging reasons
func WithLogger(logger *zap.Logger) Option {
	return func(r *sqliteRepository) {
		r.config.Logger = logger
	}
}

// WithAutoMigrate enables or disables auto-migration
func WithAutoMigrate(enable bool) Option {
	return func(r *sqliteRepository) {
		r.config.AutoMigrate = enable
	}
}

// WithConnectionPool configures the connection pool
func WithConnectionPool(maxOpen, maxIdle int) Option {
	return func(r *sqliteRepository) {
		r.config.MaxOpenConns = maxOpen
		r.config.MaxIdleConns = maxIdle
	}
}

// WithConnectionLifetime configures connection lifetimes
func WithConnectionLifetime(maxLifetime, maxIdleTime time.Duration) Option {
	return func(r *sqliteRepository) {
		r.config.ConnMaxLifetime = maxLifetime
		r.config.ConnMaxIdleTime = maxIdleTime
	}
}

// WithMigrationTimeout sets the migration timeout
func WithMigrationTimeout(timeout time.Duration) Option {
	return func(r *sqliteRepository) {
		r.config.MigrationTimeout = timeout
	}
}
