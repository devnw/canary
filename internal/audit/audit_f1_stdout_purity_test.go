// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestAuditF1SearchJSONStdoutPurity covers F1: `canary search <nomatch>
// --json` must put ONLY valid JSON on stdout, even when there are zero
// results. The pre-fix implementation printed "No tokens found for: ..."
// straight to stdout before ever consulting the --json flag, so a machine
// caller parsing stdout as JSON would fail regardless of exit code.
func TestAuditF1SearchJSONStdoutPurity(t *testing.T) {
	root := indexedEmptyRepo(t)
	bin := buildCanary(t)

	// search's human-mode "no tokens found" case exits 0; that convention
	// must be preserved in --json mode too (only the stdout SHAPE was wrong).
	out := runStdout(t, root, bin, "search", "zzz-no-such-keyword-zzz", "--json")

	var tokens []map[string]any
	if err := json.Unmarshal([]byte(out), &tokens); err != nil {
		t.Fatalf("search --json zero-result stdout is not valid JSON: %v\nstdout: %q", err, out)
	}
	if len(tokens) != 0 {
		t.Fatalf("expected empty array, got %d entries", len(tokens))
	}
	if strings.Contains(out, "No tokens found") {
		t.Fatalf("prose leaked onto stdout: %q", out)
	}
}

// TestAuditF1ShowJSONStdoutPurity covers F1: `canary show <missing> --json`
// must put ONLY valid JSON on stdout. show's human-mode not-found case exits
// 1 (returns an error) -- that exit convention must be preserved, but the
// "No tokens found" / "Suggestions" prose that used to print unconditionally
// before the --json check must move to stderr, and stdout must still decode
// as JSON.
func TestAuditF1ShowJSONStdoutPurity(t *testing.T) {
	root := indexedEmptyRepo(t)
	bin := buildCanary(t)

	cmd := exec.Command(bin, "show", "CBIN-NOSUCHREQ", "--json")
	cmd.Dir = root
	home := t.TempDir()
	cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err == nil {
		t.Fatalf("expected non-zero exit for missing requirement (human-mode convention), got success\nstdout: %s", stdout.String())
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Fatalf("expected exit code 1 (existing human-mode convention), got %d", exitErr.ExitCode())
		}
	} else {
		t.Fatalf("unexpected error type: %v", err)
	}

	out := stdout.String()
	if strings.Contains(out, "No tokens found") || strings.Contains(out, "Suggestions") {
		t.Fatalf("prose leaked onto stdout: %q", out)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		t.Fatal("expected a valid JSON document on stdout, got nothing")
	}
	var decoded any
	if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
		t.Fatalf("show --json not-found stdout is not valid JSON: %v\nstdout: %q", err, out)
	}
}
