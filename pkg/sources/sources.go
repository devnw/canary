// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package sources resolves requirement-ID prefixes to their origin: a local
// flatfile series (e.g. CBIN-105) or an external ticket system (JIRA, GitLab,
// GitHub) configured in .canary/project.yaml under `sources:`.
// CANARY: REQ=CBIN-201; FEATURE="TicketSources"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_201_RegistryPattern; UPDATED=2026-08-28
package sources

import (
	"fmt"
	"regexp"
	"strings"

	"go.devnw.com/canary/pkg/config"
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
// project.key; if the key is empty or invalid, falls back to Default() (CBIN).
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
