package init

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.devnw.com/canary/embedded"
	"go.devnw.com/canary/cli/internal/utils"
)

// CANARY: REQ=CBIN-105; FEATURE="InitWorkflow"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
// copyCanaryStructure copies the embedded base/ directory structure to the target .canary/ project directory
func copyCanaryStructure(targetDir string) error {
	targetCanary := filepath.Join(targetDir, ".canary")

	// Create root .canary directory
	if err := os.MkdirAll(targetCanary, 0755); err != nil {
		return err
	}

	// The embedded files are in "base/" and should be copied to ".canary/"
	// Walk the embedded base directory
	return fs.WalkDir(embedded.CanaryFS, "base", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the base directory itself
		if path == "base" {
			return nil
		}

		// Skip certain files that are handled separately
		if strings.HasSuffix(path, ".canaryignore") {
			return nil // This is handled separately in the init command
		}

		// Get relative path from base/
		relPath := strings.TrimPrefix(path, "base/")
		if relPath == "" {
			return nil
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
			content = utils.FilterCanaryTokens(content)
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
// localInstall: if true, install in project directory; if false, install globally in home directory
// If both are empty/false, auto-detect existing agent directories
func installSlashCommands(targetDir string, agentsList []string, allAgentsFlag bool, localInstall bool) error {
	sourceDir := filepath.Join(targetDir, ".canary", "templates", "commands")

	// Determine base directory for installation
	var baseDir string
	if localInstall {
		// Local installation: use project directory
		baseDir = targetDir
		fmt.Println("📍 Installing commands locally in project directory...")
	} else {
		// Global installation: use home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home directory: %w", err)
		}
		baseDir = homeDir
		fmt.Println("🌍 Installing commands globally in home directory...")
	}

	// Agent configurations - matching spec-kit-repo structure
	allAgents := map[string]AgentConfig{
		"claude":    {Dir: filepath.Join(baseDir, ".claude", "commands"), Prefix: "canary."},
		"cursor":    {Dir: filepath.Join(baseDir, ".cursor", "commands"), Prefix: "canary."},
		"copilot":   {Dir: filepath.Join(baseDir, ".github", "prompts"), Prefix: "canary-"},
		"windsurf":  {Dir: filepath.Join(baseDir, ".windsurf", "workflows"), Prefix: "canary-"},
		"kilocode":  {Dir: filepath.Join(baseDir, ".kilocode", "rules"), Prefix: "canary-"},
		"roo":       {Dir: filepath.Join(baseDir, ".roo", "rules"), Prefix: "canary-"},
		"opencode":  {Dir: filepath.Join(baseDir, ".opencode", "command"), Prefix: "canary-"},
		"codex":     {Dir: filepath.Join(baseDir, ".codex", "commands"), Prefix: "canary."},
		"auggie":    {Dir: filepath.Join(baseDir, ".augment", "rules"), Prefix: "canary-"},
		"codebuddy": {Dir: filepath.Join(baseDir, ".codebuddy", "commands"), Prefix: "canary."},
		"amazonq":   {Dir: filepath.Join(baseDir, ".amazonq", "prompts"), Prefix: "canary-"},
	}

	agentRootDirs := map[string]string{
		"claude":    filepath.Join(baseDir, ".claude"),
		"cursor":    filepath.Join(baseDir, ".cursor"),
		"copilot":   filepath.Join(baseDir, ".github"),
		"windsurf":  filepath.Join(baseDir, ".windsurf"),
		"kilocode":  filepath.Join(baseDir, ".kilocode"),
		"roo":       filepath.Join(baseDir, ".roo"),
		"opencode":  filepath.Join(baseDir, ".opencode"),
		"codex":     filepath.Join(baseDir, ".codex"),
		"auggie":    filepath.Join(baseDir, ".augment"),
		"codebuddy": filepath.Join(baseDir, ".codebuddy"),
		"amazonq":   filepath.Join(baseDir, ".amazonq"),
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
	// Agent files are in base/agents/
	sourceAgentsDir := "base/agents"
	targetAgentsDir := filepath.Join(targetDir, ".canary", "agents")

	// Create target agents directory
	if err := os.MkdirAll(targetAgentsDir, 0755); err != nil {
		return fmt.Errorf("create agents directory: %w", err)
	}

	// Read agent files from embedded FS
	entries, err := embedded.CanaryFS.ReadDir(sourceAgentsDir)
	if err != nil {
		return fmt.Errorf("read agents directory: %w", err)
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
		processedContent = string(utils.FilterCanaryTokens([]byte(processedContent)))

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
func installAgentFilesToSystems(targetDir string, agentsList []string, allAgentsFlag bool, agentPrefix, agentModel, agentColor string, localInstall bool) error {
	// Agent files are in base/agents/
	sourceAgentsDir := "base/agents"
	entries, err := embedded.CanaryFS.ReadDir(sourceAgentsDir)
	if err != nil {
		return fmt.Errorf("read agents directory: %w", err)
	}

	// Determine base directory for installation
	var baseDir string
	if localInstall {
		// Local installation: use project directory
		baseDir = targetDir
	} else {
		// Global installation: use home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home directory: %w", err)
		}
		baseDir = homeDir
	}

	// Agent configurations with agents subdirectory paths
	allAgents := map[string]string{
		"claude":    filepath.Join(baseDir, ".claude", "agents"),
		"cursor":    filepath.Join(baseDir, ".cursor", "agents"),
		"copilot":   filepath.Join(baseDir, ".github", "copilot", "agents"),
		"windsurf":  filepath.Join(baseDir, ".windsurf", "agents"),
		"kilocode":  filepath.Join(baseDir, ".kilocode", "agents"),
		"roo":       filepath.Join(baseDir, ".roo", "agents"),
		"opencode":  filepath.Join(baseDir, ".opencode", "agents"),
		"codex":     filepath.Join(baseDir, ".codex", "agents"),
		"auggie":    filepath.Join(baseDir, ".augment", "agents"),
		"codebuddy": filepath.Join(baseDir, ".codebuddy", "agents"),
		"amazonq":   filepath.Join(baseDir, ".amazonq", "agents"),
	}

	agentRootDirs := map[string]string{
		"claude":    filepath.Join(baseDir, ".claude"),
		"cursor":    filepath.Join(baseDir, ".cursor"),
		"copilot":   filepath.Join(baseDir, ".github"),
		"windsurf":  filepath.Join(baseDir, ".windsurf"),
		"kilocode":  filepath.Join(baseDir, ".kilocode"),
		"roo":       filepath.Join(baseDir, ".roo"),
		"opencode":  filepath.Join(baseDir, ".opencode"),
		"codex":     filepath.Join(baseDir, ".codex"),
		"auggie":    filepath.Join(baseDir, ".augment"),
		"codebuddy": filepath.Join(baseDir, ".codebuddy"),
		"amazonq":   filepath.Join(baseDir, ".amazonq"),
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
			processedContent = string(utils.FilterCanaryTokens([]byte(processedContent)))

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
