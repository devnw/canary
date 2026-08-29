// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package drift wires the pkg/drift engine into `canary drift`: a fresh
// scan (matching `canary scan`'s own default-skip + .canaryignore
// behavior), drift detection, and a bounded, agent-friendly report.
// CANARY: REQ=CP-278; FEATURE="DriftCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_305_Execute_CodeDrift,TestCANARY_CBIN_305_SummaryLine_Format,TestCANARY_CBIN_305_Cmd_JSON,TestCANARY_CBIN_305_StrictShouldFail,TestCANARY_CBIN_305_Cmd_Strict_NoFindings_NoError,TestCANARY_CBIN_305_Cmd_Table_NoDrift,TestCANARY_CBIN_305_Cmd_TimestampEnv_StaleFlips; UPDATED=2026-08-29
package drift

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/drift"
	"devnw.dev/canary/pkg/sources"
)

// DefaultLimit bounds the requirement-grouped table by default; raise it
// with --limit when more is needed.
const DefaultLimit = 20

// scanAndDetect runs a fresh scan of root (matching `canary scan`'s default
// skip regex + .canaryignore behavior) and returns drift findings. staleDays
// <= 0 falls back to .canary/project.yaml's verification.staleness_days,
// then canaryscan.DefaultStaleDays — the same precedence `canary scan
// --strict` uses.
func scanAndDetect(root string, staleDays int) ([]drift.Finding, error) {
	if root == "" {
		root = "."
	}
	skip := canaryscan.DefaultSkipRegex()
	ignorePatterns, err := canaryscan.LoadCanaryIgnore(root)
	if err != nil {
		return nil, fmt.Errorf("load .canaryignore: %w", err)
	}

	rep, err := canaryscan.Scan(root, skip, nil, ignorePatterns)
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	canaryscan.AnnotateSources(&rep, sources.LoadFromRoot(root))

	refTime := canaryscan.RefTimeFromEnv()
	if refTime.IsZero() {
		refTime = time.Now().UTC()
	}

	return drift.Detect(root, rep, resolveStaleDays(root, staleDays), refTime)
}

// resolveStaleDays mirrors canary scan's staleness-window precedence:
// the explicit flag, then .canary/project.yaml's verification.staleness_days,
// then canaryscan.DefaultStaleDays.
func resolveStaleDays(root string, flag int) int {
	if flag > 0 {
		return flag
	}
	if projCfg, err := canaryscan.LoadProjectConfig(root); err == nil && projCfg != nil && projCfg.Verification.StalenessDays > 0 {
		return projCfg.Verification.StalenessDays
	}
	return canaryscan.DefaultStaleDays
}

// summary holds the drift counts shown in the CANARY_DRIFT line and the JSON
// summary object. Requirements counts unique requirement IDs that have at
// least one finding (not the total requirements scanned).
type summary struct {
	Requirements int `json:"requirements"`
	CodeDrift    int `json:"code_drift"`
	Stale        int `json:"stale"`
	DocDrift     int `json:"doc_drift"`
}

func summarize(findings []drift.Finding) summary {
	var s summary
	reqs := map[string]struct{}{}
	for _, f := range findings {
		reqs[f.ReqID] = struct{}{}
		switch f.Kind {
		case drift.KindCodeDrift:
			s.CodeDrift++
		case drift.KindStale:
			s.Stale++
		case drift.KindDocDrift:
			s.DocDrift++
		}
	}
	s.Requirements = len(reqs)
	return s
}

// summaryLine returns the one-line CANARY_DRIFT summary agents can read from
// stdout without parsing the table or JSON.
func summaryLine(findings []drift.Finding) string {
	s := summarize(findings)
	return fmt.Sprintf("CANARY_DRIFT requirements=%d code_drift=%d stale=%d doc_drift=%d",
		s.Requirements, s.CodeDrift, s.Stale, s.DocDrift)
}

// strictShouldFail is the pure --strict decision: fail when any drift
// finding exists. Kept separate from the os.Exit(2) side effect in RunE so
// the decision itself stays unit-testable (mirrors scan.go's exit-code
// pattern, where the os.Exit branch is untested by design).
func strictShouldFail(findings []drift.Finding) bool {
	return len(findings) > 0
}

