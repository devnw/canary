// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/config"
	"devnw.dev/canary/pkg/storage"
)

// CANARY: REQ=CBIN-307; FEATURE="ProjectScopeCLI"; ASPECT=CLI; STATUS=TESTED; TEST=TestAuditF08CLI,TestAuditF07; UPDATED=2026-08-30

// ProjectFlag is the flag every project-scoped command registers.
const ProjectFlag = "project"

// AddProjectFlag registers --project on cmd.
func AddProjectFlag(cmd *cobra.Command) {
	cmd.Flags().String(ProjectFlag, "", "scope the query to one project id (default: unscoped -- required only when the index holds more than one project)")
}

// ReadProjectID resolves the project a *reading* command scopes to:
//
//	--project <id> -> that id
//	otherwise      -> "" (unscoped)
//
// The unscoped default is deliberate, and it is not the same as guessing a
// name. A database holding one project answers an unscoped query exactly as
// it always did; a database holding several refuses with
// storage.ErrProjectRequired and names --project as the fix. Defaulting the
// scope to the configured project.key instead would be worse than useless:
// it answers no question a single-project database could not already answer,
// and the day someone adds a key to project.yaml every row indexed before
// that change would silently vanish from every read.
func ReadProjectID(cmd *cobra.Command, root string) (string, error) {
	if cmd != nil {
		if f := cmd.Flags().Lookup(ProjectFlag); f != nil {
			if v, err := cmd.Flags().GetString(ProjectFlag); err == nil && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v), nil
			}
		}
	}
	return "", nil
}

// WriteProjectID resolves the project a *writing* command stores rows under.
// A writer must name a project -- a row with no identity is reachable by no
// scoped read and no scoped delete -- so an unconfigured repository resolves
// to config's "default", the same value migration 000007 backfills onto
// pre-scoping rows.
func WriteProjectID(cmd *cobra.Command, root string) (string, error) {
	if cmd != nil {
		if f := cmd.Flags().Lookup(ProjectFlag); f != nil {
			if v, err := cmd.Flags().GetString(ProjectFlag); err == nil && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v), nil
			}
		}
	}
	cfg, err := config.Load(rootOrDot(root))
	if err != nil {
		return "", fmt.Errorf("load .canary/project.yaml: %w", err)
	}
	return cfg.ProjectID(), nil
}

func rootOrDot(root string) string {
	if strings.TrimSpace(root) == "" {
		return "."
	}
	return root
}

// ErrNoIndex is what a read-only command returns when the project has never
// been indexed. It names the fix rather than creating the database itself: a
// read must never have a write side effect.
var ErrNoIndex = errors.New("no index; run 'canary index'")

// OpenIndexRO opens the token index read-only for cmd. A missing index
// yields ErrNoIndex; nothing is created on any path.
//
// cmd may be nil. When it is not, usage output is suppressed for this error:
// "no index; run 'canary index'" is a state problem, and answering it with
// the command's full flag list buries the one line that matters.
func OpenIndexRO(cmd *cobra.Command, dbPath string) (*storage.DB, error) {
	db, err := storage.OpenRO(dbPath)
	if err != nil {
		if cmd != nil {
			cmd.SilenceUsage = true
		}
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrNoIndex
		}
		return nil, fmt.Errorf("open database: %w", err)
	}
	return db, nil
}

// contractResponse is the machine-readable refusal body.
type contractResponse struct {
	OK      bool   `json:"ok"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ExitCodeContract is the exit status a contract refusal uses. It is distinct
// from 1 so a caller can tell "canary refused because you asked for something
// it will not do" from "the command failed".
const ExitCodeContract = 2

// FailContract prints a one-line JSON contract on stdout and exits 2. Nothing
// else may reach stdout on that path: the caller is a program parsing the
// line, and a stray banner would corrupt it.
func FailContract(code, message string) {
	body, err := json.Marshal(contractResponse{OK: false, Code: code, Message: message})
	if err != nil {
		// Marshalling three strings cannot fail; if it somehow does, say so
		// on stderr rather than emitting a broken contract on stdout.
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitCodeContract)
	}
	fmt.Fprintf(os.Stdout, "%s\n", body)
	os.Exit(ExitCodeContract)
}

// GuardContract turns a storage contract error into its machine-readable
// refusal on stdout (exit 2) and passes every other error through unchanged.
// It exists so the two contracts are emitted identically wherever they can
// arise, rather than being re-spelled at each call site.
func GuardContract(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, storage.ErrProjectRequired):
		FailProjectRequired()
	case errors.Is(err, storage.ErrInvalidOrderBy):
		FailInvalidOrderBy()
	}
	return err
}

// FailInvalidOrderBy emits the INVALID_ORDER_BY contract.
func FailInvalidOrderBy() {
	FailContract("INVALID_ORDER_BY", "allowed values: "+strings.Join(storage.OrderKeys(), ","))
}

// FailProjectRequired emits the PROJECT_REQUIRED contract.
func FailProjectRequired() {
	FailContract("PROJECT_REQUIRED", "duplicate requirement id across projects; pass --project")
}
