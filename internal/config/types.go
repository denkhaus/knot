package config

import (
	"time"
)

// Configurations for different components
type MCPConfig struct {
	Address   string          `json:"address"`
	Port      int             `json:"port"`
	Timeout   time.Duration   `json:"timeout"`
	Database  DatabaseConfig  `json:"database"`
	Session   SessionConfig   `json:"session"`
	Hints     HintsConfig     `json:"hints"`
	Transport TransportConfig `json:"transport"`
	Tasks     TasksConfig     `json:"tasks"`
}

type DatabaseConfig struct {
	Backend  string `json:"backend"`
	Endpoint string `json:"endpoint"`
}

type SessionConfig struct {
	Timeout     time.Duration `json:"timeout"`
	MaxSessions int           `json:"max_sessions"`
}

type HintsConfig struct {
	Enabled    bool     `json:"enabled"`
	MaxHints   int      `json:"max_hints"`
	Categories []string `json:"categories"`
}

// TransportType defines the supported transport types
type TransportType string

const (
	TransportTypeStdio TransportType = "stdio"
	TransportTypeHTTP  TransportType = "http"
	TransportTypeSSE   TransportType = "sse"
)

// String returns the string representation of the transport type
func (t TransportType) String() string {
	return string(t)
}

// IsValid checks if the transport type is valid
func (t TransportType) IsValid() bool {
	switch t {
	case TransportTypeStdio, TransportTypeHTTP, TransportTypeSSE:
		return true
	default:
		return false
	}
}

// TransportConfig holds transport configuration
type TransportConfig struct {
	// Primary transport mode (stdio, http, sse, all)
	Mode TransportType `json:"mode"`

	// HTTP-specific configuration
	HTTP HTTPTransportConfig `json:"http"`

	// SSE-specific configuration
	SSE SSETransportConfig `json:"sse"`
}

// HTTPTransportConfig holds configuration for HTTP transport
// Note: mcp-go package handles HTTP implementation, so minimal config needed
type HTTPTransportConfig struct {
	// Request timeout in seconds
	RequestTimeout int `json:"request_timeout"`
}

// SSETransportConfig holds configuration for SSE transport
// Note: mcp-go package handles SSE implementation, so minimal config needed
type SSETransportConfig struct {
	// Heartbeat interval in seconds for keep-alive messages
	HeartbeatInterval int `json:"heartbeat_interval"`

	// Client timeout in seconds (for HTTP timeouts)
	ClientTimeout int `json:"client_timeout"`
}

// TasksConfig holds configuration for task management
type TasksConfig struct {
	// DefaultComplexity is the default task complexity (1-10) when not specified
	DefaultComplexity int `json:"default_complexity"`
}

// SyncConfig represents configuration for sync operations
type SyncConfig struct {
	// ServerURL is the base URL of the sync server (e.g., "http://localhost:8080")
	ServerURL string `json:"server_url"`

	// AuthToken is the optional bearer token for authentication
	AuthToken string `json:"auth_token,omitempty"`

	// PreferredFormat is "json" or "msgpack" (default: "json")
	PreferredFormat string `json:"preferred_format,omitempty"`

	// Timeout is the HTTP request timeout
	Timeout time.Duration `json:"timeout"`

	// RetryAttempts is the number of retry attempts for failed requests
	RetryAttempts int `json:"retry_attempts"`

	// RetryDelay is the base delay between retries
	RetryDelay time.Duration `json:"retry_delay"`

	// MaxRetryDelay is the maximum retry delay with exponential backoff
	MaxRetryDelay time.Duration `json:"max_retry_delay,omitempty"`

	// MaxIdleConns is the maximum number of idle connections
	MaxIdleConns int `json:"max_idle_conns,omitempty"`

	// IdleConnTimeout is the timeout for idle connections
	IdleConnTimeout time.Duration `json:"idle_conn_timeout,omitempty"`

	// ConflictStrategy determines how to resolve conflicts
	// Options: "last-writer-wins", "prefer-local", "prefer-remote", "manual"
	ConflictStrategy string `json:"conflict_strategy,omitempty"`

	// BatchSize is the number of operations to process in a batch
	BatchSize int `json:"batch_size,omitempty"`

	// Since filters data to only return changes after this timestamp
	Since *time.Time `json:"since,omitempty"`

	// UserAgent is the user agent string for HTTP requests
	UserAgent string `json:"user_agent,omitempty"`
}
