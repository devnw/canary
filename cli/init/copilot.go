package init

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"go.devnw.com/canary/cli/internal/utils"
)

// CANARY: REQ=CBIN-148; FEATURE="CopilotInstructionCreator"; ASPECT=CLI; STATUS=BENCHED; TEST=TestCreateCopilotInstructions; BENCH=BenchmarkCreateCopilotInstructions; UPDATED=2025-10-19
// createCopilotInstructions generates GitHub Copilot instruction files for the project
func createCopilotInstructions(projectName, projectKey string) error {
	instructionsDir := filepath.Join(projectName, ".github", "instructions")

	// Create .github/instructions/ directory structure
	if err := os.MkdirAll(instructionsDir, 0755); err != nil {
		return fmt.Errorf("create .github/instructions: %w", err)
	}

	// Define instruction files to create
	instructionFiles := map[string]string{
		// Repository-wide instruction
		"repository.md": "base/copilot/repository.md",

		// Path-specific instructions (nested directories)
		".canary/specs/instruction.md": "base/copilot/specs.md",
		".canary/instruction.md":       "base/copilot/canary.md",
		"tests/instruction.md":         "base/copilot/tests.md",
	}

	// Template data for variable substitution
	type TemplateData struct {
		ProjectKey string
	}
	data := TemplateData{ProjectKey: projectKey}

	for targetPath, templatePath := range instructionFiles {
		fullTargetPath := filepath.Join(instructionsDir, targetPath)

		// Check if file already exists (preserve user customizations)
		if _, err := os.Stat(fullTargetPath); err == nil {
			fmt.Printf("⏭️  Skipping existing instruction file: %s\n", targetPath)
			continue
		}

		// Create parent directories for path-specific instructions
		if err := os.MkdirAll(filepath.Dir(fullTargetPath), 0755); err != nil {
			return fmt.Errorf("create directory for %s: %w", targetPath, err)
		}

		// Read template from embedded filesystem
		templateContent, err := utils.ReadEmbeddedFile(templatePath)
		if err != nil {
			return fmt.Errorf("read template %s: %w", templatePath, err)
		}

		// Parse and execute template
		tmpl, err := template.New(targetPath).Parse(string(templateContent))
		if err != nil {
			return fmt.Errorf("parse template %s: %w", templatePath, err)
		}

		// Write to file
		outFile, err := os.Create(fullTargetPath)
		if err != nil {
			return fmt.Errorf("create file %s: %w", fullTargetPath, err)
		}

		if err := tmpl.Execute(outFile, data); err != nil {
			outFile.Close()
			return fmt.Errorf("execute template %s: %w", templatePath, err)
		}
		outFile.Close()

		fmt.Printf("✅ Created Copilot instruction: %s\n", targetPath)
	}

	fmt.Println("✅ GitHub Copilot instructions configured")
	return nil
}
