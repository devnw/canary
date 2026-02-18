package canaryscan

// Report is the scan output (JSON/CSV shape).
type Report struct {
	GeneratedAt  string        `json:"generated_at"`
	Requirements []Requirement `json:"requirements"`
	Summary      Summary       `json:"summary"`
}

// Requirement groups features by requirement ID.
type Requirement struct {
	ID       string    `json:"id"`
	Features []Feature `json:"features"`
}

// Feature is one feature/aspect/status aggregation.
type Feature struct {
	Feature string   `json:"feature"`
	Aspect  string   `json:"aspect"`
	Status  string   `json:"status"`
	Files   []string `json:"files"`
	Tests   []string `json:"tests"`
	Benches []string `json:"benches"`
	Owner   string   `json:"owner,omitempty"`
	Updated string   `json:"updated"`
}

// Summary holds aggregate counts.
type Summary struct {
	ByStatus           map[string]int `json:"by_status"`
	ByAspect           map[string]int `json:"by_aspect"`
	TotalTokens        int            `json:"total_tokens"`
	UniqueRequirements int            `json:"unique_requirements"`
}
