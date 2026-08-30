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
	"devnw.dev/canary/pkg/contract"
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
//
// It reads one flag off cmd and cannot fail, so it returns no error: an
// error return every caller was obliged to check and no caller could ever
// see is noise that hides the checks that matter. It takes no root either --
// unlike WriteProjectID it never consults project.yaml.
func ReadProjectID(cmd *cobra.Command) string {
	if cmd != nil {
		if f := cmd.Flags().Lookup(ProjectFlag); f != nil {
			if v, err := cmd.Flags().GetString(ProjectFlag); err == nil && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
	}
	return ""
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
		// A stale schema is the same shape of problem as a missing index --
		// one sentence naming the one fix -- so it is passed through
		// verbatim rather than buried under "open database:".
		if errors.Is(err, storage.ErrSchemaOutOfDate) {
			return nil, err
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
const ExitCodeContract = contract.ExitCode

// ErrContractFailed is the sentinel a command returns once it has already
// printed its contract line. main() maps it to ExitCodeContract; nothing
// prints it, because the JSON line on stdout is the entire message.
var ErrContractFailed = contract.ErrFailed

// FailContract prints the one-line JSON contract on stdout and returns
// ErrContractFailed. Nothing else may reach stdout on that path: the caller
// is a program parsing the line, and a stray banner would corrupt it.
//
// It returns rather than calling os.Exit because it runs inside RunE, where
// exiting skips every deferred database Close and temp-file cleanup between
// here and main. cmd is silenced so cobra adds neither the error nor the
// command's usage block to the output.
func FailContract(cmd *cobra.Command, code, message string) error {
	if cmd != nil {
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
	}
	body, err := json.Marshal(contractResponse{OK: false, Code: code, Message: message})
	if err != nil {
		// Marshalling three strings cannot fail; if it somehow does, say so
		// on stderr rather than emitting a broken contract on stdout.
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return ErrContractFailed
	}
	fmt.Fprintf(os.Stdout, "%s\n", body)
	return ErrContractFailed
}

// GuardContract turns a storage contract error into its machine-readable
// refusal on stdout (ErrContractFailed, exit 2) and passes every other error
// through unchanged. It exists so the two contracts are emitted identically
// wherever they can arise, rather than being re-spelled at each call site.
func GuardContract(cmd *cobra.Command, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, storage.ErrProjectRequired):
		return FailProjectRequired(cmd)
	case errors.Is(err, storage.ErrInvalidOrderBy):
		return FailInvalidOrderBy(cmd)
	}
	return err
}

// FailInvalidOrderBy emits the INVALID_ORDER_BY contract.
func FailInvalidOrderBy(cmd *cobra.Command) error {
	return FailContract(cmd, "INVALID_ORDER_BY", "allowed values: "+strings.Join(storage.OrderKeys(), ","))
}

// FailProjectRequired emits the PROJECT_REQUIRED contract.
func FailProjectRequired(cmd *cobra.Command) error {
	return FailContract(cmd, "PROJECT_REQUIRED", "duplicate requirement id across projects; pass --project")
}
