// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package main

// CANARY: REQ=CP-235; FEATURE="ScannerCore"; ASPECT=Engine; STATUS=BENCHED; TEST=TestCANARY_CBIN_101_Engine_ScanBasic; BENCH=BenchmarkCANARY_CBIN_101_Engine_Scan; OWNER=canary; UPDATED=2026-08-29

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
	strict := flag.Bool("strict", false, "enforce staleness on TESTED/BENCHED tokens (default 30d; see --stale-days)")
	skipExpr := flag.String("skip", canaryscan.DefaultSkipRegex().String(), "skip path regex (RE2)")
	updateStale := flag.Bool("update-stale", false, "rewrite UPDATED field for stale TESTED/BENCHED tokens")
	projectOnly := flag.Bool("project-only", false, "filter by project requirement ID pattern from .canary/project.yaml")
	staleDays := flag.Int("stale-days", 0, "staleness window in days (0 = use .canary/project.yaml verification.staleness_days, else 30)")
	flag.Parse()

	code := canaryscan.RunFromArgs(*root, *outJSON, *outCSV, *verifyPath, *skipExpr, *strict, *updateStale, *projectOnly, *staleDays)
	os.Exit(code)
}
