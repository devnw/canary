// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package show

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"devnw.dev/canary"
	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/cmds/internal/utils"
	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/storage"
)

// CANARY: REQ=CBIN-CLI-001; FEATURE="ShowCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_ShowCmd; UPDATED=2026-08-30
var ShowCmd = &cobra.Command{
	Use:   "show <REQ-ID>",
	Short: "Display all CANARY tokens for a requirement",
	Long: `Show displays all CANARY tokens for a specific requirement ID.

Displays:
- Feature name, aspect, status
- File location and line number
- Test and benchmark references
- Owner and priority

Resolution:
- An exact requirement ID is shown directly.
- A prefix that uniquely identifies one requirement is resolved to it.
- A prefix that matches several requirements is reported as ambiguous, with
  the sorted candidate list, and exits non-zero.
- When there is no index (no .canary/canary.db), show falls back to a live
  filesystem scan of the working directory and marks the output accordingly.

Grouping:
- By default, groups by aspect (CLI, API, Engine, etc.)
- Use --group-by status to group by implementation status
- Use --json for machine-readable output

Examples:
  canary show CBIN-133
  canary show CBIN-133 --group-by status
  canary show CBIN-133 --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt, _ := cmd.Flags().GetString("prompt")
		var promptContent string
		if prompt != "" {
			c, err := utils.LoadPrompt(prompt)
			if err != nil {
				return err
			}
			promptContent = c
		}
		reqID := args[0]
		groupBy, _ := cmd.Flags().GetString("group-by")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if noColor, _ := cmd.Flags().GetBool("no-color"); noColor {
			color.NoColor = true
		}

		dbPath, _ := cmd.Flags().GetString("db")
		projectID := utils.ReadProjectID(cmd)

		tokens, source, err := loadTokens(cmd, dbPath, projectID, reqID)
		if err != nil {
			return err
		}

		if len(tokens) == 0 {
			// In JSON mode stdout carries only the (empty) token array,
			// matching the success shape below; the diagnostic prose goes
			// to stderr so a caller parsing stdout as JSON never chokes on
			// it. The exit-1 "requirement not found" contract is unchanged
			// in both modes.
			if jsonOutput {
				fmt.Fprintf(os.Stderr, "No tokens found for %s\n", reqID)
				fmt.Fprintln(os.Stderr, "\nSuggestions:")
				fmt.Fprintln(os.Stderr, "  • Run: canary list")
				fmt.Fprintln(os.Stderr, "  • Check requirement ID format (e.g., CBIN-XXX)")
				if jerr := outputTokensJSON([]*storage.Token{}); jerr != nil {
					return jerr
				}
				return fmt.Errorf("requirement not found")
			}

			fmt.Printf("No tokens found for %s\n", reqID)
			fmt.Println("\nSuggestions:")
			fmt.Println("  • Run: canary list")
			fmt.Println("  • Check requirement ID format (e.g., CBIN-XXX)")
			return fmt.Errorf("requirement not found")
		}

		// Format output
		if jsonOutput {
			return outputTokensJSON(tokens)
		}

		if promptContent != "" {
			fmt.Println(promptContent)
			fmt.Println()
		}
		fmt.Printf("Tokens for %s (source: %s):\n\n", tokens[0].ReqID, source)
		output := canary.FormatTokensTable(tokens, groupBy)
		fmt.Println(output)

		return nil
	},
}

// loadTokens resolves reqID against whichever token source is available and
// returns the tokens for the resolved requirement plus a source label
// ("index" or "filesystem"). On no index (fs.ErrNotExist) it falls back to a
// live filesystem scan; any other database error is returned. An ambiguous
// prefix returns a contract error after printing the sorted candidates.
func loadTokens(cmd *cobra.Command, dbPath, projectID, reqID string) ([]*storage.Token, string, error) {
	db, err := storage.OpenRO(dbPath)
	switch {
	case err == nil:
		defer db.Close()
		resolved, cerr := resolveFromDB(cmd, db, projectID, reqID)
		if cerr != nil {
			return nil, "", cerr
		}
		if resolved == "" {
			return nil, "index", nil // not found
		}
		toks, terr := db.GetTokensByReqID(projectID, resolved)
		if terr != nil {
			return nil, "", utils.GuardContract(cmd, terr)
		}
		return toks, "index", nil

	case errors.Is(err, fs.ErrNotExist):
		cmd.SilenceUsage = true
		return resolveFromFilesystem(cmd, reqID)

	case errors.Is(err, storage.ErrSchemaOutOfDate):
		cmd.SilenceUsage = true
		return nil, "", err

	default:
		cmd.SilenceUsage = true
		return nil, "", fmt.Errorf("open database: %w", err)
	}
}

// resolveFromDB returns the concrete requirement id to show from the index, or
// "" when nothing matches. An ambiguous prefix prints candidates and returns a
// contract error.
func resolveFromDB(cmd *cobra.Command, db *storage.DB, projectID, reqID string) (string, error) {
	// include_hidden=true so a requirement whose only tokens live in test
	// files or template dirs is still a candidate; limit 0 = unbounded.
	all, err := db.ListTokens(projectID, map[string]any{"include_hidden": "true"}, "", "", 0)
	if err != nil {
		return "", utils.GuardContract(cmd, err)
	}
	ids := make([]string, 0, len(all))
	for _, t := range all {
		ids = append(ids, t.ReqID)
	}
	return resolveReqID(cmd, reqID, ids)
}

// resolveFromFilesystem scans the working directory and returns the tokens for
// the resolved requirement.
func resolveFromFilesystem(cmd *cobra.Command, reqID string) ([]*storage.Token, string, error) {
	root := "."
	reg, err := sources.LoadFromRoot(root)
	if err != nil {
		return nil, "", fmt.Errorf("load sources: %w", err)
	}
	ignorePatterns, err := canaryscan.LoadCanaryIgnore(root)
	if err != nil {
		return nil, "", fmt.Errorf("load .canaryignore: %w", err)
	}
	rep, err := canaryscan.Scan(root, canaryscan.DefaultSkipRegex(), nil, ignorePatterns, reg)
	if err != nil {
		return nil, "", fmt.Errorf("scan filesystem: %w", err)
	}

	ids := make([]string, 0, len(rep.Requirements))
	for _, r := range rep.Requirements {
		ids = append(ids, r.ID)
	}
	resolved, rerr := resolveReqID(cmd, reqID, ids)
	if rerr != nil {
		return nil, "", rerr
	}
	if resolved == "" {
		return nil, "filesystem", nil // not found
	}
	for _, r := range rep.Requirements {
		if r.ID == resolved {
			return tokensFromRequirement(r), "filesystem", nil
		}
	}
	return nil, "filesystem", nil
}

// resolveReqID resolves arg against the known requirement ids. An exact match
// wins; otherwise arg is a prefix -- a unique prefix resolves to that id, and a
// prefix matching several ids prints the sorted candidates and returns a
// contract error. "" (no error) means nothing matched.
func resolveReqID(cmd *cobra.Command, arg string, ids []string) (string, error) {
	seen := map[string]struct{}{}
	uniq := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	if _, ok := seen[arg]; ok {
		return arg, nil
	}
	var pref []string
	for _, id := range uniq {
		if strings.HasPrefix(id, arg) {
			pref = append(pref, id)
		}
	}
	sort.Strings(pref)
	switch len(pref) {
	case 0:
		return "", nil
	case 1:
		return pref[0], nil
	default:
		cmd.SilenceUsage = true
		fmt.Fprintln(os.Stderr, "ambiguous requirement id; candidates:")
		for _, c := range pref {
			fmt.Fprintf(os.Stderr, "  %s\n", c)
		}
		return "", fmt.Errorf("ambiguous requirement id %q", arg)
	}
}

// tokensFromRequirement adapts a scanned requirement's features into the
// storage.Token shape FormatTokensTable renders, so the filesystem fallback
// produces the same output as the indexed path.
func tokensFromRequirement(r canaryscan.Requirement) []*storage.Token {
	toks := make([]*storage.Token, 0, len(r.Features))
	for _, f := range r.Features {
		t := &storage.Token{
			ReqID:     r.ID,
			Feature:   f.Feature,
			Aspect:    f.Aspect,
			Status:    f.Status,
			Owner:     f.Owner,
			Test:      strings.Join(f.Tests, ","),
			Bench:     strings.Join(f.Benches, ","),
			UpdatedAt: f.Updated,
		}
		if len(f.Files) > 0 {
			t.FilePath = f.Files[0]
		}
		if f.Priority != nil {
			t.Priority = *f.Priority
		}
		toks = append(toks, t)
	}
	return toks
}

// outputTokensJSON outputs tokens as JSON
func outputTokensJSON(tokens []*storage.Token) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(tokens)
}

func init() {
	ShowCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name to print above the token table")
	ShowCmd.Flags().String("group-by", "aspect", "Group tokens by field (aspect, status)")
	ShowCmd.Flags().Bool("json", false, "Output in JSON format")
	ShowCmd.Flags().Bool("no-color", false, "Disable colored output")
	ShowCmd.Flags().String("db", ".canary/canary.db", "Path to database file")
	utils.AddProjectFlag(ShowCmd)
}
