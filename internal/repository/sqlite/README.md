# PostgreSQL Repository for Project Management Tool

This package provides a PostgreSQL-based repository implementation for the project management tool, replacing the in-memory repository with persistent storage.

## Features

- **Full PostgreSQL Support**: Uses PostgreSQL as the primary database with proper indexing and constraints
- **Automatic Schema Migration**: Automatically creates and maintains database schema on startup
- **Connection Pool Management**: Configurable connection pooling for optimal performance
- **Transaction Support**: Complex operations use transactions to ensure data consistency
- **Error Handling**: Comprehensive error handling with custom error types
- **Circular Dependency Prevention**: Automatic detection and prevention of circular task dependencies
- **Hierarchical Task Support**: Full support for parent-child task relationships with depth tracking
- **Metrics and Analytics**: Built-in project progress calculation and task depth analysis

## Architecture

The repository implementation follows the existing `Repository` interface and provides these key components:

### Database Schema

- **projects**: Project metadata with progress tracking
- **tasks**: Hierarchical task structure with state management
- **task_dependencies**: Many-to-many task dependency relationships

### Key Files

- `repository.go` - Main repository implementation and database setup
- `options.go` - Configuration options and builder patterns
- `errors.go` - Custom error types and error handling utilities
- `project_operations.go` - Project CRUD operations
- `task_operations.go` - Task CRUD operations with hierarchy support
- `task_queries.go` - Advanced task query operations
- `dependency_operations.go` - Task dependency management
- `hierarchy_operations.go` - Hierarchical operations (subtree deletion, metrics)

## Usage

### Basic Setup

```go
package main

import (
    "context"
    "log"
    
    "github.com/denkhaus/agents/pkg/tools/project/repository"
)

func main() {
    // Create repository with default configuration
    repo, err := repository.NewPostgresRepository(
        "postgres://user:password@localhost/project_db?sslmode=disable",
    )
    if err != nil {
        log.Fatal(err)
    }
    defer repo.Close()

    // Use the repository
    ctx := context.Background()
    project := &project.Project{
        ID:          uuid.New(),
        Title:       "My Project",
        Description: "A sample project",
    }
    
    err = repo.CreateProject(ctx, project)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Advanced Configuration

```go
repo, err := repository.NewPostgresRepository(
    "postgres://user:password@localhost/project_db?sslmode=disable",
    repository.WithAutoMigrate(true),
    repository.WithConnectionPool(25, 5),
    repository.WithConnectionLifetime(time.Hour, time.Minute*15),
    repository.WithTaskLimits(20, 5), // max 20 tasks per depth, max depth 5
    repository.WithComplexityThreshold(8),
)
```

### Configuration Options

- `WithAutoMigrate(bool)` - Enable/disable automatic schema migration
- `WithConnectionPool(maxOpen, maxIdle int)` - Configure connection pool size
- `WithConnectionLifetime(maxLifetime, maxIdleTime time.Duration)` - Configure connection lifetimes
- `WithTaskLimits(maxTasksPerDepth, maxDepth int)` - Set task hierarchy limits
- `WithComplexityThreshold(threshold int)` - Set complexity threshold for task breakdown

## Database Requirements

### PostgreSQL Version
- PostgreSQL 13+ recommended
- Requires support for recursive CTEs (for hierarchy operations)
- UUID extension recommended

### Database Setup

```sql
-- Create database
CREATE DATABASE project_management;

-- Create user (optional)
CREATE USER project_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE project_management TO project_user;

-- Enable UUID extension (optional, for better UUID performance)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

### Connection String Format

```
postgres://username:password@host:port/database?options
```

Examples:
- `postgres://localhost/project_db?sslmode=disable` (local, no auth)
- `postgres://user:pass@localhost:5432/project_db?sslmode=require` (with auth and SSL)
- `postgres://user:pass@prod-db.example.com/project_db?sslmode=require&application_name=project-tool`

## Performance Considerations

### Indexing Strategy
The repository automatically creates optimized indexes for:
- Primary keys and foreign keys
- Common query patterns (project_id + state, project_id + agent, etc.)
- Dependency lookups
- Hierarchical queries

### Connection Pooling
Default configuration:
- MaxOpenConns: 25
- MaxIdleConns: 5
- ConnMaxLifetime: 1 hour
- ConnMaxIdleTime: 15 minutes

Adjust based on your application load and database capacity.

### Query Optimization
- Uses prepared statements for all queries
- Efficient recursive CTEs for hierarchical operations
- Batch operations where possible
- Proper transaction boundaries

## Error Handling

The repository provides custom error types for different scenarios:

```go
import "github.com/denkhaus/agents/pkg/tools/project/repository"

// Check specific error types
err := repo.GetProject(ctx, nonExistentID)
if repository.IsNotFoundError(err) {
    // Handle not found
}

// Other error types
repository.IsConstraintViolationError(err)
repository.IsCircularDependencyError(err)
repository.IsMaxDepthExceededError(err)
repository.IsConnectionError(err)
repository.IsTransactionError(err)
```

## Transaction Behavior

Complex operations automatically use transactions:
- Project deletion (cascades to tasks and dependencies)
- Task subtree deletion
- Task dependency modifications
- Bulk operations

## Limitations and Considerations

1. **Migration Strategy**: Currently uses auto-migration for development. Production deployments should consider controlled migration strategies.

2. **Ent Framework**: The initial design included Facebook's ent ORM, but the current implementation uses standard database/sql for simplicity and reliability.

3. **Concurrent Access**: The repository is thread-safe, but complex business logic operations should handle concurrent modifications appropriately.

4. **Resource Cleanup**: Always call `Close()` on the repository to properly close database connections.

## Testing

Run the integration tests with a real PostgreSQL database:

```bash
# Set up test database
createdb project_test

# Run tests
go test -v ./...

# Run with custom database URL
POSTGRES_TEST_URL=postgres://localhost/my_test_db?sslmode=disable go test -v ./...
```

## Migration from In-Memory Repository

The PostgreSQL repository is a drop-in replacement for the in-memory repository. Simply replace the repository initialization:

```go
// Old: in-memory repository
repo := project.NewMemoryRepository()

// New: PostgreSQL repository
repo, err := repository.NewPostgresRepository(databaseURL)
if err != nil {
    log.Fatal(err)
}
defer repo.Close()
```

All existing code using the `Repository` interface will continue to work without changes.