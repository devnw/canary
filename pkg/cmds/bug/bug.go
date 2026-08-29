// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package bug

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/storage"
)

// BugCmd is the parent command for bug-related operations
var BugCmd = &cobra.Command{
	Use:   "bug",
	Short: "Manage BUG-* CANARY tokens for tracking defects",
	Long: `Bug command manages BUG-* CANARY tokens for defect tracking.

Unlike project-specific tokens (e.g., CBIN-XXX), bug tokens always use
the BUG- prefix followed by aspect and ID (e.g., BUG-API-001).

Subcommands:
  list    - List all bug canaries with filtering
  create  - Create a new bug canary token
  update  - Update an existing bug canary's status
  show    - Display details for a specific bug`,
}

// Helper functions

// buildBugToken constructs a CANARY bug token comment string (single-line format).
// The format is: // CANARY: BUG=<id>; TITLE="<title>"; FEATURE="<feature>"; ASPECT=<aspect>; STATUS=<status>; SEVERITY=<sev>; PRIORITY=<pri>; UPDATED=<date>
func buildBugToken(bugID, title, feature, aspect, status, severity, priority, updated string) string {
	return fmt.Sprintf("// CANARY: BUG=%s; TITLE=\"%s\"; FEATURE=\"%s\"; ASPECT=%s; STATUS=%s; SEVERITY=%s; PRIORITY=%s; UPDATED=%s\n",
		bugID, title, feature, aspect, status, severity, priority, updated)
}

func generateBugID(aspect string, dbPath string) (string, error) {
	// Normalize aspect to uppercase
	aspect = strings.ToUpper(aspect)

	// Open database to check existing IDs
	db, err := storage.Open(dbPath)
	if err != nil {
		// If no database, start from 001
		return fmt.Sprintf("BUG-%s-001", aspect), nil
	}
	defer db.Close()

	return generateBugIDWithDB(aspect, db)
}

func generateBugIDWithDB(aspect string, db *storage.DB) (string, error) {
	// Normalize aspect to uppercase
	aspect = strings.ToUpper(aspect)

	// Query ALL tokens (no pattern filter since ListTokens is hardcoded for CBIN)
	tokens, err := db.ListTokens(nil, "", "req_id DESC", 0) // Get all tokens
	if err != nil {
		// If database doesn't have tokens table yet, start from 001
		return fmt.Sprintf("BUG-%s-001", aspect), nil
	}

	// Find highest number for this aspect
	maxNum := 0
	pattern := fmt.Sprintf(`BUG-%s-([0-9]{3})`, aspect)
	re := regexp.MustCompile(pattern)
	for _, token := range tokens {
		// Extract number from matching tokens
		if matches := re.FindStringSubmatch(token.ReqID); len(matches) > 1 {
			num, _ := strconv.Atoi(matches[1])
			if num > maxNum {
				maxNum = num
			}
		}
	}

	// Generate next ID
	nextNum := maxNum + 1
	return fmt.Sprintf("BUG-%s-%03d", aspect, nextNum), nil
}

func filterBugTokens(tokens []*storage.Token, severity, priority string) []*storage.Token {
	if severity == "" && priority == "" {
		return tokens
	}

	var filtered []*storage.Token
	for _, token := range tokens {
		sev, pri := parseBugMetadata(token.Keywords)

		// Check severity filter
		if severity != "" && !strings.Contains(severity, sev) {
			continue
		}

		// Check priority filter
		if priority != "" && !strings.Contains(priority, pri) {
			continue
		}

		filtered = append(filtered, token)
	}
	return filtered
}

func parseBugMetadata(metadata string) (severity, priority string) {
	severity = "S3" // default
	priority = "P2" // default

	parts := strings.Split(metadata, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "SEVERITY=") {
			severity = strings.TrimPrefix(part, "SEVERITY=")
		} else if strings.HasPrefix(part, "PRIORITY=") {
			priority = strings.TrimPrefix(part, "PRIORITY=")
		}
	}
	return severity, priority
}

func parsePriorityValue(priority string) int {
	// Convert P0, P1, P2, P3 to numeric values
	switch priority {
	case "P0":
		return 0
	case "P1":
		return 1
	case "P2":
		return 2
	case "P3":
		return 3
	default:
		return 2 // default
	}
}

