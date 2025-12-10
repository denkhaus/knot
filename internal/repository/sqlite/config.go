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

// FindWorkspaceRoot searches for the workspace root by traversing up the directory tree
// It looks for .knot directory first, then .git directory as a fallback
// Returns the workspace root path or error if not found
func FindWorkspaceRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	dir := cwd

	for {
		// Check if .knot exists in current directory
		knotDir := filepath.Join(dir, ProjectDirName)
		if stat, err := os.Stat(knotDir); err == nil && stat.IsDir() {
			return dir, nil
		}

		// Check if .git exists (likely workspace root)
		gitDir := filepath.Join(dir, ".git")
		if stat, err := os.Stat(gitDir); err == nil && stat.IsDir() {
			return dir, nil
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("workspace root not found (no .knot or .git directory found in path hierarchy)")
}

// GetWorkspaceProjectDir returns the .knot directory path in the workspace root
// This is the preferred method as it ensures only one .knot per workspace
func GetWorkspaceProjectDir() (string, error) {
	workspaceRoot, err := FindWorkspaceRoot()
	if err != nil {
		return "", err
	}

	projectDir := filepath.Join(workspaceRoot, ProjectDirName)
	return projectDir, nil
}

// EnsureWorkspaceProjectDir creates the .knot directory in the workspace root if it doesn't exist
// This is the preferred method as it ensures only one .knot per workspace
func EnsureWorkspaceProjectDir() (string, error) {
	projectDir, err := GetWorkspaceProjectDir()
	if err != nil {
		return "", err
	}

	// Create directory if it doesn't exist with secure permissions (owner only)
	if err := os.MkdirAll(projectDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create workspace project directory: %w", err)
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

// GetWorkspaceDatabasePath returns the full path to the SQLite database file in workspace root
// This is the preferred method as it ensures only one database per workspace
func GetWorkspaceDatabasePath() (string, error) {
	projectDir, err := EnsureWorkspaceProjectDir()
	if err != nil {
		return "", err
	}

	dbPath := filepath.Join(projectDir, DatabaseName)
	return dbPath, nil
}

// GetWorkspaceSQLiteConnectionString returns the SQLite connection string for workspace
// This is the preferred method as it ensures only one database per workspace
func GetWorkspaceSQLiteConnectionString() (string, error) {
	dbPath, err := GetWorkspaceDatabasePath()
	if err != nil {
		return "", err
	}

	// SQLite connection string - simple path format
	return dbPath, nil
}
