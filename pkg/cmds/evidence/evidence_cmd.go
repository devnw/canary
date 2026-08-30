// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package evidence implements `canary evidence`: the commands that produce
// and accumulate the passing-test records `canary verify` consumes.
//
//   - `evidence from-go-test` turns a `go test -json` stream into evidence
//     records for the tokens that declare those tests.
//   - `evidence ingest` strictly validates a produced file and merges it into
//     the project's evidence store.
//
// Neither command ever invents a record: from-go-test only emits records for
// tests that actually reported Action="pass", and ingest refuses any file
// that does not satisfy the record grammar in pkg/evidence.
package evidence

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/config"
	ev "devnw.dev/canary/pkg/evidence"
	"devnw.dev/canary/pkg/safewrite"
	"devnw.dev/canary/pkg/sources"
)

// DefaultStorePath is the evidence store's path relative to the project root.
const DefaultStorePath = ".canary/evidence.json"

// maxEventLineBytes bounds one `go test -json` event line. Test output is
// embedded in events, so a test that prints a large blob produces a large
// line; anything past this is not an event worth reading.
const maxEventLineBytes = 4 << 20

// commitPattern is the accepted --commit shape: a full 40-hex SHA. Evidence
// binds to an exact commit; an abbreviated SHA is ambiguous and is refused.
var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

// EvidenceCmd is the `canary evidence` command group.
var EvidenceCmd = &cobra.Command{
	Use:   "evidence",
	Short: "Produce and accumulate passing-test evidence records",
	Long: `Produce and accumulate the evidence records that 'canary verify' consumes.

Subcommands:
  from-go-test   Map a 'go test -json' stream to evidence records (stdout)
  ingest         Validate an evidence file and merge it into the store`,
	// PersistentPreRunE overrides the root command's database auto-creation
	// for this command group (and, by inheritance, its subcommands). Evidence
	// production and ingestion are DB-free, and from-go-test's stdout is the
	// evidence file itself -- nothing else may be written there.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
}

// CANARY: REQ=CP-236; FEATURE="EvidenceIngest"; ASPECT=CLI; STATUS=TESTED; TEST=TestAuditF22,TestIngest_RejectsMalformed,TestIngest_MergesAndDedupes,TestIngest_MalformedStoreIsNeverOverwritten; UPDATED=2026-08-30
// ingestCmd validates an evidence file and merges it into the store.
var ingestCmd = &cobra.Command{
	Use:   "ingest --in <file> [--out <store>]",
	Short: "Validate an evidence file and merge it into the evidence store",
	Long: `Strictly validate an evidence file and merge its records into the store.

Any violation of the record grammar (unknown or duplicate fields, a result
other than PASS, a malformed commit SHA or artifact digest, a non-UTC
timestamp) rejects the whole file: nothing is written and the parse error is
printed to stderr with exit 1. Records identical to ones already in the store
are merged, not duplicated.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		in, _ := cmd.Flags().GetString("in")
		out, _ := cmd.Flags().GetString("out")
		if code := RunIngest(in, out, cmd.ErrOrStderr()); code != 0 {
			os.Exit(code)
		}
		return nil
	},
}

// CANARY: REQ=CP-236; FEATURE="EvidenceFromGoTest"; ASPECT=CLI; STATUS=TESTED; TEST=TestFromGoTest_MapsPassingTests,TestFromGoTest_RejectsShortCommit,TestFromGoTest_ObservedAtFlag,TestPassingTests_SkipsNonEvents; UPDATED=2026-08-30
// fromGoTestCmd maps a `go test -json` stream to evidence records.
var fromGoTestCmd = &cobra.Command{
	Use:   "from-go-test --commit <sha> [flags]",
	Short: "Map a 'go test -json' stream on stdin to evidence records on stdout",
	Long: `Read a 'go test -json' stream on stdin and emit an evidence file on stdout.

Every token that declares TEST=<name> for a test that reported Action="pass"
in the stream yields one record. Tests that did not pass yield nothing, and a
name that reported Action="fail" anywhere in the stream is vetoed everywhere
-- even if it also reported "pass" for a different package -- so this command
cannot manufacture evidence for a test that did not run, did not pass, or
also failed somewhere else in the same run.

The artifact digest is the SHA-256 of the raw input stream, so a record is
tied to the exact test output it was derived from.

Flags:
  --root <dir>            Project root to scan for declarations (default ".")
  --project <id>          Project ID (default: .canary/project.yaml project.key)
  --commit <sha>          Commit the run attests to (required, 40-hex)
  --runner <name>         Runner label recorded on each record (default "local")
  --observed-at <rfc3339> Observation timestamp (default: CANARY_TEST_TIMESTAMP, else now UTC)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := FromGoTestOptions{}
		opts.Root, _ = cmd.Flags().GetString("root")
		opts.ProjectID, _ = cmd.Flags().GetString("project")
		opts.Commit, _ = cmd.Flags().GetString("commit")
		opts.Runner, _ = cmd.Flags().GetString("runner")
		opts.ObservedAt, _ = cmd.Flags().GetString("observed-at")
		if code := RunFromGoTest(opts, cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr()); code != 0 {
			os.Exit(code)
		}
		return nil
	},
}

