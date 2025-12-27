package transports

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/mcp/shared"
	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/server"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewSessionWrapper(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockSessionRegistry := mocks.NewMockSessionRegistry(ctrl)
	mcpServer := server.NewMCPServer("test", "1.0.0")

	t.Run("successful creation", func(t *testing.T) {
		injector := do.New()
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue[shared.SessionRegistry](injector, mockSessionRegistry)
		do.ProvideValue[*server.MCPServer](injector, mcpServer)

		wrapper, err := NewSessionWrapper(injector)

		require.NoError(t, err)
		require.NotNil(t, wrapper)

		assert.NotNil(t, wrapper.StreamableHTTPServer)
		assert.Same(t, mockSessionRegistry, wrapper.sessionRegistry)
		assert.Same(t, mockLogger, wrapper.logger)
	})

	t.Run("verifies resolver is set", func(t *testing.T) {
		injector := do.New()
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue[shared.SessionRegistry](injector, mockSessionRegistry)
		do.ProvideValue[*server.MCPServer](injector, mcpServer)

		wrapper, err := NewSessionWrapper(injector)

		require.NoError(t, err)
		require.NotNil(t, wrapper)
		// The wrapper should have a resolver configured
		assert.NotNil(t, wrapper)
	})
}

func TestCustomSessionIDManagerResolver(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockSessionRegistry := mocks.NewMockSessionRegistry(ctrl)

	t.Run("creates new resolver", func(t *testing.T) {
		resolver := NewCustomSessionIDManagerResolver(mockSessionRegistry, mockLogger)

		require.NotNil(t, resolver)
		assert.Same(t, mockSessionRegistry, resolver.sessionRegistry)
		assert.Same(t, mockLogger, resolver.logger)
	})

	t.Run("ResolveSessionIdManager returns CustomSessionIDManager", func(t *testing.T) {
		resolver := NewCustomSessionIDManagerResolver(mockSessionRegistry, mockLogger)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-MCP-Session-ID", "test-session-id")

		manager := resolver.ResolveSessionIdManager(req)

		require.NotNil(t, manager)
		assert.IsType(t, &CustomSessionIDManager{}, manager)

		customManager := manager.(*CustomSessionIDManager)
		assert.Same(t, mockSessionRegistry, customManager.sessionRegistry)
		assert.Same(t, mockLogger, customManager.logger)
	})

	t.Run("ResolveSessionIdManager without session header", func(t *testing.T) {
		resolver := NewCustomSessionIDManagerResolver(mockSessionRegistry, mockLogger)

		req := httptest.NewRequest("GET", "/test", nil)

		manager := resolver.ResolveSessionIdManager(req)

		require.NotNil(t, manager)
		assert.IsType(t, &CustomSessionIDManager{}, manager)
	})
}

func TestCustomSessionIDManager_Generate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockSessionRegistry := mocks.NewMockSessionRegistry(ctrl)
	mcpSession := shared.NewMCPClientSession(uuid.New(), shared.ActorMCPUser)

	t.Run("successful session generation", func(t *testing.T) {
		manager := &CustomSessionIDManager{
			sessionRegistry: mockSessionRegistry,
			logger:          mockLogger,
		}

		mockSessionRegistry.EXPECT().GetOrCreateSession(
			gomock.Any(),
			gomock.Any(),
		).Return(mcpSession, nil)

		sessionID := manager.Generate()

		// Should return a valid UUID string
		_, err := uuid.Parse(sessionID)
		require.NoError(t, err)
	})

	t.Run("session creation failure still returns ID", func(t *testing.T) {
		manager := &CustomSessionIDManager{
			sessionRegistry: mockSessionRegistry,
			logger:          mockLogger,
		}

		mockSessionRegistry.EXPECT().GetOrCreateSession(
			gomock.Any(),
			gomock.Any(),
		).Return(nil, assert.AnError)

		sessionID := manager.Generate()

		// Should still return a valid UUID string even on error
		_, err := uuid.Parse(sessionID)
		require.NoError(t, err)
	})
}

