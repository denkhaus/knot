package postgres

import (
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
)

// provider provides PostgreSQL repository instances with DI support
// Private implementation of RepositoryProvider interface
type provider struct{}

// Ensure provider implements RepositoryProvider
var _ types.RepositoryProvider = (*provider)(nil)

// NewProvider creates a new PostgreSQL repository provider
func NewProvider(injector do.Injector) (types.RepositoryProvider, error) {
	return &provider{}, nil
}

// NewRepository creates a new PostgreSQL repository with DI dependencies
func (p *provider) NewRepository(dsn string, opts ...interface{}) (types.Repository, error) {
	// Convert interface{} options to Option type
	var postgresOpts []Option
	for _, opt := range opts {
		if option, ok := opt.(Option); ok {
			postgresOpts = append(postgresOpts, option)
		}
	}
	return NewRepository(dsn, postgresOpts...)
}
