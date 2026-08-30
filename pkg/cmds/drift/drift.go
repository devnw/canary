// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package drift wires the pkg/drift engine into `canary drift`: it reads the
// token index built by `canary index` and reports, per requirement, whether
// the tree still matches the baseline — CURRENT, DRIFTED, or UNKNOWN — decided
// by content hash and git availability, never by a token's UPDATED= date.
// Staleness and documentation rollup are printed as a clearly separate
// advisory section and never affect the drift verdict or the exit code.
// CANARY: REQ=CP-278; FEATURE="DriftCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CP_278_Cmd_JSON_Clean,TestCANARY_CP_278_Cmd_JSON_Drifted,TestCANARY_CP_278_Cmd_Table_NoDrift,TestCANARY_CP_278_Cmd_NoIndex,TestCANARY_CP_278_SummaryLine_Format,TestCANARY_CP_278_StrictShouldFail; UPDATED=2026-08-30
package drift

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/cmds/internal/utils"
	"devnw.dev/canary/pkg/drift"
	"devnw.dev/canary/pkg/sources"
)

// DefaultLimit bounds the requirement list in table mode by default; raise it
// with --limit when more is needed.
const DefaultLimit = 20

// formatJSON and formatTable are the two --format values.
const (
	formatJSON  = "json"
	formatTable = "table"
)

// summaryLine returns the one-line CANARY_DRIFT summary agents can read from
// stdout without parsing the table: how many requirements landed in each
// drift state.
func summaryLine(states []drift.ReqState) string {
	var current, drifted, unknown int
	for _, s := range states {
		switch s.State {
		case drift.StateCurrent:
			current++
		case drift.StateDrifted:
			drifted++
		case drift.StateUnknown:
			unknown++
		}
	}
	return fmt.Sprintf("CANARY_DRIFT requirements=%d current=%d drifted=%d unknown=%d",
		len(states), current, drifted, unknown)
}

// strictShouldFail is the pure --strict decision: fail when any requirement is
// not CURRENT (DRIFTED or UNKNOWN). Kept separate from the os.Exit(2) side
// effect in RunE so the decision itself stays unit-testable.
func strictShouldFail(states []drift.ReqState) bool {
	for _, s := range states {
		if s.State != drift.StateCurrent {
			return true
		}
	}
	return false
}

// notCurrentCount is the number of requirements a --strict run fails on.
func notCurrentCount(states []drift.ReqState) int {
	n := 0
	for _, s := range states {
		if s.State != drift.StateCurrent {
			n++
		}
	}
	return n
}

