package repository

// ProviderType defines well-defined types for named dependency injection providers
type ProviderType string

// RepositoryProvider constants for named DI providers
const (
	// Repository providers
	SQLiteProvider    ProviderType = "sqlite_provider"
	PostgreSQLProvider ProviderType = "postgres_provider"
	InMemoryProvider   ProviderType = "inmemory_provider"

	// Repository instances
	LocalRepository ProviderType = "local_repository"
	MCPRepository   ProviderType = "mcp_repository"
)

// String returns the string representation of the provider type
func (p ProviderType) String() string {
	return string(p)
}

// IsSQLiteProvider checks if the provider type is SQLite
func (p ProviderType) IsSQLiteProvider() bool {
	return p == SQLiteProvider
}

// IsPostgreSQLProvider checks if the provider type is PostgreSQL
func (p ProviderType) IsPostgreSQLProvider() bool {
	return p == PostgreSQLProvider
}

// IsInMemoryProvider checks if the provider type is InMemory
func (p ProviderType) IsInMemoryProvider() bool {
	return p == InMemoryProvider
}

// IsLocalRepository checks if this is a local repository
func (p ProviderType) IsLocalRepository() bool {
	return p == LocalRepository
}

// IsMCPRepository checks if this is an MCP repository
func (p ProviderType) IsMCPRepository() bool {
	return p == MCPRepository
}