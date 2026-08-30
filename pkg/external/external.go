// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package external resolves whether a CANARY requirement ID is satisfied as
// an external dependency — one owned by a ticket-source (e.g. JIRA) or a
// peer project (see resolvePeer) rather than this project's own flatfile
// series. It answers purely from on-disk sources: the remote-status cache
// (.canary/remote-status.json, written by `canary ticket sync --apply` and
// `canary ticket status --refresh` in pkg/cmds/ticket) and configured
// peers' status.json files. This package performs NO network I/O, ever — a
// missing or stale cache, or an unreachable/malformed peer file, degrades
// to State=unknown (or falls through to the next source) rather than
// blocking or erroring.
// CANARY: REQ=ENG-3959; FEATURE="ExternalResolve"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_ENG_3959_Cache_RoundTrip,TestCANARY_ENG_3959_Cache_LoadMissing_NoError,TestCANARY_ENG_3959_Cache_LoadCorrupt_Error,TestCANARY_ENG_3959_Resolve_FlatfileSource_Unknown,TestCANARY_ENG_3959_Resolve_UnresolvedPrefix_Unknown,TestCANARY_ENG_3959_Resolve_NilRegistry_Unknown,TestCANARY_ENG_3959_Resolve_NoCacheFile_Unknown,TestCANARY_ENG_3959_Resolve_CachedDone_Satisfied,TestCANARY_ENG_3959_Resolve_CachedNotDone_Unsatisfied,TestCANARY_ENG_3959_Resolve_AbsentFromCache_Unknown,TestCANARY_ENG_3959_Resolve_CustomStatusMap_DoneSet,TestCANARY_ENG_3959_Resolve_StaleCache_DetailNote,TestCANARY_ENG_3959_Resolve_FreshCache_NoStaleNote,TestCANARY_ENG_3960_Resolution_IsExternal,TestCANARY_ENG_3960_Resolution_ShortDetail; UPDATED=2026-08-29
// CANARY: REQ=ENG-3961; FEATURE="PeerProjects"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_ENG_3961_Resolve_PeerSatisfied,TestCANARY_ENG_3961_Resolve_PeerUnsatisfied,TestCANARY_ENG_3961_Resolve_PeerNotFound,TestCANARY_ENG_3961_Resolve_PeerMissingFile,TestCANARY_ENG_3961_Resolve_PeerMalformedJSON,TestCANARY_ENG_3961_Resolve_PeerRelativeRoot,TestCANARY_ENG_3961_Resolve_PeerBeatsTicketCache,TestCANARY_ENG_3961_Resolve_UnknownPrefixResolvedByPeer,TestCANARY_ENG_3961_Resolve_LocalFlatfileNeverConsultsPeer,TestCANARY_ENG_3961_Resolve_SecondPeerFallsThroughFirst,TestCANARY_ENG_3961_Resolve_PeerLegacyExportUnknown,TestCANARY_ENG_3961_Resolve_PeerStaleExportUnknown; UPDATED=2026-08-30
package external

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devnw.dev/canary/pkg/config"
	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/ticket"
)

// CacheFileName is the remote-status cache's basename under .canary/.
const CacheFileName = "remote-status.json"

// staleAfter is the age past which a cache's Detail carries a staleness
// note. The cache is still used past this age — degradation is sacred, so a
// stale cache is strictly better than no cache.
const staleAfter = 24 * time.Hour

// refreshHint is appended to Detail whenever a requirement's remote status
// could not be determined from the cache (missing file, or the id simply
// isn't a cached key), pointing the caller at how to fix it.
const refreshHint = "no cached ticket status (run canary ticket status --refresh)"

// Resolution is the outcome of resolving one requirement ID as a possible
// external dependency.
type Resolution struct {
	ID     string
	State  string // satisfied | unsatisfied | unknown
	Detail string
}

