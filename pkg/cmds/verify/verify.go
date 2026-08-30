// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package verify implements `canary verify`: the evidence-backed answer to
// "are this project's claims true right now?". It reads declarations from the
// tree, claims from a gap-analysis file, and passing-test evidence from the
// evidence store, then delegates the verdict to pkg/evidence.Complete — the
// single completion function. Nothing here decides completeness on its own.
package verify

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/config"
	"devnw.dev/canary/pkg/evidence"
	"devnw.dev/canary/pkg/external"
	"devnw.dev/canary/pkg/sources"
)

// DefaultEvidencePath is the evidence store's path relative to the project
// root.
const DefaultEvidencePath = ".canary/evidence.json"

// Format values accepted by --format.
const (
	FormatJSON = "json"
	FormatText = "text"
)

// scanIncomplete is the verdict emitted whenever verification state is
// unknowable: the tree could not be fully scanned, the config could not be
// read, the claims file could not be read, the evidence store is malformed,
// or the current commit could not be determined. In every one of those cases
// the honest answer is "unknown", never "verified".
var scanIncomplete = evidence.Verdict{
	OK:      false,
	State:   "UNKNOWN",
	Code:    "SCAN_INCOMPLETE",
	Message: "scan incomplete",
}

// externalUnknown is the verdict emitted when a claimed requirement depends
// on an external/peer requirement whose state cannot be resolved from disk.
var externalUnknown = evidence.Verdict{
	OK:      false,
	State:   "UNKNOWN",
	Code:    "EXTERNAL_UNKNOWN",
	Message: "external dependency state unknown",
}

// Options are the resolved inputs of one verification run.
type Options struct {
	Root         string
	ClaimsPath   string
	EvidencePath string
	Format       string
	// ProjectID overrides the configured project.key when non-empty.
	ProjectID            string
	AllowEmpty           bool
	AllowUnknownExternal bool
}

// CANARY: REQ=CP-236; FEATURE="EvidenceVerify"; ASPECT=CLI; STATUS=TESTED; TEST=TestAuditF01CLI,TestAuditF15CLI,TestAuditF22,TestVerifyRun_EvidenceMissing,TestVerifyRun_Verified,TestVerifyRun_WrongCommitIsNotVerified,TestVerifyRun_TextFormat,TestVerifyRun_ScanIssueIsIncomplete,TestVerifyRun_UndeclaredClaim,TestVerifyRun_MissingClaimsFileIsUnknown,TestVerifyRun_ExternalUnknownBlocks,TestVerifyRun_ProjectOverride; UPDATED=2026-08-30
// VerifyCmd is the `canary verify` command.
var VerifyCmd = &cobra.Command{
	Use:   "verify [flags]",
	Short: "Verify claimed requirements against passing-test evidence",
	Long: `Verify that every requirement claimed in a gap-analysis file has passing
evidence recorded at the current commit.

Claims are lines like "✅ <REQ-ID>" in the claims file. For each claimed
requirement, every feature/aspect it declares in source must have a PASS
evidence record for this project at HEAD. Declarations alone (STATUS=TESTED,
TEST=...) never verify anything.

Exit codes: 0 verified, 1 not verified (or unknowable).

Flags:
  --root <dir>              Project root to scan (default ".")
  --claims <file>           Claims file (default "GAP_ANALYSIS.md")
  --evidence <file>         Evidence store (default ".canary/evidence.json")
  --format json|text        Output format (default "json")
  --project <id>            Override the configured project key
  --allow-empty             Treat a claims file with no claims as verified
  --allow-unknown-external  Do not block on unresolvable external dependencies

Examples:
  canary verify
  canary verify --format text --claims GAP_ANALYSIS.md`,
	// PersistentPreRunE overrides the root command's database auto-creation.
	// Verification is read-only and DB-free: it must work in a project that
	// has never had a canary database, and in JSON mode stdout must carry the
	// verdict line and nothing else -- the root hook prints creation notices
	// there.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := Options{}
		opts.Root, _ = cmd.Flags().GetString("root")
		opts.ClaimsPath, _ = cmd.Flags().GetString("claims")
		opts.EvidencePath, _ = cmd.Flags().GetString("evidence")
		opts.Format, _ = cmd.Flags().GetString("format")
		opts.ProjectID, _ = cmd.Flags().GetString("project")
		opts.AllowEmpty, _ = cmd.Flags().GetBool("allow-empty")
		opts.AllowUnknownExternal, _ = cmd.Flags().GetBool("allow-unknown-external")

		if opts.Format != FormatJSON && opts.Format != FormatText {
			return fmt.Errorf("unknown --format %q (want %q or %q)", opts.Format, FormatJSON, FormatText)
		}
		if code := Run(opts, cmd.OutOrStdout(), cmd.ErrOrStderr()); code != 0 {
			os.Exit(code)
		}
		return nil
	},
}

