// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadPrompt loads a custom prompt from file or embedded prompt name.
// This is a stub implementation that will be expanded in the future to support:
// - Loading prompts from embedded FS
// - Loading prompts from filesystem
// - Template variable substitution
// - Prompt validation and caching
func LoadPrompt(promptArg string) (string, error) {
	if promptArg == "" {
		return "", fmt.Errorf("no prompt specified")
	}

	// Check if it's a file path
	if strings.Contains(promptArg, "/") || strings.Contains(promptArg, "\\") {
		return loadPromptFromFile(promptArg)
	}

	// Otherwise, treat as embedded prompt name
	return loadEmbeddedPrompt(promptArg)
}

// loadPromptFromFile loads a prompt from a file path
func loadPromptFromFile(path string) (string, error) {
	// Expand relative paths
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	// Read file
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("read prompt file: %w", err)
	}

	return string(data), nil
}

// loadEmbeddedPrompt loads a prompt from embedded prompts
func loadEmbeddedPrompt(name string) (string, error) {
	// TODO: Implement embedded prompt loading
	// This would use the embedded FS to load prompts from:
	// - prompts/sys/*.md for system prompts
	// - prompts/commands/*.md for command-specific prompts
	// - .canary/templates/*.md for custom project prompts

	return "", fmt.Errorf("embedded prompt loading not yet implemented: %s", name)
}

// ValidatePromptArg validates a prompt argument format
func ValidatePromptArg(promptArg string) error {
	if promptArg == "" {
		return nil // Empty is valid (no custom prompt)
	}

	// Check if it looks like a file path
	if strings.Contains(promptArg, "/") || strings.Contains(promptArg, "\\") {
		// Validate file exists
		_, err := os.Stat(promptArg)
		if err != nil {
			return fmt.Errorf("prompt file not found: %s", promptArg)
		}
		return nil
	}

	// For embedded prompt names, just validate format
	// (actual validation happens during load)
	if len(promptArg) == 0 || len(promptArg) > 100 {
		return fmt.Errorf("invalid prompt name length")
	}

	return nil
}

// GetAvailablePrompts lists available embedded prompts
func GetAvailablePrompts() ([]string, error) {
	// TODO: Implement listing of embedded prompts
	// This would scan:
	// - embedded FS prompts directory
	// - .canary/templates/ directory
	// - prompts/ directory

	return []string{}, fmt.Errorf("prompt listing not yet implemented")
}
