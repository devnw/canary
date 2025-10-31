// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package utils

import (
	"fmt"
	"regexp"
	"strings"

	"go.devnw.com/canary/embedded"
	"go.devnw.com/canary/internal/config"
)

// FilterCanaryTokens removes CANARY tokens with OWNER=canary from file content
// This strips out CANARY CLI internal tracking tokens when copying templates to user projects
func FilterCanaryTokens(content []byte) []byte {
	lines := strings.Split(string(content), "\n")
	filtered := make([]string, 0, len(lines))

	for _, line := range lines {
		// Check if line contains a CANARY token with OWNER=canary
		if strings.Contains(line, "CANARY:") && strings.Contains(line, "OWNER=canary") {
			// Skip this line - it's a CANARY CLI internal token
			continue
		}
		filtered = append(filtered, line)
	}

	return []byte(strings.Join(filtered, "\n"))
}

// ReadEmbeddedFile safely reads a file from the embedded filesystem
// It tries with and without the "base/" prefix to handle different embed scenarios
func ReadEmbeddedFile(path string) ([]byte, error) {
	// Try the path as-is
	if content, err := embedded.CanaryFS.ReadFile(path); err == nil {
		return content, nil
	}

	// If the path starts with "base/", try without it
	if strings.HasPrefix(path, "base/") {
		trimmed := strings.TrimPrefix(path, "base/")
		if content, err := embedded.CanaryFS.ReadFile(trimmed); err == nil {
			return content, nil
		}
	}

	// If the path doesn't start with "base/", try with it
	if !strings.HasPrefix(path, "base/") {
		withBase := "base/" + path
		if content, err := embedded.CanaryFS.ReadFile(withBase); err == nil {
			return content, nil
		}
	}

	return nil, fmt.Errorf("file not found in embedded filesystem: %s", path)
}

// LoadProjectConfig loads the .canary/project.yaml configuration
func LoadProjectConfig() (*config.ProjectConfig, error) {
	return config.Load(".")
}

// ExtractField extracts a field value from a CANARY token string
func ExtractField(token, field string) string {
	// Look for FIELD="value" or FIELD=value
	pattern := field + `="([^"]+)"`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(token)
	if len(matches) > 1 {
		return matches[1]
	}

	// Try without quotes
	pattern = field + `=([^;\s]+)`
	re = regexp.MustCompile(pattern)
	matches = re.FindStringSubmatch(token)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}
