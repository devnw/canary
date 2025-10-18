// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-131; FEATURE="MigrateFrom"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
package migrate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.devnw.com/canary/embedded"
)

// SystemType represents the type of system being migrated from
type SystemType string

const (
	SystemTypeSpecKit      SystemType = "spec-kit"
	SystemTypeLegacyCanary SystemType = "legacy-canary"
	SystemTypeMigrated     SystemType = "migrated"
	SystemTypeUnknown      SystemType = "unknown"
)

// DetectSystemType identifies what type of system exists in a directory
func DetectSystemType(rootDir string) (SystemType, string) {
	// First check if already migrated (has .canary/ with modern structure)
	canaryDB := filepath.Join(rootDir, ".canary/canary.db")
	canaryTemplates := filepath.Join(rootDir, ".canary/templates")

	hasDB := false
	hasTemplates := false

	if _, err := os.Stat(canaryDB); err == nil {
		hasDB = true
	}
	if _, err := os.Stat(canaryTemplates); err == nil {
		hasTemplates = true
	}

	// If has database OR templates in .canary/, it's already migrated
	if hasDB || hasTemplates {
		details := "System already migrated to unified CANARY"
		if hasDB {
			details += " (has .canary/canary.db)"
		}
		if hasTemplates {
			details += " (has .canary/templates/)"
		}
		return SystemTypeMigrated, details
	}

	// Check for spec-kit indicators
	specKitIndicators := []string{
		"memory/constitution.md",
		"templates/spec-template.md",
		"templates/plan-template.md",
		"templates/commands/specify.md",
		"templates/commands/plan.md",
	}

	specKitScore := 0
	for _, indicator := range specKitIndicators {
		if _, err := os.Stat(filepath.Join(rootDir, indicator)); err == nil {
			specKitScore++
		}
	}

	if specKitScore >= 3 {
		return SystemTypeSpecKit, fmt.Sprintf("Detected spec-kit system (%d/5 indicators found)", specKitScore)
	}

	// Check for legacy canary indicators
	legacyIndicators := []string{
		"tools/canary",
		"tools/canary/main.go",
		"status.json",
		"GAP_ANALYSIS.md",
	}

	legacyScore := 0
	for _, indicator := range legacyIndicators {
		if _, err := os.Stat(filepath.Join(rootDir, indicator)); err == nil {
			legacyScore++
		}
	}

	if legacyScore >= 2 {
		return SystemTypeLegacyCanary, fmt.Sprintf("Detected legacy CANARY system (%d/4 indicators found)", legacyScore)
	}

	return SystemTypeUnknown, "No recognized system type detected"
}

// MigrationPlan describes what will be migrated
type MigrationPlan struct {
	SystemType    SystemType
	FilesToCopy   []FileCopy
	FilesToMerge  []FileMerge
	FilesToCreate []string
	Warnings      []string
}

// FileCopy represents a file to be copied
type FileCopy struct {
	Source string
	Dest   string
}

// FileMerge represents a file that needs merging
type FileMerge struct {
	Source      string
	Dest        string
	Description string
}

// PlanMigration creates a migration plan
func PlanMigration(rootDir string, systemType SystemType, dryRun bool) (*MigrationPlan, error) {
	plan := &MigrationPlan{
		SystemType:    systemType,
		FilesToCopy:   []FileCopy{},
		FilesToMerge:  []FileMerge{},
		FilesToCreate: []string{},
		Warnings:      []string{},
	}

	switch systemType {
	case SystemTypeSpecKit:
		return planSpecKitMigration(rootDir, plan)
	case SystemTypeLegacyCanary:
		return planLegacyCanaryMigration(rootDir, plan)
	case SystemTypeMigrated, SystemTypeUnknown:
		return nil, fmt.Errorf("cannot migrate from type: %s", systemType)
	default:
		return nil, fmt.Errorf("unsupported system type: %s", systemType)
	}
}

// planSpecKitMigration plans migration from spec-kit
func planSpecKitMigration(rootDir string, plan *MigrationPlan) (*MigrationPlan, error) {
	// Check if .canary/ already exists
	canaryDir := filepath.Join(rootDir, ".canary")
	if _, err := os.Stat(canaryDir); err == nil {
		plan.Warnings = append(plan.Warnings, ".canary/ directory already exists - will merge content")
	}

	// Files to copy directly
	copyFiles := []struct{ src, dest string }{
		{"memory/constitution.md", ".canary/memory/constitution.md"},
		{"templates/spec-template.md", ".canary/templates/spec-template.md"},
		{"templates/plan-template.md", ".canary/templates/plan-template.md"},
		{"templates/tasks-template.md", ".canary/templates/tasks-template.md"},
		{"templates/checklist-template.md", ".canary/templates/checklist-template.md"},
	}

	for _, f := range copyFiles {
		srcPath := filepath.Join(rootDir, f.src)
		if _, err := os.Stat(srcPath); err == nil {
			plan.FilesToCopy = append(plan.FilesToCopy, FileCopy{
				Source: f.src,
				Dest:   f.dest,
			})
		}
	}

	// Slash commands to copy
	commands := []string{"specify", "plan", "tasks", "implement", "clarify", "analyze", "checklist", "constitution"}
	for _, cmd := range commands {
		srcPath := filepath.Join(rootDir, "templates", "commands", cmd+".md")
		if _, err := os.Stat(srcPath); err == nil {
			plan.FilesToCopy = append(plan.FilesToCopy, FileCopy{
				Source: filepath.Join("templates", "commands", cmd+".md"),
				Dest:   filepath.Join(".canary", "templates", "commands", cmd+".md"),
			})
		}
	}

	// Scripts to adapt
	if _, err := os.Stat(filepath.Join(rootDir, "scripts")); err == nil {
		plan.Warnings = append(plan.Warnings, "scripts/ directory found - will need manual review for compatibility")
	}

	// Files to merge
	if _, err := os.Stat(filepath.Join(rootDir, "README.md")); err == nil {
		plan.FilesToMerge = append(plan.FilesToMerge, FileMerge{
			Source:      "README.md",
			Dest:        "README.md",
			Description: "Merge spec-kit README with CANARY token documentation",
		})
	}

	// Files to create
	// Note: For spec-kit migration, these files are typically created by user
	// after migration via slash commands, so we don't need to create them here.
	// The .canary/ templates were already copied from templates/ above.
	plan.FilesToCreate = []string{}

	return plan, nil
}

