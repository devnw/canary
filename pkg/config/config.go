// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CP-257; FEATURE="ProjectConfig"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-16
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SourceConfig describes one requirement-ID source: a flatfile prefix or an
// external ticket system (jira, github, gitlab) whose keys appear in REQ= fields.
// CANARY: REQ=CP-267; FEATURE="TicketSources"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_201_LoadSources; UPDATED=2026-08-28
// CANARY: REQ=CP-279; FEATURE="TicketSync"; ASPECT=Storage; STATUS=TESTED; TEST=TestCANARY_CBIN_306_LoadSources_TicketSyncFields; UPDATED=2026-08-29
type SourceConfig struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // flatfile | jira | github | gitlab
	Key  string `yaml:"key"`  // ID prefix, e.g. "CBIN", "PLAT", "GH"
	URL  string `yaml:"url,omitempty"`
	// API is the REST base URL used by `canary ticket sync` when it differs
	// from URL (which is the human browse-link template). Empty means the
	// ticket-sync client falls back to its own default (e.g. JIRA_BASE_URL).
	API string `yaml:"api,omitempty"`
	// StatusMap overrides the default CANARY-status -> remote-status-name
	// mapping (STUB/IMPL/TESTED/BENCHED keys) for this source only.
	StatusMap map[string]string `yaml:"status_map,omitempty"`
}

// ProjectConfig represents the .canary/project.yaml configuration
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
