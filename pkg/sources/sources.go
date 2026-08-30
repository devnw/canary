// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package sources resolves requirement-ID prefixes to their origin: a local
// flatfile series (e.g. CBIN-105) or an external ticket system (JIRA, GitLab,
// GitHub) configured in .canary/project.yaml under `sources:`.
// CANARY: REQ=ENG-4322; FEATURE="TicketSources"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_201_RegistryPattern; UPDATED=2026-08-28
package sources

import (
	"fmt"
	"regexp"
	"strings"

	"devnw.dev/canary/pkg/config"
)

// Source is one configured requirement-ID origin.
type Source struct {
	Name string
	Type string // flatfile | jira | github | gitlab
	Key  string // uppercase ID prefix, e.g. "CBIN", "PLAT"
	URL  string // optional template; {id} = full ID, {num} = numeric part

	// API is the ticket system's REST base URL, used by `canary ticket
	// sync` when set (falls back to its own defaults, e.g. env vars,
	// otherwise). Empty for flatfile sources.
	API string
	// StatusMap overrides the default CANARY-status -> remote-status-name
	// mapping (STUB/IMPL/TESTED/BENCHED keys) for this source only.
	StatusMap map[string]string

	// Project is this source's ticket-system project key (e.g. a JIRA
	// project). Optional; empty means this source contributes no project
	// of its own to `canary ticket sync`.
	Project string
	// Destination marks this source as the target for create_issue
	// actions. At most one source may set this — see
	// Registry.DestinationSource.
	Destination bool
}

// keyRe is the requirement-ID prefix shape. It is pkg/config's pattern, not a
// second copy: the config file and the registry must agree on what a key is.
var keyRe = config.SourceKeyPattern

// Registry answers ID-shaped questions for a set of sources.
type Registry struct {
	list    []Source
	byKey   map[string]Source
	pattern *regexp.Regexp
	claim   *regexp.Regexp
	peers   []config.PeerConfig
}

// NewRegistry validates the sources and builds the combined ID pattern.
func NewRegistry(list []Source) (*Registry, error) {
	if len(list) == 0 {
		return nil, fmt.Errorf("sources: at least one source required")
	}
	// One rule set, defined in pkg/config, applied both when the config file
	// is loaded and here when a registry is built from any source list.
	specs := make([]config.SourceSpec, 0, len(list))
	for _, s := range list {
		specs = append(specs, config.SourceSpec{Name: s.Name, Type: s.Type, Key: s.Key, Destination: s.Destination})
	}
	if err := config.ValidateSources(specs); err != nil {
		return nil, err
	}
	byKey := make(map[string]Source, len(list))
	keys := make([]string, 0, len(list))
	for _, s := range list {
		byKey[s.Key] = s
		keys = append(keys, regexp.QuoteMeta(s.Key))
	}
	alt := strings.Join(keys, "|")
	return &Registry{
		list:    list,
		byKey:   byKey,
		pattern: regexp.MustCompile(`\b((?:` + alt + `)-\d+)\b`),
		claim:   regexp.MustCompile(`(?m)^\s*✅\s+((?:` + alt + `)-\d+)\b`),
		peers:   []config.PeerConfig{},
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
// project.key; an unset key falls back to Default() (CBIN). A configured but
// invalid key or source list is an error: a misconfigured registry silently
// downgraded to CBIN-only makes every non-CBIN requirement invisible.
// CANARY: REQ=ENG-4322; FEATURE="TicketSources"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_201_FromProjectConfigInvalidKeyError,TestCANARY_CBIN_201_FromProjectConfigMalformedSourceError; UPDATED=2026-08-30
func FromProjectConfig(cfg *config.ProjectConfig) (*Registry, error) {
	if cfg == nil {
		return Default(), nil
	}
	if len(cfg.Sources) == 0 {
		key := strings.TrimSpace(cfg.Project.Key)
		if key == "" {
			return Default(), nil
		}
		if !keyRe.MatchString(key) {
			return nil, fmt.Errorf("sources: project.key %q must be uppercase alphanumeric starting with a letter", key)
		}
		r, err := NewRegistry([]Source{{Name: "default", Type: "flatfile", Key: key}})
		if err != nil {
			return nil, err
		}
		r.peers = cfg.Peers
		return r, nil
	}
	list := make([]Source, 0, len(cfg.Sources))
	for _, s := range cfg.Sources {
		list = append(list, Source{
			Name:        s.Name,
			Type:        s.Type,
			Key:         s.Key,
			URL:         s.URL,
			API:         s.API,
			StatusMap:   s.StatusMap,
			Project:     s.Project,
			Destination: s.Destination,
		})
	}
	r, err := NewRegistry(list)
	if err != nil {
		return nil, err
	}
	r.peers = cfg.Peers
	return r, nil
}

// LoadFromRoot reads .canary/project.yaml under root and builds its registry.
// A missing config is legal and yields Default(); a malformed or invalid one
// is an error the caller must surface — degrading to Default() would silently
// drop every configured source.
// CANARY: REQ=ENG-4322; FEATURE="TicketSources"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_201_LoadFromRootNoCanaryDir,TestCANARY_CBIN_201_LoadFromRootInvalidConfigError; UPDATED=2026-08-30
func LoadFromRoot(root string) (*Registry, error) {
	cfg, err := config.Load(root)
	if err != nil {
		return nil, err
	}
	return FromProjectConfig(cfg)
}

// Pattern matches any configured requirement ID (captured in group 1).
func (r *Registry) Pattern() *regexp.Regexp { return r.pattern }

// ClaimPattern matches "✅ <ID>" gap-analysis claim lines (ID in group 1).
func (r *Registry) ClaimPattern() *regexp.Regexp { return r.claim }

// Sources returns the configured sources in declaration order.
func (r *Registry) Sources() []Source { return r.list }

// Peers returns the configured peer projects in declaration order.
func (r *Registry) Peers() []config.PeerConfig { return r.peers }

// SetPeers sets the peer projects for this registry (used for testing).
func (r *Registry) SetPeers(peers []config.PeerConfig) {
	r.peers = peers
}

// DestinationSource returns the source that `canary ticket sync` targets
// when promoting a flatfile requirement (create_issue actions): the source
// with Destination=true, if one is marked (NewRegistry rejects more than
// one); otherwise the first non-flatfile source in declaration order.
// Returns false when no source is marked and only flatfile sources are
// configured — there is nothing to promote a requirement to.
// CANARY: REQ=ENG-3958; FEATURE="TicketDestination"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_ENG_3958_DestinationSource_MarkedWins,TestCANARY_ENG_3958_DestinationSource_DefaultFirstNonFlatfile,TestCANARY_ENG_3958_DestinationSource_NoneWhenOnlyFlatfile,TestCANARY_ENG_3958_NewRegistry_DuplicateDestinationError,TestCANARY_ENG_3958_NewRegistry_FlatfileDestinationError; UPDATED=2026-08-29
func (r *Registry) DestinationSource() (Source, bool) {
	for _, s := range r.list {
		if s.Destination {
			return s, true
		}
	}
	for _, s := range r.list {
		if s.Type != "flatfile" {
			return s, true
		}
	}
	return Source{}, false
}

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
