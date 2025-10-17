// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-140; FEATURE="ProjectConfig"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-16
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProjectConfig represents the .canary/project.yaml configuration
type ProjectConfig struct {
	Project struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	} `yaml:"project"`
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
