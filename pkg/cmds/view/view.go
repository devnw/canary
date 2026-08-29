// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package view aggregates everything known about one requirement — tokens,
// files, tests, dependencies, spec/plan, diagrams, ticket link — into one
// bounded, agent-friendly answer.
// CANARY: REQ=CP-270; FEATURE="RequirementView"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_204_BuildView; UPDATED=2026-08-28
// CANARY: REQ=CP-274; FEATURE="MigrateNotesView"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_301_BuildView_MigrateNotes,TestCANARY_CBIN_301_BuildView_MigrateNotesCap; UPDATED=2026-08-29
// CANARY: REQ=CP-278; FEATURE="DriftedView"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_305_BuildView_Drifted,TestCANARY_CBIN_305_BuildView_NotDrifted,TestCANARY_CBIN_305_BuildView_NonGitRootSoftSkip; UPDATED=2026-08-29
package view

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/external"
	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/storage"
)

// View is the aggregate answer for one requirement: everything an agent or a
// human needs to orient on CANARY tokens, files, tests, dependencies,
// spec/plan location, diagram refs, and (when configured) the owning ticket
// system — in one bounded call.
type View struct {
	ReqID             string         `json:"req_id"`
	Source            string         `json:"source,omitempty"`
	TicketURL         string         `json:"ticket_url,omitempty"`
	Statuses          map[string]int `json:"statuses"`       // status -> token count
	Completion        int            `json:"completion_pct"` // TESTED+BENCHED tokens / total
	Features          []string       `json:"features"`       // "Feature (ASPECT, STATUS)"
	Files             []string       `json:"files"`          // capped at limit
	FilesTotal        int            `json:"files_total"`
	Tests             []string       `json:"tests"`
	Benches           []string       `json:"benches,omitempty"`
	DependsOn         []string       `json:"depends_on,omitempty"`
	Blocks            []string       `json:"blocks,omitempty"`
	RelatedTo         []string       `json:"related_to,omitempty"`
	SpecPath          string         `json:"spec_path,omitempty"`
	PlanPath          string         `json:"plan_path,omitempty"`
	Diagrams          []string       `json:"diagrams,omitempty"` // "file:line"
	DiagramsTotal     int            `json:"diagrams_total,omitempty"`
	MigrateNotes      []string       `json:"migrate_notes,omitempty"` // "file:line: text"
	MigrateNotesTotal int            `json:"migrate_notes_total,omitempty"`
	Drifted           bool           `json:"drifted,omitempty"`
	DriftReason       string         `json:"drift_reason,omitempty"`
}

// DefaultViewLimit bounds list sections (files, diagrams) by default; agents
// raise it with --limit when they need the full list.
const DefaultViewLimit = 10

