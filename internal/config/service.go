package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/samber/do/v2"
	"github.com/urfave/cli/v2"
)

// ManagerConfig represents configuration for project and task management operations
type ManagerConfig struct {
	MaxTasksPerDepth     int  // Maximum tasks allowed per depth level (applies to all depths)
	ComplexityThreshold  int  // Threshold for task breakdown suggestions
	MaxDepth             int  // Maximum allowed depth
	MaxDescriptionLength int  // Maximum length for descriptions
	AutoReduceComplexity bool // Automatically reduce parent task complexity when subtasks are added
}

// ValidateConfig checks if the configuration values are valid
func (c *ManagerConfig) Validate() error {
	if c.MaxTasksPerDepth < 1 {
		return fmt.Errorf("max_tasks_per_depth must be at least 1, got %d", c.MaxTasksPerDepth)
	}
	if c.ComplexityThreshold < 1 || c.ComplexityThreshold > 10 {
		return fmt.Errorf("complexity_threshold must be between 1 and 10, got %d", c.ComplexityThreshold)
	}
	if c.MaxDepth < 1 {
		return fmt.Errorf("max_depth must be at least 1, got %d", c.MaxDepth)
	}
	if c.MaxDescriptionLength < 1 {
		return fmt.Errorf("max_description_length must be at least 1, got %d", c.MaxDescriptionLength)
	}
	return nil
}

// Service defines the configuration service interface.
// This abstracts configuration management to enable dependency injection and testing.
type Service interface {
	// Logger configuration methods
	GetLogLevel() string
	SetLogLevel(level string)

	// Database configuration methods
	GetDatabasePath() string
	SetDatabasePath(path string)

	// MCP configuration methods
	GetMCPConfig() *MCPConfig
	SetMCPConfig(config *MCPConfig)
	GetPostgresEndpoint(c *cli.Context) string
	IsMCPMode() bool
	SetMCPMode(enabled bool)

	// Templates configuration
	GetTemplatesPath() string

	// Manager configuration
	GetManagerConfig() *ManagerConfig
	SetManagerConfig(config *ManagerConfig)

	// CLI context initialization
	InitializeFromCLIContext(c *cli.Context) error
	GetConfigPath() (string, error)
}

// serviceImpl is the private implementation of the Service interface
type serviceImpl struct {
	mu            sync.RWMutex
	logLevel      string
	dbPath        string
	isMCPMode     bool
	managerConfig *ManagerConfig // Cache the manager config
	mcpConfig     *MCPConfig     // Cache the MCP config
}

// NewService creates a new configuration service instance.
func NewService(cliCtx *cli.Context) func(injector do.Injector) (Service, error) {
	return func(injector do.Injector) (Service, error) {
		svc := &serviceImpl{
			logLevel:  "info", // Default log level
			dbPath:    "",     // Will use default SQLite path
			isMCPMode: false,
		}
		if err := svc.InitializeFromCLIContext(cliCtx); err != nil {
			return nil, fmt.Errorf("failed to initialize config service: %v", err)
		}

		return svc, nil
	}
}

// GetLogLevel returns the current log level
func (s *serviceImpl) GetLogLevel() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.logLevel
}

// SetLogLevel sets the log level (can be updated from CLI flags)
func (s *serviceImpl) SetLogLevel(level string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logLevel = level
}

// GetDatabasePath returns the database path
func (s *serviceImpl) GetDatabasePath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.dbPath != "" {
		return s.dbPath
	}

	// Return default path if not set
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, ".knot", "knot.db")
}

// SetDatabasePath sets the database path (can be updated from CLI flags)
func (s *serviceImpl) SetDatabasePath(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dbPath = path
}

// GetMCPConfig returns the MCP configuration
func (s *serviceImpl) GetMCPConfig() *MCPConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.mcpConfig != nil {
		return s.mcpConfig
	}

	// Return default config if not set
	return s.getDefaultMCPConfig()
}

// GetPostgresEndpoint returns the PostgreSQL connection string from CLI context
func (s *serviceImpl) GetPostgresEndpoint(c *cli.Context) string {
	if c != nil {
		return c.String("postgres-endpoint")
	}
	return ""
}

