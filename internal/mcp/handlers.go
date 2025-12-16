package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// ToolHandler interface for consistent tool implementation
// TODO: Implement tool handler interface for all MCP tools
type ToolHandler interface {
	Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	GetHints(result *mcp.CallToolResult) []Hint
}

// Hint represents a suggestion for next actions
// TODO: Implement hint system for agent guidance
type Hint struct {
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	NextTools   []string `json:"next_tools,omitempty"`
}

// Registry to manage all tool handlers
// TODO: Implement handler registry for tool management
type HandlerRegistry struct {
	handlers map[string]ToolHandler
}

// NewHandlerRegistry creates a new handler registry
// TODO: Implement registry initialization
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string]ToolHandler),
	}
}

// RegisterHandler registers a tool handler
// TODO: Implement handler registration with validation
func (r *HandlerRegistry) RegisterHandler(name string, handler ToolHandler) {
	r.handlers[name] = handler
}