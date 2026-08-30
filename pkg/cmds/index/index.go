package index

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	ignore "github.com/sabhiram/go-gitignore"
	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/cmds/internal/utils"
	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/storage"
)

// CANARY: REQ=ENG-4307; FEATURE="IndexCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestAuditF12,TestAuditF12CleanRunCommits,TestCANARY_CP_285_IndexRespectsCanaryIgnore; UPDATED=2026-08-30
var IndexCmd = &cobra.Command{
	Use:   "index [flags]",
	Short: "Build or rebuild the CANARY token database",
	Long: `Scan the codebase for CANARY tokens and store metadata in SQLite database.

This enables advanced features like priority ordering, keyword search, and checkpoints.
The database is stored at .canary/canary.db by default.

The rebuild is one transaction: either the new index fully replaces the old
one, or the old one is left exactly as it was. A token the scanner cannot
parse fails the run rather than being skipped with a warning.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt, _ := cmd.Flags().GetString("prompt")
		if prompt != "" {
			if _, err := utils.LoadPrompt(prompt); err != nil {
				return err
			}
		}
		dbPath, _ := cmd.Flags().GetString("db")
		rootPath, _ := cmd.Flags().GetString("root")

		projectID, err := utils.WriteProjectID(cmd, rootPath)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Indexing CANARY tokens from: %s\n", rootPath)

		reg, err := sources.LoadFromRoot(rootPath)
		if err != nil {
			return fmt.Errorf("load .canary/project.yaml: %w", err)
		}

		ignorePatterns, err := canaryscan.LoadCanaryIgnore(rootPath)
		if err != nil {
			return fmt.Errorf("load .canaryignore: %w", err)
		}

		// One canonical scan. The command used to shell out to grep with a
		// hardcoded extension list and its own field extraction, so the index
		// and `canary scan` could disagree about what the tree contains.
		records, files, issues, err := canaryscan.ScanTokenRecords(rootPath, nil, nil, ignorePatterns, reg)
		if err != nil {
			return fmt.Errorf("scan %s: %w", rootPath, err)
		}
		if err := reportIssues("token", issues); err != nil {
			return err
		}

		commitHash, branch := gitMetadata(rootPath)

		indexedAt := time.Now().UTC().Format(time.RFC3339)
		tokens := make([]*storage.Token, 0, len(records))
		for _, rec := range records {
			tokens = append(tokens, toToken(rec, projectID, commitHash, branch, indexedAt))
		}

		refs, err := collectRefs(rootPath, reg, ignorePatterns)
		if err != nil {
			return err
		}

		db, err := storage.OpenRW(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer func() { _ = db.Close() }()

		meta := storage.IndexMeta{
			Root:         rootPath,
			ProjectID:    projectID,
			CommitSHA:    commitHash,
			ParserSchema: canaryscan.ParserSchemaVersion,
			ScanDigest:   canaryscan.ScanDigest(files),
			IndexedAt:    indexedAt,
		}

		if err := db.ReplaceIndex(projectID, tokens, refs, meta); err != nil {
			return fmt.Errorf("rebuild index: %w", err)
		}

		fmt.Fprintf(out, "Indexed %d diagram reference(s)\n", len(refs["diagram"]))
		fmt.Fprintf(out, "Indexed %d migration note(s)\n", len(refs["migrate"]))
		fmt.Fprintf(out, "\n✅ Indexed %d CANARY tokens\n", len(tokens))
		fmt.Fprintf(out, "Database: %s\n", dbPath)
		fmt.Fprintf(out, "Project: %s\n", projectID)
		if commitHash != "" {
			fmt.Fprintf(out, "Commit: %s\n", commitHash[:8])
		}
		if branch != "" {
			fmt.Fprintf(out, "Branch: %s\n", branch)
		}

		return nil
	},
}

// gitMetadata resolves the commit and branch of the tree at rootPath. Every
// invocation sets cmd.Dir: the old code built the command and then tested
// `gitCmd.Dir == ""`, which was always true, so it read whatever repository
// the process happened to be standing in -- or nothing at all. A tree that is
// not a git repository yields empty strings, never a fabricated SHA.
func gitMetadata(rootPath string) (commit, branch string) {
	run := func(args ...string) string {
		gitCmd := exec.Command("git", args...) //nolint:gosec // fixed argv
		gitCmd.Dir = rootPath
		out, err := gitCmd.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	return run("rev-parse", "HEAD"), run("rev-parse", "--abbrev-ref", "HEAD")
}

// toToken maps one scanned record onto a storage row.
//
// Absent fields stay absent. UPDATED in particular is stored verbatim -- the
// old code substituted today's date for a token that never declared one,
// which turned "we do not know when this was last touched" into a freshness
// claim the author never made. ScanTokenRecords already rejects a token with
// no UPDATED, so an empty value here can only mean a field this index does
// not require.
func toToken(rec canaryscan.TokenRecord, projectID, commitHash, branch, indexedAt string) *storage.Token {
	docPath := rec.Field("DOC")
	docType := rec.Field("DOC_TYPE")
	if docPath != "" && docType == "" {
		// Infer the type from the first path's "type:path" prefix.
		firstPath := strings.Split(docPath, ",")[0]
		if strings.Contains(firstPath, ":") {
			docType = strings.Split(firstPath, ":")[0]
		}
	}

	priority := defaultPriority
	if p, err := strconv.Atoi(rec.Field("PRIORITY")); err == nil {
		priority = p
	}

	specStatus := rec.Field("SPEC_STATUS")
	if specStatus == "" {
		specStatus = "draft"
	}

	return &storage.Token{
		ReqID:       rec.ReqID,
		Feature:     rec.Field("FEATURE"),
		Aspect:      rec.Field("ASPECT"),
		Status:      rec.Field("STATUS"),
		FilePath:    rec.File,
		LineNumber:  rec.Line,
		Test:        rec.Field("TEST"),
		Bench:       rec.Field("BENCH"),
		Owner:       rec.Field("OWNER"),
		Priority:    priority,
		Phase:       rec.Field("PHASE"),
		Keywords:    rec.Field("KEYWORDS"),
		SpecStatus:  specStatus,
		CreatedAt:   rec.Field("CREATED"),
		UpdatedAt:   rec.Field("UPDATED"),
		StartedAt:   rec.Field("STARTED"),
		CompletedAt: rec.Field("COMPLETED"),
		CommitHash:  commitHash,
		Branch:      branch,
		DependsOn:   rec.Field("DEPENDS_ON"),
		Blocks:      rec.Field("BLOCKS"),
		RelatedTo:   rec.Field("RELATED_TO"),
		RawToken:    rec.Raw,
		IndexedAt:   indexedAt,
		DocPath:     docPath,
		DocHash:     rec.Field("DOC_HASH"),
		DocType:     docType,
		ProjectID:   projectID,
		ContentHash: rec.ContentHash,
	}
}

// defaultPriority is the priority a token that declares none is indexed with.
const defaultPriority = 5

// collectRefs gathers the non-token requirement references (mermaid diagrams,
// CANARY:MIGRATE guidance notes) that `canary view` answers from, keyed by
// kind so they can be replaced inside the same transaction as the tokens.
func collectRefs(rootPath string, reg *sources.Registry, ignorePatterns *ignore.GitIgnore) (map[string][]storage.Ref, error) {
	refs := map[string][]storage.Ref{}

	diagRefs, diagIssues, err := canaryscan.ScanDiagramRefs(rootPath, nil, reg, ignorePatterns)
	if err != nil {
		return nil, fmt.Errorf("scan diagram references: %w", err)
	}
	if err := reportIssues("diagram", diagIssues); err != nil {
		return nil, err
	}
	diagrams := make([]storage.Ref, 0, len(diagRefs))
	for _, r := range diagRefs {
		diagrams = append(diagrams, storage.Ref{ReqID: r.ReqID, Kind: "diagram", FilePath: r.File, LineNumber: r.Line})
	}
	refs["diagram"] = diagrams

	// CANARY: REQ=ENG-4325; FEATURE="MigrateNotesIndex"; ASPECT=CLI; STATUS=IMPL; UPDATED=2026-08-30
	// One ref row per (note, associated ReqID); a note that matched no
	// requirement still gets one row with req_id='' so it isn't dropped.
	notes, noteIssues, err := canaryscan.ScanMigrateNotes(rootPath, nil, ignorePatterns, reg)
	if err != nil {
		return nil, fmt.Errorf("scan migration notes: %w", err)
	}
	if err := reportIssues("migrate", noteIssues); err != nil {
		return nil, err
	}
	migrate := make([]storage.Ref, 0, len(notes))
	for _, n := range notes {
		if len(n.ReqIDs) == 0 {
			migrate = append(migrate, storage.Ref{ReqID: "", Kind: "migrate", FilePath: n.File, LineNumber: n.Line, Context: n.Text})
			continue
		}
		for _, id := range n.ReqIDs {
			migrate = append(migrate, storage.Ref{ReqID: id, Kind: "migrate", FilePath: n.File, LineNumber: n.Line, Context: n.Text})
		}
	}
	refs["migrate"] = migrate

	return refs, nil
}

// reportIssues prints every scan issue on stderr and fails the run if any of
// them is a parse error.
//
// The distinction is what "strict" means here. A parse error says a CANARY
// token exists and the parser rejects it -- indexing the rest would publish
// an index that silently disagrees with the source, which is exactly the
// warn-and-continue behaviour this replaced. Every other reason (binary,
// oversized, unreadable) says a *file* was skipped, and a file with no
// readable tokens contributes nothing to lose: an image or a database in the
// tree must not be able to block the index.
// A parse error is printed per file because each one names a token someone
// has to go fix. The benign reasons are not: a repository with a vendor tree
// or a pile of fixtures produces hundreds of identical "skipped a binary"
// lines, and burying the handful of actionable lines under them is how the
// actionable ones stop being read. Those are counted and summarised, one
// line per reason.
func reportIssues(kind string, issues []canaryscan.ScanIssue) error {
	// Fixed order so the summary is stable run to run rather than following
	// map iteration.
	benignOrder := []string{
		canaryscan.IssueBinary,
		canaryscan.IssueFileTooLarge,
		canaryscan.IssueReadError,
		canaryscan.IssueLineTooLarge,
	}
	benign := make(map[string]int, len(benignOrder))
	for _, reason := range benignOrder {
		benign[reason] = 0
	}

	parseErrors := 0
	for _, is := range issues {
		if _, isBenign := benign[is.Reason]; isBenign {
			benign[is.Reason]++
			continue
		}
		fmt.Fprintf(os.Stderr, "CANARY_SCAN_ISSUE kind=%s path=%s reason=%s detail=%s\n", kind, is.Path, is.Reason, is.Detail)
		if is.Reason == canaryscan.IssueParseError {
			parseErrors++
		}
	}

	for _, reason := range benignOrder {
		if n := benign[reason]; n > 0 {
			fmt.Fprintf(os.Stderr, "CANARY_SCAN_SKIPPED kind=%s reason=%s files=%d\n", kind, reason, n)
		}
	}

	if parseErrors > 0 {
		return fmt.Errorf("refusing to index: %d unparseable CANARY token(s) in the %s scan", parseErrors, kind)
	}
	return nil
}

func init() {
	IndexCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	IndexCmd.Flags().String("db", ".canary/canary.db", "path to database file")
	IndexCmd.Flags().String("root", ".", "root directory to scan")
	utils.AddProjectFlag(IndexCmd)
}
