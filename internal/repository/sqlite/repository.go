package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/denkhaus/knot/internal/repository/sqlite/ent"
	"github.com/denkhaus/knot/internal/types"
	_ "modernc.org/sqlite"
	"go.uber.org/zap"
)

// sqliteRepository implements the Repository interface using ent ORM
type sqliteRepository struct {
	client *ent.Client
	config *Config
	logger *zap.Logger
}

// NewRepository creates a new SQLite repository using ent ORM
func NewRepository(opts ...Option) (types.Repository, error) {
	config := OptimizedConfig() // Use optimized config by default

	repo := &sqliteRepository{
		config: config,
		logger: config.Logger,
	}

	// Apply options
	for _, opt := range opts {
		opt(repo)
	}

	if err := repo.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize repository: %w", err)
	}

	return repo, nil
}

// initialize sets up the ent client and performs migrations
func (r *sqliteRepository) initialize() error {
	// Get SQLite connection string
	connStr, err := GetSQLiteConnectionString()
	if err != nil {
		return fmt.Errorf("failed to get SQLite connection string: %w", err)
	}

	r.config.Logger.Info("initialize database", zap.String("database_path", connStr))

	db, err := sql.Open("sqlite", connStr)
	if err != nil {
		return NewConnectionError("failed to open database connection", err)
	}
	
	// Configure SQLite for optimal performance
	if err := r.configureSQLiteOptimizations(db); err != nil {
		return NewConnectionError("failed to configure SQLite optimizations", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(r.config.MaxOpenConns)
	db.SetMaxIdleConns(r.config.MaxIdleConns)
	db.SetConnMaxLifetime(r.config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(r.config.ConnMaxIdleTime)

	// Test connection with comprehensive validation
	if err := r.validateInitialConnection(db); err != nil {
		return NewConnectionError("database connection validation failed", err)
	}

	// Create ent client with SQLite driver
	drv := entsql.OpenDB(dialect.SQLite, db)
	r.client = ent.NewClient(ent.Driver(drv))

	// Run auto-migration if enabled
	if r.config.AutoMigrate {
		ctx, cancel := context.WithTimeout(context.Background(), r.config.MigrationTimeout)
		defer cancel()

		// Use safe migration options to add new tables without affecting existing data
		if err := r.client.Schema.Create(ctx, 
			schema.WithDropIndex(false),
			schema.WithDropColumn(false),
		); err != nil {
			return NewMigrationError("auto-migration failed", err)
		}
		
		r.logger.Info("Database schema migration completed successfully")
	}

	return nil
}

// Close closes the ent client and database connection
func (r *sqliteRepository) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// mapError converts ent/database errors to repository errors
func (r *sqliteRepository) mapError(operation string, err error) error {
	if err == nil {
		return nil
	}

	if ent.IsNotFound(err) {
		return NewNotFoundError("resource", "unknown")
	}

	if ent.IsConstraintError(err) {
		return NewConstraintViolationError("constraint violation", err)
	}

	// TODO: Add more specific error mapping for different ent error types
	return NewConnectionError(fmt.Sprintf("database operation failed: %s", operation), err)
}

// configureSQLiteOptimizations applies SQLite-specific performance optimizations
func (r *sqliteRepository) configureSQLiteOptimizations(db *sql.DB) error {
	optimizations := []struct {
		name   string
		pragma string
	}{
		{"foreign_keys", "PRAGMA foreign_keys = ON"},
		{"wal_mode", "PRAGMA journal_mode = WAL"},           // Enable WAL mode for better concurrency
		{"synchronous", "PRAGMA synchronous = NORMAL"},      // Balance between safety and performance
		{"cache_size", "PRAGMA cache_size = -64000"},        // 64MB cache (negative = KB)
		{"temp_store", "PRAGMA temp_store = MEMORY"},        // Store temp tables in memory
		{"mmap_size", "PRAGMA mmap_size = 268435456"},       // 256MB memory-mapped I/O
		{"optimize", "PRAGMA optimize"},                     // Analyze and optimize query planner
	}

	for _, opt := range optimizations {
		if _, err := db.Exec(opt.pragma); err != nil {
			r.config.Logger.Warn("Failed to apply SQLite optimization", 
				zap.String("optimization", opt.name),
				zap.String("pragma", opt.pragma),
				zap.Error(err))
			// Continue with other optimizations even if one fails
		} else {
			r.config.Logger.Debug("Applied SQLite optimization", 
				zap.String("optimization", opt.name))
		}
	}

	return nil
}

// validateInitialConnection performs initial connection validation during setup
func (r *sqliteRepository) validateInitialConnection(db *sql.DB) error {
	// Basic ping test
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Test basic query execution
	var result int
	if err := db.QueryRow("SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("basic query test failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected query result: got %d, expected 1", result)
	}

	// Verify database file is writable by creating a test table
	testTableSQL := `
		CREATE TABLE IF NOT EXISTS _knot_connection_test (
			id INTEGER PRIMARY KEY,
			test_value TEXT
		)
	`
	if _, err := db.Exec(testTableSQL); err != nil {
		return fmt.Errorf("write test failed: %w", err)
	}

	// Clean up test table
	if _, err := db.Exec("DROP TABLE IF EXISTS _knot_connection_test"); err != nil {
		r.config.Logger.Warn("Failed to clean up test table", zap.Error(err))
		// Don not fail initialization for cleanup failure
	}

	r.config.Logger.Info("Database connection validation successful")
	return nil
}
