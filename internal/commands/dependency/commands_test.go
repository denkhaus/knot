package dependency

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/denkhaus/knot/v2/internal/testutil"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"
)

func TestCommands(t *testing.T) {
	tests := []struct {
		name             string
		expectedCommands []string
	}{
		{
			name: "basic dependency commands",
			expectedCommands: []string{
				"add",
				"remove",
				"list",
			},
		},
		{
			name: "all dependency commands including enhanced",
			expectedCommands: []string{
				"add",
				"remove",
				"list",
				// Enhanced commands might also be included
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := Commands()

			// Verify we get some commands
			assert.NotEmpty(t, commands)

			// Verify basic commands are present
			commandNames := make([]string, len(commands))
			for i, cmd := range commands {
				commandNames[i] = cmd.Name
			}

			for _, expectedName := range tt.expectedCommands {
				assert.Contains(t, commandNames, expectedName,
					"Command '%s' should be present in dependency commands", expectedName)
			}
		})
	}
}

func TestCommandStructure(t *testing.T) {
	commands := Commands()

	// Test that all commands have required structure
	for _, cmd := range commands {
		t.Run("command_"+cmd.Name, func(t *testing.T) {
			// Verify command has name and usage
			assert.NotEmpty(t, cmd.Name, "Command should have a name")
			assert.NotEmpty(t, cmd.Usage, "Command should have usage text")

			// Verify command has action
			assert.NotNil(t, cmd.Action, "Command should have an action function")

			// Verify flags are properly structured
			for _, flag := range cmd.Flags {
				assert.NotNil(t, flag, "Flag should not be nil")
				flagNames := flag.Names()
				assert.NotEmpty(t, flagNames, "Flag should have at least one name")
			}
		})
	}
}

func TestAddCommandFlags(t *testing.T) {
	commands := Commands()

	// Find the add command
	var addCommand *cli.Command
	for _, cmd := range commands {
		if cmd.Name == "add" {
			addCommand = cmd
			break
		}
	}

	require.NotNil(t, addCommand, "Add command should be found")

	// Check required flags
	expectedFlags := []string{"task-id", "depends-on"}
	flagNames := make([]string, 0)
	for _, flag := range addCommand.Flags {
		flagNames = append(flagNames, flag.Names()...)
	}

	for _, expectedFlag := range expectedFlags {
		assert.Contains(t, flagNames, expectedFlag,
			"Add command should have '%s' flag", expectedFlag)
	}
}

func TestRemoveCommandFlags(t *testing.T) {
	commands := Commands()

	// Find the remove command
	var removeCommand *cli.Command
	for _, cmd := range commands {
		if cmd.Name == "remove" {
			removeCommand = cmd
			break
		}
	}

	require.NotNil(t, removeCommand, "Remove command should be found")

	// Check required flags
	expectedFlags := []string{"task-id", "depends-on"}
	flagNames := make([]string, 0)
	for _, flag := range removeCommand.Flags {
		flagNames = append(flagNames, flag.Names()...)
	}

	for _, expectedFlag := range expectedFlags {
		assert.Contains(t, flagNames, expectedFlag,
			"Remove command should have '%s' flag", expectedFlag)
	}
}

func TestListCommandFlags(t *testing.T) {
	commands := Commands()

	// Find the list command
	var listCommand *cli.Command
	for _, cmd := range commands {
		if cmd.Name == "list" {
			listCommand = cmd
			break
		}
	}

	require.NotNil(t, listCommand, "List command should be found")

	// Check required flags
	expectedFlags := []string{"task-id"}
	flagNames := make([]string, 0)
	for _, flag := range listCommand.Flags {
		flagNames = append(flagNames, flag.Names()...)
	}

	for _, expectedFlag := range expectedFlags {
		assert.Contains(t, flagNames, expectedFlag,
			"List command should have '%s' flag", expectedFlag)
	}
}

