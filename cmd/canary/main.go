// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-CLI-104; FEATURE="CanaryCLI"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.spyder.org/canary/embedded"
	"go.spyder.org/canary/internal/config"
	"go.spyder.org/canary/internal/gap"
	"go.spyder.org/canary/internal/migrate"
	"go.spyder.org/canary/internal/reqid"
	"go.spyder.org/canary/internal/storage"
)

var (
	version = "dev"
	rootCmd = &cobra.Command{
		Use:   "canary",
		Short: "Track requirements via CANARY tokens in source code",
		Long: `Canary tracks requirements through CANARY tokens embedded in source code.

Inspired by spec-kit's specification-driven development, canary provides
commands for scanning, creating, and managing requirement tokens.`,
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip auto-migration for commands that don't use the database
			skipCommands := map[string]bool{
				"init":         true,
				"create":       true,
				"constitution": true,
				"specify":      true,
				"plan":         true,
				"implement":    true,
				"scan":         true,
				"help":         true,
				"completion":   true,
				"migrate":      true, // migrate command manages migrations itself
				"rollback":     true, // rollback command manages migrations itself
				"detect":       true, // detect command just reads, doesn't need DB
				"migrate-from": true, // migrate-from creates .canary/, shouldn't auto-migrate first
			}

			if skipCommands[cmd.Name()] {
				return nil
			}

			// Check if database commands have --db flag
			dbPath := ".canary/canary.db" // default
			if cmd.Flags().Lookup("db") != nil {
				dbPath, _ = cmd.Flags().GetString("db")
			}

			// Auto-migrate if needed
			if err := storage.AutoMigrate(dbPath); err != nil {
				return fmt.Errorf("auto-migration failed: %w", err)
			}

			return nil
		},
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// CANARY: REQ=CBIN-111; FEATURE="ScanCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17
// scanCmd wraps the existing tools/canary scanner
var scanCmd = &cobra.Command{
	Use:   "scan [flags]",
	Short: "Scan for CANARY tokens and generate reports",
	Long: `Scan source code for CANARY tokens and generate status reports.

This command scans your codebase for CANARY tokens and generates JSON/CSV reports.

Flags:
  --root <dir>            Root directory to scan (default ".")
  --out <file>            Output status.json path (default "status.json")
  --csv <file>            Optional status.csv path
  --verify <file>         GAP_ANALYSIS file to verify claims
  --strict                Enforce staleness on TESTED/BENCHED tokens (30 days)
  --update-stale          Rewrite UPDATED field for stale tokens
  --skip <regex>          Skip path regex (RE2)
  --project-only          Filter by project requirement ID pattern

Examples:
  # Basic scan
  canary scan --root . --out status.json

  # Verify GAP_ANALYSIS.md claims
  canary scan --verify GAP_ANALYSIS.md

  # Update stale tokens
  canary scan --update-stale

  # Strict mode with staleness enforcement
  canary scan --strict`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Build path to the canary scanner
		scanner := filepath.Join("tools", "canary", "main.go")

		// Get all flags
		rootDir, _ := cmd.Flags().GetString("root")
		out, _ := cmd.Flags().GetString("out")
		csv, _ := cmd.Flags().GetString("csv")
		verify, _ := cmd.Flags().GetString("verify")
		strict, _ := cmd.Flags().GetBool("strict")
		updateStale, _ := cmd.Flags().GetBool("update-stale")
		skip, _ := cmd.Flags().GetString("skip")
		projectOnly, _ := cmd.Flags().GetBool("project-only")

		// Build scanner arguments
		scanArgs := []string{"run", scanner}

		if rootDir != "" {
			scanArgs = append(scanArgs, "-root", rootDir)
		}
		if out != "" {
			scanArgs = append(scanArgs, "-out", out)
		}
		if csv != "" {
			scanArgs = append(scanArgs, "-csv", csv)
		}
		if verify != "" {
			scanArgs = append(scanArgs, "-verify", verify)
		}
		if strict {
			scanArgs = append(scanArgs, "-strict")
		}
		if updateStale {
			scanArgs = append(scanArgs, "-update-stale")
		}
		if skip != "" {
			scanArgs = append(scanArgs, "-skip", skip)
		}
		if projectOnly {
			scanArgs = append(scanArgs, "-project-only")
		}

		// Pass through any additional args
		scanArgs = append(scanArgs, args...)

		goCmd := exec.Command("go", scanArgs...)
		goCmd.Stdout = os.Stdout
		goCmd.Stderr = os.Stderr
		goCmd.Stdin = os.Stdin

		return goCmd.Run()
	},
}

// initCmd bootstraps a new project with CANARY token conventions
var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new project with full CANARY workflow",
	Long: `Bootstrap a new project with CANARY spec-kit-inspired workflow.

Creates:
- .canary/ directory with templates, scripts, agents, and slash commands
- .canary/agents/ directory with pre-configured CANARY agent definitions
- README.md with CANARY token format specification
- GAP_ANALYSIS.md template for tracking requirements
- CLAUDE.md for AI agent integration (slash commands)

The agent files support template variables that can be customized:
  --agent-prefix: Agent name prefix (default: project key)
  --agent-model:  AI model to use (default: sonnet)
  --agent-color:  Agent color theme (default: blue)`,
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
			if err := os.MkdirAll(projectName, 0755); err != nil {
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
		canaryignoreContent, err := embedded.CanaryFS.ReadFile("base/.canaryignore")
		if err == nil {
			canaryignorePath := filepath.Join(projectName, ".canaryignore")
			if err := os.WriteFile(canaryignorePath, canaryignoreContent, 0644); err != nil {
				return fmt.Errorf("write .canaryignore: %w", err)
			}
		}

		// Customize project.yaml with the project key
		if err := customizeProjectYaml(projectYamlPath, projectName, projectKey); err != nil {
			return fmt.Errorf("customize project.yaml: %w", err)
		}

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
			agentModel = "claude-3-5-sonnet-20241022"
		}
		if agentColor == "" {
			agentColor = "blue"
		}

		// Copy and process agent files to .canary/agents/ with template substitution
		if err := copyAndProcessAgentFiles(projectName, agentPrefix, agentModel, agentColor); err != nil {
			return fmt.Errorf("copy agent files: %w", err)
		}

		// Install/update slash commands to agent directories
		if err := installSlashCommands(projectName, agentsList, allAgents); err != nil {
			return fmt.Errorf("install slash commands: %w", err)
		}

		// Install agent files to each agent system's directory
		if err := installAgentFilesToSystems(projectName, agentsList, allAgents, agentPrefix, agentModel, agentColor); err != nil {
			return fmt.Errorf("install agent files to systems: %w", err)
		}

		// Rebuild canary binary if we're updating
		if isUpdate {
			fmt.Println("\n🔧 Rebuilding canary binary...")
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
			"// CANARY: REQ=CBIN-###; FEATURE=\"Name\"; ASPECT=API; STATUS=IMPL; [TEST=TestName]; [BENCH=BenchName]; [OWNER=team]; UPDATED=YYYY-MM-DD\n" +
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
		if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
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
		if err := os.WriteFile(gapPath, []byte(gap), 0644); err != nil {
			return fmt.Errorf("write GAP_ANALYSIS.md: %w", err)
		}

		// Create CLAUDE.md for AI agent integration
		claudeMD := createClaudeMD()
		claudePath := filepath.Join(projectName, "CLAUDE.md")
		if err := os.WriteFile(claudePath, []byte(claudeMD), 0644); err != nil {
			return fmt.Errorf("write CLAUDE.md: %w", err)
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
		fmt.Println("  ✅ Agent Files - Installed to detected AI agent systems")

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

		installedAgents := []string{}
		for dir, name := range agentDirs {
			if _, err := os.Stat(filepath.Join(projectName, dir)); err == nil {
				installedAgents = append(installedAgents, name)
			}
		}

		if len(installedAgents) > 0 {
			fmt.Printf("  ✅ AI Agent Integration (%d systems detected):\n", len(installedAgents))
			for _, agent := range installedAgents {
				fmt.Printf("     • %s (commands + agent files)\n", agent)
			}
		}

		if !isUpdate {
			fmt.Println("  ✅ README_CANARY.md - Token format specification")
			fmt.Println("  ✅ GAP_ANALYSIS.md - Requirements tracking template")
			fmt.Println("  ✅ CLAUDE.md - AI agent slash command integration")
		}
		fmt.Println("\nAvailable Slash Commands for AI Agents:")
		fmt.Println("  /canary.constitution - Create/update project principles")
		fmt.Println("  /canary.specify - Create requirement specification")
		fmt.Println("  /canary.plan - Generate implementation plan")
		fmt.Println("  /canary.scan - Scan for CANARY tokens")
		fmt.Println("  /canary.verify - Verify GAP_ANALYSIS.md claims")
		fmt.Println("  /canary.update-stale - Update stale tokens")
		fmt.Println("\nNext Steps:")
		fmt.Println("  1. Open in AI agent (Claude Code, Cursor, etc.)")
		fmt.Println("  2. Run: /canary.constitution to establish principles")
		fmt.Println("  3. Run: /canary.specify \"your feature description\"")
		fmt.Println("  4. Follow the spec-driven workflow!")

		return nil
	},
}

