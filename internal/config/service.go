package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

	// In MCP mode, we don't use SQLite database, return empty string
	if s.isMCPMode {
		return ""
	}

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

	// Validate MCP configuration if in MCP mode
	if s.isMCPMode {
		if err := s.validateMCPConfig(); err != nil {
			return fmt.Errorf("failed to validate MCP config: %v", err)
		}
	}

	return nil
}

// initializeMCPConfig initializes MCP configuration from CLI flags
func (s *serviceImpl) initializeMCPConfig(c *cli.Context) {
	if c == nil {
		s.mcpConfig = s.getDefaultMCPConfig()
		return
	}

	// Get transport mode from CLI flag
	transportMode := TransportType(c.String("mcp-transport-mode"))
	if !transportMode.IsValid() {
		transportMode = TransportTypeStdio // fallback to stdio if invalid
	}

	s.mcpConfig = &MCPConfig{
		Address: c.String("mcp-endpoint"),
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
			Enabled:  true, // Hints are always enabled as core feature
			MaxHints: c.Int("mcp-hints-max"),
			// Use default categories if not provided
			Categories: []string{"next_action", "recovery", "workflow"},
		},
		Transport: TransportConfig{
			Mode:         transportMode,
			StdioEnabled: transportMode == TransportTypeStdio,
			HTTPEnabled:  transportMode == TransportTypeHTTP,
			SSEEnabled:   transportMode == TransportTypeSSE,
			HTTP: HTTPTransportConfig{
				RequestTimeout: 30, // default 30 seconds
			},
			SSE: SSETransportConfig{
				HeartbeatInterval: 30, // default 30 seconds
				ClientTimeout:     120, // default 2 minutes
			},
		},
	}
}

// getDefaultMCPConfig returns the default MCP configuration
func (s *serviceImpl) getDefaultMCPConfig() *MCPConfig {
	return &MCPConfig{
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
			Enabled:  true, // Hints are always enabled as core feature
			MaxHints: 5,
			Categories: []string{"next_action", "recovery", "workflow"},
		},
		Transport: TransportConfig{
			Mode:         TransportTypeStdio, // Default to stdio for backwards compatibility
			StdioEnabled: true,               // Enable stdio by default
			HTTPEnabled:  false,              // HTTP disabled by default
			SSEEnabled:   false,              // SSE disabled by default
			HTTP: HTTPTransportConfig{
				RequestTimeout: 30, // default 30 seconds
			},
			SSE: SSETransportConfig{
				HeartbeatInterval: 30, // default 30 seconds
				ClientTimeout:     120, // default 2 minutes
			},
		},
	}
}

// validateMCPConfig validates the MCP configuration
func (s *serviceImpl) validateMCPConfig() error {
	if s.mcpConfig == nil {
		return fmt.Errorf("MCP config is not initialized")
	}

	// Validate PostgreSQL connection string if provided
	if err := s.validatePostgresConnectionString(s.mcpConfig.Database.Endpoint); err != nil {
		return fmt.Errorf("invalid postgres-endpoint: %w", err)
	}

	// Check if port is available
	if err := s.checkPortAvailable(s.mcpConfig.Address, s.mcpConfig.Port); err != nil {
		// Provide suggestions for alternative ports
		alternatives := s.suggestAlternativePorts(s.mcpConfig.Port)
		suggestion := ""
		if len(alternatives) > 0 {
			suggestion = fmt.Sprintf(" Suggested alternatives: %s", strings.Join(alternatives, ", "))
		}
		return fmt.Errorf("%s%s", err.Error(), suggestion)
	}

	return nil
}

// validatePostgresConnectionString validates the PostgreSQL connection string format
func (s *serviceImpl) validatePostgresConnectionString(connStr string) error {
	if connStr == "" {
		return fmt.Errorf("postgres-endpoint is required. Use --postgres-endpoint flag or KNOT_POSTGRES_ENDPOINT environment variable")
	}

	// Basic validation for PostgreSQL connection string
	if !strings.HasPrefix(connStr, "postgres://") && !strings.HasPrefix(connStr, "postgresql://") {
		return fmt.Errorf("must start with 'postgres://' or 'postgresql://'")
	}

	// Check for common required parts
	parts := strings.Split(connStr, "://")
	if len(parts) != 2 {
		return fmt.Errorf("malformed connection string format")
	}

	// Extract the part after protocol and check for dbname
	remaining := parts[1]
	if strings.Contains(remaining, "/") {
		dbPart := strings.Split(remaining, "/")[1]
		// Remove query parameters if present
		if strings.Contains(dbPart, "?") {
			dbPart = strings.Split(dbPart, "?")[0]
		}
		if dbPart == "" {
			return fmt.Errorf("database name is required")
		}
	}

	return nil
}

// checkPortAvailable checks if the port is available for binding
func (s *serviceImpl) checkPortAvailable(address string, port int) error {
	addr := fmt.Sprintf("%s:%d", address, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("port %d is already in use or cannot be bound", port)
	}
	listener.Close()
	return nil
}

// suggestAlternativePorts suggests alternative ports if the requested port is unavailable
func (s *serviceImpl) suggestAlternativePorts(port int) []string {
	alternatives := make([]string, 0, 5)
	for i := 1; i <= 5; i++ {
		altPort := port + i
		addr := fmt.Sprintf("localhost:%d", altPort)
		if listener, err := net.Listen("tcp", addr); err == nil {
			alternatives = append(alternatives, strconv.Itoa(altPort))
			listener.Close()
			if len(alternatives) >= 3 {
				break
			}
		}
	}
	return alternatives
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
