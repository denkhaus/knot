package shared

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
