package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
)

// SyncEndpoints provides the sync HTTP endpoints
type SyncEndpoints interface {
	// Middleware for request timeout and logging
	Middleware() func(http.Handler) http.Handler

	// FullSync handles bidirectional synchronization requests
	FullSync(w http.ResponseWriter, r *http.Request)

	// PushSync handles push-only synchronization (local to remote)
	PushSync(w http.ResponseWriter, r *http.Request)

	// PullSync handles pull-only synchronization (remote to local)
	PullSync(w http.ResponseWriter, r *http.Request)

	// Health handles health check for sync service
	Health(w http.ResponseWriter, r *http.Request)
}

// syncEndpointsImpl provides private implementation of SyncEndpoints
type syncEndpointsImpl struct {
	handler     SyncHandler
	syncService SyncService
}

// Ensure syncEndpointsImpl implements SyncEndpoints
var _ SyncEndpoints = (*syncEndpointsImpl)(nil)

// NewSyncEndpoints creates new sync endpoints following DI pattern
func NewSyncEndpoints(injector do.Injector) (SyncEndpoints, error) {
	// Resolve dependencies using do.MustInvoke as per DI pattern
	handler := do.MustInvoke[SyncHandler](injector)
	syncService := do.MustInvoke[SyncService](injector)

	return &syncEndpointsImpl{
		handler:     handler,
		syncService: syncService,
	}, nil
}

// Middleware for request timeout and logging
func (e *syncEndpointsImpl) Middleware() func(http.Handler) http.Handler {
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
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

			// Handle OPTIONS preflight
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Continue with modified request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FullSync handles bidirectional synchronization requests
// POST /api/sync/full
func (e *syncEndpointsImpl) FullSync(w http.ResponseWriter, r *http.Request) {
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

	// Check direction
	if request.Direction == shared.SyncBidirectional {
		// For bidirectional sync, we need to perform full sync logic
		response := e.performFullSync(r.Context(), *request)
		e.handler.writeResponse(w, r, response)
	} else {
		// For push/pull, redirect to appropriate handler
		switch request.Direction {
		case shared.SyncLocalToMCP:
			e.PushSync(w, r)
		case shared.SyncMcpToLocal:
			e.PullSync(w, r)
		default:
			e.handler.writeErrorResponse(w, r, request.RequestID, http.StatusBadRequest,
				&ValidationError{Message: "invalid sync direction"})
		}
	}
}

// PushSync handles push-only synchronization (local to remote)
// POST /api/sync/push
func (e *syncEndpointsImpl) PushSync(w http.ResponseWriter, r *http.Request) {
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

	// Validate direction
	if request.Direction != shared.SyncLocalToMCP && request.Direction != shared.SyncBidirectional {
		e.handler.writeErrorResponse(w, r, request.RequestID, http.StatusBadRequest,
			&ValidationError{Message: "invalid direction for push sync"})
		return
	}

	// Perform push sync
	response := e.performPushSync(r.Context(), *request)
	e.handler.writeResponse(w, r, response)
}

// PullSync handles pull-only synchronization (remote to local)
// POST /api/sync/pull
func (e *syncEndpointsImpl) PullSync(w http.ResponseWriter, r *http.Request) {
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

	// Validate direction
	if request.Direction != shared.SyncMcpToLocal && request.Direction != shared.SyncBidirectional {
		e.handler.writeErrorResponse(w, r, request.RequestID, http.StatusBadRequest,
			&ValidationError{Message: "invalid direction for pull sync"})
		return
	}

	// Perform pull sync
	response := e.performPullSync(r.Context(), *request)
	e.handler.writeResponse(w, r, response)
}

// Health handles health check for sync service
// GET /api/sync/health
func (e *syncEndpointsImpl) Health(w http.ResponseWriter, r *http.Request) {
	// Only accept GET
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Perform health check using sync service
	if err := e.syncService.HealthCheck(r.Context()); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"unhealthy","error":"` + err.Error() + `","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
		return
	}

	// Write healthy response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `","version":"1.0.0"}`))
}

// Sync business logic methods - delegate to sync service

func (e *syncEndpointsImpl) performFullSync(ctx context.Context, request shared.SyncRequest) *shared.SyncResponse {
	response, err := e.syncService.PerformFullSync(ctx, request)
	if err != nil {
		// Return error response
		return &shared.SyncResponse{
			Success:   false,
			RequestID: request.RequestID,
			Timestamp: time.Now(),
			Duration:  0,
			// Error details would be included here
		}
	}
	return response
}

func (e *syncEndpointsImpl) performPushSync(ctx context.Context, request shared.SyncRequest) *shared.SyncResponse {
	response, err := e.syncService.PerformPushSync(ctx, request)
	if err != nil {
		// Return error response
		return &shared.SyncResponse{
			Success:   false,
			RequestID: request.RequestID,
			Timestamp: time.Now(),
			Duration:  0,
		}
	}
	return response
}

func (e *syncEndpointsImpl) performPullSync(ctx context.Context, request shared.SyncRequest) *shared.SyncResponse {
	response, err := e.syncService.PerformPullSync(ctx, request)
	if err != nil {
		// Return error response
		return &shared.SyncResponse{
			Success:   false,
			RequestID: request.RequestID,
			Timestamp: time.Now(),
			Duration:  0,
		}
	}
	return response
}

// Helper function to count entities in a dataset
func countEntities(dataset *shared.SyncDataSet) int {
	if dataset == nil {
		return 0
	}

	count := 0
	if dataset.Projects != nil {
		count += len(dataset.Projects)
	}
	if dataset.Tasks != nil {
		count += len(dataset.Tasks)
	}
	return count
}