// State values for Resolution.State.
const (
	StateSatisfied   = "satisfied"
	StateUnsatisfied = "unsatisfied"
	StateUnknown     = "unknown"
)

// IsExternal reports whether r describes an actual external (ticket-source)
// dependency, as opposed to a local/flatfile id or an unconfigured prefix —
// the case Resolve marks with Detail "not external". Callers (deps/next/view)
// use this to decide whether to apply external-dependency display/blocking
// rules at all.
func (r Resolution) IsExternal() bool {
	return r.Detail != "not external"
}

// ShortDetail returns a short, display-friendly rendering of r.Detail: the
// cached remote-status name (with any appended staleness note stripped) for
// satisfied/unsatisfied, or the fixed short note "no cached ticket status"
// for unknown — dropping the longer refresh-hint command suggestion so
// callers like `deps check` and `view` stay on one line.
func (r Resolution) ShortDetail() string {
	if r.State == StateUnknown {
		// A peer that owns the id and cannot vouch for it (stale or
		// export-less) is a different problem from an uncached ticket, and
		// naming the wrong fix sends the reader to the wrong command.
		if strings.HasPrefix(r.Detail, "peer:") || strings.HasPrefix(r.Detail, "peers:") {
			return r.Detail
		}
		return "no cached ticket status"
	}
	if i := strings.Index(r.Detail, "; "); i >= 0 {
		return r.Detail[:i]
	}
	return r.Detail
}

// Cache is the on-disk shape of .canary/remote-status.json: the last time a
// fetch succeeded, and the issue-key -> remote-status-name snapshot it
// produced.
type Cache struct {
	FetchedAt string            `json:"fetched_at"`
	Statuses  map[string]string `json:"statuses"`
}

// CachePath returns the remote-status cache path under root.
func CachePath(root string) string {
	return filepath.Join(root, ".canary", CacheFileName)
}

// LoadCache reads the cache at root's .canary/remote-status.json. A missing
// file is not an error — it returns (nil, nil), the "no cache yet" case
// callers must degrade gracefully on.
func LoadCache(root string) (*Cache, error) {
	data, err := os.ReadFile(CachePath(root))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("external: read cache: %w", err)
	}
	var c Cache
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("external: parse cache %s: %w", CachePath(root), err)
	}
	return &c, nil
}

// SaveCache writes statuses to root's .canary/remote-status.json with
// fetchedAt (converted to UTC) as fetched_at, creating .canary/ if needed.
// Written with mode 0600 — this file only ever holds ticket-system status
// names, but is treated as project-local state, not something to expose
// group/world-readable.
func SaveCache(root string, statuses map[string]string, fetchedAt time.Time) error {
	if statuses == nil {
		statuses = map[string]string{}
	}
	c := Cache{
		FetchedAt: fetchedAt.UTC().Format(time.RFC3339),
		Statuses:  statuses,
	}
	data, err := json.MarshalIndent(&c, "", "  ")
	if err != nil {
		return fmt.Errorf("external: encode cache: %w", err)
	}
	dir := filepath.Join(root, ".canary")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("external: create %s: %w", dir, err)
	}
	if err := os.WriteFile(CachePath(root), data, 0o600); err != nil {
		return fmt.Errorf("external: write cache: %w", err)
	}
	return nil
}

// doneSet computes the set of remote status names that count as "done" for
// src: its StatusMap's TESTED and BENCHED entries (canary's two "complete"
// statuses), falling back per-entry to ticket.DefaultStatusMap when src
// doesn't override that status. When neither resolves to a non-empty name
// (StatusMap explicitly blanks both), falls back to the historical default
// {"Done"} rather than an empty (permanently unsatisfiable) set.
func doneSet(src sources.Source) map[string]struct{} {
	set := map[string]struct{}{}
	for _, canaryStatus := range [...]string{"TESTED", "BENCHED"} {
		name, ok := src.StatusMap[canaryStatus]
		if !ok {
			name = ticket.DefaultStatusMap[canaryStatus]
		}
		if name != "" {
			set[name] = struct{}{}
		}
	}
	if len(set) == 0 {
		set["Done"] = struct{}{}
	}
	return set
}

