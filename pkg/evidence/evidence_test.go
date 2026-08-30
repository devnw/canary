// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package evidence

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
)

var (
	validCommit = strings.Repeat("ab", 20)
	validDigest = "sha256:" + strings.Repeat("00", 32)
)

func validRecordJSON() string {
	return `{
		"project_id":"p",
		"requirement_id":"CBIN-001",
		"feature":"A",
		"aspect":"API",
		"test_id":"TestX",
		"command":"go test ./...",
		"result":"PASS",
		"commit_sha":"` + validCommit + `",
		"observed_at":"2026-08-30T00:00:00Z",
		"runner":"local",
		"artifact_digest":"` + validDigest + `"
	}`
}

func TestParseValid(t *testing.T) {
	raw := `{"schema_version":1,"records":[` + validRecordJSON() + `]}`
	f, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if f.SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", f.SchemaVersion)
	}
	if len(f.Records) != 1 {
		t.Fatalf("Records = %d, want 1", len(f.Records))
	}
	rec := f.Records[0]
	if rec.Result != "PASS" {
		t.Fatalf("Result = %q, want PASS", rec.Result)
	}
	if rec.CommitSHA != validCommit {
		t.Fatalf("CommitSHA = %q", rec.CommitSHA)
	}
	if rec.RequirementID != "CBIN-001" {
		t.Fatalf("RequirementID = %q", rec.RequirementID)
	}
}

func TestParseRejectsUnknownField(t *testing.T) {
	raw := `{"schema_version":1,"records":[{
		"project_id":"p","requirement_id":"CBIN-001","feature":"A","aspect":"API",
		"test_id":"T","command":"go test","result":"PASS",
		"commit_sha":"` + validCommit + `","observed_at":"2026-08-30T00:00:00Z",
		"runner":"local","artifact_digest":"` + validDigest + `","extra":"nope"
	}]}`
	if _, err := Parse(strings.NewReader(raw)); err == nil {
		t.Fatal("want error for unknown field")
	}
}

// TestParseRejectsDuplicateField uses a raw JSON string (not a marshaled
// struct) with "result" appearing twice in one record object -- the only way
// to exercise encoding/json's silent last-value-wins behavior, which Parse
// must catch itself via a token walk.
func TestParseRejectsDuplicateField(t *testing.T) {
	raw := `{"schema_version":1,"records":[{
		"project_id":"p","requirement_id":"CBIN-001","feature":"A","aspect":"API",
		"test_id":"T","command":"go test","result":"PASS","result":"FAIL",
		"commit_sha":"` + validCommit + `","observed_at":"2026-08-30T00:00:00Z",
		"runner":"local","artifact_digest":"` + validDigest + `"
	}]}`
	_, err := Parse(strings.NewReader(raw))
	if err == nil {
		t.Fatal("want error for duplicate field")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("error = %v, want mention of duplicate", err)
	}
}

// TestParseRejectsDuplicateFieldTopLevel exercises the same detection at the
// top level of the file, not just inside a record.
func TestParseRejectsDuplicateFieldTopLevel(t *testing.T) {
	raw := `{"schema_version":1,"schema_version":1,"records":[]}`
	_, err := Parse(strings.NewReader(raw))
	if err == nil {
		t.Fatal("want error for top-level duplicate field")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("error = %v, want mention of duplicate", err)
	}
}

func TestParseRejectsNonPass(t *testing.T) {
	raw := `{"schema_version":1,"records":[{
		"project_id":"p","requirement_id":"CBIN-001","feature":"A","aspect":"API",
		"test_id":"T","command":"go test","result":"FAIL",
		"commit_sha":"` + validCommit + `","observed_at":"2026-08-30T00:00:00Z",
		"runner":"local","artifact_digest":"` + validDigest + `"
	}]}`
	_, err := Parse(strings.NewReader(raw))
	if err == nil {
		t.Fatal("want error for non-PASS result")
	}
	if !strings.Contains(err.Error(), "result") {
		t.Fatalf("error = %v, want mention of result", err)
	}
}

func TestParseRejectsShortCommit(t *testing.T) {
	shortCommit := strings.Repeat("a", 39) // 39 hex chars, one short
	raw := `{"schema_version":1,"records":[{
		"project_id":"p","requirement_id":"CBIN-001","feature":"A","aspect":"API",
		"test_id":"T","command":"go test","result":"PASS",
		"commit_sha":"` + shortCommit + `","observed_at":"2026-08-30T00:00:00Z",
		"runner":"local","artifact_digest":"` + validDigest + `"
	}]}`
	_, err := Parse(strings.NewReader(raw))
	if err == nil {
		t.Fatal("want error for short commit_sha")
	}
	if !strings.Contains(err.Error(), "commit_sha") {
		t.Fatalf("error = %v, want mention of commit_sha", err)
	}
	if !strings.Contains(err.Error(), "records[0]") {
		t.Fatalf("error = %v, want indexed record", err)
	}
}

