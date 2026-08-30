// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package drift decides, per requirement, whether the tree in front of it
// still matches the index baseline `canary index` recorded. The verdict is
// one of three states — CURRENT, DRIFTED, UNKNOWN — and it is decided by
// evidence, never by a token's self-reported UPDATED= date:
//
//   - the content hash of every file a requirement's tokens live in is
//     compared against the hash stored at index time (a mismatch is DRIFTED,
//     which is why a same-day edit that never touches a date is still caught);
//   - a file whose hash matches is only CURRENT when git can also confirm its
//     history — when git is unavailable, not a repository, or the file is
//     untracked, the verdict is UNKNOWN, never CURRENT.
//
// Staleness and documentation rollup remain available as clearly separate
// advisory output (see Advisories); they are deliberately not folded into the
// drift State, because neither is evidence that the indexed code changed.
// CANARY: REQ=CP-278; FEATURE="DriftDetect"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CP_278_Check_CleanRepoCurrent,TestCANARY_CP_278_Check_ChangedFileDrifted,TestCANARY_CP_278_Check_GitFailureUnknown,TestCANARY_CP_278_Check_MissingBaselineUnknown,TestCANARY_CP_278_Check_UnreadableFileUnknown,TestCANARY_CP_278_Check_Precedence,TestCANARY_CP_278_Check_NoIndex,TestCANARY_CP_278_Advisories_Stale,TestCANARY_CP_278_Advisories_DocDrift,TestAuditF20; UPDATED=2026-08-30
package drift

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/storage"
)

// State is the drift verdict for a file or a requirement.
type State string

const (
	// StateCurrent means every file's content hash matches the index baseline
	// and git could confirm each file's history.
	StateCurrent State = "CURRENT"
	// StateDrifted means at least one file's content hash differs from the
	// index baseline: the code moved on since it was indexed.
	StateDrifted State = "DRIFTED"
	// StateUnknown means the verdict could not be decided — a missing
	// baseline, an unreadable file, or git being unable to answer for a file
	// whose hash otherwise matches. A git failure is UNKNOWN, never CURRENT.
	StateUnknown State = "UNKNOWN"
)

// ErrNoIndex is returned by Check when the database carries no index metadata,
// i.e. `canary index` has never populated it. Reporting UNKNOWN for every
// requirement would be technically true and practically useless; naming the
// one fix is better.
var ErrNoIndex = errors.New("no index; run 'canary index'")

// FileState is the drift verdict for one file a requirement's tokens live in.
type FileState struct {
	Path   string `json:"path"`
	State  State  `json:"state"`
	Detail string `json:"detail"`
}

// ReqState is the drift verdict for one requirement, rolled up from its files.
type ReqState struct {
	RequirementID string      `json:"requirement_id"`
	State         State       `json:"state"`
	Files         []FileState `json:"files"`
}

// Check returns a drift verdict for every requirement in the index at db,
// comparing the tree rooted at root against the baseline `canary index`
// recorded. projectID scopes the check to one project's tokens; "" checks
// every project's tokens.
//
// For each file a requirement's tokens live in, Check compares the file's
// current SHA-256 against the content_hash stored at index time. A mismatch is
// DRIFTED. A matching hash is CURRENT only when git can also confirm the
// file's history; when git cannot (no repository, git absent, file untracked)
// the file is UNKNOWN. A missing baseline hash or an unreadable file is also
// UNKNOWN. The requirement's state is the worst of its files: any DRIFTED
// makes it DRIFTED, else any UNKNOWN makes it UNKNOWN, else CURRENT.
//
// Check is read-only. It never re-scans the tree, never writes the database,
// and never fabricates a verdict from a token's UPDATED= date.
func Check(root string, db *storage.DB, projectID string) ([]ReqState, error) {
	if root == "" {
		root = "."
	}

	meta, err := db.GetIndexMeta()
	if err != nil {
		return nil, fmt.Errorf("read index metadata: %w", err)
	}
	if meta == nil {
		return nil, ErrNoIndex
	}

	tokens, err := db.GetAllTokens()
	if err != nil {
		return nil, fmt.Errorf("read tokens: %w", err)
	}

	// reqID -> (relative file path -> baseline content hash). One file can
	// carry several tokens; they share the whole-file hash, so the first
	// non-empty hash for a path is the baseline and later ones agree with it.
	baselines := map[string]map[string]string{}
	for _, tok := range tokens {
		if projectID != "" && tok.ProjectID != projectID {
			continue
		}
		if tok.FilePath == "" {
			continue
		}
		files := baselines[tok.ReqID]
		if files == nil {
			files = map[string]string{}
			baselines[tok.ReqID] = files
		}
		if existing, ok := files[tok.FilePath]; !ok || existing == "" {
			files[tok.FilePath] = tok.ContentHash
		}
	}

	gitCache := map[string]bool{} // file -> git could confirm its history
	states := make([]ReqState, 0, len(baselines))
	for reqID, files := range baselines {
		paths := make([]string, 0, len(files))
		for p := range files {
			paths = append(paths, p)
		}
		sort.Strings(paths)

		fileStates := make([]FileState, 0, len(paths))
		for _, p := range paths {
			fileStates = append(fileStates, checkFile(root, p, files[p], gitCache))
		}
		states = append(states, ReqState{
			RequirementID: reqID,
			State:         rollup(fileStates),
			Files:         fileStates,
		})
	}
	sort.Slice(states, func(i, j int) bool {
		return states[i].RequirementID < states[j].RequirementID
	})
	return states, nil
}

