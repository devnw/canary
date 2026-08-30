// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"strings"
	"testing"

	"devnw.dev/canary/pkg/evidence"
)

func TestAuditF02(t *testing.T) {
	commit := strings.Repeat("ab", 20)
	rec := func(req, feat string) evidence.Record {
		return evidence.Record{ProjectID: "p", RequirementID: req, Feature: feat, Aspect: "API",
			TestID: "TestX", Command: "go test ./...", Result: "PASS", CommitSHA: commit,
			ObservedAt: "2026-08-30T00:00:00Z", Runner: "local",
			ArtifactDigest: "sha256:" + strings.Repeat("00", 32)}
	}
	required := map[string][]evidence.FeatureKey{
		"CBIN-001": {{Feature: "A", Aspect: "API"}},
		"CBIN-002": {{Feature: "B", Aspect: "API"}},
	}
	v := evidence.Complete(required, []evidence.Record{rec("CBIN-001", "A")}, "p", commit, false)
	if v.OK || v.Code != "EVIDENCE_MISSING" {
		t.Fatalf("verdict %+v", v)
	}
	if len(v.Missing) != 1 || v.Missing[0].RequirementID != "CBIN-002" {
		t.Fatalf("must list only the missing requirement: %+v", v.Missing)
	}
	if v.Message != "no passing evidence at current commit" {
		t.Fatalf("message %q", v.Message)
	}
}
