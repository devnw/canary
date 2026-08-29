// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package external

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"devnw.dev/canary/pkg/sources"
)

// peerTestRoot lays out a temp local repo root with .canary/project.yaml
// declaring the given peers block (raw YAML under "peers:", indented by the
// caller) plus any sources the caller wants alongside it. It returns the
// root.
func peerTestRoot(t *testing.T, peersYAML string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".canary"), 0o750); err != nil {
		t.Fatal(err)
	}
	projectYAML := "project:\n  name: \"demo\"\n  key: \"CBIN\"\n" + peersYAML
	if err := os.WriteFile(filepath.Join(root, ".canary", "project.yaml"), []byte(projectYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// writePeerStatus writes a canaryscan-shaped status.json at
// <peerDir>/status.json holding the given requirements (id -> feature
// statuses).
func writePeerStatus(t *testing.T, peerDir string, requirements map[string][]string) {
	t.Helper()
	if err := os.MkdirAll(peerDir, 0o750); err != nil {
		t.Fatal(err)
	}
	type feature struct {
		Feature string `json:"feature"`
		Aspect  string `json:"aspect"`
		Status  string `json:"status"`
	}
	type requirement struct {
		ID       string    `json:"id"`
		Features []feature `json:"features"`
	}
	type report struct {
		Requirements []requirement `json:"requirements"`
	}

	var rep report
	for id, statuses := range requirements {
		var feats []feature
		for _, s := range statuses {
			feats = append(feats, feature{Feature: "X", Aspect: "Engine", Status: s})
		}
		rep.Requirements = append(rep.Requirements, requirement{ID: id, Features: feats})
	}
	data, err := json.Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(peerDir, "status.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestCANARY_ENG_3961_Resolve_PeerSatisfied proves an id found in a peer's
// status.json with a TESTED feature resolves satisfied with Detail
// "peer:<name>".
func TestCANARY_ENG_3961_Resolve_PeerSatisfied(t *testing.T) {
	root := peerTestRoot(t, "peers:\n  - name: upstream\n    root: \"peer\"\n")
	writePeerStatus(t, filepath.Join(root, "peer"), map[string][]string{
		"UP-1": {"TESTED"},
	})
	reg := newRegistry(t, []sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "up", Type: "jira", Key: "UP"},
	})

	res := Resolve("UP-1", reg, root)
	if res.State != StateSatisfied {
		t.Fatalf("State = %q, want satisfied", res.State)
	}
	if res.Detail != "peer:upstream" {
		t.Errorf("Detail = %q, want \"peer:upstream\"", res.Detail)
	}
	if !res.IsExternal() {
		t.Error("peer-resolved id must be external")
	}
}

// TestCANARY_ENG_3961_Resolve_PeerUnsatisfied proves an id found in a peer's
// status.json with only STUB/IMPL features resolves unsatisfied with Detail
// carrying the worst (least complete) status.
func TestCANARY_ENG_3961_Resolve_PeerUnsatisfied(t *testing.T) {
	root := peerTestRoot(t, "peers:\n  - name: upstream\n    root: \"peer\"\n")
	writePeerStatus(t, filepath.Join(root, "peer"), map[string][]string{
		"UP-2": {"IMPL", "STUB"},
	})
	reg := newRegistry(t, []sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "up", Type: "jira", Key: "UP"},
	})

	res := Resolve("UP-2", reg, root)
	if res.State != StateUnsatisfied {
		t.Fatalf("State = %q, want unsatisfied", res.State)
	}
	if res.Detail != "peer:upstream (STUB)" {
		t.Errorf("Detail = %q, want \"peer:upstream (STUB)\"", res.Detail)
	}
}