// filterCanaryTokens removes CANARY tokens with OWNER=canary from file content
// This strips out CANARY CLI internal tracking tokens when copying templates to user projects
func filterCanaryTokens(content []byte) []byte {
	lines := strings.Split(string(content), "\n")
	filtered := make([]string, 0, len(lines))

	for _, line := range lines {
		// Check if line contains a CANARY token with OWNER=canary
		if strings.Contains(line, "CANARY:") && strings.Contains(line, "OWNER=canary") {
			// Skip this line - it's a CANARY CLI internal token
			continue
		}
		filtered = append(filtered, line)
	}

	return []byte(strings.Join(filtered, "\n"))
}

// CANARY: REQ=CBIN-105; FEATURE="InitWorkflow"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
// copyCanaryStructure copies the embedded .canary/ directory structure to the target project
func copyCanaryStructure(targetDir string) error {
	targetCanary := filepath.Join(targetDir, ".canary")

	// Create root .canary directory
	if err := os.MkdirAll(targetCanary, 0755); err != nil {
		return err
	}

	// Walk embedded filesystem
	return fs.WalkDir(embedded.CanaryFS, "base/.canary", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from base/.canary
		relPath := strings.TrimPrefix(path, "base/.canary")
		relPath = strings.TrimPrefix(relPath, "/")
		if relPath == "" {
			return nil // Skip root
		}

		targetPath := filepath.Join(targetCanary, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		// Read embedded file
		content, err := embedded.CanaryFS.ReadFile(path)
		if err != nil {
			return err
		}

		// Filter out CANARY CLI internal tokens (OWNER=canary) for markdown, Go, and shell script files
		if strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".sh") {
			content = filterCanaryTokens(content)
		}

		// Write to target with appropriate permissions
		mode := fs.FileMode(0644)
		if strings.HasSuffix(path, ".sh") {
			mode = 0755
		}
		return os.WriteFile(targetPath, content, mode)
	})
}

// CANARY: REQ=CBIN-105; FEATURE="InitWorkflow"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
// customizeProjectYaml replaces placeholders in project.yaml with actual values
func customizeProjectYaml(path, projectName, projectKey string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read project.yaml: %w", err)
	}

	// Replace placeholders
	text := string(content)
	text = strings.ReplaceAll(text, "{{PROJECT_NAME}}", projectName)
	text = strings.ReplaceAll(text, "{{PROJECT_DESCRIPTION}}", fmt.Sprintf("%s project with CANARY requirement tracking", projectName))
	text = strings.ReplaceAll(text, "{{PROJECT_KEY}}", projectKey)

	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		return fmt.Errorf("write project.yaml: %w", err)
	}

	return nil
}

// CANARY: REQ=CBIN-105; FEATURE="InitWorkflow"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
// AgentConfig defines configuration for each supported AI agent
type AgentConfig struct {
	Dir    string // Directory for agent files
	Prefix string // Prefix for command files (e.g., "canary.")
}

// installSlashCommands copies slash commands to agent systems based on selection mode
// agentsList: specific agents to install for (e.g., ["claude", "cursor"])
// allAgentsFlag: if true, install for all supported agents
// If both are empty/false, auto-detect existing agent directories
func installSlashCommands(targetDir string, agentsList []string, allAgentsFlag bool) error {
	sourceDir := filepath.Join(targetDir, ".canary", "templates", "commands")

	// Agent configurations - matching spec-kit-repo structure
	allAgents := map[string]AgentConfig{
		"claude":    {Dir: filepath.Join(targetDir, ".claude", "commands"), Prefix: "canary."},
		"cursor":    {Dir: filepath.Join(targetDir, ".cursor", "commands"), Prefix: "canary."},
		"copilot":   {Dir: filepath.Join(targetDir, ".github", "prompts"), Prefix: "canary-"},
		"windsurf":  {Dir: filepath.Join(targetDir, ".windsurf", "workflows"), Prefix: "canary-"},
		"kilocode":  {Dir: filepath.Join(targetDir, ".kilocode", "rules"), Prefix: "canary-"},
		"roo":       {Dir: filepath.Join(targetDir, ".roo", "rules"), Prefix: "canary-"},
		"opencode":  {Dir: filepath.Join(targetDir, ".opencode", "command"), Prefix: "canary-"},
		"codex":     {Dir: filepath.Join(targetDir, ".codex", "commands"), Prefix: "canary."},
		"auggie":    {Dir: filepath.Join(targetDir, ".augment", "rules"), Prefix: "canary-"},
		"codebuddy": {Dir: filepath.Join(targetDir, ".codebuddy", "commands"), Prefix: "canary."},
		"amazonq":   {Dir: filepath.Join(targetDir, ".amazonq", "prompts"), Prefix: "canary-"},
	}

	agentRootDirs := map[string]string{
		"claude":    filepath.Join(targetDir, ".claude"),
		"cursor":    filepath.Join(targetDir, ".cursor"),
		"copilot":   filepath.Join(targetDir, ".github"),
		"windsurf":  filepath.Join(targetDir, ".windsurf"),
		"kilocode":  filepath.Join(targetDir, ".kilocode"),
		"roo":       filepath.Join(targetDir, ".roo"),
		"opencode":  filepath.Join(targetDir, ".opencode"),
		"codex":     filepath.Join(targetDir, ".codex"),
		"auggie":    filepath.Join(targetDir, ".augment"),
		"codebuddy": filepath.Join(targetDir, ".codebuddy"),
		"amazonq":   filepath.Join(targetDir, ".amazonq"),
	}

	// Determine which agents to install for
	var selectedAgents map[string]AgentConfig

	if allAgentsFlag {
		// Install for all agents
		selectedAgents = allAgents
	} else if len(agentsList) > 0 {
		// Install for specific agents
		selectedAgents = make(map[string]AgentConfig)
		for _, agentName := range agentsList {
			if config, ok := allAgents[agentName]; ok {
				selectedAgents[agentName] = config
			} else {
				return fmt.Errorf("unknown agent: %s (valid: claude, cursor, copilot, windsurf, kilocode, roo, opencode, codex, auggie, codebuddy, amazonq)", agentName)
			}
		}
	} else {
		// Auto-detect existing agent directories
		selectedAgents = make(map[string]AgentConfig)
		for agentName, rootDir := range agentRootDirs {
			if _, err := os.Stat(rootDir); err == nil {
				selectedAgents[agentName] = allAgents[agentName]
			}
		}
	}

	// If no agents selected, nothing to install
	if len(selectedAgents) == 0 {
		fmt.Println("⚠️  No AI agent directories detected - skipping slash command installation")
		fmt.Println("   Create an agent directory (e.g., .claude/, .cursor/) or use --agents or --all-agents flag")
		return nil
	}

	// Read all command files from source
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("read commands directory: %w", err)
	}

	// Install commands for selected agents
	for agentName, config := range selectedAgents {
		// Create agent directory
		if err := os.MkdirAll(config.Dir, 0755); err != nil {
			return fmt.Errorf("create %s directory: %w", agentName, err)
		}

		// Copy each command file with appropriate prefix
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			sourcePath := filepath.Join(sourceDir, entry.Name())

			// Remove .md extension, add prefix, then add .md back
			baseName := strings.TrimSuffix(entry.Name(), ".md")
			targetName := config.Prefix + baseName + ".md"
			targetPath := filepath.Join(config.Dir, targetName)

			// Read source file
			content, err := os.ReadFile(sourcePath)
			if err != nil {
				return fmt.Errorf("read command file %s: %w", entry.Name(), err)
			}

			// Write to target with prefix
			if err := os.WriteFile(targetPath, content, 0644); err != nil {
				return fmt.Errorf("write command file %s for %s: %w", targetName, agentName, err)
			}
		}
	}

	return nil
}

