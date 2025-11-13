package validation

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

// InputValidator provides validation for user inputs
type InputValidator struct {
	MaxTitleLength       int
	MaxDescriptionLength int
	AllowHTML            bool
}

// NewInputValidator creates a new input validator with default limits
func NewInputValidator() *InputValidator {
	return &InputValidator{
		MaxTitleLength:       200,   // Reasonable limit for task titles
		MaxDescriptionLength: 2000,  // Reasonable limit for descriptions
		AllowHTML:            false, // Disable HTML by default for security
	}
}

// ValidateTaskTitle validates a task title
func (v *InputValidator) ValidateTaskTitle(title string) error {
	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	// Check length
	if utf8.RuneCountInString(title) > v.MaxTitleLength {
		return fmt.Errorf("title too long: %d characters (max: %d)",
			utf8.RuneCountInString(title), v.MaxTitleLength)
	}

	// Check for dangerous content
	if err := v.validateContent(title, "title"); err != nil {
		return err
	}

	return nil
}

// ValidateTaskDescription validates a task description
func (v *InputValidator) ValidateTaskDescription(description string) error {
	// Empty description is allowed
	if description == "" {
		return nil
	}

	// Check length
	if utf8.RuneCountInString(description) > v.MaxDescriptionLength {
		return fmt.Errorf("description too long: %d characters (max: %d)",
			utf8.RuneCountInString(description), v.MaxDescriptionLength)
	}

	// Check for dangerous content
	if err := v.validateContent(description, "description"); err != nil {
		return err
	}

	return nil
}

// ValidateProjectTitle validates a project title
func (v *InputValidator) ValidateProjectTitle(title string) error {
	if title == "" {
		return fmt.Errorf("project title cannot be empty")
	}

	// Use same limits as task title
	if utf8.RuneCountInString(title) > v.MaxTitleLength {
		return fmt.Errorf("project title too long: %d characters (max: %d)",
			utf8.RuneCountInString(title), v.MaxTitleLength)
	}

	// Check for dangerous content
	if err := v.validateContent(title, "project title"); err != nil {
		return err
	}

	return nil
}

// ValidateProjectDescription validates a project description
func (v *InputValidator) ValidateProjectDescription(description string) error {
	// Empty description is allowed
	if description == "" {
		return nil
	}

	// Use same limits as task description
	if utf8.RuneCountInString(description) > v.MaxDescriptionLength {
		return fmt.Errorf("project description too long: %d characters (max: %d)",
			utf8.RuneCountInString(description), v.MaxDescriptionLength)
	}

	// Check for dangerous content
	if err := v.validateContent(description, "project description"); err != nil {
		return err
	}

	return nil
}

// validateContent checks for potentially dangerous content
func (v *InputValidator) validateContent(content, fieldName string) error {
	// Check for null bytes (can cause issues in some contexts)
	if strings.Contains(content, "\x00") {
		return fmt.Errorf("%s contains null bytes", fieldName)
	}

	// Check for control characters (except common whitespace)
	for _, r := range content {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return fmt.Errorf("%s contains invalid control characters", fieldName)
		}
	}

	// If HTML is not allowed, check for HTML-like content
	if !v.AllowHTML {
		if err := v.checkForHTML(content, fieldName); err != nil {
			return err
		}
	}

	return nil
}

// checkForHTML detects and blocks HTML-like content
func (v *InputValidator) checkForHTML(content, fieldName string) error {
	// Simple HTML tag detection
	htmlTagPattern := regexp.MustCompile(`<[^>]*>`)
	if htmlTagPattern.MatchString(content) {
		return fmt.Errorf("%s contains HTML tags which are not allowed", fieldName)
	}

	// Check for script-like patterns
	scriptPatterns := []string{
		"javascript:",
		"data:",
		"vbscript:",
		"onload=",
		"onerror=",
		"onclick=",
	}

	lowerContent := strings.ToLower(content)
	for _, pattern := range scriptPatterns {
		if strings.Contains(lowerContent, pattern) {
			return fmt.Errorf("%s contains potentially dangerous script content", fieldName)
		}
	}

	return nil
}

// SanitizeContent sanitizes content by escaping HTML if needed
func (v *InputValidator) SanitizeContent(content string) string {
	if v.AllowHTML {
		return content
	}

	// Escape HTML entities
	return html.EscapeString(content)
}

// ValidateComplexity validates task complexity value
func (v *InputValidator) ValidateComplexity(complexity int) error {
	if complexity < 1 || complexity > 10 {
		return fmt.Errorf("complexity must be between 1 and 10, got %d", complexity)
	}
	return nil
}

// ValidateTaskPriority validates task priority value
func (v *InputValidator) ValidateTaskPriority(priority string) error {
	validPriorities := []string{"low", "medium", "high"}
	for _, valid := range validPriorities {
		if priority == valid {
			return nil
		}
	}
	return fmt.Errorf("priority must be one of: %s, got %q", strings.Join(validPriorities, ", "), priority)
}

// ValidateActor validates actor name
func (v *InputValidator) ValidateActor(actor string) error {
	if actor == "" {
		return fmt.Errorf("actor cannot be empty")
	}

	// Reasonable limit for actor names
	if utf8.RuneCountInString(actor) > 100 {
		return fmt.Errorf("actor name too long: %d characters (max: 100)",
			utf8.RuneCountInString(actor))
	}

	// Check for dangerous content
	if err := v.validateContent(actor, "actor name"); err != nil {
		return err
	}

	return nil
}
