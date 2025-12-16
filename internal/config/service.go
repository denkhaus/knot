package config

import (
	"os"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
)

// ConfigService provides centralized configuration management for MCP mode
// Relies entirely on urfave/cli/v2 flags with built-in defaults and environment variable support
type ConfigService struct {
	mu sync.RWMutex

	// MCP mode state
	isMCPMode bool
}

var (
	// Global instance of the configuration service
	instance *ConfigService
	once     sync.Once
)

// GetConfigService returns the singleton configuration service instance
func GetConfigService() *ConfigService {
	once.Do(func() {
		instance = &ConfigService{
			isMCPMode: false,
		}
	})
	return instance
}

// InitializeFromCLIContext detects MCP mode
// This should be called early (e.g., in Before function) to detect MCP mode
func InitializeFromCLIContext(c *cli.Context) {
	service := GetConfigService()
	service.mu.Lock()
	defer service.mu.Unlock()

	// Simple detection: check if "mcp" appears in os.Args
	// This works regardless of where the flags are defined (app level or command level)
	for i, arg := range os.Args {
		if arg == "mcp" && i < len(os.Args)-1 {
			service.isMCPMode = true
			break
		}
	}
}

// IsMCPMode returns true if the application is running in MCP mode
func (s *ConfigService) IsMCPMode() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isMCPMode
}

// GetMCPConfig returns the MCP configuration by reading directly from the CLI context
// This leverages urfave/cli/v2's built-in handling of defaults, env vars, and CLI flags
func (s *ConfigService) GetMCPConfig(c *cli.Context) *MCPConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &MCPConfig{
		Address:  c.String("address"),
		Port:     c.Int("port"),
		LogLevel: c.String("log-level"),
		Database: DatabaseConfig{
			Backend:   "postgres", // Always postgres in MCP mode
			Endpoint:  c.String("postgres-endpoint"),
		},
		Session: SessionConfig{
			Timeout:     30 * time.Minute, // Default from CLI flag if needed
			MaxSessions: 100,              // Default from CLI flag if needed
		},
	}
}

// GetPostgresEndpoint returns the PostgreSQL connection string from CLI context
func (s *ConfigService) GetPostgresEndpoint(c *cli.Context) string {
	return c.String("postgres-endpoint")
}

// Configurations for different components
type MCPConfig struct {
	Address  string        `json:"address"`
	Port     int           `json:"port"`
	LogLevel string        `json:"log_level"`
	Database DatabaseConfig `json:"database"`
	Session  SessionConfig  `json:"session"`
}

type DatabaseConfig struct {
	Backend  string `json:"backend"`
	Endpoint string `json:"endpoint"`
}

type SessionConfig struct {
	Timeout     time.Duration `json:"timeout"`
	MaxSessions int           `json:"max_sessions"`
}

// Reset resets the configuration service to defaults
// Useful for testing
func (s *ConfigService) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isMCPMode = false
}