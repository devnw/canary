// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=ENG-4317; FEATURE="ProjectConfig"; ASPECT=Storage; STATUS=IMPL; UPDATED=2026-08-30
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultStaleDays is the staleness window (in days) used when
// verification.staleness_days is not configured. It lives here, in the single
// config type, so pkg/canaryscan can reference it without pkg/config having
// to import pkg/canaryscan.
const DefaultStaleDays = 30

// SourceConfig describes one requirement-ID source: a flatfile prefix or an
// external ticket system (jira, github, gitlab) whose keys appear in REQ= fields.
// CANARY: REQ=ENG-4322; FEATURE="TicketSources"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_201_LoadSources; UPDATED=2026-08-28
// CANARY: REQ=CP-279; FEATURE="TicketSync"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_306_LoadSources_TicketSyncFields; UPDATED=2026-08-29
// CANARY: REQ=ENG-3958; FEATURE="TicketDestination"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_ENG_3958_LoadSources_ProjectDestinationFields; UPDATED=2026-08-29
type SourceConfig struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // flatfile | jira | github | gitlab
	Key  string `yaml:"key"`  // ID prefix, e.g. "CBIN", "PLAT", "GH"
	URL  string `yaml:"url,omitempty"`
	// API is the REST base URL used by `canary ticket sync` when it differs
	// from URL (which is the human browse-link template). Precedence is
	// env > source.API: if JIRA_BASE_URL is set, it always wins; API is
	// only consulted as a fallback when JIRA_BASE_URL is unset. Email and
	// Token have no config-file fallback — they must come from
	// JIRA_EMAIL/JIRA_API_TOKEN regardless of what this field holds.
	API string `yaml:"api,omitempty"`
	// StatusMap overrides the default CANARY-status -> remote-status-name
	// mapping (STUB/IMPL/TESTED/BENCHED keys) for this source only.
	StatusMap map[string]string `yaml:"status_map,omitempty"`
	// Project is the ticket-system project key this source creates issues
	// in and fetches remote status for (e.g. a JIRA project key). Optional;
	// when unset, this source contributes no project of its own to `canary
	// ticket sync`.
	Project string `yaml:"project,omitempty"`
	// Destination marks this source as the target for create_issue actions
	// promoting flatfile requirements. At most one source may set this; see
	// Registry.DestinationSource in pkg/sources for the resolution rule
	// when no source is marked.
	Destination bool `yaml:"destination,omitempty"`
}

// PeerConfig is one peer project this repo is inter-dependent with: a
// sibling repo whose own `canary scan --out status.json` this project reads
// (read-only, never written to) to resolve requirement ids that peer owns —
// including ids under a prefix this project's own `sources:` list doesn't
// recognize at all. See pkg/external's peer-resolution layer.
// CANARY: REQ=ENG-3961; FEATURE="PeerProjects"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_ENG_3961_LoadPeers,TestCANARY_ENG_3961_LoadPeers_AbsentIsEmpty; UPDATED=2026-08-29
type PeerConfig struct {
	Name string `yaml:"name"`
	// Root is the peer project's root directory, resolved relative to
	// this project's own root when not absolute. Its status.json is read
	// from <Root>/status.json.
	Root string `yaml:"root"`
}

// ProjectConfig represents the .canary/project.yaml configuration
type ProjectConfig struct {
	Project struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Key         string `yaml:"key"`
	} `yaml:"project"`
	Sources []SourceConfig `yaml:"sources"`
	// Peers lists sibling projects consulted for requirement ids this
	// project doesn't own itself. Optional; empty when unconfigured.
	Peers        []PeerConfig `yaml:"peers"`
	Requirements struct {
		IDPattern string `yaml:"id_pattern"`
	} `yaml:"requirements"`
	Verification struct {
		StalenessDays int `yaml:"staleness_days"`
	} `yaml:"verification"`
	Agent struct {
		DefaultModel string `yaml:"default_model"`
	} `yaml:"agent"`
}

// StalenessDays is the resolved staleness window in days: the configured
// verification.staleness_days when set, else DefaultStaleDays.
func (c *ProjectConfig) StalenessDays() int {
	if c == nil || c.Verification.StalenessDays <= 0 {
		return DefaultStaleDays
	}
	return c.Verification.StalenessDays
}

// ProjectID is the resolved project identifier: project.key when set, else
// "default".
func (c *ProjectConfig) ProjectID() string {
	if c == nil {
		return "default"
	}
	if key := strings.TrimSpace(c.Project.Key); key != "" {
		return key
	}
	return "default"
}

