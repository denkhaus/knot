package shared_test

import (
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/mcp/shared"
	"github.com/denkhaus/knot/v2/internal/session"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMCPClientSession(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		sessionID := uuid.New()
		actor := shared.ActorMCPUser

		mcpSession := shared.NewMCPClientSession(sessionID, actor)

		require.NotNil(t, mcpSession)
		assert.Equal(t, sessionID.String(), mcpSession.SessionID())
		assert.Equal(t, actor, mcpSession.GetActor())
		assert.Nil(t, mcpSession.GetProjectID())
		assert.False(t, mcpSession.Initialized())
		assert.NotNil(t, mcpSession.NotificationChannel())
	})

	t.Run("initializes timestamps", func(t *testing.T) {
		sessionID := uuid.New()
		before := time.Now()

		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		after := time.Now()

		createdAt := mcpSession.GetCreatedAt()
		lastActive := mcpSession.GetLastActive()

		assert.True(t, createdAt.After(before))
		assert.True(t, createdAt.Before(after))
		assert.True(t, lastActive.After(before))
		assert.True(t, lastActive.Before(after))
	})

	t.Run("initializes maps", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		assert.NotNil(t, mcpSession.GetSessionTools())
		assert.NotNil(t, mcpSession.GetSessionResources())
		assert.NotNil(t, mcpSession.GetSessionResourceTemplates())
		assert.Empty(t, mcpSession.GetSessionTools())
		assert.Empty(t, mcpSession.GetSessionResources())
		assert.Empty(t, mcpSession.GetSessionResourceTemplates())
	})
}

func TestNewMCPClientSessionFromInternal(t *testing.T) {
	t.Run("creates from internal session", func(t *testing.T) {
		projectID := uuid.New()
		now := time.Now()
		actor := "test-actor"

		internalSession := &session.SessionContext{
			SessionID:    uuid.New(),
			ClientID:     "client-123",
			ProjectID:    &projectID,
			Actor:        actor,
			CreatedAt:    now,
			LastActivity: now,
		}

		mcpSession := shared.NewMCPClientSessionFromInternal(internalSession)

		require.NotNil(t, mcpSession)
		assert.Equal(t, internalSession.SessionID.String(), mcpSession.SessionID())
		assert.Equal(t, actor, mcpSession.GetActor())
		assert.Equal(t, &projectID, mcpSession.GetProjectID())
		assert.True(t, mcpSession.Initialized())
		assert.Equal(t, now, mcpSession.GetCreatedAt())
		assert.Equal(t, now, mcpSession.GetLastActive())
	})

	t.Run("handles nil project ID", func(t *testing.T) {
		actor := "another-actor"
		internalSession := &session.SessionContext{
			SessionID:    uuid.New(),
			ClientID:     "client-123",
			ProjectID:    nil,
			Actor:        actor,
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}

		mcpSession := shared.NewMCPClientSessionFromInternal(internalSession)

		assert.Nil(t, mcpSession.GetProjectID())
		assert.Equal(t, actor, mcpSession.GetActor())
	})
}

func TestMCPClientSession_SessionID(t *testing.T) {
	t.Run("returns correct session ID", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		assert.Equal(t, sessionID.String(), mcpSession.SessionID())
	})

	t.Run("consistent return value", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		id1 := mcpSession.SessionID()
		id2 := mcpSession.SessionID()

		assert.Equal(t, id1, id2)
	})
}

func TestMCPClientSession_NotificationChannel(t *testing.T) {
	t.Run("returns channel", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		channel := mcpSession.NotificationChannel()

		require.NotNil(t, channel)
		// NotificationChannel returns a send-only channel (chan<-)
		assert.IsType(t, make(chan<- mcp.JSONRPCNotification), channel)
	})

	t.Run("channel is buffered", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		channel := mcpSession.NotificationChannel()

		// Should be able to send without blocking
		select {
		case channel <- mcp.JSONRPCNotification{}:
			// Success
		default:
			t.Error("channel should be buffered")
		}
	})
}

func TestMCPClientSession_Initialize(t *testing.T) {
	t.Run("initializes session", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		assert.False(t, mcpSession.Initialized())

		mcpSession.Initialize()

		assert.True(t, mcpSession.Initialized())
	})

	t.Run("idempotent", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		mcpSession.Initialize()
		mcpSession.Initialize()
		mcpSession.Initialize()

		assert.True(t, mcpSession.Initialized())
	})

	t.Run("concurrent initialization", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		done := make(chan bool)

		for i := 0; i < 10; i++ {
			go func() {
				mcpSession.Initialize()
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}

		assert.True(t, mcpSession.Initialized())
	})
}

