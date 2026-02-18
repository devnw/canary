package canaryscan

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"time"
)

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

	var projectFilter *regexp.Regexp
	if cfg.ProjectOnly {
		projCfg, err := LoadProjectConfig(cfg.Root)
		if err != nil {
			fmt.Fprintf(stderr, "Warning: --project-only specified but failed to load .canary/project.yaml: %v\n", err)
			fmt.Fprintf(stderr, "Scanning all requirements. Run 'canary init' to create project config.\n")
		} else if projCfg != nil && projCfg.Requirements.IDPattern != "" {
			projectFilter, err = regexp.Compile(projCfg.Requirements.IDPattern)
			if err != nil {
				fmt.Fprintf(stderr, "CANARY_PARSE_ERROR err=%q\n", err)
				return 3
			}
			fmt.Fprintf(stderr, "Filtering by project pattern: %s\n", projCfg.Requirements.IDPattern)
		}
	}

	ignorePatterns, err := LoadCanaryIgnore(cfg.Root)
	if err != nil {
		fmt.Fprintf(stderr, "Warning: failed to load .canaryignore: %v\n", err)
	}
	if ignorePatterns != nil {
		fmt.Fprintf(stderr, "Loaded .canaryignore patterns\n")
	}

	rep, err := Scan(cfg.Root, cfg.SkipRegex, projectFilter, ignorePatterns)
	if err != nil {
		fmt.Fprintf(stderr, "CANARY_PARSE_ERROR err=%q\n", err)
		return 3
	}

	refTime := RefTimeFromEnv()
	if refTime.IsZero() {
		refTime = time.Now().UTC()
	}
	if cfg.UpdateStale {
		staleDiags := Stale(rep, 30*24*time.Hour, refTime)
		if len(staleDiags) > 0 {
			updatedFiles, err := UpdateStaleTokens(cfg.Root, cfg.SkipRegex, staleDiags)
			if err != nil {
				fmt.Fprintf(stderr, "CANARY_UPDATE_ERROR: %v\n", err)
				return 3
			}
			fmt.Fprintf(stderr, "Updated %d stale tokens in %d files\n", len(staleDiags), len(updatedFiles))
			rep, err = Scan(cfg.Root, cfg.SkipRegex, projectFilter, ignorePatterns)
			if err != nil {
				fmt.Fprintf(stderr, "CANARY_PARSE_ERROR err=%q\n", err)
				return 3
			}
		} else {
			fmt.Fprintln(stderr, "No stale tokens found")
		}
	}

	if err := WriteJSON(cfg.Out, rep); err != nil {
		fmt.Fprintf(stderr, "CANARY_PARSE_ERROR err=%q\n", err)
		return 3
	}
	if cfg.CSV != "" {
		if err := WriteCSV(cfg.CSV, rep); err != nil {
			fmt.Fprintf(stderr, "CANARY_PARSE_ERROR err=%q\n", err)
			return 3
		}
	}

	var diags []string
	if cfg.VerifyPath != "" {
		diags = append(diags, VerifyClaims(rep, cfg.VerifyPath)...)
	}
	if cfg.Strict && !cfg.UpdateStale {
		diags = append(diags, Stale(rep, 30*24*time.Hour, refTime)...)
	}
	if len(diags) > 0 {
		for _, d := range diags {
			fmt.Fprintln(stderr, d)
		}
		return 2
	}
	return 0
}

// RunFromArgs is a convenience that builds Config from flag-like args and runs with os.Stdout/os.Stderr.
// Used by tools/canary main. skipExpr is the regex string; if empty, default is used.
func RunFromArgs(root, out, csv, verifyPath, skipExpr string, strict, updateStale, projectOnly bool) int {
	skip := DefaultSkipRegex()
	if skipExpr != "" {
		var err error
		skip, err = regexp.Compile(skipExpr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "CANARY_PARSE_ERROR err=%q\n", err)
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
	}
	return Run(cfg, os.Stdout, os.Stderr)
}
