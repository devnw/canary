// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-CLI-104; FEATURE="CanaryCLI"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/cli"
	"go.devnw.com/canary/cli/db"
	canaryinit "go.devnw.com/canary/cli/init"
	"go.devnw.com/canary/cli/legacy"
	"go.devnw.com/canary/cli/next"
	"go.devnw.com/canary/internal/storage"
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

func init() {
	// Configure slog to use ERROR level by default to reduce noise
	opts := &slog.HandlerOptions{
		Level: slog.LevelError,
	}
	handler := slog.NewTextHandler(os.Stderr, opts)
	slog.SetDefault(slog.New(handler))

	// Add all commands using the centralized Commands() function
	// This automatically includes all subcommands which are registered
	// in their respective package init() functions
	rootCmd.AddCommand(cli.Commands()...)

	// initCmd flags
	canaryinit.InitCmd.Flags().Bool("local", false, "install commands locally in project directory (default: global in home directory)")
	canaryinit.InitCmd.Flags().StringSlice("agents", []string{}, "comma-separated list of agents to install for (claude,cursor,copilot,windsurf,kilocode,roo,opencode,codex,auggie,codebuddy,amazonq)")
	canaryinit.InitCmd.Flags().Bool("all-agents", false, "install commands for all supported agents")
	canaryinit.InitCmd.Flags().String("key", "", "project requirement ID prefix (e.g., CBIN, PROJ, ACME)")
	canaryinit.InitCmd.Flags().String("agent-prefix", "", "agent name prefix for CANARY agents (default: project key)")
	canaryinit.InitCmd.Flags().String("agent-model", "sonnet", "AI model for CANARY agents")
	canaryinit.InitCmd.Flags().String("agent-color", "blue", "color for CANARY agents")

	// migrateCmd flags
	db.MigrateCmd.Flags().String("db", ".canary/canary.db", "path to database file")

	// rollbackCmd flags
	db.RollbackCmd.Flags().String("db", ".canary/canary.db", "path to database file")

	// migrateFromCmd flags
	legacy.MigrateFromCmd.Flags().Bool("dry-run", false, "preview changes without applying them")
	legacy.MigrateFromCmd.Flags().Bool("force", false, "force migration even if system type doesn't match detection")

	// nextCmd flags
	next.NextCmd.Flags().String("db", ".canary/canary.db", "path to database file")
	next.NextCmd.Flags().Bool("prompt", false, "generate full implementation prompt (default: summary only)")
	next.NextCmd.Flags().Bool("json", false, "output in JSON format")
	next.NextCmd.Flags().Bool("dry-run", false, "show what would be selected without generating prompt")
	next.NextCmd.Flags().String("status", "", "filter by status (STUB, IMPL, TESTED, BENCHED)")
	next.NextCmd.Flags().String("aspect", "", "filter by aspect (API, CLI, Engine, Storage, etc.)")
}
