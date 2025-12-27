// Package client provides REST-based HTTP client for sync operations
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/sync/shared"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"go.uber.org/zap"
)

// Default configuration values
const (
	// DefaultTimeout is the default request timeout
	DefaultTimeout = 30 * time.Second
	// DefaultRetryAttempts is the default number of retry attempts
	DefaultRetryAttempts = 3
	// DefaultRetryDelay is the default delay between retries
	DefaultRetryDelay = 1 * time.Second
	// DefaultMaxIdleConns is the default maximum idle connections
	DefaultMaxIdleConns = 10
	// DefaultIdleConnTimeout is the default idle connection timeout
	DefaultIdleConnTimeout = 90 * time.Second
)

// RESTSyncClient defines the interface for REST sync client operations
type RESTSyncClient interface {
	// Main sync operation - sends sync request to remote server
	Sync(ctx context.Context, req *shared.SyncRequest) (*shared.SyncResponse, error)

	// Read-only operations for data retrieval
	GetRemoteData(ctx context.Context, projectID *uuid.UUID) (*shared.SyncDataSet, error)
	GetProjects(ctx context.Context) (*shared.SyncDataSet, error)
	GetTasks(ctx context.Context, projectID uuid.UUID) (*shared.SyncDataSet, error)

	// Health check
	HealthCheck(ctx context.Context) error
}

// restSyncClientImpl provides HTTP-based communication with sync API endpoints
type restSyncClientImpl struct {
	config     *config.SyncConfig
	httpClient *http.Client
	serializer shared.SyncDataSerializer
	logger     logger.Logger
	mu         sync.RWMutex
}

// NewRESTSyncClient creates a new REST sync client instance
func NewRESTSyncClient(injector do.Injector) (RESTSyncClient, error) {
	cnf := do.MustInvoke[config.Service](injector)
	log := do.MustInvoke[logger.Logger](injector)
	serializer := do.MustInvoke[shared.SyncDataSerializer](injector)

	cfg := cnf.GetSyncConfig()

	// Create HTTP client with timeout and connection pool settings
	httpClient := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.MaxIdleConns,
			MaxIdleConnsPerHost: cfg.MaxIdleConns,
			IdleConnTimeout:     cfg.IdleConnTimeout,
		},
	}

	return &restSyncClientImpl{
		config:     cfg,
		httpClient: httpClient,
		serializer: serializer,
		logger:     log,
	}, nil
}

// Sync performs a sync operation with the remote server
func (c *restSyncClientImpl) Sync(ctx context.Context, req *shared.SyncRequest) (*shared.SyncResponse, error) {
	c.logger.Debug("Starting sync operation",
		zap.String("request_id", req.RequestID.String()),
		zap.String("project_id", req.ProjectID.String()),
		zap.String("direction", string(req.Direction)))

	// Determine endpoint based on direction
	var endpointPath string
	switch req.Direction {
	case shared.SyncLocalToMCP:
		endpointPath = "/api/sync/push"
	case shared.SyncMcpToLocal:
		endpointPath = "/api/sync/pull"
	case shared.SyncBidirectional:
		endpointPath = "/api/sync/full"
	default:
		return nil, fmt.Errorf("unsupported sync direction: %s", req.Direction)
	}

	// Build request URL
	endpointURL, err := c.buildEndpointURL(endpointPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build endpoint URL: %w", err)
	}

	// Serialize request to JSON (msgpack no longer supported)
	requestData, err := c.serializer.SerializeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpointURL, bytes.NewReader(requestData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	c.setRequestHeaders(httpReq, "json")

	// Execute request with retry logic
	responseData, err := c.executeWithRetry(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute sync request: %w", err)
	}

	// Deserialize response (JSON only)
	resp, err := c.serializer.DeserializeResponse(responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize response: %w", err)
	}

	c.logger.Debug("Sync operation completed",
		zap.String("request_id", req.RequestID.String()),
		zap.Bool("success", resp.Success),
		zap.Int("processed", resp.Processed))

	return resp, nil
}

// GetProjects retrieves all projects from the sync server
func (c *restSyncClientImpl) GetProjects(ctx context.Context) (*shared.SyncDataSet, error) {
	c.logger.Debug("Fetching projects from sync server")

	// Build request URL
	endpointURL, err := c.buildEndpointURL("/api/sync/projects")
	if err != nil {
		return nil, fmt.Errorf("failed to build endpoint URL: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", endpointURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Accept", "application/json")
	if c.config.AuthToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.config.AuthToken)
	}

	// Execute request with retry logic
	responseData, err := c.executeWithRetry(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get projects request: %w", err)
	}

	// Deserialize response
	var dataSet shared.SyncDataSet
	if err := json.Unmarshal(responseData, &dataSet); err != nil {
		return nil, fmt.Errorf("failed to deserialize projects response: %w", err)
	}

	return &dataSet, nil
}

// GetTasks retrieves all tasks for a project from the sync server
func (c *restSyncClientImpl) GetTasks(ctx context.Context, projectID uuid.UUID) (*shared.SyncDataSet, error) {
	c.logger.Debug("Fetching tasks from sync server",
		zap.String("project_id", projectID.String()))

	// Build request URL with query parameters
	endpointURL, err := c.buildEndpointURL("/api/sync/tasks")
	if err != nil {
		return nil, fmt.Errorf("failed to build endpoint URL: %w", err)
	}

	// Add query parameters
	urlObj, err := url.Parse(endpointURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint URL: %w", err)
	}
	query := urlObj.Query()
	query.Add("project_id", projectID.String())
	if c.config.Since != nil {
		query.Add("since", c.config.Since.Format(time.RFC3339))
	}
	urlObj.RawQuery = query.Encode()

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", urlObj.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Accept", "application/json")
	if c.config.AuthToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.config.AuthToken)
	}

	// Execute request with retry logic
	responseData, err := c.executeWithRetry(ctx, httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get tasks request: %w", err)
	}

	// Deserialize response
	var dataSet shared.SyncDataSet
	if err := json.Unmarshal(responseData, &dataSet); err != nil {
		return nil, fmt.Errorf("failed to deserialize tasks response: %w", err)
	}

	return &dataSet, nil
}

