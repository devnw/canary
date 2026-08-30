// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=ENG-4314; FEATURE="SpecModification"; ASPECT=CLI; STATUS=IMPL; DOC=user:docs/user/spec-modification-guide.md; DOC_HASH=676eb2a18c9d002a; UPDATED=2025-10-17
package specify

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/cmds/internal/utils"
	"devnw.dev/canary/pkg/specs"
	"devnw.dev/canary/pkg/storage"
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
				if db, dbErr := storage.OpenRO(dbPath); dbErr == nil {
					defer db.Close()
					projectID, pErr := utils.ReadProjectID(cmd, ".")
					if pErr != nil {
						return pErr
					}
					specPath, err = specs.FindSpecInDB(db, projectID, query)
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

	// Add updateCmd as subcommand of SpecifyCmd
	SpecifyCmd.AddCommand(updateCmd)
}
