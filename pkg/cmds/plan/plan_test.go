// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-142; FEATURE="PlanTemplatePlaceholderContract"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_142_CLI_PlanRendersRealID; UPDATED=2026-08-29
package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCANARY_CBIN_142_CLI_PlanRendersRealID verifies that `canary plan`
// substitutes the requirement ID into plan.md and leaves no unrendered Go
// template placeholders ("{{") behind. This guards the placeholder contract
// between plan.go's strings.ReplaceAll target and the literal ID placeholder
// used in base/templates/plan-template.md.
func TestCANARY_CBIN_142_CLI_PlanRendersRealID(t *testing.T) {
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

	const reqID = "CBIN-API-105"
	specDir := filepath.Join(tmpDir, ".canary", "specs", reqID+"-user-auth")
	if err := os.MkdirAll(specDir, 0750); err != nil {
		t.Fatalf("mkdir spec dir: %v", err)
	}

	specContent := "# Feature Specification: UserAuth\n\n**Aspect:** [API|CLI|Engine|Storage|Security|Docs]\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0640); err != nil {
		t.Fatalf("write spec.md: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cmd := PlanCmd
	// Pass --aspect explicitly: the spec.md "**Aspect:**" line still carries
	// the template's bracketed placeholder list here, so we bypass the
	// spec-parsing fallback rather than depend on it.
	if err := cmd.Flags().Set("aspect", "API"); err != nil {
		t.Fatalf("set aspect flag: %v", err)
	}
	if err := cmd.RunE(cmd, []string{reqID}); err != nil {
		t.Fatalf("PlanCmd.RunE: %v", err)
	}

	planFile := filepath.Join(specDir, "plan.md")
	content, err := os.ReadFile(planFile)
	if err != nil {
		t.Fatalf("read plan.md: %v", err)
	}

	body := string(content)

	if !strings.Contains(body, reqID) {
		t.Errorf("plan.md does not contain the requirement ID %q", reqID)
	}

	if strings.Contains(body, "{{") {
		t.Errorf("plan.md contains unrendered template syntax \"{{\":\n%s", body)
	}
}
