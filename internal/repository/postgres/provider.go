package postgres

import (
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
)

// Provider provides PostgreSQL repository instances with DI support
// Used for MCP Mode (multi-project server with PostgreSQL backend)
type Provider struct{}

// NewProvider creates a new PostgreSQL repository provider
func NewProvider(injector do.Injector) (*Provider, error) {
	return &Provider{}, nil
}

// NewRepository creates a new PostgreSQL repository with DI dependencies
func (p *Provider) NewRepository(dsn string, opts ...Option) (types.Repository, error) {
	// For now, create repository with default configuration
	// In a full implementation, we would inject logger and config dependencies
	return NewRepository(dsn, opts...)
}

// ProvideRepository is a convenience function for DI registration
func ProvideRepository(injector do.Injector) (types.Repository, error) {
	provider := do.MustInvoke[*Provider](injector)
	loggerService := do.MustInvoke[logger.Logger](injector)

	// Use logger from DI - DSN will come from config in real implementation
	return provider.NewRepository("",
		WithLogger(loggerService.ToZap()),
		WithAutoMigrate(true),
	)
}