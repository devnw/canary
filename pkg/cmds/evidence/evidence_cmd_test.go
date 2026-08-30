// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package evidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ev "devnw.dev/canary/pkg/evidence"
)

const (
	testCommit    = "0123456789abcdef0123456789abcdef01234567"
	fixtureDigest = "sha256:" + "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
)

// recordJSON renders one evidence record, letting each test vary exactly the
// field it is about.
func recordJSON(result, commit, observedAt string) string {
	return fmt.Sprintf(`{"project_id":"default","requirement_id":"CBIN-020","feature":"C","aspect":"API",`+
		`"test_id":"TestC","command":"go test ./...","result":%q,"commit_sha":%q,`+
		`"observed_at":%q,"runner":"local","artifact_digest":%q}`,
		result, commit, observedAt, fixtureDigest)
}

func fileJSON(records ...string) string {
	return `{"schema_version":1,"records":[` + strings.Join(records, ",") + `]}`
}

// writeTemp writes body to a new file under dir and returns its path.
func writeTemp(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestIngest_RejectsMalformed proves every record-grammar violation rejects
// the whole file and writes nothing.
func TestIngest_RejectsMalformed(t *testing.T) {
	valid := recordJSON("PASS", testCommit, "2026-08-30T00:00:00Z")
	cases := map[string]string{
		"duplicate field":  fileJSON(strings.Replace(valid, `"runner":"local"`, `"runner":"local","runner":"ci"`, 1)),
		"short commit":     fileJSON(recordJSON("PASS", testCommit[:39], "2026-08-30T00:00:00Z")),
		"failing result":   fileJSON(recordJSON("FAIL", testCommit, "2026-08-30T00:00:00Z")),
		"non-utc time":     fileJSON(recordJSON("PASS", testCommit, "2026-08-30T00:00:00-05:00")),
		"unknown field":    fileJSON(strings.Replace(valid, `"runner":"local"`, `"runner":"local","extra":1`, 1)),
		"bad schema":       `{"schema_version":2,"records":[]}`,
		"empty test_id":    fileJSON(strings.Replace(valid, `"test_id":"TestC"`, `"test_id":""`, 1)),
		"malformed digest": fileJSON(strings.Replace(valid, fixtureDigest, "sha256:nope", 1)),
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			in := writeTemp(t, dir, "in.json", body)
			out := filepath.Join(dir, "store.json")
			var stderr bytes.Buffer
			if code := RunIngest(in, out, &stderr); code != 1 {
				t.Fatalf("exit=%d, want 1", code)
			}
			if !strings.Contains(stderr.String(), "evidence:") {
				t.Errorf("stderr must carry the parse error: %q", stderr.String())
			}
			if _, err := os.Stat(out); !os.IsNotExist(err) {
				t.Errorf("a rejected file must not create the store")
			}
		})
	}
}

// TestIngest_MergesAndDedupes proves ingest accumulates: an identical record
// is not duplicated, a different one is appended, and the store stays
// parseable.
func TestIngest_MergesAndDedupes(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, ".canary", "evidence.json")
	first := recordJSON("PASS", testCommit, "2026-08-30T00:00:00Z")
	second := strings.Replace(first, `"requirement_id":"CBIN-020"`, `"requirement_id":"CBIN-021"`, 1)

	var stderr bytes.Buffer
	if code := RunIngest(writeTemp(t, dir, "a.json", fileJSON(first)), out, &stderr); code != 0 {
		t.Fatalf("first ingest exit=%d stderr=%q", code, stderr.String())
	}
	// The same record again: merged, not duplicated.
	if code := RunIngest(writeTemp(t, dir, "b.json", fileJSON(first)), out, &stderr); code != 0 {
		t.Fatalf("second ingest exit=%d stderr=%q", code, stderr.String())
	}
	if code := RunIngest(writeTemp(t, dir, "c.json", fileJSON(first, second)), out, &stderr); code != 0 {
		t.Fatalf("third ingest exit=%d stderr=%q", code, stderr.String())
	}

	stored, err := ev.Load(out)
	if err != nil {
		t.Fatalf("store must stay strictly parseable: %v", err)
	}
	if len(stored.Records) != 2 {
		t.Fatalf("records = %d, want 2 (dedupe by exact record)", len(stored.Records))
	}
	if stored.Records[0].RequirementID != "CBIN-020" || stored.Records[1].RequirementID != "CBIN-021" {
		t.Errorf("existing order must be preserved: %+v", stored.Records)
	}
}

