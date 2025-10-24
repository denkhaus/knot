package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/denkhaus/knot/internal/shared"
	"github.com/denkhaus/knot/internal/types"
	internalValidation "github.com/denkhaus/knot/internal/validation"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// Commands returns validation related CLI commands
func Commands(appCtx *shared.AppContext) []*cli.Command {
	return []*cli.Command{
		{
			Name:   "states",
			Usage:  "Show valid task states and transitions",
			Action: statesAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "from",
					Usage: "Show valid transitions from specific state",
				},
				&cli.BoolFlag{
					Name:  "matrix",
					Usage: "Show complete transition matrix",
					Value: false,
				},
			},
		},
		{
			Name:   "transition",
			Usage:  "Validate a state transition without applying it",
			Action: transitionAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "task-id",
					Usage:    "Task ID to validate transition for",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "to",
					Usage:    "Target state to validate",
					Required: true,
				},
				&cli.BoolFlag{
					Name:  "lenient",
					Usage: "Use lenient validation (show warnings instead of errors)",
					Value: false,
				},
			},
		},
		{
			Name:   "project",
			Usage:  "Validate all task states in a project",
			Action: projectAction(appCtx),
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "fix",
					Usage: "Attempt to fix invalid states automatically",
					Value: false,
				},
			},
		},
	}
}

func statesAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		fromState := c.String("from")
		showMatrix := c.Bool("matrix")

		validator := internalValidation.NewStateValidator()

		if showMatrix {
			fmt.Println("Task State Transition Matrix:")
			matrix := validator.GetStateTransitionMatrix()

			for state, transitions := range matrix {
				fmt.Printf("* %s:\n", strings.ToUpper(state))
				if len(transitions) == 0 {
					fmt.Printf("   (no valid transitions)\n")
				} else {
					for _, target := range transitions {
						fmt.Printf("   ✅ → %s\n", target)
					}
				}
				fmt.Println()
			}
			return nil
		}

		if fromState != "" {
			if !validator.IsValidState(fromState) {
				return fmt.Errorf("invalid state: %s", fromState)
			}

			matrix := validator.GetStateTransitionMatrix()
			transitions := matrix[fromState]

			fmt.Printf("Valid transitions from '%s':\n\n", strings.ToUpper(fromState))
			if len(transitions) == 0 {
				fmt.Printf("   (no valid transitions)\n")
			} else {
				for _, target := range transitions {
					fmt.Printf("   ✅ %s → %s\n", fromState, target)
				}
			}
			return nil
		}

		// Show all valid states
		fmt.Println("Valid Task States:")
		states := validator.GetAllValidStates()
		for _, state := range states {
			fmt.Printf("   * %s\n", strings.ToUpper(string(state)))
		}

		fmt.Println("\nUse --matrix to see all valid transitions")
		fmt.Println("Use --from <state> to see transitions from a specific state")

		return nil
	}
}

func transitionAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		taskIDStr := c.String("task-id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			return fmt.Errorf("invalid task ID: %w", err)
		}

		toState := c.String("to")
		lenient := c.Bool("lenient")

		// Get task
		task, err := appCtx.ProjectManager.GetTask(context.Background(), taskID)
		if err != nil {
			return fmt.Errorf("failed to get task: %w", err)
		}

		validator := internalValidation.NewStateValidator()

		// Basic state validation
		if !validator.IsValidState(toState) {
			return fmt.Errorf("invalid target state: %s", toState)
		}

		targetState := types.TaskState(toState)

		fmt.Printf("Validating transition for task '%s':\n", task.Title)
		fmt.Printf("  Current State: %s\n", strings.ToUpper(string(task.State)))
		fmt.Printf("  Target State:  %s\n", strings.ToUpper(string(targetState)))
		fmt.Println()

		if lenient {
			err, warnings := validator.ValidateTransitionLenient(task.State, targetState, task)
			if err != nil {
				fmt.Printf("Transition INVALID: %v\n", err)
				return err
			}

			fmt.Printf("Transition VALID\n")
			if len(warnings) > 0 {
				fmt.Printf("\nWarnings:\n")
				for _, warning := range warnings {
					fmt.Printf("   %s\n", warning)
				}
			}
		} else {
			err := validator.ValidateTransition(task.State, targetState, task)
			if err != nil {
				fmt.Printf("Transition INVALID: %v\n", err)
				return err
			}
			fmt.Printf("Transition VALID\n")
		}

		fmt.Printf("\nTo apply this transition:\n")
		fmt.Printf("   knot task update-state --id %s --state %s\n", taskID, toState)

		return nil
	}
}

func projectAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		projectID, err := shared.ResolveProjectID(c, appCtx)
		if err != nil {
			return err
		}

		fix := c.Bool("fix")

		appCtx.Logger.Info("Validating project task states", zap.String("projectID", projectID.String()))

		// Get all tasks
		tasks, err := appCtx.ProjectManager.ListTasksForProject(context.Background(), projectID)
		if err != nil {
			return fmt.Errorf("failed to get project tasks: %w", err)
		}

		validator := internalValidation.NewStateValidator()
		var issues []string
		var fixedCount int

		fmt.Printf("Validating %d tasks in project %s:\n\n", len(tasks), projectID)

		for _, task := range tasks {
			// Check if state is valid
			if !validator.IsValidState(string(task.State)) {
				issue := fmt.Sprintf("Task '%s' has invalid state: %s", task.Title, task.State)
				issues = append(issues, issue)

				if fix {
					// Attempt to fix by setting to pending
					appCtx.Logger.Info("Fixing invalid state",
						zap.String("taskID", task.ID.String()),
						zap.String("invalidState", string(task.State)))

					_, err := appCtx.ProjectManager.UpdateTaskState(context.Background(), task.ID, types.TaskStatePending, appCtx.Actor)
					if err != nil {
						fmt.Printf("Failed to fix task %s: %v\n", task.ID, err)
					} else {
						fmt.Printf("Fixed task '%s': %s → pending\n", task.Title, task.State)
						fixedCount++
					}
				}
			}

			// Additional validations could be added here
			// e.g., check if blocked tasks have dependencies, etc.
		}

		fmt.Printf("\n📊 Validation Summary:\n")
		fmt.Printf("   Total Tasks: %d\n", len(tasks))
		fmt.Printf("   Issues Found: %d\n", len(issues))

		if fix && fixedCount > 0 {
			fmt.Printf("   Issues Fixed: %d\n", fixedCount)
		}

		if len(issues) > 0 {
			fmt.Printf("\nIssues Found:\n")
			for i, issue := range issues {
				fmt.Printf("   %d. %s\n", i+1, issue)
			}

			if !fix {
				fmt.Printf("\nUse --fix to automatically repair invalid states\n")
			}
		} else {
			fmt.Printf("\nAll task states are valid!\n")
		}

		return nil
	}
}
