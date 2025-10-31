package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/gap"
	"go.devnw.com/canary/internal/reqid"
	"go.devnw.com/canary/internal/storage"
	"go.devnw.com/canary/cli/internal/utils"
)

// CANARY: REQ=CBIN-121; FEATURE="PlanCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var PlanCmd = &cobra.Command{
	Use:   "plan <CBIN-XXX> [tech-stack]",
	Short: "Generate technical implementation plan for a requirement",
	Long: `Generate a technical implementation plan from a requirement specification.

Creates a plan.md file in the spec directory with implementation details,
tech stack decisions, and CANARY token placement instructions.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reqID := args[0]
		techStack := ""
		if len(args) > 1 {
			techStack = strings.Join(args[1:], " ")
		}

		// Get aspect flag
		aspect, _ := cmd.Flags().GetString("aspect")

		// Find spec directory
		specsDir := ".canary/specs"
		entries, err := os.ReadDir(specsDir)
		if err != nil {
			return fmt.Errorf("read specs directory: %w", err)
		}

		var specDir string
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), reqID) && entry.IsDir() {
				specDir = filepath.Join(specsDir, entry.Name())
				break
			}
		}

		if specDir == "" {
			return fmt.Errorf("specification not found for %s", reqID)
		}

		planFile := filepath.Join(specDir, "plan.md")
		if _, err := os.Stat(planFile); err == nil {
			return fmt.Errorf("plan already exists: %s", planFile)
		}

		// Read template
		templateContent, err := utils.ReadEmbeddedFile("base/templates/plan-template.md")
		if err != nil {
			return fmt.Errorf("read plan template: %w", err)
		}

		// Read spec to get feature name and aspect if not provided
		specFile := filepath.Join(specDir, "spec.md")
		specContent, err := os.ReadFile(specFile)
		if err != nil {
			return fmt.Errorf("read spec file: %w", err)
		}

		// Extract feature name and aspect from spec
		featureName := "Feature"
		specAspect := ""
		for _, line := range strings.Split(string(specContent), "\n") {
			if strings.HasPrefix(line, "# Feature Specification:") {
				featureName = strings.TrimPrefix(line, "# Feature Specification: ")
				featureName = strings.TrimSpace(featureName)
			}
			if strings.HasPrefix(line, "**Aspect:**") {
				// Extract aspect from markdown like "**Aspect:** API" or "**Aspect:** [API|CLI|...]"
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					aspectVal := strings.TrimSpace(parts[1])
					// Remove brackets and extract first option if it's a list
					aspectVal = strings.TrimPrefix(aspectVal, "[")
					aspectVal = strings.Split(aspectVal, "|")[0]
					aspectVal = strings.TrimSpace(aspectVal)
					if aspectVal != "" {
						specAspect = aspectVal
					}
				}
			}
		}

		// Use aspect from flag, or fall back to spec, or default to "Engine"
		if aspect == "" {
			if specAspect != "" {
				aspect = specAspect
			} else {
				aspect = "Engine"
			}
		}

		// Validate and normalize aspect
		if err := reqid.ValidateAspect(aspect); err != nil {
			return fmt.Errorf("invalid aspect: %w", err)
		}
		aspect = reqid.NormalizeAspect(aspect)

		content := string(templateContent)
		content = strings.ReplaceAll(content, "CBIN-XXX", reqID)
		content = strings.ReplaceAll(content, "[FEATURE NAME]", featureName)
		content = strings.ReplaceAll(content, "YYYY-MM-DD", time.Now().UTC().Format("2006-01-02"))
		content = strings.ReplaceAll(content, "SECURITY_REVIEW", aspect)

		if techStack != "" {
			content = strings.ReplaceAll(content, "[Go/Python/JavaScript/etc.]", techStack)
		}

		// CANARY: REQ=CBIN-140; FEATURE="PlanGapInjection"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
		// Inject gap analysis if available
		dbPath := ".canary/canary.db"
		if _, err := os.Stat(dbPath); err == nil {
			db, err := storage.Open(dbPath)
			if err == nil {
				defer db.Close()
				repo := storage.NewGapRepository(db)
				service := gap.NewService(repo)
				gapContent, err := service.FormatGapsForInjection(reqID)
				if err == nil && gapContent != "" {
					// Inject gaps at the end of the plan content
					content += "\n" + gapContent
				}
			}
		}

		if err := os.WriteFile(planFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("write plan file: %w", err)
		}

		fmt.Printf("✅ Created implementation plan: %s\n", planFile)
		fmt.Printf("\nRequirement: %s\n", reqID)
		fmt.Println("\nNext steps:")
		fmt.Printf("  1. Edit %s to complete the plan\n", planFile)
		fmt.Println("  2. Implement following TDD (test-first)")
		fmt.Println("  3. Add CANARY tokens to source code")
		fmt.Println("  4. Run: canary scan")

		return nil
	},
}
