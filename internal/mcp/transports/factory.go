package transports

import (
	"fmt"

	"github.com/denkhaus/knot/v2/internal/config"
	"github.com/samber/do/v2"
)


// NewTransportProvider creates a transport provider that selects the right transport based on configuration
func NewTransportProvider(injector do.Injector) (Transport, error) {
	// Get configuration to determine which transport to create
	configService := do.MustInvoke[config.Service](injector)
	transportMode := configService.GetMCPConfig().Transport.Mode

	if !transportMode.IsValid() {
		return nil, fmt.Errorf("invalid transport type: %s", transportMode.String())
	}

	// Create the appropriate transport based on mode
	switch transportMode {
	case config.TransportTypeStdio:
		return NewStdioTransportProvider(injector)
	case config.TransportTypeHTTP:
		return NewHTTPTransportProvider(injector)
	case config.TransportTypeSSE:
		return NewSSETransportProvider(injector)
	default:
		return nil, fmt.Errorf("unsupported transport type: %s", transportMode.String())
	}
}