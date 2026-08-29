// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=ENG-4317; FEATURE="ProjectConfig"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-16
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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
	Scanner struct {
		ExcludePaths []string `yaml:"exclude_paths"`
	} `yaml:"scanner"`
	Verification struct {
		RequireTestField  bool `yaml:"require_test_field"`
		RequireBenchField bool `yaml:"require_bench_field"`
		StalenessDays     int  `yaml:"staleness_days"`
	} `yaml:"verification"`
	Agent struct {
		DefaultModel string `yaml:"default_model"`
	} `yaml:"agent"`
}

// Load reads and parses the project.yaml configuration file
func Load(rootDir string) (*ProjectConfig, error) {
	configPath := filepath.Join(rootDir, ".canary", "project.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Return default config if file doesn't exist
		if os.IsNotExist(err) {
			return &ProjectConfig{}, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &cfg, nil
}
