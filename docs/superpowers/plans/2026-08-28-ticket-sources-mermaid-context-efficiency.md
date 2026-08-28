# Ticket Sources, Mermaid References & Context Efficiency Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let CANARY requirement IDs come from external ticket systems (JIRA/GitLab/GitHub) as well as flatfile prefixes, make requirement references first-class inside mermaid diagrams (both extraction and generation), add a single `canary view <REQ>` full-picture command, and cap the unbounded CLI/MCP outputs so agents get complete context without grepping.

**Architecture:** A new `internal/sources` package turns `.canary/project.yaml` `sources:` config into a `Registry` (ID pattern, source resolution, ticket-URL expansion, normalization). Every place that currently hardcodes `CBIN` (`internal/canaryscan` claim/normalize regexes, `internal/storage.ListTokens` SQL, `internal/specs` dependency parser, stale-update regex) is rewired to the registry or made prefix-agnostic. The scanner learns to extract requirement references from ```` ```mermaid ```` fences in markdown; `internal/specs.GraphGenerator` learns to render mermaid. A `refs` DB table (migration 000006) persists diagram references so `canary view` and MCP can answer "everything about REQ X" in one bounded call.

**Tech Stack:** Go 1.27, cobra, modernc.org/sqlite + sqlx + golang-migrate (embedded iofs migrations), gopkg.in/yaml.v3, github.com/modelcontextprotocol/go-sdk v1.4.1.

## Global Constraints

- Go version: `go 1.27.0` (go.mod) — do not change.
- Backwards compatibility is a hard requirement: a repo with no `sources:` in project.yaml must behave exactly as today (CBIN flatfile). All existing tests must keep passing (`go test ./...`).
- The stdout contracts are frozen: `CANARY_SCAN tokens=N requirements=M STUB=... IMPL=... TESTED=... BENCHED=...`, `CANARY_VERIFY_OK`, `CANARY_VERIFY_FAIL count=N`, exit codes 0/2/3.
- Every new exported feature gets a CANARY token using these reserved IDs: CBIN-201 TicketSources, CBIN-202 MermaidRefs, CBIN-203 MermaidGraph, CBIN-204 RequirementView, CBIN-205 ContextCaps, CBIN-206 DiagramRefsIndex. Token format: `// CANARY: REQ=CBIN-2XX; FEATURE="Name"; ASPECT=<aspect>; STATUS=IMPL; TEST=<TestName>; UPDATED=2026-08-28` (add `TEST=` and bump STATUS→TESTED once the test exists).
- New files start with the repo license header (copy the 4-line header from `internal/config/config.go:1-4`).
- Commit style: conventional commits (`feat:`, `fix:`, `test:`, `docs:`) ending with the Claude co-author trailer used by this harness.
- `gofmt` clean; run `go vet ./...` before each commit.
- External ticket IDs are NEVER zero-padded or rewritten — `PROJ-42` stays `PROJ-42`. Only flatfile-source IDs (and legacy CBIN/REQ/TASK/BUG) get 3-digit padding.
- **Small-by-default output** (user directive): every command/tool that lists things defaults to a SMALL bound (≤25 items, deliberately smaller than "probably enough") and exposes an explicit `--limit` flag / `limit` param the agent can raise when it needs more. When output is truncated, always print a `… +N more (use --limit)` style hint so the agent knows more exists and how to get it. Never default to unlimited.
- Do not fix the 5 MCP stub tools (specify/plan/index/bug-create/gap-mark) — out of scope. Do not restructure the 5-way duplicated agent docs — out of scope except where a task says otherwise.

---

## Background you need (read once)

Current hardcoded-CBIN sites this plan removes or generalizes:

| Site | File | What it does today |
|---|---|---|
| Claim regex | `internal/canaryscan/parse.go:15` | `claimRe = regexp.MustCompile("(?m)^\\s*✅\\s+(CBIN-\\d{3})\\b")` — verify only recognizes CBIN claims |
| ID padding | `internal/canaryscan/parse.go:104-124` | `normalizeREQ` pads `CBIN-`, `REQ-`, `TASK-`, `BUG-` prefixes to 3 digits |
| Stale updater | `internal/canaryscan/update.go:16` | `REQ=([A-Z]+-\d{3})` — 3-digit only |
| SQL GLOBs | `internal/storage/storage.go:253-320` (`ListTokens`) | Injects `req_id GLOB 'CBIN-[0-9][0-9][0-9]*'`-style filters; the `idPattern` argument is accepted and ignored |
| Dependency parser | `internal/specs/parser_dependency.go:19-20` | `^-\s+(CBIN-\d+)` — non-CBIN projects silently parse zero dependencies |

Key structs (verbatim, current):

```go
// internal/canaryscan/types.go
type Report struct {
	GeneratedAt  string        `json:"generated_at"`
	Requirements []Requirement `json:"requirements"`
	Summary      Summary       `json:"summary"`
}
type Requirement struct {
	ID       string    `json:"id"`
	Features []Feature `json:"features"`
}
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
```

`internal/canaryscan/run.go:22 Run(cfg Config, stdout, stderr io.Writer) int` is the scan pipeline; `run.go:102` calls `VerifyClaims(rep, cfg.VerifyPath)`. `internal/canaryscan` has NO test files — its tests live in `tools/canary/internal/acceptance_test.go` and root `scan_test.go`. New canaryscan behavior gets tested in-package (create `internal/canaryscan/*_test.go`).

`.canary/project.yaml` is loaded by `internal/config.Load(rootDir)` (`internal/config/config.go:40`) and separately by `canaryscan.LoadProjectConfig` (a local mini-copy in `internal/canaryscan/scan.go:17-23`).

---

### Task 1: Sources config schema (`internal/config`)

**Files:**
- Modify: `internal/config/config.go`
- Test: `internal/config/config_test.go` (create)

**Interfaces:**
- Consumes: nothing new.
- Produces: `config.SourceConfig{Name, Type, Key, URL string}`, `ProjectConfig.Sources []SourceConfig`, `ProjectConfig.Project.Key string` — later tasks read these.

- [ ] **Step 1: Write the failing test**

Create `internal/config/config_test.go`:

```go
// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeProjectYAML(t *testing.T, content string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".canary")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "project.yaml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCANARY_CBIN_201_LoadSources(t *testing.T) {
	root := writeProjectYAML(t, `
project:
  name: "demo"
  key: "CBIN"
sources:
  - name: core
    type: flatfile
    key: "CBIN"
  - name: platform
    type: jira
    key: "PLAT"
    url: "https://company.atlassian.net/browse/{id}"
  - name: app
    type: gitlab
    key: "GL"
    url: "https://gitlab.com/devnw/app/-/issues/{num}"
  - name: oss
    type: github
    key: "GH"
    url: "https://github.com/devnw/app/issues/{num}"
`)
	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Project.Key != "CBIN" {
		t.Errorf("Project.Key = %q, want CBIN", cfg.Project.Key)
	}
	if len(cfg.Sources) != 4 {
		t.Fatalf("len(Sources) = %d, want 4", len(cfg.Sources))
	}
	want := SourceConfig{Name: "platform", Type: "jira", Key: "PLAT", URL: "https://company.atlassian.net/browse/{id}"}
	if cfg.Sources[1] != want {
		t.Errorf("Sources[1] = %+v, want %+v", cfg.Sources[1], want)
	}
}

func TestCANARY_CBIN_201_LoadSources_AbsentIsEmpty(t *testing.T) {
	root := writeProjectYAML(t, "project:\n  name: demo\n")
	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Sources) != 0 {
		t.Errorf("Sources should be empty when absent, got %d", len(cfg.Sources))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -run TestCANARY_CBIN_201 -v`
Expected: FAIL — compile error `cfg.Project.Key undefined` / `undefined: SourceConfig`.

- [ ] **Step 3: Write minimal implementation**

In `internal/config/config.go`, add `Key` to the Project struct, and the `SourceConfig` type + `Sources` field:

```go
// SourceConfig describes one requirement-ID source: a flatfile prefix or an
// external ticket system (jira, github, gitlab) whose keys appear in REQ= fields.
// CANARY: REQ=CBIN-201; FEATURE="TicketSources"; ASPECT=Storage; STATUS=IMPL; TEST=TestCANARY_CBIN_201_LoadSources; UPDATED=2026-08-28
type SourceConfig struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // flatfile | jira | github | gitlab
	Key  string `yaml:"key"`  // ID prefix, e.g. "CBIN", "PLAT", "GH"
	URL  string `yaml:"url,omitempty"`
}
```

```go
type ProjectConfig struct {
	Project struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Key         string `yaml:"key"`
	} `yaml:"project"`
	Sources      []SourceConfig `yaml:"sources"`
	Requirements struct {
		IDPattern string `yaml:"id_pattern"`
	} `yaml:"requirements"`
	// ... keep the existing Scanner / Verification / Agent fields unchanged
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/config/ -v`
Expected: PASS (both tests).

- [ ] **Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): parse sources and project.key from project.yaml"
```

---

### Task 2: Source registry package (`internal/sources`)

**Files:**
- Create: `internal/sources/sources.go`
- Test: `internal/sources/sources_test.go`

**Interfaces:**
- Consumes: `config.SourceConfig`, `config.Load` from Task 1.
- Produces (used by Tasks 3, 4, 6, 9, 10, 12, 13):
  - `type Source struct { Name, Type, Key, URL string }`
  - `func NewRegistry(list []Source) (*Registry, error)`
  - `func Default() *Registry` — CBIN flatfile fallback
  - `func LoadFromRoot(root string) *Registry` — reads `.canary/project.yaml`, falls back to `Default()` on any error/absence
  - `func (r *Registry) Pattern() *regexp.Regexp` — matches any configured ID, e.g. `\b((?:CBIN|PLAT|GL)-\d+)\b`
  - `func (r *Registry) ClaimPattern() *regexp.Regexp` — `(?m)^\s*✅\s+((?:CBIN|PLAT|GL)-\d+)\b`
  - `func (r *Registry) Resolve(id string) (Source, bool)`
  - `func (r *Registry) TicketURL(id string) string` — expands `{id}` (full ID) and `{num}` (numeric part); `""` when none
  - `func (r *Registry) Normalize(id string) string` — pads flatfile IDs to 3 digits, leaves ticket IDs verbatim
  - `func (r *Registry) Sources() []Source`

- [ ] **Step 1: Write the failing test**

Create `internal/sources/sources_test.go`:

```go
// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package sources

import "testing"

func testRegistry(t *testing.T) *Registry {
	t.Helper()
	r, err := NewRegistry([]Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "platform", Type: "jira", Key: "PLAT", URL: "https://company.atlassian.net/browse/{id}"},
		{Name: "app", Type: "gitlab", Key: "GL", URL: "https://gitlab.com/devnw/app/-/issues/{num}"},
		{Name: "oss", Type: "github", Key: "GH", URL: "https://github.com/devnw/app/issues/{num}"},
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	return r
}

func TestCANARY_CBIN_201_RegistryPattern(t *testing.T) {
	r := testRegistry(t)
	for _, id := range []string{"CBIN-105", "PLAT-4521", "GL-88", "GH-7"} {
		if !r.Pattern().MatchString(id) {
			t.Errorf("Pattern should match %s", id)
		}
	}
	if r.Pattern().MatchString("OTHER-123") {
		t.Error("Pattern should not match unconfigured prefix OTHER-123")
	}
}

func TestCANARY_CBIN_201_RegistryResolve(t *testing.T) {
	r := testRegistry(t)
	src, ok := r.Resolve("PLAT-4521")
	if !ok || src.Type != "jira" {
		t.Errorf("Resolve(PLAT-4521) = %+v, %v; want jira source", src, ok)
	}
	if _, ok := r.Resolve("OTHER-1"); ok {
		t.Error("Resolve should fail for unconfigured prefix")
	}
}

func TestCANARY_CBIN_201_TicketURL(t *testing.T) {
	r := testRegistry(t)
	cases := map[string]string{
		"PLAT-4521": "https://company.atlassian.net/browse/PLAT-4521",
		"GL-88":     "https://gitlab.com/devnw/app/-/issues/88",
		"GH-7":      "https://github.com/devnw/app/issues/7",
		"CBIN-105":  "",
		"OTHER-9":   "",
	}
	for id, want := range cases {
		if got := r.TicketURL(id); got != want {
			t.Errorf("TicketURL(%s) = %q, want %q", id, got, want)
		}
	}
}

func TestCANARY_CBIN_201_Normalize(t *testing.T) {
	r := testRegistry(t)
	cases := map[string]string{
		"CBIN-42":   "CBIN-042", // flatfile: padded
		"CBIN-105":  "CBIN-105",
		"PLAT-42":   "PLAT-42", // jira: verbatim, never padded
		"GL-8":      "GL-8",
		"OTHER-7":   "OTHER-7", // unknown: verbatim
	}
	for id, want := range cases {
		if got := r.Normalize(id); got != want {
			t.Errorf("Normalize(%s) = %q, want %q", id, got, want)
		}
	}
}

func TestCANARY_CBIN_201_ClaimPattern(t *testing.T) {
	r := testRegistry(t)
	gap := "✅ CBIN-105\n  ✅ PLAT-4521\n- [ ] GL-88\n✅ OTHER-1\n"
	got := map[string]bool{}
	for _, m := range r.ClaimPattern().FindAllStringSubmatch(gap, -1) {
		got[m[1]] = true
	}
	if !got["CBIN-105"] || !got["PLAT-4521"] {
		t.Errorf("claims missing: %v", got)
	}
	if got["GL-88"] || got["OTHER-1"] {
		t.Errorf("false claims matched: %v", got)
	}
}

