package manager

import (
	"context"
	"testing"

	"github.com/denkhaus/knot/internal/repository/inmemory"
	"github.com/google/uuid"
)

func TestProjectSelection(t *testing.T) {
	// Setup test environment
	repo := inmemory.NewMemoryRepository()
	config := DefaultConfig()
	manager := NewManagerWithRepository(repo, config)

	// Create test projects
	project1, err := manager.CreateProject(context.Background(), "Project 1", "First test project", "test-actor")
	if err != nil {
		t.Fatalf("Failed to create project 1: %v", err)
	}

	project2, err := manager.CreateProject(context.Background(), "Project 2", "Second test project", "test-actor")
	if err != nil {
		t.Fatalf("Failed to create project 2: %v", err)
	}

	t.Run("No project selected initially", func(t *testing.T) {
		hasSelected, err := manager.HasSelectedProject(context.Background())
		if err != nil {
			t.Errorf("Unexpected error checking selection: %v", err)
		}
		if hasSelected {
			t.Errorf("Expected no project selected initially")
		}

		selectedID, err := manager.GetSelectedProject(context.Background())
		if err != nil {
			t.Errorf("Unexpected error getting selection: %v", err)
		}
		if selectedID != nil {
			t.Errorf("Expected nil selected project, got %v", selectedID)
		}
	})

	t.Run("Select project successfully", func(t *testing.T) {
		err := manager.SetSelectedProject(context.Background(), project1.ID, "test-actor")
		if err != nil {
			t.Errorf("Failed to select project: %v", err)
		}

		hasSelected, err := manager.HasSelectedProject(context.Background())
		if err != nil {
			t.Errorf("Unexpected error checking selection: %v", err)
		}
		if !hasSelected {
			t.Errorf("Expected project to be selected")
		}

		selectedID, err := manager.GetSelectedProject(context.Background())
		if err != nil {
			t.Errorf("Unexpected error getting selection: %v", err)
		}
		if selectedID == nil || *selectedID != project1.ID {
			t.Errorf("Expected selected project %v, got %v", project1.ID, selectedID)
		}
	})

	t.Run("Change selected project", func(t *testing.T) {
		err := manager.SetSelectedProject(context.Background(), project2.ID, "test-actor")
		if err != nil {
			t.Errorf("Failed to change selected project: %v", err)
		}

		selectedID, err := manager.GetSelectedProject(context.Background())
		if err != nil {
			t.Errorf("Unexpected error getting selection: %v", err)
		}
		if selectedID == nil || *selectedID != project2.ID {
			t.Errorf("Expected selected project %v, got %v", project2.ID, selectedID)
		}
	})

	t.Run("Clear selected project", func(t *testing.T) {
		err := manager.ClearSelectedProject(context.Background())
		if err != nil {
			t.Errorf("Failed to clear selected project: %v", err)
		}

		hasSelected, err := manager.HasSelectedProject(context.Background())
		if err != nil {
			t.Errorf("Unexpected error checking selection: %v", err)
		}
		if hasSelected {
			t.Errorf("Expected no project selected after clearing")
		}

		selectedID, err := manager.GetSelectedProject(context.Background())
		if err != nil {
			t.Errorf("Unexpected error getting selection: %v", err)
		}
		if selectedID != nil {
			t.Errorf("Expected nil selected project after clearing, got %v", selectedID)
		}
	})

	t.Run("Select non-existent project", func(t *testing.T) {
		nonExistentID := uuid.New()
		err := manager.SetSelectedProject(context.Background(), nonExistentID, "test-actor")
		if err == nil {
			t.Errorf("Expected error when selecting non-existent project")
		}
	})
}

func TestProjectSelectionPersistence(t *testing.T) {
	// Setup test environment
	repo := inmemory.NewMemoryRepository()
	config := DefaultConfig()
	manager := NewManagerWithRepository(repo, config)

	// Create test project
	project, err := manager.CreateProject(context.Background(), "Test Project", "Test Description", "test-actor")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Select project
	err = manager.SetSelectedProject(context.Background(), project.ID, "test-actor")
	if err != nil {
		t.Fatalf("Failed to select project: %v", err)
	}

	// Verify selection persists across multiple calls
	for i := 0; i < 3; i++ {
		selectedID, err := manager.GetSelectedProject(context.Background())
		if err != nil {
			t.Errorf("Iteration %d: Unexpected error getting selection: %v", i, err)
		}
		if selectedID == nil || *selectedID != project.ID {
			t.Errorf("Iteration %d: Expected selected project %v, got %v", i, project.ID, selectedID)
		}
	}
}

func TestProjectSelectionWithDeletedProject(t *testing.T) {
	// Setup test environment
	repo := inmemory.NewMemoryRepository()
	config := DefaultConfig()
	manager := NewManagerWithRepository(repo, config)

	// Create test project
	project, err := manager.CreateProject(context.Background(), "Test Project", "Test Description", "test-actor")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Select project
	err = manager.SetSelectedProject(context.Background(), project.ID, "test-actor")
	if err != nil {
		t.Fatalf("Failed to select project: %v", err)
	}

	// Delete the project
	err = manager.DeleteProject(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	// Check that selection is cleared or returns error gracefully
	selectedID, err := manager.GetSelectedProject(context.Background())
	// Either the selection should be cleared (selectedID == nil) or an error should be returned
	// The exact behavior depends on implementation, but it shouldn't crash
	if err == nil && selectedID != nil {
		// If no error, verify the project actually exists
		_, getErr := manager.GetProject(context.Background(), *selectedID)
		if getErr != nil {
			t.Errorf("Selected project ID points to non-existent project")
		}
	}
}