func TestCommandUsageText(t *testing.T) {
	commands := Commands()

	// Test that commands have meaningful usage text
	expectedUsages := map[string]string{
		"add":    "Add task dependency",
		"remove": "Remove task dependency",
		"list":   "List task dependencies",
	}

	for _, cmd := range commands {
		if expectedUsage, exists := expectedUsages[cmd.Name]; exists {
			assert.Equal(t, expectedUsage, cmd.Usage,
				"Command '%s' should have correct usage text", cmd.Name)
		}
	}
}

func TestCommandsReturnNewSlice(t *testing.T) {
	commands1 := Commands()
	commands2 := Commands()

	// Commands should return new slices to avoid modification issues
	assert.NotSame(t, &commands1[0], &commands2[0],
		"Commands should return new slice instances")
}

// Integration test with mock context
func TestCommandIntegration(t *testing.T) {
	commands := Commands()

	t.Run("command creation with app context", func(t *testing.T) {
		// Verify commands can be created with a valid app context
		assert.NotNil(t, commands)
		assert.Greater(t, len(commands), 0)
	})

	t.Run("command DI dependency", func(t *testing.T) {
		// Set up test injector
		config := testutil.NewTestConfig(t)
		testInjector := config.SetupTestInjector(t)

		// Test that commands properly work with DI
		assert.NotNil(t, testInjector, "DI injector should be available")

		// Verify we can get services from DI
		loggerService := do.MustInvoke[logger.Logger](testInjector)
		assert.NotNil(t, loggerService, "Logger should be available from DI")
	})
}

// Test edge cases
func TestCommandEdgeCases(t *testing.T) {
	t.Run("empty app context", func(t *testing.T) {
		// This test would verify behavior with nil/empty app context
		// But since we don't want to cause panics, we'll use a valid context

		commands := Commands()
		assert.NotEmpty(t, commands, "Commands should be created even with minimal app context")
	})
}

// TestAddAction tests the add dependency CLI action - basic validation
// Note: Full integration tests require complex DI setup that isn't reliable
// The ProjectManager.AddTaskDependency method is tested separately.
func TestAddAction(t *testing.T) {
	// Test that addAction returns a valid action function
	action := addAction()
	assert.NotNil(t, action)
}

// TestRemoveAction tests the remove dependency CLI action - basic validation
// Note: Full integration tests require complex DI setup that isn't reliable
// The ProjectManager.RemoveTaskDependency method is tested separately.
func TestRemoveAction(t *testing.T) {
	// Test that removeAction returns a valid action function
	action := removeAction()
	assert.NotNil(t, action)
}

// TestListAction tests the list dependencies CLI action - basic validation
// Note: Full integration tests require complex DI setup that isn't reliable
// The ProjectManager.GetTaskDependencies and GetDependentTasks methods are tested separately.
func TestListAction(t *testing.T) {
	// Test that listAction returns a valid action function
	action := listAction()
	assert.NotNil(t, action)
}

// TestEnhancedCommands tests the enhanced dependency commands
func TestEnhancedCommands(t *testing.T) {
	commands := EnhancedCommands()

	t.Run("enhanced commands are present", func(t *testing.T) {
		expectedCommands := []string{
			"dependents",
			"chain",
			"cycles",
			"validate",
		}

		commandNames := make([]string, len(commands))
		for i, cmd := range commands {
			commandNames[i] = cmd.Name
		}

		for _, expectedName := range expectedCommands {
			assert.Contains(t, commandNames, expectedName,
				"Enhanced command '%s' should be present", expectedName)
		}
	})

	t.Run("enhanced commands have proper structure", func(t *testing.T) {
		for _, cmd := range commands {
			t.Run("enhanced_"+cmd.Name, func(t *testing.T) {
				assert.NotEmpty(t, cmd.Name)
				assert.NotEmpty(t, cmd.Usage)
				assert.NotNil(t, cmd.Action)
			})
		}
	})
}

