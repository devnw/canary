package canary

import "time"

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
