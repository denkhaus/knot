package shared

// Configuration parameter names
const (
	// Config keys used in CLI commands
	ConfigKeyComplexityThreshold  = "complexity-threshold"
	ConfigKeyMaxDepth             = "max-depth"
	ConfigKeyMaxTasksPerDepth     = "max-tasks-per-depth"
	ConfigKeyMaxDescriptionLength = "max-description-length"
	ConfigKeyAutoReduceComplexity = "auto-reduce-complexity"
)

// Task state values
const (
	TaskStateCompleted = "completed"
	TaskStatePending   = "pending"
)

// Selection strategy names
const (
	StrategyCreationOrder   = "creation-order"
	StrategyDependencyAware = "dependency-aware"
	StrategyPriority        = "priority"
	StrategyDepthFirst      = "depth-first"
	StrategyCriticalPath    = "critical-path"
)

// Actor types
const (
	ActorUnknown = "unknown"
)

// Tree formatting characters
const (
	TreePrefixRight     = "└── "
	TreePrefixMiddle    = "├── "
	TreePrefixSeparator = "│"
)

// Health test parameters
const (
	HealthTestTimeout = "timeout"
)

// Colors
const (
	ColorGreen  = "green"
	ColorYellow = "yellow"
	ColorRed    = "red"
)
