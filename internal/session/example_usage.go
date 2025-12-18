package session

import (
	"fmt"

	"github.com/samber/do/v2"
)

// ExampleUsage demonstrates how to use the typed session providers
// This shows the benefits of well-typed providers over magic strings
func ExampleUsage(injector do.Injector) {
	// ✅ GOOD: Using typed constants - compile-time safe, IDE friendly
	memorySessionManager, err := do.InvokeNamed[Manager](injector, MemorySessionProvider.String())
	if err != nil {
		fmt.Printf("Failed to get memory session manager: %v\n", err)
		return
	}

	databaseSessionManager, err := do.InvokeNamed[Manager](injector, DatabaseSessionProvider.String())
	if err != nil {
		fmt.Printf("Failed to get database session manager: %v\n", err)
		return
	}

	_ = memorySessionManager
	_ = databaseSessionManager

	// ✅ GOOD: Type checking with helper methods
	var providerType SessionProviderType = MemorySessionProvider
	if providerType.IsMemorySessionProvider() {
		fmt.Println("Using memory session provider")
	}

	// ❌ BAD: Magic strings (old approach) - no type safety, prone to typos
	// memoryManager, err := do.InvokeNamed[Manager](injector, "memory-session") // Compile error if typo!
	// databaseManager, err := do.InvokeNamed[Manager](injector, "database-session") // Compile error if typo!

	// ✅ GOOD: Using typed constants in switch statements
	switch providerType {
	case MemorySessionProvider:
		fmt.Println("Memory session provider selected")
	case DatabaseSessionProvider:
		fmt.Println("Database session provider selected")
	default:
		fmt.Printf("Unknown session provider: %s\n", providerType)
	}
}

// ProviderTypeValidation demonstrates how the typed constants help with validation
func ProviderTypeValidation(providerType SessionProviderType) error {
	// ✅ GOOD: Type-safe validation
	switch providerType {
	case MemorySessionProvider, DatabaseSessionProvider:
		return nil
	default:
		return fmt.Errorf("invalid session provider type: %s", providerType)
	}
}