// CANARY: REQ=CBIN-105; FEATURE="InitWorkflow"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17
// copyAndProcessAgentFiles copies agent files from embedded/.canary/agents/ to .canary/agents/
// and performs template variable substitution for {{ .AgentPrefix }}, {{ .AgentModel }}, {{ .AgentColor }}
func copyAndProcessAgentFiles(targetDir, agentPrefix, agentModel, agentColor string) error {
	// Try new path first, fall back to old path
	sourceAgentsDir := ".canary/agents"
	targetAgentsDir := filepath.Join(targetDir, ".canary", "agents")

	// Create target agents directory
	if err := os.MkdirAll(targetAgentsDir, 0755); err != nil {
		return fmt.Errorf("create agents directory: %w", err)
	}

	// Read agent files from embedded FS
	entries, err := embedded.CanaryFS.ReadDir(sourceAgentsDir)
	if err != nil {
		// Fall back to base path
		sourceAgentsDir = "base/.canary/agents"
		entries, err = embedded.CanaryFS.ReadDir(sourceAgentsDir)
		if err != nil {
			return fmt.Errorf("read agents directory: %w", err)
		}
	}

	// Process each agent file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		sourcePath := filepath.Join(sourceAgentsDir, entry.Name())
		targetPath := filepath.Join(targetAgentsDir, entry.Name())

		// Read the agent file
		content, err := embedded.CanaryFS.ReadFile(sourcePath)
		if err != nil {
			return fmt.Errorf("read agent file %s: %w", entry.Name(), err)
		}

		// Perform template substitution
		processedContent := string(content)
		processedContent = strings.ReplaceAll(processedContent, "{{ .AgentPrefix }}", agentPrefix)
		processedContent = strings.ReplaceAll(processedContent, "{{ .AgentModel }}", agentModel)
		processedContent = strings.ReplaceAll(processedContent, "{{ .AgentColor }}", agentColor)

		// Filter out CANARY CLI internal tokens (OWNER=canary)
		processedContent = string(filterCanaryTokens([]byte(processedContent)))

		// Write processed content to target
		if err := os.WriteFile(targetPath, []byte(processedContent), 0644); err != nil {
			return fmt.Errorf("write agent file %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// CANARY: REQ=CBIN-105; FEATURE="InitWorkflow"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17
// installAgentFilesToSystems copies agent files from embedded/.canary/agents/ to each agent system's agents directory
// This ensures agent definitions are available in each AI agent system (Claude, Cursor, etc.)
func installAgentFilesToSystems(targetDir string, agentsList []string, allAgentsFlag bool, agentPrefix, agentModel, agentColor string) error {
	// Try new path first, fall back to old path
	sourceAgentsDir := ".canary/agents"
	entries, err := embedded.CanaryFS.ReadDir(sourceAgentsDir)
	if err != nil {
		// Fall back to base path
		sourceAgentsDir = "base/.canary/agents"
		entries, err = embedded.CanaryFS.ReadDir(sourceAgentsDir)
		if err != nil {
			return fmt.Errorf("read agents directory: %w", err)
		}
	}

	// Agent configurations with agents subdirectory paths
	allAgents := map[string]string{
		"claude":    filepath.Join(targetDir, ".claude", "agents"),
		"cursor":    filepath.Join(targetDir, ".cursor", "agents"),
		"copilot":   filepath.Join(targetDir, ".github", "copilot", "agents"),
		"windsurf":  filepath.Join(targetDir, ".windsurf", "agents"),
		"kilocode":  filepath.Join(targetDir, ".kilocode", "agents"),
		"roo":       filepath.Join(targetDir, ".roo", "agents"),
		"opencode":  filepath.Join(targetDir, ".opencode", "agents"),
		"codex":     filepath.Join(targetDir, ".codex", "agents"),
		"auggie":    filepath.Join(targetDir, ".augment", "agents"),
		"codebuddy": filepath.Join(targetDir, ".codebuddy", "agents"),
		"amazonq":   filepath.Join(targetDir, ".amazonq", "agents"),
	}

	agentRootDirs := map[string]string{
		"claude":    filepath.Join(targetDir, ".claude"),
		"cursor":    filepath.Join(targetDir, ".cursor"),
		"copilot":   filepath.Join(targetDir, ".github"),
		"windsurf":  filepath.Join(targetDir, ".windsurf"),
		"kilocode":  filepath.Join(targetDir, ".kilocode"),
		"roo":       filepath.Join(targetDir, ".roo"),
		"opencode":  filepath.Join(targetDir, ".opencode"),
		"codex":     filepath.Join(targetDir, ".codex"),
		"auggie":    filepath.Join(targetDir, ".augment"),
		"codebuddy": filepath.Join(targetDir, ".codebuddy"),
		"amazonq":   filepath.Join(targetDir, ".amazonq"),
	}

	// Determine which agents to install for
	var selectedAgents map[string]string

	if allAgentsFlag {
		// Install for all agents
		selectedAgents = allAgents
	} else if len(agentsList) > 0 {
		// Install for specific agents
		selectedAgents = make(map[string]string)
		for _, agentName := range agentsList {
			if agentDir, ok := allAgents[agentName]; ok {
				selectedAgents[agentName] = agentDir
			} else {
				return fmt.Errorf("unknown agent: %s (valid: claude, cursor, copilot, windsurf, kilocode, roo, opencode, codex, auggie, codebuddy, amazonq)", agentName)
			}
		}
	} else {
		// Auto-detect existing agent directories
		selectedAgents = make(map[string]string)
		for agentName, rootDir := range agentRootDirs {
			if _, err := os.Stat(rootDir); err == nil {
				selectedAgents[agentName] = allAgents[agentName]
			}
		}
	}

	// If no agents selected, nothing to install
	if len(selectedAgents) == 0 {
		return nil
	}

	// Install agent files for selected agents
	for agentName, agentDir := range selectedAgents {
		// Create agents directory
		if err := os.MkdirAll(agentDir, 0755); err != nil {
			return fmt.Errorf("create %s agents directory: %w", agentName, err)
		}

		// Copy each agent file with template substitution
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			sourcePath := filepath.Join(sourceAgentsDir, entry.Name())
			targetPath := filepath.Join(agentDir, entry.Name())

			// Read the agent file
			content, err := embedded.CanaryFS.ReadFile(sourcePath)
			if err != nil {
				return fmt.Errorf("read agent file %s: %w", entry.Name(), err)
			}

			// Perform template substitution
			processedContent := string(content)
			processedContent = strings.ReplaceAll(processedContent, "{{ .AgentPrefix }}", agentPrefix)
			processedContent = strings.ReplaceAll(processedContent, "{{ .AgentModel }}", agentModel)
			processedContent = strings.ReplaceAll(processedContent, "{{ .AgentColor }}", agentColor)

			// Filter out CANARY CLI internal tokens (OWNER=canary)
			processedContent = string(filterCanaryTokens([]byte(processedContent)))

			// Write to target
			if err := os.WriteFile(targetPath, []byte(processedContent), 0644); err != nil {
				return fmt.Errorf("write agent file %s for %s: %w", entry.Name(), agentName, err)
			}
		}
	}

	return nil
}

// CANARY: REQ=CBIN-106; FEATURE="AgentContext"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
// createClaudeMD generates the CLAUDE.md file for AI agent integration
func createClaudeMD() string {
	return `# CANARY Development - AI Agent Guide

**Context File for AI Coding Agents**

This project uses CANARY requirement tracking with spec-kit-inspired workflows.

## Available Slash Commands

See [.canary/AGENT_CONTEXT.md](./.canary/AGENT_CONTEXT.md) for detailed information.

### Workflow Commands

- **/canary.constitution** - Create or update project governing principles
- **/canary.specify** - Create a new requirement specification from feature description
- **/canary.plan** - Generate technical implementation plan for a requirement
- **/canary.scan** - Scan codebase for CANARY tokens and generate reports
- **/canary.verify** - Verify GAP_ANALYSIS.md claims against actual implementation
- **/canary.update-stale** - Auto-update UPDATED field for stale tokens (>30 days)

### Command Definitions

All slash commands are defined in:
- ` + "`.canary/templates/commands/constitution.md`" + `
- ` + "`.canary/templates/commands/specify.md`" + `
- ` + "`.canary/templates/commands/plan.md`" + `
- ` + "`.canary/templates/commands/scan.md`" + `
- ` + "`.canary/templates/commands/verify.md`" + `
- ` + "`.canary/templates/commands/update-stale.md`" + `

## Quick Start Workflow

1. **Establish Principles**: ` + "`/canary.constitution Create principles for code quality and testing`" + `
2. **Define Requirement**: ` + "`/canary.specify Add user authentication with OAuth2 support`" + `
3. **Create Plan**: ` + "`/canary.plan CBIN-001 Use Go standard library with bcrypt`" + `
4. **Scan & Verify**: ` + "`/canary.scan`" + ` then ` + "`/canary.verify`" + `
5. **Update Stale**: ` + "`/canary.update-stale`" + ` (as needed)

## CANARY Token Format

` + "```" + `
// CANARY: REQ=CBIN-###; FEATURE="Name"; ASPECT=API; STATUS=IMPL; UPDATED=YYYY-MM-DD
` + "```" + `

**Status Progression:**
- STUB → IMPL → TESTED → BENCHED

**Valid Aspects:**
API, CLI, Engine, Storage, Security, Docs, Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist

## Constitutional Principles

See [.canary/memory/constitution.md](./.canary/memory/constitution.md) for full details.

**Core Principles:**
1. **Requirement-First**: Every feature starts with a CANARY token
2. **Test-First**: Tests written before implementation (Article IV)
3. **Evidence-Based**: Status promoted based on TEST=/BENCH= fields
4. **Simplicity**: Minimal complexity, prefer standard library
5. **Documentation Currency**: Keep tokens current with UPDATED field

## CLI Commands

` + "```bash" + `
# Initialize new project
canary init my-project

# Create requirement token
canary create CBIN-105 "FeatureName" --aspect API --status IMPL

# Scan for tokens
canary scan --root . --out status.json --csv status.csv

# Verify claims
canary scan --root . --verify GAP_ANALYSIS.md --strict

# Update stale tokens
canary scan --root . --update-stale
` + "```" + `

## Project Structure

` + "```" + `
.canary/
├── memory/
│   └── constitution.md          # Project principles
├── scripts/
│   └── create-new-requirement.sh # Automation
├── templates/
│   ├── commands/                # Slash command definitions
│   ├── spec-template.md         # Requirement template
│   └── plan-template.md         # Implementation plan template
└── specs/
    └── CBIN-XXX-feature/        # Individual requirements
        ├── spec.md
        └── plan.md

GAP_ANALYSIS.md                   # Requirement tracking
status.json                       # Scanner output
` + "```" + `

## For AI Agents

**Before implementing:**
1. Reference ` + "`.canary/memory/constitution.md`" + `
2. Use ` + "`/canary.specify`" + ` to create structured requirements
3. Follow test-first approach (Article IV)

**After implementing:**
1. Update CANARY tokens as code evolves
2. Run ` + "`/canary.scan`" + ` to verify status
3. Run ` + "`/canary.verify`" + ` to confirm claims

**Key Files:**
- [.canary/AGENT_CONTEXT.md](./.canary/AGENT_CONTEXT.md) - Complete context for AI agents
- [.canary/memory/constitution.md](./.canary/memory/constitution.md) - Constitutional principles
- [GAP_ANALYSIS.md](./GAP_ANALYSIS.md) - Requirement tracking
`
}

// createCmd generates a new CANARY token template
var createCmd = &cobra.Command{
	Use:   "create <req-id> <feature-name>",
	Short: "Generate a new CANARY token template",
	Long: `Create a properly formatted CANARY token for a new requirement.

Example:
  canary create CBIN-CLI-105 "UserProfile" --aspect CLI --status IMPL

Outputs a ready-to-paste CANARY token comment.`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		reqID := args[0]
		feature := args[1]

		aspect, _ := cmd.Flags().GetString("aspect")
		status, _ := cmd.Flags().GetString("status")
		owner, _ := cmd.Flags().GetString("owner")
		test, _ := cmd.Flags().GetString("test")
		bench, _ := cmd.Flags().GetString("bench")

		// Validate aspect
		if err := reqid.ValidateAspect(aspect); err != nil {
			return fmt.Errorf("invalid aspect: %w", err)
		}

		// Normalize aspect to canonical form
		aspect = reqid.NormalizeAspect(aspect)

		// Get today's date
		today := time.Now().UTC().Format("2006-01-02")

		// Build token
		token := fmt.Sprintf("// CANARY: REQ=%s; FEATURE=\"%s\"; ASPECT=%s; STATUS=%s",
			reqID, feature, aspect, status)

		if test != "" {
			token += fmt.Sprintf("; TEST=%s", test)
		}
		if bench != "" {
			token += fmt.Sprintf("; BENCH=%s", bench)
		}
		if owner != "" {
			token += fmt.Sprintf("; OWNER=%s", owner)
		}

		token += fmt.Sprintf("; UPDATED=%s", today)

		fmt.Println(token)
		fmt.Println("\n// Paste this above your implementation:")
		fmt.Printf("// func %s() { ... }\n", feature)

		return nil
	},
}

// CANARY: REQ=CBIN-119; FEATURE="ConstitutionCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var constitutionCmd = &cobra.Command{
	Use:   "constitution [description]",
	Short: "Create or update project governing principles",
	Long: `Create or update the project's constitutional principles in .canary/memory/constitution.md.

If no arguments are provided, creates the default constitution.
If arguments are provided, updates or adds specific principles.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		constitutionPath := ".canary/memory/constitution.md"

		if _, err := os.Stat(constitutionPath); os.IsNotExist(err) {
			// Read template from embedded FS
			content, err := embedded.CanaryFS.ReadFile("base/.canary/memory/constitution.md")
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

// CANARY: REQ=CBIN-120; FEATURE="SpecifyCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var specifyCmd = &cobra.Command{
	Use:   "specify <feature-description>",
	Short: "Create a new requirement specification",
	Long: `Create a new CANARY requirement specification from a feature description.

Generates a new requirement ID with aspect-based format (CBIN-<ASPECT>-XXX),
creates a spec directory, and populates it with a specification template.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
		templateContent, err := embedded.CanaryFS.ReadFile("base/.canary/templates/spec-template.md")
		if err != nil {
			return fmt.Errorf("read spec template: %w", err)
		}

		content := string(templateContent)
		content = strings.ReplaceAll(content, "CBIN-XXX", generatedID)
		content = strings.ReplaceAll(content, "[FEATURE NAME]", featureDesc)
		content = strings.ReplaceAll(content, "YYYY-MM-DD", time.Now().UTC().Format("2006-01-02"))
		content = strings.ReplaceAll(content, "<ASPECT>", aspect)

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

// CANARY: REQ=CBIN-133; FEATURE="ImplementCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_133_CLI_ExactMatch; OWNER=canary; DOC=user:docs/user/implement-command-guide.md; DOC_HASH=ed68fb1d97cf0562; UPDATED=2025-10-17
var implementCmd = &cobra.Command{
	Use:   "implement <query>",
	Short: "Generate implementation guidance for a requirement",
	Long: `Generate comprehensive implementation guidance for a requirement specification.

This command:
- Accepts requirement by ID (CBIN-XXX), name, or fuzzy search query
- Uses fuzzy matching with auto-selection for strong matches
- Generates complete implementation prompt including:
  - Specification details
  - Implementation plan
  - Constitutional principles
  - Implementation checklist
  - Progress tracking
  - Test-first guidance

Examples:
  canary implement CBIN-105              # Exact ID match
  canary implement "user auth"           # Fuzzy search
  canary implement UserAuthentication    # Feature name match
  canary implement --list                # List all unimplemented requirements`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		listFlag, _ := cmd.Flags().GetBool("list")
		promptFlag, _ := cmd.Flags().GetBool("prompt")

		// Handle --list flag
		if listFlag {
			return listUnimplemented()
		}

		// Require query argument if not listing
		if len(args) < 1 {
			return fmt.Errorf("requirement query is required (use --list to show all unimplemented)")
		}

		query := args[0]

		// Find requirement spec
		spec, err := findRequirement(query)
		if err != nil {
			return fmt.Errorf("find requirement: %w", err)
		}

		// Generate prompt
		flags := &ImplementFlags{
			Prompt: promptFlag,
		}

		prompt, err := renderImplementPrompt(spec, flags)
		if err != nil {
			return fmt.Errorf("generate prompt: %w", err)
		}

		fmt.Println(prompt)

		return nil
	},
}

