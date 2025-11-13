package sqlite

import (
	"os"
	"runtime"
	"strconv"
	"time"
)

// OptimizedConfig returns SQLite-optimized configuration based on environment and workload
func OptimizedConfig() *Config {
	config := DefaultConfig()

	// Check for environment variable overrides
	if maxOpenStr := os.Getenv("KNOT_SQLITE_MAX_OPEN_CONNS"); maxOpenStr != "" {
		if maxOpen, err := strconv.Atoi(maxOpenStr); err == nil && maxOpen > 0 {
			config.MaxOpenConns = maxOpen
		}
	}

	if maxIdleStr := os.Getenv("KNOT_SQLITE_MAX_IDLE_CONNS"); maxIdleStr != "" {
		if maxIdle, err := strconv.Atoi(maxIdleStr); err == nil && maxIdle > 0 {
			config.MaxIdleConns = maxIdle
		}
	}

	// Auto-tune based on system resources and workload patterns
	config = autoTuneForSQLite(config)

	return config
}

// autoTuneForSQLite optimizes connection pool settings for SQLite workloads
func autoTuneForSQLite(config *Config) *Config {
	numCPU := runtime.NumCPU()

	// SQLite connection pool optimization rules:
	// 1. SQLite uses file-level locking, so too many connections can cause contention
	// 2. Read operations can be concurrent, but writes are serialized
	// 3. WAL mode allows better concurrency than default journal mode

	// For read-heavy workloads, allow more connections
	// For write-heavy workloads, keep it minimal
	if config.MaxOpenConns == 1 { // Only auto-tune if using default
		// Conservative approach: limit to CPU count but cap at reasonable maximum
		readConnections := min(numCPU, 4)
		config.MaxOpenConns = readConnections
		config.MaxIdleConns = min(readConnections/2, 2)
	}

	// Optimize connection lifetimes for SQLite
	if config.ConnMaxLifetime == 0 {
		// SQLite doesn't need connection rotation like network databases
		config.ConnMaxLifetime = 0 // Unlimited lifetime
	}

	// Longer idle time for file-based databases
	if config.ConnMaxIdleTime == time.Minute*30 {
		config.ConnMaxIdleTime = time.Hour // Keep connections longer
	}

	return config
}

// GetWorkloadOptimizedConfig returns configuration optimized for specific workload patterns
func GetWorkloadOptimizedConfig(workloadType WorkloadType) *Config {
	config := DefaultConfig()

	switch workloadType {
	case WorkloadReadHeavy:
		// Optimize for read-heavy workloads (CLI queries, reports)
		config.MaxOpenConns = min(runtime.NumCPU(), 6)
		config.MaxIdleConns = 3
		config.ConnMaxIdleTime = time.Hour

	case WorkloadWriteHeavy:
		// Optimize for write-heavy workloads (bulk operations, imports)
		config.MaxOpenConns = 1 // SQLite serializes writes anyway
		config.MaxIdleConns = 1
		config.ConnMaxIdleTime = time.Minute * 15

	case WorkloadMixed:
		// Balanced configuration for mixed workloads
		config.MaxOpenConns = min(runtime.NumCPU()/2+1, 3)
		config.MaxIdleConns = 2
		config.ConnMaxIdleTime = time.Minute * 30

	case WorkloadBatch:
		// Optimize for batch processing
		config.MaxOpenConns = 1 // Single connection for consistency
		config.MaxIdleConns = 1
		config.ConnMaxLifetime = 0 // Long-running operations
		config.ConnMaxIdleTime = time.Hour * 2
	}

	return config
}

// WorkloadType defines different database workload patterns
type WorkloadType int

const (
	WorkloadReadHeavy  WorkloadType = iota // Mostly SELECT operations
	WorkloadWriteHeavy                     // Mostly INSERT/UPDATE/DELETE operations
	WorkloadMixed                          // Balanced read/write operations
	WorkloadBatch                          // Long-running batch operations
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
