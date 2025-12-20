package config

import (
	"time"
)

// Configurations for different components
type MCPConfig struct {
	Address string         `json:"address"`
	Port    int            `json:"port"`
	Timeout time.Duration  `json:"timeout"`
	Database DatabaseConfig `json:"database"`
	Session  SessionConfig  `json:"session"`
	Hints    HintsConfig    `json:"hints"`
	Transport TransportConfig `json:"transport"`
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

	// Individual transport enable flags
	StdioEnabled bool `json:"stdio_enabled"`
	HTTPEnabled  bool `json:"http_enabled"`
	SSEEnabled   bool `json:"sse_enabled"`

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