// TestCANARY_ENG_3961_Resolve_PeerNotFound proves an id absent from every
// configured peer's status.json falls through to the existing "no cached
// ticket status" ticket-cache path rather than being treated as resolved by
// a peer.
func TestCANARY_ENG_3961_Resolve_PeerNotFound(t *testing.T) {
	root := peerTestRoot(t, "peers:\n  - name: upstream\n    root: \"peer\"\n")
	writePeerStatus(t, filepath.Join(root, "peer"), map[string][]string{
		"UP-1": {"TESTED"},
	})
	reg := newRegistry(t, []sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "up", Type: "jira", Key: "UP"},
	})

	res := Resolve("UP-99", reg, root)
	if res.State != StateUnknown {
		t.Fatalf("State = %q, want unknown", res.State)
	}
	if !contains(res.Detail, "canary ticket status --refresh") {
		t.Errorf("Detail = %q, want ticket-cache refresh hint (peer miss falls through)", res.Detail)
	}
}

// TestCANARY_ENG_3961_Resolve_PeerMissingFile proves a configured peer
// whose Root has no status.json at all is soft-skipped -- Resolve falls
// through to the next peer, then the ticket cache, without error.
func TestCANARY_ENG_3961_Resolve_PeerMissingFile(t *testing.T) {
	root := peerTestRoot(t, "peers:\n  - name: ghost\n    root: \"nowhere\"\n")
	reg := newRegistry(t, []sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "up", Type: "jira", Key: "UP"},
	})

	res := Resolve("UP-1", reg, root)
	if res.State != StateUnknown {
		t.Fatalf("State = %q, want unknown (missing peer file soft-skipped)", res.State)
	}
	if !contains(res.Detail, "canary ticket status --refresh") {
		t.Errorf("Detail = %q, want ticket-cache refresh hint", res.Detail)
	}
}

// TestCANARY_ENG_3961_Resolve_PeerMalformedJSON proves a peer's status.json
// that isn't valid JSON in the expected shape is soft-skipped, same as a
// missing file.
func TestCANARY_ENG_3961_Resolve_PeerMalformedJSON(t *testing.T) {
	root := peerTestRoot(t, "peers:\n  - name: broken\n    root: \"peer\"\n")
	if err := os.MkdirAll(filepath.Join(root, "peer"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "peer", "status.json"), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	reg := newRegistry(t, []sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "up", Type: "jira", Key: "UP"},
	})

	res := Resolve("UP-1", reg, root)
	if res.State != StateUnknown {
		t.Fatalf("State = %q, want unknown (malformed peer file soft-skipped)", res.State)
	}
	if !contains(res.Detail, "canary ticket status --refresh") {
		t.Errorf("Detail = %q, want ticket-cache refresh hint", res.Detail)
	}
}

// TestCANARY_ENG_3961_Resolve_PeerRelativeRoot proves a peer's Root is
// resolved relative to the local repo root (not the process cwd or the
// peer name).
func TestCANARY_ENG_3961_Resolve_PeerRelativeRoot(t *testing.T) {
	root := peerTestRoot(t, "peers:\n  - name: upstream\n    root: \"../sibling-repo\"\n")
	// root is <tmp>/<n>; place the peer beside it at <tmp>/sibling-repo so
	// filepath.Join(root, "../sibling-repo") resolves to it.
	siblingDir := filepath.Join(filepath.Dir(root), "sibling-repo")
	writePeerStatus(t, siblingDir, map[string][]string{
		"UP-1": {"BENCHED"},
	})
	reg := newRegistry(t, []sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "up", Type: "jira", Key: "UP"},
	})

	res := Resolve("UP-1", reg, root)
	if res.State != StateSatisfied {
		t.Fatalf("State = %q, want satisfied (relative peer root resolved against local repo root)", res.State)
	}
	if res.Detail != "peer:upstream" {
		t.Errorf("Detail = %q, want \"peer:upstream\"", res.Detail)
	}
}

