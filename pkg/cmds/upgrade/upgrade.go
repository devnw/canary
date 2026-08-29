// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package upgrade wires the pkg/upgrade legacy-token rewriter into the CLI.
// CANARY: REQ=CP-275; FEATURE="TokenUpgrade"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_302_CLI_DryRunDefault,TestCANARY_CBIN_302_CLI_RuleFlag,TestCANARY_CBIN_302_CLI_MapFlag,TestCANARY_CBIN_302_CLI_InvalidRule; UPDATED=2026-08-29
package upgrade

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/upgrade"
)

// maxOldNewDisplay bounds how much of a Change's Old/New text is printed
// per line so a huge join-multiline block doesn't flood the terminal.
const maxOldNewDisplay = 100

// UpgradeCmd rewrites legacy on-disk CANARY token shapes into the current
// parseable form. Dry run by default; pass --write to modify files.
var UpgradeCmd = newUpgradeCmd()

// newUpgradeCmd builds a fresh *cobra.Command instance. Production code uses
// the single package-level UpgradeCmd; tests that Execute() the command
// multiple times with different flags call this directly, since pflag's
// StringArray flag value appends across repeated Set calls on the same
// FlagSet rather than resetting.
func newUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade [flags]",
		Short: "Rewrite legacy on-disk CANARY token shapes",
		Long: `Rewrite legacy on-disk CANARY token shapes into the current parseable form:
markdown "# CANARY:" headings, unicode hyphens inside IDs, unpadded flatfile
IDs, bare legacy ID segments, bug tokens missing FEATURE=, STATUS=FIXED,
missing UPDATED=, and the old multi-line bug-create continuation shape. With
--map, also remaps old requirement IDs to new ones across CANARY tokens and
GAP_ANALYSIS.md "✅ <ID>" claim lines.

Dry run by default — nothing is written until --write is passed.

Flags:
  --root <dir>     Root directory to scan (default ".")
  --write          Apply changes to disk (default: dry run)
  --map <file>     JSON object of {"OLD-ID": "NEW-ID", ...} for the remap rule
  --rule <name>    Restrict to one named rule (repeatable); default all rules
  --limit <n>      Bound how many changes are printed (default 20)

Rules: join-multiline, md-heading, unicode-hyphen, bare-id, bug-alias,
status-fixed, pad-flatfile, add-updated, remap

Examples:
  # Preview every legacy shape this repo has
  canary upgrade --root .

  # Apply the fixes
  canary upgrade --root . --write

  # Only fix STATUS=FIXED tokens
  canary upgrade --rule status-fixed --write

  # Remap old CBIN IDs to a new source's IDs
  canary upgrade --map id-map.json --write`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, _ := cmd.Flags().GetString("root")
			write, _ := cmd.Flags().GetBool("write")
			mapPath, _ := cmd.Flags().GetString("map")
			rules, _ := cmd.Flags().GetStringArray("rule")
			limit, _ := cmd.Flags().GetInt("limit")
			if root == "" {
				root = "."
			}
			if limit <= 0 {
				limit = 20
			}

			for _, r := range rules {
				if !upgrade.ValidRule(r) {
					return fmt.Errorf("unknown --rule %q (valid: %s)", r, strings.Join(upgrade.AllRules, ", "))
				}
			}

			var idMap map[string]string
			if mapPath != "" {
				b, err := os.ReadFile(mapPath)
				if err != nil {
					return fmt.Errorf("read --map %s: %w", mapPath, err)
				}
				if err := json.Unmarshal(b, &idMap); err != nil {
					return fmt.Errorf("parse --map %s: %w", mapPath, err)
				}
			}

			skip := canaryscan.DefaultSkipRegex()
			ignorePatterns, err := canaryscan.LoadCanaryIgnore(root)
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to load .canaryignore: %v\n", err)
			}
			reg := sources.LoadFromRoot(root)

			changes, err := upgrade.Run(upgrade.Options{
				Root:     root,
				Skip:     skip,
				Ignore:   ignorePatterns,
				Registry: reg,
				Map:      idMap,
				Write:    write,
				Rules:    rules,
			})
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			files := map[string]struct{}{}
			for i, c := range changes {
				files[c.File] = struct{}{}
				if i >= limit {
					continue
				}
				fmt.Fprintf(out, "%s:%d %s: %s → %s\n", c.File, c.Line, c.Rule, bound(c.Old), bound(c.New))
			}
			if len(changes) > limit {
				fmt.Fprintf(out, "... and %d more (raise with --limit)\n", len(changes)-limit)
			}
			fmt.Fprintf(out, "CANARY_UPGRADE files=%d changes=%d written=%t\n", len(files), len(changes), write)
			return nil
		},
	}

	cmd.Flags().String("root", ".", "root directory to scan")
	cmd.Flags().Bool("write", false, "apply changes to disk (default: dry run)")
	cmd.Flags().String("map", "", "JSON file mapping old REQ/BUG ids to new ids, for the remap rule")
	cmd.Flags().StringArray("rule", nil, "restrict to one named rule (repeatable); default all rules")
	cmd.Flags().Int("limit", 20, "bound how many changes are printed")
	return cmd
}

func bound(s string) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	if len(s) <= maxOldNewDisplay {
		return s
	}
	return s[:maxOldNewDisplay] + "…"
}
