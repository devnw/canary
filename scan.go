// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package canary

import (
	"encoding/csv"
	"os"
	"sort"
	"time"

	"go.devnw.com/canary/gate"
)

// Scan now delegates to the generic gate.Scanner, converting the result into the legacy report.
// Backward-compatible: existing tests and callers expect auto-promotion logic preserved.
func Scan(root string) (Report, error) {
    sc := gate.NewScanner() // default with promotion
    res, err := sc.ScanRepository(root)
    if err != nil {
        return Report{}, err
    }
    // Convert gate result to legacy report shape
    var rows []requirementRow
    for _, r := range res.Requirements {
        feats := make([]featureEntry, len(r.Features))
        for i, f := range r.Features {
            feats[i] = featureEntry{
                Feature: f.Feature,
                Aspect:  f.Aspect,
                Status:  f.Status,
                Files:   append([]string(nil), f.Files...),
                Tests:   append([]string(nil), f.Tests...),
                Benches: append([]string(nil), f.Benches...),
                Owner:   f.Owner,
                Updated: f.Updated,
            }
        }
        // Already sorted by gate; defensive sort for determinism
        sort.Slice(feats, func(i, j int) bool {
            if feats[i].Feature == feats[j].Feature {
                return feats[i].Aspect < feats[j].Aspect
            }
            return feats[i].Feature < feats[j].Feature
        })
        rows = append(rows, requirementRow{ID: r.ID, Features: feats})
    }
    sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
    return Report{Requirements: rows, Summary: summary{ByStatus: res.Summary.ByStatus, ByAspect: res.Summary.ByAspect}}, nil
}

// CheckStaleness delegates to gate.CheckStalenessTokens for consistency.
func CheckStaleness(rep Report, dur time.Duration) error {
    // Convert report -> ScanResult minimal for staleness check
    var reqs []gate.RequirementTokens
    for _, r := range rep.Requirements {
        var feats []gate.FeatureTokens
        for _, f := range r.Features {
            feats = append(feats, gate.FeatureTokens{Feature: f.Feature, Aspect: f.Aspect, Status: f.Status, Files: f.Files, Tests: f.Tests, Benches: f.Benches, Owner: f.Owner, Updated: f.Updated})
        }
        reqs = append(reqs, gate.RequirementTokens{ID: r.ID, Features: feats})
    }
    sr := gate.ScanResult{Requirements: reqs, Summary: gate.SummaryTokens{ByStatus: rep.Summary.ByStatus, ByAspect: rep.Summary.ByAspect}}
    return gate.CheckStalenessTokens(sr, int64(dur/time.Second))
}

// WriteCSV unchanged; moved here after refactor for legacy consumers.
func WriteCSV(rep Report, path string) error {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()
    w := csv.NewWriter(f)
    defer w.Flush()
    _ = w.Write([]string{"req", "feature", "aspect", "status", "file", "test", "bench", "owner", "updated"})
    for _, r := range rep.Requirements {
        for _, f := range r.Features {
            rowBase := []string{r.ID, f.Feature, f.Aspect, f.Status, "", "", "", f.Owner, f.Updated}
            max := max3(len(f.Files), len(f.Tests), len(f.Benches))
            for i := 0; i < max; i++ {
                row := make([]string, len(rowBase))
                copy(row, rowBase)
                if i < len(f.Files) {
                    row[4] = f.Files[i]
                }
                if i < len(f.Tests) {
                    row[5] = f.Tests[i]
                }
                if i < len(f.Benches) {
                    row[6] = f.Benches[i]
                }
                if err := w.Write(row); err != nil {
                    return err
                }
            }
            if max == 0 {
                if err := w.Write(rowBase); err != nil {
                    return err
                }
            }
        }
    }
    return nil
}

func max3(a, b, c int) int { if a < b { a = b }; if a < c { a = c }; return a }
