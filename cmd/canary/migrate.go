// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-145; FEATURE="MigrateCommand"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/migrate"
	"go.devnw.com/canary/internal/storage"
)

// CANARY: REQ=CBIN-145; FEATURE="MigrateParentCommand"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
var orphanCmd = &cobra.Command{
	Use:   "orphan",
	Short: "Manage orphaned CANARY tokens (tokens without specifications)",
	Long: `Generate specifications and plans for requirements with orphaned tokens.

Orphaned requirements are those that have CANARY tokens in the codebase but no
formal specification file in .canary/specs/. This typically happens when migrating
legacy code or when tokens were added without following the requirement-first workflow.

The orphan command:
1. Detects requirements with tokens but no specs
2. Generates spec.md from existing token metadata
3. Generates plan.md reflecting current implementation state
4. Filters out documentation/example tokens automatically`,
	Example: `  # Detect orphaned requirements
  canary orphan detect

  # Migrate a specific requirement
  canary orphan run CBIN-105

  # Migrate all orphaned requirements
  canary orphan run --all

  # Preview migration without creating files
  canary orphan run --all --dry-run`,
}

// CANARY: REQ=CBIN-145; FEATURE="MigrateDetectCommand"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
var migrateDetectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect orphaned requirements",
	Long: `List all requirements that have CANARY tokens but no specification files.

This command scans the database for tokens and checks if corresponding spec.md
files exist in .canary/specs/. Tokens in documentation directories are automatically
excluded.`,
	Example: `  canary migrate detect
  canary migrate detect --show-features`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := cmd.Flag("db").Value.String()
		rootDir, _ := cmd.Flags().GetString("root")
		showFeatures, _ := cmd.Flags().GetBool("show-features")

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		// Detect orphans with path filtering
		excludePaths := []string{"/docs/", "/.claude/", "/.cursor/", "/.canary/specs/"}
		orphans, err := migrate.DetectOrphans(db, rootDir, excludePaths)
		if err != nil {
			return fmt.Errorf("failed to detect orphans: %w", err)
		}

		if len(orphans) == 0 {
			fmt.Println("✅ No orphaned requirements found!")
			fmt.Println("All CANARY tokens have corresponding specifications.")
			return nil
		}

		fmt.Printf("🔍 Found %d orphaned requirement(s):\n\n", len(orphans))

		for _, orphan := range orphans {
			confidenceEmoji := "🟢"
			if orphan.Confidence == migrate.ConfidenceMedium {
				confidenceEmoji = "🟡"
			} else if orphan.Confidence == migrate.ConfidenceLow {
				confidenceEmoji = "🔴"
			}

			fmt.Printf("%s %s (Confidence: %s)\n", confidenceEmoji, orphan.ReqID, orphan.Confidence)
			fmt.Printf("   Features: %d\n", orphan.FeatureCount)

			if showFeatures {
				for _, token := range orphan.Features {
					fmt.Printf("   - %s (%s, %s) at %s:%d\n",
						token.Feature, token.Aspect, token.Status, token.FilePath, token.LineNumber)
				}
			}
			fmt.Println()
		}

		fmt.Printf("💡 To migrate these requirements:\n")
		fmt.Printf("   Single:  canary migrate <REQ-ID>\n")
		fmt.Printf("   All:     canary migrate --all\n")
		fmt.Printf("   Preview: canary migrate --all --dry-run\n")

		return nil
	},
}

