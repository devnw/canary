// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package canaryscan

import (
	"sort"

	"devnw.dev/canary/pkg/evidence"
)

// DeclaredFeatures returns rep's declared feature/aspect keys for reqID,
// deduplicated. Two tokens differing only in OWNER or UPDATED describe one
// thing to prove, so the key is required once, not twice. An id with no
// declarations returns nil.
func DeclaredFeatures(rep Report, reqID string) []evidence.FeatureKey {
	for _, r := range rep.Requirements {
		if r.ID != reqID {
			continue
		}
		keys := make([]evidence.FeatureKey, 0, len(r.Features))
		seen := make(map[evidence.FeatureKey]struct{}, len(r.Features))
		for _, f := range r.Features {
			key := evidence.FeatureKey{Feature: f.Feature, Aspect: f.Aspect}
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			keys = append(keys, key)
		}
		return keys
	}
	return nil
}

// VerifiedRequirements returns, sorted, the IDs of every requirement in rep
// whose declared features all have passing evidence for projectID at commit.
// It asks pkg/evidence.Complete -- the single completion function -- once per
// requirement, so a peer's export and this project's own `canary verify`
// answer the same question the same way.
//
// The result is always non-nil (an empty slice when nothing is verified) so a
// caller can hand it straight to Report.Verified, where nil carries the
// distinct meaning "not checked".
func VerifiedRequirements(rep Report, recs []evidence.Record, projectID, commit string) []string {
	verified := make([]string, 0, len(rep.Requirements))
	for _, r := range rep.Requirements {
		keys := DeclaredFeatures(rep, r.ID)
		if len(keys) == 0 {
			continue
		}
		v := evidence.Complete(map[string][]evidence.FeatureKey{r.ID: keys}, recs, projectID, commit, false)
		if v.OK {
			verified = append(verified, r.ID)
		}
	}
	sort.Strings(verified)
	return verified
}
