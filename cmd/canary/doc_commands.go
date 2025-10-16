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

// CANARY: REQ=CBIN-136; FEATURE="DocCLICommands"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_136_CLI_DocWorkflow; DOC=user:docs/user/documentation-tracking-guide.md; DOC_HASH=1e32f44252c80284; UPDATED=2025-10-16

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.spyder.org/canary/internal/docs"
	"go.spyder.org/canary/internal/storage"
)

// CANARY: REQ=CBIN-136; FEATURE="DocParentCommand"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-16
var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "Documentation management commands",
	Long: `Manage documentation tracking, creation, and verification for CANARY requirements.

Documentation tracking ensures that each CANARY token references up-to-date documentation
files. The system uses SHA256 hashing to detect staleness and keep docs in sync with code.`,
	Example: `  # Create documentation from template
  canary doc create CBIN-105 --type user --output docs/user/authentication.md

  # Update documentation hash after editing
  canary doc update CBIN-105

  # Check documentation status
  canary doc status CBIN-105
  canary doc status --all`,
}

// CANARY: REQ=CBIN-136; FEATURE="DocCreateCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_136_CLI_DocCreate; UPDATED=2025-10-16
var docCreateCmd = &cobra.Command{
	Use:   "create <REQ-ID> --type <doc-type> --output <path>",
	Short: "Create documentation from template",
	Long: `Create a new documentation file from a template and link it to a requirement.

Supported documentation types:
  - user:         User-facing documentation
  - technical:    Technical design documentation
  - feature:      Feature specification documentation
  - api:          API reference documentation
  - architecture: Architecture decision records (ADR)

The command will:
1. Create the documentation file from the appropriate template
2. Update the CANARY token with DOC= field
3. Calculate and store the initial DOC_HASH=`,
	Example: `  canary doc create CBIN-105 --type user --output docs/user/auth.md
  canary doc create CBIN-200 --type api --output docs/api/rest.md`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reqID := strings.ToUpper(args[0])
		docType, _ := cmd.Flags().GetString("type")
		outputPath, _ := cmd.Flags().GetString("output")

		if docType == "" {
			return fmt.Errorf("--type flag is required (user, technical, feature, api, architecture)")
		}
		if outputPath == "" {
			return fmt.Errorf("--output flag is required")
		}

		// Validate doc type
		validTypes := map[string]bool{
			"user": true, "technical": true, "feature": true,
			"api": true, "architecture": true,
		}
		if !validTypes[docType] {
			return fmt.Errorf("invalid doc type: %s (must be user, technical, feature, api, or architecture)", docType)
		}

		// Load template
		templatePath := filepath.Join(".canary", "templates", "docs", docType+"-template.md")
		templateContent, err := os.ReadFile(templatePath)
		if err != nil {
			// If template doesn't exist, create a basic one
			templateContent = []byte(fmt.Sprintf(`# %s Documentation

**Requirement:** %s
**Type:** %s
**Created:** %s

## Overview

TODO: Provide an overview of this feature/component.

## Usage

TODO: Describe how to use this feature.

## Examples

TODO: Provide concrete examples.

## Notes

TODO: Additional notes, caveats, or considerations.
`, reqID, reqID, docType, time.Now().Format("2006-01-02")))
		}

		// Create output directory if needed
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Write documentation file
		if err := os.WriteFile(outputPath, templateContent, 0644); err != nil {
			return fmt.Errorf("failed to write documentation file: %w", err)
		}

		// Calculate hash
		hash, err := docs.CalculateHash(outputPath)
		if err != nil {
			return fmt.Errorf("failed to calculate hash: %w", err)
		}

		fmt.Printf("✅ Created documentation: %s\n", outputPath)
		fmt.Printf("   Requirement: %s\n", reqID)
		fmt.Printf("   Type: %s\n", docType)
		fmt.Printf("   Hash: %s\n", hash)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("  1. Edit the documentation file: %s\n", outputPath)
		fmt.Println("  2. Add DOC= field to your CANARY token:")
		fmt.Printf("     DOC=%s:%s; DOC_HASH=%s\n", docType, outputPath, hash)
		fmt.Println("  3. After editing, run: canary doc update", reqID)

		return nil
	},
}

