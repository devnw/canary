package constitution

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/cli/internal/utils"
)

// CANARY: REQ=CBIN-119; FEATURE="ConstitutionCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var ConstitutionCmd = &cobra.Command{
	Use:   "constitution [description]",
	Short: "Create or update project governing principles",
	Long: `Create or update the project's constitutional principles in .canary/memory/constitution.md.

If no arguments are provided, creates the default constitution.
If arguments are provided, updates or adds specific principles.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		constitutionPath := ".canary/memory/constitution.md"

		if _, err := os.Stat(constitutionPath); os.IsNotExist(err) {
			// Read template from embedded FS
			content, err := utils.ReadEmbeddedFile("base/memory/constitution.md")
			if err != nil {
				return fmt.Errorf("read constitution template: %w", err)
			}

			// Ensure directory exists
			if err := os.MkdirAll(filepath.Dir(constitutionPath), 0755); err != nil {
				return fmt.Errorf("create memory directory: %w", err)
			}

			// Write constitution
			if err := os.WriteFile(constitutionPath, content, 0644); err != nil {
				return fmt.Errorf("write constitution: %w", err)
			}

			fmt.Printf("✅ Created constitution at: %s\n", constitutionPath)
			fmt.Println("\nConstitutional Principles:")
			fmt.Println("  I. Requirement-First Development")
			fmt.Println("  II. Specification Discipline")
			fmt.Println("  III. Token-Driven Planning")
			fmt.Println("  IV. Test-First Imperative")
			fmt.Println("  V. Simplicity and Anti-Abstraction")
			fmt.Println("  VI. Integration-First Testing")
			fmt.Println("  VII. Documentation Currency")
			fmt.Println("  VIII. Continuous Improvement")
			fmt.Println("  IX. Amendment Process")
		} else {
			fmt.Printf("✅ Constitution already exists at: %s\n", constitutionPath)
			if len(args) > 0 {
				fmt.Println("\nTo update specific principles, edit the file directly.")
			}
		}

		return nil
	},
}
