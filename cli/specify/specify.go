package specify

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/cli/internal/utils"
	"go.devnw.com/canary/internal/reqid"
)

// CANARY: REQ=CBIN-120; FEATURE="SpecifyCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var SpecifyCmd = &cobra.Command{
	Use:   "specify <feature-description>",
	Short: "Create a new requirement specification",
	Long: `Create a new CANARY requirement specification from a feature description.

Generates a new requirement ID with aspect-based format (CBIN-SECURITY_REVIEW-XXX),
creates a spec directory, and populates it with a specification template.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement --prompt flag to load custom prompts
		prompt, _ := cmd.Flags().GetString("prompt")
		_ = prompt // Stubbed for future use

		featureDesc := strings.Join(args, " ")
		aspect, _ := cmd.Flags().GetString("aspect")

		// Validate aspect
		if err := reqid.ValidateAspect(aspect); err != nil {
			return fmt.Errorf("invalid aspect: %w", err)
		}

		// Normalize aspect to canonical form
		aspect = reqid.NormalizeAspect(aspect)

		// Generate next requirement ID for this aspect
		generatedID, err := reqid.GenerateNextID("CBIN", aspect)
		if err != nil {
			return fmt.Errorf("generate requirement ID: %w", err)
		}

		// Create sanitized feature name for directory
		featureName := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				return r
			}
			return '-'
		}, featureDesc)
		if len(featureName) > 50 {
			featureName = featureName[:50]
		}
		featureName = strings.Trim(featureName, "-")

		specsDir := ".canary/specs"
		specDir := filepath.Join(specsDir, fmt.Sprintf("%s-%s", generatedID, featureName))
		specFile := filepath.Join(specDir, "spec.md")

		// Create directory
		if err := os.MkdirAll(specDir, 0755); err != nil {
			return fmt.Errorf("create spec directory: %w", err)
		}

		// Read and populate template
		templateContent, err := utils.ReadEmbeddedFile("base/templates/spec-template.md")
		if err != nil {
			return fmt.Errorf("read spec template: %w", err)
		}

		content := string(templateContent)
		content = strings.ReplaceAll(content, "CBIN-XXX", generatedID)
		content = strings.ReplaceAll(content, "[FEATURE NAME]", featureDesc)
		content = strings.ReplaceAll(content, "YYYY-MM-DD", time.Now().UTC().Format("2006-01-02"))
		content = strings.ReplaceAll(content, "SECURITY_REVIEW", aspect)

		if err := os.WriteFile(specFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("write spec file: %w", err)
		}

		fmt.Printf("✅ Created specification: %s\n", specFile)
		fmt.Printf("\nRequirement ID: %s\n", generatedID)
		fmt.Printf("Aspect: %s\n", aspect)
		fmt.Printf("Feature: %s\n", featureDesc)
		fmt.Println("\nNext steps:")
		fmt.Printf("  1. Edit %s to complete the specification\n", specFile)
		fmt.Printf("  2. Run: canary plan %s\n", generatedID)

		return nil
	},
}

func init() {
	SpecifyCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	SpecifyCmd.Flags().String("aspect", "Engine", "requirement aspect (API, CLI, Engine, Storage, Security, Docs, Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist)")
}
