// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package storage

import "testing"

// TestIsHiddenPath verifies that hidden path detection works correctly
func TestIsHiddenPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Test files should be hidden
		{"test file go", "cmd/canary/main_test.go", true},
		{"test file in tests dir", "internal/tests/test.go", true},
		{"test dir slash", "/test/file.go", true},

		// Template files should be hidden
		{"canary templates", ".canary/templates/spec-template.md", true},
		{"base templates", "base/.canary/templates/plan.md", true},
		{"embedded base", "embedded/base/template.md", true},

		// Documentation examples should be hidden
		{"implementation summary", "IMPLEMENTATION_SUMMARY.md", true},
		{"final summary", "FINAL_SUMMARY.md", true},
		{"readme canary", "README_CANARY.md", true},
		{"gap analysis", "GAP_ANALYSIS.md", true},

		// AI agent directories should be hidden
		{"claude commands", ".claude/commands/canary.specify.md", true},
		{"cursor commands", ".cursor/commands/canary.plan.md", true},
		{"github prompts", ".github/prompts/canary-scan.md", true},
		{"windsurf workflows", ".windsurf/workflows/canary-verify.md", true},

		// Production code should NOT be hidden
		{"main go file", "cmd/canary/main.go", false},
		{"storage file", "internal/storage/storage.go", false},
		{"api file", "pkg/api/api.go", false},
		{"regular markdown", "docs/architecture.md", false},
		{"spec file", ".canary/specs/CBIN-105-feature/spec.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHiddenPath(tt.path)
			if result != tt.expected {
				t.Errorf("isHiddenPath(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}
