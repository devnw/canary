# CLI Commands() Refactor Summary

## Overview
Refactored the command registration system to use a centralized `cli.Commands()` function that can be called with `rootCmd.AddCommand(cli.Commands()...)`.

## Changes Made

### 1. Created `cli/cmds.go`
New file that exports all top-level commands from a single function:

```go
package cli

func Commands() []*cobra.Command {
    return []*cobra.Command{
        // All commands listed here
    }
}
```

**Key Features:**
- Single source of truth for all commands
- Subcommands automatically registered via `init()` functions
- Well-documented with comments explaining command hierarchy
- Organized into logical groups (workflow, database, query, etc.)

### 2. Simplified `cmd/canary/main.go`

#### Before:
```go
import (
    // 30+ individual command imports
    "go.devnw.com/canary/cli/bug"
    "go.devnw.com/canary/cli/create"
    // ... etc
)

func init() {
    rootCmd.AddCommand(scan.ScanCmd)
    rootCmd.AddCommand(canaryinit.InitCmd)
    rootCmd.AddCommand(create.CreateCmd)
    // ... 30+ more AddCommand calls
}
```

#### After:
```go
import (
    "go.devnw.com/canary/cli"
    "go.devnw.com/canary/cli/db"
    canaryinit "go.devnw.com/canary/cli/init"
    "go.devnw.com/canary/cli/legacy"
    "go.devnw.com/canary/cli/next"
    // Only 5 imports needed for flag configuration
)

func init() {
    // Add all commands with subcommands automatically included
    rootCmd.AddCommand(cli.Commands()...)
    
    // Configure flags for specific commands that need it
    canaryinit.InitCmd.Flags().Bool("local", false, "...")
    // ... etc
}
```

**Benefits:**
- Reduced imports from 30+ to 5
- Reduced code from 40+ lines to 1 line for command registration
- Cleaner, more maintainable structure
- Easy to add new commands (just add to `cli/cmds.go`)

### 3. Command Hierarchy

The system now properly handles command hierarchies:

#### Parent Commands with Subcommands:
- **bug** ? list, create, update, show
- **gap** ? mark, query, report, helpful, unhelpful, config, categories
- **deps** ? (dynamically created subcommands)
- **project** ? DbCmd, ProjectCmd
- **db** ? MigrateCmd, RollbackCmd
- **doc** ? (document management subcommands)
- **legacy** ? DetectCmd, MigrateFromCmd
- **migrate** ? OrphanCmd

All subcommands are registered in their respective package `init()` functions, so they're automatically available when the parent command is added.

## File Structure

```
cli/
??? cmds.go              # NEW: Central command registry
??? bug/
?   ??? bug.go           # Parent + helpers
?   ??? list.go          # Subcommand
?   ??? create.go        # Subcommand
?   ??? update.go        # Subcommand
?   ??? show.go          # Subcommand
??? gap/
?   ??? gap.go           # Parent + all subcommands
??? create/
?   ??? create.go        # Standalone command
??? list/
?   ??? list.go          # Standalone command
??? ... (other commands)

cmd/canary/
??? main.go              # SIMPLIFIED: Uses cli.Commands()
```

## Usage Pattern

### For Adding New Commands:

1. **Create command in its package:**
   ```go
   // cli/mycommand/mycommand.go
   package mycommand
   
   var MyCmd = &cobra.Command{
       Use: "mycommand",
       // ...
   }
   
   func init() {
       MyCmd.Flags().String("prompt", "", "...")
       // ... other flags
   }
   ```

2. **Add to `cli/cmds.go`:**
   ```go
   import "go.devnw.com/canary/cli/mycommand"
   
   func Commands() []*cobra.Command {
       return []*cobra.Command{
           // ... existing commands
           mycommand.MyCmd,
       }
   }
   ```

3. **Done!** No need to modify `cmd/canary/main.go`

### For Adding Subcommands:

1. **Create subcommand file:**
   ```go
   // cli/parent/subcommand.go
   package parent
   
   var subCmd = &cobra.Command{
       Use: "sub",
       // ...
   }
   
   func init() {
       ParentCmd.AddCommand(subCmd)
       subCmd.Flags().String("prompt", "", "...")
   }
   ```

2. **Already works!** Subcommand is automatically included when parent is added

## Testing

All functionality verified:
- ? All commands available: `canary --help`
- ? Bug subcommands work: `canary bug --help`
- ? Gap subcommands work: `canary gap --help`
- ? All --prompt flags present
- ? Binary builds successfully
- ? No duplicate command registrations

## Benefits

1. **Maintainability**
   - Single place to see all commands
   - Easy to add/remove commands
   - Clear command hierarchy

2. **Cleaner Code**
   - Reduced imports in main.go (30+ ? 5)
   - Reduced code lines (40+ ? 1 for registration)
   - Better separation of concerns

3. **Flexibility**
   - Easy to reorganize commands
   - Easy to add command groups
   - Subcommands automatically included

4. **Documentation**
   - Self-documenting structure in `cli/cmds.go`
   - Comments explain command relationships
   - Easy to understand command hierarchy

## Migration Notes

### What Changed:
- `cmd/canary/main.go`: Uses `cli.Commands()` instead of individual `AddCommand` calls
- `cli/cmds.go`: NEW file with centralized command registry
- Imports: Reduced to only what's needed for flag configuration

### What Stayed the Same:
- All command functionality
- All flag definitions
- All subcommand relationships
- All tests and behavior

### Backward Compatibility:
- ? 100% backward compatible
- No breaking changes
- All commands work identically
- All flags preserved

## Future Enhancements

With this structure, it's easy to add:
- Command groups/categories
- Dynamic command loading
- Plugin system for external commands
- Command aliases
- Command deprecation warnings

## Example: Adding a New Command

```go
// 1. Create cli/export/export.go
package export

import "github.com/spf13/cobra"

var ExportCmd = &cobra.Command{
    Use:   "export",
    Short: "Export CANARY tokens",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Implementation
        return nil
    },
}

func init() {
    ExportCmd.Flags().String("prompt", "", "Custom prompt (future use)")
    ExportCmd.Flags().String("format", "json", "Export format")
}

// 2. Update cli/cmds.go
import "go.devnw.com/canary/cli/export"

func Commands() []*cobra.Command {
    return []*cobra.Command{
        // ... existing commands
        export.ExportCmd,  // Add new command here
    }
}

// 3. Done! Command is now available
```

## Conclusion

The refactored command system provides a clean, maintainable structure that:
- Centralizes command registration
- Simplifies the main entry point
- Properly handles command hierarchies
- Makes it easy to add new commands
- Maintains full backward compatibility