func TestCANARY_CBIN_201_DefaultRegistry(t *testing.T) {
	r := Default()
	if !r.Pattern().MatchString("CBIN-105") {
		t.Error("Default registry must match CBIN IDs")
	}
	if got := r.Normalize("CBIN-42"); got != "CBIN-042" {
		t.Errorf("Default Normalize(CBIN-42) = %q, want CBIN-042", got)
	}
}

func TestCANARY_CBIN_201_NewRegistryValidation(t *testing.T) {
	if _, err := NewRegistry([]Source{{Name: "a", Type: "jira", Key: "bad-key"}}); err == nil {
		t.Error("lowercase/hyphen key must be rejected")
	}
	if _, err := NewRegistry([]Source{
		{Name: "a", Type: "flatfile", Key: "X"},
		{Name: "b", Type: "jira", Key: "X"},
	}); err == nil {
		t.Error("duplicate keys must be rejected")
	}
	if _, err := NewRegistry([]Source{{Name: "a", Type: "svn", Key: "X"}}); err == nil {
		t.Error("unknown type must be rejected")
	}
	if _, err := NewRegistry(nil); err == nil {
		t.Error("empty registry must be rejected")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/sources/ -v`
Expected: FAIL — package does not exist / undefined identifiers.

- [ ] **Step 3: Write the implementation**

Create `internal/sources/sources.go`:

```go
// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package sources resolves requirement-ID prefixes to their origin: a local
// flatfile series (e.g. CBIN-105) or an external ticket system (JIRA, GitLab,
// GitHub) configured in .canary/project.yaml under `sources:`.
// CANARY: REQ=CBIN-201; FEATURE="TicketSources"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_201_RegistryPattern; UPDATED=2026-08-28
package sources

import (
	"fmt"
	"regexp"
	"strings"

	"go.devnw.com/canary/internal/config"
)

// Source is one configured requirement-ID origin.
type Source struct {
	Name string
	Type string // flatfile | jira | github | gitlab
	Key  string // uppercase ID prefix, e.g. "CBIN", "PLAT"
	URL  string // optional template; {id} = full ID, {num} = numeric part
}

var validTypes = map[string]struct{}{"flatfile": {}, "jira": {}, "github": {}, "gitlab": {}}
var keyRe = regexp.MustCompile(`^[A-Z][A-Z0-9]*$`)

// Registry answers ID-shaped questions for a set of sources.
type Registry struct {
	list    []Source
	byKey   map[string]Source
	pattern *regexp.Regexp
	claim   *regexp.Regexp
}

// NewRegistry validates the sources and builds the combined ID pattern.
func NewRegistry(list []Source) (*Registry, error) {
	if len(list) == 0 {
		return nil, fmt.Errorf("sources: at least one source required")
	}
	byKey := make(map[string]Source, len(list))
	keys := make([]string, 0, len(list))
	for _, s := range list {
		if _, ok := validTypes[s.Type]; !ok {
			return nil, fmt.Errorf("sources: %q has unknown type %q", s.Name, s.Type)
		}
		if !keyRe.MatchString(s.Key) {
			return nil, fmt.Errorf("sources: %q key %q must be uppercase alphanumeric starting with a letter", s.Name, s.Key)
		}
		if _, dup := byKey[s.Key]; dup {
			return nil, fmt.Errorf("sources: duplicate key %q", s.Key)
		}
		byKey[s.Key] = s
		keys = append(keys, regexp.QuoteMeta(s.Key))
	}
	alt := strings.Join(keys, "|")
	return &Registry{
		list:    list,
		byKey:   byKey,
		pattern: regexp.MustCompile(`\b((?:` + alt + `)-\d+)\b`),
		claim:   regexp.MustCompile(`(?m)^\s*✅\s+((?:` + alt + `)-\d+)\b`),
	}, nil
}

// Default returns the registry used when no sources are configured:
// the historical CBIN flatfile series.
func Default() *Registry {
	r, _ := NewRegistry([]Source{{Name: "default", Type: "flatfile", Key: "CBIN"}})
	return r
}

// FromProjectConfig builds a registry from a parsed project config. When the
// config declares no sources, a flatfile source is synthesized from
// project.key (default "CBIN") so existing projects keep working unchanged.
func FromProjectConfig(cfg *config.ProjectConfig) *Registry {
	if cfg == nil {
		return Default()
	}
	if len(cfg.Sources) == 0 {
		key := strings.TrimSpace(cfg.Project.Key)
		if !keyRe.MatchString(key) {
			return Default()
		}
		r, err := NewRegistry([]Source{{Name: "default", Type: "flatfile", Key: key}})
		if err != nil {
			return Default()
		}
		return r
	}
	list := make([]Source, 0, len(cfg.Sources))
	for _, s := range cfg.Sources {
		list = append(list, Source{Name: s.Name, Type: s.Type, Key: s.Key, URL: s.URL})
	}
	r, err := NewRegistry(list)
	if err != nil {
		return Default()
	}
	return r
}

// LoadFromRoot reads .canary/project.yaml under root. Any error falls back to
// Default() — sources config must never break a scan.
func LoadFromRoot(root string) *Registry {
	cfg, err := config.Load(root)
	if err != nil {
		return Default()
	}
	return FromProjectConfig(cfg)
}

// Pattern matches any configured requirement ID (captured in group 1).
func (r *Registry) Pattern() *regexp.Regexp { return r.pattern }

// ClaimPattern matches "✅ <ID>" gap-analysis claim lines (ID in group 1).
func (r *Registry) ClaimPattern() *regexp.Regexp { return r.claim }

// Sources returns the configured sources in declaration order.
func (r *Registry) Sources() []Source { return r.list }

// Resolve returns the source owning id's prefix.
func (r *Registry) Resolve(id string) (Source, bool) {
	i := strings.LastIndex(id, "-")
	if i <= 0 {
		return Source{}, false
	}
	s, ok := r.byKey[id[:i]]
	return s, ok
}

// TicketURL expands the owning source's URL template. Empty when the source
// is flatfile, unknown, or has no URL configured.
func (r *Registry) TicketURL(id string) string {
	s, ok := r.Resolve(id)
	if !ok || s.URL == "" {
		return ""
	}
	num := id[strings.LastIndex(id, "-")+1:]
	out := strings.ReplaceAll(s.URL, "{id}", id)
	return strings.ReplaceAll(out, "{num}", num)
}

// Normalize zero-pads flatfile IDs to at least 3 digits (CBIN-42 → CBIN-042).
// External ticket IDs and unknown prefixes are returned verbatim.
func (r *Registry) Normalize(id string) string {
	s, ok := r.Resolve(id)
	if !ok || s.Type != "flatfile" {
		return id
	}
	i := strings.LastIndex(id, "-")
	num := id[i+1:]
	for len(num) < 3 {
		num = "0" + num
	}
	return id[:i+1] + num
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/sources/ -v`
Expected: PASS (7 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/sources/
git commit -m "feat(sources): registry for flatfile and ticket-system requirement IDs"
```

---

### Task 3: Registry-driven scanning & verification (`internal/canaryscan`)

**Files:**
- Modify: `internal/canaryscan/parse.go` (claimRe removal, normalizeREQ), `internal/canaryscan/verify.go` (VerifyClaims signature), `internal/canaryscan/run.go` (load registry), `internal/canaryscan/types.go` (Requirement gains Source/TicketURL), `internal/canaryscan/scan.go` (buildReport annotation)
- Modify callers: `mcp/tools.go:452` (handleScan) if it calls `canaryscan.Scan` directly — check with `grep -rn "canaryscan\.\(Scan\|VerifyClaims\)" --include="*.go" .`
- Test: `internal/canaryscan/sources_test.go` (create)

**Interfaces:**
- Consumes: `sources.Registry`, `sources.LoadFromRoot`, `(*Registry).ClaimPattern/Normalize/Resolve/TicketURL` from Task 2.
- Produces:
  - `VerifyClaims(rep Report, gapPath string, reg *sources.Registry) []string` (new param; nil reg → `sources.Default()`)
  - `Requirement` gains `Source string \`json:"source,omitempty"\`` and `TicketURL string \`json:"ticket_url,omitempty"\``
  - `Scan` keeps its signature; annotation happens in `Run` and in a new exported helper `AnnotateSources(rep *Report, reg *sources.Registry)` so direct `Scan` callers can opt in.

- [ ] **Step 1: Write the failing test**

Create `internal/canaryscan/sources_test.go`:

```go
package canaryscan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.devnw.com/canary/internal/sources"
)

func ticketRegistry(t *testing.T) *sources.Registry {
	t.Helper()
	r, err := sources.NewRegistry([]sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "platform", Type: "jira", Key: "PLAT", URL: "https://company.atlassian.net/browse/{id}"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestCANARY_CBIN_201_VerifyClaimsTicketSource(t *testing.T) {
	reg := ticketRegistry(t)
	gap := filepath.Join(t.TempDir(), "GAP_ANALYSIS.md")
	if err := os.WriteFile(gap, []byte("✅ PLAT-4521\n✅ CBIN-105\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	rep := Report{Requirements: []Requirement{
		{ID: "PLAT-4521", Features: []Feature{{Feature: "F", Aspect: "API", Status: "TESTED"}}},
		{ID: "CBIN-105", Features: []Feature{{Feature: "G", Aspect: "API", Status: "IMPL"}}},
	}}
	diags := VerifyClaims(rep, gap, reg)
	if len(diags) != 1 {
		t.Fatalf("diags = %v, want exactly 1 (CBIN-105 overclaim)", diags)
	}
	if !strings.Contains(diags[0], "REQ=CBIN-105") {
		t.Errorf("diag should name CBIN-105: %s", diags[0])
	}
}

func TestCANARY_CBIN_201_VerifyClaimsNilRegistryDefaultsToCBIN(t *testing.T) {
	gap := filepath.Join(t.TempDir(), "GAP.md")
	if err := os.WriteFile(gap, []byte("✅ CBIN-101\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	rep := Report{Requirements: []Requirement{
		{ID: "CBIN-101", Features: []Feature{{Status: "TESTED"}}},
	}}
	if diags := VerifyClaims(rep, gap, nil); len(diags) != 0 {
		t.Errorf("nil registry must keep legacy CBIN behavior, got %v", diags)
	}
}

func TestCANARY_CBIN_201_NormalizeREQTicketIDsVerbatim(t *testing.T) {
	reg := ticketRegistry(t)
	if got := normalizeREQWithRegistry("PLAT-42", reg); got != "PLAT-42" {
		t.Errorf("jira ID must not be padded: got %q", got)
	}
	if got := normalizeREQWithRegistry("CBIN-42", reg); got != "CBIN-042" {
		t.Errorf("flatfile ID must be padded: got %q", got)
	}
	// legacy prefixes keep padding even without registry entry
	if got := normalizeREQWithRegistry("TASK-7", reg); got != "TASK-007" {
		t.Errorf("legacy TASK padding lost: got %q", got)
	}
}

func TestCANARY_CBIN_201_AnnotateSources(t *testing.T) {
	reg := ticketRegistry(t)
	rep := Report{Requirements: []Requirement{
		{ID: "PLAT-4521"}, {ID: "CBIN-105"}, {ID: "OTHER-1"},
	}}
	AnnotateSources(&rep, reg)
	if rep.Requirements[0].Source != "platform" ||
		rep.Requirements[0].TicketURL != "https://company.atlassian.net/browse/PLAT-4521" {
		t.Errorf("PLAT annotation wrong: %+v", rep.Requirements[0])
	}
	if rep.Requirements[1].Source != "core" || rep.Requirements[1].TicketURL != "" {
		t.Errorf("CBIN annotation wrong: %+v", rep.Requirements[1])
	}
	if rep.Requirements[2].Source != "" {
		t.Errorf("unknown prefix must stay unannotated: %+v", rep.Requirements[2])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/canaryscan/ -v`
Expected: FAIL — `VerifyClaims` wrong arity, `normalizeREQWithRegistry`/`AnnotateSources` undefined, `Requirement.Source` undefined.

- [ ] **Step 3: Implement**

3a. `internal/canaryscan/types.go` — extend Requirement:

```go
type Requirement struct {
	ID        string    `json:"id"`
	Source    string    `json:"source,omitempty"`
	TicketURL string    `json:"ticket_url,omitempty"`
	Diagrams  []string  `json:"diagrams,omitempty"` // filled by Task 4
	Features  []Feature `json:"features"`
}
```

3b. `internal/canaryscan/parse.go` — delete the `claimRe` var (line 15). Replace `normalizeREQ` with a registry-aware pair; keep the old function delegating so existing call sites still compile:

```go
// package-level registry used by parse-time normalization; set by Run.
var activeRegistry *sources.Registry

func normalizeREQ(v string) string {
	return normalizeREQWithRegistry(v, activeRegistry)
}

func normalizeREQWithRegistry(v string, reg *sources.Registry) string {
	v = strings.TrimSpace(v)
	v = strings.ReplaceAll(v, "‑", "-")
	v = strings.ReplaceAll(v, "–", "-")
	if reg != nil {
		if s, ok := reg.Resolve(v); ok {
			if s.Type == "flatfile" {
				return reg.Normalize(v)
			}
			return v // external ticket IDs are verbatim, never padded
		}
	}
	pad := func(prefix, num string) string {
		for len(num) < 3 {
			num = "0" + num
		}
		return prefix + num
	}
	if m := regexp.MustCompile(`^(CBIN-)(\d{1,3})$`).FindStringSubmatch(v); len(m) == 3 {
		return pad(m[1], m[2])
	}
	if m := regexp.MustCompile(`^(REQ(?:-[A-Z]+)?-)(\d{1,3})$`).FindStringSubmatch(v); len(m) == 3 {
		return pad(m[1], m[2])
	}
	if m := regexp.MustCompile(`^((?:TASK|BUG)-)(\d{1,3})$`).FindStringSubmatch(v); len(m) == 3 {
		return pad(m[1], m[2])
	}
	return v
}
```

Add the import `go.devnw.com/canary/internal/sources`.

3c. `internal/canaryscan/verify.go` — new signature:

```go
// VerifyClaims reads the GAP file and returns diagnostics for claimed-but-not-
// TESTED/BENCHED requirements. Claims are lines like "✅ <ID>" where <ID>
// matches any configured source key; a nil registry means the default (CBIN).
// CANARY: REQ=CBIN-201; FEATURE="TicketSources"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_201_VerifyClaimsTicketSource; UPDATED=2026-08-28
func VerifyClaims(rep Report, gapPath string, reg *sources.Registry) []string {
	if reg == nil {
		reg = sources.Default()
	}
	b, err := os.ReadFile(gapPath)
	if err != nil {
		return []string{fmt.Sprintf("CANARY_PARSE_ERROR file=%s err=%q", gapPath, err)}
	}
	matches := reg.ClaimPattern().FindAllStringSubmatch(string(b), -1)
	// ... rest identical to the current body from line 17 down
}
```

3d. New helper in `internal/canaryscan/run.go` (or a new `annotate.go`):

```go
// AnnotateSources stamps each requirement with its source name and ticket URL.
func AnnotateSources(rep *Report, reg *sources.Registry) {
	if reg == nil {
		return
	}
	for i := range rep.Requirements {
		if s, ok := reg.Resolve(rep.Requirements[i].ID); ok {
			rep.Requirements[i].Source = s.Name
			rep.Requirements[i].TicketURL = reg.TicketURL(rep.Requirements[i].ID)
		}
	}
}
```

3e. In `Run` (`run.go:22`), immediately after the config defaults block, add:

```go
	reg := sources.LoadFromRoot(cfg.Root)
	activeRegistry = reg
	defer func() { activeRegistry = nil }()
```

After the successful `Scan` call(s) (both the initial one at line 57 and the re-scan after update-stale at line 76), add `AnnotateSources(&rep, reg)`. Change line 102 to `VerifyClaims(rep, cfg.VerifyPath, reg)`.

3f. Fix any other `VerifyClaims(` caller found by `grep -rn "VerifyClaims(" --include="*.go" .` — pass `nil` for legacy behavior.

- [ ] **Step 4: Run tests**

Run: `go test ./internal/canaryscan/ ./internal/sources/ -v && go test ./tools/... ./. `
Expected: PASS, including the existing acceptance tests in `tools/canary/internal/acceptance_test.go` (they exercise CBIN fixtures through `Run` — the Default fallback must keep them green).

- [ ] **Step 5: Commit**

```bash
git add internal/canaryscan/ mcp/
git commit -m "feat(scan): registry-driven claim verification, normalization, and source annotation"
```

---

### Task 4: Mermaid reference extraction in the scanner

**Files:**
- Create: `internal/canaryscan/mermaid.go`
- Modify: `internal/canaryscan/scan.go` (collect refs during walk, attach to Requirements in buildReport), `internal/canaryscan/run.go` (thread registry — already loaded in Task 3)
- Test: `internal/canaryscan/mermaid_test.go`

**Interfaces:**
- Consumes: `sources.Registry.Pattern()`, `Registry.Normalize()` from Task 2; `Requirement.Diagrams` field added in Task 3.
- Produces (used by Tasks 9, 10):
  - `type DiagramRef struct { ReqID, File string; Line int }`
  - `func ExtractDiagramRefs(relPath, content string, reg *sources.Registry) []DiagramRef` — finds requirement IDs inside ```` ```mermaid ```` fenced blocks in markdown content
  - `func ScanDiagramRefs(root string, skip *regexp.Regexp, reg *sources.Registry) ([]DiagramRef, error)` — walks root for `.md`/`.markdown`/`.mmd` files and extracts (`.mmd` files are whole-file mermaid, no fence needed)
  - `Requirement.Diagrams` populated as sorted unique `"file:line"` strings in scan output (status.json)

- [ ] **Step 1: Write the failing test**

Create `internal/canaryscan/mermaid_test.go`:

```go
package canaryscan

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const archMD = "# Architecture\n" +
	"CBIN-105 mentioned in prose is NOT a diagram ref.\n" +
	"```mermaid\n" +
	"flowchart TD\n" +
	"  A[CBIN-105 Scanner] --> B[CBIN-042 Storage]\n" +
	"  B --> C[PLAT-4521 Ingest]\n" +
	"```\n" +
	"```go\n" +
	"// CBIN-999 in a go fence is not a diagram ref\n" +
	"```\n" +
	"```mermaid\n" +
	"sequenceDiagram\n" +
	"  participant S as CBIN-105\n" +
	"```\n"

func TestCANARY_CBIN_202_ExtractDiagramRefs(t *testing.T) {
	reg := ticketRegistry(t) // from sources_test.go: CBIN flatfile + PLAT jira
	refs := ExtractDiagramRefs("docs/arch.md", archMD, reg)
	got := map[string][]int{}
	for _, r := range refs {
		if r.File != "docs/arch.md" {
			t.Errorf("File = %q", r.File)
		}
		got[r.ReqID] = append(got[r.ReqID], r.Line)
	}
	// CBIN-042 normalized (flatfile padding), PLAT verbatim, go-fence and prose excluded
	want := map[string][]int{"CBIN-105": {5, 13}, "CBIN-042": {5}, "PLAT-4521": {6}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("refs = %v, want %v", got, want)
	}
}

func TestCANARY_CBIN_202_ScanAttachesDiagrams(t *testing.T) {
	root := t.TempDir()
	code := "package x\n// CANARY: REQ=CBIN-105; FEATURE=\"Scanner\"; ASPECT=Engine; STATUS=IMPL; UPDATED=2026-08-28\n"
	if err := os.WriteFile(filepath.Join(root, "x.go"), []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}
	md := "```mermaid\nflowchart TD\n  A[CBIN-105] --> B[other]\n```\n"
	if err := os.WriteFile(filepath.Join(root, "arch.md"), []byte(md), 0o600); err != nil {
		t.Fatal(err)
	}
	rep, err := Scan(root, DefaultSkipRegex(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	AnnotateSources(&rep, nil)
	var found *Requirement
	for i := range rep.Requirements {
		if rep.Requirements[i].ID == "CBIN-105" {
			found = &rep.Requirements[i]
		}
	}
	if found == nil {
		t.Fatal("CBIN-105 not in report")
	}
	if !reflect.DeepEqual(found.Diagrams, []string{"arch.md:3"}) {
		t.Errorf("Diagrams = %v, want [arch.md:3]", found.Diagrams)
	}
}
```

Note: `TestCANARY_CBIN_202_ScanAttachesDiagrams` calls `Scan` with the current 4-arg signature and no registry — the scanner must use `activeRegistry` when set and `sources.Default()` otherwise, so diagram extraction works in both paths.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/canaryscan/ -run TestCANARY_CBIN_202 -v`
Expected: FAIL — `ExtractDiagramRefs` undefined.

- [ ] **Step 3: Implement**

Create `internal/canaryscan/mermaid.go`:

```go
package canaryscan

// CANARY: REQ=CBIN-202; FEATURE="MermaidRefs"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_202_ExtractDiagramRefs; UPDATED=2026-08-28

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"regexp"

	"go.devnw.com/canary/internal/sources"
)

// DiagramRef records one requirement-ID mention inside a mermaid diagram.
type DiagramRef struct {
	ReqID string
	File  string
	Line  int // 1-based
}

// ExtractDiagramRefs finds requirement IDs inside ```mermaid fenced blocks.
// relPath ending in .mmd is treated as a whole-file mermaid diagram.
// IDs are normalized through reg (flatfile padding); reg nil = Default.
func ExtractDiagramRefs(relPath, content string, reg *sources.Registry) []DiagramRef {
	if reg == nil {
		reg = sources.Default()
	}
	wholeFile := strings.HasSuffix(relPath, ".mmd")
	var refs []DiagramRef
	inMermaid := wholeFile
	for i, line := range strings.Split(content, "\n") {
		if !wholeFile {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "```") {
				if inMermaid {
					inMermaid = false
				} else if strings.HasPrefix(trimmed, "```mermaid") {
					inMermaid = true
				} else {
					inMermaid = false // entering a non-mermaid fence
					// skip until its closing fence; handled by the toggle above
				}
				continue
			}
		}
		if !inMermaid {
			continue
		}
		for _, m := range reg.Pattern().FindAllString(line, -1) {
			refs = append(refs, DiagramRef{ReqID: reg.Normalize(m), File: relPath, Line: i + 1})
		}
	}
	return refs
}

// ScanDiagramRefs walks root for markdown/mermaid files and extracts all refs.
// Paths in the result are root-relative with forward slashes.
func ScanDiagramRefs(root string, skip *regexp.Regexp, reg *sources.Registry) ([]DiagramRef, error) {
	if skip == nil {
		skip = DefaultSkipRegex()
	}
	var out []DiagramRef
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skip.MatchString(path) {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".markdown" && ext != ".mmd" {
			return nil
		}
		if skip.MatchString(path) {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil // unreadable markdown is not fatal to a scan
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			rel = path
		}
		out = append(out, ExtractDiagramRefs(filepath.ToSlash(rel), string(b), reg)...)
		return nil
	})
	return out, err
}
```

**Careful with the non-mermaid-fence toggle:** the simple toggle above is wrong for non-mermaid fences (a ```` ```go ```` fence would flip state). Implement it with an explicit two-state machine instead:

```go
	const (
		outside = iota
		inMermaidBlock
		inOtherBlock
	)
	state := outside
	if wholeFile {
		state = inMermaidBlock
	}
	for i, line := range strings.Split(content, "\n") {
		if !wholeFile {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "```") {
				switch state {
				case outside:
					if strings.HasPrefix(trimmed, "```mermaid") {
						state = inMermaidBlock
					} else {
						state = inOtherBlock
					}
				default:
					state = outside
				}
				continue
			}
		}
		if state != inMermaidBlock {
			continue
		}
		// ... FindAllString as above
	}
