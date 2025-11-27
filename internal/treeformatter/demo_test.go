package treeformatter

import (
	"fmt"
	"testing"

	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
)

// TestDemoTreeFormat demonstrates the target format as specified in the brain memory
func TestDemoTreeFormat(t *testing.T) {
	formatter := NewDefaultFormatter()
	formatter.config.ShowEmojis = true

	// Create sample tasks matching the brain memory specification
	rootTask := &types.Task{
		ID:    uuid.MustParse("0a4f4544-2802-493b-9986-e0d7b7a41d2e"),
		Title: "Core Architecture Implementation",
		State: types.TaskStateInProgress,
	}

	child1 := &types.Task{
		ID:    uuid.MustParse("3f22bffe-5004-4002-877c-da7be9276688"),
		Title: "ConsolidationScheduler Service",
		State: types.TaskStateCompleted,
	}

	child2 := &types.Task{
		ID:    uuid.MustParse("7415bb39-7832-4dd7-b2a9-4fe56ad30985"),
		Title: "InMemoryState Management",
		State: types.TaskStatePending,
	}

	child3 := &types.Task{
		ID:    uuid.MustParse("0ffa272b-8417-4d09-ae44-509d8a0caa74"),
		Title: "MCP Control Tools",
		State: types.TaskStatePending,
	}

	child4 := &types.Task{
		ID:    uuid.MustParse("bc71846b-8e11-45f0-b733-b16d76065294"),
		Title: "VectorService Integration",
		State: types.TaskStatePending,
	}

	// Build the tree structure as it would appear in the final output
	fmt.Println("=== Demo: Target Tree Format ===")
	fmt.Println()

	// Root task
	rootLine := formatter.FormatTaskLine(rootTask)
	fmt.Println(rootLine)

	// Child tasks
	parentPrefix := ""
	children := []*types.Task{child1, child2, child3, child4}

	for i, child := range children {
		isLast := i == len(children)-1
		prefix := formatter.GetTreePrefix(isLast, 1, parentPrefix)
		childLine := formatter.FormatTaskLine(child)
		fmt.Printf("%s%s\n", prefix, childLine)
	}

	fmt.Println()
	fmt.Println("=== Expected Output ===")
	fmt.Println("Core Architecture Implementation (ID: 0a4f4544-2802-493b-9986-e0d7b7a41d2e) - in-progress")
	fmt.Println("  ├── ✅ ConsolidationScheduler Service (ID: 3f22bffe-5004-4002-877c-da7be9276688) - completed")
	fmt.Println("  ├── ⏳ InMemoryState Management (ID: 7415bb39-7832-4dd7-b2a9-4fe56ad30985) - pending")
	fmt.Println("  ├── ⏳ MCP Control Tools (ID: 0ffa272b-8417-4d09-ae44-509d8a0caa74) - pending")
	fmt.Println("  └── ⏳ VectorService Integration (ID: bc71846b-8e11-45f0-b733-b16d76065294) - pending")
}
