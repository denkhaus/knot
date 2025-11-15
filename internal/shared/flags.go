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

func NewTaskLimitFlag() cli.Flag {
	return &cli.IntFlag{
		Name:    "limit",
		Aliases: []string{"l"},
		Usage:   "Maximum number of tasks to show (default: 10)",
		Value:   10,
		EnvVars: []string{"KNOT_TASK_LIMIT"},
	}
}

func NewTaskIDFlag() cli.Flag {
	return &cli.StringFlag{
		Name:     "id",
		Usage:    "Task ID",
		Required: true,
	}
}

func NewLogLevelFlag() cli.Flag {
	return &cli.StringFlag{
		Name:  "log-level",
		Usage: "Log level (off, error, warn, info, debug)",
		Value: "off",
	}
}
