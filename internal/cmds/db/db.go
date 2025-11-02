package db

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/storage"
)

// CANARY: REQ=CBIN-129; FEATURE="MigrateCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var MigrateCmd = &cobra.Command{
	Use:   "migrate <steps>",
	Short: "Run database migrations",
	Long: `Apply database migrations to the CANARY database.

Steps can be:
  - "all" to migrate to the latest version
  - A positive integer to migrate forward by that many steps
  - A negative integer to roll back by that many steps`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		steps := args[0]

		fmt.Printf("Running migrations on: %s\n", dbPath)

		if err := storage.MigrateDB(dbPath, steps); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		fmt.Println("✅ Migrations completed successfully")
		return nil
	},
}

// CANARY: REQ=CBIN-129; FEATURE="RollbackCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var RollbackCmd = &cobra.Command{
	Use:   "rollback <steps>",
	Short: "Roll back database migrations",
	Long: `Roll back database migrations.

Steps can be:
  - "all" to roll back all migrations
  - A positive integer to roll back by that many steps`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		steps := args[0]

		fmt.Printf("Rolling back migrations on: %s\n", dbPath)

		if err := storage.TeardownDB(dbPath, steps); err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}

		fmt.Println("✅ Rollback completed successfully")
		return nil
	},
}
