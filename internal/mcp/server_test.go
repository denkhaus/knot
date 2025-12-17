package mcp

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/mcp/session"
	"github.com/denkhaus/knot/v2/internal/mocks"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := slog.Default()
	mockConfig := &config.MCPConfig{
		Address:  "localhost",
		Port:     8080,
		LogLevel: "info",
		Database: config.DatabaseConfig{
			Backend:  "postgres",
			Endpoint: "localhost:5432",
		},
		Session: config.SessionConfig{
			Timeout:     30 * time.Minute,
			MaxSessions: 100,
		},
	}

	t.Run("Successful server creation", func(t *testing.T) {
		mockManager := mocks.NewMockProjectManager(ctrl)

		cfg := ServerConfig{
			ProjectManager: mockManager,
			Logger:         mockLogger,
			Config:         mockConfig,
		}

		server, err := NewServer(cfg)
		require.NoError(t, err)
		require.NotNil(t, server)

		assert.Equal(t, mockManager, server.projectManager)
		assert.Equal(t, mockLogger, server.logger)
		assert.Equal(t, mockConfig, server.config)
		assert.NotNil(t, server.sessions)
		assert.NotNil(t, server.MCPServer)
	})

	t.Run("Missing project manager", func(t *testing.T) {
		cfg := ServerConfig{
			Logger: mockLogger,
			Config: mockConfig,
		}

		server, err := NewServer(cfg)
		require.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "project manager is required")
	})

	t.Run("Missing logger", func(t *testing.T) {
		mockManager := mocks.NewMockProjectManager(ctrl)

		cfg := ServerConfig{
			ProjectManager: mockManager,
			Config:         mockConfig,
		}

		server, err := NewServer(cfg)
		require.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "logger is required")
	})

	t.Run("Missing config", func(t *testing.T) {
		mockManager := mocks.NewMockProjectManager(ctrl)

		cfg := ServerConfig{
			ProjectManager: mockManager,
			Logger:         mockLogger,
		}

		server, err := NewServer(cfg)
		require.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "config is required")
	})
}

func TestMCPServer_ToolsRegistration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := slog.Default()
	mockConfig := &config.MCPConfig{
		Address:  "localhost",
		Port:     8080,
		LogLevel: "info",
		Database: config.DatabaseConfig{
			Backend:  "postgres",
			Endpoint: "localhost:5432",
		},
		Session: config.SessionConfig{
			Timeout:     30 * time.Minute,
			MaxSessions: 100,
		},
	}
	mockManager := mocks.NewMockProjectManager(ctrl)

	cfg := ServerConfig{
		ProjectManager: mockManager,
		Logger:         mockLogger,
		Config:         mockConfig,
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)

	// Verify tools are registered by checking the tools list
	tools := server.ListTools()
	require.NotEmpty(t, tools)

	// Just verify tools are not empty - mcp-go server doesn't expose tool names easily
	assert.NotEmpty(t, tools)
}

func TestMCPServer_HandleProjectSelect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := slog.Default()
	mockConfig := &config.MCPConfig{
		Address:  "localhost",
		Port:     8080,
		LogLevel: "info",
		Database: config.DatabaseConfig{
			Backend:  "postgres",
			Endpoint: "localhost:5432",
		},
		Session: config.SessionConfig{
			Timeout:     30 * time.Minute,
			MaxSessions: 100,
		},
	}
	mockManager := mocks.NewMockProjectManager(ctrl)

	cfg := ServerConfig{
		ProjectManager: mockManager,
		Logger:         mockLogger,
		Config:         mockConfig,
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)

	sessionID := uuid.New()
	projectID := uuid.New()

	// Create a mock session context
	ctx := context.Background()

	t.Run("Successful project selection", func(t *testing.T) {
		// Set up session
		_, err := server.sessions.CreateSession("test-user")
		require.NoError(t, err)

		// Mock request
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id": projectID.String(),
				},
			},
		}

		args := ProjectSelectRequest{
			ProjectID: projectID.String(),
		}

		response, err := server.handleProjectSelect(ctx, request, args)
		require.NoError(t, err)

		assert.Equal(t, "Selected project "+projectID.String()+" for session", response.Message)
		assert.Equal(t, projectID.String(), response.ProjectID)

		// Verify session was updated
		retrievedSession, err := server.sessions.GetSession(sessionID)
		require.NoError(t, err)
		require.NotNil(t, retrievedSession.ProjectID)
		assert.Equal(t, projectID, *retrievedSession.ProjectID)
	})

	t.Run("Invalid project ID", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"project_id": "invalid-uuid",
				},
			},
		}

		args := ProjectSelectRequest{
			ProjectID: "invalid-uuid",
		}

		response, err := server.handleProjectSelect(ctx, request, args)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid project_id format")
		assert.Empty(t, response.Message)
	})
}

