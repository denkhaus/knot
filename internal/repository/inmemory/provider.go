package inmemory

import (
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
)

// provider provides in-memory repository instances with DI support
// Private implementation of RepositoryProvider interface
type provider struct{}

// Ensure provider implements RepositoryProvider
var _ types.RepositoryProvider = (*provider)(nil)

// NewProvider creates a new in-memory repository provider
func NewProvider(injector do.Injector) (types.RepositoryProvider, error) {
	return &provider{}, nil
}

// NewRepository creates a new in-memory repository
func (p *provider) NewRepository(dsn string, opts ...interface{}) (types.Repository, error) {
	// In-memory repository doesn't use DSN or options
	return NewMemoryRepository(), nil
}
