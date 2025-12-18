package session

import (
	"github.com/samber/do/v2"
)

// NewSessionManager creates a provider function for the SessionManager
// This follows the dependency injection pattern used throughout the application
func NewSessionManager(injector do.Injector) (Manager, error) {
	// Session manager doesn't have dependencies currently
	// but this pattern allows for future dependencies
	return newManager(), nil
}