func init() {
	ingestCmd.Flags().String("in", "", "evidence file to ingest (required)")
	ingestCmd.Flags().String("out", DefaultStorePath, "evidence store to merge into")
	_ = ingestCmd.MarkFlagRequired("in")

	fromGoTestCmd.Flags().String("root", ".", "project root to scan for declarations")
	fromGoTestCmd.Flags().String("project", "", "project ID (default: configured project.key)")
	fromGoTestCmd.Flags().String("commit", "", "commit SHA the run attests to (required)")
	fromGoTestCmd.Flags().String("runner", "local", "runner label recorded on each record")
	fromGoTestCmd.Flags().String("observed-at", "", "observation timestamp (RFC3339 UTC)")
	_ = fromGoTestCmd.MarkFlagRequired("commit")

	EvidenceCmd.AddCommand(ingestCmd, fromGoTestCmd)
}

// RunIngest validates the evidence file at in and merges it into the store at
// out, returning the process exit code (0 accepted, 1 rejected). Nothing is
// written when the input is rejected.
func RunIngest(in, out string, stderr io.Writer) int {
	if in == "" {
		fmt.Fprintln(stderr, "evidence: --in is required")
		return 1
	}
	if out == "" {
		out = DefaultStorePath
	}
	incoming, err := ev.Load(in)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	existing, err := ev.Load(out)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		// An unparseable store is never overwritten: that would destroy the
		// very record trail the store exists to preserve.
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	var current []ev.Record
	if existing != nil {
		current = existing.Records
	}
	merged := Merge(current, incoming.Records)
	if err := WriteStore(out, merged); err != nil {
		fmt.Fprintf(stderr, "evidence: write %s: %v\n", out, err)
		return 1
	}
	return 0
}

// Merge appends the records of incoming that the store does not already hold
// verbatim, preserving the store's existing order. Two records are the same
// record only when every field matches: a differing digest, timestamp or
// runner describes a different observation and is kept.
func Merge(current, incoming []ev.Record) []ev.Record {
	seen := make(map[ev.Record]struct{}, len(current))
	out := make([]ev.Record, 0, len(current)+len(incoming))
	for _, rec := range current {
		if _, dup := seen[rec]; dup {
			continue
		}
		seen[rec] = struct{}{}
		out = append(out, rec)
	}
	for _, rec := range incoming {
		if _, dup := seen[rec]; dup {
			continue
		}
		seen[rec] = struct{}{}
		out = append(out, rec)
	}
	return out
}

// WriteStore writes records to path atomically (staged in the same directory,
// fsynced, then renamed), creating the directory when needed. A crash
// mid-write therefore leaves the previous store intact rather than a
// truncated one. The store is this command's own output, so replacing it is
// intended; only a partial write would be a defect.
func WriteStore(path string, records []ev.Record) error {
	if records == nil {
		records = []ev.Record{}
	}
	data, err := json.MarshalIndent(ev.File{SchemaVersion: 1, Records: records}, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	_, err = safewrite.Write(path, data, 0o600, safewrite.Options{
		Root:  filepath.Dir(path),
		Force: true,
	})
	return err
}

// FromGoTestOptions are the resolved inputs of one from-go-test run.
type FromGoTestOptions struct {
	Root       string
	ProjectID  string
	Commit     string
	Runner     string
	ObservedAt string
}

// goTestEvent is the subset of `go test -json` events this command reads.
type goTestEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
}

// RunFromGoTest reads a `go test -json` stream from stdin and writes an
// evidence file to stdout, returning the process exit code.
func RunFromGoTest(opts FromGoTestOptions, stdin io.Reader, stdout, stderr io.Writer) int {
	if opts.Root == "" {
		opts.Root = "."
	}
	if opts.Runner == "" {
		opts.Runner = "local"
	}
	if !commitPattern.MatchString(opts.Commit) {
		fmt.Fprintf(stderr, "evidence: --commit %q: want 40 lowercase hex\n", opts.Commit)
		return 1
	}
	observedAt, err := resolveObservedAt(opts.ObservedAt)
	if err != nil {
		fmt.Fprintf(stderr, "evidence: --observed-at: %v\n", err)
		return 1
	}

	raw, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(stderr, "evidence: read stdin: %v\n", err)
		return 1
	}
	sum := sha256.Sum256(raw)
	digest := "sha256:" + hex.EncodeToString(sum[:])

	passed := PassingTests(raw)

	projCfg, err := config.Load(opts.Root)
	if err != nil {
		fmt.Fprintf(stderr, "evidence: config: %v\n", err)
		return 1
	}
	reg, err := sources.FromProjectConfig(projCfg)
	if err != nil {
		fmt.Fprintf(stderr, "evidence: sources: %v\n", err)
		return 1
	}
	projectID := opts.ProjectID
	if projectID == "" {
		projectID = projCfg.ProjectID()
	}
	ignorePatterns, err := canaryscan.LoadCanaryIgnore(opts.Root)
	if err != nil {
		fmt.Fprintf(stderr, "evidence: .canaryignore: %v\n", err)
		return 1
	}
	rep, err := canaryscan.Scan(opts.Root, canaryscan.StateSkipRegex(), nil, ignorePatterns, reg)
	if err != nil {
		fmt.Fprintf(stderr, "evidence: scan: %v\n", err)
		return 1
	}
	for _, is := range rep.Issues {
		fmt.Fprintf(stderr, "CANARY_SCAN_ISSUE path=%s reason=%s\n", is.Path, is.Reason)
	}

	records := Records(rep, passed, projectID, opts.Commit, opts.Runner, observedAt, digest)
	data, err := json.MarshalIndent(ev.File{SchemaVersion: 1, Records: records}, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "evidence: encode: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "%s\n", data)
	return 0
}

