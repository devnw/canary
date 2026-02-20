// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.devnw.com/canary/internal/prompts"
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

// loadEmbeddedPrompt loads a prompt from embedded prompts (internal/prompts).
// Tries command prompts first (e.g. "scan", "list"), then system prompts ("init", "policy", "requirements", "evaluate").
func loadEmbeddedPrompt(name string) (string, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return "", fmt.Errorf("embedded prompt name is empty")
	}
	// Command prompts (prompts/commands/*/*.md)
	content, err := prompts.GetCommand(name)
	if err == nil {
		return content, nil
	}
	// System prompts (prompts/sys/*.md)
	all := prompts.All()
	if content, ok := all[name]; ok && content != "" {
		return content, nil
	}
	return "", fmt.Errorf("embedded prompt not found: %s", name)
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

// GetAvailablePrompts returns the list of available embedded prompt names
// (command prompts from prompts/commands/* and system prompts from prompts/sys/*).
func GetAvailablePrompts() ([]string, error) {
	commands, err := prompts.ListCommands()
	if err != nil {
		return nil, fmt.Errorf("list command prompts: %w", err)
	}
	seen := make(map[string]bool)
	for _, c := range commands {
		seen[c] = true
	}
	for k := range prompts.All() {
		if k != "" {
			seen[k] = true
		}
	}
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	sort.Strings(names)
	return names, nil
}
