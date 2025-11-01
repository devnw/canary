# Canary Init - Intelligent Markdown Updates

## ? Complete: Idempotent Agent Context File Updates

Successfully updated `canary init` to intelligently update agent context files using gated sections with HTML comment markers, allowing the command to be run multiple times without losing user customizations.

## Changes Made

### 1. New Markdown Section Updater (`cli/init/markdown_updater.go`)

Created three key functions for intelligent markdown updates:

#### `updateMarkdownSection(filePath, sectionContent string)`
- Updates or inserts a single gated CANARY section
- Uses markers: `<!-- CANARY:START -->` ... `<!-- CANARY:END -->`
- Preserves all non-CANARY content
- Creates file if it doesn't exist
- Idempotent: can be run multiple times safely

#### `updateMultipleMarkdownSections(filePath string, sections map[string]string)`
- Updates multiple named sections in one file
- Uses keyed markers: `<!-- CANARY:key:START -->` ... `<!-- CANARY:key:END -->`
- Each section can be updated independently
- Preserves user content outside markers

#### `removeMarkdownSection(filePath, sectionKey string)`
- Cleanly removes a gated section if needed
- Useful for deprecating sections

### 2. Updated Init Flow (`cli/init/init.go`)

Changed from:
```go
// OLD: Overwrites entire file
claudeMD := createClaudeMD()
os.WriteFile(claudePath, []byte(claudeMD), 0644)
```

To:
```go
// NEW: Updates only CANARY section
if err := updateAgentContextFiles(projectName); err != nil {
    return fmt.Errorf("update agent context files: %w", err)
}
```

### 3. New Function: `updateAgentContextFiles()` (`cli/init/canary.go`)

Intelligently updates all agent context files:
- **CLAUDE.md** - Gated CANARY section
- **CURSOR.md** - Gated CANARY section  
- **.canary/AGENT_CONTEXT.md** - Full file (embedded)
- **.github/copilot-instructions.md** - Gated CANARY section (if .github exists)

### 4. New Content Functions

Created specialized content generators:
- `createClaudeMD()` - Claude-specific CANARY guide
- `createCursorMD()` - Cursor-specific CANARY guide
- `createCopilotInstructionsMD()` - GitHub Copilot guide

## Gated Section Format

### Single Section
```markdown
# My Custom Content

This is preserved across updates.

<!-- CANARY:START -->
## CANARY Development Guide

[CANARY-specific content here - updated by canary init]
<!-- CANARY:END -->

More custom content below.
```

### Multiple Sections
```markdown
<!-- CANARY:intro:START -->
Introduction content
<!-- CANARY:intro:END -->

User content in between

<!-- CANARY:commands:START -->
Commands list
<!-- CANARY:commands:END -->
```

## Behavior

### First Run (New File)
```
canary init
```
Creates files with gated CANARY sections.

### Subsequent Runs (Updates)
```
canary init
```
- ? Updates CANARY sections with latest content
- ? Preserves all user customizations outside markers
- ? Maintains file structure
- ? Idempotent - safe to run multiple times

### Example Workflow

```bash
# First time
canary init myproject
# Creates: CLAUDE.md with CANARY section

# User edits CLAUDE.md, adds custom sections:
# My Custom Content
# <!-- CANARY:START -->
# (CANARY content)
# <!-- CANARY:END -->
# More custom notes

# Later, run again to update CANARY sections
canary init
# Result: Custom content preserved, CANARY section updated
```

## Files Updated

| File | Update Method | Behavior |
|------|---------------|----------|
| `CLAUDE.md` | Gated section | Preserves user content |
| `CURSOR.md` | Gated section | Preserves user content |
| `.canary/AGENT_CONTEXT.md` | Full overwrite | Embedded file (no user edits expected) |
| `.github/copilot-instructions.md` | Gated section | Preserves user content |
| `README_CANARY.md` | Full overwrite | Generated file |
| `GAP_ANALYSIS.md` | Full overwrite | Template file |

## Testing

### Unit Tests (`cli/init/markdown_updater_test.go`)

Comprehensive test coverage:

1. **CreateNewFile**: Creates file with gated content
2. **PreserveUserContent**: Updates CANARY section, preserves user content
3. **AddSectionToExistingFile**: Adds CANARY section to file without markers
4. **IdempotentUpdates**: Multiple runs produce identical results
5. **CreateMultipleSections**: Handles multiple named sections
6. **UpdateOneSectionPreserveOthers**: Updates one section without affecting others

