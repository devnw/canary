// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package drift detects mismatches between what a CANARY token claims and
// what the repository actually shows: code committed after a token's UPDATED
// date (code-drift), tokens that have aged past the staleness window
// (stale), and — when a CANARY index database is present — documentation
// that has fallen out of sync with its source (doc-drift).
// CANARY: REQ=CP-278; FEATURE="DriftDetect"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_305_Detect_CodeDriftPositive,TestCANARY_CBIN_305_Detect_CodeDriftNegative,TestCANARY_CBIN_305_Detect_CodeDriftSameDayNotDrift,TestCANARY_CBIN_305_Detect_NonGitRootSoftSkip,TestCANARY_CBIN_305_Detect_UntrackedFileSoftSkip,TestCANARY_CBIN_305_Detect_Stale,TestCANARY_CBIN_305_Detect_StaleIgnoresNonTestedBenched,TestCANARY_CBIN_305_Detect_DocDrift,TestCANARY_CBIN_305_Detect_DocDriftNoDB,TestCANARY_CBIN_305_Detect_CodeDriftDedupesPerFile,TestCANARY_CBIN_305_Detect_CachesGitLogPerFile; UPDATED=2026-08-29
package drift

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/storage"
)

// Kind values for Finding.Kind.
const (
	KindCodeDrift = "code-drift"
	KindStale     = "stale"
	KindDocDrift  = "doc-drift"
)

// Finding is one detected drift signal for a requirement.
type Finding struct {
	ReqID  string `json:"req_id"`
	File   string `json:"file"`
	Kind   string `json:"kind"` // code-drift | stale | doc-drift
	Detail string `json:"detail"`
}

// Detect scans rep for drift signals relative to root:
//
//   - code-drift: a token's file was committed (per git history) after the
//     token's UPDATED date. Skips silently (no error, no finding) when root
//     is not a git repository, git is unavailable, or the file is untracked.
//   - stale: TESTED/BENCHED tokens older than staleDays, using the same rule
//     as canaryscan.Stale. staleDays <= 0 falls back to canaryscan.DefaultStaleDays.
//   - doc-drift: when root/.canary/canary.db exists, tokens whose DocStatus
//     is DOC_STALE or DOC_MISSING. Skips silently (no error, no finding)
//     when no database is present.
//
// If refTime is zero, time.Now().UTC() is used (mirroring canaryscan.Stale).
func Detect(root string, rep canaryscan.Report, staleDays int, refTime time.Time) ([]Finding, error) {
	if root == "" {
		root = "."
	}
	if refTime.IsZero() {
		refTime = time.Now().UTC()
	}

	var findings []Finding
	findings = append(findings, detectCodeDrift(root, rep)...)
	findings = append(findings, detectStale(rep, staleDays, refTime)...)

	docFindings, err := detectDocDrift(root)
	if err != nil {
		return nil, err
	}
	findings = append(findings, docFindings...)

	return findings, nil
}

// detectCodeDrift compares each unique token file's last git commit date
// against the UPDATED date of every token feature referencing it.
func detectCodeDrift(root string, rep canaryscan.Report) []Finding {
	return detectCodeDriftWith(root, rep, lastCommitDate)
}

// detectCodeDriftWith is detectCodeDrift with an injectable git-log lookup.
// Production code always goes through detectCodeDrift (which defaults
// gitLogFn to the real lastCommitDate); tests use this entry point to
// substitute a counting stub and verify lastCommitDate is invoked at most
// once per unique file, not once per (requirement, file) pair.
func detectCodeDriftWith(root string, rep canaryscan.Report, gitLogFn func(root, file string) string) []Finding {
	cache := map[string]string{} // file -> last commit date (YYYY-MM-DD, "" = none/untracked/no-git)
	seen := map[string]bool{}    // dedupe per (req, file)
	var findings []Finding

	for _, r := range rep.Requirements {
		for _, f := range r.Features {
			tokenDate, err := time.Parse("2006-01-02", f.Updated)
			if err != nil {
				continue
			}
			for _, file := range f.Files {
				key := r.ID + "\x00" + file
				if seen[key] {
					continue
				}
				commitDateStr, ok := cache[file]
				if !ok {
					commitDateStr = gitLogFn(root, file)
					cache[file] = commitDateStr
				}
				if commitDateStr == "" {
					continue
				}
				commitDate, err := time.Parse("2006-01-02", commitDateStr)
				if err != nil {
					continue
				}
				if commitDate.After(tokenDate) {
					seen[key] = true
					findings = append(findings, Finding{
						ReqID:  r.ID,
						File:   file,
						Kind:   KindCodeDrift,
						Detail: fmt.Sprintf("file committed %s, token updated %s", commitDateStr, f.Updated),
					})
				}
			}
		}
	}
	return findings
}

// lastCommitDate returns the last commit date (YYYY-MM-DD, from `git log
// --format=%cs`) for file relative to root. It returns "" — never an error —
// when root is not a git repository, git is not on PATH, or file is
// untracked; callers treat "" as "nothing to compare, skip".
func lastCommitDate(root, file string) string {
	rel, err := filepath.Rel(root, file)
	if err != nil {
		rel = file
	}
	cmd := exec.Command("git", "-C", root, "log", "-1", "--format=%cs", "--", rel)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// detectStale mirrors canaryscan.Stale's TESTED/BENCHED + threshold rule,
// but reports a Finding per (requirement, file) instead of a diagnostic
// string, so callers get the drifted file alongside the requirement.
func detectStale(rep canaryscan.Report, staleDays int, refTime time.Time) []Finding {
	if staleDays <= 0 {
		staleDays = canaryscan.DefaultStaleDays
	}
	maxAge := time.Duration(staleDays) * 24 * time.Hour
	cut := refTime.Add(-maxAge)

	var findings []Finding
	for _, r := range rep.Requirements {
		for _, f := range r.Features {
			if f.Status != "TESTED" && f.Status != "BENCHED" {
				continue
			}
			t, err := time.Parse("2006-01-02", f.Updated)
			if err != nil || !t.Before(cut) {
				continue
			}
			age := int(refTime.Sub(t).Hours() / 24)
			file := ""
			if len(f.Files) > 0 {
				file = f.Files[0]
			}
			findings = append(findings, Finding{
				ReqID:  r.ID,
				File:   file,
				Kind:   KindStale,
				Detail: fmt.Sprintf("updated %s, age %dd (threshold %dd)", f.Updated, age, staleDays),
			})
		}
	}
	return findings
}

// detectDocDrift reports tokens whose documentation has fallen out of sync,
// when root/.canary/canary.db exists. It never creates the database; a
// missing DB is treated as "not opted into doc tracking" and skipped
// silently.
func detectDocDrift(root string) ([]Finding, error) {
	dbPath := filepath.Join(root, ".canary", "canary.db")
	if _, err := os.Stat(dbPath); err != nil {
		return nil, nil
	}

	db, err := storage.OpenRO(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open index db: %w", err)
	}
	defer func() { _ = db.Close() }()

	tokens, err := db.GetAllTokens()
	if err != nil {
		return nil, fmt.Errorf("read tokens: %w", err)
	}

	var findings []Finding
	for _, tok := range tokens {
		if tok.DocStatus != "DOC_STALE" && tok.DocStatus != "DOC_MISSING" {
			continue
		}
		findings = append(findings, Finding{
			ReqID:  tok.ReqID,
			File:   tok.FilePath,
			Kind:   KindDocDrift,
			Detail: fmt.Sprintf("doc_status=%s doc_path=%s", tok.DocStatus, tok.DocPath),
		})
	}
	return findings, nil
}