// CANARY: REQ=CBIN-145; FEATURE="MigrateRunCommand"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
var migrateRunCmd = &cobra.Command{
	Use:   "run [REQ-ID]",
	Short: "Migrate orphaned requirements to specifications",
	Long: `Generate specification and plan files for orphaned requirements.

For each orphaned requirement:
1. Creates .canary/specs/REQ-ID-name/ directory
2. Generates spec.md from existing tokens
3. Generates plan.md reflecting implementation status
4. Marks requirement as having a specification

Use --all to migrate all orphaned requirements at once.
Use --dry-run to preview changes without creating files.`,
	Example: `  # Migrate single requirement
  canary migrate run CBIN-105

  # Migrate all orphaned requirements
  canary migrate run --all

  # Preview without creating files
  canary migrate run --all --dry-run`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := cmd.Flag("db").Value.String()
		rootDir, _ := cmd.Flags().GetString("root")
		migrateAll, _ := cmd.Flags().GetBool("all")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Validate arguments
		if len(args) == 0 && !migrateAll {
			return fmt.Errorf("provide REQ-ID or use --all flag")
		}
		if len(args) > 0 && migrateAll {
			return fmt.Errorf("cannot specify REQ-ID with --all flag")
		}

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		// Detect orphans
		excludePaths := []string{"/docs/", "/.claude/", "/.cursor/", "/.canary/specs/"}
		orphans, err := migrate.DetectOrphans(db, rootDir, excludePaths)
		if err != nil {
			return fmt.Errorf("failed to detect orphans: %w", err)
		}

		if len(orphans) == 0 {
			fmt.Println("✅ No orphaned requirements to migrate!")
			return nil
		}

		// Filter to specific requirement if provided
		var toMigrate []*migrate.OrphanedRequirement
		if len(args) == 1 {
			reqID := strings.ToUpper(args[0])
			found := false
			for _, orphan := range orphans {
				if orphan.ReqID == reqID {
					toMigrate = append(toMigrate, orphan)
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("requirement %s is not orphaned or does not exist", reqID)
			}
		} else {
			toMigrate = orphans
		}

		// Dry run mode
		if dryRun {
			fmt.Printf("🔍 Dry run: would migrate %d requirement(s)\n\n", len(toMigrate))
			for _, orphan := range toMigrate {
				dirName := orphan.ReqID + "-" + slugify(orphan.Features[0].Feature)
				fmt.Printf("Would create:\n")
				fmt.Printf("  📁 .canary/specs/%s/\n", dirName)
				fmt.Printf("  📄 .canary/specs/%s/spec.md (Confidence: %s)\n", dirName, orphan.Confidence)
				fmt.Printf("  📄 .canary/specs/%s/plan.md\n", dirName)
				fmt.Println()
			}
			fmt.Println("✅ Dry run complete (no files created)")
			return nil
		}

		// Perform migration
		specsDir := filepath.Join(rootDir, ".canary", "specs")
		if err := os.MkdirAll(specsDir, 0755); err != nil {
			return fmt.Errorf("failed to create specs directory: %w", err)
		}

		migratedCount := 0
		for _, orphan := range toMigrate {
			fmt.Printf("🔄 Migrating %s...\n", orphan.ReqID)

			// Generate spec
			specContent, err := migrate.GenerateSpec(orphan)
			if err != nil {
				fmt.Printf("⚠️  Failed to generate spec for %s: %v\n", orphan.ReqID, err)
				continue
			}

			// Generate plan
			planContent, err := migrate.GeneratePlan(orphan)
			if err != nil {
				fmt.Printf("⚠️  Failed to generate plan for %s: %v\n", orphan.ReqID, err)
				continue
			}

			// Create directory
			primaryFeature := orphan.Features[0].Feature
			dirName := orphan.ReqID + "-" + slugify(primaryFeature)
			specDir := filepath.Join(specsDir, dirName)

			if err := os.MkdirAll(specDir, 0755); err != nil {
				fmt.Printf("⚠️  Failed to create directory for %s: %v\n", orphan.ReqID, err)
				continue
			}

			// Write spec
			specPath := filepath.Join(specDir, "spec.md")
			if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
				fmt.Printf("⚠️  Failed to write spec for %s: %v\n", orphan.ReqID, err)
				continue
			}

			// Write plan
			planPath := filepath.Join(specDir, "plan.md")
			if err := os.WriteFile(planPath, []byte(planContent), 0644); err != nil {
				fmt.Printf("⚠️  Failed to write plan for %s: %v\n", orphan.ReqID, err)
				continue
			}

			fmt.Printf("✅ Migrated %s (Confidence: %s)\n", orphan.ReqID, orphan.Confidence)
			fmt.Printf("   📄 %s\n", specPath)
			fmt.Printf("   📄 %s\n", planPath)

			if orphan.Confidence == migrate.ConfidenceLow {
				fmt.Printf("   ⚠️  Low confidence - please review and update manually\n")
			}

			fmt.Println()
			migratedCount++
		}

		// Summary
		fmt.Printf("\n✅ Successfully migrated %d requirement(s)\n", migratedCount)

		if migratedCount > 0 {
			fmt.Println("\n💡 Next steps:")
			fmt.Println("   1. Review generated specifications for accuracy")
			fmt.Println("   2. Update spec.md files with detailed requirements")
			fmt.Println("   3. Update plan.md files with implementation details")
			fmt.Println("   4. Run 'canary scan' to reindex the database")
		}

		return nil
	},
}

// slugify converts a string to a slug (lowercase, alphanumeric + hyphens)
func slugify(s string) string {
	result := ""
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result += string(c)
		} else if c == ' ' || c == '-' || c == '_' {
			result += "-"
		}
	}
	// Convert to lowercase
	lower := ""
	for _, c := range result {
		if c >= 'A' && c <= 'Z' {
			lower += string(c + 32)
		} else {
			lower += string(c)
		}
	}
	// Limit length
	if len(lower) > 40 {
		return lower[:40]
	}
	return lower
}

func init() {
	// Add subcommands
	orphanCmd.AddCommand(migrateDetectCmd)
	orphanCmd.AddCommand(migrateRunCmd)

	// Global flags
	migrateDetectCmd.Flags().String("db", ".canary/canary.db", "Path to database file")
	migrateDetectCmd.Flags().String("root", ".", "Root directory for project")
	migrateDetectCmd.Flags().Bool("show-features", false, "Show feature details for each orphan")

	migrateRunCmd.Flags().String("db", ".canary/canary.db", "Path to database file")
	migrateRunCmd.Flags().String("root", ".", "Root directory for project")
	migrateRunCmd.Flags().Bool("all", false, "Migrate all orphaned requirements")
	migrateRunCmd.Flags().Bool("dry-run", false, "Preview migration without creating files")
}
