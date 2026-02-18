// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package files

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.devnw.com/canary"
	"go.devnw.com/canary/internal/cmds/internal/utils"
	"go.devnw.com/canary/internal/storage"
)

// CANARY: REQ=CBIN-CLI-001; FEATURE="FilesCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_FilesCmd; UPDATED=2025-10-16
var FilesCmd = &cobra.Command{
	Use:   "files <REQ-ID>",
	Short: "List implementation files for a requirement",
	Long: `Files lists all implementation files containing tokens for a requirement.

By default, excludes spec and template files, showing only actual implementation.
Files are grouped by aspect and show token counts.

Examples:
  canary files CBIN-133
  canary files CBIN-133 --all  # Include spec/template files`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt, _ := cmd.Flags().GetString("prompt")
		if prompt != "" {
			if _, err := utils.LoadPrompt(prompt); err != nil {
				return err
			}
		}
		reqID := args[0]
		includeAll, _ := cmd.Flags().GetBool("all")
		dbPath, _ := cmd.Flags().GetString("db")

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Database not found\n")
			fmt.Fprintf(os.Stderr, "   Suggestion: Run 'canary index' to build database\n\n")
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		// Query file groups
		excludeSpecs := !includeAll
		fileGroups, err := db.GetFilesByReqID(reqID, excludeSpecs)
		if err != nil {
			return fmt.Errorf("query files: %w", err)
		}

		if len(fileGroups) == 0 {
			fmt.Printf("No implementation files found for %s\n", reqID)
			if !includeAll {
				fmt.Println("\nTip: Use --all to include spec/template files")
			}
			return fmt.Errorf("no files found")
		}

		// Format output
		fmt.Printf("Implementation files for %s:\n\n", reqID)
		fmt.Print(canary.FormatFilesList(fileGroups))

		return nil
	},
}

// formatFilesList formats file groups by aspect
func init() {
	FilesCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	FilesCmd.Flags().Bool("all", false, "Include spec and template files")
	FilesCmd.Flags().String("db", ".canary/canary.db", "Path to database file")
}
