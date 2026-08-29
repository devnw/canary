package scan

import (
	"os"
	"regexp"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/canaryscan"
)

// CANARY: REQ=CP-240; FEATURE="ScanCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17
// ScanCmd uses the built-in canaryscan package so scan works from any CWD (e.g. after go install).
var ScanCmd = &cobra.Command{
	Use:   "scan [flags]",
	Short: "Scan for CANARY tokens and generate reports",
	Long: `Scan source code for CANARY tokens and generate status reports.

This command scans your codebase for CANARY tokens and generates JSON/CSV reports.
Uses the built-in scanner; works from any directory when canary is installed via go install.

Flags:
  --root <dir>            Root directory to scan (default ".")
  --out <file>            Output status.json path (default "status.json")
  --csv <file>            Optional status.csv path
  --verify <file>         GAP_ANALYSIS file to verify claims
  --strict                Enforce staleness on TESTED/BENCHED tokens (default 30 days; see --stale-days)
  --update-stale          Rewrite UPDATED field for stale tokens
  --skip <regex>          Skip path regex (RE2)
  --project-only          Filter by project requirement ID pattern
  --stale-days <n>        Staleness window in days (0 = use .canary/project.yaml verification.staleness_days, else 30)

Examples:
  # Basic scan
  canary scan --root . --out status.json

  # Verify GAP_ANALYSIS.md claims
  canary scan --verify GAP_ANALYSIS.md

  # Update stale tokens
  canary scan --update-stale

  # Strict mode with staleness enforcement
  canary scan --strict`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rootDir, _ := cmd.Flags().GetString("root")
		out, _ := cmd.Flags().GetString("out")
		csv, _ := cmd.Flags().GetString("csv")
		verify, _ := cmd.Flags().GetString("verify")
		strict, _ := cmd.Flags().GetBool("strict")
		updateStale, _ := cmd.Flags().GetBool("update-stale")
		skipStr, _ := cmd.Flags().GetString("skip")
		projectOnly, _ := cmd.Flags().GetBool("project-only")
		staleDays, _ := cmd.Flags().GetInt("stale-days")

		skip := canaryscan.DefaultSkipRegex()
		if skipStr != "" {
			var err error
			skip, err = regexp.Compile(skipStr)
			if err != nil {
				return err
			}
		}

		cfg := canaryscan.Config{
			Root:        rootDir,
			Out:         out,
			CSV:         csv,
			VerifyPath:  verify,
			Strict:      strict,
			SkipRegex:   skip,
			UpdateStale: updateStale,
			ProjectOnly: projectOnly,
			StaleDays:   staleDays,
		}
		code := canaryscan.Run(cfg, os.Stdout, os.Stderr)
		if code != 0 {
			os.Exit(code)
		}
		return nil
	},
}

func init() {
	ScanCmd.Flags().String("root", ".", "root directory to scan")
	ScanCmd.Flags().String("out", "status.json", "output status.json path")
	ScanCmd.Flags().String("csv", "", "optional status.csv path")
	ScanCmd.Flags().String("verify", "", "GAP_ANALYSIS file to verify claims")
	ScanCmd.Flags().Bool("strict", false, "enforce staleness on TESTED/BENCHED tokens (default 30 days; see --stale-days)")
	ScanCmd.Flags().Bool("update-stale", false, "rewrite UPDATED field for stale tokens")
	ScanCmd.Flags().String("skip", "", "skip path regex (RE2); default excludes .git, node_modules, vendor, bin, etc.")
	ScanCmd.Flags().Bool("project-only", false, "filter by project requirement ID pattern")
	ScanCmd.Flags().Int("stale-days", 0, "staleness window in days (0 = use .canary/project.yaml verification.staleness_days, else 30)")
}
