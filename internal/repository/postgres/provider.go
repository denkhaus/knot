package postgres

import (
	"fmt"

	"github.com/denkhaus/knot/v2/internal/config"
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
	return NewRepository(dsn, opts...)
}

// ProvideRepository is a convenience function for DI registration using MCP config
func ProvideRepository(injector do.Injector) (types.Repository, error) {
	provider := do.MustInvoke[*Provider](injector)
	configService := do.MustInvoke[config.Service](injector)
	loggerService := do.MustInvoke[logger.Logger](injector)

	// Get PostgreSQL DSN from MCP config
	dsn := configService.GetMCPConfig().Database.Endpoint
	if dsn == "" {
		return nil, fmt.Errorf("PostgreSQL DSN not configured in MCP config")
	}

	// Create repository with DI dependencies and MCP config
	return provider.NewRepository(dsn,
		WithLogger(loggerService.ToZap()),
		WithAutoMigrate(true),
		WithConnectionPool(25, 5), // Higher concurrency for MCP server
		WithMigrationTimeout(configService.GetMCPConfig().Session.Timeout),
	)
}