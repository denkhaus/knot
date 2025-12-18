package mcp

// Hint represents a suggestion for next actions
// TODO(knot-c8y): Implement hint system for agent guidance
type Hint struct {
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	NextTools   []string `json:"next_tools,omitempty"`
}