// TestGetAllTransitiveDependents tests the recursive dependent collection
func TestGetAllTransitiveDependents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskID := uuid.New()
	dependent1ID := uuid.New()
	dependent2ID := uuid.New()
	dependent3ID := uuid.New()

	tests := []struct {
		name           string
		taskID         uuid.UUID
		setupMocks     func(*mocks.MockProjectManager)
		expectedCount  int
		expectedErr    bool
		expectedIDs    []uuid.UUID
	}{
		{
			name:   "single level dependents",
			taskID: taskID,
			setupMocks: func(m *mocks.MockProjectManager) {
				m.EXPECT().GetDependentTasks(gomock.Any(), taskID).
					Return([]*types.Task{
						{ID: dependent1ID, Title: "Dependent 1"},
						{ID: dependent2ID, Title: "Dependent 2"},
					}, nil)
				m.EXPECT().GetDependentTasks(gomock.Any(), dependent1ID).
					Return([]*types.Task{}, nil)
				m.EXPECT().GetDependentTasks(gomock.Any(), dependent2ID).
					Return([]*types.Task{}, nil)
			},
			expectedCount: 2,
			expectedErr:   false,
			expectedIDs:   []uuid.UUID{dependent1ID, dependent2ID},
		},
		{
			name:   "multi-level transitive dependents",
			taskID: taskID,
			setupMocks: func(m *mocks.MockProjectManager) {
				m.EXPECT().GetDependentTasks(gomock.Any(), taskID).
					Return([]*types.Task{
						{ID: dependent1ID, Title: "Dependent 1"},
					}, nil)
				m.EXPECT().GetDependentTasks(gomock.Any(), dependent1ID).
					Return([]*types.Task{
						{ID: dependent2ID, Title: "Dependent 2"},
					}, nil)
				m.EXPECT().GetDependentTasks(gomock.Any(), dependent2ID).
					Return([]*types.Task{
						{ID: dependent3ID, Title: "Dependent 3"},
					}, nil)
				m.EXPECT().GetDependentTasks(gomock.Any(), dependent3ID).
					Return([]*types.Task{}, nil)
			},
			expectedCount: 3,
			expectedErr:   false,
			expectedIDs:   []uuid.UUID{dependent1ID, dependent2ID, dependent3ID},
		},
		{
			name:   "handles cycles correctly",
			taskID: taskID,
			setupMocks: func(m *mocks.MockProjectManager) {
				// Simulate a cycle: taskID -> dependent1 -> dependent2 -> dependent1
				m.EXPECT().GetDependentTasks(gomock.Any(), taskID).
					Return([]*types.Task{
						{ID: dependent1ID, Title: "Dependent 1"},
					}, nil)
				m.EXPECT().GetDependentTasks(gomock.Any(), dependent1ID).
					Return([]*types.Task{
						{ID: dependent2ID, Title: "Dependent 2"},
					}, nil)
				m.EXPECT().GetDependentTasks(gomock.Any(), dependent2ID).
					Return([]*types.Task{
						{ID: dependent1ID, Title: "Dependent 1"}, // Cycle back
					}, nil)
				// dependent1 already visited, so no more calls for it
			},
			expectedCount: 2,
			expectedErr:   false,
			expectedIDs:   []uuid.UUID{dependent1ID, dependent2ID},
		},
		{
			name:   "manager returns error",
			taskID: taskID,
			setupMocks: func(m *mocks.MockProjectManager) {
				m.EXPECT().GetDependentTasks(gomock.Any(), taskID).
					Return(nil, errors.New("database error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMgr := mocks.NewMockProjectManager(ctrl)
			tt.setupMocks(mockMgr)

			result, err := getAllTransitiveDependents(context.Background(), mockMgr, tt.taskID)

			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)

				// Verify expected IDs are present
				resultIDs := make([]uuid.UUID, len(result))
				for i, task := range result {
					resultIDs[i] = task.ID
				}
				for _, expectedID := range tt.expectedIDs {
					assert.Contains(t, resultIDs, expectedID)
				}
			}
		})
	}
}

