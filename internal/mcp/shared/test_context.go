package shared

import (
	"context"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Test helper for creating contexts with mock sessions.
// This file provides test utilities for simulating MCP session contexts.

// ContextWithTestSession creates a context with a mock session for testing purposes.
// This allows tests to simulate an MCP session context without needing a full MCP server.
// Uses the mcp-go server's WithContext method to properly create a session context.
func ContextWithTestSession(sessionID uuid.UUID) context.Context {
	mockSession := &testClientSession{sessionID: sessionID.String()}
	mcpServer := server.NewMCPServer("test", "1.0.0")
	return mcpServer.WithContext(context.Background(), mockSession)
}

// testClientSession is a minimal implementation of server.ClientSession for testing.
type testClientSession struct {
	sessionID string
	actor     string
}

func (m *testClientSession) SessionID() string {
	return m.sessionID
}

func (m *testClientSession) GetActor() string {
	return m.actor
}

// Implement other required ClientSession interface methods with minimal implementations

func (m *testClientSession) Initialize()                                        {}
func (m *testClientSession) Initialized() bool                                 { return true }
func (m *testClientSession) NotificationChannel() chan<- mcp.JSONRPCNotification {
	return nil
}
func (m *testClientSession) GetClientCapabilities() mcp.ClientCapabilities {
	return mcp.ClientCapabilities{}
}
func (m *testClientSession) SetClientCapabilities(c mcp.ClientCapabilities) {}
func (m *testClientSession) GetClientInfo() mcp.Implementation {
	return mcp.Implementation{}
}
func (m *testClientSession) SetClientInfo(i mcp.Implementation) {}
func (m *testClientSession) GetLogLevel() mcp.LoggingLevel {
	return mcp.LoggingLevelInfo
}
func (m *testClientSession) SetLogLevel(l mcp.LoggingLevel) {}
func (m *testClientSession) GetSessionTools() map[string]server.ServerTool {
	return nil
}
func (m *testClientSession) SetSessionTools(tools map[string]server.ServerTool) {}
func (m *testClientSession) GetSessionResources() map[string]server.ServerResource {
	return nil
}
func (m *testClientSession) SetSessionResources(resources map[string]server.ServerResource) {}
func (m *testClientSession) GetSessionResourceTemplates() map[string]server.ServerResourceTemplate {
	return nil
}
func (m *testClientSession) SetSessionResourceTemplates(templates map[string]server.ServerResourceTemplate) {}
func (m *testClientSession) GetCreatedAt() string {
	return ""
}
func (m *testClientSession) GetLastActive() string {
	return ""
}
func (m *testClientSession) RequestElicitation(ctx context.Context, request mcp.ElicitationRequest) (*mcp.ElicitationResult, error) {
	return nil, nil
}
func (m *testClientSession) RequestSampling(ctx context.Context, request mcp.CreateMessageRequest) (*mcp.CreateMessageResult, error) {
	return nil, nil
}
func (m *testClientSession) ListRoots(ctx context.Context, request mcp.ListRootsRequest) (*mcp.ListRootsResult, error) {
	return nil, nil
}
func (m *testClientSession) Close() {}
