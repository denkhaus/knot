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
	"strconv"

	knotconfig "github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/urfave/cli/v2"
)

// Commands returns the config management commands
func Commands() []*cli.Command {
	return []*cli.Command{
		{
			Name:   "show",
			Usage:  "Show current configuration",
			Action: ShowAction(),
		},
		{
			Name:   "set",
			Usage:  "Set configuration value",
			Action: SetAction(),
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
			Action: ResetAction(),
		},
	}
}

// ShowAction displays the current configuration
func ShowAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		// Get container from CLI context
		container := shared.GetContainerFromContext(c)
		configService, err := container.GetConfigService()
		if err != nil {
			return fmt.Errorf("Configuration error: %w", err)
		}
		cfg := configService.GetManagerConfig()

		fmt.Println("Current Knot Configuration:")
		fmt.Println()
		fmt.Printf("  Complexity Threshold:    %d (tasks >= this need breakdown)\n", cfg.ComplexityThreshold)
		fmt.Printf("  Max Depth:               %d (maximum hierarchy levels)\n", cfg.MaxDepth)
		fmt.Printf("  Max Tasks Per Depth:     %d (maximum tasks per level)\n", cfg.MaxTasksPerDepth)
		fmt.Printf("  Max Description Length:  %d (maximum characters)\n", cfg.MaxDescriptionLength)
		fmt.Printf("  Auto Reduce Complexity:  %t (automatically reduce parent complexity when subtasks added)\n", cfg.AutoReduceComplexity)
		fmt.Println()

		// Show config file location
		configPath, err := configService.GetConfigPath()
		if err == nil {
			fmt.Printf("Configuration file:        %s\n", configPath)
		} else {
			fmt.Printf("Configuration file:        Unable to determine path (%s)\n", err.Error())
		}

		return nil
	}
}

// SetAction sets a configuration value
func SetAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		// Get container from CLI context
		container := shared.GetContainerFromContext(c)
		configService, err := container.GetConfigService()
		if err != nil {
			return fmt.Errorf("Configuration error: %w", err)
		}
		projectManager := container.GetProjectManager()
		key := c.String("key")
		valueStr := c.String("value")

		// Parse the string value to int
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid value '%s' for key '%s': must be a number", valueStr, key)
		}

		// Get current config
		currentConfig := configService.GetManagerConfig()
		newConfig := *currentConfig // Copy current config

		// Update the specified key
		switch key {
		case shared.ConfigKeyComplexityThreshold:
			if value < 1 || value > 10 {
				return fmt.Errorf("complexity-threshold must be between 1 and 10, got %d", value)
			}
			newConfig.ComplexityThreshold = int(value)
		case shared.ConfigKeyMaxDepth:
			if value < 1 {
				return fmt.Errorf("max-depth must be at least 1, got %d", value)
			}
			newConfig.MaxDepth = int(value)
		case shared.ConfigKeyMaxTasksPerDepth:
			if value < 1 {
				return fmt.Errorf("max-tasks-per-depth must be at least 1, got %d", value)
			}
			newConfig.MaxTasksPerDepth = int(value)
		case shared.ConfigKeyMaxDescriptionLength:
			if value < 1 {
				return fmt.Errorf("max-description-length must be at least 1, got %d", value)
			}
			newConfig.MaxDescriptionLength = int(value)
		case shared.ConfigKeyAutoReduceComplexity:
			// Convert int to bool: 0 = false, 1 = true
			if value != 0 && value != 1 {
				return fmt.Errorf("auto-reduce-complexity must be 0 (false) or 1 (true), got %d", value)
			}
			newConfig.AutoReduceComplexity = value == 1
		default:
			return fmt.Errorf("unknown configuration key: %s. Valid keys: "+shared.ConfigKeyComplexityThreshold+", "+shared.ConfigKeyMaxDepth+", "+shared.ConfigKeyMaxTasksPerDepth+", "+shared.ConfigKeyMaxDescriptionLength+", "+shared.ConfigKeyAutoReduceComplexity, key)
		}

		// Update config
		configService.SetManagerConfig(&newConfig)
		// Also update project manager to reflect the new config
		projectManager.UpdateConfig(&newConfig)

		fmt.Printf("Configuration updated: %s = %s\n", key, valueStr)
		return nil
	}
}

// ResetAction resets configuration to defaults
func ResetAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		// Get container from CLI context
		container := shared.GetContainerFromContext(c)
		configService, err := container.GetConfigService()
		if err != nil {
			return fmt.Errorf("Configuration error: %w", err)
		}
		// Reset to default config
		defaultConfig := knotconfig.DefaultConfig()
		configService.SetManagerConfig(defaultConfig)

		fmt.Println("Configuration reset to defaults:")
		fmt.Printf("  Complexity Threshold:    %d\n", defaultConfig.ComplexityThreshold)
		fmt.Printf("  Max Depth:               %d\n", defaultConfig.MaxDepth)
		fmt.Printf("  Max Tasks Per Depth:     %d\n", defaultConfig.MaxTasksPerDepth)
		fmt.Printf("  Max Description Length:  %d\n", defaultConfig.MaxDescriptionLength)
		fmt.Printf("  Auto Reduce Complexity:  %t\n", defaultConfig.AutoReduceComplexity)

		return nil
	}
}
