package scan

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// CANARY: REQ=CBIN-111; FEATURE="ScanCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17
// ScanCmd wraps the existing tools/canary scanner
var ScanCmd = &cobra.Command{
	Use:   "scan [flags]",
	Short: "Scan for CANARY tokens and generate reports",
	Long: `Scan source code for CANARY tokens and generate status reports.

This command scans your codebase for CANARY tokens and generates JSON/CSV reports.

Flags:
  --root <dir>            Root directory to scan (default ".")
  --out <file>            Output status.json path (default "status.json")
  --csv <file>            Optional status.csv path
  --verify <file>         GAP_ANALYSIS file to verify claims
  --strict                Enforce staleness on TESTED/BENCHED tokens (30 days)
  --update-stale          Rewrite UPDATED field for stale tokens
  --skip <regex>          Skip path regex (RE2)
  --project-only          Filter by project requirement ID pattern

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
		// Build path to the canary scanner
		scanner := filepath.Join("tools", "canary", "main.go")

		// Get all flags
		rootDir, _ := cmd.Flags().GetString("root")
		out, _ := cmd.Flags().GetString("out")
		csv, _ := cmd.Flags().GetString("csv")
		verify, _ := cmd.Flags().GetString("verify")
		strict, _ := cmd.Flags().GetBool("strict")
		updateStale, _ := cmd.Flags().GetBool("update-stale")
		skip, _ := cmd.Flags().GetString("skip")
		projectOnly, _ := cmd.Flags().GetBool("project-only")

		// Build scanner arguments
		scanArgs := []string{"run", scanner}

		if rootDir != "" {
			scanArgs = append(scanArgs, "-root", rootDir)
		}
		if out != "" {
			scanArgs = append(scanArgs, "-out", out)
		}
		if csv != "" {
			scanArgs = append(scanArgs, "-csv", csv)
		}
		if verify != "" {
			scanArgs = append(scanArgs, "-verify", verify)
		}
		if strict {
			scanArgs = append(scanArgs, "-strict")
		}
		if updateStale {
			scanArgs = append(scanArgs, "-update-stale")
		}
		if skip != "" {
			scanArgs = append(scanArgs, "-skip", skip)
		}
		if projectOnly {
			scanArgs = append(scanArgs, "-project-only")
		}

		// Pass through any additional args
		scanArgs = append(scanArgs, args...)

		goCmd := exec.Command("go", scanArgs...)
		goCmd.Stdout = os.Stdout
		goCmd.Stderr = os.Stderr
		goCmd.Stdin = os.Stdin

		return goCmd.Run()
	},
}
