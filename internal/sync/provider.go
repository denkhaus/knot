// Package sync provider provides dependency injection setup for sync components
package sync

import (
	"github.com/denkhaus/knot/v2/internal/sync/client"
	"github.com/samber/do/v2"
)

// NewRESTSyncClientProvider creates a REST sync client from DI dependencies
func NewRESTSyncClientProvider(injector do.Injector) (client.RESTSyncClient, error) {
	// The client package has its own provider that handles DI
	return client.NewRESTSyncClient(injector)
}
