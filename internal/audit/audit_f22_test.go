// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestAuditF22 covers F-22: evidence is a strict, tamper-evident record.
// `canary evidence ingest` refuses anything that does not satisfy the record
// grammar -- duplicate fields, a short commit SHA, a non-PASS result, a
// non-UTC timestamp -- and only a clean file reaches the evidence store,
// after which `canary verify` reports VERIFIED.
func TestAuditF22(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()
	initGitRepo(t, root)
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(
		"// CANARY: REQ=CBIN-020; FEATURE=\"C\"; ASPECT=API; STATUS=TESTED; TEST=TestC; UPDATED=2026-01-01\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gap := filepath.Join(root, "GAP_ANALYSIS.md")
	if err := os.WriteFile(gap, []byte("✅ CBIN-020 - C\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommitAll(t, root)
	commit := headCommit(t, root)

	// record builds one evidence record body with the given commit,
	// result and observed_at, so each rejection case differs in exactly
	// one field from the accepted one.
	record := func(commit, result, observedAt string) string {
		return fmt.Sprintf(`{"project_id":"default","requirement_id":"CBIN-020","feature":"C","aspect":"API",`+
			`"test_id":"TestC","command":"go test ./...","result":%q,"commit_sha":%q,`+
			`"observed_at":%q,"runner":"local","artifact_digest":%q}`,
			result, commit, observedAt, fixtureDigest)
	}
	file := func(recordJSON string) string {
		return `{"schema_version":1,"records":[` + recordJSON + `]}`
	}

	out := filepath.Join(root, ".canary", "evidence.json")
	ingest := func(t *testing.T, name, content string) (string, int) {
		t.Helper()
		in := filepath.Join(root, name)
		if err := os.WriteFile(in, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(bin, "evidence", "ingest", "--in", in, "--out", out)
		cmd.Dir = root
		stdout, err := cmd.Output()
		code := 0
		stderr := ""
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
			stderr = string(ee.Stderr)
		} else if err != nil {
			t.Fatalf("run ingest: %v", err)
		}
		if len(stdout) != 0 {
			t.Fatalf("ingest wrote to stdout: %q", stdout)
		}
		return stderr, code
	}

	rejected := []struct {
		name    string
		content string
	}{
		{"dup.json", file(strings.Replace(record(commit, "PASS", "2026-08-30T00:00:00Z"),
			`"runner":"local"`, `"runner":"local","runner":"ci"`, 1))},
		{"shortsha.json", file(record(commit[:39], "PASS", "2026-08-30T00:00:00Z"))},
		{"fail.json", file(record(commit, "FAIL", "2026-08-30T00:00:00Z"))},
		{"nonutc.json", file(record(commit, "PASS", "2026-08-30T00:00:00-05:00"))},
	}
	for _, tc := range rejected {
		t.Run("rejects "+tc.name, func(t *testing.T) {
			stderr, code := ingest(t, tc.name, tc.content)
			if code != 1 {
				t.Fatalf("exit=%d, want 1 (stderr %q)", code, stderr)
			}
			if !strings.Contains(stderr, "evidence:") {
				t.Fatalf("stderr must carry the parse error: %q", stderr)
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Fatalf("rejected input must not create the evidence store")
			}
		})
	}

	t.Run("accepts the valid file", func(t *testing.T) {
		stderr, code := ingest(t, "good.json", file(record(commit, "PASS", "2026-08-30T00:00:00Z")))
		if code != 0 {
			t.Fatalf("exit=%d stderr=%q", code, stderr)
		}
		if _, err := os.Stat(out); err != nil {
			t.Fatalf("evidence store not written: %v", err)
		}
	})

	t.Run("verify passes on ingested evidence", func(t *testing.T) {
		cmd := exec.Command(bin, "verify", "--root", root, "--claims", gap,
			"--evidence", out, "--format", "json")
		cmd.Dir = root
		stdout, err := cmd.Output()
		if err != nil {
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				t.Fatalf("exit=%d stdout=%q stderr=%q", ee.ExitCode(), stdout, ee.Stderr)
			}
			t.Fatal(err)
		}
		want := `{"ok":true,"state":"VERIFIED","code":"OK","message":"all claims verified"}` + "\n"
		if string(stdout) != want {
			t.Fatalf("stdout:\n%q\nwant:\n%q", stdout, want)
		}
	})
}