func formatBugList(tokens []*storage.Token, noColor bool) {
	// Group by status
	statusGroups := make(map[string][]*storage.Token)
	for _, token := range tokens {
		statusGroups[token.Status] = append(statusGroups[token.Status], token)
	}

	// Display in order: OPEN, IN_PROGRESS, FIXED, others
	statusOrder := []string{"OPEN", "IN_PROGRESS", "FIXED", "VERIFIED", "BLOCKED", "WONTFIX", "DUPLICATE"}

	for _, status := range statusOrder {
		if bugs, ok := statusGroups[status]; ok && len(bugs) > 0 {
			fmt.Printf("\n## %s (%d)\n\n", status, len(bugs))
			for _, bug := range bugs {
				severity, priority := parseBugMetadata(bug.Keywords)

				// Format with colors if enabled
				var line string
				if !noColor {
					switch severity {
					case "S1", "S1-Critical":
						line = color.RedString("🔴")
					case "S2", "S2-High":
						line = color.YellowString("🟠")
					case "S3", "S3-Medium":
						line = color.BlueString("🔵")
					default:
						line = "⚪"
					}
				} else {
					line = "•"
				}

				line += fmt.Sprintf(" %s: %s", bug.ReqID, bug.Feature)
				if bug.Owner != "" {
					line += fmt.Sprintf(" [%s]", bug.Owner)
				}
				line += fmt.Sprintf(" (%s/%s)", severity, priority)
				fmt.Println(line)
			}
		}
	}

	// Show any status not in the predefined order
	for status, bugs := range statusGroups {
		found := false
		for _, s := range statusOrder {
			if s == status {
				found = true
				break
			}
		}
		if !found && len(bugs) > 0 {
			fmt.Printf("\n## %s (%d)\n\n", status, len(bugs))
			for _, bug := range bugs {
				severity, priority := parseBugMetadata(bug.Keywords)
				fmt.Printf("• %s: %s (%s/%s)\n", bug.ReqID, bug.Feature, severity, priority)
			}
		}
	}

	fmt.Printf("\n📊 Total bugs: %d\n", len(tokens))
}

func listBugsFromFilesystem(aspect, status, severity, priority string, jsonOutput, noColor bool, limit int) error {
	// Fallback when database is not available: use canary scanner and filter BUG tokens.
	fmt.Fprintf(os.Stderr, "⚠️  Database not found, using filesystem search (slower)\n")
	fmt.Fprintf(os.Stderr, "   Suggestion: Run 'canary index' to build database\n\n")

	root := "."
	if wd, err := os.Getwd(); err == nil {
		root = wd
	}
	skip := canaryscan.DefaultSkipRegex()
	rep, err := canaryscan.Scan(root, skip, nil, nil)
	if err != nil {
		return fmt.Errorf("filesystem scan: %w", err)
	}

	bugPattern := regexp.MustCompile(`^BUG-[A-Za-z0-9]+-[0-9]{3}$`)
	var tokens []*storage.Token
	for _, req := range rep.Requirements {
		if !bugPattern.MatchString(req.ID) {
			continue
		}
		for _, f := range req.Features {
			if aspect != "" && f.Aspect != aspect {
				continue
			}
			if status != "" && f.Status != status {
				continue
			}
			filePath := "."
			if len(f.Files) > 0 {
				filePath = f.Files[0]
				if !filepath.IsAbs(filePath) {
					filePath = filepath.Join(root, filePath)
				}
			}
			tokens = append(tokens, &storage.Token{
				ReqID:     req.ID,
				Feature:   f.Feature,
				Aspect:    f.Aspect,
				Status:    f.Status,
				FilePath:  filePath,
				UpdatedAt: f.Updated,
			})
		}
	}

	filtered := filterBugTokens(tokens, severity, priority)
	truncated := limit > 0 && len(filtered) > limit
	if truncated {
		filtered = filtered[:limit]
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(filtered)
	}
	if len(filtered) == 0 {
		fmt.Println("No bug tokens found")
		return nil
	}
	formatBugList(filtered, noColor)
	if truncated {
		fmt.Printf("(showing %d; use --limit -1 for all)\n", limit)
	}
	return nil
}

func createBugCanaryComment(token *storage.Token, severity, priority string) error {
	// Create CANARY comment in the specified file
	canaryComment := buildBugToken(token.ReqID, token.Feature, token.Feature, token.Aspect, token.Status, severity, priority, token.UpdatedAt)

	fmt.Printf("✅ Bug token created: %s\n", token.ReqID)
	fmt.Printf("\nAdd this CANARY comment to %s:%d:\n\n", token.FilePath, token.LineNumber)
	fmt.Println(canaryComment)

	return nil
}

func init() {
	// Add subcommands
	BugCmd.AddCommand(bugListCmd)
	BugCmd.AddCommand(bugCreateCmd)
	BugCmd.AddCommand(bugUpdateCmd)
	BugCmd.AddCommand(bugShowCmd)
}
