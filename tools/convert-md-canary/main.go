package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	count := 0
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories we don't want to process
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process markdown files
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Read file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		modified := false

		for scanner.Scan() {
			line := scanner.Text()
			// Check if line starts with # CANARY:
			if strings.HasPrefix(line, "# CANARY:") {
				// Convert to HTML comment
				line = strings.Replace(line, "# CANARY:", "<!-- CANARY:", 1)
				line = line + " -->"
				modified = true
			}
			lines = append(lines, line)
		}

		if err := scanner.Err(); err != nil {
			return err
		}
		file.Close()

		// Write back if modified
		if modified {
			output := strings.Join(lines, "\n") + "\n"
			if err := os.WriteFile(path, []byte(output), 0644); err != nil {
				return err
			}
			count++
			fmt.Printf("Converted: %s\n", path)
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Converted %d files\n", count)
}