// refTime returns CANARY_TEST_TIMESTAMP (RFC3339) when set and valid,
// otherwise the current time in UTC — the same test-pinning convention used
// throughout the codebase (see pkg/canaryscan.RefTimeFromEnv,
// pkg/upgrade.resolveToday).
func refTime() time.Time {
	if ts := os.Getenv("CANARY_TEST_TIMESTAMP"); ts != "" {
		if t, err := time.Parse(time.RFC3339, ts); err == nil {
			return t.UTC()
		}
	}
	return time.Now().UTC()
}

// stalenessNote returns a note to append to Detail when fetchedAt (parsed
// as RFC3339) is more than staleAfter old relative to refTime(); "" when
// fresh or unparsable (an unparsable fetched_at is a cache-format problem,
// not this call's to report).
func stalenessNote(fetchedAt string) string {
	t, err := time.Parse(time.RFC3339, fetchedAt)
	if err != nil {
		return ""
	}
	age := refTime().Sub(t)
	if age <= staleAfter {
		return ""
	}
	return fmt.Sprintf("cache stale (fetched %s, %s old)", t.Format(time.RFC3339), age.Round(time.Hour))
}

// appendNote joins detail and note with "; ", omitting empty parts.
func appendNote(detail, note string) string {
	if note == "" {
		return detail
	}
	if detail == "" {
		return note
	}
	return detail + "; " + note
}

// Resolve determines whether id is satisfied as an external dependency.
//
//  0. id is external (resolves to a non-flatfile source) OR has an unknown
//     prefix (resolves to no configured source at all, including when reg
//     is nil): every configured peer project is consulted, in declaration
//     order, BEFORE the ticket cache below. This is the inter-dependent-repos
//     case: a peer may own an id under a key this project's own sources:
//     list never heard of. Peers are drawn from reg.Peers() when reg != nil
//     (no additional config.Load); otherwise peers are loaded fresh from
//     root's .canary/project.yaml for backward compatibility. A known local
//     (flatfile) id never consults peers — it is unambiguously this project's
//     own. See resolvePeer for the peer-file contract.
//  1. id resolves to a flatfile source, or to no configured source at all
//     (and no peer claimed it): State=unknown, Detail="not external" —
//     callers treat this as a local requirement, not an external one.
//  2. id resolves to a ticket-source (jira/github/gitlab) and no peer
//     claimed it: the cache is read (never fetched — this package performs
//     no network I/O).
//     - No cache file, or the cache holds no entry for id: State=unknown,
//     Detail points at `canary ticket status --refresh`.
//     - id's cached status is in the source's done-set (its StatusMap's
//     TESTED/BENCHED names, defaulting to {"Done"}): State=satisfied,
//     Detail is the cached status name.
//     - id's cached status is present but not in the done-set:
//     State=unsatisfied, Detail is the cached status name.
//
// Whenever a cache file exists and its fetched_at is more than 24h old
// (relative to CANARY_TEST_TIMESTAMP when set, else now in UTC), a
// staleness note is appended to Detail — the cache is still used; a stale
// answer is strictly better than none (degradation is sacred).
func Resolve(id string, reg *sources.Registry, root string) Resolution {
	var src sources.Source
	var ok bool
	if reg != nil {
		src, ok = reg.Resolve(id)
	}
	localFlatfile := ok && src.Type == "flatfile"

	if !localFlatfile {
		var peers []config.PeerConfig
		if reg != nil {
			peers = reg.Peers()
		} else {
			// Backward compatibility: when reg == nil, load peers fresh from
			// config. A config that cannot be read is reported as unknown
			// rather than swallowed into an empty peer list: "this project
			// declares no peers" and "this project's peer list could not be
			// read" are different answers, and only one of them is safe to
			// act on.
			cfg, err := config.Load(root)
			if err != nil {
				return Resolution{ID: id, State: StateUnknown, Detail: "peers: config unreadable: " + err.Error()}
			}
			peers = cfg.Peers
		}
		if res, found := resolvePeer(id, root, peers); found {
			return res
		}
	}

	if !ok || src.Type == "flatfile" {
		return Resolution{ID: id, State: StateUnknown, Detail: "not external"}
	}

	cache, err := LoadCache(root)
	if err != nil || cache == nil {
		return Resolution{ID: id, State: StateUnknown, Detail: refreshHint}
	}

	var note string
	if cache.FetchedAt != "" {
		note = stalenessNote(cache.FetchedAt)
	}

	status, present := cache.Statuses[id]
	if !present {
		return Resolution{ID: id, State: StateUnknown, Detail: appendNote(refreshHint, note)}
	}

	state := StateUnsatisfied
	if _, done := doneSet(src)[status]; done {
		state = StateSatisfied
	}
	return Resolution{ID: id, State: state, Detail: appendNote(status, note)}
}