// planLegacyCanaryMigration plans migration from legacy canary
func planLegacyCanaryMigration(rootDir string, plan *MigrationPlan) (*MigrationPlan, error) {
	// Check if .canary/ already exists
	canaryDir := filepath.Join(rootDir, ".canary")
	if _, err := os.Stat(canaryDir); err == nil {
		plan.Warnings = append(plan.Warnings, ".canary/ directory already exists - will merge content")
	}

	// Preserve existing files
	preserveFiles := []string{"status.json", "status.csv", "GAP_ANALYSIS.md"}
	for _, f := range preserveFiles {
		srcPath := filepath.Join(rootDir, f)
		if _, err := os.Stat(srcPath); err == nil {
			plan.FilesToCopy = append(plan.FilesToCopy, FileCopy{
				Source: f,
				Dest:   f, // Keep in same location
			})
		}
	}

	// Files to create from embedded templates
	plan.FilesToCreate = []string{
		".canary/memory/constitution.md",
		".canary/templates/spec-template.md",
		".canary/templates/plan-template.md",
		".canary/templates/commands/constitution.md",
		".canary/templates/commands/plan.md",
		".canary/templates/commands/scan.md",
		".canary/templates/commands/specify.md",
		".canary/templates/commands/update-stale.md",
		".canary/templates/commands/verify.md",
	}

	// Warning about tools/canary
	if _, err := os.Stat(filepath.Join(rootDir, "tools/canary")); err == nil {
		plan.Warnings = append(plan.Warnings, "tools/canary/ scanner found - can be removed after migration (new binary has built-in scanner)")
	}

	return plan, nil
}

// ExecuteMigration performs the migration
func ExecuteMigration(rootDir string, plan *MigrationPlan, dryRun bool) error {
	if dryRun {
		fmt.Println("\n🔍 DRY RUN MODE - No changes will be made")
	}

	// Create directories
	dirsToCreate := []string{
		".canary",
		".canary/memory",
		".canary/templates",
		".canary/templates/commands",
		".canary/scripts",
		".canary/specs",
	}

	for _, dir := range dirsToCreate {
		dirPath := filepath.Join(rootDir, dir)
		if dryRun {
			fmt.Printf("Would create: %s\n", dir)
		} else {
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
			fmt.Printf("✅ Created: %s\n", dir)
		}
	}

	// Copy files
	for _, fc := range plan.FilesToCopy {
		srcPath := filepath.Join(rootDir, fc.Source)
		destPath := filepath.Join(rootDir, fc.Dest)

		if dryRun {
			fmt.Printf("Would copy: %s -> %s\n", fc.Source, fc.Dest)
		} else {
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create destination directory for %s: %w", fc.Dest, err)
			}
			if err := copyFile(srcPath, destPath); err != nil {
				return fmt.Errorf("failed to copy %s to %s: %w", fc.Source, fc.Dest, err)
			}
			fmt.Printf("✅ Copied: %s -> %s\n", fc.Source, fc.Dest)
		}
	}

	// Create files from templates
	for _, filename := range plan.FilesToCreate {
		destPath := filepath.Join(rootDir, filename)

		if dryRun {
			fmt.Printf("Would create: %s\n", filename)
		} else {
			// Map file path to embedded template path
			embeddedPath := filepath.Join("base", filename)

			content, err := embedded.CanaryFS.ReadFile(embeddedPath)
			if err != nil {
				return fmt.Errorf("failed to read embedded template %s: %w", embeddedPath, err)
			}

			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create directory for %s: %w", filename, err)
			}

			if err := os.WriteFile(destPath, content, 0644); err != nil {
				return fmt.Errorf("failed to create file %s: %w", filename, err)
			}

			fmt.Printf("✅ Created: %s\n", filename)
		}
	}

	// Show merge requirements
	if len(plan.FilesToMerge) > 0 {
		fmt.Println("\n⚠️  Files requiring manual merge:")
		for _, fm := range plan.FilesToMerge {
			fmt.Printf("   - %s: %s\n", fm.Source, fm.Description)
		}
	}

	// Show warnings
	if len(plan.Warnings) > 0 {
		fmt.Println("\n⚠️  Warnings:")
		for _, warning := range plan.Warnings {
			fmt.Printf("   - %s\n", warning)
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}

	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Preserve file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

// GetMigrationSummary returns a human-readable summary
func GetMigrationSummary(plan *MigrationPlan) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Migration Plan for %s:\n\n", plan.SystemType))
	sb.WriteString(fmt.Sprintf("Files to copy: %d\n", len(plan.FilesToCopy)))
	sb.WriteString(fmt.Sprintf("Files to merge: %d\n", len(plan.FilesToMerge)))
	sb.WriteString(fmt.Sprintf("Files to create: %d\n", len(plan.FilesToCreate)))
	sb.WriteString(fmt.Sprintf("Warnings: %d\n", len(plan.Warnings)))

	return sb.String()
}