// loadProjectConfig loads the .canary/project.yaml configuration
func loadProjectConfig() (*config.ProjectConfig, error) {
	return config.Load(".")
}

// extractField extracts a field value from a CANARY token string
func extractField(token, field string) string {
	// Look for FIELD="value" or FIELD=value
	pattern := field + `="([^"]+)"`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(token)
	if len(matches) > 1 {
		return matches[1]
	}

	// Try without quotes
	pattern = field + `=([^;\s]+)`
	re = regexp.MustCompile(pattern)
	matches = re.FindStringSubmatch(token)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// CANARY: REQ=CBIN-121; FEATURE="PlanCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var planCmd = &cobra.Command{
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
		templateContent, err := embedded.CanaryFS.ReadFile("base/.canary/templates/plan-template.md")
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
		content = strings.ReplaceAll(content, "<ASPECT>", aspect)

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

// CANARY: REQ=CBIN-124; FEATURE="IndexCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var indexCmd = &cobra.Command{
	Use:   "index [flags]",
	Short: "Build or rebuild the CANARY token database",
	Long: `Scan the codebase for CANARY tokens and store metadata in SQLite database.

This enables advanced features like priority ordering, keyword search, and checkpoints.
The database is stored at .canary/canary.db by default.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		rootPath, _ := cmd.Flags().GetString("root")

		fmt.Printf("Indexing CANARY tokens from: %s\n", rootPath)

		// Open or create database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}

		defer db.Close()

		// Get git info if in a repo
		var commitHash, branch string
		if gitCmd := exec.Command("git", "rev-parse", "HEAD"); gitCmd.Dir == "" {
			if output, err := gitCmd.Output(); err == nil {
				commitHash = strings.TrimSpace(string(output))
			}
		}
		if gitCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD"); gitCmd.Dir == "" {
			if output, err := gitCmd.Output(); err == nil {
				branch = strings.TrimSpace(string(output))
			}
		}

		// Scan for all CANARY tokens
		grepCmd := exec.Command("grep",
			"-rn",
			"--include=*.go", "--include=*.md", "--include=*.py",
			"--include=*.js", "--include=*.ts", "--include=*.java",
			"--include=*.rb", "--include=*.rs", "--include=*.c",
			"--include=*.cpp", "--include=*.h", "--include=*.sql",
			"CANARY:",
			rootPath,
		)

		output, err := grepCmd.CombinedOutput()
		if err != nil && len(output) == 0 {
			fmt.Println("No CANARY tokens found")
			return nil
		}

		// Parse and store tokens
		indexed := 0
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}

			// Parse grep output: file:line:content
			parts := strings.SplitN(line, ":", 3)
			if len(parts) < 3 {
				continue
			}

			file := parts[0]
			lineNum := 0
			//nolint:errcheck // Best-effort parse, default to 0 on failure
			fmt.Sscanf(parts[1], "%d", &lineNum)
			content := parts[2]

			// Extract all CANARY fields
			reqID := extractField(content, "REQ")
			feature := extractField(content, "FEATURE")
			aspect := extractField(content, "ASPECT")
			status := extractField(content, "STATUS")

			if reqID == "" || feature == "" {
				continue // Skip malformed tokens
			}

			// Build token struct
			docPath := extractField(content, "DOC")
			docType := extractField(content, "DOC_TYPE")

			// Auto-infer DOC_TYPE from type prefix if not explicitly set
			if docPath != "" && docType == "" {
				// Extract type from first doc path (e.g., "user:docs/file.md" -> "user")
				firstPath := strings.Split(docPath, ",")[0]
				if strings.Contains(firstPath, ":") {
					docType = strings.Split(firstPath, ":")[0]
				}
			}

			token := &storage.Token{
				ReqID:       reqID,
				Feature:     feature,
				Aspect:      aspect,
				Status:      status,
				FilePath:    file,
				LineNumber:  lineNum,
				Test:        extractField(content, "TEST"),
				Bench:       extractField(content, "BENCH"),
				Owner:       extractField(content, "OWNER"),
				Phase:       extractField(content, "PHASE"),
				Keywords:    extractField(content, "KEYWORDS"),
				SpecStatus:  extractField(content, "SPEC_STATUS"),
				UpdatedAt:   extractField(content, "UPDATED"),
				CreatedAt:   extractField(content, "CREATED"),
				StartedAt:   extractField(content, "STARTED"),
				CompletedAt: extractField(content, "COMPLETED"),
				CommitHash:  commitHash,
				Branch:      branch,
				DependsOn:   extractField(content, "DEPENDS_ON"),
				Blocks:      extractField(content, "BLOCKS"),
				RelatedTo:   extractField(content, "RELATED_TO"),
				DocPath:     docPath,
				DocHash:     extractField(content, "DOC_HASH"),
				DocType:     docType,
				RawToken:    content,
				IndexedAt:   time.Now().UTC().Format(time.RFC3339),
			}

			// Parse priority
			if priorityStr := extractField(content, "PRIORITY"); priorityStr != "" {
				if p, err := strconv.Atoi(priorityStr); err == nil {
					token.Priority = p
				} else {
					token.Priority = 5 // default
				}
			} else {
				token.Priority = 5 // default
			}

			// Set defaults
			if token.UpdatedAt == "" {
				token.UpdatedAt = time.Now().UTC().Format("2006-01-02")
			}
			if token.SpecStatus == "" {
				token.SpecStatus = "draft"
			}

			// Store in database
			if err := db.UpsertToken(token); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to store token %s/%s: %v\n", reqID, feature, err)
				continue
			}

			indexed++
		}

		fmt.Printf("\n✅ Indexed %d CANARY tokens\n", indexed)
		fmt.Printf("Database: %s\n", dbPath)

		if commitHash != "" {
			fmt.Printf("Commit: %s\n", commitHash[:8])
		}
		if branch != "" {
			fmt.Printf("Branch: %s\n", branch)
		}

		return nil
	},
}

// CANARY: REQ=CBIN-125; FEATURE="ListCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var listCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "List CANARY tokens with filtering and ordering",
	Long: `List tokens from the database with priority ordering and filtering.

Supports filtering by status, aspect, phase, owner, and spec status.
Results are ordered by priority (1=highest) and updated date by default.

By default, hides requirements from:
- Test files (*_test.go, /tests/, /test/)
- Template directories (.canary/templates/, /base/, /embedded/)
- Documentation examples (IMPLEMENTATION_SUMMARY, FINAL_SUMMARY, etc.)
- AI agent directories (.claude/, .cursor/, .github/prompts/, etc.)

Use --include-hidden to show all requirements including hidden ones.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		filterStatus, _ := cmd.Flags().GetString("status")
		filterAspect, _ := cmd.Flags().GetString("aspect")
		filterPhase, _ := cmd.Flags().GetString("phase")
		filterOwner, _ := cmd.Flags().GetString("owner")
		filterSpecStatus, _ := cmd.Flags().GetString("spec-status")
		priorityMin, _ := cmd.Flags().GetInt("priority-min")
		priorityMax, _ := cmd.Flags().GetInt("priority-max")
		orderBy, _ := cmd.Flags().GetString("order-by")
		limit, _ := cmd.Flags().GetInt("limit")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		includeHidden, _ := cmd.Flags().GetBool("include-hidden")

		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}

		defer db.Close()

		// Load project config to get ID pattern
		cfg, err := loadProjectConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load project config: %v\n", err)
		}
		idPattern := ""
		if cfg != nil && cfg.Requirements.IDPattern != "" {
			idPattern = cfg.Requirements.IDPattern
		}

		// Build filters
		filters := make(map[string]string)
		if filterStatus != "" {
			filters["status"] = filterStatus
		}
		if filterAspect != "" {
			filters["aspect"] = filterAspect
		}
		if filterPhase != "" {
			filters["phase"] = filterPhase
		}
		if filterOwner != "" {
			filters["owner"] = filterOwner
		}
		if filterSpecStatus != "" {
			filters["spec_status"] = filterSpecStatus
		}
		if priorityMin > 0 {
			filters["priority_min"] = strconv.Itoa(priorityMin)
		}
		if priorityMax > 0 {
			filters["priority_max"] = strconv.Itoa(priorityMax)
		}
		if includeHidden {
			filters["include_hidden"] = "true"
		}

		tokens, err := db.ListTokens(filters, idPattern, orderBy, limit)
		if err != nil {
			return fmt.Errorf("list tokens: %w", err)
		}

		if len(tokens) == 0 {
			fmt.Println("No tokens found")
			return nil
		}

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(tokens)
		}

		// Display as table
		fmt.Printf("Found %d tokens:\n\n", len(tokens))
		for _, token := range tokens {
			fmt.Printf("📌 %s - %s\n", token.ReqID, token.Feature)
			fmt.Printf("   Status: %s | Aspect: %s | Priority: %d", token.Status, token.Aspect, token.Priority)
			if token.Phase != "" {
				fmt.Printf(" | Phase: %s", token.Phase)
			}
			fmt.Println()
			fmt.Printf("   Location: %s:%d\n", token.FilePath, token.LineNumber)
			if token.Test != "" {
				fmt.Printf("   Test: %s\n", token.Test)
			}
			if token.Owner != "" {
				fmt.Printf("   Owner: %s\n", token.Owner)
			}
			if token.Keywords != "" {
				fmt.Printf("   Keywords: %s\n", token.Keywords)
			}
			fmt.Println()
		}

		return nil
	},
}

