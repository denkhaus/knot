// Package sqlite provides SQLite database repository implementation for KNOT.
//
// This package implements the repository interface using SQLite as the persistent
// storage backend. It offers data persistence, ACID compliance, and
// production-ready storage for the KNOT project management system.
//
// Key Components:
//   - SQLite Repository: Main repository implementation
//   - Database Configuration: Connection and optimization settings
//   - Migration Support: Database schema management
//   - Transaction Handling: ACID-compliant operations
//
// Cross-reference: Knot Task 86f3ba2d-3a87-493b-b8fc-96d19f344e89
package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	ProjectDirName = ".knot"
	DatabaseName   = "knot.db"
)

// GetProjectDir returns the .knot directory path in the current working directory
func GetProjectDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	projectDir := filepath.Join(cwd, ProjectDirName)
	return projectDir, nil
}

// EnsureProjectDir creates the .knot directory if it doesn't exist
func EnsureProjectDir() (string, error) {
	projectDir, err := GetProjectDir()
	if err != nil {
		return "", err
	}

	// Create directory if it doesn't exist with secure permissions (owner only)
	if err := os.MkdirAll(projectDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create project directory: %w", err)
	}

	// Verify and fix directory permissions are secure, even for existing directories
	return ensureSecureDirectory(projectDir)
}

// ensureSecureDirectory verifies and fixes directory permissions for security
func ensureSecureDirectory(dirPath string) (string, error) {
	// Check current permissions
	info, err := os.Stat(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat directory: %w", err)
	}

	currentPerms := info.Mode().Perm()
	securePerms := os.FileMode(0o700)

	// If permissions are not secure, fix them
	if currentPerms != securePerms {
		if err := os.Chmod(dirPath, securePerms); err != nil {
			return "", fmt.Errorf("failed to set secure permissions on directory: %w", err)
		}
	}

	return dirPath, nil
}

// GetDatabasePath returns the full path to the SQLite database file
// Automatically migrates legacy projects.db to knot.db if found
func GetDatabasePath() (string, error) {
	projectDir, err := EnsureProjectDir()
	if err != nil {
		return "", err
	}

	dbPath := filepath.Join(projectDir, DatabaseName)
	return dbPath, nil
}

// GetSQLiteConnectionString returns the SQLite connection string
func GetSQLiteConnectionString() (string, error) {
	dbPath, err := GetDatabasePath()
	if err != nil {
		return "", err
	}

	// SQLite connection string - simple path format
	return dbPath, nil
}
