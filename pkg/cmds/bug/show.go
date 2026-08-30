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

	"devnw.dev/canary/pkg/cmds/internal/utils"
)

var bugShowCmd = &cobra.Command{
	Use:   "show <BUG-ID>",
	Short: "Display details for a specific bug",
	Long: `Show detailed information about a specific BUG-* CANARY token.

Examples:
  canary bug show BUG-API-001
  canary bug show BUG-CLI-002 --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt, _ := cmd.Flags().GetString("prompt")
		if prompt != "" {
			if _, err := utils.LoadPrompt(prompt); err != nil {
				return err
			}
		}
		bugID := args[0]
		jsonOutput, _ := cmd.Flags().GetBool("json")
		dbPath, _ := cmd.Flags().GetString("db")

		// Validate bug ID format
		if !regexp.MustCompile(`^BUG-[A-Za-z]+-[0-9]{3}$`).MatchString(bugID) {
			return fmt.Errorf("invalid bug ID format: %s (expected BUG-ASPECT-XXX)", bugID)
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

		// Get bug token
		tokens, err := db.GetTokensByReqID(projectID, bugID)
		if err != nil {
			return utils.GuardContract(err)
		}
		if len(tokens) == 0 {
			return fmt.Errorf("bug not found: %s", bugID)
		}

		token := tokens[0]

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(token)
		}

		// Parse keywords for severity/priority
		severity, priority := parseBugMetadata(token.Keywords)

		// Format output
		fmt.Printf("?? Bug Details: %s\n\n", bugID)
		fmt.Printf("?? Title: %s\n", token.Feature)
		fmt.Printf("?? Status: %s | Aspect: %s\n", token.Status, token.Aspect)
		fmt.Printf("??  Severity: %s | Priority: %s\n", severity, priority)
		fmt.Printf("?? Location: %s:%d\n", token.FilePath, token.LineNumber)
		if token.Owner != "" {
			fmt.Printf("?? Owner: %s\n", token.Owner)
		}
		fmt.Printf("?? Updated: %s\n", token.UpdatedAt)
		if token.Test != "" {
			fmt.Printf("?? Test: %s\n", token.Test)
		}

		return nil
	},
}

func init() {
	bugShowCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	bugShowCmd.Flags().Bool("json", false, "Output in JSON format")
	bugShowCmd.Flags().String("db", ".canary/canary.db", "Path to database file")
	utils.AddProjectFlag(bugShowCmd)
}