```

Use the state-machine version; delete the toggle sketch.

Then in `internal/canaryscan/scan.go`, at the end of `Scan` (after the report is built, before return), collect and attach refs:

```go
	// Attach mermaid diagram references (Task CBIN-202).
	reg := activeRegistry
	if reg == nil {
		reg = sources.Default()
	}
	diagRefs, _ := ScanDiagramRefs(root, skip, reg)
	byID := map[string]map[string]struct{}{}
	for _, r := range diagRefs {
		key := fmt.Sprintf("%s:%d", r.File, r.Line)
		if byID[r.ReqID] == nil {
			byID[r.ReqID] = map[string]struct{}{}
		}
		byID[r.ReqID][key] = struct{}{}
	}
	for i := range rep.Requirements {
		if set, ok := byID[rep.Requirements[i].ID]; ok {
			rep.Requirements[i].Diagrams = sortedKeys(set) // sort.Strings over map keys
		}
	}
```

(Add a small `sortedKeys(map[string]struct{}) []string` helper next to `mapKeys` in parse.go, or reuse `mapKeys` — it already does exactly this.)

- [ ] **Step 4: Run tests**

Run: `go test ./internal/canaryscan/ -v && go test ./tools/...`
Expected: PASS. If acceptance tests fail because fixture markdown now contributes `diagrams` fields, inspect — `Diagrams` is `omitempty`, and fixtures contain no mermaid fences, so they must not change.

- [ ] **Step 5: Commit**

```bash
git add internal/canaryscan/
git commit -m "feat(scan): extract requirement references from mermaid diagrams into status.json"
```

---

### Task 5: Prefix-agnostic dependency parsing (`internal/specs`)

**Files:**
- Modify: `internal/specs/parser_dependency.go:14-21` (the two regexes)
- Test: `internal/specs/parser_dependency_test.go` (append to existing file)

**Interfaces:**
- Consumes: nothing new (pure regex generalization — registry not needed here; any `[A-Z][A-Z0-9]*-\d+` shaped ID is accepted, matching how spec dirs are keyed by literal REQ-ID prefix).
- Produces: `ParseDependencies` accepting any-prefix IDs; all existing signatures unchanged.

- [ ] **Step 1: Write the failing test**

Append to `internal/specs/parser_dependency_test.go` (match the existing test style in that file — table-driven with `strings.NewReader`):

```go
func TestCANARY_CBIN_201_ParseDependencies_TicketPrefixes(t *testing.T) {
	spec := `# Spec
## Dependencies
- PLAT-4521 (JIRA upstream ingest)
- GL-88:Auth (GitLab auth feature)
- CBIN-105 (core scanner)
## Next Section
- OTHER-1 (outside the section, ignored)
`
	deps, err := ParseDependencies("PLAT-9000", strings.NewReader(spec))
	if err != nil {
		t.Fatalf("ParseDependencies: %v", err)
	}
	if len(deps) != 3 {
		t.Fatalf("got %d deps, want 3: %+v", len(deps), deps)
	}
	targets := map[string]bool{}
	for _, d := range deps {
		targets[d.Target] = true
		if d.Source != "PLAT-9000" {
			t.Errorf("Source = %q, want PLAT-9000", d.Source)
		}
	}
	for _, want := range []string{"PLAT-4521", "GL-88", "CBIN-105"} {
		if !targets[want] {
			t.Errorf("missing dependency target %s in %v", want, targets)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/specs/ -run TestCANARY_CBIN_201 -v`
Expected: FAIL — only CBIN-105 parses (got 1 dep, want 3).

- [ ] **Step 3: Implement**

In `internal/specs/parser_dependency.go` replace lines 14-21:

```go
var (
	// Regex patterns for parsing dependency lines. Any configured requirement
	// prefix is accepted ([A-Z][A-Z0-9]*-digits), not just CBIN — ticket-system
	// IDs (JIRA PLAT-4521, GitLab GL-88, GitHub GH-7) are valid dependencies.
	// Format: "- PREFIX-123 (Description)" for full dependencies
	// Format: "- PREFIX-123:Feature1,Feature2 (Description)" for partial feature dependencies
	// Format: "- PREFIX-123:AspectName (Description)" for partial aspect dependencies
	fullDependencyPattern    = regexp.MustCompile(`^-\s+([A-Z][A-Z0-9]*-\d+)\s*(?:\(([^)]+)\))?`)
	partialDependencyPattern = regexp.MustCompile(`^-\s+([A-Z][A-Z0-9]*-\d+):([^(\s]+)\s*(?:\(([^)]+)\))?`)
)
```

Update the CANARY token on line 12: append `,TestCANARY_CBIN_201_ParseDependencies_TicketPrefixes` to its `TEST=` list and set `UPDATED=2026-08-28`.

- [ ] **Step 4: Run tests**

Run: `go test ./internal/specs/ -v`
Expected: PASS — all existing dependency/graph/integration tests still green (CBIN is a subset of the general pattern).

- [ ] **Step 5: Commit**

```bash
git add internal/specs/parser_dependency.go internal/specs/parser_dependency_test.go
git commit -m "fix(specs): accept any requirement prefix in spec dependencies, not just CBIN"
```

---

### Task 6: Mermaid graph rendering (`deps graph --format mermaid`)

**Files:**
- Modify: `internal/specs/graph_generator.go` (add `FormatMermaid`), `internal/cmds/deps/deps.go:131-189` (`createDepsGraphCommand` gains `--format`)
- Test: `internal/specs/graph_generator_test.go` (append)

**Interfaces:**
- Consumes: `DependencyGraph`, `Dependency{Source, Target, Type, Description}`, `GraphGenerator.GetTransitiveDependencies` (existing, `graph_generator.go:43,68`); `sources.LoadFromRoot(".").TicketURL` in the CLI layer.
- Produces: `func (gg *GraphGenerator) FormatMermaid(graph *DependencyGraph, rootReqID string, urlFor func(string) string) string` — `urlFor` may be nil; when it returns non-empty, a `click` line is emitted.

- [ ] **Step 1: Write the failing test**

Append to `internal/specs/graph_generator_test.go` (reuse that file's existing mock loader helpers — read the top of the file first and construct the graph the same way neighboring tests do):

```go
func TestCANARY_CBIN_203_FormatMermaid(t *testing.T) {
	// Build a two-level graph: CBIN-300 -> CBIN-200 -> CBIN-100, CBIN-300 -> PLAT-4521
	graph := &DependencyGraph{
		Nodes: map[string][]Dependency{
			"CBIN-300": {
				{Source: "CBIN-300", Target: "CBIN-200", Type: DependencyTypeFull, Description: "storage"},
				{Source: "CBIN-300", Target: "PLAT-4521", Type: DependencyTypeFull, Description: "upstream"},
			},
			"CBIN-200": {
				{Source: "CBIN-200", Target: "CBIN-100", Type: DependencyTypeFull},
			},
		},
	}
	gg := NewGraphGenerator(nil)
	urlFor := func(id string) string {
		if id == "PLAT-4521" {
			return "https://company.atlassian.net/browse/PLAT-4521"
		}
		return ""
	}
	out := gg.FormatMermaid(graph, "CBIN-300", urlFor)

	for _, want := range []string{
		"flowchart TD",
		`CBIN_300["CBIN-300"]`,
		"CBIN_300 --> CBIN_200",
		"CBIN_300 --> PLAT_4521",
		"CBIN_200 --> CBIN_100",
		`click PLAT_4521 "https://company.atlassian.net/browse/PLAT-4521"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("mermaid output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, `click CBIN_300`) {
		t.Error("no click line should be emitted for empty URLs")
	}
}

func TestCANARY_CBIN_203_FormatMermaid_CycleSafe(t *testing.T) {
	graph := &DependencyGraph{
		Nodes: map[string][]Dependency{
			"CBIN-100": {{Source: "CBIN-100", Target: "CBIN-200", Type: DependencyTypeFull}},
			"CBIN-200": {{Source: "CBIN-200", Target: "CBIN-100", Type: DependencyTypeFull}},
		},
	}
	gg := NewGraphGenerator(nil)
	out := gg.FormatMermaid(graph, "CBIN-100", nil) // must terminate
	if c := strings.Count(out, "CBIN_100 --> CBIN_200"); c != 1 {
		t.Errorf("edge emitted %d times, want 1", c)
	}
}
```

**Adjust the struct-literal construction to the real `DependencyGraph` shape** — read `internal/specs/types.go:14-160` first; if `DependencyGraph` uses different field names (e.g. `Dependencies map[string][]Dependency`), use those. If `DependencyTypeFull` is named differently (e.g. `FullDependency`), use the real constant.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/specs/ -run TestCANARY_CBIN_203 -v`
Expected: FAIL — `FormatMermaid` undefined.

- [ ] **Step 3: Implement**

Append to `internal/specs/graph_generator.go`:

```go
// CANARY: REQ=CBIN-203; FEATURE="MermaidGraph"; ASPECT=Engine; STATUS=IMPL; TEST=TestCANARY_CBIN_203_FormatMermaid; UPDATED=2026-08-28

// mermaidNodeID converts a requirement ID to a mermaid-safe node identifier.
func mermaidNodeID(reqID string) string {
	return strings.NewReplacer("-", "_", ".", "_", "/", "_", "#", "_").Replace(reqID)
}

// FormatMermaid renders the dependency graph rooted at rootReqID as a mermaid
// flowchart. urlFor (optional) supplies a ticket/docs URL per requirement ID;
// non-empty URLs become mermaid click directives.
func (gg *GraphGenerator) FormatMermaid(graph *DependencyGraph, rootReqID string, urlFor func(string) string) string {
	var b strings.Builder
	b.WriteString("flowchart TD\n")

	seenNode := map[string]bool{}
	seenEdge := map[string]bool{}
	var declare func(id string)
	declare = func(id string) {
		if seenNode[id] {
			return
		}
		seenNode[id] = true
		fmt.Fprintf(&b, "    %s[\"%s\"]\n", mermaidNodeID(id), id)
	}

	var walk func(id string)
	walk = func(id string) {
		declare(id)
		for _, dep := range graph.Nodes[id] {
			edge := dep.Source + "->" + dep.Target
			if seenEdge[edge] {
				continue
			}
			seenEdge[edge] = true
			declare(dep.Target)
			label := ""
			if dep.Description != "" {
				label = fmt.Sprintf("|%q|", dep.Description)
			}
			fmt.Fprintf(&b, "    %s -->%s %s\n", mermaidNodeID(dep.Source), label, mermaidNodeID(dep.Target))
			walk(dep.Target)
		}
	}
	walk(rootReqID)

	if urlFor != nil {
		ids := make([]string, 0, len(seenNode))
		for id := range seenNode {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			if u := urlFor(id); u != "" {
				fmt.Fprintf(&b, "    click %s \"%s\"\n", mermaidNodeID(id), u)
			}
		}
	}
	return b.String()
}
```

(Adapt `graph.Nodes` to the real field name. The `label` with `%q` inside `|...|` produces `|"storage"|` which mermaid renders as an edge label; if existing style prefers plain, use `fmt.Sprintf("|%s|", dep.Description)`.)

Then in `internal/cmds/deps/deps.go` `createDepsGraphCommand` (line 131): add a `--format` string flag (default `"ascii"`, accepted values `ascii|mermaid`). Where the command currently prints `FormatASCIITree(...)`, branch:

```go
		if format == "mermaid" {
			reg := sources.LoadFromRoot(".")
			fmt.Println(gg.FormatMermaid(graph, reqID, reg.TicketURL))
			return nil
		}
		// existing ascii output unchanged
```

Add the `go.devnw.com/canary/internal/sources` import.

- [ ] **Step 4: Run tests**

Run: `go test ./internal/specs/ -v && go build ./... && go vet ./...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/specs/graph_generator.go internal/specs/graph_generator_test.go internal/cmds/deps/deps.go
git commit -m "feat(deps): mermaid output for dependency graphs with ticket click-through"
```

---

### Task 7: Pattern-driven storage queries + bounded search

**Files:**
- Modify: `internal/storage/storage.go` — `ListTokens` (line 253, remove hardcoded CBIN GLOBs, honor `idPattern`), `SearchTokens` (line 374, add `limit int` param)
- Modify callers of `SearchTokens`: `internal/cmds/search/search.go:41`, `ops.go` (`GrepTokens` line ~21-57), `mcp/tools.go:330` (handleSearch), `mcp/tools_extended.go` (handleGrep ~line 371) — find all with `grep -rn "SearchTokens(" --include="*.go" .`
- Test: `internal/storage/queries_test.go` (append; the file exists and has DB test helpers — reuse `testutil`)

**Interfaces:**
- Consumes: existing `storage.Token`, `testutil` helpers (`internal/storage/testutil/testutil.go`).
- Produces:
  - `ListTokens(filters map[string]string, idPattern string, orderBy string, limit int)` — same signature, but `idPattern` (a Go regexp string) is now actually applied (in Go, post-query) and the hardcoded `CBIN` GLOB block is deleted. Empty `idPattern` = no ID filtering beyond placeholder exclusion.
  - `SearchTokens(keywords string, limit int) ([]*Token, error)` — `limit <= 0` applies `DefaultSearchLimit = 100` (a new exported const); SQL gains `LIMIT ?`.

- [ ] **Step 1: Write the failing test**

Append to `internal/storage/queries_test.go` (mirror the setup pattern used by existing tests in that file — open a temp DB via testutil, upsert tokens, query):

```go
func TestCANARY_CBIN_201_ListTokensHonorsIDPattern(t *testing.T) {
	db := testutil.OpenTestDB(t) // use the real helper name from testutil.go
	seed := []*Token{
		{ReqID: "CBIN-105", Feature: "A", Aspect: "API", Status: "IMPL", FilePath: "a.go", LineNumber: 1},
		{ReqID: "PLAT-4521", Feature: "B", Aspect: "API", Status: "IMPL", FilePath: "b.go", LineNumber: 1},
		{ReqID: "GL-88", Feature: "C", Aspect: "API", Status: "IMPL", FilePath: "c.go", LineNumber: 1},
		{ReqID: "CBIN-XXX", Feature: "Tmpl", Aspect: "API", Status: "IMPL", FilePath: "t.go", LineNumber: 1},
	}
	for _, tok := range seed {
		if err := db.UpsertToken(tok); err != nil {
			t.Fatal(err)
		}
	}

	// Pattern matching two prefixes
	got, err := db.ListTokens(nil, `^(CBIN|PLAT)-\d+$`, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, tok := range got {
		ids[tok.ReqID] = true
	}
	if !ids["CBIN-105"] || !ids["PLAT-4521"] || ids["GL-88"] {
		t.Errorf("pattern filter wrong: %v", ids)
	}

	// Empty pattern: everything except placeholders
	got, err = db.ListTokens(nil, "", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	ids = map[string]bool{}
	for _, tok := range got {
		ids[tok.ReqID] = true
	}
	if !ids["GL-88"] {
		t.Error("empty pattern must include ticket-source tokens (GL-88)")
	}
	if ids["CBIN-XXX"] {
		t.Error("placeholder CBIN-XXX must stay excluded")
	}
}

func TestCANARY_CBIN_205_SearchTokensLimit(t *testing.T) {
	db := testutil.OpenTestDB(t)
	for i := 0; i < 30; i++ {
		tok := &Token{ReqID: fmt.Sprintf("CBIN-%03d", 100+i), Feature: "needle", Aspect: "API",
			Status: "IMPL", FilePath: fmt.Sprintf("f%d.go", i), LineNumber: 1}
		if err := db.UpsertToken(tok); err != nil {
			t.Fatal(err)
		}
	}
	got, err := db.SearchTokens("needle", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 10 {
		t.Errorf("len = %d, want 10", len(got))
	}
	got, err = db.SearchTokens("needle", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != DefaultSearchLimit {
		t.Errorf("limit 0 should apply DefaultSearchLimit(%d), got %d of 30", DefaultSearchLimit, len(got))
	}
}
```

**Adapt names:** read `internal/storage/testutil/testutil.go` first for the real helper (it may be `testutil.NewTestDB(t)` or similar) and check `Token` field names against `storage.go:19-57` (`FilePath`, `LineNumber` etc.). Tokens may need `Updated`/`Indexed` fields set if UpsertToken requires them.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/storage/ -run 'TestCANARY_CBIN_201_ListTokens|TestCANARY_CBIN_205_Search' -v`
Expected: FAIL — `SearchTokens` wrong arity; `ListTokens` returns GL-88-excluding CBIN-hardcoded results.

- [ ] **Step 3: Implement**

In `ListTokens` (`storage.go:253`): delete the hardcoded block that injects `req_id GLOB 'CBIN-...'` / `BUG-` GLOBs and the `CBIN-XXX`-specific LIKEs; replace with generic placeholder exclusion in SQL, keeping all other filters (status/aspect/hidden-path/etc.) untouched:

```go
	// Exclude template/placeholder tokens regardless of project prefix.
	query += " AND req_id NOT LIKE '%XXX%' AND req_id NOT LIKE '%###%' AND req_id NOT LIKE '{{%' AND req_id NOT LIKE '%}}%'"
```

After scanning rows, apply the pattern in Go:

```go
	if idPattern != "" {
		re, err := regexp.Compile(idPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid id pattern %q: %w", idPattern, err)
		}
		filtered := tokens[:0]
		for _, tok := range tokens {
			if re.MatchString(tok.ReqID) {
				filtered = append(filtered, tok)
			}
		}
		tokens = filtered
	}
```

**Important:** when a `limit` was applied in SQL AND a pattern filters rows out, results can undershoot the limit — acceptable; do not over-engineer with SQL regexp.

Note: existing callers pass patterns like `CBIN-[1-9][0-9]{2,}` (unanchored). `MatchString` is substring semantics, which preserves current intent.

`SearchTokens`:

```go
// DefaultSearchLimit caps keyword searches to protect agent context.
// Deliberately small; callers raise it explicitly (--limit / limit param)
// when they need more.
const DefaultSearchLimit = 25

// CANARY: REQ=CBIN-205; FEATURE="ContextCaps"; ASPECT=Storage; STATUS=IMPL; TEST=TestCANARY_CBIN_205_SearchTokensLimit; UPDATED=2026-08-28
func (db *DB) SearchTokens(keywords string, limit int) ([]*Token, error) {
	if limit <= 0 {
		limit = DefaultSearchLimit
	}
	// ... existing query construction ...
	query += " LIMIT ?"
	args = append(args, limit)
	// ... existing row scan ...
}
```

Update every `SearchTokens(` caller: CLI `search` passes its new `--limit` flag value (wired in Task 11 — for now pass `0`); `ops.go` GrepTokens passes `0`; MCP handlers pass `0` (tightened in Task 12).

Also fix `GrepTokens` in `ops.go` (~line 21-57): it currently calls `db.ListTokens(nil, "", "", 0)` to load the entire table. Replace that with `db.SearchTokens(pattern, 0)` results only (drop the full-table union) — the union added nothing except unbounded memory use, since SearchTokens already does LIKE matching over the same columns. Check `ops.go`'s actual merge logic first; keep the output formatting identical.

- [ ] **Step 4: Run tests**

Run: `go test ./internal/storage/ -v && go build ./... && go test ./...`
Expected: PASS. Watch `internal/cmds` and `mcp` compile errors from the arity change — fix all callers.

- [ ] **Step 5: Commit**

```bash
git add internal/storage/ internal/cmds/ mcp/ ops.go
git commit -m "feat(storage): honor id_pattern in ListTokens, bound SearchTokens, drop CBIN-only SQL"
```

---

### Task 8: Prefix-agnostic stale updater

**Files:**
- Modify: `internal/canaryscan/update.go:16`
- Test: `internal/canaryscan/update_test.go` (create)

**Interfaces:**
- Consumes/Produces: no signature changes — pure regex generalization.

- [ ] **Step 1: Write the failing test**

Create `internal/canaryscan/update_test.go`:

```go
package canaryscan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCANARY_CBIN_201_UpdateStaleTicketIDs(t *testing.T) {
	t.Setenv("CANARY_TEST_TIMESTAMP", "2026-08-28T00:00:00Z")
	root := t.TempDir()
	src := "package x\n" +
		"// CANARY: REQ=PLAT-4521; FEATURE=\"Ingest\"; ASPECT=API; STATUS=TESTED; TEST=TestIngest; UPDATED=2020-01-01\n"
	path := filepath.Join(root, "x.go")
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	diags := []string{"CANARY_STALE REQ=PLAT-4521 updated=2020-01-01 age_days=2431 threshold=30"}
	if _, err := UpdateStaleTokens(root, DefaultSkipRegex(), diags); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	if strings.Contains(string(b), "UPDATED=2020-01-01") {
		t.Errorf("stale ticket-sourced token was not updated:\n%s", b)
	}
	if !strings.Contains(string(b), "UPDATED=2026-08-28") {
		t.Errorf("UPDATED not rewritten to test timestamp:\n%s", b)
	}
}
```

**Check `UpdateStaleTokens`'s actual behavior first** (`internal/canaryscan/update.go:14-95`): confirm the exact date format it writes (`2006-01-02` from the timestamp) and the diag format it parses; adjust the assertion accordingly.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/canaryscan/ -run TestCANARY_CBIN_201_UpdateStale -v`
Expected: FAIL — the `REQ=([A-Z]+-\d{3})` regex won't match `PLAT-4521` (4 digits).

- [ ] **Step 3: Implement**

In `internal/canaryscan/update.go:16` change:

```go
var reqFromDiag = regexp.MustCompile(`REQ=([A-Z][A-Z0-9]*-\d+)`)
```

(match the actual var name at that line) — and check the same file for any other `\d{3}`-anchored regex that touches REQ IDs (e.g. the in-file token matcher used to locate `UPDATED=`); generalize those the same way: `[A-Z][A-Z0-9]*-\d+`.

- [ ] **Step 4: Run tests**

Run: `go test ./internal/canaryscan/ -v && go test ./tools/...`
Expected: PASS including `TestAcceptance_UpdateStale`.

- [ ] **Step 5: Commit**

```bash
git add internal/canaryscan/update.go internal/canaryscan/update_test.go
git commit -m "fix(scan): update-stale handles any requirement prefix and ID width"
```

---

### Task 9: Diagram refs in the database (migration 000006 + index)

**Files:**
- Create: `internal/storage/migrations/000006_add_refs_table.up.sql`, `internal/storage/migrations/000006_add_refs_table.down.sql`
- Modify: `internal/storage/db.go` (`LatestVersion` 5 → 6), `internal/storage/storage.go` (add `Ref` type + `ReplaceRefs` + `GetRefsByReqID`), `internal/cmds/index/index.go` (populate refs after token indexing)
- Test: `internal/storage/refs_test.go` (create)

**Interfaces:**
- Consumes: `canaryscan.ScanDiagramRefs` + `canaryscan.DiagramRef` from Task 4, `sources.LoadFromRoot` from Task 2.
- Produces (used by Tasks 10, 13):
  - `type Ref struct { ReqID string \`db:"req_id"\`; Kind string \`db:"kind"\`; FilePath string \`db:"file_path"\`; LineNumber int \`db:"line_number"\`; Context string \`db:"context"\` }`
  - `func (db *DB) ReplaceRefs(kind string, refs []Ref) error` — transactional delete-by-kind + bulk insert
  - `func (db *DB) GetRefsByReqID(reqID string) ([]*Ref, error)` — ordered by file_path, line_number

- [ ] **Step 1: Write the migration files**

`internal/storage/migrations/000006_add_refs_table.up.sql`:

```sql
-- Requirement references found outside CANARY tokens (e.g. mermaid diagrams).
-- kind: 'diagram' today; future kinds may include 'doc', 'adr'.
CREATE TABLE IF NOT EXISTS refs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    req_id TEXT NOT NULL,
    kind TEXT NOT NULL DEFAULT 'diagram',
    file_path TEXT NOT NULL,
    line_number INTEGER NOT NULL DEFAULT 0,
    context TEXT DEFAULT '',
    indexed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(req_id, kind, file_path, line_number)
);

CREATE INDEX IF NOT EXISTS idx_refs_req_id ON refs(req_id);
CREATE INDEX IF NOT EXISTS idx_refs_kind ON refs(kind);
```

`internal/storage/migrations/000006_add_refs_table.down.sql`:

```sql
DROP INDEX IF EXISTS idx_refs_kind;
DROP INDEX IF EXISTS idx_refs_req_id;
DROP TABLE IF EXISTS refs;
```

Bump `LatestVersion = 6` in `internal/storage/db.go` (currently `= 5` around line 33).

- [ ] **Step 2: Write the failing test**

Create `internal/storage/refs_test.go` (reuse the same testutil DB helper as Task 7):

```go
package storage

import "testing"

func TestCANARY_CBIN_206_RefsRoundTrip(t *testing.T) {
	db := testutil.OpenTestDB(t)
	refs := []Ref{
		{ReqID: "CBIN-105", Kind: "diagram", FilePath: "docs/arch.md", LineNumber: 12, Context: "flowchart"},
		{ReqID: "CBIN-105", Kind: "diagram", FilePath: "docs/flow.mmd", LineNumber: 3},
		{ReqID: "PLAT-4521", Kind: "diagram", FilePath: "docs/arch.md", LineNumber: 14},
	}
	if err := db.ReplaceRefs("diagram", refs); err != nil {
		t.Fatalf("ReplaceRefs: %v", err)
	}
	got, err := db.GetRefsByReqID("CBIN-105")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d refs, want 2", len(got))
	}
	if got[0].FilePath != "docs/arch.md" || got[0].LineNumber != 12 {
		t.Errorf("ordering/content wrong: %+v", got[0])
	}

	// Replace is idempotent and clears old rows of the same kind.
	if err := db.ReplaceRefs("diagram", refs[:1]); err != nil {
		t.Fatal(err)
	}
	got, _ = db.GetRefsByReqID("CBIN-105")
	if len(got) != 1 {
		t.Errorf("after replace: got %d refs, want 1", len(got))
	}
	got, _ = db.GetRefsByReqID("PLAT-4521")
	if len(got) != 0 {
		t.Errorf("PLAT refs should be cleared by replace, got %d", len(got))
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./internal/storage/ -run TestCANARY_CBIN_206 -v`
Expected: FAIL — `Ref`/`ReplaceRefs`/`GetRefsByReqID` undefined.

- [ ] **Step 4: Implement storage API**

Append to `internal/storage/storage.go` (or a new `internal/storage/refs.go` with license header):

```go
// CANARY: REQ=CBIN-206; FEATURE="DiagramRefsIndex"; ASPECT=Storage; STATUS=IMPL; TEST=TestCANARY_CBIN_206_RefsRoundTrip; UPDATED=2026-08-28

// Ref is a requirement reference found outside CANARY tokens (diagrams, docs).
type Ref struct {
	ReqID      string `db:"req_id" json:"req_id"`
	Kind       string `db:"kind" json:"kind"`
	FilePath   string `db:"file_path" json:"file_path"`
	LineNumber int    `db:"line_number" json:"line_number"`
	Context    string `db:"context" json:"context,omitempty"`
}

// ReplaceRefs atomically replaces all refs of the given kind.
func (db *DB) ReplaceRefs(kind string, refs []Ref) error {
	tx, err := db.conn.Beginx() // match the actual sqlx handle field name in DB
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`DELETE FROM refs WHERE kind = ?`, kind); err != nil {
		return err
	}
	for _, r := range refs {
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO refs (req_id, kind, file_path, line_number, context) VALUES (?, ?, ?, ?, ?)`,
			r.ReqID, kind, r.FilePath, r.LineNumber, r.Context,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetRefsByReqID returns refs for one requirement, ordered by file then line.
func (db *DB) GetRefsByReqID(reqID string) ([]*Ref, error) {
	rows, err := db.conn.Queryx(
		`SELECT req_id, kind, file_path, line_number, COALESCE(context,'') AS context
		 FROM refs WHERE req_id = ? ORDER BY file_path, line_number`, reqID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []*Ref
	for rows.Next() {
		var r Ref
		if err := rows.StructScan(&r); err != nil {
			return nil, err
		}
		out = append(out, &r)
	}
	return out, rows.Err()
}
```

**Adapt `db.conn`** to the real unexported field on `storage.DB` (check `storage.go:75-78`; it may be `db.db` or an embedded `*sqlx.DB`).

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/storage/ -v`
Expected: PASS (migration 6 applies in testutil's MigrateDB path).

- [ ] **Step 6: Wire into `canary index`**

In `internal/cmds/index/index.go`, after the token-indexing loop completes (before the summary print), add:

```go
	// Index diagram references (mermaid) so `canary view` can answer without grepping.
	reg := sources.LoadFromRoot(root)
	diagRefs, derr := canaryscan.ScanDiagramRefs(root, nil, reg)
	if derr == nil {
		refs := make([]storage.Ref, 0, len(diagRefs))
		for _, r := range diagRefs {
			refs = append(refs, storage.Ref{ReqID: r.ReqID, Kind: "diagram", FilePath: r.File, LineNumber: r.Line})
		}
		if err := db.ReplaceRefs("diagram", refs); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to index diagram refs: %v\n", err)
		} else if len(refs) > 0 {
			fmt.Printf("Indexed %d diagram reference(s)\n", len(refs))
		}
	}
```

(Match the file's existing variable names for root/db; add imports `go.devnw.com/canary/internal/canaryscan`, `go.devnw.com/canary/internal/sources`.)

- [ ] **Step 7: Verify build + full storage/index tests, then commit**

Run: `go build ./... && go test ./internal/storage/ ./internal/cmds/... -count=1`
Expected: PASS.

```bash
git add internal/storage/ internal/cmds/index/
git commit -m "feat(storage): refs table (migration 6) and diagram-ref indexing"
```

---

### Task 10: `canary view <REQ>` — the one-call full picture

**Files:**
- Create: `internal/cmds/view/view.go`
- Modify: `cli/cmds.go` (register `view.CreateViewCommand()` in the `Commands()` slice)
- Test: `internal/cmds/view/view_test.go`

**Interfaces:**
- Consumes: `storage.Open`, `db.GetTokensByReqID`, `db.GetRefsByReqID` (Task 9), `sources.LoadFromRoot` / `Resolve` / `TicketURL` (Task 2), `specs` spec-path glob convention (`.canary/specs/<REQ-ID>-*/spec.md`).
- Produces:
  - CLI: `canary view <REQ-ID> [--json] [--limit N]` — compact (~20 line) aggregate: header, source+ticket URL, status rollup, files (grouped, capped at N default 10 with `+N more` hint), tests, deps (DEPENDS_ON/BLOCKS/RELATED_TO from tokens + spec `## Dependencies` when a spec exists), spec/plan paths, diagram refs.
  - Go API for MCP (Task 13): `func BuildView(dbPath, root, reqID string, limit int) (*View, error)` and:

```go
type View struct {
	ReqID     string   `json:"req_id"`
	Source    string   `json:"source,omitempty"`
	TicketURL string   `json:"ticket_url,omitempty"`
	Statuses  map[string]int `json:"statuses"`       // status -> token count
	Completion int     `json:"completion_pct"`        // TESTED+BENCHED tokens / total
	Features  []string `json:"features"`              // "Feature (ASPECT, STATUS)"
	Files     []string `json:"files"`                 // capped at limit
	FilesTotal int     `json:"files_total"`
	Tests     []string `json:"tests"`
	Benches   []string `json:"benches,omitempty"`
	DependsOn []string `json:"depends_on,omitempty"`
	Blocks    []string `json:"blocks,omitempty"`
	RelatedTo []string `json:"related_to,omitempty"`
	SpecPath  string   `json:"spec_path,omitempty"`
	PlanPath  string   `json:"plan_path,omitempty"`
	Diagrams  []string `json:"diagrams,omitempty"`    // "file:line"
}
```

- [ ] **Step 1: Write the failing test**

Create `internal/cmds/view/view_test.go`:

```go
// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package view

import (
	"os"
	"path/filepath"
	"testing"

	"go.devnw.com/canary/internal/storage"
)

func seedDB(t *testing.T) (dbPath, root string) {
	t.Helper()
	root = t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".canary"), 0o750); err != nil {
		t.Fatal(err)
	}
	dbPath = filepath.Join(root, ".canary", "canary.db")
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	// storage.Open must run migrations; if it does not, call the migrate
	// helper used by cmd/canary/main.go:34-48 here.
	toks := []*storage.Token{
		{ReqID: "CBIN-105", Feature: "Scanner", Aspect: "Engine", Status: "TESTED",
			FilePath: "scan.go", LineNumber: 10, Test: "TestScan", DependsOn: "CBIN-101"},
		{ReqID: "CBIN-105", Feature: "ScannerCLI", Aspect: "CLI", Status: "IMPL",
			FilePath: "cli.go", LineNumber: 5, RelatedTo: "PLAT-4521"},
	}
	for _, tok := range toks {
		if err := db.UpsertToken(tok); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.ReplaceRefs("diagram", []storage.Ref{
		{ReqID: "CBIN-105", Kind: "diagram", FilePath: "docs/arch.md", LineNumber: 7},
	}); err != nil {
		t.Fatal(err)
	}
	return dbPath, root
}

func TestCANARY_CBIN_204_BuildView(t *testing.T) {
	dbPath, root := seedDB(t)
	v, err := BuildView(dbPath, root, "CBIN-105", 10)
	if err != nil {
		t.Fatalf("BuildView: %v", err)
	}
	if v.ReqID != "CBIN-105" {
		t.Errorf("ReqID = %q", v.ReqID)
	}
	if v.Statuses["TESTED"] != 1 || v.Statuses["IMPL"] != 1 {
		t.Errorf("Statuses = %v", v.Statuses)
	}
	if v.Completion != 50 {
		t.Errorf("Completion = %d, want 50", v.Completion)
	}
	if len(v.Files) != 2 || v.FilesTotal != 2 {
		t.Errorf("Files = %v (total %d)", v.Files, v.FilesTotal)
	}
	if len(v.Tests) != 1 || v.Tests[0] != "TestScan" {
		t.Errorf("Tests = %v", v.Tests)
	}
	if len(v.DependsOn) != 1 || v.DependsOn[0] != "CBIN-101" {
		t.Errorf("DependsOn = %v", v.DependsOn)
	}
	if len(v.RelatedTo) != 1 || v.RelatedTo[0] != "PLAT-4521" {
		t.Errorf("RelatedTo = %v", v.RelatedTo)
	}
	if len(v.Diagrams) != 1 || v.Diagrams[0] != "docs/arch.md:7" {
		t.Errorf("Diagrams = %v", v.Diagrams)
	}
}

func TestCANARY_CBIN_204_BuildView_FileCap(t *testing.T) {
	dbPath, root := seedDB(t)
	v, err := BuildView(dbPath, root, "CBIN-105", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Files) != 1 || v.FilesTotal != 2 {
		t.Errorf("cap not applied: files=%v total=%d", v.Files, v.FilesTotal)
	}
}

func TestCANARY_CBIN_204_BuildView_NotFound(t *testing.T) {
	dbPath, root := seedDB(t)
	if _, err := BuildView(dbPath, root, "CBIN-999", 10); err == nil {
		t.Error("unknown requirement must return an error")
	}
}
```

**Adapt Token field names** (`Test`, `DependsOn`, `RelatedTo`) to the real struct tags in `storage.go:19-57` (they may be `TestName`, or DependsOn may hold comma-lists — check; if comma-list, seed `DependsOn: "CBIN-101,CBIN-102"` and assert the split). Check whether `storage.Open` runs migrations; if not, use the same migrate call `cmd/canary/main.go:34-48` uses.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cmds/view/ -v`
Expected: FAIL — package/BuildView do not exist.

- [ ] **Step 3: Implement**

Create `internal/cmds/view/view.go`:

```go
// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package view aggregates everything known about one requirement — tokens,
// files, tests, dependencies, spec/plan, diagrams, ticket link — into one
// bounded, agent-friendly answer.
// CANARY: REQ=CBIN-204; FEATURE="RequirementView"; ASPECT=CLI; STATUS=IMPL; TEST=TestCANARY_CBIN_204_BuildView; UPDATED=2026-08-28
package view

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"go.devnw.com/canary/internal/sources"
	"go.devnw.com/canary/internal/storage"
)

// View is the aggregate answer for one requirement. (struct as in Interfaces block above)

// DefaultViewLimit bounds list sections (files, diagrams) by default; agents
// raise it with --limit when they need the full list.
const DefaultViewLimit = 10

// BuildView assembles the view from the index DB plus filesystem conventions.
func BuildView(dbPath, root, reqID string, limit int) (*View, error) {
	if limit <= 0 {
		limit = DefaultViewLimit
	}
	reg := sources.LoadFromRoot(root)
	reqID = reg.Normalize(strings.TrimSpace(reqID))

	db, err := storage.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open index db (run 'canary index' first): %w", err)
	}
	defer func() { _ = db.Close() }()

	tokens, err := db.GetTokensByReqID(reqID)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens found for %s (run 'canary index' to refresh)", reqID)
	}

	v := &View{ReqID: reqID, Statuses: map[string]int{}}
	if s, ok := reg.Resolve(reqID); ok {
		v.Source = s.Name
		v.TicketURL = reg.TicketURL(reqID)
	}

	fileSet, testSet, benchSet := map[string]struct{}{}, map[string]struct{}{}, map[string]struct{}{}
	depSet, blockSet, relSet := map[string]struct{}{}, map[string]struct{}{}, map[string]struct{}{}
	done := 0
	for _, tok := range tokens {
		v.Statuses[tok.Status]++
		if tok.Status == "TESTED" || tok.Status == "BENCHED" {
			done++
		}
		v.Features = append(v.Features, fmt.Sprintf("%s (%s, %s)", tok.Feature, tok.Aspect, tok.Status))
		if tok.FilePath != "" {
			fileSet[tok.FilePath] = struct{}{}
		}
		for _, s := range splitCSV(tok.Test) {
			testSet[s] = struct{}{}
		}
		for _, s := range splitCSV(tok.Bench) {
			benchSet[s] = struct{}{}
		}
		for _, s := range splitCSV(tok.DependsOn) {
			depSet[s] = struct{}{}
		}
		for _, s := range splitCSV(tok.Blocks) {
			blockSet[s] = struct{}{}
		}
		for _, s := range splitCSV(tok.RelatedTo) {
			relSet[s] = struct{}{}
		}
	}
	v.Completion = done * 100 / len(tokens)
	sort.Strings(v.Features)

	allFiles := sortedSet(fileSet)
	v.FilesTotal = len(allFiles)
	if len(allFiles) > limit {
		allFiles = allFiles[:limit]
	}
	v.Files = allFiles
	v.Tests = sortedSet(testSet)
	v.Benches = sortedSet(benchSet)
	v.DependsOn = sortedSet(depSet)
	v.Blocks = sortedSet(blockSet)
	v.RelatedTo = sortedSet(relSet)

	// Spec/plan by directory convention: .canary/specs/<REQ-ID>-<slug>/
	if matches, _ := filepath.Glob(filepath.Join(root, ".canary", "specs", reqID+"-*", "spec.md")); len(matches) > 0 {
		v.SpecPath = matches[0]
		if p := filepath.Join(filepath.Dir(matches[0]), "plan.md"); fileExists(p) {
			v.PlanPath = p
		}
	}

	if refs, err := db.GetRefsByReqID(reqID); err == nil {
		for _, r := range refs {
			v.Diagrams = append(v.Diagrams, fmt.Sprintf("%s:%d", r.FilePath, r.LineNumber))
		}
		if len(v.Diagrams) > limit {
			v.Diagrams = v.Diagrams[:limit]
		}
	}
	return v, nil
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func sortedSet(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
```

And the cobra command in the same file:

```go
// CreateViewCommand returns the `canary view` command.
func CreateViewCommand() *cobra.Command {
	var jsonOut bool
	var limit int
	cmd := &cobra.Command{
		Use:   "view <REQ-ID>",
		Short: "Full picture of one requirement: tokens, files, tests, deps, spec, diagrams, ticket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := BuildView(".canary/canary.db", ".", args[0], limit)
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				return enc.Encode(v) // compact: no SetIndent
			}
			printView(cmd, v, limit)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "compact JSON output")
	cmd.Flags().IntVar(&limit, "limit", DefaultViewLimit, "max entries per list section (raise when you need more)")
	return cmd
}

func printView(cmd *cobra.Command, v *View, limit int) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "%s — %d%% complete", v.ReqID, v.Completion)
	if v.Source != "" {
		fmt.Fprintf(out, " (source: %s)", v.Source)
	}
	fmt.Fprintln(out)
	if v.TicketURL != "" {
		fmt.Fprintf(out, "Ticket:   %s\n", v.TicketURL)
	}
	statuses := make([]string, 0, len(v.Statuses))
	for s, n := range v.Statuses {
		statuses = append(statuses, fmt.Sprintf("%s=%d", s, n))
	}
	sort.Strings(statuses)
	fmt.Fprintf(out, "Tokens:   %s\n", strings.Join(statuses, " "))
	fmt.Fprintf(out, "Features: %s\n", strings.Join(v.Features, "; "))
	fmt.Fprintf(out, "Files:    %s", strings.Join(v.Files, ", "))
	if v.FilesTotal > len(v.Files) {
		fmt.Fprintf(out, " … +%d more (use --limit %d)", v.FilesTotal-len(v.Files), v.FilesTotal)
	}
	fmt.Fprintln(out)
	if len(v.Tests) > 0 {
		fmt.Fprintf(out, "Tests:    %s\n", strings.Join(v.Tests, ", "))
	}
	if len(v.Benches) > 0 {
		fmt.Fprintf(out, "Benches:  %s\n", strings.Join(v.Benches, ", "))
	}
	if len(v.DependsOn) > 0 {
		fmt.Fprintf(out, "Depends:  %s\n", strings.Join(v.DependsOn, ", "))
	}
	if len(v.Blocks) > 0 {
		fmt.Fprintf(out, "Blocks:   %s\n", strings.Join(v.Blocks, ", "))
	}
	if len(v.RelatedTo) > 0 {
		fmt.Fprintf(out, "Related:  %s\n", strings.Join(v.RelatedTo, ", "))
	}
	if v.SpecPath != "" {
		fmt.Fprintf(out, "Spec:     %s\n", v.SpecPath)
	}
	if v.PlanPath != "" {
		fmt.Fprintf(out, "Plan:     %s\n", v.PlanPath)
	}
	if len(v.Diagrams) > 0 {
		fmt.Fprintf(out, "Diagrams: %s\n", strings.Join(v.Diagrams, ", "))
	}
}
```

Register in `cli/cmds.go` — add `view.CreateViewCommand(),` to the returned slice (follow the import + call pattern of the neighboring commands, e.g. `show`).

- [ ] **Step 4: Run tests + smoke**

Run: `go test ./internal/cmds/view/ -v && go build -o bin/canary ./cmd/canary && ./bin/canary index --root . >/dev/null 2>&1; ./bin/canary view CBIN-105`
Expected: tests PASS; the smoke command prints a bounded (~10-15 line) view of a real requirement.

- [ ] **Step 5: Commit**

```bash
git add internal/cmds/view/ cli/cmds.go
git commit -m "feat(cli): canary view — bounded one-call requirement picture"
```

---

### Task 11: CLI context caps — small defaults, explicit raise

**Files:**
- Modify: `internal/cmds/list/list.go` (default `--limit` 0 → 20; truncation hint; compact JSON), `internal/cmds/search/search.go` (add `--limit` default 20; compact JSON), `internal/cmds/grep/grep.go` (add `--limit` default 20), `internal/cmds/status/status.go` (add `--json`), `internal/cmds/files/files.go` (add `--json`), `internal/cmds/gap/gap.go` (default `--limit` 0 → 20), `internal/cmds/bug/list.go` (default `--limit` 0 → 20)
- Test: `internal/cmds/list/list_test.go` (create — only for the limit default; the rest are flag plumbing verified by build + smoke)

**Interfaces:**
- Consumes: `SearchTokens(keywords, limit)` from Task 7.
- Produces: uniform flag semantics for later doc task: `--limit N` (default 20, `0` = default not unlimited; `--limit -1` = explicitly unlimited), truncation hint `… +N more (use --limit)` on stderr-free stdout.

- [ ] **Step 1: Write the failing test**

Create `internal/cmds/list/list_test.go` asserting the flag default:

```go
package list

import "testing"

func TestCANARY_CBIN_205_ListDefaultLimitIsSmall(t *testing.T) {
	cmd := CreateListCommand() // match the real constructor name in list.go
	f := cmd.Flags().Lookup("limit")
	if f == nil {
		t.Fatal("list must have a --limit flag")
	}
	if f.DefValue != "20" {
		t.Errorf("--limit default = %s, want 20 (small-by-default for agent context)", f.DefValue)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cmds/list/ -v`
Expected: FAIL — DefValue is "0".

- [ ] **Step 3: Implement across the seven commands**

Apply this uniform pattern (shown for `list`; repeat for the others):

```go
// CANARY: REQ=CBIN-205; FEATURE="ContextCaps"; ASPECT=CLI; STATUS=IMPL; TEST=TestCANARY_CBIN_205_ListDefaultLimitIsSmall; UPDATED=2026-08-28
const defaultListLimit = 20

cmd.Flags().IntVar(&limit, "limit", defaultListLimit,
	"max results (default 20 to protect agent context; -1 = unlimited)")
```

Semantics normalization where the command runs:

```go
	if limit == 0 {
		limit = defaultListLimit
	}
	if limit < 0 {
		limit = 0 // storage layer treats 0 as its own default; use a big cap instead
	}
```

**Careful:** `storage.ListTokens` treats `limit == 0` as "no LIMIT clause". Map the CLI semantics explicitly: CLI `-1` → pass `0` to storage (unlimited); CLI `0`/unset → pass 20. Write a tiny shared helper in `internal/cmds/internal/utils/utils.go`:

```go
// EffectiveLimit maps CLI --limit semantics (0/unset => def, -1 => unlimited)
// to the storage layer's convention (0 => unlimited).
func EffectiveLimit(flag, def int) int {
	switch {
	case flag < 0:
		return 0
	case flag == 0:
		return def
	default:
		return flag
	}
}
```

Per command:
- `list.go`: default 20 via `EffectiveLimit`; after printing, if the DB returned exactly the limit, print `(showing %d; use --limit -1 for all)`. Change `--json` encoder to drop `SetIndent` (compact, one object per line via `json.NewEncoder(os.Stdout).Encode(tokens)`).
- `search.go`: add `--limit` (default 20) → `db.SearchTokens(kw, utils.EffectiveLimit(limit, 20))`; compact JSON same as list.
- `grep.go`: add `--limit` (default 20); pass through to the (Task 7-rewritten) `GrepTokens`; add a `limit` parameter to `GrepTokens` in `ops.go` and truncate its results.
- `status.go`: add `--json` printing a compact `{"req_id":..., "completion_pct":..., "by_status":{...}, "incomplete":[...]}` object built from the data it already computes; cap `incomplete` at 20 entries.
- `files.go`: add `--json` printing `{"req_id":...,"files":{"<aspect>":["path",...]}}` compactly.
- `gap.go`: change `--limit` default from 0 to 20 (line ~446).
- `bug/list.go`: change `--limit` default 0 → 20.

- [ ] **Step 4: Run everything**

Run: `go test ./internal/cmds/... -count=1 && go build -o bin/canary ./cmd/canary && ./bin/canary list | head -30 && ./bin/canary search scanner | head -10`
Expected: tests PASS; `list` prints at most 20 items + hint; `search` bounded.

- [ ] **Step 5: Commit**

```bash
git add internal/cmds/ ops.go
git commit -m "feat(cli): small-by-default output limits with explicit --limit raise"
```

---

### Task 12: MCP context caps + correctness fixes

**Files:**
- Modify: `mcp/tools.go` — `handleShow` (line 121), `handleStatus` (line 259), `handleSearch` (line 330), `handleNext` (line 379); `mcp/tools_extended.go` — `handleGrep` (line 371), `handleBugList` (line 426)
- Test: `mcp/mcp_test.go` (append — follow the existing test style there, which exercises handlers directly with a temp DB)

**Interfaces:**
- Consumes: `storage.SearchTokens(kw, limit)` (Task 7).
- Produces: every list-shaped MCP tool takes optional `limit` (default 20, hard cap 100) and returns `total` alongside the capped array so agents know to raise `limit`; `next` works with no args; `bug-list` returns only BUG-prefixed tokens.

- [ ] **Step 1: Write the failing tests**

Append to `mcp/mcp_test.go` (reuse its existing DB-seeding helpers; if none exist for these handlers, seed via `storage.Open` on a temp `.canary/canary.db` and `os.Chdir` to the temp root, matching how existing tests in that file arrange the hardcoded relative db path):

```go
func TestCANARY_CBIN_205_SearchCapped(t *testing.T) {
	// seed 30 matching tokens in a temp project dir (chdir into it)
	// call handleSearch with SearchParams{Keywords: "needle"}
	// assert len(result.Tokens) == 20 and result.Total == 30
}

func TestCANARY_CBIN_205_SearchLimitRaised(t *testing.T) {
	// same seed; SearchParams{Keywords: "needle", Limit: 100}
	// assert len(result.Tokens) == 30
}

func TestCANARY_CBIN_205_NextDefaultFindsWork(t *testing.T) {
	// seed one STUB token; call handleNext with empty NextParams
	// assert result.ReqID is the seeded token, NOT the "all complete" message
}

func TestCANARY_CBIN_205_BugListOnlyBugs(t *testing.T) {
	// seed one BUG-001 token and one CBIN-100 token
	// call handleBugList with empty params
	// assert result.Count == 1 and result.Bugs[0].ReqID == "BUG-001"
}
```

Write these as real tests (the sketches above name the exact behavior; expand them with the file's existing helper conventions — read `mcp/mcp_test.go` first).

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./mcp/ -run TestCANARY_CBIN_205 -v`
Expected: FAIL — search unbounded/no Total field; next returns nothing; bug-list returns both tokens.

- [ ] **Step 3: Implement**

Shared cap helper at the top of `mcp/tools.go`:

```go
// CANARY: REQ=CBIN-205; FEATURE="ContextCaps"; ASPECT=API; STATUS=IMPL; TEST=TestCANARY_CBIN_205_SearchCapped; UPDATED=2026-08-28
const (
	defaultToolLimit = 20  // small by default to protect agent context
	maxToolLimit     = 100 // explicit ceiling even when the agent asks for more
)

func capLimit(requested int) int {
	switch {
	case requested <= 0:
		return defaultToolLimit
	case requested > maxToolLimit:
		return maxToolLimit
	default:
		return requested
	}
}
```

- `SearchParams`/`GrepParams` gain `Limit int \`json:"limit,omitempty" jsonschema:"maximum results (default 20, max 100)"\`` (match the schema-tag style used by `ListParams`). Handlers: fetch `db.SearchTokens(kw, maxToolLimit+1)`, set `Total = len(all)` (clamped display), truncate to `capLimit(params.Limit)`. Add `Total int \`json:"total"\`` to `SearchResult`/`GrepResult`.
- `handleShow`: truncate `tokens` to `capLimit(0)`=20 unless `params.Limit` raises; add `Total`.
- `handleStatus`: keep the stats + completion; truncate the embedded `Tokens` array to 20 and add `Total` (stats remain computed over ALL tokens).
- `handleNext` fix (`tools.go:392`): replace `filters["status"] = "STUB,IMPL"` with two sequential queries like the CLI does (`next.go:198-238`): query `status=STUB` first, then `status=IMPL` if empty, order by priority.
- `handleBugList` fix (`tools_extended.go:440`): drop the bogus `req_id_prefix` filter; call `ListTokens(nil, "", "", 0)` then filter in Go with `strings.HasPrefix(tok.ReqID, "BUG-")`, apply `capLimit(params.Limit)`, set `Total`.
- `handleGrep`: honor the `field` param — when set, filter matches in Go by the named field (`req`→ReqID, `feature`→Feature, `aspect`→Aspect, `owner`→Owner) instead of ignoring it; otherwise current behavior. Bounded like search.

- [ ] **Step 4: Run tests**

Run: `go test ./mcp/ -count=1 -v`
Expected: PASS (new + existing).

- [ ] **Step 5: Commit**

```bash
git add mcp/
git commit -m "fix(mcp): bound all list-shaped tools, fix next default and bug-list filtering"
```

---

### Task 13: MCP `view` and `deps` tools

**Files:**
- Modify: `mcp/mcp.go` (register two tools), `mcp/tools_extended.go` (two handlers)
- Test: `mcp/mcp_test.go` (append)

**Interfaces:**
- Consumes: `view.BuildView` (Task 10), `specs` graph via the same helpers `internal/cmds/deps/deps.go` uses (`buildDependencyGraph` is unexported — either export a thin `deps.BuildGraph()` helper from `internal/cmds/deps` or reconstruct with `specs.NewGraphGenerator`; prefer exporting `func BuildGraph() (*specs.DependencyGraph, error)` from the deps package).
- Produces:
  - MCP tool `view`: params `{reqId string, limit?: int}` → returns the `view.View` struct as structured content + a 1-line text summary (`CBIN-105: 50% complete, 2 files, 1 test, 1 diagram, depends on CBIN-101`).
  - MCP tool `deps`: params `{reqId string, direction?: "forward"|"reverse"}` → returns `{reqId, direction, dependencies: []string, count}` (IDs only — no token payloads).

- [ ] **Step 1: Write the failing tests**

Append to `mcp/mcp_test.go`:

```go
func TestCANARY_CBIN_204_MCPView(t *testing.T) {
	// seed temp project (tokens for CBIN-105 incl. a diagram ref via ReplaceRefs)
	// call handleView(ctx, req, &ViewParams{ReqID: "CBIN-105"})
	// assert result.ReqID == "CBIN-105", result.Completion computed,
	// and the CallToolResult text content is a single line
}

func TestCANARY_CBIN_204_MCPViewUnknown(t *testing.T) {
	// handleView with ReqID "CBIN-999" → error (not a panic, not empty success)
}
```

(Expand with the file's conventions as in Task 12.)

- [ ] **Step 2: Run to verify they fail**

Run: `go test ./mcp/ -run TestCANARY_CBIN_204 -v`
Expected: FAIL — handleView undefined.

- [ ] **Step 3: Implement**

In `mcp/tools_extended.go`:

```go
// ViewParams identifies the requirement to aggregate.
type ViewParams struct {
	ReqID string `json:"reqId" jsonschema:"requirement ID, e.g. CBIN-105 or PLAT-4521"`
	Limit int    `json:"limit,omitempty" jsonschema:"max entries per list section (default 10)"`
}

// handleView returns the full bounded picture of one requirement.
// CANARY: REQ=CBIN-204; FEATURE="RequirementView"; ASPECT=API; STATUS=IMPL; TEST=TestCANARY_CBIN_204_MCPView; UPDATED=2026-08-28
func handleView(ctx context.Context, req *mcp.CallToolRequest, params *ViewParams) (*mcp.CallToolResult, *view.View, error) {
	v, err := view.BuildView(".canary/canary.db", ".", params.ReqID, params.Limit)
	if err != nil {
		return nil, nil, err
	}
	summary := fmt.Sprintf("%s: %d%% complete, %d files, %d tests, %d diagrams",
		v.ReqID, v.Completion, v.FilesTotal, len(v.Tests), len(v.Diagrams))
	if len(v.DependsOn) > 0 {
		summary += ", depends on " + strings.Join(v.DependsOn, ",")
	}
	if v.TicketURL != "" {
		summary += ", ticket " + v.TicketURL
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: summary}},
	}, v, nil
}

// DepsParams selects a requirement and traversal direction.
type DepsParams struct {
	ReqID     string `json:"reqId" jsonschema:"requirement ID"`
	Direction string `json:"direction,omitempty" jsonschema:"forward (what it depends on, default) or reverse (what depends on it)"`
}

// DepsResult carries dependency IDs only — deliberately no token payloads.
type DepsResult struct {
	ReqID        string   `json:"reqId"`
	Direction    string   `json:"direction"`
	Dependencies []string `json:"dependencies"`
	Count        int      `json:"count"`
}

func handleDeps(ctx context.Context, req *mcp.CallToolRequest, params *DepsParams) (*mcp.CallToolResult, *DepsResult, error) {
	graph, err := deps.BuildGraph() // exported helper added to internal/cmds/deps
	if err != nil {
		return nil, nil, err
	}
	gg := specs.NewGraphGenerator(nil)
	dir := params.Direction
	if dir == "" {
		dir = "forward"
	}
	var ids []string
	if dir == "reverse" {
		ids = reverseDependencies(graph, params.ReqID) // walk graph.Nodes for edges targeting ReqID
	} else {
		ids = gg.GetTransitiveDependencies(graph, params.ReqID)
	}
	res := &DepsResult{ReqID: params.ReqID, Direction: dir, Dependencies: ids, Count: len(ids)}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("%s %s deps: %d", params.ReqID, dir, len(ids))}},
	}, res, nil
}
```

For `reverse`, check whether `internal/cmds/deps/deps.go:191` (`createDepsReverseCommand`) already has a reverse-walk helper to export; reuse it rather than writing `reverseDependencies` fresh. Export `BuildGraph` from the deps package by renaming `buildDependencyGraph` → `BuildGraph` (keep a thin unexported alias if other in-package callers exist).

Register both in `mcp/mcp.go` next to the existing registrations (mirror lines 55-58's pattern):

```go
	mcp.AddTool(server, &mcp.Tool{Name: "view",
		Description: "Full picture of one requirement: status, files, tests, deps, spec/plan, diagrams, ticket URL. Use this FIRST instead of separate show/status/files calls."},
		handleView)
	mcp.AddTool(server, &mcp.Tool{Name: "deps",
		Description: "Dependency IDs for a requirement (forward or reverse). IDs only; follow up with view for detail."},
		handleDeps)
```

- [ ] **Step 4: Run tests**

Run: `go test ./mcp/ ./internal/cmds/deps/ -count=1 && go build ./...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add mcp/ internal/cmds/deps/
git commit -m "feat(mcp): view and deps tools for one-call hierarchical context"
```

---

### Task 14: Templates, init assets & agent docs

**Files:**
- Modify: `.canary/project.yaml` (live), `internal/cmds/init/base/project.yaml` (template — keep byte-identical structure to live file apart from `{{PROJECT_KEY}}` substitution), `.canary/AGENT_CONTEXT.md`, `internal/cmds/init/base/AGENT_CONTEXT.md`, `CLAUDE.md` (the `CANARY:START` block), `claude-plugin/skills/canary-token-format/SKILL.md`
- Create: `docs/user/ticket-sources-guide.md`
- Test: none (docs) — but `canary scan` must stay green.

**Interfaces:** none — documentation of Tasks 1-13's surface.

- [ ] **Step 1: project.yaml template gains a commented sources section**

Append to BOTH `.canary/project.yaml` and `internal/cmds/init/base/project.yaml` (template uses `{{PROJECT_KEY}}` where the live file uses `CBIN`):

```yaml
# Requirement-ID sources. Every prefix used in REQ= fields should be declared.
# type: flatfile (local series) | jira | github | gitlab
# url templates: {id} = full ID (PLAT-4521), {num} = numeric part (4521)
sources:
  - name: core
    type: flatfile
    key: "CBIN"
#  - name: platform
#    type: jira
#    key: "PLAT"
#    url: "https://company.atlassian.net/browse/{id}"
#  - name: app
#    type: gitlab
#    key: "GL"
#    url: "https://gitlab.com/group/project/-/issues/{num}"
#  - name: oss
#    type: github
#    key: "GH"
#    url: "https://github.com/owner/repo/issues/{num}"
```

- [ ] **Step 2: Create `docs/user/ticket-sources-guide.md`**

Short guide (~80 lines): sources config reference (the yaml above), how REQ= fields reference tickets (`REQ=PLAT-4521` — same token grammar, no new syntax), what changes in output (`status.json` gains `source`/`ticket_url`, `canary view` prints `Ticket:`), verification (`✅ PLAT-4521` claims work in GAP_ANALYSIS.md), mermaid references (IDs inside ```` ```mermaid ```` blocks are indexed; `canary deps graph --format mermaid` emits clickable diagrams), and the note that external IDs are never padded.

- [ ] **Step 3: Update agent docs (keep them SHORT — these are loaded into agent context)**

In `.canary/AGENT_CONTEXT.md` and `internal/cmds/init/base/AGENT_CONTEXT.md`, in the commands section add exactly:

```markdown
- `canary view <REQ-ID>` — **use this first**: one bounded call returns status, files, tests, deps, spec/plan paths, diagrams, ticket URL. `--json` for structured. Avoid separate show/status/files calls and avoid grep.
- All list commands default to 20 results; raise with `--limit N` (or `--limit -1` for all) only when needed.
```

In `CLAUDE.md` (both the top copy and the `CANARY:START` block — they are intentionally duplicated) add one line to the low-context section:

```markdown
- **Requirement lookup:** run `canary view <REQ-ID>` (one bounded call: files, tests, deps, diagrams, ticket). Do not grep the tree for requirement context.
```

In `claude-plugin/skills/canary-token-format/SKILL.md`, add a short "Ticket-system IDs" subsection: REQ= accepts any configured source key (`REQ=PLAT-4521` for JIRA, `REQ=GH-7` for GitHub, `REQ=GL-88` for GitLab); configured in `.canary/project.yaml` `sources:`; external IDs are written verbatim (no zero-padding).

- [ ] **Step 4: Full verification sweep**

Run:
```bash
go build ./... && go vet ./... && go test ./... -count=1
go build -o bin/canary ./cmd/canary
./bin/canary scan --root . --out status.json          # expect: CANARY_SCAN tokens=N ... (exit 0)
./bin/canary scan --root . --verify GAP_ANALYSIS.md   # expect: CANARY_VERIFY_OK or pre-existing failures only
./bin/canary index --root . && ./bin/canary view CBIN-201
```
Expected: all green; `view CBIN-201` shows the new TicketSources tokens with their tests.

- [ ] **Step 5: Commit**

```bash
git add .canary/ internal/cmds/init/base/ CLAUDE.md claude-plugin/ docs/user/ticket-sources-guide.md
git commit -m "docs: ticket sources config, view-first agent guidance, small-limit defaults"
```

---

## Self-Review

1. **Spec coverage:** Ticket sources (JIRA/GitLab/GitHub) → Tasks 1-3, 7, 14 (config, registry, scan/verify/normalize, storage, docs). Flatfile still supported → Default()/FromProjectConfig fallbacks + frozen stdout contracts + existing test suites. Mermaid references → Task 4 (extraction), Task 6 (generation + click-through), Task 9 (DB index). Full requirement view tying code/tests/deps/diagrams/tickets → Tasks 10, 13. Context-bloat minimization with small defaults an agent can raise → Tasks 7, 11, 12 + Global Constraints bullet.
2. **Placeholder scan:** Task 12/13 Step 1 test bodies are deliberately behavioral sketches with exact assertions named, because `mcp/mcp_test.go`'s helper conventions must be read first — the assertions to encode are fully specified. Everything else contains full code.
3. **Type consistency:** `sources.Registry` methods (`Pattern/ClaimPattern/Resolve/TicketURL/Normalize/Sources`) used in Tasks 3, 4, 6, 9, 10 match Task 2's definitions. `storage.Ref`+`ReplaceRefs`+`GetRefsByReqID` (Task 9) match Task 10/13 usage. `view.BuildView(dbPath, root, reqID, limit)` (Task 10) matches Task 13. `SearchTokens(kw, limit)` (Task 7) matches Tasks 11/12. Field names on `storage.Token` are flagged for verification against the real struct in Tasks 7 and 10.

### Task 15: Promote `internal/` packages to importable `pkg/` tree

**Files:**
- Move: every directory under `internal/` → `pkg/` (`git mv internal/canaryscan pkg/canaryscan` etc. for: canaryscan, cmds, config, docs, embedded, gap, matcher, migrate, prompts, reqid, sources, specs, storage). `internal/cmds/internal/utils` stays nested (`pkg/cmds/internal/utils`).
- Modify: every `.go` file importing `go.devnw.com/canary/internal/...` → `go.devnw.com/canary/pkg/...`.
- Test: no new tests — the full existing suite is the gate.

**Why `pkg/` and not top-level:** `internal/docs` collides with the markdown `docs/` directory and `internal/embedded` collides with the root `embedded/` directory. `pkg/` is outside `internal/`, so all packages become importable by other systems, which is the requirement.

- [ ] **Step 1: Move and rewrite imports**

```bash
mkdir -p pkg
for d in internal/*/; do git mv "$d" "pkg/$(basename $d)"; done
grep -rl 'go.devnw.com/canary/internal/' --include='*.go' . | xargs sed -i 's|go.devnw.com/canary/internal/|go.devnw.com/canary/pkg/|g'
```

Also grep non-Go references: `grep -rn "canary/internal/" --include='*.md' --include='*.yaml' --include='*.yml' .` — update stale doc references in CLAUDE.md/docs if they name import paths (prose mentions of file paths in historical docs may stay).

Check `//go:embed` directives still resolve (embedded assets moved with their packages — `pkg/cmds/init/base/**`, `pkg/storage/migrations/*.sql`, `pkg/prompts/...`).

- [ ] **Step 2: Verify**

Run: `go build ./... && go vet ./... && go test ./... -count=1`
Expected: all green, zero remaining `canary/internal` imports (`grep -rn "canary/internal" --include='*.go' .` is empty).

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "refactor: promote internal packages to importable pkg/ tree"
```

---

### Task 16: Module rename to devnw.dev/canary + dependency updates

**Files:**
- Modify: `go.mod` (module line), every `.go` import of `go.devnw.com/canary/...` → `devnw.dev/canary/...`, `go.sum` (via tidy).
- Modify: non-Go references to the module path (`CLAUDE.md`, `README*.md`, docs, plugin manifests) — `grep -rn "go.devnw.com/canary" .` and update remaining hits that describe the import path.

- [ ] **Step 1: Rename module**

```bash
go mod edit -module devnw.dev/canary
grep -rl 'go.devnw.com/canary' --include='*.go' . | xargs sed -i 's|go.devnw.com/canary|devnw.dev/canary|g'
```

- [ ] **Step 2: Update all dependencies**

```bash
go get -u ./... && go mod tidy
```

If a major-version bump breaks the build, pin that one dependency back (`go get dep@<previous>`) and note it in the commit message rather than fighting an upstream API migration in this task.

- [ ] **Step 3: Verify**

Run: `go build ./... && go vet ./... && go test ./... -count=1`
Expected: all green; `grep -rn "go.devnw.com/canary" --include='*.go' .` empty.

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "feat!: rename module to devnw.dev/canary and update dependencies"
```

---

### Task 17: Publish (controller-executed — not a subagent task)

Preconditions verified during planning: `origin` = git@gitlab.com:devnw/codepros/oss/canary.git, `gh` = github.com/devnw/canary (SSH). `~/.netrc` does NOT exist in this environment. `devnw.dev/canary?go-get=1` currently returns "not found" (no vanity meta); `go.devnw.com/canary` meta points at github.com/devnw/canary.

- [ ] Merge feature branch to main (after final whole-branch review), push to `origin` and `gh`.
- [ ] Tag the release: next semver from `git tag --list` (breaking module rename → major or minor bump per existing scheme), annotated tag, push to both remotes.
- [ ] Attempt public index: `GOPROXY=https://proxy.golang.org GONOSUMDB= go mod download devnw.dev/canary@<tag>` — EXPECTED TO FAIL until devnw.dev serves `<meta name="go-import" content="devnw.dev/canary git https://github.com/devnw/canary">`. Report this to the user as their infra step, with the exact meta tag needed.
- [ ] Confirm the GitHub repo is public (`gh repo view devnw/canary --json visibility` or curl the repo URL unauthenticated).

---

## Known adaptation points (implementers: verify, don't assume)

- `storage.DB`'s sqlx handle field name (Task 9), `testutil` helper names (Tasks 7, 9), `storage.Token` field names for Test/Bench/DependsOn/Blocks/RelatedTo (Tasks 7, 10), `DependencyGraph` field/constant names (Task 6), `mcp/mcp_test.go` seeding conventions (Tasks 12, 13), exact constructor name `CreateListCommand` (Task 11), whether `storage.Open` auto-migrates (Task 10).
