// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package evidence

import "sort"

// FeatureKey identifies one claimed feature/aspect within a requirement.
type FeatureKey struct {
	Feature string `json:"feature"`
	Aspect  string `json:"aspect"`
}

// Missing describes one required feature/aspect that Complete could not
// find satisfying evidence for.
type Missing struct {
	RequirementID string     `json:"requirement_id"`
	Key           FeatureKey `json:"key"`
	Reason        string     `json:"reason"` // "no_evidence" | "wrong_commit" | "scope_mismatch"
}

// Verdict is the result of a completion check.
type Verdict struct {
	OK      bool      `json:"ok"`
	State   string    `json:"state"` // "VERIFIED" | "UNVERIFIED" | "UNKNOWN"
	Code    string    `json:"code"`  // "OK" | "EVIDENCE_MISSING" | "EMPTY_CLAIMS" | "SCAN_INCOMPLETE" | "EXTERNAL_UNKNOWN"
	Message string    `json:"message"`
	Missing []Missing `json:"-"`
}

// Complete is THE completion function: every required feature/aspect of
// every required requirement must have at least one record with
// Result=PASS (guaranteed by Parse), ProjectID==projectID, and
// CommitSHA==commit. An empty required map yields EMPTY_CLAIMS unless
// allowEmpty is set, in which case it is treated as trivially satisfied.
//
// The Missing list, when non-empty, is sorted deterministically by
// (RequirementID, Feature, Aspect) so repeated runs over the same inputs
// produce byte-identical output.
func Complete(required map[string][]FeatureKey, recs []Record, projectID, commit string, allowEmpty bool) Verdict {
	if len(required) == 0 {
		if allowEmpty {
			return Verdict{OK: true, State: "VERIFIED", Code: "OK", Message: "no claims (allowed)"}
		}
		return Verdict{OK: false, State: "UNVERIFIED", Code: "EMPTY_CLAIMS", Message: "no claims found"}
	}

	var missing []Missing
	for reqID, keys := range required {
		for _, key := range keys {
			reason, satisfied := evaluateKey(recs, reqID, key, projectID, commit)
			if satisfied {
				continue
			}
			missing = append(missing, Missing{RequirementID: reqID, Key: key, Reason: reason})
		}
	}

	sort.Slice(missing, func(i, j int) bool {
		a, b := missing[i], missing[j]
		if a.RequirementID != b.RequirementID {
			return a.RequirementID < b.RequirementID
		}
		if a.Key.Feature != b.Key.Feature {
			return a.Key.Feature < b.Key.Feature
		}
		return a.Key.Aspect < b.Key.Aspect
	})

	if len(missing) == 0 {
		return Verdict{OK: true, State: "VERIFIED", Code: "OK", Message: "all claims verified"}
	}
	return Verdict{
		OK:      false,
		State:   "UNVERIFIED",
		Code:    "EVIDENCE_MISSING",
		Message: "no passing evidence at current commit",
		Missing: missing,
	}
}

// evaluateKey checks whether any record satisfies (reqID, key) at
// projectID+commit, and when not, classifies why, in this precedence order:
//
//   - "wrong_commit": at least one record matches the key and this project,
//     but not this commit.
//   - "scope_mismatch": at least one record matches the key and this commit,
//     but not this project (only checked when "wrong_commit" does not
//     apply).
//   - "no_evidence": everything else -- either no record exists for the key
//     at all, OR every record for the key matches neither this project nor
//     this commit. A record that matches neither dimension carries no
//     probative value for this project+commit, so it counts as no evidence
//     rather than as a project/commit-specific mismatch.
func evaluateKey(recs []Record, reqID string, key FeatureKey, projectID, commit string) (reason string, satisfied bool) {
	var (
		anyForKey       bool
		anyProjectMatch bool // project matches; commit does not
		anyCommitMatch  bool // commit matches; project does not
	)
	for _, rec := range recs {
		if rec.RequirementID != reqID || rec.Feature != key.Feature || rec.Aspect != key.Aspect {
			continue
		}
		anyForKey = true
		projectOK := rec.ProjectID == projectID
		commitOK := rec.CommitSHA == commit
		if projectOK && commitOK {
			return "", true
		}
		if projectOK {
			anyProjectMatch = true
		}
		if commitOK {
			anyCommitMatch = true
		}
	}
	switch {
	case !anyForKey:
		return "no_evidence", false
	case anyProjectMatch:
		return "wrong_commit", false
	case anyCommitMatch:
		return "scope_mismatch", false
	default:
		return "no_evidence", false
	}
}
