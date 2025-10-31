// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-146; FEATURE="ProjectCLI"; ASPECT=CLI; STATUS=IMPL; UPDATED=2025-10-18
package project

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/storage"
)

// DbCmd manages database initialization
var DbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
	Long:  `Manage the CANARY database (global or local modes).`,
}

// dbInitCmd initializes a database
var dbInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize database",
	Long: `Initialize the CANARY database in global or local mode.

Modes:
  Global (default): ~/.canary/canary.db - shared across all projects
  Local (--local):  ./.canary/canary.db - project-specific database

Examples:
  canary db init           # Initialize global database
  canary db init --global  # Initialize global database (explicit)
  canary db init --local   # Initialize local database`,
	RunE: func(cmd *cobra.Command, args []string) error {
		local, _ := cmd.Flags().GetBool("local")
		global, _ := cmd.Flags().GetBool("global")

		// Determine mode (default to global)
		mode := storage.GlobalMode
		modeStr := "global"

		if local {
			mode = storage.LocalMode
			modeStr = "local"
		} else if global {
			mode = storage.GlobalMode
			modeStr = "global"
		}

		// Initialize database
		manager := storage.NewDatabaseManager()
		if err := manager.Initialize(mode); err != nil {
			return fmt.Errorf("initialize database: %w", err)
		}
		defer manager.Close()

		// Success message
		fmt.Fprintf(cmd.OutOrStdout(), "✅ Initialized %s database at: %s\n", modeStr, manager.Location())

		return nil
	},
}

// ProjectCmd manages projects
var ProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Project management commands",
	Long:  `Register, list, remove, and switch between projects.`,
}

// projectRegisterCmd registers a new project
var projectRegisterCmd = &cobra.Command{
	Use:   "register <name> <path>",
	Short: "Register a new project",
	Long: `Register a new project in the global database.

The project will be assigned a unique slug-based ID generated from the name.

Examples:
  canary project register "My Project" /path/to/project
  canary project register Backend ./backend`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		path := args[1]

		// Open global database
		manager := storage.NewDatabaseManager()
		if err := manager.Discover(); err != nil {
			return fmt.Errorf("database not found: %w (run 'canary db init' first)", err)
		}
		defer manager.Close()

		// Register project
		registry := storage.NewProjectRegistry(manager)
		project := &storage.Project{
			Name: name,
			Path: path,
		}

		if err := registry.Register(project); err != nil {
			return fmt.Errorf("register project: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "✅ Registered project: %s (ID: %s)\n", project.Name, project.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "   Path: %s\n", project.Path)

		return nil
	},
}

// projectListCmd lists all registered projects
var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered projects",
	Long:  `List all projects registered in the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Open database
		manager := storage.NewDatabaseManager()
		if err := manager.Discover(); err != nil {
			return fmt.Errorf("database not found: %w (run 'canary db init' first)", err)
		}
		defer manager.Close()

		// List projects
		registry := storage.NewProjectRegistry(manager)
		projects, err := registry.List()
		if err != nil {
			return fmt.Errorf("list projects: %w", err)
		}

		if len(projects) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No projects registered.")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Registered Projects (%d):\n\n", len(projects))
		for _, p := range projects {
			active := ""
			if p.Active {
				active = " (active)"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  %s: %s%s\n", p.ID, p.Name, active)
			fmt.Fprintf(cmd.OutOrStdout(), "     Path: %s\n", p.Path)
			fmt.Fprintln(cmd.OutOrStdout())
		}

		return nil
	},
}

// projectRemoveCmd removes a project
var projectRemoveCmd = &cobra.Command{
	Use:   "remove <project-id>",
	Short: "Remove a project",
	Long: `Remove a project from the registry.

Examples:
  canary project remove my-project
  canary project remove backend-api`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID := args[0]

		// Open database
		manager := storage.NewDatabaseManager()
		if err := manager.Discover(); err != nil {
			return fmt.Errorf("database not found: %w", err)
		}
		defer manager.Close()

		// Remove project
		registry := storage.NewProjectRegistry(manager)
		if err := registry.Remove(projectID); err != nil {
			return fmt.Errorf("remove project: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "✅ Project removed: %s\n", projectID)

		return nil
	},
}

// projectSwitchCmd switches active project context
var projectSwitchCmd = &cobra.Command{
	Use:   "switch <project-id>",
	Short: "Switch to a different project",
	Long: `Switch the active project context.

Only one project can be active at a time.

Examples:
  canary project switch my-project
  canary project switch backend-api`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID := args[0]

		// Open database
		manager := storage.NewDatabaseManager()
		if err := manager.Discover(); err != nil {
			return fmt.Errorf("database not found: %w", err)
		}
		defer manager.Close()

		// Switch context
		ctx := storage.NewContextManager(manager)
		if err := ctx.SwitchTo(projectID); err != nil {
			return fmt.Errorf("switch project: %w", err)
		}

		// Get project details
		registry := storage.NewProjectRegistry(manager)
		project, err := registry.GetByID(projectID)
		if err != nil {
			return fmt.Errorf("get project: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "✅ Switched to project: %s\n", project.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "   Path: %s\n", project.Path)

		return nil
	},
}

// projectCurrentCmd shows the current active project
var projectCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current active project",
	Long:  `Display the currently active project context.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Open database
		manager := storage.NewDatabaseManager()
		if err := manager.Discover(); err != nil {
			return fmt.Errorf("database not found: %w", err)
		}
		defer manager.Close()

		// Get current project
		ctx := storage.NewContextManager(manager)
		current, err := ctx.GetCurrent()
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), "No active project context set.")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Current Project: %s\n", current.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "  ID: %s\n", current.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "  Path: %s\n", current.Path)

		return nil
	},
}

func init() {
	// Database command flags
	dbInitCmd.Flags().Bool("global", false, "Initialize global database (default)")
	dbInitCmd.Flags().Bool("local", false, "Initialize local database in current directory")

	// Add subcommands
	DbCmd.AddCommand(dbInitCmd)

	ProjectCmd.AddCommand(projectRegisterCmd)
	ProjectCmd.AddCommand(projectListCmd)
	ProjectCmd.AddCommand(projectRemoveCmd)
	ProjectCmd.AddCommand(projectSwitchCmd)
	ProjectCmd.AddCommand(projectCurrentCmd)
}

// Constructor functions for testing
func newDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database management commands",
	}

	init := &cobra.Command{
		Use:   "init",
		Short: "Initialize database",
		RunE:  dbInitCmd.RunE,
	}
	init.Flags().Bool("global", false, "Initialize global database (default)")
	init.Flags().Bool("local", false, "Initialize local database")

	cmd.AddCommand(init)
	return cmd
}

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Project management commands",
	}

	register := &cobra.Command{
		Use:   "register <name> <path>",
		Short: "Register a new project",
		Args:  cobra.ExactArgs(2),
		RunE:  projectRegisterCmd.RunE,
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "List all registered projects",
		RunE:  projectListCmd.RunE,
	}

	remove := &cobra.Command{
		Use:   "remove <project-id>",
		Short: "Remove a project",
		Args:  cobra.ExactArgs(1),
		RunE:  projectRemoveCmd.RunE,
	}

	switchCmd := &cobra.Command{
		Use:   "switch <project-id>",
		Short: "Switch to a different project",
		Args:  cobra.ExactArgs(1),
		RunE:  projectSwitchCmd.RunE,
	}

	current := &cobra.Command{
		Use:   "current",
		Short: "Show current active project",
		RunE:  projectCurrentCmd.RunE,
	}

	cmd.AddCommand(register, list, remove, switchCmd, current)
	return cmd
}