// BuildView assembles the view from the index DB plus filesystem conventions.
// It assumes dbPath is already migrated (the CLI's PersistentPreRunE handles
// this before any command runs); callers such as tests that open a fresh
// database must migrate it themselves first.
func BuildView(dbPath, root, reqID string, limit int) (*View, error) {
	if limit <= 0 {
		limit = DefaultViewLimit
	}
	reg := sources.LoadFromRoot(root)
	reqID = reg.Normalize(strings.TrimSpace(reqID))

	db, err := storage.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open index db (run 'canary index' first): %w", err)
	}
	defer func() { _ = db.Close() }()

	tokens, err := db.GetTokensByReqID(reqID)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens found for %s (run 'canary index' to refresh)", reqID)
	}

	v := &View{ReqID: reqID, Statuses: map[string]int{}}
	if s, ok := reg.Resolve(reqID); ok {
		v.Source = s.Name
		v.TicketURL = reg.TicketURL(reqID)
	}

	fileSet, testSet, benchSet := map[string]struct{}{}, map[string]struct{}{}, map[string]struct{}{}
	depSet, blockSet, relSet := map[string]struct{}{}, map[string]struct{}{}, map[string]struct{}{}
	fileUpdated := map[string]string{} // file -> most recent token UPDATED referencing it
	done := 0
	for _, tok := range tokens {
		v.Statuses[tok.Status]++
		if tok.Status == "TESTED" || tok.Status == "BENCHED" {
			done++
		}
		v.Features = append(v.Features, fmt.Sprintf("%s (%s, %s)", tok.Feature, tok.Aspect, tok.Status))
		if tok.FilePath != "" {
			fileSet[tok.FilePath] = struct{}{}
			if tok.UpdatedAt > fileUpdated[tok.FilePath] {
				fileUpdated[tok.FilePath] = tok.UpdatedAt
			}
		}
		for _, s := range splitCSV(tok.Test) {
			testSet[s] = struct{}{}
		}
		for _, s := range splitCSV(tok.Bench) {
			benchSet[s] = struct{}{}
		}
		for _, s := range splitCSV(tok.DependsOn) {
			depSet[s] = struct{}{}
		}
		for _, s := range splitCSV(tok.Blocks) {
			blockSet[s] = struct{}{}
		}
		for _, s := range splitCSV(tok.RelatedTo) {
			relSet[s] = struct{}{}
		}
	}
	v.Completion = done * 100 / len(tokens)
	sort.Strings(v.Features)

	allFiles := sortedSet(fileSet)
	v.FilesTotal = len(allFiles)
	if len(allFiles) > limit {
		allFiles = allFiles[:limit]
	}
	v.Files = allFiles
	v.Tests = sortedSet(testSet)
	v.Benches = sortedSet(benchSet)
	v.DependsOn = annotateExternal(sortedSet(depSet), reg, root)
	v.Blocks = sortedSet(blockSet)
	v.RelatedTo = sortedSet(relSet)

	// Cheap, best-effort drift signal: compare the requirement's first file
	// (alphabetically) against its own most recent UPDATED claim via one git
	// log call. Soft-fails silently when git is unavailable, root isn't a
	// git repo, or the file is untracked — see driftCheck.
	if len(v.Files) > 0 {
		v.Drifted, v.DriftReason = driftCheck(root, v.Files[0], fileUpdated[v.Files[0]])
	}

	// Spec/plan by directory convention: .canary/specs/<REQ-ID>-<slug>/
	if matches, _ := filepath.Glob(filepath.Join(root, ".canary", "specs", reqID+"-*", "spec.md")); len(matches) > 0 {
		v.SpecPath = matches[0]
		if p := filepath.Join(filepath.Dir(matches[0]), "plan.md"); fileExists(p) {
			v.PlanPath = p
		}
	}

	if refs, err := db.GetRefsByReqID(reqID); err == nil {
		for _, r := range refs {
			switch r.Kind {
			case "diagram":
				v.Diagrams = append(v.Diagrams, fmt.Sprintf("%s:%d", r.FilePath, r.LineNumber))
			case "migrate":
				v.MigrateNotes = append(v.MigrateNotes, fmt.Sprintf("%s:%d: %s", r.FilePath, r.LineNumber, r.Context))
			}
		}
		v.DiagramsTotal = len(v.Diagrams)
		if len(v.Diagrams) > limit {
			v.Diagrams = v.Diagrams[:limit]
		}
		v.MigrateNotesTotal = len(v.MigrateNotes)
		if len(v.MigrateNotes) > limit {
			v.MigrateNotes = v.MigrateNotes[:limit]
		}
	}
	return v, nil
}

