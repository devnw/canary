package canaryscan

// Scanner and token-grammar limits. These bound what a single scan will read
// and what a single token field may hold, so a hostile or accidental input
// (a multi-gigabyte artifact, a minified one-line bundle) cannot exhaust
// memory or wedge a scan.
const (
	MaxFileBytes  = 16 << 20 // 16 MiB per file
	MaxLineBytes  = 1 << 20  // 1 MiB per logical line
	MaxFieldBytes = 1024     // per token field value, UTF-8 bytes
	// ParserSchemaVersion identifies the token grammar/semantics. Bumped to
	// 2 when STATUS became a pure declaration (no lexical promotion).
	ParserSchemaVersion = 2
)

// ScanIssue records a file the scanner could not fully process. Issues never
// abort a scan: the file is skipped and the rest of the tree still scans.
type ScanIssue struct {
	Path   string `json:"path"`
	Reason string `json:"reason"` // "binary" | "file_too_large" | "line_too_large" | "read_error" | "parse_error"
	Detail string `json:"detail,omitempty"`
}

// Scan issue reasons.
const (
	IssueBinary       = "binary"
	IssueFileTooLarge = "file_too_large"
	IssueLineTooLarge = "line_too_large"
	IssueReadError    = "read_error"
	IssueParseError   = "parse_error"
)

// Report is the scan output (JSON/CSV shape). Fields are only ever added:
// status.json is a cross-package wire contract.
type Report struct {
	GeneratedAt    string        `json:"generated_at"`
	Requirements   []Requirement `json:"requirements"`
	Summary        Summary       `json:"summary"`
	MigrationNotes []MigrateNote `json:"migration_notes,omitempty"`
	Issues         []ScanIssue   `json:"issues,omitempty"`
}

// Requirement groups features by requirement ID.
type Requirement struct {
	ID        string    `json:"id"`
	Source    string    `json:"source,omitempty"`
	TicketURL string    `json:"ticket_url,omitempty"`
	Diagrams  []string  `json:"diagrams,omitempty"` // filled by Task 4
	Features  []Feature `json:"features"`
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
