package init

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"devnw.dev/canary/pkg/cmds/internal/utils"
	"devnw.dev/canary/pkg/safewrite"
)

// writeManagedFile writes one file this command authors, staged and confined
// to root.
//
// An existing file whose bytes differ is replaced only when force is set;
// otherwise it is kept and the skip is reported on stderr. A bootstrap command
// runs repeatedly over a repository people are editing, so overwriting by
// default is how customized files get destroyed -- keeping them, and saying
// so, is the safe default. With --force the prior bytes are preserved in a
// .bak beside the file.
// CANARY: REQ=ENG-4331; FEATURE="InitForce"; ASPECT=CLI; STATUS=TESTED; TEST=TestAuditF03,TestAuditF03Force; UPDATED=2026-08-30
func writeManagedFile(root, path string, data []byte, mode fs.FileMode, force bool) error {
	res, err := safewrite.Write(path, data, mode, safewrite.Options{
		Root:   root,
		Force:  force,
		Backup: force,
	})
	if errors.Is(err, safewrite.ErrWouldReplace) {
		fmt.Fprintf(os.Stderr, "notice: kept existing %s (use --force to overwrite)\n", path)
		return nil
	}
	if err != nil {
		return err
	}
	if res.Replaced {
		fmt.Fprintf(os.Stderr, "notice: replaced %s (previous content saved to %s)\n", path, res.BackupPath)
		printDiff(os.Stderr, res.Diff)
	}
	return nil
}

// maxDiffLines bounds how much of a replaced file's diff is echoed, so
// overwriting a large file reports what changed without burying the rest of
// the command's output.
const maxDiffLines = 20

// printDiff writes a bounded rendering of a safewrite diff.
func printDiff(w io.Writer, diff string) {
	if diff == "" {
		return
	}
	lines := strings.Split(diff, "\n")
	shown := lines
	if len(shown) > maxDiffLines {
		shown = shown[:maxDiffLines]
	}
	for _, l := range shown {
		fmt.Fprintf(w, "  %s\n", l)
	}
	if len(lines) > len(shown) {
		fmt.Fprintf(w, "  ... and %d more changed line(s)\n", len(lines)-len(shown))
	}
}

// CANARY: REQ=ENG-4300; FEATURE="InitWorkflow"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
// copyCanaryStructure copies the embedded base/ directory structure to the target .canary/ project directory
func copyCanaryStructure(targetDir string) error {
	targetCanary := filepath.Join(targetDir, ".canary")

	// Create root .canary directory
	if err := os.MkdirAll(targetCanary, 0750); err != nil {
		return err
	}

	// The embedded files are in "base/" and should be copied to ".canary/"
	// Walk the embedded base directory
	return fs.WalkDir(embedded, "base", func(path string, d fs.DirEntry, err error) error {
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
			return os.MkdirAll(targetPath, 0750)
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(targetPath), 0750); err != nil {
			return err
		}

		// Read embedded file
		content, err := embedded.ReadFile(path)
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

// CANARY: REQ=ENG-4300; FEATURE="InitWorkflow"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
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

	if err := os.WriteFile(path, []byte(text), 0640); err != nil {
		return fmt.Errorf("write project.yaml: %w", err)
	}

	return nil
}

// CANARY: REQ=ENG-4300; FEATURE="InitWorkflow"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
// AgentConfig defines configuration for each supported AI agent
type AgentConfig struct {
	Dir    string // Directory for agent files
	Prefix string // Prefix for command files (e.g., "canary.")
}

func codexPromptsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".codex", "prompts"), nil
}