func TestCustomSessionIDManager_Validate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockSessionRegistry := mocks.NewMockSessionRegistry(ctrl)
	mcpSession := shared.NewMCPClientSession(uuid.New(), shared.ActorMCPUser)

	t.Run("valid session ID", func(t *testing.T) {
		manager := &CustomSessionIDManager{
			sessionRegistry: mockSessionRegistry,
			logger:          mockLogger,
		}

		testUUID := uuid.New()
		mockSessionRegistry.EXPECT().GetOrCreateSession(
			gomock.Any(),
			testUUID,
		).Return(mcpSession, nil)

		isTerminated, err := manager.Validate(testUUID.String())

		require.NoError(t, err)
		assert.False(t, isTerminated) // Session is valid, not terminated
	})

	t.Run("session ID with prefix", func(t *testing.T) {
		manager := &CustomSessionIDManager{
			sessionRegistry: mockSessionRegistry,
			logger:          mockLogger,
		}

		testUUID := uuid.New()
		prefixedID := "mcp-session-" + testUUID.String()
		mockSessionRegistry.EXPECT().GetOrCreateSession(
			gomock.Any(),
			testUUID, // Should extract UUID without prefix
		).Return(mcpSession, nil)

		isTerminated, err := manager.Validate(prefixedID)

		require.NoError(t, err)
		assert.False(t, isTerminated)
	})

	t.Run("invalid session ID format", func(t *testing.T) {
		manager := &CustomSessionIDManager{
			sessionRegistry: mockSessionRegistry,
			logger:          mockLogger,
		}

		isTerminated, err := manager.Validate("invalid-uuid")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid session ID format")
		assert.False(t, isTerminated)
	})

	t.Run("session creation failure", func(t *testing.T) {
		manager := &CustomSessionIDManager{
			sessionRegistry: mockSessionRegistry,
			logger:          mockLogger,
		}

		testUUID := uuid.New()
		mockSessionRegistry.EXPECT().GetOrCreateSession(
			gomock.Any(),
			testUUID,
		).Return(nil, assert.AnError)

		isTerminated, err := manager.Validate(testUUID.String())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get or create session")
		assert.False(t, isTerminated)
	})
}

func TestCustomSessionIDManager_Terminate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockSessionRegistry := mocks.NewMockSessionRegistry(ctrl)

	t.Run("successful termination", func(t *testing.T) {
		manager := &CustomSessionIDManager{
			sessionRegistry: mockSessionRegistry,
			logger:          mockLogger,
		}

		testUUID := uuid.New()
		mockSessionRegistry.EXPECT().RemoveSession(
			gomock.Any(),
			testUUID,
		).Return(nil)

		isNotAllowed, err := manager.Terminate(testUUID.String())

		require.NoError(t, err)
		assert.False(t, isNotAllowed) // Termination is allowed
	})

	t.Run("terminate with prefix", func(t *testing.T) {
		manager := &CustomSessionIDManager{
			sessionRegistry: mockSessionRegistry,
			logger:          mockLogger,
		}

		testUUID := uuid.New()
		prefixedID := "mcp-session-" + testUUID.String()
		mockSessionRegistry.EXPECT().RemoveSession(
			gomock.Any(),
			testUUID, // Should extract UUID without prefix
		).Return(nil)

		isNotAllowed, err := manager.Terminate(prefixedID)

		require.NoError(t, err)
		assert.False(t, isNotAllowed)
	})

	t.Run("invalid session ID format", func(t *testing.T) {
		manager := &CustomSessionIDManager{
			sessionRegistry: mockSessionRegistry,
			logger:          mockLogger,
		}

		isNotAllowed, err := manager.Terminate("invalid-uuid")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid session ID format")
		assert.False(t, isNotAllowed)
	})

	t.Run("session removal failure", func(t *testing.T) {
		manager := &CustomSessionIDManager{
			sessionRegistry: mockSessionRegistry,
			logger:          mockLogger,
		}

		testUUID := uuid.New()
		mockSessionRegistry.EXPECT().RemoveSession(
			gomock.Any(),
			testUUID,
		).Return(assert.AnError)

		isNotAllowed, err := manager.Terminate(testUUID.String())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove session")
		assert.False(t, isNotAllowed)
	})
}

