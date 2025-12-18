package session

// SessionProviderType defines the type for session provider identifiers
type SessionProviderType string

// Session provider constants for named DI providers
const (
	// MemorySessionProvider represents in-memory session storage
	MemorySessionProvider SessionProviderType = "memory_session_provider"

	// DatabaseSessionProvider represents database-backed session storage
	DatabaseSessionProvider SessionProviderType = "database_session_provider"
)

// String returns the string representation of the session provider type
func (s SessionProviderType) String() string {
	return string(s)
}

// IsMemorySessionProvider checks if the provider type is MemorySessionProvider
func (s SessionProviderType) IsMemorySessionProvider() bool {
	return s == MemorySessionProvider
}

// IsDatabaseSessionProvider checks if the provider type is DatabaseSessionProvider
func (s SessionProviderType) IsDatabaseSessionProvider() bool {
	return s == DatabaseSessionProvider
}