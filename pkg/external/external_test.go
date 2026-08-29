// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package external

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"devnw.dev/canary/pkg/sources"
)

func newRegistry(t *testing.T, list []sources.Source) *sources.Registry {
	t.Helper()
	reg, err := sources.NewRegistry(list)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	return reg
}

// TestCANARY_ENG_3959_Cache_RoundTrip proves SaveCache followed by LoadCache
// round-trips fetched_at and statuses, and that the file is written with
// mode 0600.
func TestCANARY_ENG_3959_Cache_RoundTrip(t *testing.T) {
	root := t.TempDir()
	fetchedAt := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	statuses := map[string]string{"CP-240": "Done", "CP-241": "In Progress"}

	if err := SaveCache(root, statuses, fetchedAt); err != nil {
		t.Fatalf("SaveCache: %v", err)
	}

	path := CachePath(root)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat cache: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("cache file mode = %o, want 0600", perm)
	}

	cache, err := LoadCache(root)
	if err != nil {
		t.Fatalf("LoadCache: %v", err)
	}
	if cache == nil {
		t.Fatal("LoadCache returned nil cache after SaveCache")
	}
	if cache.FetchedAt != fetchedAt.UTC().Format(time.RFC3339) {
		t.Errorf("FetchedAt = %q, want %q", cache.FetchedAt, fetchedAt.UTC().Format(time.RFC3339))
	}
	if len(cache.Statuses) != 2 || cache.Statuses["CP-240"] != "Done" || cache.Statuses["CP-241"] != "In Progress" {
		t.Errorf("Statuses = %+v, want %+v", cache.Statuses, statuses)
	}
}

// TestCANARY_ENG_3959_Cache_LoadMissing_NoError proves that a missing cache
// file is not an error: LoadCache returns (nil, nil), matching the "no cache
// yet" degradation case.
func TestCANARY_ENG_3959_Cache_LoadMissing_NoError(t *testing.T) {
	root := t.TempDir()
	cache, err := LoadCache(root)
	if err != nil {
		t.Fatalf("LoadCache on missing file returned error: %v", err)
	}
	if cache != nil {
		t.Errorf("cache = %+v, want nil", cache)
	}
}

// TestCANARY_ENG_3959_Cache_LoadCorrupt_Error proves a malformed cache file
// surfaces an error rather than panicking or silently returning zero values.
func TestCANARY_ENG_3959_Cache_LoadCorrupt_Error(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".canary"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(CachePath(root), []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCache(root); err == nil {
		t.Fatal("LoadCache on corrupt file should return an error")
	}
}

// TestCANARY_ENG_3959_Resolve_FlatfileSource_Unknown proves that an ID
// resolving to a flatfile source is treated as "not external" — callers
// treat it as a local requirement.
func TestCANARY_ENG_3959_Resolve_FlatfileSource_Unknown(t *testing.T) {
	reg := newRegistry(t, []sources.Source{{Name: "core", Type: "flatfile", Key: "CBIN"}})
	res := Resolve("CBIN-105", reg, t.TempDir())
	if res.State != StateUnknown || res.Detail != "not external" {
		t.Errorf("Resolve = %+v, want State=unknown Detail=\"not external\"", res)
	}
}

// TestCANARY_ENG_3959_Resolve_UnresolvedPrefix_Unknown proves an ID whose
// prefix matches no configured source also resolves to "not external".
func TestCANARY_ENG_3959_Resolve_UnresolvedPrefix_Unknown(t *testing.T) {
	reg := newRegistry(t, []sources.Source{{Name: "core", Type: "flatfile", Key: "CBIN"}})
	res := Resolve("ZZZ-1", reg, t.TempDir())
	if res.State != StateUnknown || res.Detail != "not external" {
		t.Errorf("Resolve = %+v, want State=unknown Detail=\"not external\"", res)
	}
}

// TestCANARY_ENG_3959_Resolve_NilRegistry_Unknown proves a nil registry
// degrades to "not external" rather than panicking.
func TestCANARY_ENG_3959_Resolve_NilRegistry_Unknown(t *testing.T) {
	res := Resolve("CP-240", nil, t.TempDir())
	if res.State != StateUnknown || res.Detail != "not external" {
		t.Errorf("Resolve = %+v, want State=unknown Detail=\"not external\"", res)
	}
}

// TestCANARY_ENG_3959_Resolve_NoCacheFile_Unknown proves a ticket-source ID
// with no cache file on disk resolves to unknown, with a Detail pointing at
// 'canary ticket status --refresh'.
func TestCANARY_ENG_3959_Resolve_NoCacheFile_Unknown(t *testing.T) {
	reg := newRegistry(t, []sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "platform", Type: "jira", Key: "CP"},
	})
	res := Resolve("CP-240", reg, t.TempDir())
	if res.State != StateUnknown {
		t.Errorf("State = %q, want unknown", res.State)
	}
	if res.Detail == "" || !contains(res.Detail, "canary ticket status --refresh") {
		t.Errorf("Detail = %q, want a hint to run canary ticket status --refresh", res.Detail)
	}
}