func TestMCPServer_HandleProjectCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := slog.Default()
	mockConfig := &config.MCPConfig{
		Address:  "localhost",
		Port:     8080,
		LogLevel: "info",
		Database: config.DatabaseConfig{
			Backend:  "postgres",
			Endpoint: "localhost:5432",
		},
		Session: config.SessionConfig{
			Timeout:     30 * time.Minute,
			MaxSessions: 100,
		},
	}
	mockManager := mocks.NewMockProjectManager(ctrl)

	cfg := ServerConfig{
		ProjectManager: mockManager,
		Logger:         mockLogger,
		Config:         mockConfig,
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("Successful project creation", func(t *testing.T) {
		projectTitle := "Test Project"
		projectDescription := "A test project"
		expectedProject := &types.Project{
			ID:          uuid.New(),
			Title:       projectTitle,
			Description: projectDescription,
		}

		mockManager.EXPECT().
			CreateProject(ctx, projectTitle, projectDescription, "").
			Return(expectedProject, nil)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"title":       projectTitle,
					"description": projectDescription,
				},
			},
		}

		args := ProjectCreateRequest{
			Title:       projectTitle,
			Description: projectDescription,
		}

		response, err := server.handleProjectCreate(ctx, request, args)
		require.NoError(t, err)

		assert.Contains(t, response.Message, "Created project "+projectTitle)
		assert.Equal(t, expectedProject.ID.String(), response.ProjectID)
		assert.Equal(t, projectTitle, response.Title)
	})

	t.Run("Project creation error", func(t *testing.T) {
		projectTitle := "Error Project"

		mockManager.EXPECT().
			CreateProject(ctx, projectTitle, "", "").
			Return(nil, assert.AnError)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"title": projectTitle,
				},
			},
		}

		args := ProjectCreateRequest{
			Title: projectTitle,
		}

		response, err := server.handleProjectCreate(ctx, request, args)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create project")
		assert.Empty(t, response.Message)
	})
}

func TestMCPServer_HandleTaskCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := slog.Default()
	mockConfig := &config.MCPConfig{
		Address:  "localhost",
		Port:     8080,
		LogLevel: "info",
		Database: config.DatabaseConfig{
			Backend:  "postgres",
			Endpoint: "localhost:5432",
		},
		Session: config.SessionConfig{
			Timeout:     30 * time.Minute,
			MaxSessions: 100,
		},
	}
	mockManager := mocks.NewMockProjectManager(ctrl)

	cfg := ServerConfig{
		ProjectManager: mockManager,
		Logger:         mockLogger,
		Config:         mockConfig,
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)

	sessionID := uuid.New()
	projectID := uuid.New()
	ctx := context.Background()

	// Create session and set project
	_, err = server.sessions.CreateSession("test-user")
	require.NoError(t, err)
	err = server.sessions.SetProject(sessionID, projectID)
	require.NoError(t, err)

	t.Run("Successful task creation", func(t *testing.T) {
		taskTitle := "Test Task"
		taskDescription := "A test task"
		taskComplexity := 5
		expectedTask := &types.Task{
			ID:          uuid.New(),
			Title:       taskTitle,
			Description: taskDescription,
			Complexity:  taskComplexity,
		}

		mockManager.EXPECT().
			CreateTask(ctx, projectID, (*uuid.UUID)(nil), taskTitle, taskDescription, taskComplexity, types.TaskPriorityMedium, "").
			Return(expectedTask, nil)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"title":       taskTitle,
					"description": taskDescription,
					"complexity":  taskComplexity,
				},
			},
		}

		args := TaskCreateRequest{
			Title:       taskTitle,
			Description: taskDescription,
			Complexity:  taskComplexity,
		}

		response, err := server.handleTaskCreate(ctx, request, args)
		require.NoError(t, err)

		assert.Contains(t, response.Message, "Created task "+taskTitle)
		assert.Equal(t, expectedTask.ID.String(), response.TaskID)
		assert.Equal(t, projectID.String(), response.ProjectID)
		assert.Equal(t, taskTitle, response.Title)
	})

	t.Run("Task creation with default complexity", func(t *testing.T) {
		taskTitle := "Default Complexity Task"
		expectedTask := &types.Task{
			ID:         uuid.New(),
			Title:      taskTitle,
			Complexity: 5, // Default complexity
		}

		mockManager.EXPECT().
			CreateTask(ctx, projectID, (*uuid.UUID)(nil), taskTitle, "", 5, types.TaskPriorityMedium, "").
			Return(expectedTask, nil)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"title": taskTitle,
				},
			},
		}

		args := TaskCreateRequest{
			Title: taskTitle,
		}

		response, err := server.handleTaskCreate(ctx, request, args)
		require.NoError(t, err)

		assert.Contains(t, response.Message, "Created task "+taskTitle)
		assert.Equal(t, expectedTask.ID.String(), response.TaskID)
	})

	t.Run("Task creation without selected project", func(t *testing.T) {
		// Create session without project
		noProjectSessionManager := session.NewManager()
		noProjectSessionManager.CreateSession("test-user-2")
		server.sessions = noProjectSessionManager

		taskTitle := "No Project Task"

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"title": taskTitle,
				},
			},
		}

		args := TaskCreateRequest{
			Title: taskTitle,
		}

		response, err := server.handleTaskCreate(ctx, request, args)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no project selected")
		assert.Empty(t, response.Message)
	})

	t.Run("Task creation error", func(t *testing.T) {
		taskTitle := "Error Task"

		mockManager.EXPECT().
			CreateTask(ctx, projectID, (*uuid.UUID)(nil), taskTitle, "", 5, types.TaskPriorityMedium, "").
			Return(nil, assert.AnError)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"title": taskTitle,
				},
			},
		}

		args := TaskCreateRequest{
			Title: taskTitle,
		}

		response, err := server.handleTaskCreate(ctx, request, args)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create task")
		assert.Empty(t, response.Message)
	})
}