// CANARY: REQ=CBIN-126; FEATURE="SearchCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var searchCmd = &cobra.Command{
	Use:   "search <keywords>",
	Short: "Search CANARY tokens by keywords",
	Long: `Search tokens by keywords in feature names, requirement IDs, and keyword tags.

Keywords are matched case-insensitively using LIKE queries.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		keywords := strings.Join(args, " ")

		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}

		defer db.Close()

		tokens, err := db.SearchTokens(keywords)
		if err != nil {
			return fmt.Errorf("search tokens: %w", err)
		}

		if len(tokens) == 0 {
			fmt.Printf("No tokens found for: %s\n", keywords)
			return nil
		}

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(tokens)
		}

		fmt.Printf("Search results for '%s' (%d tokens):\n\n", keywords, len(tokens))
		for _, token := range tokens {
			fmt.Printf("📌 %s - %s\n", token.ReqID, token.Feature)
			fmt.Printf("   Status: %s | Priority: %d | %s:%d\n",
				token.Status, token.Priority, token.FilePath, token.LineNumber)
			if token.Keywords != "" {
				fmt.Printf("   Tags: %s\n", token.Keywords)
			}
			fmt.Println()
		}

		return nil
	},
}

// CANARY: REQ=CBIN-127; FEATURE="PrioritizeCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var prioritizeCmd = &cobra.Command{
	Use:   "prioritize <REQ-ID> <feature> <priority>",
	Short: "Update priority of a CANARY token",
	Long: `Update the priority of a specific token (1=highest, 10=lowest).

