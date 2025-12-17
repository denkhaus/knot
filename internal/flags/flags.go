package flags

import "github.com/urfave/cli/v2"

// NewJSONFlag creates a consistent JSON flag for all commands
func NewJSONFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:    "json",
		Aliases: []string{"j"},
		Usage:   "Output in JSON format",
	}
}

// NewQuietFlag creates a consistent quiet flag for all commands
func NewQuietFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:  "quiet",
		Usage: "Suppress project context display",
	}
}

// NewTaskLimitFlag creates a flag for limiting the number of tasks shown
func NewTaskLimitFlag() cli.Flag {
	return &cli.IntFlag{
		Name:    "limit",
		Aliases: []string{"l"},
		Usage:   "Maximum number of tasks to show (default: 10)",
		Value:   10,
		EnvVars: []string{"KNOT_TASK_LIMIT"},
	}
}

// NewTaskIDFlag creates a flag for specifying a task ID
func NewTaskIDFlag() cli.Flag {
	return &cli.StringFlag{
		Name:     "id",
		Usage:    "Task ID",
		Required: true,
	}
}

// NewLogLevelFlag creates a flag for setting the log level
func NewLogLevelFlag() cli.Flag {
	return &cli.StringFlag{
		Name:  "log-level",
		Usage: "Log level (off, error, warn, info, debug)",
		Value: "off",
	}
}

// NewManagerConfigFlags creates flags for manager configuration
func NewManagerConfigFlags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:    "max-tasks-per-depth",
			Usage:   "Maximum tasks allowed per depth level",
			Value:   100,
			EnvVars: []string{"KNOT_MAX_TASKS_PER_DEPTH"},
		},
		&cli.IntFlag{
			Name:    "complexity-threshold",
			Usage:   "Threshold for task breakdown suggestions (1-10)",
			Value:   8,
			EnvVars: []string{"KNOT_COMPLEXITY_THRESHOLD"},
		},
		&cli.IntFlag{
			Name:    "max-depth",
			Usage:   "Maximum allowed task depth",
			Value:   5,
			EnvVars: []string{"KNOT_MAX_DEPTH"},
		},
		&cli.IntFlag{
			Name:    "max-description-length",
			Usage:   "Maximum length for task descriptions",
			Value:   2000,
			EnvVars: []string{"KNOT_MAX_DESCRIPTION_LENGTH"},
		},
		&cli.BoolFlag{
			Name:    "auto-reduce-complexity",
			Usage:   "Automatically reduce parent task complexity when subtasks are added",
			Value:   true,
			EnvVars: []string{"KNOT_AUTO_REDUCE_COMPLEXITY"},
		},
	}
}
