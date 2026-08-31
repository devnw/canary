// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestAuditF24 covers F-24: the README must be honest. It must not carry
// unbacked numeric performance claims (no benchmark artifacts exist to support
// them), must not reference package paths that no longer exist, must not point
// at a repository/CI surface that is not real, and must document the MCP
// server it actually ships.
func TestAuditF24(t *testing.T) {
	b, err := os.ReadFile(repoPath(t, "README.md"))
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	s := string(b)

	// No unbacked numeric performance claims. These regexes match the shapes
	// of the deleted claims ("<209ms for 500", "<10s for 50k", "<50ms",
	// "≤512 MiB").
	perfClaims := []*regexp.Regexp{
		regexp.MustCompile(`\d+\s*(ms|s)\s+for\s+\d+`),
		regexp.MustCompile(`[<≤]\s*\d+\s*ms`),
		regexp.MustCompile(`[<≤]\s*\d+\s*s\b`),
		regexp.MustCompile(`[≤<]\s*\d+\s*MiB`),
		regexp.MustCompile(`\d+\s*MiB\s+RSS`),
	}
	for _, re := range perfClaims {
		if loc := re.FindString(s); loc != "" {
			t.Errorf("README carries unbacked performance claim matching %q: %q", re, loc)
		}
	}

	// No references to package paths that no longer exist.
	for _, dead := range []string{"internal/specs", "internal/storage", "cmd/canary/deps.go"} {
		if strings.Contains(s, dead) {
			t.Errorf("README references stale path %q", dead)
		}
	}

	// No clone URL against a repository surface the project does not publish
	// from (CI is GitLab; the GitHub Actions build badge was removed).
	if strings.Contains(s, "github.com/devnw/canary.git") {
		t.Error("README references github.com/devnw/canary.git")
	}
	if strings.Contains(s, "actions/workflows") {
		t.Error("README still carries a GitHub Actions badge/link")
	}

	// The Go badge must claim the real toolchain (1.27).
	if strings.Contains(s, "Go-1.25") || strings.Contains(s, "Go-1.24") {
		t.Error("README Go badge is not 1.27")
	}

	// An MCP section must exist and point at the generated tool list.
	if !regexp.MustCompile(`(?m)^#{1,3}\s+.*MCP`).MatchString(s) {
		t.Error("README has no MCP heading")
	}
	if !strings.Contains(s, "docs/MCP_TOOLS.md") {
		t.Error("README MCP section does not reference the generated docs/MCP_TOOLS.md")
	}
}
