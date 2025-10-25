package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestNewTaskLimitFlag(t *testing.T) {
	flag := NewTaskLimitFlag()

	assert.Equal(t, "limit", flag.Names()[0])
	assert.Equal(t, []string{"l"}, flag.Names()[1:])
	assert.Equal(t, "Maximum number of tasks to show (default: 10)", getUsage(flag))

	// Test that it's an IntFlag
	intFlag, ok := flag.(*cli.IntFlag)
	assert.True(t, ok)
	assert.Equal(t, 10, intFlag.Value)
}

func TestNewJSONFlag(t *testing.T) {
	flag := NewJSONFlag()

	assert.Equal(t, "json", flag.Names()[0])
	assert.Equal(t, []string{"j"}, flag.Names()[1:])
	assert.Equal(t, "Output in JSON format", getUsage(flag))

	// Test that it's a BoolFlag
	_, ok := flag.(*cli.BoolFlag)
	assert.True(t, ok)
}

func TestNewLogLevelFlag(t *testing.T) {
	flag := NewLogLevelFlag()

	assert.Equal(t, "log-level", flag.Names()[0])
	assert.Equal(t, "Log level (off, error, warn, info, debug)", getUsage(flag))

	// Test that it's a StringFlag
	stringFlag, ok := flag.(*cli.StringFlag)
	assert.True(t, ok)
	assert.Equal(t, "off", stringFlag.Value)
}

func TestTaskLimitFlagDefaultValue(t *testing.T) {
	flag := NewTaskLimitFlag()
	intFlag, ok := flag.(*cli.IntFlag)
	assert.True(t, ok)
	assert.Equal(t, 10, intFlag.Value)
}

func TestJSONFlagDefaultValue(t *testing.T) {
	flag := NewJSONFlag()
	boolFlag, ok := flag.(*cli.BoolFlag)
	assert.True(t, ok)
	assert.False(t, boolFlag.Value) // Should default to false
}

func TestLogLevelFlagDefaultValue(t *testing.T) {
	flag := NewLogLevelFlag()
	stringFlag, ok := flag.(*cli.StringFlag)
	assert.True(t, ok)
	assert.Equal(t, "off", stringFlag.Value)
}

func TestFlagAliases(t *testing.T) {
	// Test task limit flag aliases
	taskLimitFlag := NewTaskLimitFlag()
	assert.Contains(t, taskLimitFlag.Names(), "limit")
	assert.Contains(t, taskLimitFlag.Names(), "l")

	// Test JSON flag aliases
	jsonFlag := NewJSONFlag()
	assert.Contains(t, jsonFlag.Names(), "json")
	assert.Contains(t, jsonFlag.Names(), "j")

	// Test log level flag has no aliases
	logLevelFlag := NewLogLevelFlag()
	assert.Contains(t, logLevelFlag.Names(), "log-level")
	assert.Len(t, logLevelFlag.Names(), 1) // Only one name, no aliases
}

func TestFlagUsageStrings(t *testing.T) {
	// Test all usage strings are properly set
	taskLimitFlag := NewTaskLimitFlag()
	assert.Equal(t, "Maximum number of tasks to show (default: 10)", getUsage(taskLimitFlag))

	jsonFlag := NewJSONFlag()
	assert.Equal(t, "Output in JSON format", getUsage(jsonFlag))

	logLevelFlag := NewLogLevelFlag()
	assert.Equal(t, "Log level (off, error, warn, info, debug)", getUsage(logLevelFlag))
}

// Helper function to get usage from any flag type
func getUsage(flag cli.Flag) string {
	switch f := flag.(type) {
	case *cli.StringFlag:
		return f.Usage
	case *cli.IntFlag:
		return f.Usage
	case *cli.BoolFlag:
		return f.Usage
	default:
		return ""
	}
}

// Integration test to ensure flags work properly when used in CLI context
func TestFlagsIntegration(t *testing.T) {
	// Test that flags can be properly initialized and used
	flags := []cli.Flag{
		NewTaskLimitFlag(),
		NewJSONFlag(),
		NewLogLevelFlag(),
	}

	// Create a mock app to test flag initialization
	app := &cli.App{
		Name:  "test",
		Flags: flags,
		Action: func(c *cli.Context) error {
			// Verify that the flags were processed correctly
			assert.Equal(t, 10, c.Int("limit"))
			assert.False(t, c.Bool("json"))
			assert.Equal(t, "off", c.String("log-level"))
			return nil
		},
	}

	// Run with default values (no flags provided)
	err := app.Run([]string{"test"})
	assert.NoError(t, err)

	// Run with custom values
	err = app.Run([]string{"test", "--limit", "5", "--json", "--log-level", "debug"})
	assert.NoError(t, err)
}

func TestTaskLimitFlagWithCustomValue(t *testing.T) {
	app := &cli.App{
		Name:  "test",
		Flags: []cli.Flag{NewTaskLimitFlag()},
		Action: func(c *cli.Context) error {
			assert.Equal(t, 25, c.Int("limit"))
			return nil
		},
	}

	err := app.Run([]string{"test", "--limit", "25"})
	assert.NoError(t, err)
}

func TestJSONFlagWithTrueValue(t *testing.T) {
	app := &cli.App{
		Name:  "test",
		Flags: []cli.Flag{NewJSONFlag()},
		Action: func(c *cli.Context) error {
			assert.True(t, c.Bool("json"))
			return nil
		},
	}

	err := app.Run([]string{"test", "--json"})
	assert.NoError(t, err)
}

func TestLogLevelFlagWithCustomValue(t *testing.T) {
	app := &cli.App{
		Name:  "test",
		Flags: []cli.Flag{NewLogLevelFlag()},
		Action: func(c *cli.Context) error {
			assert.Equal(t, "info", c.String("log-level"))
			return nil
		},
	}

	err := app.Run([]string{"test", "--log-level", "info"})
	assert.NoError(t, err)
}

func TestFlagShortAliases(t *testing.T) {
	app := &cli.App{
		Name: "test",
		Flags: []cli.Flag{
			NewTaskLimitFlag(),
			NewJSONFlag(),
		},
		Action: func(c *cli.Context) error {
			assert.Equal(t, 15, c.Int("limit"))
			assert.True(t, c.Bool("json"))
			return nil
		},
	}

	err := app.Run([]string{"test", "-l", "15", "-j"})
	assert.NoError(t, err)
}