// SourceKeyPattern is the required shape of a requirement-ID prefix: uppercase
// alphanumeric starting with a letter (e.g. "CBIN", "ENG", "GH2").
var SourceKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]*$`)

// validSourceTypes enumerates the source backends canary knows how to resolve.
var validSourceTypes = map[string]struct{}{"flatfile": {}, "jira": {}, "github": {}, "gitlab": {}}

// SourceSpec is the minimal source shape ValidateSources checks. It exists so
// the rules have exactly one implementation, shared by config parsing and by
// pkg/sources' registry construction (which validates its own Source type).
type SourceSpec struct {
	Name        string
	Type        string
	Key         string
	Destination bool
}

// ValidateSources enforces the source rules: every type must be known, every
// key well-formed and unique, and at most one source may be the ticket
// destination (which must not be a flatfile source).
func ValidateSources(specs []SourceSpec) error {
	seen := make(map[string]struct{}, len(specs))
	destinations := 0
	for _, s := range specs {
		if _, ok := validSourceTypes[s.Type]; !ok {
			return fmt.Errorf("sources: %q has unknown type %q", s.Name, s.Type)
		}
		if !SourceKeyPattern.MatchString(s.Key) {
			return fmt.Errorf("sources: %q key %q must be uppercase alphanumeric starting with a letter", s.Name, s.Key)
		}
		if _, dup := seen[s.Key]; dup {
			return fmt.Errorf("sources: duplicate key %q", s.Key)
		}
		if s.Destination {
			destinations++
			if s.Type == "flatfile" {
				return fmt.Errorf("sources: destination source %q must be a ticket source, not flatfile", s.Name)
			}
		}
		seen[s.Key] = struct{}{}
	}
	if destinations > 1 {
		return fmt.Errorf("sources: at most one source may set destination: true, found %d", destinations)
	}
	return nil
}

// specs projects the configured sources onto the shape ValidateSources checks.
func (c *ProjectConfig) specs() []SourceSpec {
	if c == nil {
		return nil
	}
	specs := make([]SourceSpec, 0, len(c.Sources))
	for _, s := range c.Sources {
		specs = append(specs, SourceSpec{Name: s.Name, Type: s.Type, Key: s.Key, Destination: s.Destination})
	}
	return specs
}

// ValidateProjectKey enforces the project.key shape rule: empty (unset) is
// legal, a non-empty key must match SourceKeyPattern. It is the single
// implementation of that rule, called both by ProjectConfig.validate (the
// config.Load path) and by pkg/sources.FromProjectConfig, which validates a
// project.key of its own -- direct construction of a *ProjectConfig (used in
// tests and by any caller that builds one without going through Load) never
// runs validate(), so FromProjectConfig cannot rely on Load having already
// checked it.
func ValidateProjectKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" || SourceKeyPattern.MatchString(key) {
		return nil
	}
	return fmt.Errorf("project.key %q must be uppercase alphanumeric starting with a letter", key)
}

// validate checks the parsed config's values.
func (c *ProjectConfig) validate() error {
	if c == nil {
		return nil
	}
	if err := ValidateProjectKey(c.Project.Key); err != nil {
		return err
	}
	if c.Verification.StalenessDays < 0 {
		return fmt.Errorf("verification.staleness_days must not be negative, got %d", c.Verification.StalenessDays)
	}
	if err := ValidateSources(c.specs()); err != nil {
		return err
	}
	for i, p := range c.Peers {
		if strings.TrimSpace(p.Root) == "" {
			return fmt.Errorf("peers[%d] (%q): root must not be empty", i, p.Name)
		}
	}
	return nil
}

// CANARY: REQ=ENG-4317; FEATURE="StrictProjectConfig"; ASPECT=Storage; STATUS=TESTED; TEST=TestLoadRejectsUnknownField,TestLoadRejectsDuplicateKey,TestLoadRejectsNegativeStaleness,TestLoadRejectsBadSourceType,TestLoadRejectsMultiDocumentYAML,TestAuditF19; UPDATED=2026-08-30
// Load reads, strictly parses and validates <rootDir>/.canary/project.yaml.
// Unknown fields, duplicate mapping keys, and invalid values are errors: a
// config that does not mean what it says must never be silently downgraded to
// defaults. A missing file is legal and yields an empty (unconfigured) config.
func Load(rootDir string) (*ProjectConfig, error) {
	configPath := filepath.Join(rootDir, ".canary", "project.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Unconfigured repos are legal; every other read error is not.
		if errors.Is(err, os.ErrNotExist) {
			return &ProjectConfig{}, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	if err := checkDuplicateKeys(data); err != nil {
		return nil, fmt.Errorf("parse %s: %w", configPath, err)
	}

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var cfg ProjectConfig
	if err := dec.Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("parse %s: %w", configPath, err)
	}

	// A second `---`-separated document must never be silently discarded: it
	// may hold configuration (including unknown fields) the author believes
	// is applied. Only a clean end-of-stream after the first document is
	// legal.
	var extra yaml.Node
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%s: multiple YAML documents are not allowed", configPath)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", configPath, err)
	}

	return &cfg, nil
}

// checkDuplicateKeys rejects repeated mapping keys anywhere in the document.
// yaml.v3 accepts them silently and keeps the last value, which quietly
// discards configuration the author believes is applied.
func checkDuplicateKeys(data []byte) error {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		// Decode below reports the syntax error with better context.
		return nil //nolint:nilerr // syntax errors are surfaced by the decoder
	}
	return walkDuplicateKeys(&root)
}

func walkDuplicateKeys(n *yaml.Node) error {
	if n == nil {
		return nil
	}
	if n.Kind == yaml.MappingNode {
		seen := make(map[string]int, len(n.Content)/2)
		for i := 0; i+1 < len(n.Content); i += 2 {
			k := n.Content[i]
			if first, dup := seen[k.Value]; dup {
				return fmt.Errorf("duplicate key %q at line %d (first defined at line %d)", k.Value, k.Line, first)
			}
			seen[k.Value] = k.Line
		}
	}
	for _, c := range n.Content {
		if err := walkDuplicateKeys(c); err != nil {
			return err
		}
	}
	return nil
}