// installSlashCommands copies slash commands to agent systems based on selection mode
// agentsList: specific agents to install for (e.g., ["claude", "cursor"])
// allAgentsFlag: if true, install for all supported agents
// localInstall: if true, install in project directory; if false, install globally in home directory
// force: allow replacing an already-installed command file whose content differs
// If both are empty/false, auto-detect existing agent directories
func installSlashCommands(targetDir string, agentsList []string, allAgentsFlag bool, localInstall, force bool) ([]string, error) {
	sourceDir := filepath.Join(targetDir, ".canary", "templates", "commands")
	notes := []string{}

	// Determine base directory for installation
	var baseDir string
	if localInstall {
		// Local installation: use project directory
		baseDir = targetDir
		fmt.Println("📍 Installing commands in project-local locations where supported...")
	} else {
		// Global installation: use home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home directory: %w", err)
		}
		baseDir = homeDir
		fmt.Println("🌍 Installing commands globally in home directory...")
	}

	codexDir, err := codexPromptsDir()
	if err != nil {
		return nil, err
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
		"codex":     {Dir: codexDir, Prefix: "canary."},
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
		"codex":     filepath.Dir(codexDir),
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
				return nil, fmt.Errorf("unknown agent: %s (valid: claude, cursor, copilot, windsurf, kilocode, roo, opencode, codex, auggie, codebuddy, amazonq)", agentName)
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
		return nil, nil
	}

	if localInstall {
		if _, ok := selectedAgents["codex"]; ok {
			notes = append(notes, fmt.Sprintf("Codex custom prompts were installed globally in %s because the current Codex CLI scans ~/.codex/prompts instead of project-local prompt directories.", codexDir))
		}
	}

	// Read all command files from source
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("read commands directory: %w", err)
	}

	// Install commands for selected agents
	for agentName, config := range selectedAgents {
		// Create agent directory
		if err := os.MkdirAll(config.Dir, 0750); err != nil {
			return nil, fmt.Errorf("create %s directory: %w", agentName, err)
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
				return nil, fmt.Errorf("read command file %s: %w", entry.Name(), err)
			}

			// Write to target with prefix
			if err := writeManagedFile(config.Dir, targetPath, content, 0640, force); err != nil {
				return nil, fmt.Errorf("write command file %s for %s: %w", targetName, agentName, err)
			}
		}
	}

	return notes, nil
}