func init() {
	VerifyCmd.Flags().String("root", ".", "project root to scan")
	VerifyCmd.Flags().String("claims", "GAP_ANALYSIS.md", "claims file to verify")
	VerifyCmd.Flags().String("evidence", DefaultEvidencePath, "evidence store path")
	VerifyCmd.Flags().String("format", FormatJSON, "output format: json or text")
	VerifyCmd.Flags().String("project", "", "override the configured project key")
	VerifyCmd.Flags().Bool("allow-empty", false, "treat a claims file with no claims as verified")
	VerifyCmd.Flags().Bool("allow-unknown-external", false, "do not block on unresolvable external dependencies")
}

// Run performs one verification and returns the process exit code: 0 when
// the verdict is OK, 1 otherwise. Exactly one line is ever written to stdout
// (the verdict); every diagnostic goes to stderr.
func Run(opts Options, stdout, stderr io.Writer) int {
	if opts.Root == "" {
		opts.Root = "."
	}
	if opts.ClaimsPath == "" {
		opts.ClaimsPath = "GAP_ANALYSIS.md"
	}
	if opts.EvidencePath == "" {
		opts.EvidencePath = DefaultEvidencePath
	}
	if opts.Format == "" {
		opts.Format = FormatJSON
	}
	// The evidence store path is resolved the same way pkg/canaryscan does:
	// relative to --root, not the process's CWD. An absolute --evidence
	// passes through unchanged; a relative one (including the default,
	// ".canary/evidence.json") is joined under --root so `canary verify
	// --root <dir>` finds evidence recorded for that project regardless of
	// where the command is invoked from.
	evidencePath := opts.EvidencePath
	if !filepath.IsAbs(evidencePath) {
		evidencePath = filepath.Join(opts.Root, evidencePath)
	}

	projCfg, err := config.Load(opts.Root)
	if err != nil {
		fmt.Fprintf(stderr, "verify: config: %v\n", err)
		return emit(stdout, stderr, opts.Format, scanIncomplete)
	}
	reg, err := sources.FromProjectConfig(projCfg)
	if err != nil {
		fmt.Fprintf(stderr, "verify: sources: %v\n", err)
		return emit(stdout, stderr, opts.Format, scanIncomplete)
	}
	projectID := projCfg.ProjectID()
	if opts.ProjectID != "" {
		projectID = opts.ProjectID
	}

	ignorePatterns, ignoreErr := canaryscan.LoadCanaryIgnore(opts.Root)
	if ignoreErr != nil {
		fmt.Fprintf(stderr, "verify: .canaryignore: %v\n", ignoreErr)
		return emit(stdout, stderr, opts.Format, scanIncomplete)
	}

	rep, err := canaryscan.Scan(opts.Root, canaryscan.StateSkipRegex(), nil, ignorePatterns, reg)
	if err != nil {
		fmt.Fprintf(stderr, "verify: scan: %v\n", err)
		return emit(stdout, stderr, opts.Format, scanIncomplete)
	}
	// A partial scan cannot yield a clean verdict: some part of the tree was
	// never examined, so its declarations are unknown.
	if len(rep.Issues) > 0 {
		for _, is := range rep.Issues {
			fmt.Fprintf(stderr, "CANARY_SCAN_ISSUE path=%s reason=%s\n", is.Path, is.Reason)
		}
		return emit(stdout, stderr, opts.Format, scanIncomplete)
	}

	commit, err := canaryscan.HeadCommit(opts.Root)
	if err != nil {
		fmt.Fprintf(stderr, "verify: cannot determine commit: %v\n", err)
		return emit(stdout, stderr, opts.Format, scanIncomplete)
	}

	claimed, err := canaryscan.Claims(opts.ClaimsPath, reg)
	if err != nil {
		fmt.Fprintf(stderr, "verify: claims: %v\n", err)
		return emit(stdout, stderr, opts.Format, scanIncomplete)
	}

	recs, err := loadEvidence(evidencePath)
	if err != nil {
		fmt.Fprintf(stderr, "verify: evidence: %v\n", err)
		return emit(stdout, stderr, opts.Format, scanIncomplete)
	}

	required := canaryscan.RequiredFeatures(rep, claimed)
	v := evidence.Complete(required, recs, projectID, commit, opts.AllowEmpty)
	for _, m := range v.Missing {
		fmt.Fprintf(stderr, "UNVERIFIED %s %s/%s reason=%s\n",
			m.RequirementID, m.Key.Feature, m.Key.Aspect, m.Reason)
	}

	// External/peer dependency state is checked last: an unresolvable
	// dependency means the verdict itself is unknown, which outranks any
	// evidence-level answer already computed.
	if unknown := unknownExternals(rep, claimed, reg, opts.Root); len(unknown) > 0 {
		for _, u := range unknown {
			fmt.Fprintf(stderr, "EXTERNAL_UNKNOWN dep=%s detail=%s\n", u.ID, u.ShortDetail())
		}
		if !opts.AllowUnknownExternal {
			return emit(stdout, stderr, opts.Format, externalUnknown)
		}
	}

	return emit(stdout, stderr, opts.Format, v)
}

