// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package specify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCANARY_CBIN_141_CLI_SpecifyRendersRealID verifies that `canary specify`
// substitutes the generated requirement ID into spec.md and leaves no
// unrendered Go template placeholders ("{{") behind. This guards the
// placeholder contract between specify.go's strings.ReplaceAll target and
// the literal ID placeholder used in base/templates/spec-template.md.
func TestCANARY_CBIN_141_CLI_SpecifyRendersRealID(t *testing.T) {
	tmpDir := t.TempDir()

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origWD); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})

	if err := os.MkdirAll(filepath.Join(tmpDir, ".canary", "specs"), 0750); err != nil {
		t.Fatalf("mkdir .canary/specs: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cmd := SpecifyCmd
	if err := cmd.Flags().Set("aspect", "API"); err != nil {
		t.Fatalf("set aspect flag: %v", err)
	}

	if err := cmd.RunE(cmd, []string{"a", "test", "feature"}); err != nil {
		t.Fatalf("SpecifyCmd.RunE: %v", err)
	}

	specsDir := filepath.Join(tmpDir, ".canary", "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		t.Fatalf("read specs dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 spec directory, got %d", len(entries))
	}

	generatedID := strings.SplitN(entries[0].Name(), "-a-test-feature", 2)[0]
	if generatedID == "" {
		t.Fatalf("could not derive generated ID from spec dir name %q", entries[0].Name())
	}
	if !strings.HasPrefix(generatedID, "CBIN-API-") {
		t.Fatalf("generated ID %q does not have expected CBIN-API- prefix", generatedID)
	}

	specFile := filepath.Join(specsDir, entries[0].Name(), "spec.md")
	content, err := os.ReadFile(specFile)
	if err != nil {
		t.Fatalf("read spec.md: %v", err)
	}

	body := string(content)

	if !strings.Contains(body, generatedID) {
		t.Errorf("spec.md does not contain the generated requirement ID %q", generatedID)
	}

	if strings.Contains(body, "{{") {
		t.Errorf("spec.md contains unrendered template syntax \"{{\":\n%s", body)
	}

	if strings.Contains(body, "REQID-") {
		t.Errorf("spec.md contains unrendered placeholder \"REQID-\"")
	}

	if strings.Contains(body, "SECURITY_REVIEW") {
		t.Errorf("spec.md contains unrendered placeholder \"SECURITY_REVIEW\"")
	}
}