// checkFile decides one file's state against its indexed baseline hash.
func checkFile(root, relPath, baselineHash string, gitCache map[string]bool) FileState {
	fs := FileState{Path: relPath}

	if baselineHash == "" {
		fs.State = StateUnknown
		fs.Detail = "no baseline content hash in index"
		return fs
	}

	current, err := hashFile(filepath.Join(root, relPath))
	if err != nil {
		fs.State = StateUnknown
		fs.Detail = fmt.Sprintf("unreadable: %v", err)
		return fs
	}
	if current != baselineHash {
		fs.State = StateDrifted
		fs.Detail = "content hash differs from index baseline"
		return fs
	}

	// The hash matches. CURRENT additionally requires git to confirm the
	// file's history; when it cannot, the honest verdict is UNKNOWN.
	decidable, ok := gitCache[relPath]
	if !ok {
		decidable = gitTracks(root, relPath)
		gitCache[relPath] = decidable
	}
	if !decidable {
		fs.State = StateUnknown
		fs.Detail = "content matches baseline but git history is unavailable"
		return fs
	}

	fs.State = StateCurrent
	fs.Detail = "content hash matches index baseline"
	return fs
}

// rollup reduces a requirement's file states to a single verdict: any DRIFTED
// wins, else any UNKNOWN wins, else CURRENT. A requirement with no files is
// CURRENT (there is nothing that could have drifted).
func rollup(files []FileState) State {
	state := StateCurrent
	for _, f := range files {
		switch f.State {
		case StateDrifted:
			return StateDrifted
		case StateUnknown:
			state = StateUnknown
		}
	}
	return state
}

// gitTracks reports whether git can name the last commit that touched relPath
// under root. It returns false — never an error — when root is not a git
// repository, git is not on PATH, or the file is untracked: every one of those
// is a case where the git leg cannot decide, which the caller turns into
// UNKNOWN.
func gitTracks(root, relPath string) bool {
	cmd := exec.Command("git", "-C", root, "log", "-1", "--format=%H", "--", relPath) //nolint:gosec // fixed argv
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// hashFile returns the hex SHA-256 of a file's contents.
func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is an indexed token file
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// --- advisories: kept separate from the drift State -----------------------

// Kind values for Finding.Kind. These are advisory signals, deliberately kept
// out of the drift State: neither is evidence that indexed code changed.
const (
	KindStale    = "stale"
	KindDocDrift = "doc-drift"
)

// Finding is one advisory signal for a requirement (staleness or a doc
// rollup). It is not a drift verdict; see ReqState for that.
type Finding struct {
	ReqID  string `json:"req_id"`
	File   string `json:"file"`
	Kind   string `json:"kind"` // stale | doc-drift
	Detail string `json:"detail"`
}

// Advisories returns the non-drift signals worth surfacing alongside the drift
// verdict but never mixed into it: TESTED/BENCHED tokens past the staleness
// window, and — when root/.canary/canary.db exists — tokens whose tracked
// documentation is DOC_STALE or DOC_MISSING. If refTime is zero,
// time.Now().UTC() is used.
func Advisories(root string, rep canaryscan.Report, staleDays int, refTime time.Time) ([]Finding, error) {
	if root == "" {
		root = "."
	}
	if refTime.IsZero() {
		refTime = time.Now().UTC()
	}

	findings := staleFindings(rep, staleDays, refTime)

	docFindings, err := docDriftFindings(root)
	if err != nil {
		return nil, err
	}
	findings = append(findings, docFindings...)
	return findings, nil
}

// staleFindings mirrors canaryscan.Stale's TESTED/BENCHED + threshold rule,
// reporting a Finding per (requirement, file) instead of a diagnostic string.
func staleFindings(rep canaryscan.Report, staleDays int, refTime time.Time) []Finding {
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

// docDriftFindings reports tokens whose documentation has fallen out of sync,
// when root/.canary/canary.db exists. It never creates the database; a missing
// DB is treated as "not opted into doc tracking" and skipped silently.
func docDriftFindings(root string) ([]Finding, error) {
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
