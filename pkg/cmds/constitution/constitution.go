package constitution

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"devnw.dev/canary/gate"
	"devnw.dev/canary/pkg/cmds/internal/utils"
)

// amendmentsKey is the gate-managed section that holds dated constitution
// amendments. It lives between <!-- CANARY:amendments:START --> and
// <!-- CANARY:amendments:END --> markers so repeated `canary constitution
// "<text>"` calls append rather than clobber.
const amendmentsKey = "amendments"

// CANARY: REQ=ENG-4303; FEATURE="ConstitutionCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2026-08-30
var ConstitutionCmd = &cobra.Command{
	Use:   "constitution [amendment text]",
	Short: "Create the constitution, or append a dated amendment",
	Long: `Create the project's constitutional principles in .canary/memory/constitution.md.

With no arguments and no existing file, writes the default constitution.
With no arguments and an existing file, reports that it already exists.
With arguments and an existing file, appends the argument text as a dated
amendment inside a managed <!-- CANARY:amendments:START/END --> section.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		constitutionPath := ".canary/memory/constitution.md"

		if _, err := os.Stat(constitutionPath); os.IsNotExist(err) {
			// Read template from embedded FS
			content, err := utils.ReadEmbeddedFile("base/memory/constitution.md")
			if err != nil {
				return fmt.Errorf("read constitution template: %w", err)
			}

			// Ensure directory exists
			if err := os.MkdirAll(filepath.Dir(constitutionPath), 0750); err != nil {
				return fmt.Errorf("create memory directory: %w", err)
			}

			// Write constitution
			if err := os.WriteFile(constitutionPath, content, 0640); err != nil { //nolint:gosec // generated doc, group-readable by design
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
			if len(args) == 0 {
				fmt.Printf("✅ Constitution already exists at: %s\n", constitutionPath)
				return nil
			}
			if err := appendAmendment(constitutionPath, strings.TrimSpace(strings.Join(args, " "))); err != nil {
				return err
			}
			fmt.Printf("✅ Appended amendment to: %s\n", constitutionPath)
		}

		return nil
	},
}

// appendAmendment adds a dated bullet carrying text to the gate-managed
// amendments section of the constitution, preserving any amendments already
// there. gate.UpdateSection replaces the whole section body, so the existing
// bullets are read out first and carried forward.
func appendAmendment(path, text string) error {
	if text == "" {
		return fmt.Errorf("amendment text is empty")
	}
	existing, err := os.ReadFile(path) //nolint:gosec // managed project doc
	if err != nil {
		return fmt.Errorf("read constitution: %w", err)
	}
	body := existingAmendments(string(existing))
	entry := fmt.Sprintf("- %s: %s", time.Now().UTC().Format("2006-01-02"), text)
	if body == "" {
		body = entry
	} else {
		body = body + "\n" + entry
	}
	if err := gate.UpdateSection(path, amendmentsKey, body); err != nil {
		return fmt.Errorf("update amendments section: %w", err)
	}
	return nil
}

// existingAmendments returns the current amendment lines (the body between the
// section markers), or "" when the section is absent or empty.
func existingAmendments(content string) string {
	start := gate.StartMarker(amendmentsKey)
	end := gate.EndMarker(amendmentsKey)
	si := strings.Index(content, start)
	if si < 0 {
		return ""
	}
	si += len(start)
	ei := strings.Index(content[si:], end)
	if ei < 0 {
		return ""
	}
	return strings.Trim(content[si:si+ei], "\n")
}
