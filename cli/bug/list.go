// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package bug

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/storage"
)

var bugListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all BUG-* CANARY tokens",
	Long: `List all BUG-* CANARY tokens with optional filtering.

Examples:
  canary bug list
  canary bug list --aspect API
  canary bug list --status OPEN --severity S1
  canary bug list --priority P0,P1
  canary bug list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement --prompt flag to load custom prompts
		prompt, _ := cmd.Flags().GetString("prompt")
		_ = prompt // Stubbed for future use

		aspect, _ := cmd.Flags().GetString("aspect")
		status, _ := cmd.Flags().GetString("status")
		severity, _ := cmd.Flags().GetString("severity")
		priority, _ := cmd.Flags().GetString("priority")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		noColor, _ := cmd.Flags().GetBool("no-color")
		limit, _ := cmd.Flags().GetInt("limit")
		dbPath, _ := cmd.Flags().GetString("db")

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			// Fallback to filesystem search if no database
			return listBugsFromFilesystem(aspect, status, severity, priority, jsonOutput, noColor, limit)
		}
		defer db.Close()

		// Build filters for BUG tokens
		filters := make(map[string]string)
		if aspect != "" {
			filters["aspect"] = aspect
		}
		if status != "" {
			filters["status"] = status
		}

		// Query database for all tokens (ListTokens is hardcoded for CBIN patterns)
		allTokens, err := db.ListTokens(filters, "", "priority ASC, updated_at DESC", 0)
		if err != nil {
			return fmt.Errorf("query bugs: %w", err)
		}

		// Filter for BUG tokens only
		var tokens []*storage.Token
		bugPattern := regexp.MustCompile(`^BUG-[A-Za-z]+-[0-9]{3}$`)
		for _, tok := range allTokens {
			if bugPattern.MatchString(tok.ReqID) {
				tokens = append(tokens, tok)
			}
		}

		// Additional filtering for severity and priority (stored in token comments or metadata)
		filteredTokens := filterBugTokens(tokens, severity, priority)

		// Apply limit if specified
		if limit > 0 && len(filteredTokens) > limit {
			filteredTokens = filteredTokens[:limit]
		}

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(filteredTokens)
		}

		// Format output
		if len(filteredTokens) == 0 {
			fmt.Println("No bug tokens found")
			return nil
		}

		formatBugList(filteredTokens, noColor)
		return nil
	},
}

func init() {
	bugListCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	bugListCmd.Flags().String("aspect", "", "Filter by aspect (API, CLI, Engine, Storage, etc.)")
	bugListCmd.Flags().String("status", "", "Filter by status (OPEN, IN_PROGRESS, FIXED, etc.)")
	bugListCmd.Flags().String("severity", "", "Filter by severity (S1, S2, S3, S4)")
	bugListCmd.Flags().String("priority", "", "Filter by priority (P0, P1, P2, P3)")
	bugListCmd.Flags().Bool("json", false, "Output in JSON format")
	bugListCmd.Flags().Bool("no-color", false, "Disable colored output")
	bugListCmd.Flags().Int("limit", 0, "Limit number of results (0 = unlimited)")
	bugListCmd.Flags().String("db", ".canary/canary.db", "Path to database file")
}