// CANARY: REQ=CBIN-136; FEATURE="DocUpdateCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_136_CLI_BatchUpdate; UPDATED=2025-10-16
var docUpdateCmd = &cobra.Command{
	Use:   "update [REQ-ID]",
	Short: "Update documentation hash after changes",
	Long: `Recalculate documentation hashes for a requirement and update the database.

This command should be run after editing documentation files to update the
DOC_HASH field in the database, marking the documentation as current.

Batch Operations:
  --all            Update all documentation in the database
  --stale-only     Only update stale documentation (requires --all)`,
	Example: `  # Update specific requirement
  canary doc update CBIN-105

  # Update all documentation
  canary doc update --all

  # Update only stale documentation
  canary doc update --all --stale-only`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := cmd.Flag("db").Value.String()
		updateAll, _ := cmd.Flags().GetBool("all")
		staleOnly, _ := cmd.Flags().GetBool("stale-only")

		// Validate flags
		if staleOnly && !updateAll {
			return fmt.Errorf("--stale-only requires --all flag")
		}
		if len(args) == 0 && !updateAll {
			return fmt.Errorf("provide REQ-ID or use --all flag")
		}
		if len(args) > 0 && updateAll {
			return fmt.Errorf("cannot specify REQ-ID with --all flag")
		}

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		var tokens []*storage.Token

		if updateAll {
			// Get all tokens with documentation
			tokens, err = db.ListTokens(map[string]string{}, "", "req_id ASC", 0)
			if err != nil {
				return fmt.Errorf("failed to query tokens: %w", err)
			}
		} else {
			// Get tokens for specific requirement
			reqID := strings.ToUpper(args[0])
			tokens, err = db.GetTokensByReqID(reqID)
			if err != nil {
				return fmt.Errorf("failed to query tokens: %w", err)
			}
			if len(tokens) == 0 {
				return fmt.Errorf("no tokens found for requirement: %s", reqID)
			}
		}

		// Update documentation hashes
		updated := 0
		skipped := 0
		for _, token := range tokens {
			if token.DocPath == "" {
				continue
			}

			// If stale-only, check if documentation is stale first
			if staleOnly {
				results, err := docs.CheckMultipleDocumentation(token)
				if err != nil {
					fmt.Printf("⚠️  Error checking %s: %v\n", token.DocPath, err)
					continue
				}

				// Check if any docs are stale
				hasStale := false
				for _, status := range results {
					if status == "DOC_STALE" {
						hasStale = true
						break
					}
				}

				if !hasStale {
					skipped++
					continue
				}
			}

			// Handle multiple documentation paths (comma-separated)
			docPaths := strings.Split(token.DocPath, ",")
			newHashes := make([]string, 0, len(docPaths))

			for _, docPath := range docPaths {
				// Strip type prefix (e.g., "user:docs/file.md" -> "docs/file.md")
				docPath = strings.TrimSpace(docPath)
				actualPath := docPath
				if strings.Contains(docPath, ":") {
					parts := strings.SplitN(docPath, ":", 2)
					if len(parts) == 2 {
						actualPath = parts[1]
					}
				}

				// Recalculate hash
				newHash, err := docs.CalculateHash(actualPath)
				if err != nil {
					fmt.Printf("⚠️  Failed to calculate hash for %s: %v\n", docPath, err)
					continue
				}

				newHashes = append(newHashes, newHash)
				if updateAll {
					fmt.Printf("✅ %s: %s (hash: %s)\n", token.ReqID, docPath, newHash)
				} else {
					fmt.Printf("✅ Updated: %s (hash: %s)\n", docPath, newHash)
				}
			}

			if len(newHashes) == 0 {
				continue
			}

			// Update token with new hashes
			token.DocHash = strings.Join(newHashes, ",")
			token.DocCheckedAt = time.Now().UTC().Format(time.RFC3339)
			token.DocStatus = "DOC_CURRENT"

			if err := db.UpsertToken(token); err != nil {
				fmt.Printf("⚠️  Failed to update token: %v\n", err)
				continue
			}

			updated++
		}

		// Display summary
		if updateAll {
			fmt.Printf("\n✅ Updated %d requirement(s)", updated)
			if skipped > 0 {
				fmt.Printf(" (skipped %d current)", skipped)
			}
			fmt.Println()
		} else {
			if updated == 0 {
				fmt.Printf("No documentation files found\n")
				fmt.Println("Use 'canary doc create' to create documentation first.")
			} else {
				fmt.Printf("\n✅ Updated %d documentation file(s)\n", updated)
			}
		}

		return nil
	},
}