func TestMCPClientSession_Initialized(t *testing.T) {
	t.Run("not initialized initially", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		assert.False(t, mcpSession.Initialized())
	})

	t.Run("returns true after initialize", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		mcpSession.Initialize()

		assert.True(t, mcpSession.Initialized())
	})

	t.Run("thread-safe", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		done := make(chan bool)

		// Initialize from multiple goroutines
		for i := 0; i < 10; i++ {
			go func() {
				mcpSession.Initialize()
				mcpSession.Initialized()
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}

		assert.True(t, mcpSession.Initialized())
	})
}

func TestMCPClientSession_GetActor(t *testing.T) {
	t.Run("returns actor", func(t *testing.T) {
		actor := "test-actor"
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, actor)

		assert.Equal(t, actor, mcpSession.GetActor())
	})

	t.Run("consistent", func(t *testing.T) {
		actor := shared.ActorMCPUser
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, actor)

		actor1 := mcpSession.GetActor()
		actor2 := mcpSession.GetActor()

		assert.Equal(t, actor1, actor2)
	})
}

func TestMCPClientSession_GetProjectID(t *testing.T) {
	t.Run("nil initially", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		assert.Nil(t, mcpSession.GetProjectID())
	})

	t.Run("returns set project ID", func(t *testing.T) {
		projectID := uuid.New()
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		mcpSession.SetProjectID(&projectID)

		assert.Equal(t, &projectID, mcpSession.GetProjectID())
	})
}

func TestMCPClientSession_SetProjectID(t *testing.T) {
	t.Run("sets project ID", func(t *testing.T) {
		projectID := uuid.New()
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		mcpSession.SetProjectID(&projectID)

		assert.Equal(t, &projectID, mcpSession.GetProjectID())
	})

	t.Run("updates project ID", func(t *testing.T) {
		projectID1 := uuid.New()
		projectID2 := uuid.New()
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		mcpSession.SetProjectID(&projectID1)
		assert.Equal(t, &projectID1, mcpSession.GetProjectID())

		mcpSession.SetProjectID(&projectID2)
		assert.Equal(t, &projectID2, mcpSession.GetProjectID())
	})

	t.Run("clears project ID", func(t *testing.T) {
		projectID := uuid.New()
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		mcpSession.SetProjectID(&projectID)
		assert.NotNil(t, mcpSession.GetProjectID())

		mcpSession.SetProjectID(nil)
		assert.Nil(t, mcpSession.GetProjectID())
	})

	t.Run("updates last active timestamp", func(t *testing.T) {
		projectID := uuid.New()
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		before := mcpSession.GetLastActive()
		time.Sleep(10 * time.Millisecond)

		mcpSession.SetProjectID(&projectID)

		after := mcpSession.GetLastActive()
		assert.True(t, after.After(before))
	})
}

func TestMCPClientSession_GetCreatedAt(t *testing.T) {
	t.Run("returns creation time", func(t *testing.T) {
		before := time.Now()
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)
		after := time.Now()

		createdAt := mcpSession.GetCreatedAt()

		assert.True(t, createdAt.After(before) || createdAt.Equal(before))
		assert.True(t, createdAt.Before(after) || createdAt.Equal(after))
	})

	t.Run("consistent value", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		t1 := mcpSession.GetCreatedAt()
		t2 := mcpSession.GetCreatedAt()

		assert.Equal(t, t1, t2)
	})
}

func TestMCPClientSession_GetLastActive(t *testing.T) {
	t.Run("returns last active time", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		lastActive := mcpSession.GetLastActive()

		assert.False(t, lastActive.IsZero())
	})

	t.Run("updates on activity", func(t *testing.T) {
		projectID := uuid.New()
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		time.Sleep(10 * time.Millisecond)
		before := mcpSession.GetLastActive()
		time.Sleep(10 * time.Millisecond)

		mcpSession.SetProjectID(&projectID)

		after := mcpSession.GetLastActive()
		assert.True(t, after.After(before))
	})
}

func TestMCPClientSession_SetLogLevel(t *testing.T) {
	t.Run("sets log level", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		mcpSession.SetLogLevel(mcp.LoggingLevelDebug)

		assert.Equal(t, mcp.LoggingLevelDebug, mcpSession.GetLogLevel())
	})

	t.Run("updates log level", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		mcpSession.SetLogLevel(mcp.LoggingLevelInfo)
		assert.Equal(t, mcp.LoggingLevelInfo, mcpSession.GetLogLevel())

		mcpSession.SetLogLevel(mcp.LoggingLevelWarning)
		assert.Equal(t, mcp.LoggingLevelWarning, mcpSession.GetLogLevel())
	})
}

func TestMCPClientSession_GetLogLevel(t *testing.T) {
	t.Run("returns default info level", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		assert.Equal(t, mcp.LoggingLevelInfo, mcpSession.GetLogLevel())
	})

	t.Run("returns set level", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		mcpSession.SetLogLevel(mcp.LoggingLevelError)

		assert.Equal(t, mcp.LoggingLevelError, mcpSession.GetLogLevel())
	})
}