// emit writes the verdict to stdout in the requested format and returns the
// exit code for it. In JSON mode this is one compact line and nothing else,
// so the output is machine-consumable without filtering: even the
// (practically unreachable) marshal-failure fallback goes to stderr, never
// stdout, so a JSON consumer never has to handle a fifth output shape.
func emit(stdout, stderr io.Writer, format string, v evidence.Verdict) int {
	if format == FormatText {
		switch {
		case v.OK:
			fmt.Fprintln(stdout, "VERIFIED")
		case v.State == "UNKNOWN":
			// An unknown verdict (SCAN_INCOMPLETE, EXTERNAL_UNKNOWN, ...) has
			// no "missing" count to report -- printing "0 missing" would
			// read as verified-except-nothing, which is not what UNKNOWN
			// means: the answer could not be determined at all.
			fmt.Fprintf(stdout, "UNKNOWN: %s\n", v.Code)
		default:
			fmt.Fprintf(stdout, "UNVERIFIED: %d missing\n", len(v.Missing))
		}
	} else {
		data, err := json.Marshal(v)
		if err != nil {
			// Verdict is four strings and a bool; Marshal cannot fail. The
			// branch exists only so the error is never silently dropped --
			// and it goes to stderr, not stdout, so stdout in JSON mode
			// never carries anything but the one verdict shape.
			fmt.Fprintf(stderr, "verify: marshal verdict: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "%s\n", data)
	}
	if v.OK {
		return 0
	}
	return 1
}

// loadEvidence reads the evidence store. A missing store is not an error: it
// simply means no evidence has been recorded yet, which Complete reports as
// EVIDENCE_MISSING rather than as an IO failure. A malformed store IS an
// error -- silently treating unparseable evidence as absent would hide
// tampering.
func loadEvidence(path string) ([]evidence.Record, error) {
	f, err := evidence.Load(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return f.Records, nil
}

// unknownExternals resolves every DEPENDS_ON edge declared by a claimed
// requirement and returns those that resolve to an external (ticket-source or
// peer) dependency whose state cannot be determined from disk. Local
// (flatfile) ids are not external and never appear here.
func unknownExternals(rep canaryscan.Report, claimed []string, reg *sources.Registry, root string) []external.Resolution {
	claimedSet := make(map[string]struct{}, len(claimed))
	for _, id := range claimed {
		claimedSet[id] = struct{}{}
	}
	seen := map[string]struct{}{}
	var deps []string
	for _, r := range rep.Requirements {
		if _, ok := claimedSet[r.ID]; !ok {
			continue
		}
		for _, f := range r.Features {
			for _, dep := range f.DependsOn {
				if _, dup := seen[dep]; dup {
					continue
				}
				seen[dep] = struct{}{}
				deps = append(deps, dep)
			}
		}
	}
	var unknown []external.Resolution
	for _, dep := range deps {
		res := external.Resolve(dep, reg, root)
		if res.IsExternal() && res.State == external.StateUnknown {
			unknown = append(unknown, res)
		}
	}
	return unknown
}
