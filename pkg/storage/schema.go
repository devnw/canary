// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=ENG-4315; FEATURE="DocDatabaseSchema"; ASPECT=Storage; STATUS=IMPL; OWNER=canary; UPDATED=2026-08-30
package storage

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// SchemaDDL returns the concatenated forward (`.up.sql`) migrations, in
// migration order, each preceded by a header naming its file. It is the single
// source of truth behind both `canary db schema` and docs/DB_SCHEMA.md, so the
// documented schema can never drift from the embedded migrations that actually
// build the database.
func SchemaDDL() (string, error) {
	entries, err := fs.ReadDir(migrationFiles, DBMigrationPath)
	if err != nil {
		return "", fmt.Errorf("read migrations: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".up.sql") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	var b strings.Builder
	for i, name := range names {
		body, rerr := fs.ReadFile(migrationFiles, DBMigrationPath+"/"+name)
		if rerr != nil {
			return "", fmt.Errorf("read migration %s: %w", name, rerr)
		}
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("-- ===== " + name + " =====\n")
		s := strings.TrimRight(string(body), "\n")
		b.WriteString(s)
		b.WriteString("\n")
	}
	return b.String(), nil
}
