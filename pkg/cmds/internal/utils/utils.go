// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"devnw.dev/canary/pkg/config"
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
	// Try reading directly from filesystem paths (development mode)
	candidates := []string{
		path,
		filepath.Join("pkg", "cmds", "init", path),
	}

	// Try variants with/without base/ prefix
	if strings.HasPrefix(path, "base/") {
		trimmed := strings.TrimPrefix(path, "base/")
		candidates = append(candidates, trimmed)
		candidates = append(candidates, filepath.Join("pkg", "cmds", "init", trimmed))
	} else {
		withBase := filepath.Join("base", path)
		candidates = append(candidates, withBase)
		candidates = append(candidates, filepath.Join("pkg", "cmds", "init", withBase))
	}

	for _, c := range candidates {
		if content, err := os.ReadFile(c); err == nil {
			return content, nil
		}
	}

	// Attempt to resolve relative to this source file's directory (helps when tests run from package dirs)
	if _, file, _, ok := runtime.Caller(0); ok {
		srcDir := filepath.Dir(file)
		// Walk up to repo root by looking for go.mod
		repoDir := srcDir
		for i := 0; i < 6; i++ {
			if _, err := os.Stat(filepath.Join(repoDir, "go.mod")); err == nil {
				break
			}
			repoDir = filepath.Join(repoDir, "..")
		}
		// Try repo-root based candidate
		repoCandidate := filepath.Join(repoDir, path)
		if content, err := os.ReadFile(repoCandidate); err == nil {
			return content, nil
		}
		// Try package-relative candidate (pkg/cmds/init/...)
		pkgCandidate := filepath.Join(repoDir, "pkg", "cmds", "init", path)
		if content, err := os.ReadFile(pkgCandidate); err == nil {
			return content, nil
		}
	}

	return nil, fmt.Errorf("file not found in embedded filesystem or disk: %s", path)
}

// LoadProjectConfig loads the .canary/project.yaml configuration
func LoadProjectConfig() (*config.ProjectConfig, error) {
	return config.Load(".")
}

// CANARY: REQ=CBIN-205; FEATURE="ContextCaps"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_205_EffectiveLimit; UPDATED=2026-08-28
// EffectiveLimit maps CLI --limit semantics (0/unset => def, -1 => unlimited)
// to the storage layer's convention (0 => unlimited). Defaults are
// deliberately small to protect agent context; callers pass -1 to
// explicitly request everything.
func EffectiveLimit(flag, def int) int {
	switch {
	case flag < 0:
		return 0
	case flag == 0:
		return def
	default:
		return flag
	}
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