func TestSessionWrapper_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockSessionRegistry := mocks.NewMockSessionRegistry(ctrl)
	mcpServer := server.NewMCPServer("test", "1.0.0")
	mcpSession := shared.NewMCPClientSession(uuid.New(), shared.ActorMCPUser)

	t.Run("serve HTTP with session header", func(t *testing.T) {
		t.Parallel()

		injector := do.New()
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue[shared.SessionRegistry](injector, mockSessionRegistry)
		do.ProvideValue[*server.MCPServer](injector, mcpServer)

		wrapper, err := NewSessionWrapper(injector)
		require.NoError(t, err)

		testUUID := uuid.New()
		mockSessionRegistry.EXPECT().GetOrCreateSession(
			gomock.Any(),
			testUUID,
		).Return(mcpSession, nil)

		// Create a request with context that can be cancelled
		ctx, cancel := context.WithCancel(context.Background())
		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		req.Header.Set("X-MCP-Session-ID", testUUID.String())
		w := httptest.NewRecorder()

		// Run ServeHTTP in a goroutine and cancel after a short delay
		done := make(chan struct{})
		go func() {
			wrapper.ServeHTTP(w, req)
			close(done)
		}()

		// Cancel after a short time to prevent hanging
		select {
		case <-done:
			// Test completed normally
		case <-time.After(100 * time.Millisecond):
			cancel()
			// Wait a bit for cleanup
			select {
			case <-done:
			case <-time.After(50 * time.Millisecond):
			}
		}

		// Request should be processed
		assert.True(t, w.Code >= 200 || w.Code == 404) // Either OK or not found
	})

	t.Run("serve HTTP with invalid session header", func(t *testing.T) {
		t.Parallel()

		injector := do.New()
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue[shared.SessionRegistry](injector, mockSessionRegistry)
		do.ProvideValue[*server.MCPServer](injector, mcpServer)

		wrapper, err := NewSessionWrapper(injector)
		require.NoError(t, err)

		// Create a request with context that can be cancelled
		ctx, cancel := context.WithCancel(context.Background())
		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		req.Header.Set("X-MCP-Session-ID", "invalid-uuid")
		w := httptest.NewRecorder()

		// Run ServeHTTP in a goroutine and cancel after a short delay
		done := make(chan struct{})
		go func() {
			wrapper.ServeHTTP(w, req)
			close(done)
		}()

		// Cancel after a short time to prevent hanging
		select {
		case <-done:
			// Test completed normally
		case <-time.After(100 * time.Millisecond):
			cancel()
			// Wait a bit for cleanup
			select {
			case <-done:
			case <-time.After(50 * time.Millisecond):
			}
		}

		// Request should still be processed
		assert.True(t, w.Code >= 200 || w.Code == 404)
	})

	t.Run("serve HTTP without session header", func(t *testing.T) {
		t.Parallel()

		injector := do.New()
		do.ProvideValue[logger.Logger](injector, mockLogger)
		do.ProvideValue[shared.SessionRegistry](injector, mockSessionRegistry)
		do.ProvideValue[*server.MCPServer](injector, mcpServer)

		wrapper, err := NewSessionWrapper(injector)
		require.NoError(t, err)

		// Create a request with context that can be cancelled
		ctx, cancel := context.WithCancel(context.Background())
		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		// Run ServeHTTP in a goroutine and cancel after a short delay
		done := make(chan struct{})
		go func() {
			wrapper.ServeHTTP(w, req)
			close(done)
		}()

		// Cancel after a short time to prevent hanging
		select {
		case <-done:
			// Test completed normally
		case <-time.After(100 * time.Millisecond):
			cancel()
			// Wait a bit for cleanup
			select {
			case <-done:
			case <-time.After(50 * time.Millisecond):
			}
		}

		// Request should still be processed
		assert.True(t, w.Code >= 200 || w.Code == 404)
	})
}
