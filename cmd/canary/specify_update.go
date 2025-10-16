// Copyright (c) 2024 by CodePros.
//
// This software is proprietary information of CodePros.
// Unauthorized use, copying, modification, distribution, and/or
// disclosure is strictly prohibited, except as provided under the terms
// of the commercial license agreement you have entered into with
// CodePros.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact CodePros at info@codepros.org.

// CANARY: REQ=CBIN-134; FEATURE="SpecModification"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-16
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.spyder.org/canary/internal/specs"
	"go.spyder.org/canary/internal/storage"
)

var updateCmd = &cobra.Command{
	Use:   "update <REQ-ID or search-query>",
	Short: "Update an existing requirement specification",
	Long: `Locate and update an existing CANARY requirement specification.

Supports exact ID lookup, fuzzy text search, and section-specific loading
to minimize context usage for AI agents.

Examples:
  canary specify update CBIN-134                     # Exact ID lookup
  canary specify update --search "spec mod"          # Fuzzy search
  canary specify update CBIN-134 --sections overview # Load specific sections`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		searchFlag, _ := cmd.Flags().GetBool("search")
		sectionsFlag, _ := cmd.Flags().GetStringSlice("sections")

		var specPath string
		var err error

		// Determine lookup method
		if searchFlag {
			// Fuzzy search mode
			matches, err := specs.FindSpecBySearch(query, 5)
			if err != nil {
				return fmt.Errorf("search specs: %w", err)
			}

			if len(matches) == 0 {
				return fmt.Errorf("no specs found matching: %s", query)
			}

			// Show matches
			fmt.Printf("Found %d matching specs:\n\n", len(matches))
			for i, match := range matches {
				fmt.Printf("  %d. %s - %s (Score: %d%%)\n",
					i+1, match.ReqID, match.FeatureName, match.Score)
			}

			// Auto-select if single strong match (>90%)
			if len(matches) == 1 || (matches[0].Score > 90 && (len(matches) == 1 || matches[0].Score-matches[1].Score > 20)) {
				specPath = filepath.Join(matches[0].SpecPath, "spec.md")
				fmt.Printf("\nAuto-selected: %s\n\n", matches[0].ReqID)
			} else {
				return fmt.Errorf("multiple matches found - please use exact REQ-ID for precision")
			}
		} else {
			// Exact ID lookup
			specPath, err = specs.FindSpecByID(query)
			if err != nil {
				// Try database fallback
				dbPath := ".canary/canary.db"
				if db, dbErr := storage.Open(dbPath); dbErr == nil {
					defer db.Close()
					specPath, err = specs.FindSpecInDB(db, query)
				}
			}

			if err != nil {
				return fmt.Errorf("spec not found: %w\n\nHint: Try fuzzy search with --search flag:\n  canary specify update --search \"%s\"", err, query)
			}
		}

		// Read spec content
		content, err := os.ReadFile(specPath)
		if err != nil {
			return fmt.Errorf("read spec: %w", err)
		}

		specContent := string(content)

		// Apply section filtering if requested
		if len(sectionsFlag) > 0 {
			specContent, err = specs.ParseSections(specContent, sectionsFlag)
			if err != nil {
				return fmt.Errorf("parse sections: %w\n\nHint: Use --sections with valid section names like: overview, user-stories, requirements", err)
			}
		}

		// Check for plan.md
		planPath := filepath.Join(filepath.Dir(specPath), "plan.md")
		hasPlan := false
		if _, err := os.Stat(planPath); err == nil {
			hasPlan = true
		}

		// Output results
		fmt.Printf("✅ Found specification: %s\n", specPath)
		if hasPlan {
			fmt.Printf("📋 Plan exists: %s\n", planPath)
		}

		// If sections were requested, show what was included
		if len(sectionsFlag) > 0 {
			fmt.Printf("📄 Sections: %v\n", sectionsFlag)
		}

		fmt.Printf("\n--- Spec Content ---\n\n")
		fmt.Println(specContent)

		if hasPlan && len(sectionsFlag) == 0 {
			fmt.Printf("\n💡 Tip: View plan with: cat %s\n", planPath)
		}

		return nil
	},
}

func init() {
	updateCmd.Flags().Bool("search", false, "use fuzzy search instead of exact ID lookup")
	updateCmd.Flags().StringSlice("sections", []string{}, "load only specific sections (comma-separated)")

	// Add updateCmd as subcommand of specifyCmd
	// specifyCmd is defined in main.go
	specifyCmd.AddCommand(updateCmd)
}
