// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CP-252; FEATURE="NextPriorityCommand"; ASPECT=CLI; STATUS=BENCHED; TEST=TestCANARY_CBIN_132_CLI_NextPrioritySelection; BENCH=BenchmarkCANARY_CBIN_132_CLI_PriorityQuery; OWNER=canary; UPDATED=2026-08-30
package next

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/cmds/internal/utils"
	"devnw.dev/canary/pkg/config"
	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/storage"
)

// Output formats accepted by --format.
const (
	FormatJSON = "json"
	FormatText = "text"
)

// Sources a selection can come from, reported verbatim in --format json's
// "source" field so a caller always knows what the answer was computed from.
const (
	SourceDatabase   = "database"
	SourceFilesystem = "filesystem"
)

// defaultScanPriority is the priority a filesystem-scanned candidate carries
// when its token declared none. It is the same neutral default `canary index`
// gives an undeclared token, so the two selection sources rank the same tree
// the same way. A token that DOES declare a PRIORITY keeps it: the scan report
// carries the declaration (canaryscan.Feature.Priority).
const defaultScanPriority = 5

// CANARY: REQ=CP-252; FEATURE="NextCmd"; ASPECT=CLI; STATUS=BENCHED; TEST=TestCANARY_CBIN_132_CLI_NextPrioritySelection; BENCH=BenchmarkCANARY_CBIN_132_CLI_PriorityQuery; OWNER=canary; DOC=user:docs/user/next-priority-guide.md; DOC_HASH=6a276e6735100236; UPDATED=2026-08-30
var NextCmd = &cobra.Command{
	Use:   "next [flags]",
	Short: "Identify and implement next highest priority requirement",
	Long: `Identify the next highest priority unimplemented requirement and generate
comprehensive implementation guidance.

This command automatically:
- Uses the token index when it still describes this tree, and scans the tree
  itself when it does not
- Identifies highest priority STUB or IMPL requirement
- Excludes hidden requirements (test files, templates, examples)
- Requires every dependency to be proven complete -- passing evidence at the
  current commit for a local one, a satisfied external/peer state otherwise
- Generates comprehensive implementation prompt with:
  - Specification details
  - Constitutional principles
  - Test-first guidance
  - Token placement examples

Priority determination factors:
1. PRIORITY field (1=highest, 10=lowest); a token that declares none is
   ranked -- and reported -- at the neutral default 5
2. STATUS (STUB before IMPL)
3. DEPENDS_ON (every dependency must be proven complete by evidence, or
   satisfied as an external/peer requirement)

The answer names its own source: "database" when the index is present and
still describes this tree, "filesystem" when the tree had to be scanned
directly. Completion is only ever claimed from a current index.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := resolveFormat(cmd)
		if err != nil {
			return err
		}
		dbPath, _ := cmd.Flags().GetString("db")
		promptFlag, _ := cmd.Flags().GetBool("prompt")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		filterStatus, _ := cmd.Flags().GetString("status")
		filterAspect, _ := cmd.Flags().GetString("aspect")
		allowUnknownExternal, _ := cmd.Flags().GetBool("allow-unknown-external")

		// Build filters
		filters := make(map[string]string)
		if filterStatus != "" {
			filters["status"] = filterStatus
		}
		if filterAspect != "" {
			filters["aspect"] = filterAspect
		}

		// Select next priority
		projectID := utils.ReadProjectID(cmd)

		sel, err := selectNext(selection{
			DBPath:               dbPath,
			Root:                 ".",
			ProjectID:            projectID,
			Filters:              filters,
			AllowUnknownExternal: allowUnknownExternal,
			Stderr:               cmd.ErrOrStderr(),
		})
		if err != nil {
			// A PROJECT_REQUIRED refusal reaching here is the contract, not
			// a failure to select: it must arrive as the machine-readable
			// line on stdout, unwrapped.
			if guarded := utils.GuardContract(cmd, err); errors.Is(guarded, utils.ErrContractFailed) {
				return guarded
			}
			return fmt.Errorf("select next priority: %w", err)
		}

		out := cmd.OutOrStdout()

		// JSON is a machine contract and outranks the human-facing modes: a
		// caller that asked for parseable output gets it, dry-run or not.
		if format == FormatJSON {
			return emitJSON(out, sel)
		}

		if sel.Token == nil {
			printNoWork(out, sel.Source, sel.Blocked)
			return nil
		}

		if dryRun {
			fmt.Fprintf(out, "Next priority (dry run): %s - %s\n", sel.Token.ReqID, sel.Token.Feature)
			fmt.Fprintf(out, "Priority: %d | Status: %s | Aspect: %s\n", sel.Token.Priority, sel.Token.Status, sel.Token.Aspect)
			fmt.Fprintf(out, "Location: %s\n", sel.Token.FilePath)
			fmt.Fprintf(out, "Source: %s\n", sel.Source)
			return nil
		}

		// The prompt template (and any --prompt-arg it embeds) is read only
		// on the path that actually renders one: a --json or --dry-run run
		// used to fail on a missing template it was never going to use.
		promptArg, _ := cmd.Flags().GetString("prompt-arg")
		output, err := renderPrompt(sel.Token, projectID, promptFlag, promptArg)
		if err != nil {
			return fmt.Errorf("render prompt: %w", err)
		}

		fmt.Fprintln(out, output)
		return nil
	},
}

// resolveFormat resolves the output format from --format, with the deprecated
// --json as an alias. An explicit --format always wins, so the two can never
// disagree silently.
//
// A --format the caller actually typed is validated as given: `--format ""`
// is an invalid value and says so, rather than quietly meaning "text" and
// switching the --json alias off along the way. Only an unset flag falls back
// to text.
func resolveFormat(cmd *cobra.Command) (string, error) {
	format, _ := cmd.Flags().GetString("format")
	if !cmd.Flags().Changed("format") {
		if format == "" {
			format = FormatText
		}
		if jsonFlag, _ := cmd.Flags().GetBool("json"); jsonFlag {
			format = FormatJSON
		}
	}
	if format != FormatJSON && format != FormatText {
		return "", fmt.Errorf("unknown --format %q (want %q or %q)", format, FormatJSON, FormatText)
	}
	return format, nil
}

// nextJSONOutput is the structure emitted for next --format json. Source is
// always present; the requirement fields are absent when there was nothing to
// select, in which case Message says why.
//
// Priority is emitted for every selected requirement, 0 included: 0 is a
// priority a token can declare, and dropping it as "empty" would erase the
// most urgent declaration in the tree. It is a pointer only so that it stays
// absent alongside the other requirement fields when nothing was selected.
// When the selected token declared no PRIORITY, the number reported is the
// neutral default (5) -- the same one selection ranked it at.
type nextJSONOutput struct {
	ReqID    string `json:"req_id,omitempty"`
	Feature  string `json:"feature,omitempty"`
	Aspect   string `json:"aspect,omitempty"`
	Status   string `json:"status,omitempty"`
	Priority *int   `json:"priority,omitempty"`
	FilePath string `json:"file_path,omitempty"`
	Updated  string `json:"updated,omitempty"`
	Source   string `json:"source"`
	Message  string `json:"message,omitempty"`
}

// emitJSON writes the one JSON object that is the whole of stdout in JSON
// mode.
func emitJSON(w io.Writer, sel result) error {
	out := nextJSONOutput{Source: sel.Source}
	if sel.Token == nil {
		out.Message = noWorkMessage(sel.Source, sel.Blocked)
	} else {
		out.ReqID = sel.Token.ReqID
		out.Feature = sel.Token.Feature
		out.Aspect = sel.Token.Aspect
		out.Status = sel.Token.Status
		priority := sel.Token.Priority
		out.Priority = &priority
		out.FilePath = sel.Token.FilePath
		out.Updated = sel.Token.UpdatedAt
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	return nil
}

// noWorkMessage is what "nothing to select" means for each source. Work that
// exists but is blocked is never completion, whatever the source; and only an
// index that still describes this tree can support the claim that everything
// is done. A filesystem scan that found no candidate says exactly that and no
// more.
func noWorkMessage(source string, blocked int) string {
	if blocked > 0 {
		return fmt.Sprintf("no unblocked requirements (%d blocked by unmet dependencies)", blocked)
	}
	if source == SourceDatabase {
		return "all requirements completed"
	}
	return "no actionable requirements found"
}

// printNoWork writes the human-facing form of noWorkMessage.
func printNoWork(w io.Writer, source string, blocked int) {
	if blocked > 0 {
		// A dependency blocks for one of two reasons, and the hint has to name
		// both: an external one whose state could not be resolved (those are
		// noted on stderr), or a local one with no passing evidence at this
		// commit -- a declaration of TESTED is not proof, so `canary verify`
		// and the evidence store are where that one is answered.
		fmt.Fprintf(w, "%s\n", noWorkMessage(source, blocked))
		fmt.Fprintln(w, "\nEvery remaining candidate is waiting on a dependency:")
		fmt.Fprintln(w, "  • Run: canary deps check <REQ-ID>   (which dependency, and in what state)")
		fmt.Fprintln(w, "  • A local dependency blocks until evidence at this commit proves it:")
		fmt.Fprintln(w, "    run its tests, record the result, then: canary verify")
		fmt.Fprintln(w, "  • Unresolved external/peer dependencies are noted on stderr above")
		return
	}
	if source == SourceDatabase {
		fmt.Fprintln(w, "🎉 All requirements completed! No work available.")
		fmt.Fprintln(w, "\nSuggestions:")
		fmt.Fprintln(w, "  • Run: canary verify")
		fmt.Fprintln(w, "  • Review completed requirements")
		fmt.Fprintln(w, "  • Consider creating new specifications")
		return
	}
	fmt.Fprintf(w, "no actionable requirements found (source=%s)\n", SourceFilesystem)
	fmt.Fprintln(w, "\nThe tree was scanned directly because no current index was available.")
	fmt.Fprintln(w, "  • Run: canary index   (then re-run canary next)")
}

// PromptData holds template variables for prompt generation
type PromptData struct {
	ReqID             string
	Feature           string
	Aspect            string
	Status            string
	Priority          int
	SpecFile          string
	SpecContent       string
	Constitution      string
	RelatedSpecs      []RelatedSpec
	Dependencies      []*storage.Token
	SuggestedFiles    []string
	TestGuidance      string
	TokenExample      string
	SuccessCriteria   []string
	Today             string
	SuggestedTestFile string
	PackageName       string
	// PromptArg is the --prompt-arg value as given, and PromptContent the
	// prompt it resolved to. Both are empty when no extra prompt was asked
	// for.
	PromptArg     string
	PromptContent string
}

// RelatedSpec represents a related specification reference
type RelatedSpec struct {
	ReqID    string
	Feature  string
	SpecFile string
}

// selection are the resolved inputs of one `next` run.
type selection struct {
	DBPath string
	// Root is the tree the answer is about: where the config, the evidence
	// store, the git HEAD and (in filesystem mode) the scan all come from.
	Root string
	// ProjectID scopes index queries. Empty means "every project", which the
	// index refuses only when it would be ambiguous.
	ProjectID            string
	Filters              map[string]string
	AllowUnknownExternal bool
	Stderr               io.Writer
}

// result is one selection outcome: the chosen token (nil when there is
// nothing to do) and the source that answer came from.
type result struct {
	Token  *storage.Token
	Source string
	// Blocked counts the candidates that were passed over because a
	// dependency was not complete. It is what separates "there is nothing
	// left to do" from "everything left is waiting on something".
	Blocked int
}

// CANARY: REQ=CP-252; FEATURE="NextSelectionAPI"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_ENG_3960_MCP_Next_ExternalUnknown_Blocking,TestCANARY_ENG_3960_MCP_Next_ExternalSatisfied_NotBlocking; UPDATED=2026-08-30

// SelectNext is the exported entry point to `canary next`'s selection: the
// same source decision (fresh index, else canonical filesystem scan) and the
// same dependency gate the CLI applies, returning the chosen token, the
// source that answered, and how many candidates were passed over as blocked.
//
// It exists so the MCP `next` tool can delegate rather than reimplement.
// The MCP tool used to carry a hand-maintained replica of the dependency
// rule, and the replica drifted: it still accepted a declared STATUS=TESTED
// as proof and still let an unresolvable external dependency pass, months
// after the CLI stopped doing either. Two answers to "may this work start?"
// is one answer too many, so there is now exactly one implementation and the
// MCP surface calls it.
//
// root is the tree the answer is about; projectID scopes index queries
// ("" means every project, refused only when ambiguous); stderr receives the
// one-line notes about unresolvable dependencies (io.Discard is fine).
func SelectNext(dbPath, root, projectID string, filters map[string]string, allowUnknownExternal bool, stderr io.Writer) (token *storage.Token, source string, blocked int, err error) {
	res, err := selectNext(selection{
		DBPath:               dbPath,
		Root:                 root,
		ProjectID:            projectID,
		Filters:              filters,
		AllowUnknownExternal: allowUnknownExternal,
		Stderr:               stderr,
	})
	return res.Token, res.Source, res.Blocked, err
}

// selectNextPriority identifies the highest-priority actionable requirement
// under the default policy, discarding the source. It exists for callers
// (tests, benchmarks) that only want the token.
func selectNextPriority(dbPath, projectID string, filters map[string]string) (*storage.Token, error) {
	res, err := selectNext(selection{DBPath: dbPath, ProjectID: projectID, Filters: filters, Stderr: io.Discard})
	return res.Token, err
}

// selectNext picks the next requirement to work on and names where the answer
// came from.
//
// The index is used only when it still describes the tree in front of us (see
// openFreshIndex); otherwise the tree is scanned with the same scanner
// `canary scan` uses. Both paths then apply one dependency rule (depGate), so
// "may this start?" cannot depend on which source answered.
// CANARY: REQ=CP-252; FEATURE="NextSourceDecision"; ASPECT=CLI; STATUS=TESTED; TEST=TestAuditF13,TestAuditF13_EmptyTreeNeverClaimsCompletion,TestAuditF13_StaleIndexFallsBackToFilesystem,TestCANARY_CBIN_132_CLI_StaleIndexIsNotUsed; UPDATED=2026-08-30
func selectNext(s selection) (result, error) {
	if s.Root == "" {
		s.Root = "."
	}
	if s.Stderr == nil {
		s.Stderr = os.Stderr
	}

	cfg, err := config.Load(s.Root)
	if err != nil {
		return result{}, fmt.Errorf("load .canary/project.yaml: %w", err)
	}
	reg, err := sources.FromProjectConfig(cfg)
	if err != nil {
		return result{}, fmt.Errorf("load .canary/project.yaml: %w", err)
	}
	recs, err := loadEvidenceRecords(s.Root)
	if err != nil {
		return result{}, err
	}
	// A tree with no readable HEAD leaves commit empty, which no evidence
	// record can match: every dependency then fails closed rather than being
	// assumed done.
	commit, _ := canaryscan.HeadCommit(s.Root)

	gate := &depGate{
		root:                 s.Root,
		evidenceProjectID:    evidenceProjectID(cfg, s.ProjectID),
		commit:               commit,
		recs:                 recs,
		reg:                  reg,
		allowUnknownExternal: s.AllowUnknownExternal,
		warned:               map[string]bool{},
		stderr:               s.Stderr,
	}

	if db := openFreshIndex(s.DBPath, s.Root); db != nil {
		defer func() { _ = db.Close() }()
		token, err := selectFromDatabase(db, s.ProjectID, s.Filters, cfg.Requirements.IDPattern, gate)
		return result{Token: token, Source: SourceDatabase, Blocked: gate.blockedCount}, err
	}

	token, err := selectFromScan(s.Root, s.Filters, cfg.Requirements.IDPattern, reg, gate)
	return result{Token: token, Source: SourceFilesystem, Blocked: gate.blockedCount}, err
}

// evidenceProjectID is the project evidence is looked up under: the --project
// override when given, else the configured project.key. It is deliberately
// not the (possibly empty) index scope -- evidence records always carry a
// concrete project, so an empty one would match nothing.
func evidenceProjectID(cfg *config.ProjectConfig, override string) string {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override)
	}
	return cfg.ProjectID()
}

// openFreshIndex opens the index read-only and returns it ONLY when it still
// describes the tree at root. It returns nil -- meaning "scan the filesystem
// instead" -- when:
//
//   - there is no database, or it cannot be opened (including a schema older
//     than this binary's);
//   - it carries no index metadata (it was created but never built);
//   - it was built from a different root;
//   - it was built at a different commit, or the current commit cannot be
//     read at all (an unknowable freshness is not freshness);
//   - it was built by a different token grammar.
//
// Nothing is ever created here: a read must not have a write side effect. The
// old code opened whatever database it found and answered from it, which is
// how `canary next` came to announce "all requirements completed" over an
// index that had never been built.
func openFreshIndex(dbPath, root string) *storage.DB {
	db, err := storage.OpenRO(dbPath)
	if err != nil {
		return nil
	}
	meta, err := db.GetIndexMeta()
	if err != nil || meta == nil {
		_ = db.Close()
		return nil
	}
	if !sameDir(meta.Root, root) || meta.ParserSchema != canaryscan.ParserSchemaVersion {
		_ = db.Close()
		return nil
	}
	head, err := canaryscan.HeadCommit(root)
	if err != nil || head == "" || head != meta.CommitSHA {
		_ = db.Close()
		return nil
	}
	return db
}

// sameDir reports whether two directory paths name the same place, comparing
// them absolute so the conventional "." recorded by `canary index` matches the
// root a later `canary next` resolves.
func sameDir(a, b string) bool {
	absA, errA := filepath.Abs(a)
	absB, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return absA == absB
}

// selectFromScan answers from a canonical filesystem scan -- the same scanner,
// ignore rules and token grammar `canary scan` uses, so the two can never
// disagree about what the tree contains.
func selectFromScan(root string, filters map[string]string, idPattern string, reg *sources.Registry, gate *depGate) (*storage.Token, error) {
	var projectFilter *regexp.Regexp
	if idPattern != "" {
		compiled, err := regexp.Compile(idPattern)
		if err != nil {
			return nil, fmt.Errorf("compile requirements.id_pattern %q: %w", idPattern, err)
		}
		projectFilter = compiled
	}
	ignorePatterns, err := canaryscan.LoadCanaryIgnore(root)
	if err != nil {
		return nil, fmt.Errorf("load .canaryignore: %w", err)
	}
	rep, err := canaryscan.Scan(root, canaryscan.StateSkipRegex(), projectFilter, ignorePatterns, reg)
	if err != nil {
		return nil, fmt.Errorf("scan %s: %w", root, err)
	}

	declared := scanDeclared(rep)
	for _, cand := range scanCandidates(rep, filters) {
		blocked, err := gate.blocked(cand.DependsOn, declared)
		if err != nil {
			return nil, err
		}
		if !blocked {
			return cand, nil
		}
	}
	return nil, nil
}

// scanCandidates turns a scan report into the ordered candidate list: filtered
// the same way the index query filters, then sorted by declared priority
// (ascending, defaultScanPriority when a token declares none), then STUB
// before IMPL, then requirement id, feature and aspect -- a total order, so
// two runs over an unchanged tree pick the same requirement.
func scanCandidates(rep canaryscan.Report, filters map[string]string) []*storage.Token {
	var candidates []*storage.Token
	_, hasStatusFilter := filters["status"]
	for _, r := range rep.Requirements {
		for _, f := range r.Features {
			if want, ok := filters["status"]; ok && f.Status != want {
				continue
			}
			if want, ok := filters["aspect"]; ok && f.Aspect != want {
				continue
			}
			if !hasStatusFilter && f.Status != "STUB" && f.Status != "IMPL" {
				continue
			}
			file := ""
			if len(f.Files) > 0 {
				file = f.Files[0]
			}
			if includeHidden, ok := filters["include_hidden"]; (!ok || includeHidden != "true") && isHiddenPath(file) {
				continue
			}
			priority := defaultScanPriority
			if f.Priority != nil {
				priority = *f.Priority
			}
			candidates = append(candidates, &storage.Token{
				ReqID:     r.ID,
				Feature:   f.Feature,
				Aspect:    f.Aspect,
				Status:    f.Status,
				Priority:  priority,
				FilePath:  file,
				Owner:     f.Owner,
				UpdatedAt: f.Updated,
				DependsOn: strings.Join(f.DependsOn, ","),
			})
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if a.Priority != b.Priority {
			return a.Priority < b.Priority
		}
		if ra, rb := statusOrder(a.Status), statusOrder(b.Status); ra != rb {
			return ra < rb
		}
		if a.ReqID != b.ReqID {
			return a.ReqID < b.ReqID
		}
		if a.Feature != b.Feature {
			return a.Feature < b.Feature
		}
		return a.Aspect < b.Aspect
	})
	return candidates
}

// statusOrder ranks statuses by how much work they represent: unstarted work
// comes before work already in progress. Anything unrecognized sorts last.
func statusOrder(status string) int {
	switch status {
	case "STUB":
		return 0
	case "IMPL":
		return 1
	case "TESTED":
		return 2
	case "BENCHED":
		return 3
	default:
		return 4
	}
}

// selectFromDatabase queries the index for next priority.
func selectFromDatabase(db *storage.DB, projectID string, filters map[string]string, idPattern string, gate *depGate) (*storage.Token, error) {
	// Build filters for incomplete requirements
	if filters == nil {
		filters = make(map[string]string)
	}
	declared := dbDeclared(db, projectID)

	// If no status filter, only select STUB or IMPL by default
	if _, hasStatusFilter := filters["status"]; !hasStatusFilter {
		// Query separately for STUB and IMPL, prioritizing STUB
		stubFilters := make(map[string]string)
		for k, v := range filters {
			stubFilters[k] = v
		}
		stubFilters["status"] = "STUB"

		// Try STUB first
		tokens, err := db.ListTokens(projectID, anyFilters(stubFilters), idPattern, "", 50)
		if err != nil {
			return nil, fmt.Errorf("query STUB tokens: %w", err)
		}

		// Filter out blocked tokens
		for _, token := range tokens {
			blocked, err := gate.blocked(token.DependsOn, declared)
			if err != nil {
				return nil, err
			}
			if !blocked {
				return token, nil
			}
		}

		// Try IMPL if no STUB available
		implFilters := make(map[string]string)
		for k, v := range filters {
			implFilters[k] = v
		}
		implFilters["status"] = "IMPL"

		tokens, err = db.ListTokens(projectID, anyFilters(implFilters), idPattern, "", 50)
		if err != nil {
			return nil, fmt.Errorf("query IMPL tokens: %w", err)
		}

		for _, token := range tokens {
			blocked, err := gate.blocked(token.DependsOn, declared)
			if err != nil {
				return nil, err
			}
			if !blocked {
				return token, nil
			}
		}

		return nil, nil // No work available
	}

	// Use provided filters
	tokens, err := db.ListTokens(projectID, anyFilters(filters), idPattern, "", 50)
	if err != nil {
		return nil, fmt.Errorf("query tokens: %w", err)
	}

	// Find first unblocked token
	for _, token := range tokens {
		blocked, err := gate.blocked(token.DependsOn, declared)
		if err != nil {
			return nil, err
		}
		if !blocked {
			return token, nil
		}
	}

	return nil, nil // No unblocked work available
}

// isHiddenPath determines if a token should be hidden based on its file path
func isHiddenPath(filePath string) bool {
	hiddenPatterns := []string{
		// Test files
		"_test.go", "Test.", "/tests/", "/test/",
		// Template directories
		".canary/templates/", "/templates/", "/base/", "/embedded/",
		// Documentation examples
		"IMPLEMENTATION_SUMMARY", "FINAL_SUMMARY", "README_CANARY.md", "GAP_ANALYSIS.md",
		// AI agent directories
		".claude/", ".cursor/", ".github/prompts/", ".windsurf/", ".kilocode/",
		".roo/", ".opencode/", ".codex/", ".augment/", ".codebuddy/", ".amazonq/",
	}

	for _, pattern := range hiddenPatterns {
		if strings.Contains(filePath, pattern) {
			return true
		}
	}
	return false
}

// renderPrompt generates the implementation prompt from the template.
//
// promptArg names an extra prompt (a file, or an embedded prompt's name) to
// make available to the template as {{.PromptArg}}/{{.PromptContent}}. It is
// resolved here, on the one path that renders a template, rather than at
// command start: a --format json or --dry-run run has no template to fill and
// must not fail on a prompt it will never use.
func renderPrompt(token *storage.Token, projectID string, promptFlag bool, promptArg string) (string, error) {
	if !promptFlag {
		// Simple summary output
		return fmt.Sprintf("Next: %s - %s (Priority: %d, Status: %s)\n"+
			"Run with --prompt for full implementation guidance.",
			token.ReqID, token.Feature, token.Priority, token.Status), nil
	}

	// Load template
	templatePath := ".canary/templates/next-prompt-template.md"
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("read template: %w", err)
	}

	tmpl, err := template.New("next-prompt").Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	// Load prompt data
	data, err := loadPromptData(token, projectID)
	if err != nil {
		return "", fmt.Errorf("load prompt data: %w", err)
	}
	if promptArg != "" {
		content, err := utils.LoadPrompt(promptArg)
		if err != nil {
			return "", err
		}
		data.PromptArg = promptArg
		data.PromptContent = content
	}

	// Render template
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// loadPromptData loads all data needed for template rendering
// projectID scopes the dependency lookups; "" spans every project.
func loadPromptData(token *storage.Token, projectID string) (*PromptData, error) {
	data := &PromptData{
		ReqID:    token.ReqID,
		Feature:  token.Feature,
		Aspect:   token.Aspect,
		Status:   token.Status,
		Priority: token.Priority,
		Today:    time.Now().UTC().Format("2006-01-02"),
	}

	// Load specification file
	specPattern := fmt.Sprintf(".canary/specs/%s-*/spec.md", token.ReqID)
	matches, err := filepath.Glob(specPattern)
	if err == nil && len(matches) > 0 {
		data.SpecFile = matches[0]
		specContent, err := os.ReadFile(matches[0])
		if err == nil {
			data.SpecContent = string(specContent)

			// Extract success criteria from spec
			data.SuccessCriteria = extractSuccessCriteria(data.SpecContent)
		}
	}

	// Load constitution
	constitutionPath := ".canary/memory/constitution.md"
	constitutionContent, err := os.ReadFile(constitutionPath)
	if err == nil {
		data.Constitution = string(constitutionContent)
	}

	// Generate suggested files based on aspect
	data.SuggestedFiles = suggestFileLocations(token.Aspect)

	// Generate test guidance
	data.TestGuidance = generateTestGuidance(token)

	// Generate token example
	data.TokenExample = generateTokenExample(token)

	// Determine package name and test file
	data.PackageName = guessPackageName(token.Aspect)
	data.SuggestedTestFile = fmt.Sprintf("cmd/canary/%s_test.go", strings.ToLower(token.Feature))

	// Load dependencies if in database
	dbPath := ".canary/canary.db"
	if db, err := storage.OpenRO(dbPath); err == nil {
		defer func() { _ = db.Close() }()
		if token.DependsOn != "" {
			deps := strings.Split(token.DependsOn, ",")
			for _, dep := range deps {
				dep = strings.TrimSpace(dep)
				if dep == "" {
					continue
				}
				depTokens, err := db.GetTokensByReqID(projectID, dep)
				if err == nil && len(depTokens) > 0 {
					data.Dependencies = append(data.Dependencies, depTokens[0])
				}
			}
		}
	}

	return data, nil
}

// extractSuccessCriteria extracts success criteria from specification
func extractSuccessCriteria(specContent string) []string {
	var criteria []string
	inSection := false

	lines := strings.Split(specContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for success criteria section
		if strings.Contains(strings.ToLower(line), "success criteria") {
			inSection = true
			continue
		}

		// Stop at next major section
		if inSection && strings.HasPrefix(line, "##") {
			break
		}

		// Extract list items
		if inSection && (strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*")) {
			criterion := strings.TrimLeft(line, "-* \t")
			if criterion != "" {
				criteria = append(criteria, criterion)
			}
		}
	}

	if len(criteria) == 0 {
		criteria = []string{
			"Implementation meets specification requirements",
			"All tests pass",
			"Code follows project conventions",
		}
	}

	return criteria
}

// suggestFileLocations suggests file locations based on aspect
func suggestFileLocations(aspect string) []string {
	suggestions := map[string][]string{
		"CLI":      {"cmd/canary/main.go", "cmd/canary/*.go"},
		"API":      {"internal/*/api.go", "pkg/*/api.go"},
		"Engine":   {"internal/engine/*.go", "pkg/engine/*.go"},
		"Storage":  {"internal/storage/*.go"},
		"Security": {"internal/security/*.go", "pkg/security/*.go"},
	}

	if files, ok := suggestions[aspect]; ok {
		return files
	}

	return []string{"cmd/", "internal/", "pkg/"}
}

// generateTestGuidance creates test-first guidance
func generateTestGuidance(token *storage.Token) string {
	return fmt.Sprintf(`Create tests that verify the %s functionality:
- Test happy path with valid inputs
- Test error cases with invalid inputs
- Test edge cases and boundary conditions
- Test integration with existing components

Use table-driven tests where appropriate for multiple scenarios.`, token.Feature)
}

// generateTokenExample creates CANARY token placement example
func generateTokenExample(token *storage.Token) string {
	today := time.Now().UTC().Format("2006-01-02")
	return fmt.Sprintf(`// CANARY: REQ=%s; FEATURE="%s"; ASPECT=%s; STATUS=STUB; UPDATED=%s
func %s() error {
    // TODO: implement
    return nil
}`, token.ReqID, token.Feature, token.Aspect, today, token.Feature)
}

// guessPackageName guesses package name from aspect
func guessPackageName(aspect string) string {
	names := map[string]string{
		"CLI":      "main",
		"API":      "api",
		"Engine":   "engine",
		"Storage":  "storage",
		"Security": "security",
	}

	if name, ok := names[aspect]; ok {
		return name
	}

	return "main"
}

// anyFilters widens a string filter map to the map[string]any shape
// storage.ListTokens takes. Filter values are bound as query parameters, so
// only their type changes here, never their meaning.
func anyFilters(in map[string]string) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
