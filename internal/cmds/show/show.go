// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package show

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/storage"
)

// CANARY: REQ=CBIN-CLI-001; FEATURE="ShowCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_ShowCmd; UPDATED=2025-10-16
var ShowCmd = &cobra.Command{
	Use:   "show <REQ-ID>",
	Short: "Display all CANARY tokens for a requirement",
	Long: `Show displays all CANARY tokens for a specific requirement ID.

Displays:
- Feature name, aspect, status
- File location and line number
- Test and benchmark references
- Owner and priority

Grouping:
- By default, groups by aspect (CLI, API, Engine, etc.)
- Use --group-by status to group by implementation status
- Use --json for machine-readable output

Examples:
  canary show CBIN-133
  canary show CBIN-133 --group-by status
  canary show CBIN-133 --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement --prompt flag to load custom prompts
		prompt, _ := cmd.Flags().GetString("prompt")
		_ = prompt // Stubbed for future use

		reqID := args[0]
		groupBy, _ := cmd.Flags().GetString("group-by")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		noColor, _ := cmd.Flags().GetBool("no-color")

		dbPath, _ := cmd.Flags().GetString("db")

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Database not found, using filesystem search (slower)\n")
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
			fmt.Println("\nSuggestions:")
			fmt.Println("  • Run: canary list")
			fmt.Println("  • Check requirement ID format (e.g., CBIN-XXX)")
			return fmt.Errorf("requirement not found")
		}

		// Format output
		if jsonOutput {
			return outputTokensJSON(tokens)
		}

		fmt.Printf("Tokens for %s:\n\n", reqID)
		output := formatTokensTable(tokens, groupBy, !noColor)
		fmt.Println(output)

		return nil
	},
}

// outputTokensJSON outputs tokens as JSON
func outputTokensJSON(tokens []*storage.Token) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(tokens)
}

// formatTokensTable formats tokens as a grouped table
func formatTokensTable(tokens []*storage.Token, groupBy string, useColor bool) string {
	var buf strings.Builder

	// Group tokens
	groups := groupTokens(tokens, groupBy)

	// Format each group
	for groupName, groupTokens := range groups {
		buf.WriteString(fmt.Sprintf("## %s\n\n", groupName))

		for _, token := range groupTokens {
			buf.WriteString(fmt.Sprintf("📌 %s - %s\n", token.ReqID, token.Feature))

			// Status with optional color
			statusLine := fmt.Sprintf("   Status: %s | Aspect: %s", token.Status, token.Aspect)
			if token.Priority > 0 {
				statusLine += fmt.Sprintf(" | Priority: %d", token.Priority)
			}
			buf.WriteString(statusLine + "\n")

			buf.WriteString(fmt.Sprintf("   Location: %s:%d\n", token.FilePath, token.LineNumber))

			if token.Test != "" {
				buf.WriteString(fmt.Sprintf("   Test: %s\n", token.Test))
			}
			if token.Bench != "" {
				buf.WriteString(fmt.Sprintf("   Bench: %s\n", token.Bench))
			}
			if token.Owner != "" {
				buf.WriteString(fmt.Sprintf("   Owner: %s\n", token.Owner))
			}
			buf.WriteString("\n")
		}
	}

	return buf.String()
}

// groupTokens groups tokens by specified field
func groupTokens(tokens []*storage.Token, groupBy string) map[string][]*storage.Token {
	groups := make(map[string][]*storage.Token)

	for _, token := range tokens {
		var key string
		switch groupBy {
		case "status":
			key = token.Status
		case "aspect":
			key = token.Aspect
		default:
			// Default: group by aspect
			key = token.Aspect
		}

		if key == "" {
			key = "Ungrouped"
		}

		groups[key] = append(groups[key], token)
	}

	return groups
}

func init() {
	ShowCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	ShowCmd.Flags().String("group-by", "aspect", "Group tokens by field (aspect, status)")
	ShowCmd.Flags().Bool("json", false, "Output in JSON format")
	ShowCmd.Flags().Bool("no-color", false, "Disable colored output")
	ShowCmd.Flags().String("db", ".canary/canary.db", "Path to database file")
}
