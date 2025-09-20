package main

import (
	"fmt"
	"path/filepath"
	"regexp"

	core "go.codepros.org/canary/internal/core"
)

func main() {
	re := regexp.MustCompile(core.DefaultSkipPattern)
	rep, _ := core.Scan(core.ScanOptions{Root: filepath.Clean("."), Skip: re})
	fmt.Printf("REQS=%d TOKENS=%d\n", len(rep.Requirements), rep.Summary.TotalTokens)
	for _, r := range rep.Requirements {
		fmt.Println(r.ID, len(r.Features))
	}
}
