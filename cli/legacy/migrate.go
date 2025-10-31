package legacy

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/migrate"
)

// CANARY: REQ=CBIN-131; FEATURE="MigrateFromCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var MigrateFromCmd = &cobra.Command{
	Use:   "migrate-from <system-type> [directory]",
	Short: "Migrate from spec-kit or legacy canary to unified canary system",
	Long: `Migrate an existing spec-kit or legacy canary project to the new unified system.

System types:
  spec-kit       - Migrate from spec-kit specification system
  legacy-canary  - Migrate from legacy CANARY token system

The migration will:
- Create .canary/ directory structure
- Copy/merge existing templates and configurations
- Preserve existing CANARY tokens and documentation
- Create missing files (constitution, slash commands, etc.)

Use --dry-run to preview changes before applying them.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		systemTypeStr := args[0]
		rootDir := "."
		if len(args) > 1 {
			rootDir = args[1]
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		force, _ := cmd.Flags().GetBool("force")

		// Parse system type
		var systemType migrate.SystemType
		switch systemTypeStr {
		case "spec-kit":
			systemType = migrate.SystemTypeSpecKit
		case "legacy-canary":
			systemType = migrate.SystemTypeLegacyCanary
		default:
			return fmt.Errorf("unsupported system type: %s (use 'spec-kit' or 'legacy-canary')", systemTypeStr)
		}

		// Detect actual system type
		detectedType, details := migrate.DetectSystemType(rootDir)

		// Check if already migrated
		if detectedType == migrate.SystemTypeMigrated {
			fmt.Printf("✅ System already migrated!\n\n")
			fmt.Printf("Details: %s\n\n", details)
			fmt.Println("This system is already using the unified CANARY system.")
			fmt.Println("No migration needed.")
			fmt.Println("\nAvailable commands:")
			fmt.Println("  canary index         # Build/rebuild token database")
			fmt.Println("  canary list          # List tokens")
			fmt.Println("  canary scan          # Scan for CANARY tokens")
			fmt.Println("  canary implement     # Show implementation locations")
			return nil
		}

		if !force && detectedType != systemType {
			if detectedType == migrate.SystemTypeUnknown {
				fmt.Printf("⚠️  Warning: No %s system detected in %s\n", systemType, rootDir)
				fmt.Printf("Details: %s\n", details)
				fmt.Println("\nUse --force to proceed anyway, or run 'canary detect' to identify the system type.")
				return fmt.Errorf("system type mismatch")
			}
			fmt.Printf("⚠️  Warning: Detected %s but trying to migrate as %s\n", detectedType, systemType)
			fmt.Println("Use --force to override detection, or specify the correct system type.")
			return fmt.Errorf("system type mismatch")
		}

		// Create migration plan
		fmt.Printf("📋 Planning migration from %s...\n\n", systemType)
		plan, err := migrate.PlanMigration(rootDir, systemType, dryRun)
		if err != nil {
			return fmt.Errorf("failed to create migration plan: %w", err)
		}

		// Show summary
		fmt.Println(migrate.GetMigrationSummary(plan))

		// Execute migration
		fmt.Printf("\n🚀 Executing migration...\n\n")
		if err := migrate.ExecuteMigration(rootDir, plan, dryRun); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		if dryRun {
			fmt.Println("\n✅ Dry run complete - no changes were made")
			fmt.Println("Run without --dry-run to apply changes")
		} else {
			fmt.Println("\n✅ Migration complete!")
			fmt.Println("\nNext steps:")
			fmt.Println("  1. Review migrated files in .canary/")
			fmt.Println("  2. Update slash commands in .canary/templates/commands/ for your workflow")
			fmt.Println("  3. Run: canary index")
			fmt.Println("  4. Run: canary scan --root . --out status.json")
		}

		return nil
	},
}

// CANARY: REQ=CBIN-131; FEATURE="DetectCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var DetectCmd = &cobra.Command{
	Use:   "detect [directory]",
	Short: "Detect what type of system exists (spec-kit or legacy canary)",
	Long: `Analyze a directory to determine if it contains a spec-kit or legacy canary system.

This helps determine the best migration strategy.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rootDir := "."
		if len(args) > 0 {
			rootDir = args[0]
		}

		systemType, details := migrate.DetectSystemType(rootDir)

		fmt.Printf("🔍 Analyzing: %s\n\n", rootDir)
		fmt.Printf("System Type: %s\n", systemType)
		fmt.Printf("Details: %s\n\n", details)

		if systemType == migrate.SystemTypeUnknown {
			fmt.Println("No recognized system found. You can initialize a new CANARY project with:")
			fmt.Println("  canary init")
			return nil
		}

		if systemType == migrate.SystemTypeMigrated {
			fmt.Println("✅ This system is already using the unified CANARY system!")
			fmt.Println("\nAvailable commands:")
			fmt.Println("  canary index         # Build/rebuild token database")
			fmt.Println("  canary list          # List tokens")
			fmt.Println("  canary scan          # Scan for CANARY tokens")
			fmt.Println("  canary implement     # Show implementation locations")
			return nil
		}

		fmt.Println("To migrate this system, run:")
		fmt.Printf("  canary migrate-from %s\n", systemType)
		fmt.Println("\nFor a dry run (preview changes):")
		fmt.Printf("  canary migrate-from %s --dry-run\n", systemType)

		return nil
	},
}
