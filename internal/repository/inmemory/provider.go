package inmemory

import (
	"github.com/denkhaus/knot/v2/internal/types"
	"github.com/samber/do/v2"
)

// Provider provides in-memory repository instances with DI support
type Provider struct{}

// NewProvider creates a new in-memory repository provider
func NewProvider(injector do.Injector) (*Provider, error) {
	return &Provider{}, nil
}

// NewRepository creates a new in-memory repository
func (p *Provider) NewRepository() types.Repository {
	return NewMemoryRepository()
}

// ProvideRepository is a convenience function for DI registration
func ProvideRepository(injector do.Injector) (types.Repository, error) {
	provider := do.MustInvoke[*Provider](injector)
	return provider.NewRepository(), nil
}