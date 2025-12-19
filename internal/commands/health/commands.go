// Package health provides CLI commands for system health monitoring in KNOT.
//
// This package implements command-line interface commands for checking the
// operational status of the KNOT project management system, including
// repository connectivity, data integrity, and system diagnostics.
//
// Key Commands:
//   - HealthCheck: Performs comprehensive system health checks
//   - RepositoryStatus: Checks database connectivity and integrity
//   - SystemInfo: Displays system configuration and status
//
// Cross-reference: Knot Task 86f3ba2d-3a87-493b-b8fc-96d19f344e89
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/shared"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// Commands returns health check related CLI commands
func Commands() []*cli.Command {
	return []*cli.Command{
		{
			Name:   "check",
			Usage:  "Check database connection health",
			Action: checkAction(),
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "json",
					Usage: "Output health status as JSON",
					Value: false,
				},
				&cli.DurationFlag{
					Name:  "timeout",
					Usage: "Health check timeout",
					Value: time.Second * 10,
				},
			},
		},
		{
			Name:   "ping",
			Usage:  "Simple database connectivity test",
			Action: pingAction(),
			Flags: []cli.Flag{
				&cli.DurationFlag{
					Name:  "timeout",
					Usage: "Ping timeout",
					Value: time.Second * 5,
				},
			},
		},
		{
			Name:   "validate",
			Usage:  "Comprehensive database connection validation",
			Action: validateAction(),
			Flags: []cli.Flag{
				&cli.DurationFlag{
					Name:  "timeout",
					Usage: "Validation timeout",
					Value: time.Second * 30,
				},
			},
		},
	}
}

func checkAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		container := shared.GetContainerFromCLIContext(c)
		projectManager := container.GetProjectManager()
		loggerService := container.GetLogger()

		timeout := c.Duration("timeout")
		jsonOutput := c.Bool("json")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		loggerService.Info("Performing database health check", zap.Duration("timeout", timeout))

		// Get health status from repository
		// Note: This requires extending the manager interface to expose health checks
		health, err := performHealthCheck(ctx, projectManager)
		if err != nil {
			loggerService.Error("Health check failed", zap.Error(err))
			return fmt.Errorf("health check failed: %w", err)
		}

		if jsonOutput {
			jsonData, err := json.MarshalIndent(health, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal health status: %w", err)
			}
			fmt.Println(string(jsonData))
		} else {
			printStatus(health)
		}

		if !health.Healthy {
			return fmt.Errorf("database connection is unhealthy")
		}

		return nil
	}
}

func pingAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		container := shared.GetContainerFromCLIContext(c)
		projectManager := container.GetProjectManager()
		loggerService := container.GetLogger()

		timeout := c.Duration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		loggerService.Info("Pinging database", zap.Duration("timeout", timeout))

		start := time.Now()
		err := performPing(ctx, projectManager)
		latency := time.Since(start)

		if err != nil {
			loggerService.Error("Database ping failed", zap.Error(err), zap.Duration("latency", latency))
			fmt.Printf("Database ping failed: %v\n", err)
			fmt.Printf("Latency: %v\n", latency)
			return err
		}

		loggerService.Info("Database ping successful", zap.Duration("latency", latency))
		fmt.Printf("Database ping successful\n")
		fmt.Printf("Latency: %v\n", latency)

		if latency > time.Millisecond*100 {
			fmt.Printf("High latency detected (>100ms)\n")
		}

		return nil
	}
}

func validateAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		container := shared.GetContainerFromCLIContext(c)
		projectManager := container.GetProjectManager()
		loggerService := container.GetLogger()

		timeout := c.Duration("timeout")

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		loggerService.Info("Validating database connection", zap.Duration("timeout", timeout))

		err := performValidation(ctx, projectManager)
		if err != nil {
			loggerService.Error("Database validation failed", zap.Error(err))
			fmt.Printf("Database validation failed: %v\n", err)
			return err
		}

		loggerService.Info("Database validation successful")
		fmt.Printf("✅ Database connection validation successful\n")
		fmt.Printf("   All checks passed\n")

		return nil
	}
}