// TestCANARY_ENG_3959_Resolve_CachedDone_Satisfied proves a ticket-source ID
// whose cached status is in the source's done-set (default "Done") resolves
// satisfied.
func TestCANARY_ENG_3959_Resolve_CachedDone_Satisfied(t *testing.T) {
	root := t.TempDir()
	if err := SaveCache(root, map[string]string{"CP-240": "Done"}, freshTime()); err != nil {
		t.Fatal(err)
	}
	reg := newRegistry(t, []sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "platform", Type: "jira", Key: "CP"},
	})
	res := Resolve("CP-240", reg, root)
	if res.State != StateSatisfied {
		t.Errorf("State = %q, want satisfied", res.State)
	}
	if res.Detail != "Done" {
		t.Errorf("Detail = %q, want \"Done\"", res.Detail)
	}
}

// TestCANARY_ENG_3959_Resolve_CachedNotDone_Unsatisfied proves a ticket
// present in the cache but not in the done-set resolves unsatisfied, with
// Detail set to the remote status.
func TestCANARY_ENG_3959_Resolve_CachedNotDone_Unsatisfied(t *testing.T) {
	root := t.TempDir()
	if err := SaveCache(root, map[string]string{"CP-240": "In Progress"}, freshTime()); err != nil {
		t.Fatal(err)
	}
	reg := newRegistry(t, []sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "platform", Type: "jira", Key: "CP"},
	})
	res := Resolve("CP-240", reg, root)
	if res.State != StateUnsatisfied {
		t.Errorf("State = %q, want unsatisfied", res.State)
	}
	if res.Detail != "In Progress" {
		t.Errorf("Detail = %q, want \"In Progress\"", res.Detail)
	}
}

// TestCANARY_ENG_3959_Resolve_AbsentFromCache_Unknown proves an ID not
// present as a key in an existing cache resolves unknown (distinct from "no
// cache file at all"), with a hint to refresh.
func TestCANARY_ENG_3959_Resolve_AbsentFromCache_Unknown(t *testing.T) {
	root := t.TempDir()
	if err := SaveCache(root, map[string]string{"CP-999": "Done"}, freshTime()); err != nil {
		t.Fatal(err)
	}
	reg := newRegistry(t, []sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "platform", Type: "jira", Key: "CP"},
	})
	res := Resolve("CP-240", reg, root)
	if res.State != StateUnknown {
		t.Errorf("State = %q, want unknown", res.State)
	}
	if !contains(res.Detail, "canary ticket status --refresh") {
		t.Errorf("Detail = %q, want a hint to run canary ticket status --refresh", res.Detail)
	}
}