Priority affects ordering in list and search results.`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		reqID := args[0]
		feature := args[1]
		priority, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid priority: %s (must be 1-10)", args[2])
		}

		if priority < 1 || priority > 10 {
			return fmt.Errorf("priority must be between 1 (highest) and 10 (lowest)")
		}

		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}

		defer db.Close()

		if err := db.UpdatePriority(reqID, feature, priority); err != nil {
			return fmt.Errorf("update priority: %w", err)
		}

		fmt.Printf("✅ Updated priority for %s/%s to %d\n", reqID, feature, priority)
		return nil
	},
}

// CANARY: REQ=CBIN-128; FEATURE="CheckpointCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var checkpointCmd = &cobra.Command{
	Use:   "checkpoint <name> [description]",
	Short: "Create a state snapshot checkpoint",
	Long: `Create a checkpoint to capture current state of all tokens.

Checkpoints include:
- Counts by status (STUB, IMPL, TESTED, BENCHED)
- Commit hash and timestamp
- Full JSON snapshot of all tokens

Useful for tracking progress over time.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		name := args[0]
		description := ""
		if len(args) > 1 {
			description = strings.Join(args[1:], " ")
		}

		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}

		defer db.Close()

		// Get current commit hash
		commitHash := ""
		if gitCmd := exec.Command("git", "rev-parse", "HEAD"); gitCmd.Dir == "" {
			if output, err := gitCmd.Output(); err == nil {
				commitHash = strings.TrimSpace(string(output))
			}
		}

		// Load project config for ID pattern filtering
		cfg, _ := loadProjectConfig()
		idPattern := ""
		if cfg != nil && cfg.Requirements.IDPattern != "" {
			idPattern = cfg.Requirements.IDPattern
		}

		// Get all tokens for snapshot
		tokens, err := db.ListTokens(nil, idPattern, "", 0)
		if err != nil {
			return fmt.Errorf("get tokens: %w", err)
		}

		snapshotJSON, err := json.Marshal(tokens)
		if err != nil {
			return fmt.Errorf("marshal snapshot: %w", err)
		}

		if err := db.CreateCheckpoint(name, description, commitHash, string(snapshotJSON)); err != nil {
			return fmt.Errorf("create checkpoint: %w", err)
		}

		fmt.Printf("✅ Created checkpoint: %s\n", name)
		if commitHash != "" {
			fmt.Printf("Commit: %s\n", commitHash[:8])
		}
		fmt.Printf("Tokens: %d\n", len(tokens))

		return nil
	},
}

