package canary

import "time"

// Report is the canonical legacy output structure retained for backward compatibility.
type Report struct {
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

// fromScanResult converts a gate.ScanResult into a legacy report (without GeneratedAt).
// GeneratedAt should be set by caller (e.g., Run). Key ordering and promotion already handled upstream.
func fromScanResult(res interface{ GetRequirements() []requirementRow; GetSummary() summary }) Report {
	// This adapter interface allows future decoupling; currently unused externally.
	return Report{Requirements: res.GetRequirements(), Summary: res.GetSummary()}
}

