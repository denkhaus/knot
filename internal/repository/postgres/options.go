package postgres

import (
	"time"

	"go.uber.org/zap"
)

// Option is a function that configures a PostgreSQL repository
type Option func(*postgresRepository)

// WithLogger sets a logger for the PostgreSQL repository
func WithLogger(logger *zap.Logger) Option {
	return func(r *postgresRepository) {
		r.logger = logger
	}
}

// WithAutoMigrate enables or disables auto-migration for PostgreSQL
func WithAutoMigrate(enable bool) Option {
	return func(r *postgresRepository) {
		r.autoMigrate = enable
	}
}

// WithConnectionPool configures the connection pool for PostgreSQL
func WithConnectionPool(maxOpen, maxIdle int) Option {
	return func(r *postgresRepository) {
		r.maxOpenConns = maxOpen
		r.maxIdleConns = maxIdle
	}
}

// WithConnectionLifetime configures connection lifetimes for PostgreSQL
func WithConnectionLifetime(maxLifetime, maxIdleTime time.Duration) Option {
	return func(r *postgresRepository) {
		r.connMaxLifetime = maxLifetime
		r.connMaxIdleTime = maxIdleTime
	}
}

// WithMigrationTimeout sets the migration timeout for PostgreSQL
func WithMigrationTimeout(timeout time.Duration) Option {
	return func(r *postgresRepository) {
		r.migrationTimeout = timeout
	}
}