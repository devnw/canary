// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package show

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"devnw.dev/canary"
	"devnw.dev/canary/pkg/cmds/internal/utils"
	"devnw.dev/canary/pkg/storage"
)

// CANARY: REQ=CBIN-CLI-001; FEATURE="ShowCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_ShowCmd; UPDATED=2026-08-29
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
		prompt, _ := cmd.Flags().GetString("prompt")
		if prompt != "" {
			if _, err := utils.LoadPrompt(prompt); err != nil {
				return err
			}
		}
		reqID := args[0]
		groupBy, _ := cmd.Flags().GetString("group-by")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		// noColor flag retained for backward compatibility; color output removed in refactor
		_, _ = cmd.Flags().GetBool("no-color")

		dbPath, _ := cmd.Flags().GetString("db")

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

		// Query tokens
		tokens, err := db.GetTokensByReqID(projectID, reqID)
		if err != nil {
			return utils.GuardContract(err)
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
		output := canary.FormatTokensTable(tokens, groupBy)
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
func init() {
	ShowCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	ShowCmd.Flags().String("group-by", "aspect", "Group tokens by field (aspect, status)")
	ShowCmd.Flags().Bool("json", false, "Output in JSON format")
	ShowCmd.Flags().Bool("no-color", false, "Disable colored output")
	ShowCmd.Flags().String("db", ".canary/canary.db", "Path to database file")
	utils.AddProjectFlag(ShowCmd)
}