// PassingTests returns the test names that reported Action="pass" in a
// `go test -json` stream, mapped to the package they ran in — but only for
// names that never also reported Action="fail" anywhere in the stream. A
// test name that passed in one package and failed in another (e.g. the same
// subtest name reused across packages) is vetoed entirely: it is a lie to
// call that name "passing" when the stream also shows it failing, so no
// record is emitted for it at all, regardless of how many packages it passed
// in.
//
// The stream is read line by line and any line that is not a JSON event is
// skipped: `go test -json` can interleave raw build/vet output with events,
// and a build log line must not cost the whole run its evidence. Skipping is
// safe in the only direction that matters — an unreadable line can only
// produce FEWER records, never a record for a test that did not pass.
func PassingTests(raw []byte) map[string]string {
	passed := map[string]string{}
	failed := map[string]bool{}
	sc := bufio.NewScanner(bytes.NewReader(raw))
	sc.Buffer(make([]byte, 0, 64<<10), maxEventLineBytes)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 || line[0] != '{' {
			continue
		}
		var e goTestEvent
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		if e.Test == "" {
			continue
		}
		switch e.Action {
		case "pass":
			passed[e.Test] = e.Package
		case "fail":
			failed[e.Test] = true
		}
	}
	// A fail anywhere in the stream vetoes the name everywhere: passing in
	// package A does not excuse failing in package B.
	for name := range failed {
		delete(passed, name)
	}
	return passed
}

// Records builds one evidence record per (requirement, feature, aspect, test)
// whose test is in passed. Records are sorted by requirement, feature, aspect
// then test so repeated runs over the same inputs produce byte-identical
// output.
func Records(rep canaryscan.Report, passed map[string]string, projectID, commit, runner, observedAt, digest string) []ev.Record {
	var records []ev.Record
	for _, r := range rep.Requirements {
		for _, f := range r.Features {
			for _, test := range f.Tests {
				pkg, ok := passed[test]
				if !ok {
					continue
				}
				records = append(records, ev.Record{
					ProjectID:      projectID,
					RequirementID:  r.ID,
					Feature:        f.Feature,
					Aspect:         f.Aspect,
					TestID:         test,
					Command:        testCommand(test, pkg),
					Result:         "PASS",
					CommitSHA:      commit,
					ObservedAt:     observedAt,
					Runner:         runner,
					ArtifactDigest: digest,
				})
			}
		}
	}
	sort.Slice(records, func(i, j int) bool {
		a, b := records[i], records[j]
		if a.RequirementID != b.RequirementID {
			return a.RequirementID < b.RequirementID
		}
		if a.Feature != b.Feature {
			return a.Feature < b.Feature
		}
		if a.Aspect != b.Aspect {
			return a.Aspect < b.Aspect
		}
		return a.TestID < b.TestID
	})
	return records
}

// testCommand renders the command that reproduces one test's result.
func testCommand(test, pkg string) string {
	if pkg == "" {
		pkg = "./..."
	}
	return fmt.Sprintf("go test -run ^%s$ %s", test, pkg)
}

// resolveObservedAt returns the observation timestamp: the --observed-at flag
// when given (validated as RFC3339 UTC), else CANARY_TEST_TIMESTAMP, else now
// in UTC. The env fallback is the same test-pinning convention used
// throughout the codebase, and is what makes a from-go-test run reproducible.
func resolveObservedAt(flagValue string) (string, error) {
	if flagValue != "" {
		t, err := time.Parse(time.RFC3339, flagValue)
		if err != nil {
			return "", err
		}
		if _, offset := t.Zone(); offset != 0 {
			return "", fmt.Errorf("want UTC (zero offset), got %q", flagValue)
		}
		return t.UTC().Format(time.RFC3339), nil
	}
	if t := canaryscan.RefTimeFromEnv(); !t.IsZero() {
		return t.UTC().Format(time.RFC3339), nil
	}
	return time.Now().UTC().Format(time.RFC3339), nil
}
