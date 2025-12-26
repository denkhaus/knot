package selection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTaskSelector(t *testing.T) {
	t.Run("ValidConfiguration", func(t *testing.T) {
		selector, err := NewTaskSelector(DefaultConfig())

		assert.NoError(t, err)
		assert.NotNil(t, selector)
		assert.NotNil(t, selector.analyzer)
		assert.NotNil(t, selector.filter)
		assert.NotNil(t, selector.config)
	})

	t.Run("NilConfiguration", func(t *testing.T) {
		selector, err := NewTaskSelector(nil)

		assert.NoError(t, err)
		assert.NotNil(t, selector)
		assert.NotNil(t, selector.config) // Should use default config
	})

	t.Run("InvalidConfiguration", func(t *testing.T) {
		config := &Config{
			DependentCountWeight: -0.1, // Invalid negative weight
			PriorityWeight:       0.4,
			DepthFirstWeight:     0.4,
			CriticalPathWeight:   0.3,
		}

		_, err := NewTaskSelector(config)

		assert.Error(t, err)
		// With -0.1 + 0.4 + 0.4 + 0.3 = 1.0, the sum check passes but the negative check fails
		assert.Contains(t, err.Error(), "all weights must be non-negative")
	})

	t.Run("WeightsDontSumToOne", func(t *testing.T) {
		config := &Config{
			DependentCountWeight: 0.5,
			PriorityWeight:       0.5,
			DepthFirstWeight:     0.5,
			CriticalPathWeight:   0.5,
		}

		_, err := NewTaskSelector(config)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "weights should sum to approximately 1.0")
	})
}

func TestValidateConfig(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		config := DefaultConfig()
		err := ValidateConfig(config)
		assert.NoError(t, err)
	})

	t.Run("NilConfig", func(t *testing.T) {
		err := ValidateConfig(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("NegativeScoreThreshold", func(t *testing.T) {
		config := DefaultConfig()
		config.ScoreThreshold = -1.0

		err := ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "score threshold cannot be negative")
	})

	t.Run("NegativeMaxDependencyDepth", func(t *testing.T) {
		config := DefaultConfig()
		config.MaxDependencyDepth = -1

		err := ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max dependency depth cannot be negative")
	})

	t.Run("NegativeCacheDuration", func(t *testing.T) {
		config := DefaultConfig()
		config.CacheDuration = -1

		err := ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache duration cannot be negative")
	})
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 0.4, config.DependentCountWeight)
	assert.Equal(t, 0.3, config.PriorityWeight)
	assert.Equal(t, 0.2, config.DepthFirstWeight)
	assert.Equal(t, 0.1, config.CriticalPathWeight)
	assert.True(t, config.PreferInProgress)
	assert.Equal(t, 10, config.MaxDependencyDepth)
	assert.Equal(t, 0.0, config.ScoreThreshold)
	assert.True(t, config.CacheGraphs)
}
