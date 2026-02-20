# CLI Refactor Summary

## Overview
Refactored the Canary CLI to improve structure and add `--prompt` argument support to all subcommands.

## Changes Made

### 1. Split Bug Command Subcommands
- **Created separate files** for each bug subcommand:
  - `cli/bug/list.go` - List bug tokens with filtering
  - `cli/bug/create.go` - Create new bug tokens
  - `cli/bug/update.go` - Update existing bug tokens
  - `cli/bug/show.go` - Display bug token details
  - `cli/bug/bug.go` - Main command and helper functions

- **Benefits**:
  - Better code organization
  - Easier to maintain and test
  - Follows single responsibility principle
  - Each subcommand is self-contained with its own init()

### 2. Added --prompt Flag to All Commands
Added `--prompt` flag (stubbed for future use) to all CLI commands:

#### Commands with --prompt added:
- `create` - Generate CANARY token template
- `list` - List tokens with filtering
- `show` - Display token details
- `status` - Show implementation progress
- `specify` - Create requirement specification
- `plan` - Generate implementation plan
- `scan` - Scan for CANARY tokens
- `search` - Search tokens by keywords
- `index` - Build token database
- `files` - List implementation files
- `grep` - Search tokens by pattern
- `implement` - Generate implementation guidance (uses --prompt-arg)
- `next` - Identify next priority (uses --prompt-arg)
- `prioritize` - Update token priority
- `checkpoint` - Create state snapshot
- `gap` (all subcommands) - Manage gap analysis

#### Commands with subcommands:
- `bug list` - Added --prompt flag
- `bug create` - Added --prompt flag
- `bug update` - Added --prompt flag
- `bug show` - Added --prompt flag
- `gap mark`, `gap query`, etc. - Added --prompt as persistent flag

### 3. Moved Flag Registrations to init()
- Moved flag registrations from `cmd/canary/main.go` to individual command `init()` functions
- Reduces duplication and centralizes command configuration
- Makes commands more self-contained

**Removed from main.go:**
- create command flags (aspect, status, owner, test, bench)
- list command flags (db, status, aspect, phase, owner, etc.)
- search command flags (db, json)
- scan command flags (root, out, csv, verify, etc.)
- specify command flags (aspect)
- plan command flags (aspect)
- index command flags (db, root)
- prioritize command flags (db)
- checkpoint command flags (db)

**Kept in main.go:**
- init command flags (local, agents, key, etc.)
- migrate/rollback command flags
- legacy command flags
- next command flags (db, prompt, json, dry-run, status, aspect)

### 4. Created Prompt Helper Utility
Created `cli/internal/utils/prompt.go` with:

#### Functions:
- `LoadPrompt(promptArg string)` - Load custom prompts from file or embedded FS
- `loadPromptFromFile(path string)` - Load from filesystem
- `loadEmbeddedPrompt(name string)` - Load from embedded prompts (stub)
- `ValidatePromptArg(promptArg string)` - Validate prompt argument format
- `GetAvailablePrompts()` - List available embedded prompts (stub)

#### Future Implementation:
The helper functions are stubbed and ready for:
- Loading prompts from embedded FS (`prompts/sys/*.md`, `prompts/commands/*.md`)
- Loading custom project prompts from `.canary/templates/*.md`
- Template variable substitution
- Prompt validation and caching

## File Structure Changes

### Before:
```
cli/bug/bug.go (single file with all subcommands)
cmd/canary/main.go (all flag registrations)
```

### After:
```
cli/bug/
??? bug.go          # Main command + helpers
??? list.go         # List subcommand
??? create.go       # Create subcommand
??? update.go       # Update subcommand
??? show.go         # Show subcommand

cli/internal/utils/
??? prompt.go       # Prompt helper utilities

cmd/canary/main.go  # Reduced flag registrations
```

## Usage Examples

### Using --prompt flag:
```bash
# Will be used to load custom prompts in the future
canary create CBIN-200 "Feature" --prompt /path/to/custom/prompt.md
canary list --prompt embedded:list-detailed
canary bug create "Issue" --prompt .canary/templates/bug-prompt.md
```

### Bug command subcommands:
```bash
# List all bugs
canary bug list

# Create new bug
canary bug create "Login fails" --aspect API --severity S1

# Update bug status
canary bug update BUG-API-001 --status FIXED

# Show bug details
canary bug show BUG-API-001
```

## Testing

All changes have been tested:
- Binary builds successfully: ?
- Bug command and subcommands work: ?
- --prompt flag appears in all command help: ?
- Existing tests pass: ?
- No duplicate flag registrations: ?

## Implementation Notes

### TODO Comments
All --prompt flag usage includes TODO comments:
```go
// TODO: Implement --prompt flag to load custom prompts
prompt, _ := cmd.Flags().GetString("prompt")
_ = prompt // Stubbed for future use
```

### Flag Pattern
Consistent pattern across all commands:
```go
func init() {
    Cmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
    // ... other flags
}
```

### Gap Command
Uses PersistentFlags for all subcommands:
```go
GapCmd.PersistentFlags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
GapCmd.PersistentFlags().String("db", ".canary/canary.db", "path to database file")
```

## Next Steps

To fully implement the --prompt functionality:

1. **Embed Prompts**: Add prompt templates to embedded FS
2. **Implement Loaders**: Complete `loadEmbeddedPrompt()` function
3. **Add Templates**: Create command-specific prompt templates
4. **Variable Substitution**: Add template variable support
5. **Caching**: Implement prompt caching for performance
6. **Validation**: Add prompt format validation
7. **Documentation**: Document available prompts and usage

## Benefits

1. **Better Organization**: Subcommands in separate files
2. **Easier Maintenance**: Self-contained command definitions
3. **Future-Ready**: --prompt flag stub ready for implementation
4. **Consistency**: All commands follow same pattern
5. **Reduced Duplication**: Flags defined once per command
6. **Testability**: Each subcommand can be tested independently

## Compatibility

All changes are backward compatible:
- No breaking changes to command syntax
- All existing flags preserved
- New --prompt flag is optional
- Tests continue to pass
