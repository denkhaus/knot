package completion

import (
	_ "embed"
	"fmt"
)

//go:embed zsh_completion.zsh
var zshCompletionScript string

// ZshCompletion handles zsh completion generation
type ZshCompletion struct{}

// NewZshCompletion creates a new zsh completion generator
func NewZshCompletion() *ZshCompletion {
	return &ZshCompletion{}
}

// Generate generates zsh completion script
func (z *ZshCompletion) Generate() error {
	_, err := fmt.Print(zshCompletionScript)
	return err
}

// LoadCompletionScript loads the zsh completion script from embedded fs
func LoadZshCompletionScript() (string, error) {
	return zshCompletionScript, nil
}