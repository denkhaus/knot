// Package config provides CLI commands for configuration management in KNOT.
//
// This package implements command-line interface commands for managing
// application settings, project preferences, and system configuration.
//
// Key Commands:
//   - SetConfig: Updates configuration values
//   - GetConfig: Retrieves current configuration settings
//   - ResetConfig: Resets configuration to default values
//
// Cross-reference: Knot Task 86f3ba2d-3a87-493b-b8fc-96d19f344e89
package config

import (
	"fmt"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/urfave/cli/v2"
)

// Commands returns the config management commands
func Commands(appCtx *shared.AppContext) []*cli.Command {
	return []*cli.Command{
		{
			Name:   "show",
			Usage:  "Show current configuration",
			Action: ShowAction(appCtx),
		},
		{
			Name:   "set",
			Usage:  "Set configuration value",
			Action: SetAction(appCtx),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "key",
					Aliases:  []string{"k"},
					Usage:    "Configuration key (" + shared.ConfigKeyComplexityThreshold + ", " + shared.ConfigKeyMaxDepth + ", " + shared.ConfigKeyMaxTasksPerDepth + ", " + shared.ConfigKeyMaxDescriptionLength + ", " + shared.ConfigKeyAutoReduceComplexity + ")",
					Required: true,
				},
				&cli.IntFlag{
					Name:     "value",
					Aliases:  []string{"v"},
					Usage:    "Configuration value",
					Required: true,
				},
			},
		},
		{
			Name:   "reset",
			Usage:  "Reset configuration to defaults",
			Action: ResetAction(appCtx),
		},
	}
}

// ShowAction displays the current configuration
func ShowAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(_ *cli.Context) error {
		cfg := appCtx.ProjectManager.GetConfig()

		fmt.Println("Current Knot Configuration:")
		fmt.Println()
		fmt.Printf("  Complexity Threshold:    %d (tasks >= this need breakdown)\n", cfg.ComplexityThreshold)
		fmt.Printf("  Max Depth:               %d (maximum hierarchy levels)\n", cfg.MaxDepth)
		fmt.Printf("  Max Tasks Per Depth:     %d (maximum tasks per level)\n", cfg.MaxTasksPerDepth)
		fmt.Printf("  Max Description Length:  %d (maximum characters)\n", cfg.MaxDescriptionLength)
		fmt.Printf("  Auto Reduce Complexity:  %t (automatically reduce parent complexity when subtasks added)\n", cfg.AutoReduceComplexity)
		fmt.Println()

		// Show config file location
		configPath, err := config.GetConfigPath()
		if err == nil {
			fmt.Printf("Configuration file:        %s\n", configPath)
		} else {
			fmt.Printf("Configuration file:        Unable to determine path (%s)\n", err.Error())
		}

		return nil
	}
}

// SetAction sets a configuration value
func SetAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(c *cli.Context) error {
		key := c.String("key")
		value := c.Int("value")

		// Get current config
		currentConfig := appCtx.ProjectManager.GetConfig()
		newConfig := *currentConfig // Copy current config

		// Update the specified key
		switch key {
		case shared.ConfigKeyComplexityThreshold:
			if value < 1 || value > 10 {
				return fmt.Errorf("complexity-threshold must be between 1 and 10, got %d", value)
			}
			newConfig.ComplexityThreshold = value
		case shared.ConfigKeyMaxDepth:
			if value < 1 {
				return fmt.Errorf("max-depth must be at least 1, got %d", value)
			}
			newConfig.MaxDepth = value
		case shared.ConfigKeyMaxTasksPerDepth:
			if value < 1 {
				return fmt.Errorf("max-tasks-per-depth must be at least 1, got %d", value)
			}
			newConfig.MaxTasksPerDepth = value
		case shared.ConfigKeyMaxDescriptionLength:
			if value < 1 {
				return fmt.Errorf("max-description-length must be at least 1, got %d", value)
			}
			newConfig.MaxDescriptionLength = value
		case shared.ConfigKeyAutoReduceComplexity:
			// Convert int to bool: 0 = false, 1 = true
			if value != 0 && value != 1 {
				return fmt.Errorf("auto-reduce-complexity must be 0 (false) or 1 (true), got %d", value)
			}
			newConfig.AutoReduceComplexity = value == 1
		default:
			return fmt.Errorf("unknown configuration key: %s. Valid keys: "+shared.ConfigKeyComplexityThreshold+", "+shared.ConfigKeyMaxDepth+", "+shared.ConfigKeyMaxTasksPerDepth+", "+shared.ConfigKeyMaxDescriptionLength+", "+shared.ConfigKeyAutoReduceComplexity, key)
		}

		// Update and save config
		appCtx.ProjectManager.UpdateConfig(&newConfig)
		if err := appCtx.ProjectManager.SaveConfigToFile(); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("Configuration updated: %s = %d\n", key, value)
		return nil
	}
}

// ResetAction resets configuration to defaults
func ResetAction(appCtx *shared.AppContext) cli.ActionFunc {
	return func(_ *cli.Context) error {
		// Reset to default config
		defaultConfig := manager.DefaultConfig()
		appCtx.ProjectManager.UpdateConfig(defaultConfig)

		// Save to file
		if err := appCtx.ProjectManager.SaveConfigToFile(); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Println("Configuration reset to defaults:")
		fmt.Printf("  Complexity Threshold:    %d\n", defaultConfig.ComplexityThreshold)
		fmt.Printf("  Max Depth:               %d\n", defaultConfig.MaxDepth)
		fmt.Printf("  Max Tasks Per Depth:     %d\n", defaultConfig.MaxTasksPerDepth)
		fmt.Printf("  Max Description Length:  %d\n", defaultConfig.MaxDescriptionLength)
		fmt.Printf("  Auto Reduce Complexity:  %t\n", defaultConfig.AutoReduceComplexity)

		return nil
	}
}
