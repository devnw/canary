package canaryscan

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"devnw.dev/canary/pkg/config"
	"devnw.dev/canary/pkg/sources"
)

// AnnotateSources stamps each requirement with its source name and ticket URL.
func AnnotateSources(rep *Report, reg *sources.Registry) {
	if reg == nil {
		return
	}
	for i := range rep.Requirements {
		if s, ok := reg.Resolve(rep.Requirements[i].ID); ok {
			rep.Requirements[i].Source = s.Name
			rep.Requirements[i].TicketURL = reg.TicketURL(rep.Requirements[i].ID)
		}
	}
}

// DefaultStaleDays is the fallback staleness window (in days) when neither
// Config.StaleDays nor .canary/project.yaml's verification.staleness_days is
// set. It is pkg/config's constant under canaryscan's historical name.
const DefaultStaleDays = config.DefaultStaleDays

// staleThreshold resolves the effective staleness window: Config.StaleDays if
// set (>0), else the project config's resolved staleness_days (which is
// DefaultStaleDays when unconfigured). projCfg may be nil.
func staleThreshold(cfg Config, projCfg *config.ProjectConfig) time.Duration {
	days := cfg.StaleDays
	if days <= 0 {
		days = projCfg.StalenessDays()
	}
	if days <= 0 {
		days = DefaultStaleDays
	}
	return time.Duration(days) * 24 * time.Hour
}

// ScanSummaryLine returns a single parseable line for rep so agents get metrics without reading status.json.
func ScanSummaryLine(rep Report) string {
	s := rep.Summary
	return fmt.Sprintf("CANARY_SCAN tokens=%d requirements=%d STUB=%d IMPL=%d TESTED=%d BENCHED=%d",
		s.TotalTokens, s.UniqueRequirements,
		s.ByStatus["STUB"], s.ByStatus["IMPL"], s.ByStatus["TESTED"], s.ByStatus["BENCHED"])
}

// printIssues writes one parseable line per scan issue to w.
func printIssues(w io.Writer, issues []ScanIssue) {
	for _, is := range issues {
		_, _ = fmt.Fprintf(w, "CANARY_SCAN_ISSUE path=%s reason=%s\n", is.Path, is.Reason)
	}
}

