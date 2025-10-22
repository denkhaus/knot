package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInputValidator(t *testing.T) {
	validator := NewInputValidator()
	
	assert.Equal(t, 200, validator.MaxTitleLength)
	assert.Equal(t, 2000, validator.MaxDescriptionLength)
	assert.False(t, validator.AllowHTML)
}

func TestValidateTaskTitle(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name        string
		title       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid title",
			title:       "Valid Task Title",
			expectError: false,
		},
		{
			name:        "empty title",
			title:       "",
			expectError: true,
			errorMsg:    "title cannot be empty",
		},
		{
			name:        "title too long",
			title:       strings.Repeat("A", 201),
			expectError: true,
			errorMsg:    "title too long",
		},
		{
			name:        "title with HTML tags",
			title:       "Task with <script>alert('xss')</script>",
			expectError: true,
			errorMsg:    "contains HTML tags",
		},
		{
			name:        "title with script content",
			title:       "Task with javascript:alert(1)",
			expectError: true,
			errorMsg:    "contains potentially dangerous script content",
		},
		{
			name:        "title with null bytes",
			title:       "Task with\x00null",
			expectError: true,
			errorMsg:    "contains null bytes",
		},
		{
			name:        "title with control characters",
			title:       "Task with\x01control",
			expectError: true,
			errorMsg:    "contains invalid control characters",
		},
		{
			name:        "title with valid whitespace",
			title:       "Task with\ttabs\nand\rnewlines",
			expectError: false,
		},
		{
			name:        "unicode title",
			title:       "Task with émojis 🚀 and ünïcödé",
			expectError: false,
		},
		{
			name:        "max length title",
			title:       strings.Repeat("A", 200),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTaskTitle(tt.title)
			
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTaskDescription(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name        string
		description string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid description",
			description: "This is a valid task description with details.",
			expectError: false,
		},
		{
			name:        "empty description",
			description: "",
			expectError: false, // Empty descriptions are allowed
		},
		{
			name:        "description too long",
			description: strings.Repeat("A", 2001),
			expectError: true,
			errorMsg:    "description too long",
		},
		{
			name:        "description with HTML",
			description: "Description with <b>bold</b> text",
			expectError: true,
			errorMsg:    "contains HTML tags",
		},
		{
			name:        "max length description",
			description: strings.Repeat("A", 2000),
			expectError: false,
		},
		{
			name:        "multiline description",
			description: "Line 1\nLine 2\nLine 3",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTaskDescription(tt.description)
			
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateProjectTitle(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name        string
		title       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid project title",
			title:       "My Project",
			expectError: false,
		},
		{
			name:        "empty project title",
			title:       "",
			expectError: true,
			errorMsg:    "project title cannot be empty",
		},
		{
			name:        "project title too long",
			title:       strings.Repeat("A", 201),
			expectError: true,
			errorMsg:    "project title too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateProjectTitle(tt.title)
			
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateComplexity(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name        string
		complexity  int
		expectError bool
	}{
		{
			name:        "valid complexity 1",
			complexity:  1,
			expectError: false,
		},
		{
			name:        "valid complexity 5",
			complexity:  5,
			expectError: false,
		},
		{
			name:        "valid complexity 10",
			complexity:  10,
			expectError: false,
		},
		{
			name:        "complexity too low",
			complexity:  0,
			expectError: true,
		},
		{
			name:        "complexity too high",
			complexity:  11,
			expectError: true,
		},
		{
			name:        "negative complexity",
			complexity:  -1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateComplexity(tt.complexity)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "complexity must be between 1 and 10")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateActor(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name        string
		actor       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid actor",
			actor:       "john.doe",
			expectError: false,
		},
		{
			name:        "empty actor",
			actor:       "",
			expectError: true,
			errorMsg:    "actor cannot be empty",
		},
		{
			name:        "actor too long",
			actor:       strings.Repeat("A", 101),
			expectError: true,
			errorMsg:    "actor name too long",
		},
		{
			name:        "max length actor",
			actor:       strings.Repeat("A", 100),
			expectError: false,
		},
		{
			name:        "actor with special chars",
			actor:       "user@domain.com",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateActor(tt.actor)
			
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSanitizeContent(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "HTML entities",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "mixed content",
			input:    "Text with <b>bold</b> & special chars",
			expected: "Text with &lt;b&gt;bold&lt;/b&gt; &amp; special chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.SanitizeContent(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidatorWithHTMLAllowed(t *testing.T) {
	validator := NewInputValidator()
	validator.AllowHTML = true

	// Should not error on HTML when allowed
	err := validator.ValidateTaskTitle("Task with <b>bold</b> text")
	assert.NoError(t, err)

	// Sanitization should return original content when HTML is allowed
	input := "Text with <b>bold</b>"
	result := validator.SanitizeContent(input)
	assert.Equal(t, input, result)
}

func TestValidatorCustomLimits(t *testing.T) {
	validator := NewInputValidator()
	validator.MaxTitleLength = 50
	validator.MaxDescriptionLength = 100

	// Should error with custom limits
	err := validator.ValidateTaskTitle(strings.Repeat("A", 51))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title too long: 51 characters (max: 50)")

	err = validator.ValidateTaskDescription(strings.Repeat("A", 101))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description too long: 101 characters (max: 100)")
}