func TestParseRejectsBadDigest(t *testing.T) {
	// Missing "sha256:" prefix; 63 hex chars (one short too).
	badDigest := strings.Repeat("0", 63)
	raw := `{"schema_version":1,"records":[{
		"project_id":"p","requirement_id":"CBIN-001","feature":"A","aspect":"API",
		"test_id":"T","command":"go test","result":"PASS",
		"commit_sha":"` + validCommit + `","observed_at":"2026-08-30T00:00:00Z",
		"runner":"local","artifact_digest":"` + badDigest + `"
	}]}`
	_, err := Parse(strings.NewReader(raw))
	if err == nil {
		t.Fatal("want error for bad artifact_digest")
	}
	if !strings.Contains(err.Error(), "artifact_digest") {
		t.Fatalf("error = %v, want mention of artifact_digest", err)
	}
}

func TestParseRejectsNonUTC(t *testing.T) {
	raw := `{"schema_version":1,"records":[{
		"project_id":"p","requirement_id":"CBIN-001","feature":"A","aspect":"API",
		"test_id":"T","command":"go test","result":"PASS",
		"commit_sha":"` + validCommit + `","observed_at":"2026-01-01T00:00:00+05:00",
		"runner":"local","artifact_digest":"` + validDigest + `"
	}]}`
	_, err := Parse(strings.NewReader(raw))
	if err == nil {
		t.Fatal("want error for non-UTC observed_at")
	}
	if !strings.Contains(err.Error(), "observed_at") {
		t.Fatalf("error = %v, want mention of observed_at", err)
	}
}

func TestParseAcceptsUTCOffsetForm(t *testing.T) {
	raw := `{"schema_version":1,"records":[{
		"project_id":"p","requirement_id":"CBIN-001","feature":"A","aspect":"API",
		"test_id":"T","command":"go test","result":"PASS",
		"commit_sha":"` + validCommit + `","observed_at":"2026-01-01T00:00:00+00:00",
		"runner":"local","artifact_digest":"` + validDigest + `"
	}]}`
	if _, err := Parse(strings.NewReader(raw)); err != nil {
		t.Fatalf("Parse: %v", err)
	}
}

func TestParseRejectsBadSchemaVersion(t *testing.T) {
	raw := `{"schema_version":2,"records":[]}`
	_, err := Parse(strings.NewReader(raw))
	if err == nil {
		t.Fatal("want error for bad schema_version")
	}
	if !strings.Contains(err.Error(), "schema_version") {
		t.Fatalf("error = %v, want mention of schema_version", err)
	}
}

func TestParseRejectsEmptyField(t *testing.T) {
	raw := `{"schema_version":1,"records":[{
		"project_id":"","requirement_id":"CBIN-001","feature":"A","aspect":"API",
		"test_id":"T","command":"go test","result":"PASS",
		"commit_sha":"` + validCommit + `","observed_at":"2026-08-30T00:00:00Z",
		"runner":"local","artifact_digest":"` + validDigest + `"
	}]}`
	_, err := Parse(strings.NewReader(raw))
	if err == nil {
		t.Fatal("want error for empty required field")
	}
	if !strings.Contains(err.Error(), "project_id") {
		t.Fatalf("error = %v, want mention of project_id", err)
	}
}

func TestLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(dir + "/does-not-exist.json")
	if err == nil {
		t.Fatal("want error for missing file")
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("error = %v, want fs.ErrNotExist", err)
	}
}

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/evidence.json"
	raw := `{"schema_version":1,"records":[` + validRecordJSON() + `]}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(f.Records) != 1 {
		t.Fatalf("Records = %d, want 1", len(f.Records))
	}
}

// TestSchemaFileMatchesParser keeps docs/canary-evidence.schema.json in
// lockstep with Parse's rules without pulling in a schema-validation
// dependency: it asserts the exact pattern/const strings Parse enforces
// appear verbatim in the schema document.
func TestSchemaFileMatchesParser(t *testing.T) {
	data, err := os.ReadFile("../../docs/canary-evidence.schema.json")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	raw := string(data)
	for _, want := range []string{
		`"pattern": "^[0-9a-f]{40}$"`,
		`"pattern": "^sha256:[0-9a-f]{64}$"`,
		`"pattern": "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}(\\.\\d+)?(Z|\\+00:00)$"`,
		`"const": "PASS"`,
		`"const": 1`,
		`"additionalProperties": false`,
	} {
		if !strings.Contains(raw, want) {
			t.Errorf("schema file missing verbatim string %q", want)
		}
	}
}