All tests passing:
```bash
$ go test ./cli/init/... -v -run TestUpdateMarkdown
=== RUN   TestUpdateMarkdownSection
=== RUN   TestUpdateMarkdownSection/CreateNewFile
=== RUN   TestUpdateMarkdownSection/PreserveUserContent
=== RUN   TestUpdateMarkdownSection/AddSectionToExistingFile
=== RUN   TestUpdateMarkdownSection/IdempotentUpdates
--- PASS: TestUpdateMarkdownSection (0.00s)
PASS
ok      go.devnw.com/canary/cli/init    0.005s
```

## Agent Context Files

### CLAUDE.md
Full CANARY development guide with:
- Available slash commands
- Workflow steps
- Token format
- Constitutional principles
- CLI commands
- Project structure

### CURSOR.md
Same content as CLAUDE.md (both use same format)

### .github/copilot-instructions.md
Streamlined guide for GitHub Copilot with:
- Token format
- Status values
- Key principles
- Basic workflow
- Essential commands

### .canary/AGENT_CONTEXT.md
Complete embedded file with:
- Detailed workflow documentation
- All available commands
- Best practices
- Examples

## Benefits

### For Users
? **Safe Updates**: Run `canary init` anytime without losing customizations
? **Latest Features**: Get updated CANARY documentation automatically
? **Flexibility**: Add custom sections anywhere outside markers
? **Consistency**: Same CANARY content across all agent systems

### For Development
? **Maintainable**: Update content in one place (embedded files)
? **Testable**: Comprehensive unit tests
? **Extensible**: Easy to add new gated sections
? **Predictable**: Idempotent behavior

## Future Enhancements

### Potential Additions
- [ ] Version markers to track CANARY content versions
- [ ] Migration helpers for old format files
- [ ] Validation of marker integrity
- [ ] Diff viewer for section changes
- [ ] Backup creation before updates

### Additional Agent Systems
- [ ] Windsurf instructions
- [ ] Amazon Q context
- [ ] Auggie rules
- [ ] CodeBuddy instructions

## Implementation Details

### Key Design Decisions

1. **HTML Comments**: Used `<!-- CANARY:START -->` markers because:
   - Invisible in rendered markdown
   - Preserves markdown validity
   - Clear, searchable markers
   - Standard HTML comment syntax

2. **Section Keys**: Support both keyed and unkeyed markers:
   - Unkeyed: `<!-- CANARY:START -->` for single section
   - Keyed: `<!-- CANARY:key:START -->` for multiple sections

3. **Idempotency**: Designed for multiple executions:
   - Replaces only content between markers
   - Never duplicates markers
   - Preserves file structure

4. **Error Handling**: Graceful failures:
   - Creates files if missing
   - Skips optional files (.github) if parent doesn't exist
   - Returns clear error messages

### Code Organization

```
cli/init/
??? init.go                    # Main init command
??? canary.go                  # CANARY-specific logic
??? markdown_updater.go        # NEW: Gated section updater
??? markdown_updater_test.go   # NEW: Comprehensive tests
??? copilot.go                 # Copilot instructions
```

## Migration from Previous Version

### Old Behavior (Before)
```bash
canary init
# Overwrites CLAUDE.md completely
# User customizations lost
```

### New Behavior (After)
```bash
canary init
# Updates only CANARY sections
# User customizations preserved
```

### For Existing Projects

If you have an existing `CLAUDE.md` without markers:
1. Run `canary init`
2. CANARY section appended to end
3. Original content preserved
4. Markers added for future updates

## Examples

### Example 1: First-Time Setup
```bash
$ canary init myproject
? Initialized CANARY project in: myproject

Created:
  ? .canary/ - Full workflow structure
  ? CLAUDE.md - AI agent slash command integration
  ? CURSOR.md - Cursor-specific integration
  ? .canary/AGENT_CONTEXT.md - Complete context
```

### Example 2: Update Existing Project
```bash
$ cd myproject
$ canary init
?? Existing CANARY project detected - updating...

? Updated CANARY project in: .

Updated:
  ? .canary/ - Full workflow structure
  ? Agent context files - CANARY sections refreshed
```

### Example 3: Custom CLAUDE.md
```markdown
# My Project - Custom Guide

I added my own content here for my team.

## Team Conventions
- Use feat/ branches
- Squash merges only

<!-- CANARY:START -->
# CANARY Development - AI Agent Guide

[Updated automatically by canary init]
<!-- CANARY:END -->

## My Custom Notes
- Remember to run tests locally
- Check coverage reports
```

After `canary init`:
- "My Project - Custom Guide" section: ? Preserved
- "Team Conventions" section: ? Preserved
- CANARY section: ? Updated
- "My Custom Notes" section: ? Preserved

## Conclusion

The updated `canary init` command now provides **intelligent, idempotent updates** to agent context files, allowing users to:
- Safely update CANARY documentation
- Preserve custom configurations
- Maintain consistency across agent systems
- Run init multiple times without fear

**All requirements met!** ?
