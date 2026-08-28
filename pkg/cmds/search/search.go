package search

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"go.devnw.com/canary/pkg/cmds/internal/utils"
	"go.devnw.com/canary/pkg/storage"
)

// CANARY: REQ=CBIN-126; FEATURE="SearchCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var SearchCmd = &cobra.Command{
	Use:   "search <keywords>",
	Short: "Search CANARY tokens by keywords",
	Long: `Search tokens by keywords across feature names, requirement IDs, keyword tags, file paths, test names, and bench names.

Keywords are matched case-insensitively using LIKE queries.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt, _ := cmd.Flags().GetString("prompt")
		if prompt != "" {
			if _, err := utils.LoadPrompt(prompt); err != nil {
				return err
			}
		}
		dbPath, _ := cmd.Flags().GetString("db")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		limit, _ := cmd.Flags().GetInt("limit")
		var effLimit int
		if limit < 0 {
			effLimit = math.MaxInt32
		} else {
			effLimit = utils.EffectiveLimit(limit, defaultSearchLimit)
		}
		keywords := strings.Join(args, " ")

		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}

		defer func() { _ = db.Close() }()

		tokens, err := db.SearchTokens(keywords, effLimit)
		if err != nil {
			return fmt.Errorf("search tokens: %w", err)
		}

		if len(tokens) == 0 {
			fmt.Printf("No tokens found for: %s\n", keywords)
			return nil
		}

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			return enc.Encode(tokens)
		}

		fmt.Printf("Search results for '%s' (%d tokens):\n\n", keywords, len(tokens))
		for _, token := range tokens {
			fmt.Printf("📌 %s - %s\n", token.ReqID, token.Feature)
			fmt.Printf("   Status: %s | Priority: %d | %s:%d\n",
				token.Status, token.Priority, token.FilePath, token.LineNumber)
			if token.Keywords != "" {
				fmt.Printf("   Tags: %s\n", token.Keywords)
			}
			fmt.Println()
		}

		if effLimit > 0 && len(tokens) == effLimit {
			fmt.Printf("(showing %d; use --limit -1 for all)\n", effLimit)
		}

		return nil
	},
}

// CANARY: REQ=CBIN-205; FEATURE="ContextCaps"; ASPECT=CLI; STATUS=IMPL; UPDATED=2026-08-28
// defaultSearchLimit caps search output to protect agent context.
// Deliberately small; pass --limit -1 to explicitly request everything.
const defaultSearchLimit = 20

func init() {
	SearchCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	SearchCmd.Flags().String("db", ".canary/canary.db", "path to database file")
	SearchCmd.Flags().Bool("json", false, "output as JSON")
	SearchCmd.Flags().Int("limit", defaultSearchLimit, "maximum number of results (default 20 to protect agent context; -1 = unlimited)")
}
