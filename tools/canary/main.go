// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package main

// CANARY: REQ=CBIN-101; FEATURE="ScannerCore"; ASPECT=Engine; STATUS=BENCHED; TEST=TestCANARY_CBIN_101_Engine_ScanBasic; BENCH=BenchmarkCANARY_CBIN_101_Engine_Scan; OWNER=canary; UPDATED=2025-10-15

import (
	"flag"
	"os"

	"devnw.dev/canary/pkg/canaryscan"
)

func main() {
	root := flag.String("root", ".", "root directory to scan")
	outJSON := flag.String("out", "status.json", "output status.json path")
	outCSV := flag.String("csv", "", "optional status.csv path")
	verifyPath := flag.String("verify", "", "GAP_ANALYSIS file to verify claims")
	strict := flag.Bool("strict", false, "enforce staleness on TESTED/BENCHED (30d)")
	skipExpr := flag.String("skip", canaryscan.DefaultSkipRegex().String(), "skip path regex (RE2)")
	updateStale := flag.Bool("update-stale", false, "rewrite UPDATED field for stale TESTED/BENCHED tokens")
	projectOnly := flag.Bool("project-only", false, "filter by project requirement ID pattern from .canary/project.yaml")
	flag.Parse()

	code := canaryscan.RunFromArgs(*root, *outJSON, *outCSV, *verifyPath, *skipExpr, *strict, *updateStale, *projectOnly)
	os.Exit(code)
}
