// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package bug

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/storage"
)

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

var bugListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all BUG-* CANARY tokens",
	Long: `List all BUG-* CANARY tokens with optional filtering.

Examples:
  canary bug list
  canary bug list --aspect API
  canary bug list --status OPEN --severity S1
  canary bug list --priority P0,P1
  canary bug list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		aspect, _ := cmd.Flags().GetString("aspect")
		status, _ := cmd.Flags().GetString("status")
		severity, _ := cmd.Flags().GetString("severity")
		priority, _ := cmd.Flags().GetString("priority")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		noColor, _ := cmd.Flags().GetBool("no-color")
		limit, _ := cmd.Flags().GetInt("limit")
		dbPath, _ := cmd.Flags().GetString("db")

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			// Fallback to filesystem search if no database
			return listBugsFromFilesystem(aspect, status, severity, priority, jsonOutput, noColor, limit)
		}
		defer db.Close()

		// Build filters for BUG tokens
		filters := make(map[string]string)
		if aspect != "" {
			filters["aspect"] = aspect
		}
		if status != "" {
			filters["status"] = status
		}

		// Query database for all tokens (ListTokens is hardcoded for CBIN patterns)
		allTokens, err := db.ListTokens(filters, "", "priority ASC, updated_at DESC", 0)
		if err != nil {
			return fmt.Errorf("query bugs: %w", err)
		}

		// Filter for BUG tokens only
		var tokens []*storage.Token
		bugPattern := regexp.MustCompile(`^BUG-[A-Za-z]+-[0-9]{3}$`)
		for _, tok := range allTokens {
			if bugPattern.MatchString(tok.ReqID) {
				tokens = append(tokens, tok)
			}
		}

		// Additional filtering for severity and priority (stored in token comments or metadata)
		filteredTokens := filterBugTokens(tokens, severity, priority)

		// Apply limit if specified
		if limit > 0 && len(filteredTokens) > limit {
			filteredTokens = filteredTokens[:limit]
		}

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(filteredTokens)
		}

		// Format output
		if len(filteredTokens) == 0 {
			fmt.Println("No bug tokens found")
			return nil
		}

		formatBugList(filteredTokens, noColor)
		return nil
	},
}

var bugCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new BUG-* CANARY token",
	Long: `Create a new BUG-* CANARY token for tracking a defect.

The bug ID will be automatically generated in the format BUG-SECURITY_REVIEW-XXX.

Examples:
  canary bug create "Login fails on first attempt"
  canary bug create "API returns 500 error" --aspect API --severity S1 --priority P0
  canary bug create "Memory leak in storage engine" --aspect Storage --file src/storage/cache.go:42`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.Join(args, " ")
		aspect, _ := cmd.Flags().GetString("aspect")
		severity, _ := cmd.Flags().GetString("severity")
		priority, _ := cmd.Flags().GetString("priority")
		status, _ := cmd.Flags().GetString("status")
		file, _ := cmd.Flags().GetString("file")
		owner, _ := cmd.Flags().GetString("owner")
		dbPath, _ := cmd.Flags().GetString("db")

		// Default values
		if aspect == "" {
			aspect = "API" // Default aspect
		}
		if status == "" {
			status = "OPEN"
		}
		if severity == "" {
			severity = "S3"
		}
		if priority == "" {
			priority = "P2"
		}

		// Generate bug ID
		bugID, err := generateBugID(aspect, dbPath)
		if err != nil {
			return fmt.Errorf("generate bug ID: %w", err)
		}

		// Parse file location if provided
		var filePath string
		var lineNum int
		if file != "" {
			parts := strings.Split(file, ":")
			filePath = parts[0]
			if len(parts) > 1 {
				lineNum, _ = strconv.Atoi(parts[1])
			}
		} else {
			// Default to main.go or most relevant file
			filePath = "main.go"
			lineNum = 1
		}

		// Create token
		token := &storage.Token{
			ReqID:      bugID,
			Feature:    title,
			Aspect:     aspect,
			Status:     status,
			FilePath:   filePath,
			LineNumber: lineNum,
			UpdatedAt:  time.Now().Format("2006-01-02"),
			Owner:      owner,
			Priority:   parsePriorityValue(priority),
			Keywords:   fmt.Sprintf("SEVERITY=%s;PRIORITY=%s", severity, priority),
		}

		// Save to database
		db, err := storage.Open(dbPath)
		if err != nil {
			// Create CANARY comment in file if no database
			return createBugCanaryComment(token, severity, priority)
		}
		defer db.Close()

		if err := db.UpsertToken(token); err != nil {
			return fmt.Errorf("save bug token: %w", err)
		}

		// Generate CANARY comment format
		canaryComment := fmt.Sprintf(
			"// CANARY: BUG=%s; TITLE=\"%s\";\n"+
				"//         ASPECT=%s; STATUS=%s;\n"+
				"//         SEVERITY=%s; PRIORITY=%s;\n"+
				"//         UPDATED=%s",
			bugID, title, aspect, status,
			severity, priority,
			time.Now().Format("2006-01-02"),
		)

		fmt.Printf("✅ Created bug token: %s\n", bugID)
		fmt.Printf("📝 Title: %s\n", title)
		fmt.Printf("📊 Severity: %s | Priority: %s\n", severity, priority)
		fmt.Printf("📍 Location: %s:%d\n", filePath, lineNum)
		fmt.Printf("\n%s CANARY comment to add:\n", color.YellowString("→"))
		fmt.Println(canaryComment)

		return nil
	},
}

var bugUpdateCmd = &cobra.Command{
	Use:   "update <BUG-ID>",
	Short: "Update a BUG-* CANARY token's status",
	Long: `Update the status or other properties of a BUG-* CANARY token.