// Run runs the full scan pipeline: load config, scan, optional update-stale, write outputs, verify, strict.
// It writes to stdout/stderr via the given writers (use os.Stdout, os.Stderr from CLI).
// Returns exit code: 0 success, 2 verify/staleness failure, 3 parse/IO error.
func Run(cfg Config, stdout, stderr io.Writer) (exitCode int) {
	if cfg.SkipRegex == nil {
		cfg.SkipRegex = DefaultSkipRegex()
	}
	if cfg.Root == "" {
		cfg.Root = "."
	}
	if cfg.Out == "" {
		cfg.Out = "status.json"
	}

	// The project config is loaded once, strictly: a config that cannot be
	// parsed or does not validate stops the scan rather than degrading it to
	// defaults that silently answer a different question.
	// CANARY: REQ=ENG-4317; FEATURE="StrictProjectConfig"; ASPECT=CLI; STATUS=TESTED; TEST=TestAuditF19_ScanFailsOnInvalidConfig,TestAuditF19_ScanWithoutConfigStillWorks; UPDATED=2026-08-30
	projCfg, err := config.Load(cfg.Root)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "CANARY_CONFIG_ERROR err=%q\n", err)
		return 3
	}
	reg, err := sources.FromProjectConfig(projCfg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "CANARY_CONFIG_ERROR err=%q\n", err)
		return 3
	}

	var projectFilter *regexp.Regexp
	if cfg.ProjectOnly {
		if projCfg.Requirements.IDPattern == "" {
			_, _ = fmt.Fprintf(stderr, "Warning: --project-only specified but .canary/project.yaml sets no requirements.id_pattern\n")
			_, _ = fmt.Fprintf(stderr, "Scanning all requirements. Run 'canary init' to create project config.\n")
		} else {
			projectFilter, err = regexp.Compile(projCfg.Requirements.IDPattern)
			if err != nil {
				_, _ = fmt.Fprintf(stderr, "CANARY_PARSE_ERROR err=%q\n", err)
				return 3
			}
			_, _ = fmt.Fprintf(stderr, "Filtering by project pattern: %s\n", projCfg.Requirements.IDPattern)
		}
	}

	ignorePatterns, ignoreErr := LoadCanaryIgnore(cfg.Root)
	if ignoreErr != nil {
		_, _ = fmt.Fprintf(stderr, "Warning: failed to load .canaryignore: %v\n", ignoreErr)
	}
	if ignorePatterns != nil {
		_, _ = fmt.Fprintf(stderr, "Loaded .canaryignore patterns\n")
	}

	rep, err := Scan(cfg.Root, cfg.SkipRegex, projectFilter, ignorePatterns, reg)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "CANARY_PARSE_ERROR err=%q\n", err)
		return 3
	}
	AnnotateSources(&rep, reg)

	refTime := RefTimeFromEnv()
	if refTime.IsZero() {
		refTime = time.Now().UTC()
	}
	threshold := staleThreshold(cfg, projCfg)
	if cfg.UpdateStale {
		staleDiags := Stale(rep, threshold, refTime)
		if len(staleDiags) > 0 {
			updatedFiles, tokenCount, updateIssues, err := UpdateStaleTokens(cfg.Root, cfg.SkipRegex, staleDiags, ignorePatterns)
			if err != nil {
				_, _ = fmt.Fprintf(stderr, "CANARY_UPDATE_ERROR: %v\n", err)
				return 3
			}
			printIssues(stderr, updateIssues)
			_, _ = fmt.Fprintf(stderr, "Updated %d stale tokens in %d files\n", tokenCount, len(updatedFiles))
			// Issues found only during the rewrite walk (e.g. a file the
			// walk could not read) must not be lost: merge them with the
			// first pass's issues before the re-scan overwrites rep, and
			// again with the re-scan's own issues below, so --strict sees
			// every issue either walk found.
			firstPassIssues := rep.Issues
			rep, err = Scan(cfg.Root, cfg.SkipRegex, projectFilter, ignorePatterns, reg)
			if err != nil {
				_, _ = fmt.Fprintf(stderr, "CANARY_PARSE_ERROR err=%q\n", err)
				return 3
			}
			AnnotateSources(&rep, reg)
			merged := make([]ScanIssue, 0, len(firstPassIssues)+len(updateIssues)+len(rep.Issues))
			merged = append(merged, firstPassIssues...)
			merged = append(merged, updateIssues...)
			merged = append(merged, rep.Issues...)
			rep.Issues = normalizeIssues(merged)
		} else {
			_, _ = fmt.Fprintln(stderr, "No stale tokens found")
		}
	}

	if err := WriteJSON(cfg.Out, rep); err != nil {
		_, _ = fmt.Fprintf(stderr, "CANARY_PARSE_ERROR err=%q\n", err)
		return 3
	}
	if cfg.CSV != "" {
		if err := WriteCSV(cfg.CSV, rep); err != nil {
			_, _ = fmt.Fprintf(stderr, "CANARY_PARSE_ERROR err=%q\n", err)
			return 3
		}
	}

	// Every file the scan could not fully read is surfaced, always — a
	// partial scan must never look like a clean one.
	printIssues(stderr, rep.Issues)

	// One-line stdout summary so agents get metrics without reading status.json.
	_, _ = fmt.Fprintln(stdout, ScanSummaryLine(rep))

	var diags []string
	if cfg.VerifyPath != "" {
		diags = append(diags, VerifyClaims(rep, cfg.VerifyPath, reg)...)
	}
	if cfg.Strict && !cfg.UpdateStale {
		diags = append(diags, Stale(rep, threshold, refTime)...)
	}
	// Under --strict an incomplete scan is a failure: the tree was not fully
	// examined, so no clean verdict can be given for it.
	if cfg.Strict {
		for _, is := range rep.Issues {
			diags = append(diags, fmt.Sprintf("SCAN_INCOMPLETE path=%s reason=%s", is.Path, is.Reason))
		}
	}
	if len(diags) > 0 {
		for _, d := range diags {
			_, _ = fmt.Fprintln(stderr, d)
		}
		if cfg.VerifyPath != "" || (cfg.Strict && len(rep.Issues) > 0) {
			_, _ = fmt.Fprintf(stdout, "CANARY_VERIFY_FAIL count=%d\n", len(diags))
		}
		return 2
	}
	if cfg.VerifyPath != "" {
		_, _ = fmt.Fprintln(stdout, "CANARY_VERIFY_OK")
	}
	return 0
}

// RunFromArgs is a convenience that builds Config from flag-like args and runs with os.Stdout/os.Stderr.
// Used by tools/canary main. skipExpr is the regex string; if empty, default is used.
// staleDays overrides the staleness window in days; 0 means "use config/default".
func RunFromArgs(root, out, csv, verifyPath, skipExpr string, strict, updateStale, projectOnly bool, staleDays int) int {
	skip := DefaultSkipRegex()
	if skipExpr != "" {
		var err error
		skip, err = regexp.Compile(skipExpr)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "CANARY_PARSE_ERROR err=%q\n", err)
			return 3
		}
	}
	cfg := Config{
		Root:        root,
		Out:         out,
		CSV:         csv,
		VerifyPath:  verifyPath,
		Strict:      strict,
		SkipRegex:   skip,
		UpdateStale: updateStale,
		ProjectOnly: projectOnly,
		StaleDays:   staleDays,
	}
	return Run(cfg, os.Stdout, os.Stderr)
}