// IsMCPMode returns true if the application is running in MCP mode
func (s *serviceImpl) IsMCPMode() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isMCPMode
}

// SetMCPConfig sets the MCP configuration
func (s *serviceImpl) SetMCPConfig(config *MCPConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mcpConfig = config
}

// SetMCPMode sets the MCP mode flag
func (s *serviceImpl) SetMCPMode(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isMCPMode = enabled
}

// GetTemplatesPath returns the templates directory path
func (s *serviceImpl) GetTemplatesPath() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, ".knot", "templates")
}

// GetManagerConfig returns the manager configuration with defaults
func (s *serviceImpl) GetManagerConfig() *ManagerConfig {
	return s.managerConfig
}

// SetManagerConfig sets the manager configuration
func (s *serviceImpl) SetManagerConfig(config *ManagerConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.managerConfig = config
}

// GetConfigPath returns the path to the knot configuration file
func (s *serviceImpl) GetConfigPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	knotDir := filepath.Join(cwd, ".knot")
	configPath := filepath.Join(knotDir, "config.json")
	return configPath, nil
}

// InitializeFromCLIContext detects MCP mode and sets configuration from CLI context
func (s *serviceImpl) InitializeFromCLIContext(c *cli.Context) error {
	// Detect MCP mode
	for i, arg := range os.Args {
		if arg == "mcp" && i < len(os.Args)-1 {
			s.SetMCPMode(true)
			break
		}
	}

	// Set log level from CLI flag
	if c.IsSet("log-level") {
		s.SetLogLevel(c.String("log-level"))
	}

	// Initialize manager config from CLI flags (defaults come from CLI)
	s.managerConfig = &ManagerConfig{
		MaxTasksPerDepth:     c.Int("max-tasks-per-depth"),
		ComplexityThreshold:  c.Int("complexity-threshold"),
		MaxDepth:             c.Int("max-depth"),
		MaxDescriptionLength: c.Int("max-description-length"),
		AutoReduceComplexity: c.Bool("auto-reduce-complexity"),
	}

	if err := s.managerConfig.Validate(); err != nil {
		return fmt.Errorf("failed to validate manager config: %v", err)
	}

	// Initialize MCP config from CLI flags if provided
	s.initializeMCPConfig(c)

	return nil
}

// initializeMCPConfig initializes MCP configuration from CLI flags
func (s *serviceImpl) initializeMCPConfig(c *cli.Context) {
	if c == nil {
		s.mcpConfig = s.getDefaultMCPConfig()
		return
	}

	s.mcpConfig = &MCPConfig{
		Enabled: c.Bool("mcp-enabled"),
		Address: c.String("mcp-address"),
		Port:    c.Int("mcp-port"),
		Database: DatabaseConfig{
			Backend:  "postgres", // Default backend
			Endpoint: c.String("postgres-endpoint"),
		},
		Session: SessionConfig{
			Timeout:     time.Duration(c.Int("mcp-timeout")) * time.Minute,
			MaxSessions: c.Int("mcp-max-sessions"),
		},
		Hints: HintsConfig{
			Enabled:  c.Bool("mcp-hints-enabled"),
			MaxHints: c.Int("mcp-hints-max"),
			// Use default categories if not provided
			Categories: []string{"next_action", "recovery", "workflow"},
		},
	}
}

// getDefaultMCPConfig returns the default MCP configuration
func (s *serviceImpl) getDefaultMCPConfig() *MCPConfig {
	return &MCPConfig{
		Enabled: false,
		Address: "localhost",
		Port:    8080,
		Timeout: 30 * time.Minute,
		Database: DatabaseConfig{
			Backend:  "postgres",
			Endpoint: "",
		},
		Session: SessionConfig{
			Timeout:     30 * time.Minute,
			MaxSessions: 100,
		},
		Hints: HintsConfig{
			Enabled:  true,
			MaxHints: 5,
			Categories: []string{"next_action", "recovery", "workflow"},
		},
	}
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *ManagerConfig {
	return &ManagerConfig{
		MaxTasksPerDepth:     50,
		ComplexityThreshold:  5,
		MaxDepth:             10,
		MaxDescriptionLength: 500,
		AutoReduceComplexity: true,
	}
}