Examples:
  canary bug update BUG-API-001 --status FIXED
  canary bug update BUG-CLI-002 --status IN_PROGRESS --owner alice
  canary bug update BUG-Storage-003 --priority P0 --severity S1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bugID := args[0]
		status, _ := cmd.Flags().GetString("status")
		severity, _ := cmd.Flags().GetString("severity")
		priority, _ := cmd.Flags().GetString("priority")
		owner, _ := cmd.Flags().GetString("owner")
		dbPath, _ := cmd.Flags().GetString("db")

		// Validate bug ID format
		if !regexp.MustCompile(`^BUG-[A-Za-z]+-[0-9]{3}$`).MatchString(bugID) {
			return fmt.Errorf("invalid bug ID format: %s (expected BUG-SECURITY_REVIEW-XXX)", bugID)
		}

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		// Get existing token
		tokens, err := db.GetTokensByReqID(bugID)
		if err != nil {
			return fmt.Errorf("find bug: %w", err)
		}
		if len(tokens) == 0 {
			return fmt.Errorf("bug not found: %s", bugID)
		}

		token := tokens[0]
		updated := false

		// Update fields if provided
		if status != "" {
			token.Status = status
			updated = true
		}
		if owner != "" {
			token.Owner = owner
			updated = true
		}
		if priority != "" {
			token.Priority = parsePriorityValue(priority)
			updated = true
		}

		// Update keywords for severity/priority
		if severity != "" || priority != "" {
			metaParts := strings.Split(token.Keywords, ";")

			// Parse existing keywords
			existingSev := "S3"
			existingPri := "P2"
			for _, part := range metaParts {
				if strings.HasPrefix(part, "SEVERITY=") {
					existingSev = strings.TrimPrefix(part, "SEVERITY=")
				} else if strings.HasPrefix(part, "PRIORITY=") {
					existingPri = strings.TrimPrefix(part, "PRIORITY=")
				}
			}

			if severity != "" {
				existingSev = severity
			}
			if priority != "" {
				existingPri = priority
			}

			token.Keywords = fmt.Sprintf("SEVERITY=%s;PRIORITY=%s", existingSev, existingPri)
			updated = true
		}

		if !updated {
			fmt.Println("No changes specified")
			return nil
		}

		// Update timestamp
		token.UpdatedAt = time.Now().Format("2006-01-02")

		// Save updated token
		if err := db.UpsertToken(token); err != nil {
			return fmt.Errorf("update bug: %w", err)
		}

		fmt.Printf("✅ Updated bug %s\n", bugID)
		fmt.Printf("📊 Status: %s\n", token.Status)
		if owner != "" {
			fmt.Printf("👤 Owner: %s\n", token.Owner)
		}
		if token.Keywords != "" {
			fmt.Printf("📈 Metadata: %s\n", token.Keywords)
		}

		return nil
	},
}

