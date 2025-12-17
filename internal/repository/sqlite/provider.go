package sqlite

import (
	"github.com/denkhaus/knot/v2/internal/logger"
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
)

// Provider provides SQLite repository instances with DI support
type Provider struct{}

// NewProvider creates a new SQLite repository provider
func NewProvider(injector do.Injector) (*Provider, error) {
	return &Provider{}, nil
}

// NewRepository creates a new SQLite repository with DI dependencies
func (p *Provider) NewRepository(dbPath string, opts ...Option) (types.Repository, error) {
	// For now, create repository with default configuration
	// In a full implementation, we would inject logger and config dependencies
	return NewRepository(dbPath, opts...)
}

// ProvideRepository is a convenience function for DI registration
func ProvideRepository(injector do.Injector) (types.Repository, error) {
	provider := do.MustInvoke[*Provider](injector)
	loggerService := do.MustInvoke[logger.Logger](injector)

	// Use logger from DI
	return provider.NewRepository("",
		WithLogger(loggerService.ToZap()),
		WithAutoMigrate(true),
	)
}