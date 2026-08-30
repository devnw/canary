// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package evidence

import (
	"strings"
	"testing"
)

func mkRecord(projectID, reqID, feature, aspect, commit string) Record {
	return Record{
		ProjectID:      projectID,
		RequirementID:  reqID,
		Feature:        feature,
		Aspect:         aspect,
		TestID:         "TestX",
		Command:        "go test ./...",
		Result:         "PASS",
		CommitSHA:      commit,
		ObservedAt:     "2026-08-30T00:00:00Z",
		Runner:         "local",
		ArtifactDigest: "sha256:" + strings.Repeat("00", 32),
	}
}

func TestCompleteAllPass(t *testing.T) {
	commit := strings.Repeat("ab", 20)
	required := map[string][]FeatureKey{
		"CBIN-001": {{Feature: "A", Aspect: "API"}},
	}
	recs := []Record{mkRecord("p", "CBIN-001", "A", "API", commit)}
	v := Complete(required, recs, "p", commit, false)
	if !v.OK || v.State != "VERIFIED" || v.Code != "OK" {
		t.Fatalf("verdict %+v", v)
	}
	if v.Message != "all claims verified" {
		t.Fatalf("message %q", v.Message)
	}
	if len(v.Missing) != 0 {
		t.Fatalf("missing %+v, want none", v.Missing)
	}
}

func TestCompleteListsOnlyMissing(t *testing.T) {
	commit := strings.Repeat("ab", 20)
	required := map[string][]FeatureKey{
		"CBIN-001": {{Feature: "A", Aspect: "API"}},
		"CBIN-002": {{Feature: "B", Aspect: "API"}},
	}
	// Only CBIN-001/A has evidence; CBIN-002/B has none at all.
	recs := []Record{mkRecord("p", "CBIN-001", "A", "API", commit)}
	v := Complete(required, recs, "p", commit, false)
	if v.OK || v.Code != "EVIDENCE_MISSING" {
		t.Fatalf("verdict %+v", v)
	}
	if len(v.Missing) != 1 {
		t.Fatalf("missing %+v, want exactly 1 entry", v.Missing)
	}
	if v.Missing[0].RequirementID != "CBIN-002" || v.Missing[0].Key.Feature != "B" || v.Missing[0].Reason != "no_evidence" {
		t.Fatalf("missing entry %+v", v.Missing[0])
	}
}

func TestCompleteWrongCommit(t *testing.T) {
	commit := strings.Repeat("ab", 20)
	other := strings.Repeat("cd", 20)
	required := map[string][]FeatureKey{
		"CBIN-001": {{Feature: "A", Aspect: "API"}},
	}
	recs := []Record{mkRecord("p", "CBIN-001", "A", "API", other)}
	v := Complete(required, recs, "p", commit, false)
	if v.OK || v.Code != "EVIDENCE_MISSING" {
		t.Fatalf("verdict %+v", v)
	}
	if len(v.Missing) != 1 || v.Missing[0].Reason != "wrong_commit" {
		t.Fatalf("missing %+v", v.Missing)
	}
}

func TestCompleteScopeMismatch(t *testing.T) {
	commit := strings.Repeat("ab", 20)
	required := map[string][]FeatureKey{
		"CBIN-001": {{Feature: "A", Aspect: "API"}},
	}
	// Evidence exists at the right commit but for a different project --
	// must not satisfy the requirement.
	recs := []Record{mkRecord("other-project", "CBIN-001", "A", "API", commit)}
	v := Complete(required, recs, "p", commit, false)
	if v.OK || v.Code != "EVIDENCE_MISSING" {
		t.Fatalf("verdict %+v", v)
	}
	if len(v.Missing) != 1 || v.Missing[0].Reason != "scope_mismatch" {
		t.Fatalf("missing %+v", v.Missing)
	}
}

func TestCompleteMissingIsSortedDeterministically(t *testing.T) {
	commit := strings.Repeat("ab", 20)
	required := map[string][]FeatureKey{
		"CBIN-002": {{Feature: "Z", Aspect: "API"}, {Feature: "A", Aspect: "API"}},
		"CBIN-001": {{Feature: "B", Aspect: "API"}},
	}
	v := Complete(required, nil, "p", commit, false)
	if len(v.Missing) != 3 {
		t.Fatalf("missing %+v, want 3 entries", v.Missing)
	}
	want := []struct{ req, feat string }{
		{"CBIN-001", "B"},
		{"CBIN-002", "A"},
		{"CBIN-002", "Z"},
	}
	for i, w := range want {
		if v.Missing[i].RequirementID != w.req || v.Missing[i].Key.Feature != w.feat {
			t.Fatalf("Missing[%d] = %+v, want req=%s feature=%s", i, v.Missing[i], w.req, w.feat)
		}
	}
}

func TestCompleteEmptyClaims(t *testing.T) {
	v := Complete(map[string][]FeatureKey{}, nil, "p", strings.Repeat("ab", 20), false)
	if v.OK || v.State != "UNVERIFIED" || v.Code != "EMPTY_CLAIMS" {
		t.Fatalf("verdict %+v", v)
	}
	if v.Message != "no claims found" {
		t.Fatalf("message %q", v.Message)
	}
}

func TestCompleteEmptyClaimsAllowed(t *testing.T) {
	v := Complete(map[string][]FeatureKey{}, nil, "p", strings.Repeat("ab", 20), true)
	if !v.OK || v.State != "VERIFIED" || v.Code != "OK" {
		t.Fatalf("verdict %+v", v)
	}
	if v.Message != "no claims (allowed)" {
		t.Fatalf("message %q", v.Message)
	}
}
