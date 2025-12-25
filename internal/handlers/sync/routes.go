package sync

import (
	"github.com/go-chi/chi/v5"
	"github.com/samber/do/v2"
)

// SetupRoutes registers sync HTTP routes with the router
func SetupRoutes(r chi.Router, injector do.Injector) error {
	// Resolve sync handlers
	handlers, err := NewSyncHTTPHandlers(injector)
	if err != nil {
		return err
	}

	// Apply middleware
	r.Group(func(r chi.Router) {
		r.Use(handlers.Middleware())

		// Register routes
		r.Route("/api/sync", func(r chi.Router) {
			// Health check
			r.Get("/health", handlers.HealthHandler())

			// Sync operations
			r.Post("/full", handlers.FullSyncHandler())
			r.Post("/push", handlers.PushSyncHandler())
			r.Post("/pull", handlers.PullSyncHandler())

			// Projects CRUD
			r.Get("/projects", handlers.ListProjectsHandler())

			// Tasks CRUD
			r.Get("/tasks", handlers.ListTasksHandler())
		})
	})

	return nil
}
