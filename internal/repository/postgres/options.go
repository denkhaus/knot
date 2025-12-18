package postgres

import (
	"time"

	"github.com/denkhaus/knot/v2/internal/types"
	"go.uber.org/zap"
)

// Option is a function that configures a PostgreSQL repository
type Option func(*Adapter)

// WithConfig sets the entire configuration
func WithConfig(config *Config) Option {
	return func(a *Adapter) {
		a.config = config
	}
}

// WithDSN sets the PostgreSQL connection string
func WithDSN(dsn string) Option {
	return func(a *Adapter) {
		a.dsn = dsn
	}
}

// WithLogger sets a logger for debugging reasons
func WithLogger(logger *zap.Logger) Option {
	return func(a *Adapter) {
		if a.config == nil {
			a.config = DefaultConfig()
		}
		a.config.Logger = logger
	}
}

// WithAutoMigrate enables or disables auto-migration
func WithAutoMigrate(enable bool) Option {
	return func(a *Adapter) {
		if a.config == nil {
			a.config = DefaultConfig()
		}
		a.config.AutoMigrate = enable
	}
}

// WithConnectionPool configures the connection pool for PostgreSQL
func WithConnectionPool(maxOpen, maxIdle int) Option {
	return func(a *Adapter) {
		if a.config == nil {
			a.config = DefaultConfig()
		}
		a.config.MaxOpenConns = maxOpen
		a.config.MaxIdleConns = maxIdle
	}
}

// WithConnectionLifetime configures connection lifetimes
func WithConnectionLifetime(maxLifetime, maxIdleTime time.Duration) Option {
	return func(a *Adapter) {
		if a.config == nil {
			a.config = DefaultConfig()
		}
		a.config.ConnMaxLifetime = maxLifetime
		a.config.ConnMaxIdleTime = maxIdleTime
	}
}

// WithMigrationTimeout sets the migration timeout
func WithMigrationTimeout(timeout time.Duration) Option {
	return func(a *Adapter) {
		if a.config == nil {
			a.config = DefaultConfig()
		}
		a.config.MigrationTimeout = timeout
	}
}