// CANARY: REQ=CBIN-136; FEATURE="DocStatusCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_136_CLI_DocStatus; UPDATED=2025-10-16
var docStatusCmd = &cobra.Command{
	Use:   "status [REQ-ID]",
	Short: "Check documentation staleness status",
	Long: `Check the staleness status of documentation for one or all requirements.

Status values:
  - DOC_CURRENT:  Documentation hash matches file content
  - DOC_STALE:    Documentation has been modified since last hash
  - DOC_MISSING:  Documentation file does not exist
  - DOC_UNHASHED: No hash tracking enabled for this documentation`,
	Example: `  canary doc status CBIN-105
  canary doc status --all
  canary doc status --stale-only`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := cmd.Flag("db").Value.String()
		showAll, _ := cmd.Flags().GetBool("all")
		staleOnly, _ := cmd.Flags().GetBool("stale-only")

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		var tokens []*storage.Token

		if len(args) == 1 {
			// Check specific requirement
			reqID := strings.ToUpper(args[0])
			tokens, err = db.GetTokensByReqID(reqID)
			if err != nil {
				return fmt.Errorf("failed to query tokens: %w", err)
			}
		} else if showAll {
			// Check all requirements with documentation
			tokens, err = db.ListTokens(map[string]string{}, "", "req_id ASC", 0)
			if err != nil {
				return fmt.Errorf("failed to query tokens: %w", err)
			}
		} else {
			return fmt.Errorf("provide REQ-ID or use --all flag")
		}

		// Check staleness for each token
		stats := map[string]int{
			"DOC_CURRENT":  0,
			"DOC_STALE":    0,
			"DOC_MISSING":  0,
			"DOC_UNHASHED": 0,
		}

		for _, token := range tokens {
			if token.DocPath == "" {
				continue
			}

			// Use CheckMultipleDocumentation to handle type prefixes and multiple paths
			results, err := docs.CheckMultipleDocumentation(token)
			if err != nil {
				fmt.Printf("⚠️  Error checking %s: %v\n", token.DocPath, err)
				continue
			}

			// Process each documentation path result
			for docPath, status := range results {
				stats[status]++

				// Filter output based on flags
				if staleOnly && status != "DOC_STALE" {
					continue
				}

				// Display result with full path (including type prefix if present)
				fullPath := docPath
				// Find the original path with type prefix
				for _, origPath := range strings.Split(token.DocPath, ",") {
					origPath = strings.TrimSpace(origPath)
					if strings.HasSuffix(origPath, docPath) {
						fullPath = origPath
						break
					}
				}

				emoji := "✅"
				if status == "DOC_STALE" {
					emoji = "⚠️"
				} else if status == "DOC_MISSING" {
					emoji = "❌"
				} else if status == "DOC_UNHASHED" {
					emoji = "ℹ️"
				}

				fmt.Printf("%s %s (%s): %s\n", emoji, token.ReqID, status, fullPath)
			}
		}

		// Summary
		total := stats["DOC_CURRENT"] + stats["DOC_STALE"] + stats["DOC_MISSING"] + stats["DOC_UNHASHED"]
		if total > 0 {
			fmt.Println()
			fmt.Printf("Summary: %d total\n", total)
			if stats["DOC_CURRENT"] > 0 {
				fmt.Printf("  ✅ Current:  %d\n", stats["DOC_CURRENT"])
			}
			if stats["DOC_STALE"] > 0 {
				fmt.Printf("  ⚠️  Stale:    %d\n", stats["DOC_STALE"])
			}
			if stats["DOC_MISSING"] > 0 {
				fmt.Printf("  ❌ Missing:  %d\n", stats["DOC_MISSING"])
			}
			if stats["DOC_UNHASHED"] > 0 {
				fmt.Printf("  ℹ️  Unhashed: %d\n", stats["DOC_UNHASHED"])
			}
		}

		return nil
	},
}