var bugShowCmd = &cobra.Command{
	Use:   "show <BUG-ID>",
	Short: "Display details for a specific bug",
	Long: `Show detailed information about a specific BUG-* CANARY token.

Examples:
  canary bug show BUG-API-001
  canary bug show BUG-CLI-002 --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bugID := args[0]
		jsonOutput, _ := cmd.Flags().GetBool("json")
		dbPath, _ := cmd.Flags().GetString("db")

		// Validate bug ID format
		if !regexp.MustCompile(`^BUG-[A-Za-z]+-[0-9]{3}$`).MatchString(bugID) {
			return fmt.Errorf("invalid bug ID format: %s (expected BUG-SECURITY_REVIEW-XXX)", bugID)
		}

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		// Get bug token
		tokens, err := db.GetTokensByReqID(bugID)
		if err != nil {
			return fmt.Errorf("query bug: %w", err)
		}
		if len(tokens) == 0 {
			return fmt.Errorf("bug not found: %s", bugID)
		}

		token := tokens[0]

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(token)
		}

		// Parse keywords for severity/priority
		severity, priority := parseBugMetadata(token.Keywords)

		// Format output
		fmt.Printf("🐛 Bug Details: %s\n\n", bugID)
		fmt.Printf("📝 Title: %s\n", token.Feature)
		fmt.Printf("📊 Status: %s | Aspect: %s\n", token.Status, token.Aspect)
		fmt.Printf("⚠️  Severity: %s | Priority: %s\n", severity, priority)
		fmt.Printf("📍 Location: %s:%d\n", token.FilePath, token.LineNumber)
		if token.Owner != "" {
			fmt.Printf("👤 Owner: %s\n", token.Owner)
		}
		fmt.Printf("📅 Updated: %s\n", token.UpdatedAt)
		if token.Test != "" {
			fmt.Printf("🧪 Test: %s\n", token.Test)
		}

		return nil
	},
}

// Helper functions

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
	// Fallback implementation for when database is not available
	fmt.Fprintf(os.Stderr, "⚠️  Database not found, using filesystem search (slower)\n")
	fmt.Fprintf(os.Stderr, "   Suggestion: Run 'canary index' to build database\n\n")

	// TODO: Implement filesystem-based search for BUG tokens
	// This would scan files for CANARY comments starting with BUG=

	return fmt.Errorf("filesystem search not yet implemented for bug tokens")
}

func createBugCanaryComment(token *storage.Token, severity, priority string) error {
	// Create CANARY comment in the specified file
	canaryComment := fmt.Sprintf(
		"// CANARY: BUG=%s; TITLE=\"%s\";\n"+
			"//         ASPECT=%s; STATUS=%s;\n"+
			"//         SEVERITY=%s; PRIORITY=%s;\n"+
			"//         UPDATED=%s",
		token.ReqID, token.Feature, token.Aspect, token.Status,
		severity, priority,
		token.UpdatedAt,
	)

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

	// List command flags
	bugListCmd.Flags().String("aspect", "", "Filter by aspect (API, CLI, Engine, Storage, etc.)")
	bugListCmd.Flags().String("status", "", "Filter by status (OPEN, IN_PROGRESS, FIXED, etc.)")
	bugListCmd.Flags().String("severity", "", "Filter by severity (S1, S2, S3, S4)")
	bugListCmd.Flags().String("priority", "", "Filter by priority (P0, P1, P2, P3)")
	bugListCmd.Flags().Bool("json", false, "Output in JSON format")
	bugListCmd.Flags().Bool("no-color", false, "Disable colored output")
	bugListCmd.Flags().Int("limit", 0, "Limit number of results (0 = unlimited)")
	bugListCmd.Flags().String("db", ".canary/canary.db", "Path to database file")

	// Create command flags
	bugCreateCmd.Flags().String("aspect", "", "Bug aspect (API, CLI, Engine, Storage, etc.)")
	bugCreateCmd.Flags().String("severity", "S3", "Severity level (S1-Critical, S2-High, S3-Medium, S4-Low)")
	bugCreateCmd.Flags().String("priority", "P2", "Priority level (P0, P1, P2, P3)")
	bugCreateCmd.Flags().String("status", "OPEN", "Initial status")
	bugCreateCmd.Flags().String("file", "", "File and line number (e.g., src/api/handler.go:42)")
	bugCreateCmd.Flags().String("owner", "", "Bug owner/assignee")
	bugCreateCmd.Flags().String("db", ".canary/canary.db", "Path to database file")

	// Update command flags
	bugUpdateCmd.Flags().String("status", "", "New status (OPEN, IN_PROGRESS, FIXED, etc.)")
	bugUpdateCmd.Flags().String("severity", "", "New severity (S1, S2, S3, S4)")
	bugUpdateCmd.Flags().String("priority", "", "New priority (P0, P1, P2, P3)")
	bugUpdateCmd.Flags().String("owner", "", "New owner/assignee")
	bugUpdateCmd.Flags().String("db", ".canary/canary.db", "Path to database file")

	// Show command flags
	bugShowCmd.Flags().Bool("json", false, "Output in JSON format")
	bugShowCmd.Flags().String("db", ".canary/canary.db", "Path to database file")
}