// TestCANARY_ENG_3961_Resolve_PeerBeatsTicketCache proves the peer layer is
// consulted BEFORE the ticket cache: when both a peer and the local
// remote-status cache hold an entry for the same id, the peer's answer
// wins.
func TestCANARY_ENG_3961_Resolve_PeerBeatsTicketCache(t *testing.T) {
	root := peerTestRoot(t, "peers:\n  - name: upstream\n    root: \"peer\"\n")
	writePeerStatus(t, filepath.Join(root, "peer"), map[string][]string{
		"UP-1": {"IMPL"}, // peer says unsatisfied
	})
	if err := SaveCache(root, map[string]string{"UP-1": "Done"}, freshTime()); err != nil {
		t.Fatal(err) // ticket cache says satisfied
	}
	reg := newRegistry(t, []sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "up", Type: "jira", Key: "UP"},
	})

	res := Resolve("UP-1", reg, root)
	if res.State != StateUnsatisfied {
		t.Fatalf("State = %q, want unsatisfied (peer must win over ticket cache)", res.State)
	}
	if res.Detail != "peer:upstream (IMPL)" {
		t.Errorf("Detail = %q, want \"peer:upstream (IMPL)\"", res.Detail)
	}
}

// TestCANARY_ENG_3961_Resolve_UnknownPrefixResolvedByPeer proves the
// inter-dependent-repos case: an id whose prefix matches NO configured
// source at all (reg.Resolve returns ok=false) is still resolvable when a
// configured peer's status.json knows it -- the peer uses a different key
// entirely.
func TestCANARY_ENG_3961_Resolve_UnknownPrefixResolvedByPeer(t *testing.T) {
	root := peerTestRoot(t, "peers:\n  - name: upstream\n    root: \"peer\"\n")
	writePeerStatus(t, filepath.Join(root, "peer"), map[string][]string{
		"ZZZ-1": {"TESTED"},
	})
	// Only CBIN is configured locally -- ZZZ is a totally unknown prefix.
	reg := newRegistry(t, []sources.Source{{Name: "core", Type: "flatfile", Key: "CBIN"}})

	res := Resolve("ZZZ-1", reg, root)
	if res.State != StateSatisfied {
		t.Fatalf("State = %q, want satisfied (unknown-prefix id resolved by peer)", res.State)
	}
	if res.Detail != "peer:upstream" {
		t.Errorf("Detail = %q, want \"peer:upstream\"", res.Detail)
	}
	if !res.IsExternal() {
		t.Error("peer-resolved unknown-prefix id must be external")
	}
}

// TestCANARY_ENG_3961_Resolve_LocalFlatfileNeverConsultsPeer proves a known
// local (flatfile) id is never routed through the peer layer, even when a
// configured peer's status.json happens to hold an entry with the same id.
func TestCANARY_ENG_3961_Resolve_LocalFlatfileNeverConsultsPeer(t *testing.T) {
	root := peerTestRoot(t, "peers:\n  - name: upstream\n    root: \"peer\"\n")
	writePeerStatus(t, filepath.Join(root, "peer"), map[string][]string{
		"CBIN-1": {"TESTED"},
	})
	reg := newRegistry(t, []sources.Source{{Name: "core", Type: "flatfile", Key: "CBIN"}})

	res := Resolve("CBIN-1", reg, root)
	if res.State != StateUnknown || res.Detail != "not external" {
		t.Errorf("Resolve = %+v, want State=unknown Detail=\"not external\" (flatfile id must never consult peers)", res)
	}
}

// TestCANARY_ENG_3961_Resolve_SecondPeerFallsThroughFirst proves peers are
// consulted in declaration order and a miss on the first peer falls through
// to the second.
func TestCANARY_ENG_3961_Resolve_SecondPeerFallsThroughFirst(t *testing.T) {
	root := peerTestRoot(t, "peers:\n  - name: first\n    root: \"peer-a\"\n  - name: second\n    root: \"peer-b\"\n")
	writePeerStatus(t, filepath.Join(root, "peer-a"), map[string][]string{
		"UP-99": {"TESTED"}, // does not have UP-1
	})
	writePeerStatus(t, filepath.Join(root, "peer-b"), map[string][]string{
		"UP-1": {"BENCHED"},
	})
	reg := newRegistry(t, []sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "up", Type: "jira", Key: "UP"},
	})

	res := Resolve("UP-1", reg, root)
	if res.State != StateSatisfied {
		t.Fatalf("State = %q, want satisfied (fell through to second peer)", res.State)
	}
	if res.Detail != "peer:second" {
		t.Errorf("Detail = %q, want \"peer:second\"", res.Detail)
	}
}
