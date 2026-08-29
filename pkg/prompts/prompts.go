// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package prompts

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// CANARY: REQ=CP-265; FEATURE="HierarchicalPrompts"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-11-01

//go:embed sys/*.md
var sysFS embed.FS

//go:embed commands/*/*.md
var commandsFS embed.FS

// System prompts (legacy - kept for compatibility)
var (
	Init         string
	Policy       string
	Requirements string
	Evaluate     string
)

func init() {
	// Load system prompts
	if content, err := fs.ReadFile(sysFS, "sys/init.md"); err == nil {
		Init = string(content)
	}
	if content, err := fs.ReadFile(sysFS, "sys/policy.md"); err == nil {
		Policy = string(content)
	}
	if content, err := fs.ReadFile(sysFS, "sys/requirements.md"); err == nil {
		Requirements = string(content)
	}
	if content, err := fs.ReadFile(sysFS, "sys/evaluate.md"); err == nil {
		Evaluate = string(content)
	}
}

// All returns all system prompts (legacy function)
func All() map[string]string {
	return map[string]string{
		"init":         Init,
		"policy":       Policy,
		"requirements": Requirements,
		"evaluate":     Evaluate,
	}
}

// GetCommand returns the prompt for a specific command
func GetCommand(command string) (string, error) {
	path := filepath.Join("commands", command, command+".md")
	content, err := fs.ReadFile(commandsFS, path)
	if err != nil {
		return "", fmt.Errorf("prompt not found for command %q: %w", command, err)
	}
	return string(content), nil
}

// ListCommands returns all available command prompts
func ListCommands() ([]string, error) {
	var commands []string

	err := fs.WalkDir(commandsFS, "commands", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && path != "commands" {
			// Extract command name from path
			parts := strings.Split(path, "/")
			if len(parts) >= 2 {
				commands = append(commands, parts[1])
			}
		}

		return nil
	})

	return commands, err
}

// GetAllCommands returns all command prompts as a map
func GetAllCommands() (map[string]string, error) {
	result := make(map[string]string)

	commands, err := ListCommands()
	if err != nil {
		return nil, err
	}

	for _, cmd := range commands {
		content, err := GetCommand(cmd)
		if err != nil {
			// Skip commands without prompts
			continue
		}
		result[cmd] = content
	}

	return result, nil
}

// CommandPrompt represents a structured command prompt
type CommandPrompt struct {
	Command     string
	Purpose     string
	Behavior    string
	Examples    []string
	Standards   string
	FullContent string
}

// ParseCommandPrompt parses a command prompt into structured fields
// This is a simple parser - could be enhanced to extract specific sections
func ParseCommandPrompt(command string) (*CommandPrompt, error) {
	content, err := GetCommand(command)
	if err != nil {
		return nil, err
	}

	return &CommandPrompt{
		Command:     command,
		FullContent: content,
		// Additional parsing could be added here
	}, nil
}
