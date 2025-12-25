package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/denkhaus/knot/v2/internal/manager"
	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

// syncEndpoints provides private implementation of SyncHTTPHandlers interface
type syncEndpoints struct {
	handler        SyncHandler
	service        SyncService
	serializer     shared.SyncDataSerializer
	projectManager manager.ProjectManager
}

// Ensure syncEndpoints implements SyncHTTPHandlers
var _ SyncHTTPHandlers = (*syncEndpoints)(nil)

// NewSyncHTTPHandlers creates a new sync HTTP handlers implementation following DI pattern
func NewSyncHTTPHandlers(injector do.Injector) (SyncHTTPHandlers, error) {
	// Resolve dependencies using do.MustInvoke as per DI pattern
	handler := do.MustInvoke[SyncHandler](injector)
	serializer := do.MustInvoke[shared.SyncDataSerializer](injector)
	service := do.MustInvoke[SyncService](injector)
	projectManager := do.MustInvoke[manager.ProjectManager](injector)

	return &syncEndpoints{
		handler:        handler,
		service:        service,
		serializer:     serializer,
		projectManager: projectManager,
	}, nil
}

// Middleware provides common middleware for sync endpoints
func (e *syncEndpoints) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set request timeout
			ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
			defer cancel()

			// Add request ID to context
			requestID := uuid.New()
			ctx = context.WithValue(ctx, "request_id", requestID)

			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

			// Handle OPTIONS preflight
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Add request time
			ctx = context.WithValue(ctx, "request_time", time.Now())

			// Continue with modified request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FullSyncHandler returns handler for full bidirectional synchronization
// POST /api/sync/full
func (e *syncEndpoints) FullSyncHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		e.handleSync(w, r, func(ctx context.Context, request shared.SyncRequest) (*shared.SyncResponse, error) {
			return e.service.PerformFullSync(ctx, request)
		})
	}
}

// PushSyncHandler returns handler for push-only synchronization (local to remote)
// POST /api/sync/push
func (e *syncEndpoints) PushSyncHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		e.handleSync(w, r, func(ctx context.Context, request shared.SyncRequest) (*shared.SyncResponse, error) {
			return e.service.PerformPushSync(ctx, request)
		})
	}
}

// PullSyncHandler returns handler for pull-only synchronization (remote to local)
// POST /api/sync/pull
func (e *syncEndpoints) PullSyncHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		e.handleSync(w, r, func(ctx context.Context, request shared.SyncRequest) (*shared.SyncResponse, error) {
			return e.service.PerformPullSync(ctx, request)
		})
	}
}

// HealthHandler returns handler for health check
// GET /api/sync/health
func (e *syncEndpoints) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only accept GET
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check health
		if err := e.service.HealthCheck(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","error":"` + err.Error() + `"}`))
			return
		}

		// Return health status
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `","version":"1.0.0"}`))
	}
}

// handleSync provides common sync request handling logic
func (e *syncEndpoints) handleSync(w http.ResponseWriter, r *http.Request, syncFunc func(context.Context, shared.SyncRequest) (*shared.SyncResponse, error)) {
	// Only accept POST
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate and deserialize request
	request, ok := e.handler.validateAndDeserializeRequest(w, r)
	if !ok {
		return
	}

	// Perform sync operation using provided function
	response, err := syncFunc(r.Context(), *request)
	if err != nil {
		e.handler.writeErrorResponse(w, r, request.RequestID, http.StatusInternalServerError, err)
		return
	}

	// Write response
	e.handler.writeResponse(w, r, response)
}

// ListProjectsHandler returns handler for listing all projects
// GET /api/sync/projects
func (e *syncEndpoints) ListProjectsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		projects, err := e.projectManager.ListProjects(r.Context())
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to list projects: %v", err), http.StatusInternalServerError)
			return
		}

		// Convert to SyncDataSet format
		dataSet := &shared.SyncDataSet{
			Projects: make(map[uuid.UUID]*types.Project),
			Tasks:    make(map[uuid.UUID]*types.Task),
		}

		for _, p := range projects {
			dataSet.Projects[p.ID] = p
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dataSet)
	}
}

// ListTasksHandler returns handler for listing tasks for a project
// GET /api/sync/tasks?project_id=uuid
func (e *syncEndpoints) ListTasksHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		projectIDStr := r.URL.Query().Get("project_id")
		if projectIDStr == "" {
			http.Error(w, "missing project_id parameter", http.StatusBadRequest)
			return
		}

		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			http.Error(w, "invalid project_id parameter", http.StatusBadRequest)
			return
		}

		tasks, err := e.projectManager.ListTasksForProject(r.Context(), projectID)
		if err != nil {
			// Check if project not found - return 404 instead of 500
			if strings.Contains(err.Error(), "project not found") {
				http.Error(w, fmt.Sprintf("failed to list tasks: %v", err), http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("failed to list tasks: %v", err), http.StatusInternalServerError)
			return
		}

		// Convert to SyncDataSet format
		dataSet := &shared.SyncDataSet{
			Projects: make(map[uuid.UUID]*types.Project),
			Tasks:    make(map[uuid.UUID]*types.Task),
		}

		for _, t := range tasks {
			dataSet.Tasks[t.ID] = t
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dataSet)
	}
}
