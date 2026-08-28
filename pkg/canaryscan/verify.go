package canaryscan

import (
	"fmt"
	"os"
	"strings"
	"time"

	"devnw.dev/canary/pkg/sources"
)

// VerifyClaims reads the GAP file and returns diagnostics for claimed-but-not-
// TESTED/BENCHED requirements. Claims are lines like "✅ <ID>" where <ID>
// matches any configured source key; a nil registry means the default (CBIN).
// CANARY: REQ=CBIN-201; FEATURE="TicketSources"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_201_VerifyClaimsTicketSource; UPDATED=2026-08-28
func VerifyClaims(rep Report, gapPath string, reg *sources.Registry) []string {
	if reg == nil {
		reg = sources.Default()
	}
	b, err := os.ReadFile(gapPath)
	if err != nil {
		return []string{fmt.Sprintf("CANARY_PARSE_ERROR file=%s err=%q", gapPath, err)}
	}
	matches := reg.ClaimPattern().FindAllStringSubmatch(string(b), -1)
	claimed := map[string]struct{}{}
	for _, m := range matches {
		claimed[m[1]] = struct{}{}
	}
	evidence := map[string]bool{}
	for _, r := range rep.Requirements {
		ok := false
		for _, f := range r.Features {
			if f.Status == "TESTED" || f.Status == "BENCHED" {
				ok = true
				break
			}
		}
		evidence[r.ID] = ok
	}
	var diags []string
	for id := range claimed {
		if !evidence[id] {
			diags = append(diags, fmt.Sprintf("CANARY_VERIFY_FAIL REQ=%s reason=claimed_but_not_TESTED_OR_BENCHED", id))
		}
	}
	return diags
}

// Stale returns diagnostics for TESTED/BENCHED tokens older than maxAge.
// If refTime is zero, time.Now().UTC() is used.
func Stale(rep Report, maxAge time.Duration, refTime time.Time) []string {
	if refTime.IsZero() {
		refTime = time.Now().UTC()
	}
	cut := refTime.Add(-maxAge)
	var diags []string
	for _, r := range rep.Requirements {
		for _, f := range r.Features {
			if f.Status == "TESTED" || f.Status == "BENCHED" {
				t, err := time.Parse("2006-01-02", f.Updated)
				if err != nil {
					diags = append(diags, fmt.Sprintf("CANARY_PARSE_ERROR file=%s err=%q", strings.Join(f.Files, ","), err))
					continue
				}
				if t.Before(cut) {
					age := int(refTime.Sub(t).Hours() / 24)
					diags = append(diags, fmt.Sprintf("CANARY_STALE REQ=%s updated=%s age_days=%d threshold=%d", r.ID, f.Updated, age, int(maxAge.Hours()/24)))
				}
			}
		}
	}
	return diags
}