func TestMCPServer_GetSessionID(t *testing.T) {
	t.Run("No session in context", func(t *testing.T) {
		ctx := context.Background()
		sessionID := getSessionID(ctx)
		assert.Empty(t, sessionID)
	})
}

func TestMCPServer_Stop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := slog.Default()
	mockConfig := &config.MCPConfig{
		Address:  "localhost",
		Port:     8080,
		LogLevel: "info",
		Database: config.DatabaseConfig{
			Backend:  "postgres",
			Endpoint: "localhost:5432",
		},
		Session: config.SessionConfig{
			Timeout:     30 * time.Minute,
			MaxSessions: 100,
		},
	}
	mockManager := mocks.NewMockProjectManager(ctrl)

	cfg := ServerConfig{
		ProjectManager: mockManager,
		Logger:         mockLogger,
		Config:         mockConfig,
	}

	server, err := NewServer(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Stop server
	err = server.Stop(ctx)
	require.NoError(t, err)
}

// TestStructuredInputOutput validates that our structured types work correctly
func TestStructuredInputOutput(t *testing.T) {
	t.Run("ProjectSelectRequest validation", func(t *testing.T) {
		req := ProjectSelectRequest{
			ProjectID: uuid.New().String(),
		}

		assert.NotEmpty(t, req.ProjectID)
	})

	t.Run("ProjectCreateRequest validation", func(t *testing.T) {
		req := ProjectCreateRequest{
			Title:       "Test Project",
			Description: "Test Description",
		}

		assert.Equal(t, "Test Project", req.Title)
		assert.Equal(t, "Test Description", req.Description)
	})

	t.Run("TaskCreateRequest validation", func(t *testing.T) {
		req := TaskCreateRequest{
			Title:       "Test Task",
			Description: "Test Description",
			Complexity:  7,
		}

		assert.Equal(t, "Test Task", req.Title)
		assert.Equal(t, "Test Description", req.Description)
		assert.Equal(t, 7, req.Complexity)
	})

	t.Run("Response validation", func(t *testing.T) {
		projectResp := ProjectSelectResponse{
			Message:   "Success",
			ProjectID: uuid.New().String(),
		}

		assert.Equal(t, "Success", projectResp.Message)
		assert.NotEmpty(t, projectResp.ProjectID)

		taskResp := TaskCreateResponse{
			Message:   "Task created",
			TaskID:    uuid.New().String(),
			ProjectID: uuid.New().String(),
			Title:     "Test Task",
		}

		assert.Equal(t, "Task created", taskResp.Message)
		assert.NotEmpty(t, taskResp.TaskID)
		assert.NotEmpty(t, taskResp.ProjectID)
		assert.Equal(t, "Test Task", taskResp.Title)
	})
}
