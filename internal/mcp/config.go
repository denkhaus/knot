package mcp

import "time"

// Config holds the MCP server configuration
// TODO: Implement full configuration with validation
type Config struct {
	// Server settings
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`

	// Session settings
	Timeout     time.Duration `yaml:"timeout"`
	MaxSessions int           `yaml:"max_sessions"`

	// Hint settings
	Hints HintsConfig `yaml:"hints"`
}

// HintsConfig holds hint system configuration
// TODO: Implement hint system configuration
type HintsConfig struct {
	Enabled  bool     `yaml:"enabled"`
	MaxHints int      `yaml:"max_hints"`
	Categories []string `yaml:"categories"`
}

// DefaultConfig returns the default MCP configuration
// TODO: Define proper default values
func DefaultConfig() *Config {
	return &Config{
		Enabled:     false,
		Address:     "localhost",
		Port:        8080,
		Timeout:     30 * time.Minute,
		MaxSessions: 100,
		Hints: HintsConfig{
			Enabled:  true,
			MaxHints: 5,
			Categories: []string{"next_action", "recovery", "workflow"},
		},
	}
}