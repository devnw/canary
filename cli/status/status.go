// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package status

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/storage"
)

// CANARY: REQ=CBIN-CLI-001; FEATURE="StatusCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_StatusCmd; UPDATED=2025-10-16
var StatusCmd = &cobra.Command{
	Use:   "status <REQ-ID>",
	Short: "Show implementation progress for a requirement",
	Long: `Status displays implementation progress summary for a requirement.

Shows:
- Total token count
- Breakdown by status (STUB, IMPL, TESTED, BENCHED)
- Completion percentage
- List of incomplete work

Examples:
  canary status CBIN-133
  canary status CBIN-133 --no-color`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement --prompt flag to load custom prompts
		prompt, _ := cmd.Flags().GetString("prompt")
		_ = prompt // Stubbed for future use

		reqID := args[0]
		noColor, _ := cmd.Flags().GetBool("no-color")
		dbPath, _ := cmd.Flags().GetString("db")

		// Disable colors if requested
		if noColor {
			color.NoColor = true
		}

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Database not found\n")
			fmt.Fprintf(os.Stderr, "   Suggestion: Run 'canary index' to build database\n\n")
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		// Query tokens
		tokens, err := db.GetTokensByReqID(reqID)
		if err != nil {
			return fmt.Errorf("query tokens: %w", err)
		}

		if len(tokens) == 0 {
			fmt.Printf("No tokens found for %s\n", reqID)
			return fmt.Errorf("requirement not found")
		}

		// Calculate statistics
		stats := calculateStats(tokens)

		// Display summary
		displayStatusSummary(reqID, stats, tokens)

		return nil
	},
}

// StatusStats holds progress statistics
type StatusStats struct {
	Total     int
	Stub      int
	Impl      int
	Tested    int
	Benched   int
	Completed int
}

// calculateStats computes statistics from tokens
func calculateStats(tokens []*storage.Token) *StatusStats {
	stats := &StatusStats{
		Total: len(tokens),
	}

	for _, token := range tokens {
		switch token.Status {
		case "STUB":
			stats.Stub++
		case "IMPL":
			stats.Impl++
		case "TESTED":
			stats.Tested++
			stats.Completed++
		case "BENCHED":
			stats.Benched++
			stats.Completed++
		}
	}

	return stats
}

// displayStatusSummary shows formatted progress summary
func displayStatusSummary(reqID string, stats *StatusStats, tokens []*storage.Token) {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Printf("Implementation Status for %s:\n\n", cyan(reqID))

	// Progress bar
	completionPct := 0
	if stats.Total > 0 {
		completionPct = (stats.Completed * 100) / stats.Total
	}

	fmt.Printf("Progress: %s\n\n", progressBar(completionPct, 40))

	// Statistics
	fmt.Printf("Total:     %d tokens\n", stats.Total)
	fmt.Printf("Completed: %s (%d%%)\n", green(fmt.Sprintf("%d", stats.Completed)), completionPct)
	fmt.Printf("In Progress:\n")
	fmt.Printf("  • IMPL:   %s\n", yellow(fmt.Sprintf("%d", stats.Impl)))
	fmt.Printf("  • STUB:   %s\n", red(fmt.Sprintf("%d", stats.Stub)))
	fmt.Printf("Status Breakdown:\n")
	fmt.Printf("  • TESTED: %s\n", green(fmt.Sprintf("%d", stats.Tested)))
	fmt.Printf("  • BENCHED: %s\n", green(fmt.Sprintf("%d", stats.Benched)))
	fmt.Println()

	// List incomplete work
	if stats.Stub > 0 || stats.Impl > 0 {
		fmt.Println("Incomplete Work:")
		for _, token := range tokens {
			if token.Status == "STUB" || token.Status == "IMPL" {
				statusColor := red
				if token.Status == "IMPL" {
					statusColor = yellow
				}
				fmt.Printf("  %s %s - %s\n",
					statusColor(token.Status),
					token.Feature,
					token.FilePath)
			}
		}
	} else {
		fmt.Println(green("✅ All features completed!"))
	}
}

// progressBar generates a text progress bar
func progressBar(pct int, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	filled := (pct * width) / 100
	empty := width - filled

	bar := "["
	for i := 0; i < filled; i++ {
		bar += "="
	}
	if filled < width {
		bar += ">"
	}
	for i := 0; i < empty-1; i++ {
		bar += " "
	}
	bar += fmt.Sprintf("] %d%%", pct)

	return bar
}

func init() {
	StatusCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	StatusCmd.Flags().Bool("no-color", false, "Disable colored output")
	StatusCmd.Flags().String("db", ".canary/canary.db", "Path to database file")
}
