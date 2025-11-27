// Package completion provides shell completion functionality for KNOT CLI.
//
// This package contains shell completion scripts and functionality for different
// shell environments including bash and zsh, enabling tab completion for
// KNOT commands and options.
//
// Key Components:
//   - Bash Completion: Bash shell completion scripts and integration
//   - ZSH Completion: ZSH shell completion scripts and integration
//   - CLI Integration: Completion registration and setup
//
// Cross-reference: Knot Task 86f3ba2d-3a87-493b-b8fc-96d19f344e89
package completion

import (
	_ "embed"
	"fmt"
)

//go:embed bash_completion.bash
var bashCompletionScript string

// BashCompletion handles bash completion generation
type BashCompletion struct{}

// NewBashCompletion creates a new bash completion generator
func NewBashCompletion() *BashCompletion {
	return &BashCompletion{}
}

// Generate generates bash completion script
func (b *BashCompletion) Generate() error {
	_, err := fmt.Print(bashCompletionScript)
	return err
}

// LoadCompletionScript loads the bash completion script from embedded fs
func LoadBashCompletionScript() (string, error) {
	return bashCompletionScript, nil
}
