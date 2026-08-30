package canaryscan

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"devnw.dev/canary/pkg/evidence"
	"devnw.dev/canary/pkg/sources"
)

// UndeclaredKey is the stand-in feature/aspect used for a claimed
// requirement that declares no tokens at all. Such a claim can never be
// satisfied by evidence — there is nothing for evidence to attest to — so it
// is reported as one missing entry rather than passing vacuously.
var UndeclaredKey = evidence.FeatureKey{Feature: "*", Aspect: "*"}

// Claims returns the requirement IDs claimed in gapPath, in first-seen order
// with duplicates removed. The claim grammar is reg.ClaimPattern():
// "✅ <KEY>-<digits>"; a nil registry means the default (CBIN).
func Claims(gapPath string, reg *sources.Registry) ([]string, error) {
	if reg == nil {
		reg = sources.Default()
	}
	b, err := os.ReadFile(gapPath) //nolint:gosec // caller-supplied claims path
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	var ids []string
	for _, m := range reg.ClaimPattern().FindAllStringSubmatch(string(b), -1) {
		if _, dup := seen[m[1]]; dup {
			continue
		}
		seen[m[1]] = struct{}{}
		ids = append(ids, m[1])
	}
	return ids, nil
}

// RequiredFeatures maps each claimed requirement to the feature/aspect pairs
// it declares in rep. A claimed requirement rep knows nothing about maps to a
// single UndeclaredKey entry so it is still checked (and fails), never
// silently skipped.
func RequiredFeatures(rep Report, claimed []string) map[string][]evidence.FeatureKey {
	declared := make(map[string][]evidence.FeatureKey, len(rep.Requirements))
	for _, r := range rep.Requirements {
		keys := make([]evidence.FeatureKey, 0, len(r.Features))
		// Two tokens for the same feature/aspect (differing only in OWNER or
		// UPDATED) aggregate separately but describe one thing to prove, so
		// the key is required once, not twice.
		seen := make(map[evidence.FeatureKey]struct{}, len(r.Features))
		for _, f := range r.Features {
			key := evidence.FeatureKey{Feature: f.Feature, Aspect: f.Aspect}
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			keys = append(keys, key)
		}
		declared[r.ID] = keys
	}
	required := make(map[string][]evidence.FeatureKey, len(claimed))
	for _, id := range claimed {
		keys := declared[id]
		if len(keys) == 0 {
			keys = []evidence.FeatureKey{UndeclaredKey}
		}
		required[id] = keys
	}
	return required
}

// HeadCommit returns root's current commit SHA. Evidence binds to a commit,
// so a repository whose HEAD cannot be read cannot be verified.
func HeadCommit(root string) (string, error) {
	out, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	sha := strings.TrimSpace(string(out))
	if sha == "" {
		return "", errors.New("git rev-parse HEAD returned no commit")
	}
	return sha, nil
}

// VerifyClaims reads the GAP file and returns one diagnostic per claimed
// requirement that lacks passing evidence at commit. The decision is made by
// evidence.Complete — the single completion function — never by inspecting
// declared STATUS values: a token saying TESTED proves nothing.
//
// An empty claims file yields no diagnostics here (allowEmpty), preserving
// `scan --verify`'s historical contract; `canary verify` is where an empty
// claims file is itself a failure.
// CANARY: REQ=ENG-4322; FEATURE="TicketSources"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_201_VerifyClaimsTicketSource; UPDATED=2026-08-30
// CANARY: REQ=CP-236; FEATURE="EvidenceVerify"; ASPECT=Engine; STATUS=TESTED; TEST=TestVerifyClaims_NoEvidenceFails,TestVerifyClaims_CurrentEvidencePasses,TestVerifyClaims_UndeclaredClaimFails,TestVerifyClaims_EmptyClaimsFileIsSilent,TestVerifyClaims_UnreadableGapFileReportsParseError; UPDATED=2026-08-30
func VerifyClaims(rep Report, gapPath string, reg *sources.Registry, recs []evidence.Record, projectID, commit string) []string {
	claimed, err := Claims(gapPath, reg)
	if err != nil {
		return []string{fmt.Sprintf("CANARY_PARSE_ERROR file=%s err=%q", gapPath, err)}
	}
	v := evidence.Complete(RequiredFeatures(rep, claimed), recs, projectID, commit, true)
	failed := map[string]struct{}{}
	for _, m := range v.Missing {
		failed[m.RequirementID] = struct{}{}
	}
	ids := make([]string, 0, len(failed))
	for id := range failed {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	diags := make([]string, 0, len(ids))
	for _, id := range ids {
		diags = append(diags, fmt.Sprintf("CANARY_VERIFY_FAIL REQ=%s reason=no_current_evidence", id))
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
