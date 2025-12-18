package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/denkhaus/knot/v2/internal/repository/ent"
	"github.com/denkhaus/knot/v2/internal/types"
	"go.uber.org/zap"
	_ "github.com/lib/pq"
)

// postgresRepository implements the Repository interface using ent ORM with PostgreSQL
type postgresRepository struct {
	client    *ent.Client
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