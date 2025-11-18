package selection

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ConfigProvider manages configuration loading, saving, and validation
type DefaultConfigProvider struct {
	config     *Config
	configPath string
	mutex      sync.RWMutex
	watchers   []ConfigWatcher
}

// ConfigWatcher defines a callback for configuration changes
type ConfigWatcher func(oldConfig, newConfig *Config) error

// NewConfigProvider creates a new configuration provider
func NewConfigProvider(configPath string) *DefaultConfigProvider {
	return &DefaultConfigProvider{
		configPath: configPath,
		watchers:   make([]ConfigWatcher, 0),
	}
}

// LoadConfig loads configuration from file, falling back to defaults
func (cp *DefaultConfigProvider) LoadConfig() (*Config, error) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	// If config is already loaded and fresh, return it
	if cp.config != nil {
		return cp.config, nil
	}

	// Try to load from file
	config, err := cp.loadFromFile()
	if err != nil {
		// Fall back to default config
		config = DefaultConfig()
	}

	// Validate the loaded config
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("loaded configuration is invalid: %w", err)
	}

	cp.config = config
	return config, nil
}

// loadFromFile loads configuration from the specified file
func (cp *DefaultConfigProvider) loadFromFile() (*Config, error) {
	if cp.configPath == "" {
		return DefaultConfig(), nil
	}

	// Check if file exists
	if _, err := os.Stat(cp.configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	// Read file
	data, err := ioutil.ReadFile(cp.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Fill in missing fields with defaults
	cp.fillDefaults(&config)

	return &config, nil
}

// fillDefaults fills in missing configuration fields with defaults
func (cp *DefaultConfigProvider) fillDefaults(config *Config) {
	defaults := DefaultConfig()

	// Fill strategy if not set
	if config.Strategy.String() == "unknown" {
		config.Strategy = defaults.Strategy
	}

	// Fill weights if empty
	if config.Weights == (Weights{}) {
		factory := &StrategyFactory{}
		config.Weights = factory.GetDefaultWeightsForStrategy(config.Strategy)
	}

	// Fill behavior settings if not set
	if config.Behavior == (BehaviorConfig{}) {
		config.Behavior = defaults.Behavior
	}

	// Fill advanced settings if not set
	if config.Advanced == (AdvancedConfig{}) {
		config.Advanced = defaults.Advanced
	}
}

// SaveConfig saves configuration to file
func (cp *DefaultConfigProvider) SaveConfig(config *Config) error {
	if err := ValidateConfig(config); err != nil {
		return fmt.Errorf("cannot save invalid configuration: %w", err)
	}

	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	// Create directory if it doesn't exist
	if cp.configPath != "" {
		dir := filepath.Dir(cp.configPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Marshal to JSON
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		// Write to file
		if err := ioutil.WriteFile(cp.configPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
	}

	// Notify watchers of the change
	oldConfig := cp.config
	cp.config = config

	for _, watcher := range cp.watchers {
		if err := watcher(oldConfig, config); err != nil {
			// Log the error but don't fail the save
			fmt.Printf("Warning: config watcher failed: %v\n", err)
		}
	}

	return nil
}

// GetConfig returns the current configuration
func (cp *DefaultConfigProvider) GetConfig() *Config {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()

	if cp.config == nil {
		// Try to load config
		config, err := cp.LoadConfig()
		if err != nil {
			return DefaultConfig()
		}
		return config
	}

	return cp.config
}

// SetConfig sets the configuration without saving to file
func (cp *DefaultConfigProvider) SetConfig(config *Config) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	oldConfig := cp.config
	cp.config = config

	// Notify watchers
	for _, watcher := range cp.watchers {
		if err := watcher(oldConfig, config); err != nil {
			fmt.Printf("Warning: config watcher failed: %v\n", err)
		}
	}
}

// ValidateConfig validates the current configuration
func (cp *DefaultConfigProvider) ValidateConfig() error {
	config := cp.GetConfig()
	return ValidateConfig(config)
}

// AddConfigWatcher adds a watcher for configuration changes
func (cp *DefaultConfigProvider) AddConfigWatcher(watcher ConfigWatcher) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	cp.watchers = append(cp.watchers, watcher)
}

// GetConfigPath returns the path to the configuration file
func (cp *DefaultConfigProvider) GetConfigPath() string {
	return cp.configPath
}

// ResetConfig resets configuration to defaults
func (cp *DefaultConfigProvider) ResetConfig() error {
	defaultConfig := DefaultConfig()
	return cp.SaveConfig(defaultConfig)
}

// ConfigTemplate represents a configuration template for different project types
type ConfigTemplate struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Config      *Config `json:"config"`
}

// GetBuiltinTemplates returns predefined configuration templates
func GetBuiltinTemplates() []ConfigTemplate {
	return []ConfigTemplate{
		{
			Name:        "default",
			Description: "Balanced approach suitable for most projects",
			Config:      DefaultConfig(),
		},
		{
			Name:        "priority-driven",
			Description: "Focus on high-priority tasks first",
			Config: &Config{
				Strategy: StrategyPriority,
				Weights: Weights{
					Priority:       0.8,
					DependentCount: 0.2,
					DepthFirst:     0.0,
					CriticalPath:   0.0,
				},
				Behavior: BehaviorConfig{
					AllowParentWithSubtasks: false,
					PreferInProgress:       true,
					BreakTiesByCreation:    true,
					StrictDependencies:     true,
				},
				Advanced: AdvancedConfig{
					MaxDependencyDepth: 10,
					ScoreThreshold:     0.0,
					CacheGraphs:       true,
					CacheDuration:     5 * time.Minute,
				},
			},
		},
		{
			Name:        "depth-first",
			Description: "Complete subtasks before moving to other branches",
			Config: &Config{
				Strategy: StrategyDepthFirst,
				Weights: Weights{
					DepthFirst:     0.7,
					Priority:       0.2,
					DependentCount: 0.1,
					CriticalPath:   0.0,
				},
				Behavior: BehaviorConfig{
					AllowParentWithSubtasks: false,
					PreferInProgress:       true,
					BreakTiesByCreation:    true,
					StrictDependencies:     true,
				},
				Advanced: AdvancedConfig{
					MaxDependencyDepth: 15,
					ScoreThreshold:     0.0,
					CacheGraphs:       true,
					CacheDuration:     5 * time.Minute,
				},
			},
		},
		{
			Name:        "dependency-focused",
			Description: "Prioritize tasks that unblock others",
			Config: &Config{
				Strategy: StrategyDependencyAware,
				Weights: Weights{
					DependentCount: 0.6,
					CriticalPath:   0.2,
					Priority:       0.1,
					DepthFirst:     0.1,
				},
				Behavior: BehaviorConfig{
					AllowParentWithSubtasks: true,
					PreferInProgress:       true,
					BreakTiesByCreation:    true,
					StrictDependencies:     true,
				},
				Advanced: AdvancedConfig{
					MaxDependencyDepth: 20,
					ScoreThreshold:     0.0,
					CacheGraphs:       true,
					CacheDuration:     10 * time.Minute,
				},
			},
		},
		{
			Name:        "critical-path",
			Description: "Focus on tasks that affect project timeline",
			Config: &Config{
				Strategy: StrategyCriticalPath,
				Weights: Weights{
					CriticalPath:   0.5,
					DependentCount: 0.3,
					Priority:       0.2,
					DepthFirst:     0.0,
				},
				Behavior: BehaviorConfig{
					AllowParentWithSubtasks: true,
					PreferInProgress:       true,
					BreakTiesByCreation:    false,
					StrictDependencies:     true,
				},
				Advanced: AdvancedConfig{
					MaxDependencyDepth: 25,
					ScoreThreshold:     1.0,
					CacheGraphs:       true,
					CacheDuration:     15 * time.Minute,
				},
			},
		},
		{
			Name:        "legacy-compatible",
			Description: "Mimics the original knot behavior",
			Config: &Config{
				Strategy: StrategyCreationOrder,
				Weights: Weights{
					DependentCount: 0.0,
					Priority:       0.0,
					DepthFirst:     0.0,
					CriticalPath:   0.0,
				},
				Behavior: BehaviorConfig{
					AllowParentWithSubtasks: false,
					PreferInProgress:       false,
					BreakTiesByCreation:    true,
					StrictDependencies:     true,
				},
				Advanced: AdvancedConfig{
					MaxDependencyDepth: 5,
					ScoreThreshold:     0.0,
					CacheGraphs:       false,
					CacheDuration:     0,
				},
			},
		},
	}
}

// ApplyTemplate applies a configuration template
func (cp *DefaultConfigProvider) ApplyTemplate(templateName string) error {
	templates := GetBuiltinTemplates()
	
	for _, template := range templates {
		if template.Name == templateName {
			return cp.SaveConfig(template.Config)
		}
	}
	
	return fmt.Errorf("template %s not found", templateName)
}

// ListTemplates returns available configuration templates
func (cp *DefaultConfigProvider) ListTemplates() []ConfigTemplate {
	return GetBuiltinTemplates()
}

// ConfigValidator provides validation utilities for configurations
type ConfigValidator struct{}

// ValidateConfigFile validates a configuration file without loading it
func (cv *ConfigValidator) ValidateConfigFile(path string) error {
	provider := NewConfigProvider(path)
	_, err := provider.LoadConfig()
	return err
}

// ValidateWeightsSum ensures weights sum to approximately 1.0 for dependency-aware strategy
func (cv *ConfigValidator) ValidateWeightsSum(weights Weights, tolerance float64) error {
	total := weights.DependentCount + weights.Priority + weights.DepthFirst + weights.CriticalPath
	if total < (1.0-tolerance) || total > (1.0+tolerance) {
		return fmt.Errorf("weights sum to %.3f, should be close to 1.0 (±%.3f)", total, tolerance)
	}
	return nil
}

// SuggestWeightAdjustment suggests weight adjustments based on project characteristics
func (cv *ConfigValidator) SuggestWeightAdjustment(taskCount, dependencyCount, hierarchyDepth int) Weights {
	// Simple heuristics for weight suggestions
	weights := Weights{}
	
	if dependencyCount > taskCount/2 {
		// High dependency projects
		weights.DependentCount = 0.5
		weights.CriticalPath = 0.2
		weights.Priority = 0.2
		weights.DepthFirst = 0.1
	} else if hierarchyDepth > 3 {
		// Deep hierarchy projects
		weights.DepthFirst = 0.4
		weights.Priority = 0.3
		weights.DependentCount = 0.2
		weights.CriticalPath = 0.1
	} else {
		// Balanced projects
		weights.DependentCount = 0.4
		weights.Priority = 0.3
		weights.DepthFirst = 0.2
		weights.CriticalPath = 0.1
	}
	
	return weights
}