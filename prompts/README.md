# Canary Prompts - Hierarchical Structure

## Overview

The Canary prompts package provides a hierarchical organization of prompts for all CLI commands. Each command has its own dedicated prompt file describing its purpose, behavior, and standards.

## Directory Structure

```
prompts/
??? commands/              # Command-specific prompts
?   ??? scan/
?   ?   ??? scan.md       # Scan command prompt
?   ??? list/
?   ?   ??? list.md       # List command prompt
?   ??? show/
?   ?   ??? show.md       # Show command prompt
?   ??? create/
?   ?   ??? create.md     # Create command prompt
?   ??? status/
?   ?   ??? status.md     # Status command prompt
?   ??? search/
?   ?   ??? search.md     # Search command prompt
?   ??? next/
?   ?   ??? next.md       # Next command prompt
?   ??? specify/
?   ?   ??? specify.md    # Specify command prompt
?   ??? plan/
?   ?   ??? plan.md       # Plan command prompt
?   ??? implement/
?   ?   ??? implement.md  # Implement command prompt
?   ??? index/
?   ?   ??? index.md      # Index command prompt
?   ??? files/
?   ?   ??? files.md      # Files command prompt
?   ??? grep/
?   ?   ??? grep.md       # Grep command prompt
?   ??? prioritize/
?   ?   ??? prioritize.md # Prioritize command prompt
?   ??? bug/
?   ?   ??? bug.md        # Bug command prompt
?   ??? gap/
?   ?   ??? gap.md        # Gap command prompt
?   ??? checkpoint/
?   ?   ??? checkpoint.md # Checkpoint command prompt
?   ??? constitution/
?   ?   ??? constitution.md # Constitution command prompt
?   ??? deps/
?   ?   ??? deps.md       # Dependencies command prompt
?   ??? doc/
?   ?   ??? doc.md        # Documentation command prompt
?   ??? migrate/
?   ?   ??? migrate.md    # Migration command prompt
?   ??? project/
?   ?   ??? project.md    # Project command prompt
?   ??? specs/
?   ?   ??? specs.md      # Specs command prompt
?   ??? db/
?   ?   ??? db.md         # Database command prompt
?   ??? mcp/
?       ??? mcp.md        # MCP server command prompt
??? sys/                   # System-level prompts (legacy)
?   ??? init.md           # Initialization prompt
?   ??? policy.md         # Policy prompt
?   ??? requirements.md   # Requirements prompt
?   ??? evaluate.md       # Evaluation prompt
??? prompts.go            # Go API for accessing prompts
??? prompts_test.go       # Tests for prompt API
??? README.md             # This file
```

## Usage

### Go API

```go
import "go.devnw.com/canary/prompts"

// Get a specific command prompt
content, err := prompts.GetCommand("scan")
if err != nil {
    log.Fatal(err)
}
fmt.Println(content)

// List all available commands
commands, err := prompts.ListCommands()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Available commands: %v\n", commands)

// Get all command prompts as a map
allPrompts, err := prompts.GetAllCommands()
if err != nil {
    log.Fatal(err)
}

// Parse a command prompt into structured data
prompt, err := prompts.ParseCommandPrompt("scan")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Command: %s\n", prompt.Command)
fmt.Printf("Full content: %s\n", prompt.FullContent)
```

### Legacy System Prompts

```go
// Legacy API still supported
all := prompts.All()
fmt.Println(all["init"])
fmt.Println(all["policy"])

// Or use individual variables
fmt.Println(prompts.Init)
fmt.Println(prompts.Policy)
fmt.Println(prompts.Requirements)
fmt.Println(prompts.Evaluate)
```

## Command Prompt Structure

Each command prompt file follows this structure:

### Header
```markdown
# [Command Name] Command Prompt

## Purpose
Brief description of what the command does.
```

### Task Description
```markdown
## Task
Detailed description of what needs to be implemented.
```

### Expected Behavior
```markdown
## Expected Behavior
Examples of how the command should work:

\`\`\`bash
canary command --flag value
\`\`\`
```

### Output Format
```markdown
## Output Format
Description and examples of command output.
```

### Standards
```markdown
## Standards
- Implementation guidelines
- Best practices
- Performance requirements
- Error handling expectations
```

## Categories

### Core Token Management
Commands that deal with CANARY tokens directly:
- **list** - List/filter tokens
- **show** - Show requirement details
- **create** - Generate token template
- **status** - Show progress stats
- **search** - Search by keywords
- **next** - Get next priority requirement

### Workflow & Development
Commands for the development workflow:
- **scan** - Scan codebase for tokens
- **specify** - Create requirement specification
- **plan** - Generate implementation plan
- **implement** - Get implementation guidance
- **index** - Index tokens to database

### Query & Navigation
Commands for exploring the codebase:
- **files** - Find files for requirement
- **grep** - Search by pattern in fields

### Management
Commands for project management:
- **prioritize** - Set requirement priority
- **checkpoint** - Create state snapshots
- **constitution** - Manage project principles
- **deps** - Manage dependencies
- **project** - Manage project configuration

### Bug Tracking
Commands for bug management:
- **bug** - Bug lifecycle management (list, create, show, update)
- **gap** - Gap analysis feedback

### Documentation
Commands for documentation:
- **doc** - Generate documentation
- **specs** - Manage specifications

### Infrastructure
Commands for system management:
- **db** - Database operations
- **migrate** - Schema migrations
- **mcp** - MCP server for AI assistants

## Adding New Commands

To add a new command prompt:

1. Create a new directory: `commands/yourcommand/`
2. Create the prompt file: `commands/yourcommand/yourcommand.md`
3. Follow the standard prompt structure
4. The prompt will be automatically embedded and available via the API

Example:
```bash
mkdir -p prompts/commands/newcmd
cat > prompts/commands/newcmd/newcmd.md << 'EOF'
# New Command Prompt

## Purpose
Description of the new command.

## Task
Implementation details.

## Expected Behavior
Usage examples.

## Standards
Guidelines and requirements.
EOF
```

The new prompt will be automatically available:
```go
content, err := prompts.GetCommand("newcmd")
```

## Testing

Run tests to verify all prompts are accessible:

```bash
go test ./prompts/...
```

## Embedding

All prompts are embedded into the binary using Go's `embed` directive. This means:
- No external files needed at runtime
- Fast access (no disk I/O)
- Version-controlled with code
- Single binary distribution

## Benefits

### For Developers
- Clear guidelines for each command
- Consistent command behavior
- Easy to reference during implementation

### For AI Assistants
- Structured prompts for each tool
- Clear expectations and standards
- Examples and formats

### For Documentation
- Single source of truth
- Auto-generated docs possible
- Version-controlled

## Future Enhancements

Potential improvements:
- [ ] Structured parsing of prompt sections
- [ ] Validation of prompt completeness
- [ ] Generation of command documentation from prompts
- [ ] Template variable substitution
- [ ] Multi-language prompt support
- [ ] Prompt versioning system

## Related Files

- `cli/` - CLI command implementations
- `mcp/` - MCP server that exposes commands to AI
- `internal/storage/` - Database layer used by commands
- `.canary/templates/` - User-facing templates

## License

Copyright (c) 2025 by Developer Network.

For more details, see the LICENSE file in the root directory of this
source code repository or contact Developer Network at info@devnw.com.
