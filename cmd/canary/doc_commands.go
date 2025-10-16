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

// CANARY: REQ=CBIN-136; FEATURE="DocCLICommands"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-16

package main

import (
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

// CANARY: REQ=CBIN-136; FEATURE="DocCreateCommand"; ASPECT=CLI; STATUS=IMPL; TEST=TestCANARY_CBIN_136_CLI_DocCreate; UPDATED=2025-10-16
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

// CANARY: REQ=CBIN-136; FEATURE="DocUpdateCommand"; ASPECT=CLI; STATUS=IMPL; TEST=TestCANARY_CBIN_136_CLI_DocUpdate; UPDATED=2025-10-16
var docUpdateCmd = &cobra.Command{
	Use:   "update <REQ-ID>",
	Short: "Update documentation hash after changes",
	Long: `Recalculate documentation hashes for a requirement and update the database.

This command should be run after editing documentation files to update the
DOC_HASH field in the database, marking the documentation as current.`,
	Example: `  canary doc update CBIN-105
  canary doc update CBIN-200`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reqID := strings.ToUpper(args[0])
		dbPath := cmd.Flag("db").Value.String()

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		// Get tokens for requirement
		tokens, err := db.GetTokensByReqID(reqID)
		if err != nil {
			return fmt.Errorf("failed to query tokens: %w", err)
		}

		if len(tokens) == 0 {
			return fmt.Errorf("no tokens found for requirement: %s", reqID)
		}

		updated := 0
		for _, token := range tokens {
			if token.DocPath == "" {
				continue
			}

			// Recalculate hash
			newHash, err := docs.CalculateHash(token.DocPath)
			if err != nil {
				fmt.Printf("⚠️  Failed to calculate hash for %s: %v\n", token.DocPath, err)
				continue
			}

			// Update token
			token.DocHash = newHash
			token.DocCheckedAt = time.Now().UTC().Format(time.RFC3339)
			token.DocStatus = "DOC_CURRENT"

			if err := db.UpsertToken(token); err != nil {
				fmt.Printf("⚠️  Failed to update token: %v\n", err)
				continue
			}

			fmt.Printf("✅ Updated: %s (hash: %s)\n", token.DocPath, newHash)
			updated++
		}

		if updated == 0 {
			fmt.Printf("No documentation files found for %s\n", reqID)
			fmt.Println("Use 'canary doc create' to create documentation first.")
		} else {
			fmt.Printf("\n✅ Updated %d documentation file(s)\n", updated)
		}

		return nil
	},
}

// CANARY: REQ=CBIN-136; FEATURE="DocStatusCommand"; ASPECT=CLI; STATUS=IMPL; TEST=TestCANARY_CBIN_136_CLI_DocStatus; UPDATED=2025-10-16
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

			status, err := docs.CheckStaleness(token)
			if err != nil {
				fmt.Printf("⚠️  Error checking %s: %v\n", token.DocPath, err)
				continue
			}

			stats[status]++

			// Filter output based on flags
			if staleOnly && status != "DOC_STALE" {
				continue
			}

			// Display result
			emoji := "✅"
			if status == "DOC_STALE" {
				emoji = "⚠️"
			} else if status == "DOC_MISSING" {
				emoji = "❌"
			} else if status == "DOC_UNHASHED" {
				emoji = "ℹ️"
			}

			fmt.Printf("%s %s (%s): %s\n", emoji, token.ReqID, status, token.DocPath)
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

func init() {
	// Add sub-commands to docCmd
	docCmd.AddCommand(docCreateCmd)
	docCmd.AddCommand(docUpdateCmd)
	docCmd.AddCommand(docStatusCmd)

	// docCreateCmd flags
	docCreateCmd.Flags().String("type", "", "Documentation type (user, technical, feature, api, architecture)")
	docCreateCmd.Flags().String("output", "", "Output path for documentation file")

	// docUpdateCmd flags
	docUpdateCmd.Flags().String("db", ".canary/canary.db", "path to database file")

	// docStatusCmd flags
	docStatusCmd.Flags().String("db", ".canary/canary.db", "path to database file")
	docStatusCmd.Flags().Bool("all", false, "Check all requirements")
	docStatusCmd.Flags().Bool("stale-only", false, "Show only stale documentation")
}