// CANARY: REQ=CBIN-129; FEATURE="MigrateCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var migrateCmd = &cobra.Command{
	Use:   "migrate <steps>",
	Short: "Run database migrations",
	Long: `Apply database migrations to the CANARY database.

Steps can be:
  - "all" to migrate to the latest version
  - A positive integer to migrate forward by that many steps
  - A negative integer to roll back by that many steps`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		steps := args[0]

		fmt.Printf("Running migrations on: %s\n", dbPath)

		if err := storage.MigrateDB(dbPath, steps); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		fmt.Println("✅ Migrations completed successfully")
		return nil
	},
}

// CANARY: REQ=CBIN-129; FEATURE="RollbackCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var rollbackCmd = &cobra.Command{
	Use:   "rollback <steps>",
	Short: "Roll back database migrations",
	Long: `Roll back database migrations.

Steps can be:
  - "all" to roll back all migrations
  - A positive integer to roll back by that many steps`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		steps := args[0]

		fmt.Printf("Rolling back migrations on: %s\n", dbPath)

		if err := storage.TeardownDB(dbPath, steps); err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}

		fmt.Println("✅ Rollback completed successfully")
		return nil
	},
}

// CANARY: REQ=CBIN-131; FEATURE="DetectCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var detectCmd = &cobra.Command{
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