// peerReport is the minimal shape decoded from a peer project's
// status.json (the pkg/canaryscan.Report JSON produced by `canary scan
// --out status.json`). Only the fields Resolve needs — the requirement ids
// and feature statuses it declares, and the verification export that says
// which of them are actually proven — are decoded; everything else in the
// peer's report (summary, diagrams, migration notes, ...) is ignored.
type peerReport struct {
	Requirements []peerRequirement `json:"requirements"`
	// Verified is a pointer so a report that omits the key (one written by a
	// canary old enough to predate the export) is distinguishable from one
	// that ran the check and verified nothing. The first is unknown; the
	// second is an answer.
	Verified *[]string `json:"verified"`
}

// peerRequirement is one requirement entry in a peer's status.json.
type peerRequirement struct {
	ID       string        `json:"id"`
	Features []peerFeature `json:"features"`
}

// peerFeature is one feature entry within a peerRequirement; only Status is
// needed to decide satisfied/unsatisfied.
type peerFeature struct {
	Status string `json:"status"`
}

// statusRank orders CANARY statuses from least to most complete, used to
// pick the "worst" (least complete) status to report in an unsatisfied peer
// Detail. It is display only: a status is a declaration, and no declaration
// decides whether a peer's requirement is done — the peer's verification
// export does.
var statusRank = map[string]int{"STUB": 0, "IMPL": 1, "TESTED": 2, "BENCHED": 3}

// peerWorstStatus returns the worst (least complete, by statusRank) feature
// status declared by r, for display in an unsatisfied Detail. An unrecognized
// status falls back to the first one seen; a requirement with no features at
// all yields "".
func peerWorstStatus(r peerRequirement) string {
	worst := ""
	worstRank := -1
	for _, f := range r.Features {
		rank, known := statusRank[f.Status]
		switch {
		case !known && worst == "":
			worst = f.Status
		case known && (worstRank == -1 || rank < worstRank):
			worstRank = rank
			worst = f.Status
		}
	}
	return worst
}

// peerStaleNote returns a note when a peer's status.json was last written
// more than staleAfter ago (relative to refTime(), so CANARY_TEST_TIMESTAMP
// pins it), and "" when it is fresh. An export that old describes a tree the
// peer has almost certainly moved past, so it answers nothing.
func peerStaleNote(modTime time.Time) string {
	age := refTime().Sub(modTime.UTC())
	if age <= staleAfter {
		return ""
	}
	return fmt.Sprintf("stale export, %s old", age.Round(time.Hour))
}

// containsID reports whether list holds id.
func containsID(list []string, id string) bool {
	for _, v := range list {
		if v == id {
			return true
		}
	}
	return false
}