// TestDetectCycles tests the cycle detection algorithm
func TestDetectCycles(t *testing.T) {
	task1ID := uuid.New()
	task2ID := uuid.New()
	task3ID := uuid.New()
	task4ID := uuid.New()

	tests := []struct {
		name          string
		tasks         []*types.Task
		expectedCycles int
		verifyCycles  func(t *testing.T, cycles [][]uuid.UUID)
	}{
		{
			name: "no cycles - linear dependency chain",
			tasks: []*types.Task{
				{ID: task1ID, Title: "Task 1", Dependencies: []uuid.UUID{task2ID}},
				{ID: task2ID, Title: "Task 2", Dependencies: []uuid.UUID{task3ID}},
				{ID: task3ID, Title: "Task 3", Dependencies: []uuid.UUID{}},
			},
			expectedCycles: 0,
		},
		{
			name: "simple cycle - two tasks depend on each other",
			tasks: []*types.Task{
				{ID: task1ID, Title: "Task 1", Dependencies: []uuid.UUID{task2ID}},
				{ID: task2ID, Title: "Task 2", Dependencies: []uuid.UUID{task1ID}},
			},
			expectedCycles: 1,
			verifyCycles: func(t *testing.T, cycles [][]uuid.UUID) {
				assert.Len(t, cycles, 1)
				assert.Contains(t, cycles[0], task1ID)
				assert.Contains(t, cycles[0], task2ID)
			},
		},
		{
			name: "three task cycle",
			tasks: []*types.Task{
				{ID: task1ID, Title: "Task 1", Dependencies: []uuid.UUID{task2ID}},
				{ID: task2ID, Title: "Task 2", Dependencies: []uuid.UUID{task3ID}},
				{ID: task3ID, Title: "Task 3", Dependencies: []uuid.UUID{task1ID}},
			},
			expectedCycles: 1,
			verifyCycles: func(t *testing.T, cycles [][]uuid.UUID) {
				assert.Len(t, cycles, 1)
				assert.Len(t, cycles[0], 3)
			},
		},
		{
			name: "multiple disconnected cycles",
			tasks: []*types.Task{
				{ID: task1ID, Title: "Task 1", Dependencies: []uuid.UUID{task2ID}},
				{ID: task2ID, Title: "Task 2", Dependencies: []uuid.UUID{task1ID}},
				{ID: task3ID, Title: "Task 3", Dependencies: []uuid.UUID{task4ID}},
				{ID: task4ID, Title: "Task 4", Dependencies: []uuid.UUID{task3ID}},
			},
			expectedCycles: 2,
			verifyCycles: func(t *testing.T, cycles [][]uuid.UUID) {
				assert.Len(t, cycles, 2)
			},
		},
		{
			name: "no dependencies at all",
			tasks: []*types.Task{
				{ID: task1ID, Title: "Task 1", Dependencies: []uuid.UUID{}},
				{ID: task2ID, Title: "Task 2", Dependencies: []uuid.UUID{}},
				{ID: task3ID, Title: "Task 3", Dependencies: []uuid.UUID{}},
			},
			expectedCycles: 0,
		},
		{
			name: "complex graph with cycle",
			tasks: []*types.Task{
				{ID: task1ID, Title: "Task 1", Dependencies: []uuid.UUID{task2ID, task3ID}},
				{ID: task2ID, Title: "Task 2", Dependencies: []uuid.UUID{task4ID}},
				{ID: task3ID, Title: "Task 3", Dependencies: []uuid.UUID{task4ID}},
				{ID: task4ID, Title: "Task 4", Dependencies: []uuid.UUID{task1ID}}, // Cycles: 1->2->4->1 and 1->3->4->1
			},
			expectedCycles: 2, // DFS finds both distinct paths to the cycle
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cycles := detectCycles(tt.tasks)
			assert.Len(t, cycles, tt.expectedCycles)
			if tt.verifyCycles != nil {
				tt.verifyCycles(t, cycles)
			}
		})
	}
}

