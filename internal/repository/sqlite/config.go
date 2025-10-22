package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	ProjectDirName    = ".knot"
	DatabaseName      = "knot.db"
	LegacyDatabaseName = "projects.db"
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
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create project directory: %w", err)
	}
	
	return projectDir, nil
}

// GetDatabasePath returns the full path to the SQLite database file
// Automatically migrates legacy projects.db to knot.db if found
func GetDatabasePath() (string, error) {
	projectDir, err := EnsureProjectDir()
	if err != nil {
		return "", err
	}
	
	dbPath := filepath.Join(projectDir, DatabaseName)
	legacyDbPath := filepath.Join(projectDir, LegacyDatabaseName)
	
	// Check if legacy database exists and current doesn't
	if _, err := os.Stat(legacyDbPath); err == nil {
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			// Migrate legacy database to new name
			if err := os.Rename(legacyDbPath, dbPath); err != nil {
				return "", fmt.Errorf("failed to migrate legacy database from %s to %s: %w", 
					LegacyDatabaseName, DatabaseName, err)
			}
		}
	}
	
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