// CANARY: REQ=CBIN-132; FEATURE="NextCmd"; ASPECT=CLI; STATUS=BENCHED; TEST=TestCANARY_CBIN_132_CLI_NextPrioritySelection; BENCH=BenchmarkCANARY_CBIN_132_CLI_PriorityQuery; OWNER=canary; DOC=user:docs/user/next-priority-guide.md; DOC_HASH=17524f7a14d2c410; UPDATED=2025-10-17
var nextCmd = &cobra.Command{
	Use:   "next [flags]",
	Short: "Identify and implement next highest priority requirement",
	Long: `Identify the next highest priority unimplemented requirement and generate
comprehensive implementation guidance.

This command automatically:
- Queries database or scans filesystem for CANARY tokens
- Identifies highest priority STUB or IMPL requirement
- Excludes hidden requirements (test files, templates, examples)
- Verifies dependencies are satisfied
- Generates comprehensive implementation prompt with:
  - Specification details
  - Constitutional principles
  - Test-first guidance
  - Token placement examples

Priority determination factors:
1. PRIORITY field (1=highest, 10=lowest)
2. STATUS (STUB > IMPL > TESTED)
3. DEPENDS_ON (dependencies must be TESTED/BENCHED)
4. UPDATED field (older tokens get priority boost)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		promptFlag, _ := cmd.Flags().GetBool("prompt")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		filterStatus, _ := cmd.Flags().GetString("status")
		filterAspect, _ := cmd.Flags().GetString("aspect")

		// Build filters
		filters := make(map[string]string)
		if filterStatus != "" {
			filters["status"] = filterStatus
		}
		if filterAspect != "" {
			filters["aspect"] = filterAspect
		}

		// Select next priority
		token, err := selectNextPriority(dbPath, filters)
		if err != nil {
			return fmt.Errorf("select next priority: %w", err)
		}

		if token == nil {
			fmt.Println("🎉 All requirements completed! No work available.")
			fmt.Println("\nSuggestions:")
			fmt.Println("  • Run: canary scan --verify GAP_ANALYSIS.md")
			fmt.Println("  • Review completed requirements")
			fmt.Println("  • Consider creating new specifications")
			return nil
		}

		if dryRun {
			fmt.Printf("Next priority (dry run): %s - %s\n", token.ReqID, token.Feature)
			fmt.Printf("Priority: %d | Status: %s | Aspect: %s\n", token.Priority, token.Status, token.Aspect)
			fmt.Printf("Location: %s\n", token.FilePath)
			return nil
		}

		// Render prompt
		output, err := renderPrompt(token, promptFlag)
		if err != nil {
			return fmt.Errorf("render prompt: %w", err)
		}

		if jsonOutput {
			// TODO: Implement JSON output format
			fmt.Println("{\"error\": \"JSON output not yet implemented\"}")
			return nil
		}

		fmt.Println(output)
		return nil
	},
}

// CANARY: REQ=CBIN-131; FEATURE="MigrateFromCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var migrateFromCmd = &cobra.Command{
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

func init() {
	// Configure slog to use ERROR level by default to reduce noise
	opts := &slog.HandlerOptions{
		Level: slog.LevelError,
	}
	handler := slog.NewTextHandler(os.Stderr, opts)
	slog.SetDefault(slog.New(handler))

	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(constitutionCmd)
	rootCmd.AddCommand(specifyCmd)
	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(implementCmd)
	rootCmd.AddCommand(nextCmd)
	rootCmd.AddCommand(indexCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(prioritizeCmd)
	rootCmd.AddCommand(checkpointCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(detectCmd)
	rootCmd.AddCommand(migrateFromCmd)
	rootCmd.AddCommand(orphanCmd)
	rootCmd.AddCommand(docCmd)
	// CANARY: REQ=CBIN-CLI-001; FEATURE="ShowCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_ShowCmd; UPDATED=2025-10-16
	rootCmd.AddCommand(showCmd)
	// CANARY: REQ=CBIN-CLI-001; FEATURE="FilesCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_FilesCmd; UPDATED=2025-10-16
	rootCmd.AddCommand(filesCmd)
	// CANARY: REQ=CBIN-CLI-001; FEATURE="StatusCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_StatusCmd; UPDATED=2025-10-16
	rootCmd.AddCommand(statusCmd)
	// CANARY: REQ=CBIN-CLI-001; FEATURE="GrepCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_GrepCmd; UPDATED=2025-10-16
	rootCmd.AddCommand(grepCmd)
	// CANARY: REQ=CBIN-140; FEATURE="GapCmd"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-17
	rootCmd.AddCommand(gapCmd)
	// CANARY: REQ=CBIN-145; FEATURE="SpecsCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_145_CLI_SpecsCmd; UPDATED=2025-10-17
	rootCmd.AddCommand(specsCmd)

	// initCmd flags
	initCmd.Flags().StringSlice("agents", []string{}, "comma-separated list of agents to install for (claude,cursor,copilot,windsurf,kilocode,roo,opencode,codex,auggie,codebuddy,amazonq)")
	initCmd.Flags().Bool("all-agents", false, "install commands for all supported agents")
	initCmd.Flags().String("key", "", "project requirement ID prefix (e.g., CBIN, PROJ, ACME)")
	initCmd.Flags().String("agent-prefix", "", "agent name prefix for CANARY agents (default: project key)")
	initCmd.Flags().String("agent-model", "sonnet", "AI model for CANARY agents")
	initCmd.Flags().String("agent-color", "blue", "color for CANARY agents")

	// specifyCmd flags
	specifyCmd.Flags().String("aspect", "Engine", "requirement aspect (API, CLI, Engine, Storage, Security, Docs, Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist)")

	// planCmd flags
	planCmd.Flags().String("aspect", "", "requirement aspect for template substitution (API, CLI, Engine, Storage, Security, Docs, Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist)")

	// createCmd flags
	createCmd.Flags().String("aspect", "API", "requirement aspect/category")
	createCmd.Flags().String("status", "IMPL", "implementation status")
	createCmd.Flags().String("owner", "", "team/person responsible")
	createCmd.Flags().String("test", "", "test function name")
	createCmd.Flags().String("bench", "", "benchmark function name")

	// implementCmd flags
	implementCmd.Flags().Bool("list", false, "list all unimplemented requirements")
	implementCmd.Flags().Bool("prompt", true, "generate full implementation prompt (default: true)")

	// indexCmd flags
	indexCmd.Flags().String("db", ".canary/canary.db", "path to database file")
	indexCmd.Flags().String("root", ".", "root directory to scan")

	// listCmd flags
	listCmd.Flags().String("db", ".canary/canary.db", "path to database file")
	listCmd.Flags().String("status", "", "filter by status (STUB, IMPL, TESTED, BENCHED)")
	listCmd.Flags().String("aspect", "", "filter by aspect (API, CLI, Engine, etc.)")
	listCmd.Flags().String("phase", "", "filter by phase (Phase0, Phase1, Phase2, Phase3)")
	listCmd.Flags().String("owner", "", "filter by owner")
	listCmd.Flags().String("spec-status", "", "filter by spec status (draft, approved, in-progress, completed, archived)")
	listCmd.Flags().Int("priority-min", 0, "filter by minimum priority (0 = no minimum)")
	listCmd.Flags().Int("priority-max", 0, "filter by maximum priority (0 = no maximum)")
	listCmd.Flags().String("order-by", "", "custom ORDER BY clause (default: priority ASC, updated_at DESC)")
	listCmd.Flags().Int("limit", 0, "maximum number of results (0 = no limit)")
	listCmd.Flags().Bool("json", false, "output as JSON")
	listCmd.Flags().Bool("include-hidden", false, "include hidden requirements (test files, templates, examples)")

	// searchCmd flags
	searchCmd.Flags().String("db", ".canary/canary.db", "path to database file")
	searchCmd.Flags().Bool("json", false, "output as JSON")

	// prioritizeCmd flags
	prioritizeCmd.Flags().String("db", ".canary/canary.db", "path to database file")

	// checkpointCmd flags
	checkpointCmd.Flags().String("db", ".canary/canary.db", "path to database file")

	// migrateCmd flags
	migrateCmd.Flags().String("db", ".canary/canary.db", "path to database file")

	// rollbackCmd flags
	rollbackCmd.Flags().String("db", ".canary/canary.db", "path to database file")

	// migrateFromCmd flags
	migrateFromCmd.Flags().Bool("dry-run", false, "preview changes without applying them")
	migrateFromCmd.Flags().Bool("force", false, "force migration even if system type doesn't match detection")

	// scanCmd flags
	scanCmd.Flags().String("root", ".", "root directory to scan")
	scanCmd.Flags().String("out", "status.json", "output status.json path")
	scanCmd.Flags().String("csv", "", "optional status.csv path")
	scanCmd.Flags().String("verify", "", "GAP_ANALYSIS file to verify claims")
	scanCmd.Flags().Bool("strict", false, "enforce staleness on TESTED/BENCHED tokens (30 days)")
	scanCmd.Flags().Bool("update-stale", false, "rewrite UPDATED field for stale tokens")
	scanCmd.Flags().String("skip", "", "skip path regex (RE2)")
	scanCmd.Flags().Bool("project-only", false, "filter by project requirement ID pattern")

	// nextCmd flags
	nextCmd.Flags().String("db", ".canary/canary.db", "path to database file")
	nextCmd.Flags().Bool("prompt", false, "generate full implementation prompt (default: summary only)")
	nextCmd.Flags().Bool("json", false, "output in JSON format")
	nextCmd.Flags().Bool("dry-run", false, "show what would be selected without generating prompt")
	nextCmd.Flags().String("status", "", "filter by status (STUB, IMPL, TESTED, BENCHED)")
	nextCmd.Flags().String("aspect", "", "filter by aspect (API, CLI, Engine, Storage, etc.)")
}