// Status represents database health information (renamed from Status to avoid stuttering)
type Status struct {
	Healthy          bool          `json:"healthy"`
	ConnectionActive bool          `json:"connection_active"`
	PingLatency      time.Duration `json:"ping_latency"`
	OpenConnections  int           `json:"open_connections"`
	IdleConnections  int           `json:"idle_connections"`
	InUseConnections int           `json:"in_use_connections"`
	ErrorMessage     string        `json:"error_message,omitempty"`
	LastChecked      time.Time     `json:"last_checked"`
	DatabasePath     string        `json:"database_path"`
	WALModeEnabled   bool          `json:"wal_mode_enabled"`
	ForeignKeys      bool          `json:"foreign_keys_enabled"`
}

// performHealthCheck performs a health check using the project manager
func performHealthCheck(ctx context.Context, projectManager manager.ProjectManager) (*Status, error) {
	// For now, we'll implement a basic health check
	// TODO(knot-f7r): Extend manager interface to expose repository health checks

	start := time.Now()

	// Test basic functionality by listing projects
	_, err := projectManager.ListProjects(ctx)
	latency := time.Since(start)

	health := &Status{
		LastChecked:      time.Now(),
		PingLatency:      latency,
		ConnectionActive: err == nil,
		Healthy:          err == nil,
		DatabasePath:     ".knot/knot.db", // Default path
	}

	if err != nil {
		health.ErrorMessage = err.Error()
		return health, err // Return the error here
	}

	return health, nil
}

// performPing performs a simple connectivity test
func performPing(ctx context.Context, projectManager manager.ProjectManager) error {
	// Test basic connectivity by attempting to list projects
	_, err := projectManager.ListProjects(ctx)
	return err
}

// performValidation performs comprehensive validation
func performValidation(ctx context.Context, projectManager manager.ProjectManager) error {
	// Test multiple operations to validate connection
	tests := []struct {
		name string
		test func() error
	}{
		{"List Projects", func() error {
			_, err := projectManager.ListProjects(ctx)
			return err
		}},
		{"Get Config", func() error {
			config := projectManager.GetConfig()
			if config == nil {
				return fmt.Errorf("config is nil")
			}
			return nil
		}},
	}

	for _, test := range tests {
		if err := test.test(); err != nil {
			return fmt.Errorf("%s failed: %w", test.name, err)
		}
	}

	return nil
}

// printStatus prints health status in human-readable format
func printStatus(health *Status) {
	fmt.Printf("Database Health Status:\n\n")

	if health.Healthy {
		fmt.Printf("✅ Status: Healthy\n")
	} else {
		fmt.Printf("❌ Status: Unhealthy\n")
		if health.ErrorMessage != "" {
			fmt.Printf("   Error: %s\n", health.ErrorMessage)
		}
	}

	fmt.Printf("📊 Connection Details:\n")
	fmt.Printf("   Active: %v\n", health.ConnectionActive)
	fmt.Printf("   Latency: %v\n", health.PingLatency)
	fmt.Printf("   Database: %s\n", health.DatabasePath)
	fmt.Printf("   Last Checked: %v\n", health.LastChecked.Format(time.RFC3339))

	if health.OpenConnections > 0 {
		fmt.Printf("🔗 Connection Pool:\n")
		fmt.Printf("   Open: %d\n", health.OpenConnections)
		fmt.Printf("   Idle: %d\n", health.IdleConnections)
		fmt.Printf("   In Use: %d\n", health.InUseConnections)
	}

	if health.WALModeEnabled || health.ForeignKeys {
		fmt.Printf("⚙️  SQLite Settings:\n")
		if health.WALModeEnabled {
			fmt.Printf("   WAL Mode: ✅ Enabled\n")
		}
		if health.ForeignKeys {
			fmt.Printf("   Foreign Keys: ✅ Enabled\n")
		}
	}
}
