// CANARY: REQ=CP-277; FEATURE="StalenessConfig"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_304_StaleDaysFromConfig,TestCANARY_CBIN_304_UpdateStaleReportsEvidenceCurrency,TestCANARY_CBIN_304_UpdateStaleMutatesNothing; UPDATED=2026-08-30
package canaryscan

import (
	"fmt"
	"regexp"
	"sort"

	"devnw.dev/canary/pkg/evidence"
)

// diagReqRe extracts the REQ ID from a CANARY_STALE diagnostic line
// (format: "CANARY_STALE REQ=<id> updated=<date> age_days=<n> threshold=<n>").
// REQ IDs are opaque, source-defined strings (e.g. "CBIN-304", legacy
// "PLAT-4521", or v2 multi-segment IDs like "CBIN-CLI-001"), so match
// anything up to the next whitespace rather than assuming a shape.
var diagReqRe = regexp.MustCompile(`REQ=(\S+)`)

// ReportEvidenceCurrency answers, for every requirement named by staleDiags,
// whether it has passing evidence at the current commit.
//
// This replaces the old --update-stale behavior, which rewrote UPDATED= dates
// in source. Rewriting a date made a stale claim *look* fresh without any new
// proof — the exact failure mode evidence-backed verification exists to
// prevent. Nothing is mutated here: the report is the whole output, one line
// per requirement, sorted by ID:
//
//	CANARY_UPDATE_STALE req=<id> evidence=current
//	CANARY_UPDATE_STALE req=<id> evidence=missing
//
// "current" means every feature/aspect the requirement declares has a PASS
// record for this project at this commit; anything less is "missing".
func ReportEvidenceCurrency(rep Report, staleDiags []string, recs []evidence.Record, projectID, commit string) []string {
	seen := map[string]struct{}{}
	var ids []string
	for _, diag := range staleDiags {
		m := diagReqRe.FindStringSubmatch(diag)
		if len(m) < 2 {
			continue
		}
		if _, dup := seen[m[1]]; dup {
			continue
		}
		seen[m[1]] = struct{}{}
		ids = append(ids, m[1])
	}
	if len(ids) == 0 {
		return nil
	}
	sort.Strings(ids)

	v := evidence.Complete(RequiredFeatures(rep, ids), recs, projectID, commit, true)
	missing := map[string]struct{}{}
	for _, m := range v.Missing {
		missing[m.RequirementID] = struct{}{}
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		state := "current"
		if _, gone := missing[id]; gone {
			state = "missing"
		}
		out = append(out, fmt.Sprintf("CANARY_UPDATE_STALE req=%s evidence=%s", id, state))
	}
	return out
}