func TestMCPClientSession_GetSessionTools(t *testing.T) {
	t.Run("returns empty map initially", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		tools := mcpSession.GetSessionTools()

		assert.NotNil(t, tools)
		assert.Empty(t, tools)
	})

	t.Run("returns copy of tools", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		tools := mcpSession.GetSessionTools()
		tools["test"] = server.ServerTool{}

		// Original should be unchanged
		tools2 := mcpSession.GetSessionTools()
		assert.Empty(t, tools2)
	})
}

func TestMCPClientSession_SetSessionTools(t *testing.T) {
	t.Run("sets tools", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		tools := map[string]server.ServerTool{
			"tool1": {},
			"tool2": {},
		}

		mcpSession.SetSessionTools(tools)

		retrieved := mcpSession.GetSessionTools()
		assert.Len(t, retrieved, 2)
		assert.Contains(t, retrieved, "tool1")
		assert.Contains(t, retrieved, "tool2")
	})

	t.Run("creates copy", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		tools := map[string]server.ServerTool{
			"tool1": {},
		}

		mcpSession.SetSessionTools(tools)

		// Modify original
		tools["tool2"] = server.ServerTool{}

		// Session should have original set
		retrieved := mcpSession.GetSessionTools()
		assert.Len(t, retrieved, 1)
		assert.Contains(t, retrieved, "tool1")
		assert.NotContains(t, retrieved, "tool2")
	})
}

func TestMCPClientSession_GetSessionResources(t *testing.T) {
	t.Run("returns empty map initially", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		resources := mcpSession.GetSessionResources()

		assert.NotNil(t, resources)
		assert.Empty(t, resources)
	})
}

func TestMCPClientSession_SetSessionResources(t *testing.T) {
	t.Run("sets resources", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		resources := map[string]server.ServerResource{
			"res1": {},
			"res2": {},
		}

		mcpSession.SetSessionResources(resources)

		retrieved := mcpSession.GetSessionResources()
		assert.Len(t, retrieved, 2)
	})
}

func TestMCPClientSession_GetSessionResourceTemplates(t *testing.T) {
	t.Run("returns empty map initially", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		templates := mcpSession.GetSessionResourceTemplates()

		assert.NotNil(t, templates)
		assert.Empty(t, templates)
	})
}

func TestMCPClientSession_SetSessionResourceTemplates(t *testing.T) {
	t.Run("sets templates", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		templates := map[string]server.ServerResourceTemplate{
			"tmpl1": {},
			"tmpl2": {},
		}

		mcpSession.SetSessionResourceTemplates(templates)

		retrieved := mcpSession.GetSessionResourceTemplates()
		assert.Len(t, retrieved, 2)
	})
}

func TestMCPClientSession_GetClientInfo(t *testing.T) {
	t.Run("returns empty info initially", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		info := mcpSession.GetClientInfo()

		assert.Equal(t, mcp.Implementation{}, info)
	})

	t.Run("returns set info", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		clientInfo := mcp.Implementation{
			Name:    "test-client",
			Version: "1.0.0",
		}

		mcpSession.SetClientInfo(clientInfo)

		info := mcpSession.GetClientInfo()
		assert.Equal(t, "test-client", info.Name)
		assert.Equal(t, "1.0.0", info.Version)
	})
}

func TestMCPClientSession_GetClientCapabilities(t *testing.T) {
	t.Run("returns empty capabilities initially", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		caps := mcpSession.GetClientCapabilities()

		assert.Equal(t, mcp.ClientCapabilities{}, caps)
	})

	t.Run("returns set capabilities", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		capabilities := mcp.ClientCapabilities{}

		mcpSession.SetClientCapabilities(capabilities)

		caps := mcpSession.GetClientCapabilities()
		assert.Equal(t, capabilities, caps)
	})
}

func TestMCPClientSession_Close(t *testing.T) {
	t.Run("closes notification channel", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		channel := mcpSession.NotificationChannel()

		mcpSession.Close()

		// Can't receive from send-only channel, but we verified it exists
		assert.NotNil(t, channel)
	})
}

func TestMCPClientSession_ClientSessionInterface(t *testing.T) {
	t.Run("implements ClientSession interface", func(t *testing.T) {
		var _ server.ClientSession = (*shared.MCPClientSession)(nil)
	})
}

func TestMCPClientSession_AtomicOperations(t *testing.T) {
	t.Run("initialized state is consistent", func(t *testing.T) {
		sessionID := uuid.New()
		mcpSession := shared.NewMCPClientSession(sessionID, shared.ActorMCPUser)

		// Initially not initialized
		assert.False(t, mcpSession.Initialized())

		mcpSession.Initialize()
		assert.True(t, mcpSession.Initialized())
	})
}
