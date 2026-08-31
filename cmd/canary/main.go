// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-CLI-104; FEATURE="CanaryCLI"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"

	"devnw.dev/canary/cli"
	"devnw.dev/canary/pkg/cmds/db"
	canaryinit "devnw.dev/canary/pkg/cmds/init"
	"devnw.dev/canary/pkg/cmds/legacy"
	"devnw.dev/canary/pkg/cmds/next"
	"devnw.dev/canary/pkg/contract"
)

// version, commit and date are overwritten at release time by goreleaser via
// -X main.version / -X main.commit / -X main.date ldflags. They MUST exist as
// package-level vars for those flags to land -- a missing var makes the linker
// silently drop the -X. For a plain `go install ...@vX` build (no ldflags),
// resolveVersion falls back to the module version stamped into the binary by
// the toolchain.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"

	rootCmd = &cobra.Command{
		Use:   "canary",
		Short: "Track requirements via CANARY tokens in source code",
		Long: `Canary tracks requirements through CANARY tokens embedded in source code.

Inspired by spec-kit's specification-driven development, canary provides
commands for scanning, creating, and managing requirement tokens.`,
		Version: resolveVersion(),

		// main() prints the error itself; without this cobra prints it too,
		// so every failure arrived twice.
		SilenceErrors: true,

		// There is deliberately no PersistentPreRunE here. The root command
		// used to run storage.AutoMigrate before *every* subcommand, so a
		// pure read -- `canary list` on a repository with no index, `canary
		// scan`, even `canary --help` -- created .canary/canary.db and
		// printed a creation banner on stdout. Each command now opens the
		// database itself: storage.OpenRO to read (never creates anything),
		// storage.OpenRW to write (creates and migrates, banners on stderr).
	}
)

// resolveVersion returns the version to report. A release build has version
// set by ldflags. A `go install devnw.dev/canary/cmd/canary@vX.Y.Z` build
// leaves it "dev", but the toolchain stamps the module version into the
// binary's build info, so fall back to that (tag-derived) value when present.
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return version
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// A refused contract has already said everything it has to say: one
		// JSON line on stdout, nothing on stderr. Exiting here rather than
		// inside RunE is what lets every deferred database Close and
		// temp-file cleanup along the way actually run.
		if errors.Is(err, contract.ErrFailed) {
			os.Exit(contract.ExitCode)
		}
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

	// Resolve the effective version once (ldflags at release, module build
	// info for `go install`, else "dev") and thread it everywhere: the
	// --version output, and every subcommand that reports a version (mcp).
	resolved := resolveVersion()
	rootCmd.Version = resolved
	rootCmd.SetVersionTemplate(fmt.Sprintf("canary %s (%s %s)\n", resolved, commit, date))

	// Add all commands using the centralized Commands() function
	// This automatically includes all subcommands which are registered
	// in their respective package init() functions
	rootCmd.AddCommand(cli.Commands(resolved)...)

	// initCmd flags
	canaryinit.InitCmd.Flags().Bool("local", false, "install commands locally in project directory (default: global in home directory)")
	canaryinit.InitCmd.Flags().StringSlice("agents", []string{}, "comma-separated list of agents to install for (claude,cursor,copilot,windsurf,kilocode,roo,opencode,codex,auggie,codebuddy,amazonq)")
	canaryinit.InitCmd.Flags().Bool("all-agents", false, "install commands for all supported agents")
	canaryinit.InitCmd.Flags().String("key", "", "project requirement ID prefix (e.g., CBIN, PROJ, ACME)")
	canaryinit.InitCmd.Flags().String("agent-prefix", "", "agent name prefix for CANARY agents (default: project key)")
	canaryinit.InitCmd.Flags().String("agent-model", "sonnet", "AI model for CANARY agents")
	canaryinit.InitCmd.Flags().String("agent-color", "blue", "color for CANARY agents")
	canaryinit.InitCmd.Flags().Bool("force", false, "overwrite files whose content differs (previous content is saved to <file>.bak)")

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
	next.NextCmd.Flags().String("prompt-arg", "", "extra prompt file or embedded prompt name to expose to the prompt template")
	next.NextCmd.Flags().String("format", next.FormatText, "output format: json or text")
	next.NextCmd.Flags().Bool("json", false, "output in JSON format (deprecated: use --format json)")
	// Registering the deprecation with cobra is what makes the alias actually
	// warn when it is used. The warning goes to the command's error output,
	// never to stdout, so `next --json` stays pipeable into a JSON parser.
	// MarkDeprecated only fails when the flag does not exist, which the line
	// above just guaranteed.
	if err := next.NextCmd.Flags().MarkDeprecated("json", "use --format json"); err != nil {
		panic(err)
	}
	next.NextCmd.Flags().Bool("dry-run", false, "show what would be selected without generating prompt")
	next.NextCmd.Flags().String("status", "", "filter by status (STUB, IMPL, TESTED, BENCHED)")
	next.NextCmd.Flags().String("aspect", "", "filter by aspect (API, CLI, Engine, Storage, etc.)")
	next.NextCmd.Flags().Bool("allow-unknown-external", false, "do not block on external (ticket-source/peer) dependencies whose state cannot be resolved from disk")
	next.NextCmd.Flags().String("project", "", "scope the query to one project id (default: unscoped -- required only when the index holds more than one project)")
}