type reqGroup struct {
	ReqID    string
	Findings []drift.Finding
}

// groupByReq buckets findings by requirement ID, sorted by ID; findings
// within a group keep Kind then File order for stable output.
func groupByReq(findings []drift.Finding) []reqGroup {
	byReq := map[string][]drift.Finding{}
	for _, f := range findings {
		byReq[f.ReqID] = append(byReq[f.ReqID], f)
	}
	groups := make([]reqGroup, 0, len(byReq))
	for reqID, fs := range byReq {
		sort.Slice(fs, func(i, j int) bool {
			if fs[i].Kind != fs[j].Kind {
				return fs[i].Kind < fs[j].Kind
			}
			return fs[i].File < fs[j].File
		})
		groups = append(groups, reqGroup{ReqID: reqID, Findings: fs})
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].ReqID < groups[j].ReqID })
	return groups
}

// CreateDriftCommand returns the `canary drift` command.
func CreateDriftCommand() *cobra.Command {
	var jsonOut bool
	var limit int
	var root string
	var staleDays int
	var strict bool

	cmd := &cobra.Command{
		Use:   "drift [flags]",
		Short: "Detect code-vs-token drift: commits after UPDATED, staleness, and doc rollup",
		Long: `Scan the repo fresh and compare CANARY tokens against reality:

  code-drift  a token's file was committed (per git history) after its
              UPDATED date — the code moved on but the token wasn't refreshed.
  stale       TESTED/BENCHED tokens older than the staleness window
              (same rule as 'canary scan --strict').
  doc-drift   when .canary/canary.db exists, tokens whose tracked
              documentation is DOC_STALE or DOC_MISSING.

Flags:
  --root <dir>        Root directory to scan (default ".")
  --json               Compact JSON: {findings:[...], summary:{...}}
  --stale-days <n>     Staleness window in days (0 = project.yaml, else 30)
  --strict             Return a non-zero-carrying error when any finding exists
  --limit <n>          Max requirement groups shown in the table (default 20)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			findings, err := scanAndDetect(root, staleDays)
			if err != nil {
				return err
			}

			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				if err := enc.Encode(struct {
					Findings []drift.Finding `json:"findings"`
					Summary  summary         `json:"summary"`
				}{Findings: findings, Summary: summarize(findings)}); err != nil {
					return err
				}
			} else {
				printTable(cmd, findings, limit)
			}
			fmt.Fprintln(cmd.OutOrStdout(), summaryLine(findings))

			if strict && strictShouldFail(findings) {
				fmt.Fprintf(cmd.ErrOrStderr(), "CANARY_DRIFT_FAIL count=%d\n", len(findings))
				os.Exit(2)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&root, "root", ".", "root directory to scan")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "compact JSON output")
	cmd.Flags().IntVar(&staleDays, "stale-days", 0, "staleness window in days (0 = use .canary/project.yaml verification.staleness_days, else 30)")
	cmd.Flags().BoolVar(&strict, "strict", false, "return an error when any drift finding exists")
	cmd.Flags().IntVar(&limit, "limit", DefaultLimit, "max requirement groups shown in the table (raise when you need more)")
	return cmd
}

func printTable(cmd *cobra.Command, findings []drift.Finding, limit int) {
	out := cmd.OutOrStdout()
	if len(findings) == 0 {
		fmt.Fprintln(out, "No drift detected")
		return
	}
	if limit <= 0 {
		limit = DefaultLimit
	}

	groups := groupByReq(findings)
	total := len(groups)
	shown := groups
	if len(shown) > limit {
		shown = shown[:limit]
	}

	for _, g := range shown {
		fmt.Fprintf(out, "%s\n", g.ReqID)
		for _, f := range g.Findings {
			if f.File != "" {
				fmt.Fprintf(out, "  [%s] %s: %s\n", f.Kind, f.File, f.Detail)
			} else {
				fmt.Fprintf(out, "  [%s] %s\n", f.Kind, f.Detail)
			}
		}
	}
	if total > len(shown) {
		fmt.Fprintf(out, "… +%d more requirement(s) with findings (use --limit %d)\n", total-len(shown), total)
	}
}
