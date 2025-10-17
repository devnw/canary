// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"go.spyder.org/canary/internal/storage"
)

// CANARY: REQ=CBIN-CLI-001; FEATURE="FilesCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_FilesCmd; UPDATED=2025-10-16
var filesCmd = &cobra.Command{
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
		formatFilesList(fileGroups)

		return nil
	},
}

// formatFilesList formats file groups by aspect
func formatFilesList(fileGroups map[string][]*storage.Token) {
	// Group files by aspect
	aspectFiles := make(map[string][]string)
	fileCounts := make(map[string]int)

	for filePath, tokens := range fileGroups {
		// Get aspect from first token (all tokens in same file may have different aspects)
		aspects := make(map[string]bool)
		for _, token := range tokens {
			aspects[token.Aspect] = true
		}

		// Add file to each unique aspect
		for aspect := range aspects {
			aspectFiles[aspect] = append(aspectFiles[aspect], filePath)
		}

		fileCounts[filePath] = len(tokens)
	}

	// Sort aspects for consistent output
	var aspects []string
	for aspect := range aspectFiles {
		aspects = append(aspects, aspect)
	}
	sort.Strings(aspects)

	// Display by aspect
	for _, aspect := range aspects {
		files := aspectFiles[aspect]
		sort.Strings(files)

		fmt.Printf("**%s:**\n", aspect)
		for _, file := range files {
			count := fileCounts[file]
			plural := "token"
			if count > 1 {
				plural = "tokens"
			}
			fmt.Printf("  %s (%d %s)\n", file, count, plural)
		}
		fmt.Println()
	}

	// Summary
	totalFiles := len(fileGroups)
	totalTokens := 0
	for _, tokens := range fileGroups {
		totalTokens += len(tokens)
	}
	fmt.Printf("Total: %d files, %d tokens\n", totalFiles, totalTokens)
}

func init() {
	filesCmd.Flags().Bool("all", false, "Include spec and template files")
	filesCmd.Flags().String("db", ".canary/canary.db", "Path to database file")
}
