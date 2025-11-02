package search

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/storage"
)

// CANARY: REQ=CBIN-126; FEATURE="SearchCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var SearchCmd = &cobra.Command{
	Use:   "search <keywords>",
	Short: "Search CANARY tokens by keywords",
	Long: `Search tokens by keywords in feature names, requirement IDs, and keyword tags.

Keywords are matched case-insensitively using LIKE queries.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement --prompt flag to load custom prompts
		prompt, _ := cmd.Flags().GetString("prompt")
		_ = prompt // Stubbed for future use

		dbPath, _ := cmd.Flags().GetString("db")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		keywords := strings.Join(args, " ")

		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}

		defer db.Close()

		tokens, err := db.SearchTokens(keywords)
		if err != nil {
			return fmt.Errorf("search tokens: %w", err)
		}

		if len(tokens) == 0 {
			fmt.Printf("No tokens found for: %s\n", keywords)
			return nil
		}

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
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

		return nil
	},
}

func init() {
	SearchCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	SearchCmd.Flags().String("db", ".canary/canary.db", "path to database file")
	SearchCmd.Flags().Bool("json", false, "output as JSON")
}
