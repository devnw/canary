// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package db

import (
	"fmt"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/storage"
)

// CANARY: REQ=ENG-4315; FEATURE="DocDatabaseSchema"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2026-08-30
var SchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Print the database schema (embedded forward migrations)",
	Long: `Print the canary database schema as the concatenated forward (.up.sql)
migrations, in order, each with a header naming its migration file.

This is the same content committed to docs/DB_SCHEMA.md, so the documented
schema can be regenerated from, and checked against, the running binary:

  canary db schema > docs/DB_SCHEMA.md`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ddl, err := storage.SchemaDDL()
		if err != nil {
			return fmt.Errorf("read schema: %w", err)
		}
		fmt.Print(ddl)
		return nil
	},
}
