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