// TestIngest_MalformedStoreIsNeverOverwritten proves an unparseable store is
// left alone rather than replaced -- the record trail is not destroyed by a
// later ingest.
func TestIngest_MalformedStoreIsNeverOverwritten(t *testing.T) {
	dir := t.TempDir()
	out := writeTemp(t, dir, "store.json", `{"schema_version":1,"records":[{`)
	in := writeTemp(t, dir, "in.json", fileJSON(recordJSON("PASS", testCommit, "2026-08-30T00:00:00Z")))

	var stderr bytes.Buffer
	if code := RunIngest(in, out, &stderr); code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"schema_version":1,"records":[{` {
		t.Errorf("malformed store was modified: %q", body)
	}
}

// goTestStream renders a minimal `go test -json` stream.
func goTestStream(events ...string) string {
	return strings.Join(events, "\n") + "\n"
}

// TestFromGoTest_MapsPassingTests proves only tests that actually passed
// produce records, that each record binds to the given commit, and that the
// artifact digest is the SHA-256 of the raw stream.
func TestFromGoTest_MapsPassingTests(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte(
		"// CANARY: REQ=CBIN-001; FEATURE=\"F\"; ASPECT=API; STATUS=TESTED; TEST=TestPasses; UPDATED=2026-01-01\n"+
			"// CANARY: REQ=CBIN-002; FEATURE=\"G\"; ASPECT=API; STATUS=TESTED; TEST=TestFails; UPDATED=2026-01-01\n"+
			"// CANARY: REQ=CBIN-003; FEATURE=\"H\"; ASPECT=API; STATUS=IMPL; TEST=TestNeverRan; UPDATED=2026-01-01\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stream := goTestStream(
		`{"Action":"run","Package":"p","Test":"TestPasses"}`,
		`{"Action":"pass","Package":"p","Test":"TestPasses"}`,
		`{"Action":"fail","Package":"p","Test":"TestFails"}`,
		`{"Action":"pass","Package":"p"}`,
		`not a json line at all`,
	)

	var stdout, stderr bytes.Buffer
	code := RunFromGoTest(FromGoTestOptions{
		Root: root, ProjectID: "default", Commit: testCommit,
		ObservedAt: "2026-08-30T00:00:00Z",
	}, strings.NewReader(stream), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit=%d stderr=%q", code, stderr.String())
	}

	f, err := ev.Parse(bytes.NewReader(stdout.Bytes()))
	if err != nil {
		t.Fatalf("emitted file must be strictly parseable: %v\n%s", err, stdout.String())
	}
	if len(f.Records) != 1 {
		t.Fatalf("records = %+v, want exactly the passing test's", f.Records)
	}
	rec := f.Records[0]
	if rec.RequirementID != "CBIN-001" || rec.TestID != "TestPasses" || rec.Aspect != "API" {
		t.Errorf("record = %+v", rec)
	}
	if rec.CommitSHA != testCommit || rec.ObservedAt != "2026-08-30T00:00:00Z" || rec.Runner != "local" {
		t.Errorf("record = %+v", rec)
	}
	sum := sha256.Sum256([]byte(stream))
	if want := "sha256:" + hex.EncodeToString(sum[:]); rec.ArtifactDigest != want {
		t.Errorf("digest = %q, want the SHA-256 of the raw stream %q", rec.ArtifactDigest, want)
	}
	if !strings.Contains(rec.Command, "TestPasses") {
		t.Errorf("command must reproduce the test: %q", rec.Command)
	}
}

// TestFromGoTest_RejectsShortCommit proves an abbreviated SHA is refused:
// evidence binds to an exact commit.
func TestFromGoTest_RejectsShortCommit(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunFromGoTest(FromGoTestOptions{Root: t.TempDir(), Commit: testCommit[:12]},
		strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit=%d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("nothing may be emitted on rejection: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "40 lowercase hex") {
		t.Errorf("stderr = %q", stderr.String())
	}
}

// TestFromGoTest_ObservedAtFlag proves --observed-at is honored, that
// CANARY_TEST_TIMESTAMP is the fallback, and that a non-UTC value is refused.
func TestFromGoTest_ObservedAtFlag(t *testing.T) {
	t.Run("flag wins", func(t *testing.T) {
		t.Setenv("CANARY_TEST_TIMESTAMP", "2020-01-01T00:00:00Z")
		got, err := resolveObservedAt("2026-08-30T12:00:00Z")
		if err != nil || got != "2026-08-30T12:00:00Z" {
			t.Fatalf("got %q err %v", got, err)
		}
	})
	t.Run("env fallback", func(t *testing.T) {
		t.Setenv("CANARY_TEST_TIMESTAMP", "2020-01-01T00:00:00Z")
		got, err := resolveObservedAt("")
		if err != nil || got != "2020-01-01T00:00:00Z" {
			t.Fatalf("got %q err %v", got, err)
		}
	})
	t.Run("non-utc refused", func(t *testing.T) {
		if _, err := resolveObservedAt("2026-08-30T12:00:00-05:00"); err == nil {
			t.Fatal("a non-UTC observed_at must be refused")
		}
	})
	t.Run("unparseable refused", func(t *testing.T) {
		if _, err := resolveObservedAt("yesterday"); err == nil {
			t.Fatal("an unparseable observed_at must be refused")
		}
	})
}

// TestPassingTests_SkipsNonEvents proves build noise interleaved with events
// costs at most the noisy line, never the whole run.
func TestPassingTests_SkipsNonEvents(t *testing.T) {
	raw := []byte("# devnw.dev/canary\n{\"Action\":\"pass\",\"Package\":\"p\",\"Test\":\"TestA\"}\n{oops\n")
	got := PassingTests(raw)
	if len(got) != 1 || got["TestA"] != "p" {
		t.Fatalf("passed = %v, want TestA in package p", got)
	}
}
