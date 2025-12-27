package transports

import (
	"fmt"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/samber/do/v2"
)

// NewTransport creates a transport provider that selects the right transport based on configuration
func NewTransport(injector do.Injector) (Transport, error) {
	// Get configuration to determine which transport to create
	configService := do.MustInvoke[config.Service](injector)
	transportMode := configService.GetMCPConfig().Transport.Mode

	if !transportMode.IsValid() {
		return nil, fmt.Errorf("invalid transport type: %s", transportMode.String())
	}

	// Create the appropriate transport based on mode
	switch transportMode {
	case config.TransportTypeStdio:
		return newStdioTransport(injector)
	case config.TransportTypeHTTP:
		return newHTTPTransport(injector)
	case config.TransportTypeSSE:
		return newSSETransport(injector)
	default:
		return nil, fmt.Errorf("unsupported transport type: %s", transportMode.String())
	}
}
