package config

import (
	"time"
)

// Configurations for different components
type MCPConfig struct {
	Enabled bool           `json:"enabled"`
	Address string         `json:"address"`
	Port    int            `json:"port"`
	Timeout time.Duration  `json:"timeout"`
	Database DatabaseConfig `json:"database"`
	Session  SessionConfig  `json:"session"`
	Hints    HintsConfig    `json:"hints"`
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