// CreateDriftCommand returns the `canary drift` command.
func CreateDriftCommand() *cobra.Command {
	var root string
	var dbPath string
	var format string
	var limit int
	var strict bool

	cmd := &cobra.Command{
		Use:   "drift [flags]",
		Short: "Report per-requirement drift: CURRENT | DRIFTED | UNKNOWN, by content hash",
		Long: `Compare the indexed baseline against the tree and report, per requirement,
whether it still matches:

  CURRENT   every file's content hash matches the index baseline and git
            could confirm each file's history.
  DRIFTED   at least one file's content hash differs from the baseline — the
            code moved on since 'canary index' last recorded it.
  UNKNOWN   the verdict could not be decided: a missing baseline, an
            unreadable file, or git unable to answer for a file whose hash
            otherwise matches. A git failure is UNKNOWN, never CURRENT.

The verdict is evidence-based; it is never decided by a token's UPDATED= date.
Staleness and documentation rollup are printed as a separate advisory section
(table mode only) and never change the verdict or the exit code.

Flags:
  --root <dir>     Root directory whose tree is checked (default ".")
  --db <path>      Index database (default "<root>/.canary/canary.db")
  --project <id>   Scope to one project id (default: all projects)
  --format <fmt>   Output format: table (default) or json (a ReqState array)
  --strict         Exit non-zero when any requirement is DRIFTED or UNKNOWN
  --limit <n>      Max requirements shown in the table (default 20)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if root == "" {
				root = "."
			}
			resolvedDB := dbPath
			if resolvedDB == "" {
				resolvedDB = filepath.Join(root, ".canary", "canary.db")
			}

			db, err := utils.OpenIndexRO(cmd, resolvedDB)
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()

			projectID := utils.ReadProjectID(cmd)
			states, err := drift.Check(root, db, projectID)
			if err != nil {
				return utils.GuardContract(cmd, err)
			}

			if format == formatJSON {
				if states == nil {
					states = []drift.ReqState{}
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				if err := enc.Encode(states); err != nil {
					return err
				}
			} else {
				printStates(cmd, states, limit)
				if err := printAdvisories(cmd, root); err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), summaryLine(states))
			}

			if strict && strictShouldFail(states) {
				fmt.Fprintf(cmd.ErrOrStderr(), "CANARY_DRIFT_FAIL count=%d\n", notCurrentCount(states))
				os.Exit(2)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&root, "root", ".", "root directory whose tree is checked")
	cmd.Flags().StringVar(&dbPath, "db", "", "index database (default \"<root>/.canary/canary.db\")")
	cmd.Flags().StringVar(&format, "format", formatTable, "output format: table or json")
	cmd.Flags().BoolVar(&strict, "strict", false, "exit non-zero when any requirement is DRIFTED or UNKNOWN")
	cmd.Flags().IntVar(&limit, "limit", DefaultLimit, "max requirements shown in the table (raise when you need more)")
	utils.AddProjectFlag(cmd)
	return cmd
}

// printStates renders the per-requirement drift verdict in table mode.
func printStates(cmd *cobra.Command, states []drift.ReqState, limit int) {
	out := cmd.OutOrStdout()
	if len(states) == 0 {
		fmt.Fprintln(out, "No requirements indexed")
		return
	}
	if limit <= 0 {
		limit = DefaultLimit
	}

	shown := states
	if len(shown) > limit {
		shown = shown[:limit]
	}
	for _, s := range shown {
		fmt.Fprintf(out, "%s: %s\n", s.RequirementID, s.State)
		for _, f := range s.Files {
			if f.State == drift.StateCurrent {
				continue
			}
			fmt.Fprintf(out, "  [%s] %s: %s\n", f.State, f.Path, f.Detail)
		}
	}
	if len(states) > len(shown) {
		fmt.Fprintf(out, "… +%d more requirement(s) (use --limit %d)\n", len(states)-len(shown), len(states))
	}
}

// printAdvisories runs a fresh scan and prints the separate, non-drift
// advisory section: staleness and documentation rollup. A scan or advisory
// error is returned; a clean tree prints nothing.
func printAdvisories(cmd *cobra.Command, root string) error {
	findings, err := scanAdvisories(root)
	if err != nil {
		return err
	}
	if len(findings) == 0 {
		return nil
	}
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Advisories (not drift):")
	for _, g := range groupByReq(findings) {
		fmt.Fprintf(out, "%s\n", g.ReqID)
		for _, f := range g.Findings {
			if f.File != "" {
				fmt.Fprintf(out, "  [%s] %s: %s\n", f.Kind, f.File, f.Detail)
			} else {
				fmt.Fprintf(out, "  [%s] %s\n", f.Kind, f.Detail)
			}
		}
	}
	return nil
}

// scanAdvisories scans root (matching `canary scan`'s default skip regex +
// .canaryignore behavior) and returns staleness + doc-rollup findings. These
// are advisory only and never contribute to the drift verdict.
func scanAdvisories(root string) ([]drift.Finding, error) {
	if root == "" {
		root = "."
	}
	skip := canaryscan.DefaultSkipRegex()
	ignorePatterns, err := canaryscan.LoadCanaryIgnore(root)
	if err != nil {
		return nil, fmt.Errorf("load .canaryignore: %w", err)
	}
	reg, err := sources.LoadFromRoot(root)
	if err != nil {
		return nil, fmt.Errorf("load .canary/project.yaml: %w", err)
	}
	rep, err := canaryscan.Scan(root, skip, nil, ignorePatterns, reg)
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	canaryscan.AnnotateSources(&rep, reg)

	refTime := canaryscan.RefTimeFromEnv()
	if refTime.IsZero() {
		refTime = time.Now().UTC()
	}
	return drift.Advisories(root, rep, 0, refTime)
}

type reqGroup struct {
	ReqID    string
	Findings []drift.Finding
}

// groupByReq buckets advisory findings by requirement ID, sorted by ID;
// findings within a group keep Kind then File order for stable output.
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