// CANARY: REQ=CBIN-136; FEATURE="DocReportCommand"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_136_CLI_DocReport; UPDATED=2025-10-16
var docReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate documentation coverage and staleness report",
	Long: `Generate comprehensive report on documentation coverage and health.

The report includes:
  - Total documentation count by type (user, api, technical, feature, architecture)
  - Staleness statistics (current, stale, missing, unhashed)
  - Coverage percentage (requirements with vs without documentation)
  - Requirements without documentation
  - Documentation age metrics`,
	Example: `  canary doc report
  canary doc report --format json
  canary doc report --show-undocumented`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := cmd.Flag("db").Value.String()
		format, _ := cmd.Flags().GetString("format")
		showUndocumented, _ := cmd.Flags().GetBool("show-undocumented")

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		// Get all tokens
		tokens, err := db.ListTokens(map[string]string{}, "", "req_id ASC", 0)
		if err != nil {
			return fmt.Errorf("failed to query tokens: %w", err)
		}

		// Statistics
		stats := struct {
			TotalTokens          int
			TokensWithDocs       int
			TokensWithoutDocs    int
			ByType               map[string]int
			ByStatus             map[string]int
			UndocumentedRequirements []string
		}{
			ByType:   make(map[string]int),
			ByStatus: make(map[string]int),
			UndocumentedRequirements: []string{},
		}

		seenRequirements := make(map[string]bool)
		requirementsWithDocs := make(map[string]bool)

		// Analyze tokens
		for _, token := range tokens {
			stats.TotalTokens++

			// Track unique requirements
			if !seenRequirements[token.ReqID] {
				seenRequirements[token.ReqID] = true
			}

			if token.DocPath == "" {
				continue
			}

			stats.TokensWithDocs++
			requirementsWithDocs[token.ReqID] = true

			// Count by type
			if token.DocType != "" {
				stats.ByType[token.DocType]++
			}

			// Check staleness for each documentation
			results, err := docs.CheckMultipleDocumentation(token)
			if err != nil {
				continue
			}

			for _, status := range results {
				stats.ByStatus[status]++
			}
		}

		// Find undocumented requirements
		for reqID := range seenRequirements {
			if !requirementsWithDocs[reqID] {
				stats.UndocumentedRequirements = append(stats.UndocumentedRequirements, reqID)
			}
		}
		stats.TokensWithoutDocs = len(stats.UndocumentedRequirements)

		// Output report
		if format == "json" {
			// JSON format output
			report := map[string]interface{}{
				"total_tokens":       stats.TotalTokens,
				"tokens_with_docs":   stats.TokensWithDocs,
				"tokens_without_docs": stats.TokensWithoutDocs,
				"coverage_percent":   float64(stats.TokensWithDocs) / float64(stats.TotalTokens) * 100,
				"by_type":            stats.ByType,
				"by_status":          stats.ByStatus,
				"undocumented_count": len(stats.UndocumentedRequirements),
			}
			if showUndocumented {
				report["undocumented_requirements"] = stats.UndocumentedRequirements
			}
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(report)
		}

		// Human-readable format
		fmt.Println("📊 Documentation Report")
		fmt.Println()

		// Coverage summary
		coveragePercent := 0.0
		if stats.TotalTokens > 0 {
			coveragePercent = float64(len(requirementsWithDocs)) / float64(len(seenRequirements)) * 100
		}
		fmt.Printf("Coverage: %d/%d requirements (%.1f%%)\n",
			len(requirementsWithDocs), len(seenRequirements), coveragePercent)
		fmt.Printf("Total Tokens: %d (%d with docs, %d without)\n\n",
			stats.TotalTokens, stats.TokensWithDocs, stats.TokensWithoutDocs)

		// Documentation by type
		if len(stats.ByType) > 0 {
			fmt.Println("📚 Documentation by Type:")
			for docType, count := range stats.ByType {
				fmt.Printf("  %s: %d\n", docType, count)
			}
			fmt.Println()
		}

		// Staleness statistics
		totalDocs := stats.ByStatus["DOC_CURRENT"] + stats.ByStatus["DOC_STALE"] +
			stats.ByStatus["DOC_MISSING"] + stats.ByStatus["DOC_UNHASHED"]

		if totalDocs > 0 {
			fmt.Println("📋 Documentation Status:")
			if stats.ByStatus["DOC_CURRENT"] > 0 {
				fmt.Printf("  ✅ Current:  %d (%.1f%%)\n",
					stats.ByStatus["DOC_CURRENT"],
					float64(stats.ByStatus["DOC_CURRENT"])/float64(totalDocs)*100)
			}
			if stats.ByStatus["DOC_STALE"] > 0 {
				fmt.Printf("  ⚠️  Stale:    %d (%.1f%%)\n",
					stats.ByStatus["DOC_STALE"],
					float64(stats.ByStatus["DOC_STALE"])/float64(totalDocs)*100)
			}
			if stats.ByStatus["DOC_MISSING"] > 0 {
				fmt.Printf("  ❌ Missing:  %d (%.1f%%)\n",
					stats.ByStatus["DOC_MISSING"],
					float64(stats.ByStatus["DOC_MISSING"])/float64(totalDocs)*100)
			}
			if stats.ByStatus["DOC_UNHASHED"] > 0 {
				fmt.Printf("  ℹ️  Unhashed: %d (%.1f%%)\n",
					stats.ByStatus["DOC_UNHASHED"],
					float64(stats.ByStatus["DOC_UNHASHED"])/float64(totalDocs)*100)
			}
			fmt.Println()
		}

		// Undocumented requirements
		if showUndocumented && len(stats.UndocumentedRequirements) > 0 {
			fmt.Printf("📝 Undocumented Requirements (%d):\n", len(stats.UndocumentedRequirements))
			for _, reqID := range stats.UndocumentedRequirements {
				fmt.Printf("  - %s\n", reqID)
			}
			fmt.Println()
		} else if len(stats.UndocumentedRequirements) > 0 {
			fmt.Printf("💡 %d requirements without documentation (use --show-undocumented to list)\n\n",
				len(stats.UndocumentedRequirements))
		}

		// Recommendations
		if stats.ByStatus["DOC_STALE"] > 0 {
			fmt.Println("💡 Recommendations:")
			fmt.Println("  Run 'canary doc update --all --stale-only' to update stale documentation")
		}

		return nil
	},
}