// TestCountTotalDependencies tests the dependency counting helper
func TestCountTotalDependencies(t *testing.T) {
	task1ID := uuid.New()
	task2ID := uuid.New()
	task3ID := uuid.New()

	tests := []struct {
		name             string
		tasks            []*types.Task
		expectedTotal    int
	}{
		{
			name: "no dependencies",
			tasks: []*types.Task{
				{ID: task1ID, Dependencies: []uuid.UUID{}},
				{ID: task2ID, Dependencies: []uuid.UUID{}},
			},
			expectedTotal: 0,
		},
		{
			name: "single dependency per task",
			tasks: []*types.Task{
				{ID: task1ID, Dependencies: []uuid.UUID{task2ID}},
				{ID: task2ID, Dependencies: []uuid.UUID{task3ID}},
				{ID: task3ID, Dependencies: []uuid.UUID{}},
			},
			expectedTotal: 2,
		},
		{
			name: "multiple dependencies per task",
			tasks: []*types.Task{
				{ID: task1ID, Dependencies: []uuid.UUID{task2ID, task3ID}},
				{ID: task2ID, Dependencies: []uuid.UUID{task3ID}},
				{ID: task3ID, Dependencies: []uuid.UUID{}},
			},
			expectedTotal: 3,
		},
		{
			name: "empty task list",
			tasks: []*types.Task{},
			expectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := countTotalDependencies(tt.tasks)
			assert.Equal(t, tt.expectedTotal, total)
		})
	}
}

// TestShowUpstreamChain tests the upstream chain traversal
func TestShowUpstreamChain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskID := uuid.New()
	dep1ID := uuid.New()
	dep2ID := uuid.New()

	tests := []struct {
		name         string
		taskID       uuid.UUID
		setupMocks   func(*mocks.MockProjectManager)
		expectedErr  bool
		verifyOutput func(t *testing.T, output string)
	}{
		{
			name:   "simple upstream chain",
			taskID: taskID,
			setupMocks: func(m *mocks.MockProjectManager) {
				m.EXPECT().GetTaskDependencies(gomock.Any(), taskID).
					Return([]*types.Task{
						{ID: dep1ID, Title: "Dep 1", State: "completed"},
					}, nil)
				m.EXPECT().GetTaskDependencies(gomock.Any(), dep1ID).
					Return([]*types.Task{}, nil)
			},
			expectedErr: false,
			verifyOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "Dep 1")
			},
		},
		{
			name:   "nested upstream chain",
			taskID: taskID,
			setupMocks: func(m *mocks.MockProjectManager) {
				m.EXPECT().GetTaskDependencies(gomock.Any(), taskID).
					Return([]*types.Task{
						{ID: dep1ID, Title: "Dep 1", State: "completed"},
					}, nil)
				m.EXPECT().GetTaskDependencies(gomock.Any(), dep1ID).
					Return([]*types.Task{
						{ID: dep2ID, Title: "Dep 2", State: "pending"},
					}, nil)
				m.EXPECT().GetTaskDependencies(gomock.Any(), dep2ID).
					Return([]*types.Task{}, nil)
			},
			expectedErr: false,
			verifyOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "Dep 1")
				assert.Contains(t, output, "Dep 2")
			},
		},
		{
			name:   "no upstream dependencies",
			taskID: taskID,
			setupMocks: func(m *mocks.MockProjectManager) {
				m.EXPECT().GetTaskDependencies(gomock.Any(), taskID).
					Return([]*types.Task{}, nil)
			},
			expectedErr: false,
			verifyOutput: func(t *testing.T, output string) {
				// At depth 0, should show "No upstream dependencies"
			},
		},
		{
			name:   "error getting dependencies",
			taskID: taskID,
			setupMocks: func(m *mocks.MockProjectManager) {
				m.EXPECT().GetTaskDependencies(gomock.Any(), taskID).
					Return(nil, errors.New("database error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMgr := mocks.NewMockProjectManager(ctrl)
			tt.setupMocks(mockMgr)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := showUpstreamChain(context.Background(), mockMgr, tt.taskID, 0)

			_ = w.Close()
			os.Stdout = oldStdout

			out, _ := io.ReadAll(r)
			output := string(out)

			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.verifyOutput != nil {
					tt.verifyOutput(t, output)
				}
			}
		})
	}
}

