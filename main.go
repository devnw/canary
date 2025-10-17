// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.


package main

// CANARY: REQ=CBIN-101; FEATURE="ScannerCore"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_101_Engine_ScanBasic; BENCH=BenchmarkCANARY_CBIN_101_Engine_Scan; OWNER=canary; UPDATED=2025-09-20
import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

type report struct {
	GeneratedAt  time.Time        `json:"generated_at"`
	Requirements []requirementRow `json:"requirements"`
	Summary      summary          `json:"summary"`
}

type requirementRow struct {
	ID       string         `json:"id"`
	Features []featureEntry `json:"features"`
}

type featureEntry struct {
	Feature string   `json:"feature"`
	Aspect  string   `json:"aspect"`
	Status  string   `json:"status"`
	Files   []string `json:"files"`
	Tests   []string `json:"tests"`
	Benches []string `json:"benches"`
	Owner   string   `json:"owner,omitempty"`
	Updated string   `json:"updated,omitempty"`
}

type summary struct {
	ByStatus map[string]int `json:"by_status"`
	ByAspect map[string]int `json:"by_aspect"`
}

func main() {
	var root, out, csv, verify string
	var strict bool
	flag.StringVar(&root, "root", ".", "repository root to scan")
	flag.StringVar(&out, "out", "status.json", "output JSON file")
	flag.StringVar(&csv, "csv", "", "optional CSV output file")
	flag.StringVar(&verify, "verify", "", "verify claims in GAP_ANALYSIS.md")
	flag.BoolVar(&strict, "strict", false, "strict mode: fail on stale UPDATED (>30 days) for TESTED/BENCHED")
	flag.Parse()

	rep, err := Scan(root)
	if err != nil {
		log.Printf("ERROR scan: %v", err)
		os.Exit(3)
	}
	rep.GeneratedAt = time.Now().UTC()

	// Strict staleness check
	if strict {
		if err := CheckStaleness(rep, 30*24*time.Hour); err != nil {
			log.Printf("CANARY_STALE %v", err)
			// still write outputs for inspection
			//nolint:errcheck // Error doesn't matter here, we're exiting anyway
			writeOutputs(rep, out, csv)
			os.Exit(2)
		}
	}

	// Verify GAP/claims
	if verify != "" {
		claims, err := ParseGAPClaims(verify)
		if err != nil {
			log.Printf("ERROR verify-parse: %v", err)
			//nolint:errcheck // Error doesn't matter here, we're exiting anyway
			writeOutputs(rep, out, csv)
			os.Exit(3)
		}
		if err := VerifyClaims(rep, claims); err != nil {
			log.Printf("CANARY_VERIFY_FAIL %v", err)
			//nolint:errcheck // Error doesn't matter here, we're exiting anyway
			writeOutputs(rep, out, csv)
			os.Exit(2)
		}
	}

	if err := writeOutputs(rep, out, csv); err != nil {
		log.Printf("ERROR write: %v", err)
		os.Exit(3)
	}
}

func writeOutputs(rep report, out, csv string) error {
	// JSON
	jf, err := os.Create(out)
	if err != nil {
		return err
	}

	defer jf.Close()
	enc := json.NewEncoder(jf)
	enc.SetEscapeHTML(false)
	// Minified JSON output (no indentation)
	if err := enc.Encode(rep); err != nil {
		return err
	}

	// CSV optional
	if csv != "" {
		if err := WriteCSV(rep, csv); err != nil {
			return err
		}
	}
	fmt.Printf("CANARY_OK wrote %s\n", out)
	return nil
}
