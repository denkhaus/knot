package hints

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHintGenerator(t *testing.T) {
	t.Run("enabled with max hints", func(t *testing.T) {
		generator := NewHintGenerator(true, 10, []string{"general", "actions"})

		require.NotNil(t, generator)
		// Test public interface
		hints := generator.GenerateGeneralHints()
		assert.NotEmpty(t, hints)
	})

	t.Run("disabled", func(t *testing.T) {
		generator := NewHintGenerator(false, 5, []string{})

		require.NotNil(t, generator)
		// Test public interface - should return nil when disabled
		hints := generator.GenerateGeneralHints()
		assert.Nil(t, hints)
	})
}

func TestHintGenerator_GenerateHints(t *testing.T) {
	t.Run("disabled returns nil", func(t *testing.T) {
		generator := NewHintGenerator(false, 10, []string{})

		hints := generator.GenerateHints(context.Background(), "project_create", nil, nil)

		assert.Nil(t, hints)
	})

	t.Run("enabled generates hints", func(t *testing.T) {
		generator := NewHintGenerator(true, 10, []string{})

		hints := generator.GenerateHints(context.Background(), "project_create", nil, nil)

		assert.NotNil(t, hints)
	})
}

func TestHintGenerator_GenerateNoProjectSelectedHints(t *testing.T) {
	t.Run("disabled returns nil", func(t *testing.T) {
		generator := NewHintGenerator(false, 10, []string{})

		hints := generator.GenerateNoProjectSelectedHints()

		assert.Nil(t, hints)
	})

	t.Run("enabled returns hints", func(t *testing.T) {
		generator := NewHintGenerator(true, 10, []string{})

		hints := generator.GenerateNoProjectSelectedHints()

		require.NotNil(t, hints)
		assert.Len(t, hints, 2)
		assert.Equal(t, string(HintTypeNextAction), hints[0].Type)
		assert.Equal(t, "No Project Selected", hints[0].Title)
	})
}

func TestHintGenerator_GenerateGeneralHints(t *testing.T) {
	t.Run("disabled returns nil", func(t *testing.T) {
		generator := NewHintGenerator(false, 10, []string{})

		hints := generator.GenerateGeneralHints()

		assert.Nil(t, hints)
	})

	t.Run("enabled returns hints", func(t *testing.T) {
		generator := NewHintGenerator(true, 10, []string{})

		hints := generator.GenerateGeneralHints()

		require.NotNil(t, hints)
		assert.Len(t, hints, 2)
		assert.Equal(t, string(HintTypeBestPractice), hints[0].Type)
		assert.Equal(t, "Project Organization", hints[0].Title)
	})
}

func TestHintGenerator_FilterHints(t *testing.T) {
	t.Run("empty categories returns all", func(t *testing.T) {
		generator := NewHintGenerator(true, 10, []string{})

		hints := []Hint{
			{Type: "next_action", Title: "Action"},
			{Type: "suggestion", Title: "Suggestion"},
		}

		filtered := generator.FilterHints(hints)

		assert.Len(t, filtered, 2)
	})

	t.Run("filters by category", func(t *testing.T) {
		generator := NewHintGenerator(true, 10, []string{"actions"})

		hints := []Hint{
			{Type: "next_action", Title: "Action"},
			{Type: "suggestion", Title: "Suggestion"},
		}

		filtered := generator.FilterHints(hints)

		// Only next_action should match "actions" category
		assert.Len(t, filtered, 1)
		assert.Equal(t, "next_action", filtered[0].Type)
	})
}

func TestHintGeneratorHintTypes(t *testing.T) {
	t.Run("all hint types defined", func(t *testing.T) {
		assert.Equal(t, HintType("next_action"), HintTypeNextAction)
		assert.Equal(t, HintType("warning"), HintTypeWarning)
		assert.Equal(t, HintType("suggestion"), HintTypeSuggestion)
		assert.Equal(t, HintType("best_practice"), HintTypeBestPractice)
	})
}

func TestHintGenerator_AppendProjectCreatedHints(t *testing.T) {
	generator := NewHintGenerator(true, 10, []string{})

	t.Run("adds project creation hints", func(t *testing.T) {
		var hints []Hint
		result := map[string]string{"project_id": "test-id"}

		hints = generator.appendProjectCreatedHints(hints, result, nil)

		require.Len(t, hints, 2)
		assert.Equal(t, string(HintTypeNextAction), hints[0].Type)
		assert.Contains(t, hints[0].Title, "Project Created Successfully")
	})
}

func TestHintGenerator_AppendProjectSelectedHints(t *testing.T) {
	generator := NewHintGenerator(true, 10, []string{})

	t.Run("empty project", func(t *testing.T) {
		var hints []Hint
		session := &SessionState{TaskCount: 0}

		hints = generator.appendProjectSelectedHints(hints, nil, session)

		require.Len(t, hints, 1)
		assert.Contains(t, hints[0].Title, "Project Selected - Ready for Tasks")
	})
}

func TestHintGenerator_AppendProjectListHints(t *testing.T) {
	generator := NewHintGenerator(true, 10, []string{})

	t.Run("no projects", func(t *testing.T) {
		var hints []Hint
		result := []interface{}{}

		hints = generator.appendProjectListHints(hints, result, nil)

		require.Len(t, hints, 1)
		assert.Contains(t, hints[0].Title, "No Projects Found")
	})

	t.Run("single project", func(t *testing.T) {
		var hints []Hint
		result := []interface{}{"project1"}

		hints = generator.appendProjectListHints(hints, result, nil)

		require.Len(t, hints, 1)
		assert.Contains(t, hints[0].Title, "Single Project Available")
	})
}

func TestSessionState(t *testing.T) {
	t.Run("create session state", func(t *testing.T) {
		projectID := "test-project"
		state := &SessionState{
			ProjectID:   &projectID,
			Actor:       "test-user",
			TaskCount:   5,
			RecentTools: []string{"tool1", "tool2"},
		}

		assert.Equal(t, "test-project", *state.ProjectID)
		assert.Equal(t, "test-user", state.Actor)
		assert.Equal(t, 5, state.TaskCount)
		assert.Len(t, state.RecentTools, 2)
	})
}
