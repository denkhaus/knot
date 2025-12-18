package sqlite

import (
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
)

// provider provides SQLite repository instances with DI support
// Private implementation of RepositoryProvider interface
type provider struct{}

// Ensure provider implements RepositoryProvider
var _ types.RepositoryProvider = (*provider)(nil)

// NewProvider creates a new SQLite repository provider
func NewProvider(injector do.Injector) (types.RepositoryProvider, error) {
	return &provider{}, nil
}

// NewRepository creates a new SQLite repository with DI dependencies
func (p *provider) NewRepository(dbPath string, opts ...interface{}) (types.Repository, error) {
	// Convert interface{} options to Option type
	var sqliteOpts []Option
	for _, opt := range opts {
		if option, ok := opt.(Option); ok {
			sqliteOpts = append(sqliteOpts, option)
		}
	}
	return NewRepository(dbPath, sqliteOpts...)
}