// TestCANARY_ENG_3959_Resolve_CustomStatusMap_DoneSet proves the done-set is
// derived from the source's StatusMap (reversed for TESTED/BENCHED) rather
// than always defaulting to "Done" — a source that maps TESTED to "Closed"
// treats a cached "Closed" status as satisfied.
func TestCANARY_ENG_3959_Resolve_CustomStatusMap_DoneSet(t *testing.T) {
	root := t.TempDir()
	if err := SaveCache(root, map[string]string{"CP-240": "Closed"}, freshTime()); err != nil {
		t.Fatal(err)
	}
	reg := newRegistry(t, []sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "platform", Type: "jira", Key: "CP", StatusMap: map[string]string{"TESTED": "Closed", "BENCHED": "Closed"}},
	})
	res := Resolve("CP-240", reg, root)
	if res.State != StateSatisfied {
		t.Errorf("State = %q, want satisfied (custom StatusMap done-set includes Closed)", res.State)
	}

	// And the default "Done" status is now NOT in the done-set for this
	// source (its TESTED/BENCHED map only to "Closed").
	if err := SaveCache(root, map[string]string{"CP-241": "Done"}, freshTime()); err != nil {
		t.Fatal(err)
	}
	res = Resolve("CP-241", reg, root)
	if res.State != StateUnsatisfied {
		t.Errorf("State = %q, want unsatisfied (Done not in custom done-set)", res.State)
	}
}

// TestCANARY_ENG_3959_Resolve_StaleCache_DetailNote proves that a cache
// older than 24h gets a staleness note appended to Detail, computed against
// CANARY_TEST_TIMESTAMP.
func TestCANARY_ENG_3959_Resolve_StaleCache_DetailNote(t *testing.T) {
	root := t.TempDir()
	fetchedAt := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	if err := SaveCache(root, map[string]string{"CP-240": "Done"}, fetchedAt); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CANARY_TEST_TIMESTAMP", "2026-08-29T00:00:00Z") // 28 days later

	reg := newRegistry(t, []sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "platform", Type: "jira", Key: "CP"},
	})
	res := Resolve("CP-240", reg, root)
	if res.State != StateSatisfied {
		t.Errorf("State = %q, want satisfied (staleness doesn't change state)", res.State)
	}
	if !contains(res.Detail, "stale") {
		t.Errorf("Detail = %q, want a staleness note", res.Detail)
	}
}

// TestCANARY_ENG_3959_Resolve_FreshCache_NoStaleNote proves a cache fetched
// under 24h ago carries no staleness note.
func TestCANARY_ENG_3959_Resolve_FreshCache_NoStaleNote(t *testing.T) {
	root := t.TempDir()
	fetchedAt := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	if err := SaveCache(root, map[string]string{"CP-240": "Done"}, fetchedAt); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CANARY_TEST_TIMESTAMP", "2026-08-29T00:00:00Z") // 11.5h later

	reg := newRegistry(t, []sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "platform", Type: "jira", Key: "CP"},
	})
	res := Resolve("CP-240", reg, root)
	if res.Detail != "Done" {
		t.Errorf("Detail = %q, want plain \"Done\" with no staleness note", res.Detail)
	}
}

func freshTime() time.Time { return time.Now().UTC() }

func contains(s, substr string) bool { return strings.Contains(s, substr) }

// TestCANARY_ENG_3959_Resolve_DegenerateIMPLToDone pins the semantics the
// doneSet comment documents: a (degenerate but legal) StatusMap that sends a
// NON-done canary status to "Done" must not make "Done" satisfy the
// dependency unless TESTED/BENCHED also target it.
func TestCANARY_ENG_3959_Resolve_DegenerateIMPLToDone(t *testing.T) {
	root := t.TempDir()
	if err := SaveCache(root, map[string]string{"ENG-7": "Done"}, freshTime()); err != nil {
		t.Fatal(err)
	}
	reg := newRegistry(t, []sources.Source{{
		Name: "eng", Type: "jira", Key: "ENG",
		StatusMap: map[string]string{
			"IMPL":    "Done",     // degenerate: non-done status targets Done
			"TESTED":  "Verified", // done-set is {Verified, Shipped}
			"BENCHED": "Shipped",
		},
	}})
	res := Resolve("ENG-7", reg, root)
	if res.State != StateUnsatisfied {
		t.Fatalf("Done must NOT satisfy when TESTED/BENCHED target other names: got %s (%s)", res.State, res.Detail)
	}
}
