// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package grep

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/storage"
)

// CANARY: REQ=CBIN-CLI-001; FEATURE="GrepCmd"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_CLI_001_CLI_GrepCmd; UPDATED=2025-10-16
var GrepCmd = &cobra.Command{
	Use:   "grep <pattern>",
	Short: "Search CANARY tokens by pattern",
	Long: `Search for CANARY tokens matching a pattern.

Searches across:
- Feature names
- File paths
- Test names
- Bench names
- Requirement IDs

The search is case-insensitive and matches substrings.

Examples:
  canary grep User              # Find all tokens related to "User"
  canary grep internal/auth     # Find tokens in auth directory
  canary grep TestAuth          # Find tokens with "TestAuth" test`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := args[0]
		dbPath, _ := cmd.Flags().GetString("db")
		groupBy, _ := cmd.Flags().GetString("group-by")

		// Open database
		db, err := storage.Open(dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Database not found\n")
			fmt.Fprintf(os.Stderr, "   Suggestion: Run 'canary index' to build database\n\n")
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		// Search for matching tokens
		tokens, err := grepTokens(db, pattern)
		if err != nil {
			return fmt.Errorf("search tokens: %w", err)
		}

		if len(tokens) == 0 {
			fmt.Printf("No tokens found matching pattern: %s\n", pattern)
			return nil
		}

		// Display results
		fmt.Printf("Found %d tokens matching '%s':\n\n", len(tokens), pattern)

		if groupBy == "requirement" {
			displayGrepResultsByRequirement(tokens)
		} else {
			displayGrepResults(tokens)
		}

		return nil
	},
}

// grepTokens searches for tokens matching the pattern
func grepTokens(db *storage.DB, pattern string) ([]*storage.Token, error) {
	if pattern == "" {
		return []*storage.Token{}, nil
	}

	// Get all tokens and filter by pattern
	// We use SearchTokens which already does keyword matching
	tokens, err := db.SearchTokens(pattern)
	if err != nil {
		return nil, err
	}

	// Additional filtering for file paths and test names
	allTokens, err := db.ListTokens(nil, "", "", 0)
	if err != nil {
		return nil, err
	}

	patternLower := strings.ToLower(pattern)
	matchMap := make(map[string]*storage.Token)

	// Add tokens from keyword search
	for _, token := range tokens {
		key := fmt.Sprintf("%s:%s:%s:%d", token.ReqID, token.Feature, token.FilePath, token.LineNumber)
		matchMap[key] = token
	}

	// Add tokens matching file path, test, or bench
	for _, token := range allTokens {
		if strings.Contains(strings.ToLower(token.FilePath), patternLower) ||
			strings.Contains(strings.ToLower(token.Test), patternLower) ||
			strings.Contains(strings.ToLower(token.Bench), patternLower) {
			key := fmt.Sprintf("%s:%s:%s:%d", token.ReqID, token.Feature, token.FilePath, token.LineNumber)
			matchMap[key] = token
		}
	}

	// Convert map back to slice
	result := make([]*storage.Token, 0, len(matchMap))
	for _, token := range matchMap {
		result = append(result, token)
	}

	return result, nil
}

// displayGrepResults shows grep results in a simple list format
func displayGrepResults(tokens []*storage.Token) {
	for _, token := range tokens {
		fmt.Printf("📌 %s - %s\n", token.ReqID, token.Feature)
		fmt.Printf("   Status: %s | Aspect: %s\n", token.Status, token.Aspect)
		fmt.Printf("   Location: %s:%d\n", token.FilePath, token.LineNumber)
		if token.Test != "" {
			fmt.Printf("   Test: %s\n", token.Test)
		}
		if token.Bench != "" {
			fmt.Printf("   Bench: %s\n", token.Bench)
		}
		fmt.Println()
	}
}

// displayGrepResultsByRequirement groups results by requirement ID
func displayGrepResultsByRequirement(tokens []*storage.Token) {
	// Group by requirement
	reqMap := make(map[string][]*storage.Token)
	for _, token := range tokens {
		reqMap[token.ReqID] = append(reqMap[token.ReqID], token)
	}

	// Display grouped results
	for reqID, reqTokens := range reqMap {
		fmt.Printf("## %s (%d tokens)\n\n", reqID, len(reqTokens))
		for _, token := range reqTokens {
			fmt.Printf("  📌 %s\n", token.Feature)
			fmt.Printf("     Status: %s | Aspect: %s | %s:%d\n",
				token.Status, token.Aspect, token.FilePath, token.LineNumber)
			if token.Test != "" {
				fmt.Printf("     Test: %s\n", token.Test)
			}
		}
		fmt.Println()
	}
}

func init() {
	GrepCmd.Flags().String("db", ".canary/canary.db", "Path to database file")
	GrepCmd.Flags().String("group-by", "none", "Group results (none, requirement)")
}