// peerStatusPath resolves peer.Root's status.json path, joining a relative
// Root against root (the local repo root passed to Resolve) — an absolute
// Root is used as-is.
func peerStatusPath(root string, peer config.PeerConfig) string {
	peerRoot := peer.Root
	if !filepath.IsAbs(peerRoot) {
		peerRoot = filepath.Join(root, peerRoot)
	}
	return filepath.Join(peerRoot, "status.json")
}

// resolvePeer consults each peer project in the peers list, in declaration
// order, for id — read-only, never writing or creating anything. For each
// peer:
//   - peer.Root/status.json missing or unreadable: soft skip, try the next
//     peer (an absent or unreachable peer must never error Resolve, and it
//     must not shadow an answer a later peer or the ticket cache can give).
//   - status.json present but not valid JSON in the expected shape: soft
//     skip, same as missing.
//   - id not found among the peer's requirements: try the next peer.
//   - id found, export written more than 24h ago (mtime vs refTime()):
//     unknown, Detail "peer:<name> (stale export, <age> old)". An export
//     that old describes a tree the peer has moved past.
//   - id found, report carries no "verified" key at all (written by a canary
//     predating the verification export): unknown, Detail "peer:<name> (no
//     verification export)". A peer that only publishes declarations cannot
//     say anything is done.
//   - id found and listed in the peer's "verified" export: satisfied, Detail
//     "peer:<name>".
//   - id found and NOT listed there: unsatisfied, Detail "peer:<name>
//     (<worst declared status>)" (the status is omitted when the requirement
//     declares no features).
//
// The first peer (in declaration order) that has an entry for id wins —
// later peers are not consulted once one has answered, including when its
// answer is unknown: a peer that owns an id and cannot vouch for it is the
// authority on that, and falling through to a ticket cache would answer a
// different question.
//
// Returns found=false when no peer in the list has an entry for id
// (including when peers is empty), letting Resolve fall through to its own
// ticket-cache logic.
//
// Reads each consulted peer's status.json fresh on every call. That is a
// known performance follow-up left for later if peer-heavy repos need it —
// Resolve is not currently called in a loop tight enough to matter.
// CANARY: REQ=ENG-3961; FEATURE="PeerProjects"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_ENG_3961_Resolve_PeerSatisfied,TestCANARY_ENG_3961_Resolve_PeerUnsatisfied,TestCANARY_ENG_3961_Resolve_PeerLegacyExportUnknown,TestCANARY_ENG_3961_Resolve_PeerStaleExportUnknown; UPDATED=2026-08-30
func resolvePeer(id, root string, peers []config.PeerConfig) (Resolution, bool) {
	for _, peer := range peers {
		path := peerStatusPath(root, peer)
		info, err := os.Stat(path)
		if err != nil {
			continue // missing/unreadable status.json: soft skip
		}
		data, err := os.ReadFile(path) //nolint:gosec // path derives from configured peer roots
		if err != nil {
			continue
		}
		var report peerReport
		if err := json.Unmarshal(data, &report); err != nil {
			continue // malformed status.json: soft skip
		}

		for _, r := range report.Requirements {
			if r.ID != id {
				continue
			}
			if note := peerStaleNote(info.ModTime()); note != "" {
				return Resolution{ID: id, State: StateUnknown, Detail: fmt.Sprintf("peer:%s (%s)", peer.Name, note)}, true
			}
			if report.Verified == nil {
				return Resolution{ID: id, State: StateUnknown, Detail: "peer:" + peer.Name + " (no verification export)"}, true
			}
			if containsID(*report.Verified, id) {
				return Resolution{ID: id, State: StateSatisfied, Detail: "peer:" + peer.Name}, true
			}
			detail := "peer:" + peer.Name
			if worst := peerWorstStatus(r); worst != "" {
				detail = fmt.Sprintf("peer:%s (%s)", peer.Name, worst)
			}
			return Resolution{ID: id, State: StateUnsatisfied, Detail: detail}, true
		}
	}
	return Resolution{}, false
}