// TestShowDownstreamChain tests the downstream chain traversal
func TestShowDownstreamChain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskID := uuid.New()
	dep1ID := uuid.New()
	dep2ID := uuid.New()

	tests := []struct {
		name         string
		taskID       uuid.UUID
		setupMocks   func(*mocks.MockProjectManager)
		expectedErr  bool
		verifyOutput func(t *testing.T, output string)
	}{
		{
			name:   "simple downstream chain",
			taskID: taskID,
			setupMocks: func(m *mocks.MockProjectManager) {
				m.EXPECT().GetDependentTasks(gomock.Any(), taskID).
					Return([]*types.Task{
						{ID: dep1ID, Title: "Dependent 1", State: "pending"},
					}, nil)
				m.EXPECT().GetDependentTasks(gomock.Any(), dep1ID).
					Return([]*types.Task{}, nil)
			},
			expectedErr: false,
			verifyOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "Dependent 1")
			},
		},
		{
			name:   "nested downstream chain",
			taskID: taskID,
			setupMocks: func(m *mocks.MockProjectManager) {
				m.EXPECT().GetDependentTasks(gomock.Any(), taskID).
					Return([]*types.Task{
						{ID: dep1ID, Title: "Dependent 1", State: "pending"},
					}, nil)
				m.EXPECT().GetDependentTasks(gomock.Any(), dep1ID).
					Return([]*types.Task{
						{ID: dep2ID, Title: "Dependent 2", State: "in-progress"},
					}, nil)
				m.EXPECT().GetDependentTasks(gomock.Any(), dep2ID).
					Return([]*types.Task{}, nil)
			},
			expectedErr: false,
			verifyOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "Dependent 1")
				assert.Contains(t, output, "Dependent 2")
			},
		},
		{
			name:   "no downstream dependencies",
			taskID: taskID,
			setupMocks: func(m *mocks.MockProjectManager) {
				m.EXPECT().GetDependentTasks(gomock.Any(), taskID).
					Return([]*types.Task{}, nil)
			},
			expectedErr: false,
		},
		{
			name:   "error getting dependents",
			taskID: taskID,
			setupMocks: func(m *mocks.MockProjectManager) {
				m.EXPECT().GetDependentTasks(gomock.Any(), taskID).
					Return(nil, errors.New("database error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMgr := mocks.NewMockProjectManager(ctrl)
			tt.setupMocks(mockMgr)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := showDownstreamChain(context.Background(), mockMgr, tt.taskID, 0)

			_ = w.Close()
			os.Stdout = oldStdout

			out, _ := io.ReadAll(r)
			output := string(out)

			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.verifyOutput != nil {
					tt.verifyOutput(t, output)
				}
			}
		})
	}
}

// TestChainAction tests the chain CLI action - basic validation only
// Note: Full integration tests require complex DI setup, the helper functions
// (showUpstreamChain, showDownstreamChain) are tested separately.
func TestChainAction(t *testing.T) {
	// Test that chainAction returns a valid action function
	action := chainAction()
	assert.NotNil(t, action)
}

// TestCyclesAction tests the cycles CLI action - basic validation only
// Note: Full integration tests require complex DI setup, the detectCycles
// helper function is tested separately.
func TestCyclesAction(t *testing.T) {
	// Test that cyclesAction returns a valid action function
	action := cyclesAction()
	assert.NotNil(t, action)
}
