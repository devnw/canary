package index

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.devnw.com/canary/internal/storage"
	"go.devnw.com/canary/cli/internal/utils"
)

// CANARY: REQ=CBIN-124; FEATURE="IndexCmd"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
var IndexCmd = &cobra.Command{
	Use:   "index [flags]",
	Short: "Build or rebuild the CANARY token database",
	Long: `Scan the codebase for CANARY tokens and store metadata in SQLite database.

This enables advanced features like priority ordering, keyword search, and checkpoints.
The database is stored at .canary/canary.db by default.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement --prompt flag to load custom prompts
		prompt, _ := cmd.Flags().GetString("prompt")
		_ = prompt // Stubbed for future use

		dbPath, _ := cmd.Flags().GetString("db")
		rootPath, _ := cmd.Flags().GetString("root")

		fmt.Printf("Indexing CANARY tokens from: %s\n", rootPath)

		// Open or create database
		db, err := storage.Open(dbPath)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}

		defer db.Close()

		// Get git info if in a repo
		var commitHash, branch string
		if gitCmd := exec.Command("git", "rev-parse", "HEAD"); gitCmd.Dir == "" {
			if output, err := gitCmd.Output(); err == nil {
				commitHash = strings.TrimSpace(string(output))
			}
		}
		if gitCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD"); gitCmd.Dir == "" {
			if output, err := gitCmd.Output(); err == nil {
				branch = strings.TrimSpace(string(output))
			}
		}

		// Scan for all CANARY tokens
		grepCmd := exec.Command("grep",
			"-rn",
			"--include=*.go", "--include=*.md", "--include=*.py",
			"--include=*.js", "--include=*.ts", "--include=*.java",
			"--include=*.rb", "--include=*.rs", "--include=*.c",
			"--include=*.cpp", "--include=*.h", "--include=*.sql",
			"CANARY:",
			rootPath,
		)

		output, err := grepCmd.CombinedOutput()
		if err != nil && len(output) == 0 {
			fmt.Println("No CANARY tokens found")
			return nil
		}

		// Parse and store tokens
		indexed := 0
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}

			// Parse grep output: file:line:content
			parts := strings.SplitN(line, ":", 3)
			if len(parts) < 3 {
				continue
			}

			file := parts[0]
			lineNum := 0
			//nolint:errcheck // Best-effort parse, default to 0 on failure
			fmt.Sscanf(parts[1], "%d", &lineNum)
			content := parts[2]

			// Extract all CANARY fields
			reqID := utils.ExtractField(content, "REQ")
			feature := utils.ExtractField(content, "FEATURE")
			aspect := utils.ExtractField(content, "ASPECT")
			status := utils.ExtractField(content, "STATUS")

			if reqID == "" || feature == "" {
				continue // Skip malformed tokens
			}

			// Build token struct
			docPath := utils.ExtractField(content, "DOC")
			docType := utils.ExtractField(content, "DOC_TYPE")

			// Auto-infer DOC_TYPE from type prefix if not explicitly set
			if docPath != "" && docType == "" {
				// Extract type from first doc path (e.g., "user:docs/file.md" -> "user")
				firstPath := strings.Split(docPath, ",")[0]
				if strings.Contains(firstPath, ":") {
					docType = strings.Split(firstPath, ":")[0]
				}
			}

			token := &storage.Token{
				ReqID:       reqID,
				Feature:     feature,
				Aspect:      aspect,
				Status:      status,
				FilePath:    file,
				LineNumber:  lineNum,
				Test:        utils.ExtractField(content, "TEST"),
				Bench:       utils.ExtractField(content, "BENCH"),
				Owner:       utils.ExtractField(content, "OWNER"),
				Phase:       utils.ExtractField(content, "PHASE"),
				Keywords:    utils.ExtractField(content, "KEYWORDS"),
				SpecStatus:  utils.ExtractField(content, "SPEC_STATUS"),
				UpdatedAt:   utils.ExtractField(content, "UPDATED"),
				CreatedAt:   utils.ExtractField(content, "CREATED"),
				StartedAt:   utils.ExtractField(content, "STARTED"),
				CompletedAt: utils.ExtractField(content, "COMPLETED"),
				CommitHash:  commitHash,
				Branch:      branch,
				DependsOn:   utils.ExtractField(content, "DEPENDS_ON"),
				Blocks:      utils.ExtractField(content, "BLOCKS"),
				RelatedTo:   utils.ExtractField(content, "RELATED_TO"),
				DocPath:     docPath,
				DocHash:     utils.ExtractField(content, "DOC_HASH"),
				DocType:     docType,
				RawToken:    content,
				IndexedAt:   time.Now().UTC().Format(time.RFC3339),
			}

			// Parse priority
			if priorityStr := utils.ExtractField(content, "PRIORITY"); priorityStr != "" {
				if p, err := strconv.Atoi(priorityStr); err == nil {
					token.Priority = p
				} else {
					token.Priority = 5 // default
				}
			} else {
				token.Priority = 5 // default
			}

			// Set defaults
			if token.UpdatedAt == "" {
				token.UpdatedAt = time.Now().UTC().Format("2006-01-02")
			}
			if token.SpecStatus == "" {
				token.SpecStatus = "draft"
			}

			// Store in database
			if err := db.UpsertToken(token); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to store token %s/%s: %v\n", reqID, feature, err)
				continue
			}

			indexed++
		}

		fmt.Printf("\n✅ Indexed %d CANARY tokens\n", indexed)
		fmt.Printf("Database: %s\n", dbPath)

		if commitHash != "" {
			fmt.Printf("Commit: %s\n", commitHash[:8])
		}
		if branch != "" {
			fmt.Printf("Branch: %s\n", branch)
		}

		return nil
	},
}

func init() {
	IndexCmd.Flags().String("prompt", "", "Custom prompt file or embedded prompt name (future use)")
	IndexCmd.Flags().String("db", ".canary/canary.db", "path to database file")
	IndexCmd.Flags().String("root", ".", "root directory to scan")
}