// GetRemoteData retrieves both projects and tasks from the sync server
func (c *restSyncClientImpl) GetRemoteData(ctx context.Context, projectID *uuid.UUID) (*shared.SyncDataSet, error) {
	c.logger.Debug("Fetching remote data from sync server")

	dataSet := &shared.SyncDataSet{
		Projects: make(map[uuid.UUID]*types.Project),
		Tasks:    make(map[uuid.UUID]*types.Task),
	}

	// Fetch projects
	projects, err := c.GetProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}
	if projects.Projects != nil {
		dataSet.Projects = projects.Projects
	}

	// Fetch tasks if project ID is specified
	if projectID != nil {
		tasks, err := c.GetTasks(ctx, *projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tasks: %w", err)
		}
		if tasks.Tasks != nil {
			dataSet.Tasks = tasks.Tasks
		}
	}

	return dataSet, nil
}

// HealthCheck performs a health check on the sync server
func (c *restSyncClientImpl) HealthCheck(ctx context.Context) error {
	c.logger.Debug("Performing health check on sync server")

	// Build request URL
	endpointURL, err := c.buildEndpointURL("/api/sync/health")
	if err != nil {
		return fmt.Errorf("failed to build endpoint URL: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", endpointURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Accept", "application/json")

	// Execute request (no retry for health check)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// Close closes the REST sync client and releases resources
func (c *restSyncClientImpl) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close the HTTP client's transport to close idle connections
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}

	c.logger.Debug("REST sync client closed")
	return nil
}

// executeWithRetry executes an HTTP request with retry logic
func (c *restSyncClientImpl) executeWithRetry(ctx context.Context, req *http.Request) ([]byte, error) {
	var lastErr error
	var responseData []byte

	for attempt := 0; attempt <= c.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(c.calculateRetryDelay(attempt)):
			case <-ctx.Done():
				return nil, ctx.Err()
			}

			c.logger.Debug("Retrying request",
				zap.String("url", req.URL.String()),
				zap.Int("attempt", attempt))
		}

		// Execute request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			c.logger.Warn("Request failed",
				zap.String("url", req.URL.String()),
				zap.Int("attempt", attempt),
				zap.Error(err))
			continue
		}

		// Read response body
		responseData, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			c.logger.Warn("Failed to read response",
				zap.String("url", req.URL.String()),
				zap.Int("attempt", attempt),
				zap.Error(err))
			continue
		}

		// Check status code
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Success
			return responseData, nil
		}

		// Check if error is retryable
		if !c.isRetryableStatusCode(resp.StatusCode) {
			// Non-retryable error
			return nil, c.formatErrorResponse(resp.StatusCode, responseData)
		}

		lastErr = c.formatErrorResponse(resp.StatusCode, responseData)
		c.logger.Warn("Request returned retryable error",
			zap.String("url", req.URL.String()),
			zap.Int("status", resp.StatusCode),
			zap.Int("attempt", attempt))
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.config.RetryAttempts+1, lastErr)
}

// calculateRetryDelay calculates the delay before retry with exponential backoff
func (c *restSyncClientImpl) calculateRetryDelay(attempt int) time.Duration {
	// Exponential backoff: delay * 2^(attempt-1)
	backoff := c.config.RetryDelay * time.Duration(1<<uint(attempt-1))

	// Cap at maximum retry delay
	if backoff > c.config.MaxRetryDelay {
		backoff = c.config.MaxRetryDelay
	}

	return backoff
}

// isRetryableStatusCode checks if a status code is retryable
func (c *restSyncClientImpl) isRetryableStatusCode(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || // 429
		statusCode == http.StatusInternalServerError || // 500
		statusCode == http.StatusBadGateway || // 502
		statusCode == http.StatusServiceUnavailable || // 503
		statusCode == http.StatusGatewayTimeout // 504
}

// formatErrorResponse formats an error response from the server
func (c *restSyncClientImpl) formatErrorResponse(statusCode int, body []byte) error {
	// Try to parse as JSON error response
	var errResp map[string]interface{}
	if err := json.Unmarshal(body, &errResp); err == nil {
		if errMsg, ok := errResp["error"].(string); ok && errMsg != "" {
			return fmt.Errorf("sync server error (status %d): %s", statusCode, errMsg)
		}
	}

	// Fallback to raw body
	return fmt.Errorf("sync server error (status %d): %s", statusCode, string(body))
}

// setRequestHeaders sets the required headers for an HTTP request (JSON only)
func (c *restSyncClientImpl) setRequestHeaders(req *http.Request, format string) {
	// format parameter is ignored - only JSON is supported
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.AuthToken)
	}

	if c.config.UserAgent != "" {
		req.Header.Set("User-Agent", c.config.UserAgent)
	}
}

// buildEndpointURL builds a full endpoint URL from a path
func (c *restSyncClientImpl) buildEndpointURL(endpointPath string) (string, error) {
	baseURL, err := url.Parse(c.config.ServerURL)
	if err != nil {
		return "", fmt.Errorf("invalid server URL: %w", err)
	}

	baseURL.Path = path.Join(baseURL.Path, endpointPath)
	return baseURL.String(), nil
}