func init() {
	// Add sub-commands to docCmd
	docCmd.AddCommand(docCreateCmd)
	docCmd.AddCommand(docUpdateCmd)
	docCmd.AddCommand(docStatusCmd)
	docCmd.AddCommand(docReportCmd)

	// docCreateCmd flags
	docCreateCmd.Flags().String("type", "", "Documentation type (user, technical, feature, api, architecture)")
	docCreateCmd.Flags().String("output", "", "Output path for documentation file")

	// docUpdateCmd flags
	docUpdateCmd.Flags().String("db", ".canary/canary.db", "path to database file")
	docUpdateCmd.Flags().Bool("all", false, "Update all documentation in database")
	docUpdateCmd.Flags().Bool("stale-only", false, "Only update stale documentation (requires --all)")

	// docStatusCmd flags
	docStatusCmd.Flags().String("db", ".canary/canary.db", "path to database file")
	docStatusCmd.Flags().Bool("all", false, "Check all requirements")
	docStatusCmd.Flags().Bool("stale-only", false, "Show only stale documentation")

	// docReportCmd flags
	docReportCmd.Flags().String("db", ".canary/canary.db", "path to database file")
	docReportCmd.Flags().String("format", "text", "Output format (text or json)")
	docReportCmd.Flags().Bool("show-undocumented", false, "Show list of undocumented requirements")
}
