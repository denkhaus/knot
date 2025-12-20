package transports

import (
	"fmt"

	"github.com/denkhaus/knot/v2/internal/config"
)

// CreateTransport creates a transport instance based on the configuration mode
func CreateTransport(deps TransportDependencies, transportMode config.TransportType) (Transport, error) {
	if !transportMode.IsValid() {
		return nil, fmt.Errorf("invalid transport type: %s", transportMode.String())
	}

	switch transportMode {
	case config.TransportTypeStdio:
		return NewStdioTransport(deps), nil
	case config.TransportTypeHTTP:
		return NewHTTPTransport(deps), nil
	case config.TransportTypeSSE:
		return NewSSETransport(deps), nil
	default:
		return nil, fmt.Errorf("unsupported transport type: %s", transportMode.String())
	}
}