// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package grep

import (
	"fmt"
	"math"

	"github.com/spf13/cobra"

	"devnw.dev/canary"
	"devnw.dev/canary/pkg/cmds/internal/utils"
)

// CANARY: REQ=CBIN-CLI-001; FEATURE="GrepCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_GrepCmd; UPDATED=2026-08-29
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

		projectID, err := utils.ReadProjectID(cmd, ".")
		if err != nil {
			return err
		}

		// Read-only: never creates the database it could not find.
		db, err := utils.OpenIndexRO(cmd, dbPath)
		if err != nil {
			return err
		}
		defer db.Close()

		// Search for matching tokens
		tokens, err := canary.GrepTokens(db, projectID, pattern, effLimit)
		if err != nil {
			return utils.GuardContract(err)
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

// CANARY: REQ=ENG-4323; FEATURE="ContextCaps"; ASPECT=CLI; STATUS=IMPL; UPDATED=2026-08-28
// defaultGrepLimit caps grep output to protect agent context. Deliberately
// small; pass --limit -1 to explicitly request everything.
const defaultGrepLimit = 20

// grepTokens searches for tokens matching the pattern
func init() {
	GrepCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	GrepCmd.Flags().String("db", ".canary/canary.db", "Path to database file")
	utils.AddProjectFlag(GrepCmd)
	GrepCmd.Flags().String("group-by", "none", "Group results (none, requirement)")
	GrepCmd.Flags().Int("limit", defaultGrepLimit, "maximum number of results (default 20 to protect agent context; -1 = unlimited)")
}
