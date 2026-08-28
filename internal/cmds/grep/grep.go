// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package grep

import (
	"fmt"
	"math"
	"os"

	"github.com/spf13/cobra"

	"go.devnw.com/canary"
	"go.devnw.com/canary/internal/cmds/internal/utils"
	"go.devnw.com/canary/internal/storage"
)

// CANARY: REQ=CBIN-CLI-001; FEATURE="GrepCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_GrepCmd; UPDATED=2025-10-16
var GrepCmd = &cobra.Command{
	Use:   "grep <pattern>",
	Short: "Search CANARY tokens by pattern",
	Long: `Search for CANARY tokens matching a pattern.

Searches across:
- Feature names
- File paths
- Test names
- Bench names
- Requirement IDs

The search is case-insensitive and matches substrings.

Examples:
  canary grep User              # Find all tokens related to "User"
  canary grep internal/auth     # Find tokens in auth directory
  canary grep TestAuth          # Find tokens with "TestAuth" test`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt, _ := cmd.Flags().GetString("prompt")
		if prompt != "" {
			if _, err := utils.LoadPrompt(prompt); err != nil {
				return err
			}
		}

		pattern := args[0]
		dbPath, _ := cmd.Flags().GetString("db")
		groupBy, _ := cmd.Flags().GetString("group-by")
		limit, _ := cmd.Flags().GetInt("limit")
		var effLimit int
		if limit < 0 {
			effLimit = math.MaxInt32
		} else {
			effLimit = utils.EffectiveLimit(limit, defaultGrepLimit)
		}

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Database not found\n")
			fmt.Fprintf(os.Stderr, "   Suggestion: Run 'canary index' to build database\n\n")
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		// Search for matching tokens
		tokens, err := canary.GrepTokens(db, pattern, effLimit)
		if err != nil {
			return fmt.Errorf("search tokens: %w", err)
		}

		if len(tokens) == 0 {
			fmt.Printf("No tokens found matching pattern: %s\n", pattern)
			return nil
		}

		// Display results
		fmt.Printf("Found %d tokens matching '%s':\n\n", len(tokens), pattern)

		if groupBy == "requirement" {
			fmt.Print(canary.FormatGrepResultsByRequirement(tokens))
		} else {
			fmt.Print(canary.FormatGrepResults(tokens))
		}

		if effLimit > 0 && len(tokens) == effLimit {
			fmt.Printf("(showing %d; use --limit -1 for all)\n", effLimit)
		}

		return nil
	},
}

// CANARY: REQ=CBIN-205; FEATURE="ContextCaps"; ASPECT=CLI; STATUS=IMPL; UPDATED=2026-08-28
// defaultGrepLimit caps grep output to protect agent context. Deliberately
// small; pass --limit -1 to explicitly request everything.
const defaultGrepLimit = 20

// grepTokens searches for tokens matching the pattern
func init() {
	GrepCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	GrepCmd.Flags().String("db", ".canary/canary.db", "Path to database file")
	GrepCmd.Flags().String("group-by", "none", "Group results (none, requirement)")
	GrepCmd.Flags().Int("limit", defaultGrepLimit, "maximum number of results (default 20 to protect agent context; -1 = unlimited)")
}
