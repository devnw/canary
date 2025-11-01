// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package cli

import (
	"github.com/spf13/cobra"
	"go.devnw.com/canary/cli/bug"
	"go.devnw.com/canary/cli/checkpoint"
	"go.devnw.com/canary/cli/constitution"
	"go.devnw.com/canary/cli/create"
	"go.devnw.com/canary/cli/db"
	"go.devnw.com/canary/cli/deps"
	"go.devnw.com/canary/cli/doc"
	"go.devnw.com/canary/cli/files"
	"go.devnw.com/canary/cli/gap"
	"go.devnw.com/canary/cli/grep"
	"go.devnw.com/canary/cli/implement"
	"go.devnw.com/canary/cli/index"
	canaryinit "go.devnw.com/canary/cli/init"
	"go.devnw.com/canary/cli/legacy"
	"go.devnw.com/canary/cli/list"
	"go.devnw.com/canary/cli/migrate"
	"go.devnw.com/canary/cli/next"
	"go.devnw.com/canary/cli/plan"
	"go.devnw.com/canary/cli/prioritize"
	"go.devnw.com/canary/cli/project"
	"go.devnw.com/canary/cli/scan"
	"go.devnw.com/canary/cli/search"
	"go.devnw.com/canary/cli/show"
	"go.devnw.com/canary/cli/specify"
	"go.devnw.com/canary/cli/specs"
	"go.devnw.com/canary/cli/status"
)

// Commands returns all top-level commands for the canary CLI.
// Subcommands are already registered with their parent commands via init() functions.
//
// Usage:
//   rootCmd.AddCommand(cli.Commands()...)
//
// Note: Commands with subcommands (bug, gap, deps, project, db, doc, legacy, migrate)
// have their subcommands registered in their respective package init() functions.
func Commands() []*cobra.Command {
	return []*cobra.Command{
		// Core workflow commands
		scan.ScanCmd,
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

		// Documentation and dependencies
		doc.DocCmd,
		deps.CreateDepsCommand(),

		// Advanced features with subcommands
		gap.GapCmd,  // Subcommands: mark, query, report, helpful, unhelpful, config, categories
		specs.SpecsCmd,
		bug.BugCmd,  // Subcommands: list, create, update, show

		// Project management
		project.DbCmd,
		project.ProjectCmd,
	}
}
