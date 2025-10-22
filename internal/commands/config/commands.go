package config

import (
	"fmt"

	"github.com/denkhaus/knot/internal/manager"
	"github.com/denkhaus/knot/internal/shared"
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
					Usage:    "Configuration key (complexity-threshold, max-depth, max-tasks-per-depth, max-description-length, auto-reduce-complexity)",
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
	return func(c *cli.Context) error {
		config := appCtx.ProjectManager.GetConfig()

		fmt.Println("Current Knot Configuration:")
		fmt.Println()
		fmt.Printf("  Complexity Threshold:    %d (tasks >= this need breakdown)\n", config.ComplexityThreshold)
		fmt.Printf("  Max Depth:               %d (maximum hierarchy levels)\n", config.MaxDepth)
		fmt.Printf("  Max Tasks Per Depth:     %d (maximum tasks per level)\n", config.MaxTasksPerDepth)
		fmt.Printf("  Max Description Length:  %d (maximum characters)\n", config.MaxDescriptionLength)
		fmt.Printf("  Auto Reduce Complexity:  %t (automatically reduce parent complexity when subtasks added)\n", config.AutoReduceComplexity)
		fmt.Println()

		// Show config file location - TODO: implement GetConfigPath method
		// configPath, err := config.GetConfigPath()
		// if err == nil {
		// 	fmt.Printf("Configuration file: %s\n", configPath)
		// }

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
		case "complexity-threshold":
			if value < 1 || value > 10 {
				return fmt.Errorf("complexity-threshold must be between 1 and 10, got %d", value)
			}
			newConfig.ComplexityThreshold = value
		case "max-depth":
			if value < 1 {
				return fmt.Errorf("max-depth must be at least 1, got %d", value)
			}
			newConfig.MaxDepth = value
		case "max-tasks-per-depth":
			if value < 1 {
				return fmt.Errorf("max-tasks-per-depth must be at least 1, got %d", value)
			}
			newConfig.MaxTasksPerDepth = value
		case "max-description-length":
			if value < 1 {
				return fmt.Errorf("max-description-length must be at least 1, got %d", value)
			}
			newConfig.MaxDescriptionLength = value
		case "auto-reduce-complexity":
			// Convert int to bool: 0 = false, 1 = true
			if value != 0 && value != 1 {
				return fmt.Errorf("auto-reduce-complexity must be 0 (false) or 1 (true), got %d", value)
			}
			newConfig.AutoReduceComplexity = value == 1
		default:
			return fmt.Errorf("unknown configuration key: %s. Valid keys: complexity-threshold, max-depth, max-tasks-per-depth, max-description-length, auto-reduce-complexity", key)
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
	return func(c *cli.Context) error {
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
