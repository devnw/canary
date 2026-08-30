// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/evidence"
)

// fixtureDigest is a syntactically valid artifact digest used by fixture
// evidence records. Fixture requirements are not repo claims, so no real
// artifact is being attested here -- the digest only has to satisfy the
// strict record grammar.
const fixtureDigest = "sha256:" + "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

// writeEvidence writes an evidence file at path holding recs.
func writeEvidence(t *testing.T, path string, recs []evidence.Record) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(evidence.File{SchemaVersion: 1, Records: recs})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestAuditF15CLI covers the CLI half of F-15: one verification path, driven
// by evidence. Claims with current-commit PASS evidence verify; claims
// without it are named on stderr and nothing else; an empty claims file is a
// failure (EMPTY_CLAIMS) unless the caller explicitly allows it.
//
// (The scan-level half of F-15 -- the single token grammar -- is
// TestAuditF15 in audit_f01_f04_test.go.)
func TestAuditF15CLI(t *testing.T) {
	bin := buildCanary(t)
	root := t.TempDir()
	initGitRepo(t, root)
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(
		"// CANARY: REQ=CBIN-010; FEATURE=\"A\"; ASPECT=API; STATUS=TESTED; TEST=TestA; UPDATED=2026-01-01\n"+
			"// CANARY: REQ=CBIN-011; FEATURE=\"B\"; ASPECT=API; STATUS=TESTED; TEST=TestB; UPDATED=2026-01-01\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gap := filepath.Join(root, "GAP_ANALYSIS.md")
	if err := os.WriteFile(gap, []byte("✅ CBIN-010 - A\n✅ CBIN-011 - B\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommitAll(t, root)
	commit := headCommit(t, root)

	// Only CBIN-010 has passing evidence at HEAD.
	writeEvidence(t, filepath.Join(root, ".canary", "evidence.json"), []evidence.Record{{
		ProjectID: "default", RequirementID: "CBIN-010", Feature: "A", Aspect: "API",
		TestID: "TestA", Command: "go test ./...", Result: "PASS", CommitSHA: commit,
		ObservedAt: "2026-08-30T00:00:00Z", Runner: "local", ArtifactDigest: fixtureDigest,
	}})

	run := func(args ...string) (string, string, int) {
		t.Helper()
		cmd := exec.Command(bin, args...)
		cmd.Dir = root
		out, err := cmd.Output()
		code := 0
		stderr := ""
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
			stderr = string(ee.Stderr)
		} else if err != nil {
			t.Fatalf("run %v: %v", args, err)
		}
		return string(out), stderr, code
	}

	t.Run("only unverified claims are reported", func(t *testing.T) {
		out, stderr, code := run("verify", "--root", root, "--claims", gap, "--format", "json")
		if code != 1 {
			t.Fatalf("exit=%d stdout=%q stderr=%q", code, out, stderr)
		}
		want := `{"ok":false,"state":"UNVERIFIED","code":"EVIDENCE_MISSING","message":"no passing evidence at current commit"}` + "\n"
		if out != want {
			t.Fatalf("stdout:\n%q\nwant:\n%q", out, want)
		}
		if !strings.Contains(stderr, "CBIN-011") {
			t.Fatalf("stderr must name the unverified claim: %q", stderr)
		}
		if strings.Contains(stderr, "CBIN-010") {
			t.Fatalf("stderr must not name the verified claim: %q", stderr)
		}
	})

	t.Run("empty claims file fails closed", func(t *testing.T) {
		empty := filepath.Join(root, "EMPTY.md")
		if err := os.WriteFile(empty, []byte("# no claims here\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		out, stderr, code := run("verify", "--root", root, "--claims", empty, "--format", "json")
		if code != 1 {
			t.Fatalf("exit=%d stdout=%q stderr=%q", code, out, stderr)
		}
		want := `{"ok":false,"state":"UNVERIFIED","code":"EMPTY_CLAIMS","message":"no claims found"}` + "\n"
		if out != want {
			t.Fatalf("stdout:\n%q\nwant:\n%q", out, want)
		}
	})

	t.Run("allow-empty opts out", func(t *testing.T) {
		empty := filepath.Join(root, "EMPTY.md")
		out, stderr, code := run("verify", "--root", root, "--claims", empty, "--format", "json", "--allow-empty")
		if code != 0 {
			t.Fatalf("exit=%d stdout=%q stderr=%q", code, out, stderr)
		}
		want := `{"ok":true,"state":"VERIFIED","code":"OK","message":"no claims (allowed)"}` + "\n"
		if out != want {
			t.Fatalf("stdout:\n%q\nwant:\n%q", out, want)
		}
	})
}
