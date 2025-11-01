// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package bug

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/storage"
)

var bugUpdateCmd = &cobra.Command{
	Use:   "update <BUG-ID>",
	Short: "Update a BUG-* CANARY token's status",
	Long: `Update the status or other properties of a BUG-* CANARY token.

Examples:
  canary bug update BUG-API-001 --status FIXED
  canary bug update BUG-CLI-002 --status IN_PROGRESS --owner alice
  canary bug update BUG-Storage-003 --priority P0 --severity S1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement --prompt flag to load custom prompts
		prompt, _ := cmd.Flags().GetString("prompt")
		_ = prompt // Stubbed for future use

		bugID := args[0]
		status, _ := cmd.Flags().GetString("status")
		severity, _ := cmd.Flags().GetString("severity")
		priority, _ := cmd.Flags().GetString("priority")
		owner, _ := cmd.Flags().GetString("owner")
		dbPath, _ := cmd.Flags().GetString("db")

		// Validate bug ID format
		if !regexp.MustCompile(`^BUG-[A-Za-z]+-[0-9]{3}$`).MatchString(bugID) {
			return fmt.Errorf("invalid bug ID format: %s (expected BUG-ASPECT-XXX)", bugID)
		}

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		// Get existing token
		tokens, err := db.GetTokensByReqID(bugID)
		if err != nil {
			return fmt.Errorf("find bug: %w", err)
		}
		if len(tokens) == 0 {
			return fmt.Errorf("bug not found: %s", bugID)
		}

		token := tokens[0]
		updated := false

		// Update fields if provided
		if status != "" {
			token.Status = status
			updated = true
		}
		if owner != "" {
			token.Owner = owner
			updated = true
		}
		if priority != "" {
			token.Priority = parsePriorityValue(priority)
			updated = true
		}

		// Update keywords for severity/priority
		if severity != "" || priority != "" {
			metaParts := strings.Split(token.Keywords, ";")

			// Parse existing keywords
			existingSev := "S3"
			existingPri := "P2"
			for _, part := range metaParts {
				if strings.HasPrefix(part, "SEVERITY=") {
					existingSev = strings.TrimPrefix(part, "SEVERITY=")
				} else if strings.HasPrefix(part, "PRIORITY=") {
					existingPri = strings.TrimPrefix(part, "PRIORITY=")
				}
			}

			if severity != "" {
				existingSev = severity
			}
			if priority != "" {
				existingPri = priority
			}

			token.Keywords = fmt.Sprintf("SEVERITY=%s;PRIORITY=%s", existingSev, existingPri)
			updated = true
		}

		if !updated {
			fmt.Println("No changes specified")
			return nil
		}

		// Update timestamp
		token.UpdatedAt = time.Now().Format("2006-01-02")

		// Save updated token
		if err := db.UpsertToken(token); err != nil {
			return fmt.Errorf("update bug: %w", err)
		}

		fmt.Printf("? Updated bug %s\n", bugID)
		fmt.Printf("?? Status: %s\n", token.Status)
		if owner != "" {
			fmt.Printf("?? Owner: %s\n", token.Owner)
		}
		if token.Keywords != "" {
			fmt.Printf("?? Metadata: %s\n", token.Keywords)
		}

		return nil
	},
}

func init() {
	bugUpdateCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	bugUpdateCmd.Flags().String("status", "", "New status (OPEN, IN_PROGRESS, FIXED, etc.)")
	bugUpdateCmd.Flags().String("severity", "", "New severity (S1, S2, S3, S4)")
	bugUpdateCmd.Flags().String("priority", "", "New priority (P0, P1, P2, P3)")
	bugUpdateCmd.Flags().String("owner", "", "New owner/assignee")
	bugUpdateCmd.Flags().String("db", ".canary/canary.db", "Path to database file")
}
