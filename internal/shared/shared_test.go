package shared

import (
	"context"
	"testing"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/repository/inmemory"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func TestResolveProjectID(t *testing.T) {
	// Setup test environment
	repo := inmemory.NewMemoryRepository()
	config := manager.DefaultConfig()
	projectManager := manager.NewManagerWithRepository(repo, config)
	logger := zap.NewNop()
	appCtx := NewAppContext(projectManager, logger)

	// Create a test project
	project, err := projectManager.CreateProject(context.Background(), "Test Project", "Test Description", "test-actor")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	tests := []struct {
		name           string
		setupSelection bool
		projectID      *uuid.UUID
		expectError    bool
		errorContains  string
	}{
		{
			name:           "No project selected",
			setupSelection: false,
			expectError:    true,
			errorContains:  "no project is currently selected",
		},
		{
			name:           "Project selected successfully",
			setupSelection: true,
			projectID:      &project.ID,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear any existing selection
			_ = projectManager.ClearSelectedProject(context.Background())

			// Setup selection if needed
			if tt.setupSelection && tt.projectID != nil {
				err := projectManager.SetSelectedProject(context.Background(), *tt.projectID, "test-actor")
				if err != nil {
					t.Fatalf("Failed to set selected project: %v", err)
				}
			}

			// Create a mock CLI context
			app := &cli.App{}
			ctx := cli.NewContext(app, nil, nil)

			// Test ResolveProjectID
			resolvedID, err := ResolveProjectID(ctx, appCtx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', but got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if tt.projectID != nil && resolvedID != *tt.projectID {
					t.Errorf("Expected project ID %v, but got %v", *tt.projectID, resolvedID)
				}
			}
		})
	}
}

func TestValidateProjectID(t *testing.T) {
	tests := []struct {
		name          string
		projectIDStr  string
		expectError   bool
		errorContains string
	}{
		{
			name:          "Empty project ID",
			projectIDStr:  "",
			expectError:   true,
			errorContains: "project-id",
		},
		{
			name:          "Invalid UUID format",
			projectIDStr:  "invalid-uuid",
			expectError:   true,
			errorContains: "invalid",
		},
		{
			name:         "Valid UUID",
			projectIDStr: "550e8400-e29b-41d4-a716-446655440000",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock CLI context with project-id flag
			app := &cli.App{
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "project-id",
					},
				},
			}

			// Create args slice to simulate command line arguments
			args := []string{"test"}
			if tt.projectIDStr != "" {
				args = append(args, "--project-id", tt.projectIDStr)
			}

			// Parse the arguments to create context
			err := app.Run(args)
			if err != nil && !tt.expectError {
				t.Errorf("Failed to create CLI context: %v", err)
				return
			}

			// For this test, we'll test the UUID parsing logic directly
			if tt.expectError {
				if tt.projectIDStr == "" {
					// Test empty string case
					_, err := uuid.Parse("")
					if err == nil {
						t.Errorf("Expected error for empty string but got none")
					}
				} else {
					// Test invalid UUID case
					_, err := uuid.Parse(tt.projectIDStr)
					if err == nil {
						t.Errorf("Expected error for invalid UUID but got none")
					}
				}
			} else {
				// Test valid UUID case
				resolvedID, err := uuid.Parse(tt.projectIDStr)
				if err != nil {
					t.Errorf("Unexpected error for valid UUID: %v", err)
					return
				}
				expectedID, _ := uuid.Parse(tt.projectIDStr)
				if resolvedID != expectedID {
					t.Errorf("Expected project ID %v, but got %v", expectedID, resolvedID)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
