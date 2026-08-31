// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package cli

import (
	"github.com/spf13/cobra"

	"devnw.dev/canary/mcp"
	"devnw.dev/canary/pkg/cmds/bug"
	"devnw.dev/canary/pkg/cmds/checkpoint"
	"devnw.dev/canary/pkg/cmds/constitution"
	"devnw.dev/canary/pkg/cmds/create"
	"devnw.dev/canary/pkg/cmds/db"
	"devnw.dev/canary/pkg/cmds/deps"
	"devnw.dev/canary/pkg/cmds/doc"
	"devnw.dev/canary/pkg/cmds/drift"
	evidencecmd "devnw.dev/canary/pkg/cmds/evidence"
	"devnw.dev/canary/pkg/cmds/files"
	"devnw.dev/canary/pkg/cmds/gap"
	"devnw.dev/canary/pkg/cmds/grep"
	"devnw.dev/canary/pkg/cmds/implement"
	"devnw.dev/canary/pkg/cmds/index"
	canaryinit "devnw.dev/canary/pkg/cmds/init"
	"devnw.dev/canary/pkg/cmds/legacy"
	"devnw.dev/canary/pkg/cmds/list"
	"devnw.dev/canary/pkg/cmds/migrate"
	"devnw.dev/canary/pkg/cmds/next"
	"devnw.dev/canary/pkg/cmds/onboard"
	"devnw.dev/canary/pkg/cmds/plan"
	"devnw.dev/canary/pkg/cmds/prioritize"
	"devnw.dev/canary/pkg/cmds/project"
	"devnw.dev/canary/pkg/cmds/scan"
	"devnw.dev/canary/pkg/cmds/search"
	"devnw.dev/canary/pkg/cmds/show"
	"devnw.dev/canary/pkg/cmds/specify"
	"devnw.dev/canary/pkg/cmds/specs"
	"devnw.dev/canary/pkg/cmds/status"
	"devnw.dev/canary/pkg/cmds/ticket"
	"devnw.dev/canary/pkg/cmds/upgrade"
	"devnw.dev/canary/pkg/cmds/verify"
	"devnw.dev/canary/pkg/cmds/view"
)

// CANARY: REQ=ENG-4326; FEATURE="CommandRegistry"; ASPECT=CLI; STATUS=IMPL; UPDATED=2026-08-29
// Commands returns all top-level commands for the canary CLI.
// Subcommands are already registered with their parent commands via init() functions.
//
// Usage:
//
//	rootCmd.AddCommand(cli.Commands(version)...)
//
// Note: Commands with subcommands (bug, gap, deps, project, db, doc, legacy, migrate)
// have their subcommands registered in their respective package init() functions.
// version is the binary's version (from ldflags). It reaches the MCP server,
// which reports it to clients as its implementation version -- it used to
// report a hardcoded "1.0.0", so every build in the field was indistinguishable
// from every other.
func Commands(version string) []*cobra.Command {
	return []*cobra.Command{
		// Core workflow commands
		scan.ScanCmd,
		verify.VerifyCmd,
		evidencecmd.EvidenceCmd,
		canaryinit.InitCmd,
		create.CreateCmd,
		constitution.ConstitutionCmd,
		specify.SpecifyCmd,
		plan.PlanCmd,
		implement.ImplementCmd,
		next.NextCmd,

		// Database and indexing
		index.IndexCmd,

		// Query and display commands
		list.ListCmd,
		search.SearchCmd,
		show.ShowCmd,
		files.FilesCmd,
		status.StatusCmd,
		grep.GrepCmd,
		view.CreateViewCommand(),
		onboard.CreateOnboardCommand(),
		drift.CreateDriftCommand(),
		ticket.CreateTicketCommand(),

		// Management commands
		prioritize.PrioritizeCmd,
		checkpoint.CheckpointCmd,

		// Database migration commands
		db.MigrateCmd,
		db.RollbackCmd,

		// Legacy migration commands
		legacy.DetectCmd,
		legacy.MigrateFromCmd,
		migrate.OrphanCmd,
		upgrade.UpgradeCmd,

		// Documentation and dependencies
		doc.DocCmd,
		deps.CreateDepsCommand(),

		// Advanced features with subcommands
		gap.GapCmd, // Subcommands: mark, query, report, helpful, unhelpful, config, categories
		specs.SpecsCmd,
		bug.BugCmd, // Subcommands: list, create, update, show

		// Project management
		project.DbCmd,
		project.ProjectCmd,

		// MCP server for AI assistant integration
		mcp.New(version),
	}
}

// init wires subcommands that live in one package under a parent command
// defined in another. `canary db schema` is defined in pkg/cmds/db but the
// `db` parent command lives in pkg/cmds/project; wiring it here (once, at
// package load) keeps that cross-package edge in the assembly layer rather
// than forcing an import between the two leaf packages.
func init() {
	project.DbCmd.AddCommand(db.SchemaCmd)
}