// CANARY: REQ=ENG-4300; FEATURE="InitWorkflow"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17
// copyAndProcessAgentFiles copies agent files from embedded/.canary/agents/ to .canary/agents/
// and performs template variable substitution for {{ .AgentPrefix }}, {{ .AgentModel }}, {{ .AgentColor }}
func copyAndProcessAgentFiles(targetDir, agentPrefix, agentModel, agentColor string, force bool) error {
	// Agent files are in base/agents/
	sourceAgentsDir := "base/agents"
	targetAgentsDir := filepath.Join(targetDir, ".canary", "agents")

	// Create target agents directory
	if err := os.MkdirAll(targetAgentsDir, 0750); err != nil {
		return fmt.Errorf("create agents directory: %w", err)
	}

	// Read agent files from embedded FS
	entries, err := embedded.ReadDir(sourceAgentsDir)
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
		content, err := embedded.ReadFile(sourcePath)
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
		if err := writeManagedFile(targetDir, targetPath, []byte(processedContent), 0640, force); err != nil {
			return fmt.Errorf("write agent file %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// CANARY: REQ=ENG-4300; FEATURE="InitWorkflow"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17
// installAgentFilesToSystems copies agent files from embedded/.canary/agents/ to each agent system's agents directory
// This ensures agent definitions are available in each AI agent system (Claude, Cursor, etc.)
func installAgentFilesToSystems(targetDir string, agentsList []string, allAgentsFlag bool, agentPrefix, agentModel, agentColor string, localInstall, force bool) error {
	// Agent files are in base/agents/
	sourceAgentsDir := "base/agents"
	entries, err := embedded.ReadDir(sourceAgentsDir)
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
		if err := os.MkdirAll(agentDir, 0750); err != nil {
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
			content, err := embedded.ReadFile(sourcePath)
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
			if err := writeManagedFile(agentDir, targetPath, []byte(processedContent), 0640, force); err != nil {
				return fmt.Errorf("write agent file %s for %s: %w", entry.Name(), agentName, err)
			}
		}
	}

	return nil
}

// CANARY: REQ=ENG-4320; FEATURE="AgentContextUpdate"; ASPECT=CLI; STATUS=TESTED; TEST=TestAuditF11,TestAuditF11NoDuplicateSection,TestUpdateAgentContextFilesCreatesAGENTS; OWNER=canary; UPDATED=2026-08-30
// updateAgentContextFiles updates all agent context files with gated CANARY
// sections. Each file is attempted independently and every failure is
// reported together: one file with malformed markers must not stop the other
// agent systems from being configured, and it must not be silently skipped
// either.
func updateAgentContextFiles(projectName string, force bool) error {
	agentContextContent, err := utils.ReadEmbeddedFile("base/AGENT_CONTEXT.md")
	if err != nil {
		return fmt.Errorf("read AGENT_CONTEXT.md: %w", err)
	}

	gated := []struct {
		path    string
		content string
	}{
		{filepath.Join(projectName, "CLAUDE.md"), createClaudeMD()},
		{filepath.Join(projectName, "CURSOR.md"), createCursorMD()},
		{filepath.Join(projectName, "AGENTS.md"), createCodexAGENTSMD()},
	}

	// Update .github/copilot-instructions.md only when .github exists.
	githubDir := filepath.Join(projectName, ".github")
	if _, serr := os.Stat(githubDir); serr == nil {
		gated = append(gated, struct {
			path    string
			content string
		}{filepath.Join(githubDir, "copilot-instructions.md"), createCopilotInstructionsMD()})
	}

	var errs []error
	for _, g := range gated {
		if uerr := updateMarkdownSection(g.path, g.content); uerr != nil {
			errs = append(errs, uerr)
		}
	}

	// Update .canary/AGENT_CONTEXT.md (embedded file is already correct)
	agentContextPath := filepath.Join(projectName, ".canary", "AGENT_CONTEXT.md")
	filteredContent := utils.FilterCanaryTokens(agentContextContent)
	if werr := writeManagedFile(projectName, agentContextPath, filteredContent, 0640, force); werr != nil {
		errs = append(errs, fmt.Errorf("write AGENT_CONTEXT.md: %w", werr))
	}

	// Create Cursor-specific files (rules + optional MCP) so Cursor IDE and plugins work well
	if cerr := createCursorRuleAndMCP(projectName, force); cerr != nil {
		errs = append(errs, fmt.Errorf("create Cursor rule/MCP: %w", cerr))
	}

	return errors.Join(errs...)
}

// createCursorRuleAndMCP creates .cursor/rules/canary-requirements.mdc and .cursor/mcp.json (if missing)
// so Cursor IDE and Cursor plugins get CANARY guidance and optional MCP tool access.
func createCursorRuleAndMCP(projectName string, force bool) error {
	cursorDir := filepath.Join(projectName, ".cursor")
	rulesDir := filepath.Join(cursorDir, "rules")
	if err := os.MkdirAll(rulesDir, 0750); err != nil {
		return err
	}

	// Cursor rule: apply when working on requirements/specs; minimal so it doesn't bloat every chat
	rulePath := filepath.Join(rulesDir, "canary-requirements.mdc")
	ruleContent := "---\ndescription: CANARY requirement tracking — use when editing requirements, specs, or CANARY tokens\n" +
		"globs: [\"**/.canary/**\", \"**/GAP_ANALYSIS.md\", \"**/*spec*.md\", \"**/plan.md\"]\nalwaysApply: false\n---\n\n" +
		"# CANARY requirements\n\nThis project uses CANARY. For the command you're running, read only that file under .canary/commands/ (e.g. scan.md for /canary.scan).\n\n" +
		"- **Scan:** Use the one-line stdout from `canary scan` (CANARY_SCAN tokens=...) for metrics; do not read status.json unless needed.\n" +
		"- **Verify:** Use stdout CANARY_VERIFY_OK or stderr CANARY_VERIFY_FAIL lines; open GAP_ANALYSIS.md only when fixing claims.\n" +
		"- **Token format:** `// CANARY: REQ=ID; FEATURE=\"Name\"; ASPECT=API; STATUS=IMPL; UPDATED=YYYY-MM-DD` — status STUB→IMPL→TESTED→BENCHED.\n"
	if err := writeManagedFile(projectName, rulePath, []byte(ruleContent), 0640, force); err != nil {
		return err
	}

	// Optional MCP config: only create if missing so we don't overwrite user's other servers
	mcpPath := filepath.Join(cursorDir, "mcp.json")
	if _, err := os.Stat(mcpPath); err == nil {
		return nil // already exists
	}
	mcpContent := `{
  "mcpServers": {
    "canary": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
`
	// Cursor may expect streamable HTTP; url alone is often enough. User must run `canary mcp` first.
	// The existence check above already guarantees this is a create, never a
	// replacement, so the user's other MCP servers are never touched.
	if err := writeManagedFile(projectName, mcpPath, []byte(mcpContent), 0640, force); err != nil {
		return err
	}
	return nil
}

// CANARY: REQ=ENG-4301; FEATURE="AgentContext"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
// createClaudeMD generates the CANARY section for CLAUDE.md (Claude Code / Claude plugins).
// Kept minimal to reduce agent context; load-only-what-you-need.
func createClaudeMD() string {
	b := "# CANARY Development - AI Agent Guide\n\n**Minimal context:** This project uses CANARY requirement tracking. Load only what you need.\n\n## Commands (load only the one you use)\n\n"
	bt := "`"
	b += "- **/canary.constitution** → " + bt + ".canary/commands/constitution.md" + bt + "\n"
	b += "- **/canary.specify** → " + bt + ".canary/commands/specify.md" + bt + "\n"
	b += "- **/canary.plan** → " + bt + ".canary/commands/plan.md" + bt + "\n"
	b += "- **/canary.scan** → " + bt + ".canary/commands/scan.md" + bt + "\n"
	b += "- **/canary.verify** → " + bt + ".canary/commands/verify.md" + bt + "\n"
	b += "- **/canary.update-stale** → " + bt + ".canary/commands/update-stale.md" + bt + "\n\n"
	b += "Read **only** the command file for the slash command you are running. Do not load AGENT_CONTEXT, full constitution, or GAP_ANALYSIS unless that command's instructions tell you to.\n\n"
	b += "## Scan & verify (low context)\n\n"
	b += "- **Scan:** Run " + bt + "canary scan --root . --out status.json" + bt + ". Use the **one line** on stdout (" + bt + "CANARY_SCAN tokens=N requirements=M STUB=..." + bt + ") for metrics. Do not read " + bt + "status.json" + bt + " unless you need per-requirement detail.\n"
	b += "- **Verify:** Run " + bt + "canary scan --root . --verify GAP_ANALYSIS.md --strict" + bt + ". Use stdout (" + bt + "CANARY_VERIFY_OK" + bt + " or " + bt + "CANARY_VERIFY_FAIL count=N" + bt + ") and stderr for failures. Open " + bt + "GAP_ANALYSIS.md" + bt + " only when fixing or updating claims.\n\n"
	b += "## Token format\n\n"
	b += bt + "// CANARY: REQ=CBIN-###; FEATURE=\"Name\"; ASPECT=API; STATUS=IMPL; UPDATED=YYYY-MM-DD" + bt + "\nStatus: STUB → IMPL → TESTED → BENCHED. Aspects: API, CLI, Engine, Storage, Security, Docs, Wire, Planner, Decode, Encode, RoundTrip, Bench, FrontEnd, Dist.\n\n"
	b += "## Principles (inline; load constitution only when editing principles)\n\n1. Requirement-First: every feature has a CANARY token\n2. Test-First: tests before implementation (Article IV)\n3. Evidence-Based: status from TEST=/BENCH=\n4. Simplicity: prefer standard library\n5. Documentation Currency: keep UPDATED current\n\n"
	b += "## Plugins & MCP (optional)\n\nFor **tool-based** access (list, show, scan, create, etc.) from Claude Code or Cursor: run " + bt + "canary mcp" + bt + " (HTTP server on port 8080), then add the MCP server in your IDE:\n"
	b += "- **Claude Code:** Add to " + bt + "~/.claude.json" + bt + " or use MCP settings; URL " + bt + "http://localhost:8080/mcp" + bt + " (streamable HTTP).\n"
	b += "- **Cursor:** Add to " + bt + ".cursor/mcp.json" + bt + " in this project; see docs/MCP_QUICK_START.md.\n"
	return b
}

// CANARY: REQ=ENG-4320; FEATURE="AgentContextUpdate"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-11-01
// createCursorMD generates the CANARY section for CURSOR.md (Cursor IDE / Cursor plugins).
func createCursorMD() string {
	bt := "`"
	return createClaudeMD() + "\n\n## Cursor-specific\n\n" +
		"- **Project rules:** CANARY rule is in " + bt + ".cursor/rules/canary-requirements.mdc" + bt + " (apply when working on requirements or specs).\n" +
		"- **MCP:** Optional " + bt + ".cursor/mcp.json" + bt + " is created by " + bt + "canary init" + bt + "; start " + bt + "canary mcp" + bt + " in a terminal, then Cursor can use list/show/scan tools.\n"
}

// createCodexAGENTSMD generates the CANARY section for AGENTS.md (Codex / repository-scoped instructions).
func createCodexAGENTSMD() string {
	bt := "`"
	return "# Repository Guidelines\n\n" +
		"This project uses CANARY requirement tracking. Load only the context you need.\n\n" +
		"## Slash Commands\n\n" +
		"- " + bt + "/canary.constitution" + bt + " -> " + bt + ".canary/commands/constitution.md" + bt + "\n" +
		"- " + bt + "/canary.specify" + bt + " -> " + bt + ".canary/commands/specify.md" + bt + "\n" +
		"- " + bt + "/canary.plan" + bt + " -> " + bt + ".canary/commands/plan.md" + bt + "\n" +
		"- " + bt + "/canary.scan" + bt + " -> " + bt + ".canary/commands/scan.md" + bt + "\n" +
		"- " + bt + "/canary.verify" + bt + " -> " + bt + ".canary/commands/verify.md" + bt + "\n" +
		"- " + bt + "/canary.update-stale" + bt + " -> " + bt + ".canary/commands/update-stale.md" + bt + "\n\n" +
		"Read only the command file for the slash command you are running. Do not load " + bt + ".canary/AGENT_CONTEXT.md" + bt + ", the full constitution, or " + bt + "GAP_ANALYSIS.md" + bt + " unless the task requires it.\n\n" +
		"## Scan and verify\n\n" +
		"- " + bt + "canary scan --root . --out status.json" + bt + " -> use the one-line stdout summary (" + bt + "CANARY_SCAN ..." + bt + ") for metrics.\n" +
		"- " + bt + "canary scan --root . --verify GAP_ANALYSIS.md --strict" + bt + " -> use stdout/stderr results before opening extra files.\n\n" +
		"## Token format\n\n" +
		bt + "// CANARY: REQ=CBIN-###; FEATURE=\"Name\"; ASPECT=API; STATUS=IMPL; UPDATED=YYYY-MM-DD" + bt + "\n" +
		"Status: STUB -> IMPL -> TESTED -> BENCHED.\n\n" +
		"## Codex\n\n" +
		"- Custom prompts are installed by " + bt + "canary init" + bt + " in " + bt + "~/.codex/prompts/" + bt + ".\n" +
		"- Keep repository-specific guidance in this file; " + bt + "canary init" + bt + " preserves content outside the CANARY gated section.\n"
}

// CANARY: REQ=ENG-4320; FEATURE="AgentContextUpdate"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-11-01
// createCopilotInstructionsMD generates the CANARY section for .github/copilot-instructions.md (GitHub Copilot / plugins).
func createCopilotInstructionsMD() string {
	return "# CANARY Development Guide for GitHub Copilot\n\n" +
		"This project uses CANARY requirement tracking. Load only the command file you need from `.canary/commands/`.\n\n" +
		"## Token format\n" +
		"`// CANARY: REQ=CBIN-###; FEATURE=\"Name\"; ASPECT=API; STATUS=IMPL; UPDATED=YYYY-MM-DD`\n" +
		"Status: STUB → IMPL → TESTED → BENCHED.\n\n" +
		"## Scan & verify (low context)\n" +
		"- **Scan:** `canary scan --root . --out status.json` — use the one-line stdout (`CANARY_SCAN tokens=...`) for metrics.\n" +
		"- **Verify:** `canary scan --root . --verify GAP_ANALYSIS.md --strict` — use stdout `CANARY_VERIFY_OK` or stderr for failures.\n\n" +
		"## Principles\n" +
		"1. Requirement-First 2. Test-First 3. Evidence-Based (TEST=/BENCH=) 4. Simplicity 5. Keep UPDATED current.\n\n" +
		"See [.canary/AGENT_CONTEXT.md](./.canary/AGENT_CONTEXT.md) for full reference.\n"
}
