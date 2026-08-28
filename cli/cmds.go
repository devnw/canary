// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package cli

import (
	"github.com/spf13/cobra"

	"go.devnw.com/canary/pkg/cmds/bug"
	"go.devnw.com/canary/pkg/cmds/checkpoint"
	"go.devnw.com/canary/pkg/cmds/constitution"
	"go.devnw.com/canary/pkg/cmds/create"
	"go.devnw.com/canary/pkg/cmds/db"
	"go.devnw.com/canary/pkg/cmds/deps"
	"go.devnw.com/canary/pkg/cmds/doc"
	"go.devnw.com/canary/pkg/cmds/files"
	"go.devnw.com/canary/pkg/cmds/gap"
	"go.devnw.com/canary/pkg/cmds/grep"
	"go.devnw.com/canary/pkg/cmds/implement"
	"go.devnw.com/canary/pkg/cmds/index"
	canaryinit "go.devnw.com/canary/pkg/cmds/init"
	"go.devnw.com/canary/pkg/cmds/legacy"
	"go.devnw.com/canary/pkg/cmds/list"
	"go.devnw.com/canary/pkg/cmds/migrate"
	"go.devnw.com/canary/pkg/cmds/next"
	"go.devnw.com/canary/pkg/cmds/plan"
	"go.devnw.com/canary/pkg/cmds/prioritize"
	"go.devnw.com/canary/pkg/cmds/project"
	"go.devnw.com/canary/pkg/cmds/scan"
	"go.devnw.com/canary/pkg/cmds/search"
	"go.devnw.com/canary/pkg/cmds/show"
	"go.devnw.com/canary/pkg/cmds/specify"
	"go.devnw.com/canary/pkg/cmds/specs"
	"go.devnw.com/canary/pkg/cmds/status"
	"go.devnw.com/canary/pkg/cmds/view"
	"go.devnw.com/canary/mcp"
)

// Commands returns all top-level commands for the canary CLI.
// Subcommands are already registered with their parent commands via init() functions.
//
// Usage:
//
//	rootCmd.AddCommand(cli.Commands()...)
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
		view.CreateViewCommand(),

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
		gap.GapCmd, // Subcommands: mark, query, report, helpful, unhelpful, config, categories
		specs.SpecsCmd,
		bug.BugCmd, // Subcommands: list, create, update, show

		// Project management
		project.DbCmd,
		project.ProjectCmd,

		// MCP server for AI assistant integration
		mcp.New(),
	}
}

func Tools() []*cobra.Command {
	return []*cobra.Command{
		// Add tool commands here
	}
}
