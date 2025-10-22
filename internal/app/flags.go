package app

import (
	"github.com/urfave/cli/v2"
)

func NewProjectIDFlag() cli.Flag {
	return &cli.StringFlag{
		Name:    "project-id",
		Aliases: []string{"p"},
		Usage:   "Project ID",
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

func NewJSONFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:    "json",
		Aliases: []string{"j"},
		Usage:   "Output in JSON format",
	}
}

func NewLogLevelFlag() cli.Flag {
	return &cli.StringFlag{
		Name:  "log-level",
		Usage: "Log level (off, error, warn, info, debug)",
		Value: "off",
	}
}
