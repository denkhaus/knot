package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInputValidator(t *testing.T) {
	validator := NewInputValidator()
	
	assert.Equal(t, 200, validator.MaxTitleLength)
	assert.Equal(t, 2000, validator.MaxDescriptionLength)
	assert.False(t, validator.AllowHTML)
}

func TestInputValidatorValidateTaskTitle(t *testing.T) {
	validator := NewInputValidator()
	
	tests := []struct {
		name        string
		title       string
		expectError bool
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
		},
		{
			name:        "very long title",
			title:       "a" + "b" + "c", // Create a string longer than 200 chars
			expectError: true,
		},
		{
			name:        "title with null bytes",
			title:       "test\x00title",
			expectError: true,
		},
		{
			name:        "title with HTML",
			title:       "test <script>alert('xss')</script>",
			expectError: true,
		},
	}

	// Create a very long title for the test
	longTitle := make([]byte, 201)
	for i := range longTitle {
		longTitle[i] = 'a'
	}

	// Run the tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var title string
			if tt.name == "very long title" {
				title = string(longTitle)
			} else {
				title = tt.title
			}
			
			err := validator.ValidateTaskTitle(title)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputValidatorValidateTaskDescription(t *testing.T) {
	validator := NewInputValidator()
	
	tests := []struct {
		name        string
		description string
		expectError bool
	}{
		{
			name:        "valid description",
			description: "Valid task description",
			expectError: false,
		},
		{
			name:        "empty description",
			description: "",
			expectError: false, // Empty description is allowed
		},
		{
			name:        "very long description",
			description: "",
			expectError: true,
		},
		{
			name:        "description with null bytes",
			description: "test\x00description",
			expectError: true,
		},
		{
			name:        "description with HTML",
			description: "test <script>alert('xss')</script>",
			expectError: true,
		},
	}

	// Create a very long description for the test
	longDescription := make([]byte, 2001)
	for i := range longDescription {
		longDescription[i] = 'a'
	}

	// Run the tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var description string
			if tt.name == "very long description" {
				description = string(longDescription)
			} else {
				description = tt.description
			}
			
			err := validator.ValidateTaskDescription(description)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputValidatorValidateProjectTitle(t *testing.T) {
	validator := NewInputValidator()
	
	longTitle := make([]byte, 201)
	for i := range longTitle {
		longTitle[i] = 'a'
	}

	tests := []struct {
		name        string
		title       string
		expectError bool
	}{
		{
			name:        "valid project title",
			title:       "Valid Project Title",
			expectError: false,
		},
		{
			name:        "empty project title",
			title:       "",
			expectError: true,
		},
		{
			name:        "very long project title",
			title:       string(longTitle),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateProjectTitle(tt.title)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputValidatorValidateProjectDescription(t *testing.T) {
	validator := NewInputValidator()
	
	longDescription := make([]byte, 2001)
	for i := range longDescription {
		longDescription[i] = 'a'
	}

	tests := []struct {
		name        string
		description string
		expectError bool
	}{
		{
			name:        "valid project description",
			description: "Valid project description",
			expectError: false,
		},
		{
			name:        "empty project description",
			description: "",
			expectError: false, // Empty description is allowed
		},
		{
			name:        "very long project description",
			description: string(longDescription),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateProjectDescription(tt.description)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputValidatorValidateContent(t *testing.T) {
	validator := NewInputValidator()
	
	tests := []struct {
		name        string
		content     string
		fieldName   string
		expectError bool
	}{
		{
			name:        "valid content",
			content:     "Valid content",
			fieldName:   "test field",
			expectError: false,
		},
		{
			name:        "content with null bytes",
			content:     "test\x00content",
			fieldName:   "test field",
			expectError: true,
		},
		{
			name:        "content with control characters",
			content:     "test\x01content",
			fieldName:   "test field",
			expectError: true,
		},
		{
			name:        "content with allowed whitespace",
			content:     "test\t\n\rcontent",
			fieldName:   "test field",
			expectError: false,
		},
		{
			name:        "content with HTML when not allowed",
			content:     "test <div>html</div>",
			fieldName:   "test field",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateContent(tt.content, tt.fieldName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputValidatorCheckForHTML(t *testing.T) {
	validator := NewInputValidator()
	validator.AllowHTML = false // Default, but being explicit
	
	htmlTests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:        "valid text without HTML",
			content:     "This is plain text",
			expectError: false,
		},
		{
			name:        "HTML tag",
			content:     "This has <div>HTML</div>",
			expectError: true,
		},
		{
			name:        "HTML self-closing tag",
			content:     "This has <img />",
			expectError: true,
		},
		{
			name:        "javascript in content",
			content:     "javascript:alert(1)",
			expectError: true,
		},
		{
			name:        "javascript capitalized",
			content:     "JAVASCRIPT:alert(1)",
			expectError: true,
		},
		{
			name:        "onload event",
			content:     "has onload= attribute",
			expectError: true,
		},
	}

	for _, tt := range htmlTests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.checkForHTML(tt.content, "test field")
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputValidatorSanitizeContent(t *testing.T) {
	validator := NewInputValidator()
	
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "plain text",
			content:  "Plain text content",
			expected: "Plain text content",
		},
		{
			name:     "content with HTML entities",
			content:  "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "content with quotes",
			content:  `Text with "quotes" and 'apostrophes'`,
			expected: `Text with &#34;quotes&#34; and &#39;apostrophes&#39;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.SanitizeContent(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInputValidatorSanitizeContentWithHTMLAllowed(t *testing.T) {
	validator := NewInputValidator()
	validator.AllowHTML = true
	
	content := "<div>HTML content</div>"
	result := validator.SanitizeContent(content)
	
	// When HTML is allowed, content should be returned as-is
	assert.Equal(t, content, result)
}

func TestInputValidatorValidateComplexity(t *testing.T) {
	validator := NewInputValidator()
	
	tests := []struct {
		name        string
		complexity  int
		expectError bool
	}{
		{
			name:        "valid complexity low",
			complexity:  1,
			expectError: false,
		},
		{
			name:        "valid complexity high",
			complexity:  10,
			expectError: false,
		},
		{
			name:        "valid complexity middle",
			complexity:  5,
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
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputValidatorValidateActor(t *testing.T) {
	validator := NewInputValidator()
	
	longActor := make([]byte, 101)
	for i := range longActor {
		longActor[i] = 'a'
	}

	tests := []struct {
		name        string
		actor       string
		expectError bool
	}{
		{
			name:        "valid actor",
			actor:       "john_doe",
			expectError: false,
		},
		{
			name:        "empty actor",
			actor:       "",
			expectError: true,
		},
		{
			name:        "very long actor",
			actor:       string(longActor),
			expectError: true,
		},
		{
			name:        "actor with null bytes",
			actor:       "test\x00actor",
			expectError: true,
		},
		{
			name:        "actor with control characters",
			actor:       "test\x01actor",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateActor(tt.actor)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputValidatorValidateTaskPriority(t *testing.T) {
	validator := NewInputValidator()
	
	tests := []struct {
		name        string
		priority    string
		expectError bool
	}{
		{
			name:        "valid low priority",
			priority:    "low",
			expectError: false,
		},
		{
			name:        "valid medium priority",
			priority:    "medium",
			expectError: false,
		},
		{
			name:        "valid high priority",
			priority:    "high",
			expectError: false,
		},
		{
			name:        "invalid priority",
			priority:    "invalid",
			expectError: true,
		},
		{
			name:        "case sensitive invalid",
			priority:    "LOW",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTaskPriority(tt.priority)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputValidatorWithCustomLimits(t *testing.T) {
	validator := &InputValidator{
		MaxTitleLength:       50,
		MaxDescriptionLength: 100,
		AllowHTML:           false,
	}
	
	// Test with custom length limits
	longTitle := make([]byte, 51)
	for i := range longTitle {
		longTitle[i] = 'a'
	}
	
	err := validator.ValidateTaskTitle(string(longTitle))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title too long")
	
	// Test with custom description limit
	longDesc := make([]byte, 101)
	for i := range longDesc {
		longDesc[i] = 'a'
	}
	
	err = validator.ValidateTaskDescription(string(longDesc))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description too long")
}

func TestInputValidatorValidateContentWithHTMLAllowed(t *testing.T) {
	validator := NewInputValidator()
	validator.AllowHTML = true // Allow HTML for this test
	
	// Should not return error when HTML is allowed
	err := validator.validateContent("test <div>html</div>", "test field")
	assert.NoError(t, err)
}

func TestInputValidatorValidationCombinations(t *testing.T) {
	validator := NewInputValidator()
	
	// Test valid combinations
	err := validator.ValidateTaskTitle("Valid Title")
	assert.NoError(t, err)
	
	err = validator.ValidateTaskDescription("Valid Description")
	assert.NoError(t, err)
	
	err = validator.ValidateProjectTitle("Valid Project Title")
	assert.NoError(t, err)
	
	err = validator.ValidateProjectDescription("Valid Project Description")
	assert.NoError(t, err)
	
	err = validator.ValidateComplexity(5)
	assert.NoError(t, err)
	
	err = validator.ValidateActor("valid_actor")
	assert.NoError(t, err)
	
	err = validator.ValidateTaskPriority("medium")
	assert.NoError(t, err)
}

func TestInputValidatorValidateContentWithVariousControlChars(t *testing.T) {
	validator := NewInputValidator()
	
	// Test with various control characters that should be rejected
	controlChars := []byte{1, 2, 3, 4, 5, 6, 7, 8, 11, 12, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	
	for i, char := range controlChars {
		t.Run("control_char_"+string(rune(i)), func(t *testing.T) {
			content := "test" + string(char) + "content"
			err := validator.validateContent(content, "test field")
			assert.Error(t, err)
		})
	}
	
	// Test allowed control characters (tab, newline, carriage return)
	allowedContent := "test\t\n\rontent" // tab, newline, carriage return
	err := validator.validateContent(allowedContent, "test field")
	assert.NoError(t, err)
}

// Benchmark tests to ensure validation is performant
func BenchmarkValidateTaskTitle(b *testing.B) {
	validator := NewInputValidator()
	
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateTaskTitle("Test Title")
	}
}

func BenchmarkValidateTaskDescription(b *testing.B) {
	validator := NewInputValidator()
	
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateTaskDescription("Test Description")
	}
}

func BenchmarkValidateComplexity(b *testing.B) {
	validator := NewInputValidator()
	
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateComplexity(5)
	}
}