// annotateExternal decorates each id in depsOn that resolves to an external
// (ticket-source) dependency with its resolution, e.g. "ENG-12 (external:
// Done)" / "ENG-13 (external: In Progress)" / "ENG-14 (external: no cached
// ticket status)". Local ids (external.Resolve's "not external" case) are
// returned verbatim.
// CANARY: REQ=ENG-3960; FEATURE="ExternalDeps"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_ENG_3960_View_DependsOn_ExternalAnnotated,TestCANARY_ENG_3960_View_DependsOn_LocalUnchanged; UPDATED=2026-08-29
func annotateExternal(depsOn []string, reg *sources.Registry, root string) []string {
	out := make([]string, len(depsOn))
	for i, id := range depsOn {
		res := external.Resolve(id, reg, root)
		if !res.IsExternal() {
			out[i] = id
			continue
		}
		out[i] = fmt.Sprintf("%s (external: %s)", id, res.ShortDetail())
	}
	return out
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func sortedSet(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// driftCheck reports whether file's last git commit date is after
// tokenUpdated (a token's UPDATED field, "YYYY-MM-DD"). It never errors:
// missing git, a non-git root, an untracked file, or an unparsable date all
// resolve to (false, "") so BuildView's drift signal stays best-effort.
func driftCheck(root, file, tokenUpdated string) (bool, string) {
	if file == "" || tokenUpdated == "" {
		return false, ""
	}
	tokenDate, err := time.Parse("2006-01-02", tokenUpdated)
	if err != nil {
		return false, ""
	}
	out, err := exec.Command("git", "-C", root, "log", "-1", "--format=%cs", "--", file).Output()
	if err != nil {
		return false, ""
	}
	commitStr := strings.TrimSpace(string(out))
	if commitStr == "" {
		return false, ""
	}
	commitDate, err := time.Parse("2006-01-02", commitStr)
	if err != nil {
		return false, ""
	}
	if commitDate.After(tokenDate) {
		return true, fmt.Sprintf("%s committed %s, token updated %s", file, commitStr, tokenUpdated)
	}
	return false, ""
}

// CreateViewCommand returns the `canary view` command.
func CreateViewCommand() *cobra.Command {
	var jsonOut bool
	var limit int
	cmd := &cobra.Command{
		Use:   "view <REQ-ID>",
		Short: "Full picture of one requirement: tokens, files, tests, deps, spec, diagrams, ticket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := BuildView(".canary/canary.db", ".", args[0], limit)
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				return enc.Encode(v) // compact: no SetIndent
			}
			printView(cmd, v, limit)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "compact JSON output")
	cmd.Flags().IntVar(&limit, "limit", DefaultViewLimit, "max entries per list section (raise when you need more)")
	return cmd
}

func printView(cmd *cobra.Command, v *View, limit int) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "%s — %d%% complete", v.ReqID, v.Completion)
	if v.Source != "" {
		fmt.Fprintf(out, " (source: %s)", v.Source)
	}
	fmt.Fprintln(out)
	if v.TicketURL != "" {
		fmt.Fprintf(out, "Ticket:   %s\n", v.TicketURL)
	}
	statuses := make([]string, 0, len(v.Statuses))
	for s, n := range v.Statuses {
		statuses = append(statuses, fmt.Sprintf("%s=%d", s, n))
	}
	sort.Strings(statuses)
	fmt.Fprintf(out, "Tokens:   %s\n", strings.Join(statuses, " "))
	fmt.Fprintf(out, "Features: %s\n", strings.Join(v.Features, "; "))
	fmt.Fprintf(out, "Files:    %s", strings.Join(v.Files, ", "))
	if v.FilesTotal > len(v.Files) {
		fmt.Fprintf(out, " … +%d more (use --limit %d)", v.FilesTotal-len(v.Files), v.FilesTotal)
	}
	fmt.Fprintln(out)
	if len(v.Tests) > 0 {
		fmt.Fprintf(out, "Tests:    %s\n", strings.Join(v.Tests, ", "))
	}
	if len(v.Benches) > 0 {
		fmt.Fprintf(out, "Benches:  %s\n", strings.Join(v.Benches, ", "))
	}
	if len(v.DependsOn) > 0 {
		fmt.Fprintf(out, "Depends:  %s\n", strings.Join(v.DependsOn, ", "))
	}
	if len(v.Blocks) > 0 {
		fmt.Fprintf(out, "Blocks:   %s\n", strings.Join(v.Blocks, ", "))
	}
	if len(v.RelatedTo) > 0 {
		fmt.Fprintf(out, "Related:  %s\n", strings.Join(v.RelatedTo, ", "))
	}
	if v.Drifted {
		fmt.Fprintf(out, "Drift:    %s\n", v.DriftReason)
	}
	if v.SpecPath != "" {
		fmt.Fprintf(out, "Spec:     %s\n", v.SpecPath)
	}
	if v.PlanPath != "" {
		fmt.Fprintf(out, "Plan:     %s\n", v.PlanPath)
	}
	if len(v.Diagrams) > 0 {
		fmt.Fprintf(out, "Diagrams: %s", strings.Join(v.Diagrams, ", "))
		if v.DiagramsTotal > len(v.Diagrams) {
			fmt.Fprintf(out, " … +%d more (use --limit %d)", v.DiagramsTotal-len(v.Diagrams), v.DiagramsTotal)
		}
		fmt.Fprintln(out)
	}
	if len(v.MigrateNotes) > 0 {
		fmt.Fprintf(out, "Migrate:  %s", strings.Join(v.MigrateNotes, ", "))
		if v.MigrateNotesTotal > len(v.MigrateNotes) {
			fmt.Fprintf(out, " … +%d more (use --limit %d)", v.MigrateNotesTotal-len(v.MigrateNotes), v.MigrateNotesTotal)
		}
		fmt.Fprintln(out)
	}
}
