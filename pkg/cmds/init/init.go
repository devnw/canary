package init

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/cmds/internal/utils"
)

//go:embed base
var embedded embed.FS

// InitCmd bootstraps a new project with CANARY token conventions
var InitCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new project with CANARY",
	Long: `Bootstrap a new project with CANARY spec-kit-inspired workflow.

Installation Modes:
  Global (default): Installs commands in ~/.claude/commands/, ~/.cursor/commands/, etc.
                    for use across all projects
  Local (--local):  Installs commands in .claude/commands/, .cursor/commands/, etc.
                    for project-specific use

Creates:
- .canary/ directory with templates, scripts, agents, and slash commands
- .canary/agents/ directory with pre-configured CANARY agent definitions
- README.md with CANARY token format specification
- GAP_ANALYSIS.md template for tracking requirements
- CLAUDE.md for AI agent integration (slash commands)

The agent files support template variables that can be customized:
  --agent-prefix: Agent name prefix (default: project key)
  --agent-model:  AI model to use (default: sonnet)
  --agent-color:  Agent color theme (default: blue)

Examples:
  canary init                  # Global install (default)
  canary init --local           # Local install in current project
  canary init myproject --local # Local install in new project`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := "."
		if len(args) > 0 {
			projectName = args[0]
		}

		// Check if this is an update to an existing project
		canaryDir := filepath.Join(projectName, ".canary")
		isUpdate := false
		if _, err := os.Stat(canaryDir); err == nil {
			isUpdate = true
			fmt.Println("📦 Existing CANARY project detected - updating...")
		}

		// Create project directory if needed
		if projectName != "." {
			if err := os.MkdirAll(projectName, 0750); err != nil {
				return fmt.Errorf("create project dir: %w", err)
			}
		}

		// Get project key and check for existing key BEFORE copying structure
		projectKey, _ := cmd.Flags().GetString("key")
		projectYamlPath := filepath.Join(projectName, ".canary", "project.yaml")

		// If updating existing project, try to read existing key from project.yaml BEFORE overwriting
		if isUpdate && projectKey == "" {
			if existingContent, err := os.ReadFile(projectYamlPath); err == nil {
				// Extract existing key from project.yaml (handles both "key:" and indented "  key:")
				for _, line := range strings.Split(string(existingContent), "\n") {
					trimmed := strings.TrimSpace(line)
					if strings.HasPrefix(trimmed, "key:") {
						parts := strings.SplitN(trimmed, ":", 2)
						if len(parts) == 2 {
							existingKey := strings.TrimSpace(parts[1])
							existingKey = strings.Trim(existingKey, "\"' ")
							if existingKey != "" && existingKey != "{{PROJECT_KEY}}" {
								projectKey = existingKey
								fmt.Printf("📦 Using existing project key: %s\n", projectKey)
								break
							}
						}
					}
				}
			} else {
				fmt.Printf("⚠️  Warning: Could not read project.yaml: %v\n", err)
			}
		}

		// Only prompt if still no key
		if projectKey == "" {
			// Prompt for project key
			fmt.Print("Enter project requirement ID prefix (e.g., CBIN, PROJ, ACME): ")
			var input string
			if _, err := fmt.Scanln(&input); err != nil {
				// Handle scan error (e.g., EOF or interrupted input)
				input = ""
			}
			projectKey = strings.TrimSpace(strings.ToUpper(input))
		}
		if projectKey == "" {
			projectKey = "PROJ" // Default
		}

		// Copy .canary/ structure from base/ (after extracting existing key)
		if err := copyCanaryStructure(projectName); err != nil {
			return fmt.Errorf("copy .canary structure: %w", err)
		}

		// Copy .canaryignore template
		canaryignoreContent, err := utils.ReadEmbeddedFile("base/.canaryignore")
		if err == nil {
			canaryignorePath := filepath.Join(projectName, ".canaryignore")
			if err := os.WriteFile(canaryignorePath, canaryignoreContent, 0640); err != nil {
				return fmt.Errorf("write .canaryignore: %w", err)
			}
		}

		// Customize project.yaml with the project key
		if err := customizeProjectYaml(projectYamlPath, projectName, projectKey); err != nil {
			return fmt.Errorf("customize project.yaml: %w", err)
		}

		// Get installation mode flag
		localInstall, _ := cmd.Flags().GetBool("local")

		// Get agent selection flags
		agentsList, _ := cmd.Flags().GetStringSlice("agents")
		allAgents, _ := cmd.Flags().GetBool("all-agents")

		// Get agent configuration flags
		agentPrefix, _ := cmd.Flags().GetString("agent-prefix")
		agentModel, _ := cmd.Flags().GetString("agent-model")
		agentColor, _ := cmd.Flags().GetString("agent-color")

		// Set defaults if not provided
		if agentPrefix == "" {
			agentPrefix = projectKey // Use project key as default agent prefix
		}
		if agentModel == "" {
			agentModel = "sonnet" // Default to Claude 4.5
		}
		if agentColor == "" {
			agentColor = "blue"
		}

		// Copy and process agent files to .canary/agents/ with template substitution
		if err := copyAndProcessAgentFiles(projectName, agentPrefix, agentModel, agentColor); err != nil {
			return fmt.Errorf("copy agent files: %w", err)
		}

		// Install/update slash commands to agent directories
		slashCommandNotes, err := installSlashCommands(projectName, agentsList, allAgents, localInstall)
		if err != nil {
			return fmt.Errorf("install slash commands: %w", err)
		}

		// Install agent files to each agent system's directory
		if err := installAgentFilesToSystems(projectName, agentsList, allAgents, agentPrefix, agentModel, agentColor, localInstall); err != nil {
			return fmt.Errorf("install agent files to systems: %w", err)
		}

		// CANARY: REQ=CBIN-148; FEATURE="CopilotInitInstructions"; ASPECT=CLI; STATUS=BENCHED; TEST=TestCreateCopilotInstructions; BENCH=BenchmarkCreateCopilotInstructions; UPDATED=2026-08-29
		// Create GitHub Copilot instruction files
		if err := createCopilotInstructions(projectName, projectKey); err != nil {
			return fmt.Errorf("create Copilot instructions: %w", err)
		}

		// Rebuild canary binary if we're updating
		if isUpdate {
			fmt.Println("🔧 Rebuilding canary binary...")
			buildCmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", "./bin/canary", "./cmd/canary")
			buildCmd.Stdout = os.Stdout
			buildCmd.Stderr = os.Stderr
			if err := buildCmd.Run(); err != nil {
				fmt.Printf("⚠️  Warning: Failed to rebuild canary binary: %v\n", err)
				fmt.Println("   Run 'make canary-build' or 'go build -o ./bin/canary ./cmd/canary/main.go' to rebuild manually")
			} else {
				fmt.Println("✅ Canary binary updated")
			}
		}

		// Create README.md
		readme := "# CANARY Token Specification\n\n" +
			"## Format\n\n" +
			"CANARY tokens track requirements directly in source code:\n\n" +
			"```\n" +
			"// CANARY: REQ=CBIN-###; FEATURE=\"Name\"; ASPECT=API; STATUS=IMPL; [TEST=TestName]; [BENCH=BenchName]; [OWNER=team]; UPDATED=<YYYY-MM-DD>\n" +
			"```\n\n" +
			"## Required Fields\n\n" +
			"- **REQ**: Requirement ID (format: CBIN-###)\n" +
			"- **FEATURE**: Short feature name\n" +
			"- **ASPECT**: Category (API, CLI, Engine, Storage, etc.)\n" +
			"- **STATUS**: Implementation state\n" +
			"- **UPDATED**: Last update date (YYYY-MM-DD)\n\n" +
			"## Status Values\n\n" +
			"- **MISSING**: Planned but not implemented\n" +
			"- **STUB**: Placeholder implementation\n" +
			"- **IMPL**: Implemented\n" +
			"- **TESTED**: Implemented with tests (auto-promoted from IMPL+TEST)\n" +
			"- **BENCHED**: Tested with benchmarks (auto-promoted from TESTED+BENCH)\n" +
			"- **REMOVED**: Deprecated/removed\n\n" +
			"## Optional Fields\n\n" +
			"- **TEST**: Test function name (promotes IMPL → TESTED)\n" +
			"- **BENCH**: Benchmark function name (promotes TESTED → BENCHED)\n" +
			"- **OWNER**: Team/person responsible\n\n" +
			"## Example\n\n" +
			"```go\n" +
			"// CANARY: REQ=CBIN-001; FEATURE=\"UserAuth\"; ASPECT=API; STATUS=TESTED; TEST=TestUserAuth; OWNER=backend; UPDATED=2025-10-16\n" +
			"func AuthenticateUser(credentials *Credentials) (*Session, error) {\n" +
			"    // implementation\n" +
			"}\n" +
			"```\n\n" +
			"## Usage\n\n" +
			"```bash\n" +
			"# Scan for tokens and generate reports\n" +
			"canary scan --root . --out status.json --csv status.csv\n\n" +
			"# Verify GAP_ANALYSIS.md claims\n" +
			"canary scan --root . --verify GAP_ANALYSIS.md\n\n" +
			"# Check for stale tokens (30-day threshold)\n" +
			"canary scan --root . --strict\n\n" +
			"# Auto-update stale TESTED/BENCHED tokens\n" +
			"canary scan --root . --update-stale\n" +
			"```\n"
		readmePath := filepath.Join(projectName, "README_CANARY.md")
		if err := os.WriteFile(readmePath, []byte(readme), 0640); err != nil {
			return fmt.Errorf("write README: %w", err)
		}

		// Create GAP_ANALYSIS.md template
		gap := "# Requirements Gap Analysis\n\n" +
			"## Claimed Requirements\n\n" +
			"List requirements that are fully implemented and verified:\n\n" +
			"✅ CBIN-001 - UserAuth API fully tested\n" +
			"✅ CBIN-002 - DataValidation with benchmarks\n\n" +
			"## Gaps\n\n" +
			"List requirements that are planned or in progress:\n\n" +
			"- [ ] CBIN-003 - ReportGeneration (STATUS=IMPL, needs tests)\n" +
			"- [ ] CBIN-004 - CacheOptimization (STATUS=STUB)\n\n" +
			"## Verification\n\n" +
			"Run verification with:\n\n" +
			"```bash\n" +
			"canary scan --root . --verify GAP_ANALYSIS.md\n" +
			"```\n\n" +
			"This will:\n" +
			"- ✅ Verify claimed requirements are TESTED or BENCHED\n" +
			"- ❌ Fail with exit code 2 if claims are overclaimed\n"
		gapPath := filepath.Join(projectName, "GAP_ANALYSIS.md")
		if err := os.WriteFile(gapPath, []byte(gap), 0640); err != nil {
			return fmt.Errorf("write GAP_ANALYSIS.md: %w", err)
		}

		// Update agent context files for AI agent integration (preserves existing content)
		if err := updateAgentContextFiles(projectName); err != nil {
			return fmt.Errorf("update agent context files: %w", err)
		}

		if isUpdate {
			fmt.Printf("\n✅ Updated CANARY project in: %s\n\n", projectName)
			fmt.Println("Updated:")
		} else {
			fmt.Printf("\n✅ Initialized CANARY project in: %s\n\n", projectName)
			fmt.Println("Created:")
		}
		fmt.Println("  ✅ .canary/ - Full workflow structure")
		fmt.Println("     ├── agents/ - Pre-configured CANARY agent definitions")
		fmt.Println("     ├── memory/constitution.md - Project principles")
		fmt.Println("     ├── scripts/ - Automation scripts")
		fmt.Println("     ├── templates/ - Spec/plan templates")
		fmt.Println("     └── templates/commands/ - Slash commands for AI agents")
		// Show installation location information
		if localInstall {
			fmt.Println("  ✅ Agent Files - Installed LOCALLY in project directory")
		} else {
			homeDir, _ := os.UserHomeDir()
			fmt.Printf("  ✅ Agent Files - Installed GLOBALLY in %s\n", homeDir)
		}

		// Show which agents had commands installed
		agentDirs := map[string]string{
			".claude":    "Claude Code",
			".cursor":    "Cursor",
			".github":    "GitHub Copilot",
			".windsurf":  "Windsurf",
			".kilocode":  "Kilocode",
			".roo":       "Roo",
			".opencode":  "opencode",
			".codex":     "Codex",
			".augment":   "Auggie",
			".codebuddy": "CodeBuddy",
			".amazonq":   "Amazon Q Developer",
		}

		// Determine where to check for agent directories
		checkDir := projectName
		if !localInstall {
			// For global install, check in home directory
			if homeDir, err := os.UserHomeDir(); err == nil {
				checkDir = homeDir
			}
		}

		installedAgents := []string{}
		for dir, name := range agentDirs {
			if _, err := os.Stat(filepath.Join(checkDir, dir)); err == nil {
				installedAgents = append(installedAgents, name)
			}
		}

		if len(installedAgents) > 0 {
			fmt.Printf("  ✅ AI Agent Integration (%d systems configured):\n", len(installedAgents))
			for _, agent := range installedAgents {
				fmt.Printf("     • %s\n", agent)
			}
			for _, note := range slashCommandNotes {
				fmt.Printf("     • %s\n", note)
			}
		}

		if !isUpdate {
			fmt.Println("  ✅ README_CANARY.md - Token format specification")
			fmt.Println("  ✅ GAP_ANALYSIS.md - Requirements tracking template")
			fmt.Println("  ✅ AGENTS.md - Codex / repository instructions")
			fmt.Println("  ✅ CLAUDE.md - Claude Code / Claude plugins")
			fmt.Println("  ✅ CURSOR.md - Cursor IDE / Cursor plugins")
			fmt.Println("  ✅ .cursor/rules/canary-requirements.mdc - Cursor rule (apply when editing requirements)")
			fmt.Println("  ✅ .cursor/mcp.json - Optional MCP (run `canary mcp` then use Cursor MCP tools)")
		}

		fmt.Print(`
Available Slash Commands for AI Agents:
	/canary.constitution - Create/update project principles
	/canary.specify - Create requirement specification
	/canary.plan - Generate implementation plan
	/canary.scan - Scan for CANARY tokens
	/canary.verify - Verify GAP_ANALYSIS.md claims
	/canary.update-stale - Update stale tokens
Next Steps:
	1. Open in AI agent (Claude Code, Cursor, etc.)
	2. Run: /canary.constitution to establish principles
	3. Run: /canary.specify "your feature description"
	4. Follow the spec-driven workflow!
`)

		return nil
	},
}
