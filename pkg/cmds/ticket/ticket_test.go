// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package ticket

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"devnw.dev/canary/pkg/external"
	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/storage"
	"devnw.dev/canary/pkg/ticket"
)

// chdir switches the process cwd to dir for the duration of the test,
// restoring it on cleanup. `canary ticket sync` resolves its registry and
// default --db path relative to the cwd, mirroring every other CANARY
// command (see pkg/cmds/view.CreateViewCommand).
func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
}

// seedProject writes .canary/project.yaml (one flatfile source CBIN plus
// one jira source PLAT) and a migrated, seeded canary.db with one flatfile
// token (CBIN-105) and one jira token (PLAT-42, IMPL) into a fresh temp
// root, returning the root.
func seedProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	canaryDir := filepath.Join(root, ".canary")
	if err := os.MkdirAll(canaryDir, 0o750); err != nil {
		t.Fatal(err)
	}

	yaml := `project:
  name: Demo
  key: CBIN
sources:
  - name: core
    type: flatfile
    key: CBIN
  - name: platform
    type: jira
    key: PLAT
`
	if err := os.WriteFile(filepath.Join(canaryDir, "project.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(canaryDir, "canary.db")
	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("MigrateDB: %v", err)
	}
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	toks := []*storage.Token{
		{ReqID: "CBIN-105", Feature: "Scanner", Aspect: "Engine", Status: "TESTED", FilePath: "scan.go"},
		{ReqID: "PLAT-42", Feature: "Sync", Aspect: "Engine", Status: "IMPL", FilePath: "sync.go"},
	}
	for _, tok := range toks {
		if err := db.UpsertToken(tok); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func execSync(t *testing.T, args ...string) (stdout string, err error) {
	t.Helper()
	// A fresh command tree per call (CreateTicketCommand mirrors
	// CreateDriftCommand/CreateViewCommand) keeps flag state isolated
	// between test cases.
	root := CreateTicketCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(append([]string{"sync"}, args...))
	err = root.Execute()
	return out.String(), err
}

func execStatus(t *testing.T, args ...string) (stdout string, err error) {
	t.Helper()
	root := CreateTicketCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(append([]string{"status"}, args...))
	err = root.Execute()
	return out.String(), err
}

func TestCANARY_CBIN_306_Sync_NoCredsPlanOnly(t *testing.T) {
	root := seedProject(t)
	chdir(t, root)
	t.Setenv("JIRA_BASE_URL", "")
	t.Setenv("JIRA_EMAIL", "")
	t.Setenv("JIRA_API_TOKEN", "")

	out, err := execSync(t, "--plan", ".canary/plan.json")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "CANARY_TICKET_PLAN actions=") || !strings.Contains(out, "applied=0 reason=plan_only") {
		t.Fatalf("output = %q", out)
	}
	data, rerr := os.ReadFile(filepath.Join(root, ".canary", "plan.json"))
	if rerr != nil {
		t.Fatalf("read plan file: %v", rerr)
	}
	var actions []ticket.Action
	if uerr := json.Unmarshal(data, &actions); uerr != nil {
		t.Fatalf("unmarshal plan: %v", uerr)
	}
	if len(actions) == 0 {
		t.Fatal("expected at least one proposed action (create_issue+remap for CBIN-105, transition for PLAT-42)")
	}
}

func TestCANARY_CBIN_306_Sync_NoCredsApplyDegradesGracefully(t *testing.T) {
	root := seedProject(t)
	chdir(t, root)
	t.Setenv("JIRA_BASE_URL", "")
	t.Setenv("JIRA_EMAIL", "")
	t.Setenv("JIRA_API_TOKEN", "")

	out, err := execSync(t, "--apply", "--plan", ".canary/plan.json")
	if err != nil {
		t.Fatalf("Execute must never error without credentials, got: %v", err)
	}
	if !strings.Contains(out, "applied=0 reason=no_credentials") {
		t.Fatalf("output = %q, want reason=no_credentials", out)
	}
	if !fileExists(filepath.Join(root, ".canary", "plan.json")) {
		t.Error("plan file should still be written under --apply without credentials")
	}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func TestCANARY_CBIN_306_Sync_ApplyWithCreds_EndToEnd(t *testing.T) {
	root := seedProject(t)
	chdir(t, root)

	var createCalls, transitionGETs, transitionPOSTs, searchCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/rest/api/3/issue":
			createCalls++
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"1","key":"CP-1"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/rest/api/3/search":
			searchCalls++
			_, _ = w.Write([]byte(`{"issues":[{"key":"PLAT-42","fields":{"status":{"name":"To Do"}}}],"total":1,"startAt":0,"maxResults":50}`))
		case r.Method == http.MethodGet && r.URL.Path == "/rest/api/3/issue/PLAT-42/transitions":
			transitionGETs++
			_, _ = w.Write([]byte(`{"transitions":[{"id":"21","name":"go","to":{"name":"In Progress"}}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/rest/api/3/issue/PLAT-42/transitions":
			transitionPOSTs++
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	t.Setenv("JIRA_BASE_URL", srv.URL)
	t.Setenv("JIRA_EMAIL", "agent@example.com")
	t.Setenv("JIRA_API_TOKEN", "sekret")

	planPath := filepath.Join(root, ".canary", "plan.json")
	out, err := execSync(t, "--apply", "--project", "CP", "--plan", ".canary/plan.json")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "CANARY_TICKET_SYNC created=1 transitioned=1 remap_pending=0") {
		t.Fatalf("output = %q", out)
	}
	if createCalls != 1 || searchCalls == 0 || transitionGETs != 1 || transitionPOSTs != 1 {
		t.Errorf("call counts: create=%d search=%d transitionGET=%d transitionPOST=%d", createCalls, searchCalls, transitionGETs, transitionPOSTs)
	}

	planData, rerr := os.ReadFile(planPath)
	if rerr != nil {
		t.Fatalf("read plan: %v", rerr)
	}
	var actions []ticket.Action
	if uerr := json.Unmarshal(planData, &actions); uerr != nil {
		t.Fatalf("unmarshal plan: %v", uerr)
	}
	foundRemap := false
	for _, a := range actions {
		if a.Type == "remap" && a.ReqID == "CBIN-105" {
			foundRemap = true
			if a.Issue != "CP-1" {
				t.Errorf("remap.Issue = %q, want CP-1 (filled from create_issue)", a.Issue)
			}
		}
	}
	if !foundRemap {
		t.Fatal("expected a remap action for CBIN-105 in the written plan")
	}

	mapData, merr := os.ReadFile(planPath + ".map.json")
	if merr != nil {
		t.Fatalf("read remap map: %v", merr)
	}
	var idMap map[string]string
	if uerr := json.Unmarshal(mapData, &idMap); uerr != nil {
		t.Fatalf("unmarshal remap map: %v", uerr)
	}
	if idMap["CBIN-105"] != "CP-1" {
		t.Errorf("remap map = %+v, want CBIN-105 -> CP-1", idMap)
	}
}

// seedProjectWithDestination mirrors seedProject but declares a single
// jira source ("platform") with an api: field (pointing at apiURL), a
// project: field, and destination: true, plus one flatfile token
// (CBIN-105) and no jira tokens — isolating the create_issue path so the
// test can assert it needs no --project flag.
func seedProjectWithDestination(t *testing.T, apiURL string) string {
	t.Helper()
	root := t.TempDir()
	canaryDir := filepath.Join(root, ".canary")
	if err := os.MkdirAll(canaryDir, 0o750); err != nil {
		t.Fatal(err)
	}

	yaml := `project:
  name: Demo
  key: CBIN
sources:
  - name: core
    type: flatfile
    key: CBIN
  - name: platform
    type: jira
    key: PLAT
    api: ` + apiURL + `
    project: CP
    destination: true
`
	if err := os.WriteFile(filepath.Join(canaryDir, "project.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(canaryDir, "canary.db")
	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("MigrateDB: %v", err)
	}
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	if err := db.UpsertToken(&storage.Token{ReqID: "CBIN-105", Feature: "Scanner", Aspect: "Engine", Status: "TESTED", FilePath: "scan.go"}); err != nil {
		t.Fatal(err)
	}
	return root
}

// TestCANARY_ENG_3958_Sync_ApplyWithDestinationSource_NoProjectFlag proves
// that --apply succeeds without a --project flag when the configured jira
// source carries "project:" and "destination: true" — the create_issue
// action is applied against that source's project.
func TestCANARY_ENG_3958_Sync_ApplyWithDestinationSource_NoProjectFlag(t *testing.T) {
	var createCalls int
	var createdProject string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/rest/api/3/issue":
			createCalls++
			var body struct {
				Fields struct {
					Project struct {
						Key string `json:"key"`
					} `json:"project"`
				} `json:"fields"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			createdProject = body.Fields.Project.Key
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"1","key":"CP-1"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/rest/api/3/search":
			_, _ = w.Write([]byte(`{"issues":[],"total":0,"startAt":0,"maxResults":50}`))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	root := seedProjectWithDestination(t, srv.URL)
	chdir(t, root)

	// Deliberately no JIRA_BASE_URL: BaseURL comes from the source's api:
	// field, mirroring the existing source-API-fallback tests.
	t.Setenv("JIRA_BASE_URL", "")
	t.Setenv("JIRA_EMAIL", "agent@example.com")
	t.Setenv("JIRA_API_TOKEN", "sekret")

	out, err := execSync(t, "--apply", "--plan", ".canary/plan.json")
	if err != nil {
		t.Fatalf("Execute: %v (no --project flag; destination source configures project=CP)", err)
	}
	if !strings.Contains(out, "CANARY_TICKET_SYNC created=1") {
		t.Fatalf("output = %q", out)
	}
	if createCalls != 1 {
		t.Errorf("createCalls = %d, want 1", createCalls)
	}
	if createdProject != "CP" {
		t.Errorf("created issue project = %q, want CP (from destination source config)", createdProject)
	}
}

// TestCANARY_ENG_3958_RemoteStatusForSources_MultiSourceMerge proves that
// remoteStatusForSources merges remote status across multiple jira sources
// and documents the merge semantics: when the SAME issue key appears in
// results from multiple sources (in source order), last-write-wins — the
// final source's status value persists.
func TestCANARY_ENG_3958_RemoteStatusForSources_MultiSourceMerge(t *testing.T) {
	var searchCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet || r.URL.Path != "/rest/api/3/search" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		searchCalls++
		switch r.URL.Query().Get("jql") {
		case "project=PLATPROJ":
			// Platform source returns both X-1 and X-2
			_, _ = w.Write([]byte(`{"issues":[{"key":"X-1","fields":{"status":{"name":"Done"}}},{"key":"X-2","fields":{"status":{"name":"In Progress"}}}],"total":2,"startAt":0,"maxResults":50}`))
		case "project=SECPROJ":
			// Security source returns X-1 (same key!) with DIFFERENT status
			// and X-3. When merged, X-1 should have Security's value (last-write-wins)
			_, _ = w.Write([]byte(`{"issues":[{"key":"X-1","fields":{"status":{"name":"To Do"}}},{"key":"X-3","fields":{"status":{"name":"Blocked"}}}],"total":2,"startAt":0,"maxResults":50}`))
		default:
			t.Errorf("unexpected jql: %s", r.URL.Query().Get("jql"))
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	reg, err := sources.NewRegistry([]sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "platform", Type: "jira", Key: "PLAT", Project: "PLATPROJ"},
		{Name: "security", Type: "jira", Key: "SEC", Project: "SECPROJ"},
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	client := &ticket.JiraClient{BaseURL: srv.URL, Email: "agent@example.com", Token: "sekret"}
	merged, err := remoteStatusForSources(client, reg, "")
	if err != nil {
		t.Fatalf("remoteStatusForSources: %v", err)
	}

	// Verify both project searches were made
	if searchCalls != 2 {
		t.Errorf("searchCalls = %d, want 2 (PLATPROJ + SECPROJ)", searchCalls)
	}

	// Verify merge results: X-1 should have "To Do" (Security source's value, last-write-wins)
	// X-2 and X-3 come from their respective sources
	want := map[string]string{"X-1": "To Do", "X-2": "In Progress", "X-3": "Blocked"}
	if len(merged) != len(want) {
		t.Fatalf("merged = %+v, want %+v", merged, want)
	}
	for k, v := range want {
		if merged[k] != v {
			t.Errorf("merged[%q] = %q, want %q", k, merged[k], v)
		}
	}
	// Specifically verify the merge semantics: X-1 from Security (last source)
	// overwrites X-1 from Platform
	if merged["X-1"] != "To Do" {
		t.Errorf("X-1 merge semantics: got %q, want Security's 'To Do' (last-write-wins)", merged["X-1"])
	}
}

// TestCANARY_ENG_3958_RemoteStatusForSources_SharedProjectSingleFetch proves that
// when multiple jira sources are configured with the SAME project value,
// remoteStatusForSources fetches that project exactly once, not repeatedly,
// deduplicating via the fetched map.
func TestCANARY_ENG_3958_RemoteStatusForSources_SharedProjectSingleFetch(t *testing.T) {
	var searchCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet || r.URL.Path != "/rest/api/3/search" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		searchCalls++
		jql := r.URL.Query().Get("jql")
		if jql == "project=SHARED" {
			_, _ = w.Write([]byte(`{"issues":[{"key":"SHARED-42","fields":{"status":{"name":"In Progress"}}}],"total":1,"startAt":0,"maxResults":50}`))
		} else {
			t.Errorf("unexpected jql: %s", jql)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	// Two jira sources pointing to the same project
	reg, err := sources.NewRegistry([]sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "alpha", Type: "jira", Key: "ALPHA", Project: "SHARED"},
		{Name: "beta", Type: "jira", Key: "BETA", Project: "SHARED"},
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	client := &ticket.JiraClient{BaseURL: srv.URL, Email: "agent@example.com", Token: "sekret"}
	merged, err := remoteStatusForSources(client, reg, "")
	if err != nil {
		t.Fatalf("remoteStatusForSources: %v", err)
	}

	// Verify exactly ONE fetch despite two sources pointing to SHARED
	if searchCalls != 1 {
		t.Errorf("searchCalls = %d, want 1 (deduped even with two sources for project=SHARED)", searchCalls)
	}

	// Verify the results are there
	if merged["SHARED-42"] != "In Progress" {
		t.Errorf("merged[SHARED-42] = %q, want In Progress", merged["SHARED-42"])
	}
}

func TestCANARY_CBIN_306_PrintActions_Bounded(t *testing.T) {
	var actions []ticket.Action
	for i := 0; i < 25; i++ {
		actions = append(actions, ticket.Action{Type: "transition", Issue: "PLAT-1", To: "Done", Source: "platform"})
	}
	cmd := CreateTicketCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	printActions(cmd, actions, 5)
	s := out.String()
	if strings.Count(s, "[transition]") != 5 {
		t.Errorf("expected 5 printed lines, got:\n%s", s)
	}
	if !strings.Contains(s, "+20 more") {
		t.Errorf("expected a truncation hint, got:\n%s", s)
	}
}

func TestCANARY_CBIN_306_Sync_ApplyNoCreds_ProjectRequired(t *testing.T) {
	root := seedProject(t)
	chdir(t, root)

	// No HTTP server needed — we must error before calling network
	var httpCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpCalls++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	t.Setenv("JIRA_BASE_URL", srv.URL)
	t.Setenv("JIRA_EMAIL", "agent@example.com")
	t.Setenv("JIRA_API_TOKEN", "sekret")

	// --apply with creds but no --project should error before making HTTP calls
	// if the plan contains create_issue or transition actions
	_, err := execSync(t, "--apply", "--plan", ".canary/plan.json")
	if err == nil {
		t.Fatal("expected error when --apply is set with credentials and plan contains mutable actions but --project is empty")
	}
	if !strings.Contains(err.Error(), "--project is required with --apply") {
		t.Errorf("error = %q, want '--project is required with --apply' in message", err.Error())
	}
	if httpCalls > 0 {
		t.Errorf("HTTP calls made despite --project being empty: %d calls", httpCalls)
	}
	// Verify no plan file was written (since we errored early)
	if fileExists(filepath.Join(root, ".canary", "plan.json")) {
		t.Error("plan file should not be written when --project validation fails")
	}
}

func TestCANARY_CBIN_306_Sync_PartialProgress(t *testing.T) {
	root := seedProject(t)
	chdir(t, root)

	var createCalls, transitionGETs, transitionPOSTs, searchCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/rest/api/3/issue":
			createCalls++
			// Fail the first create_issue call
			if createCalls == 1 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"errorMessages":["Invalid field"]}`))
				return
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"1","key":"CP-1"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/rest/api/3/search":
			searchCalls++
			_, _ = w.Write([]byte(`{"issues":[{"key":"PLAT-42","fields":{"status":{"name":"To Do"}}}],"total":1,"startAt":0,"maxResults":50}`))
		case r.Method == http.MethodGet && r.URL.Path == "/rest/api/3/issue/PLAT-42/transitions":
			transitionGETs++
			_, _ = w.Write([]byte(`{"transitions":[{"id":"21","name":"go","to":{"name":"In Progress"}}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/rest/api/3/issue/PLAT-42/transitions":
			transitionPOSTs++
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	t.Setenv("JIRA_BASE_URL", srv.URL)
	t.Setenv("JIRA_EMAIL", "agent@example.com")
	t.Setenv("JIRA_API_TOKEN", "sekret")

	planPath := filepath.Join(root, ".canary", "plan.json")
	out, err := execSync(t, "--apply", "--project", "CP", "--plan", ".canary/plan.json")

	// Should error but still print summary and persist the map with partial results
	if err == nil {
		t.Fatal("expected error due to create_issue failure, got nil")
	}

	// Output should contain the summary with errors count
	if !strings.Contains(out, "CANARY_TICKET_SYNC created=0 transitioned=1 remap_pending=") {
		t.Fatalf("output = %q, want CANARY_TICKET_SYNC with created=0 transitioned=1", out)
	}
	if !strings.Contains(out, "errors=1") {
		t.Fatalf("output = %q, want errors=1 in summary", out)
	}

	// Verify partial progress persisted to map: only PLAT-42 transitioned, CBIN-105 failed
	// The remap map should be empty since CBIN-105's create_issue failed (only successful remaps go into the map)
	mapData, merr := os.ReadFile(planPath + ".map.json")
	if merr != nil {
		t.Fatalf("read remap map: %v", merr)
	}
	var idMap map[string]string
	if uerr := json.Unmarshal(mapData, &idMap); uerr != nil {
		t.Fatalf("unmarshal remap map: %v", uerr)
	}
	// Since CBIN-105's create_issue failed, the remap map should be empty
	if len(idMap) != 0 {
		t.Errorf("remap map = %+v, want empty (create failed so no successful remaps)", idMap)
	}
}

// seedProjectWithSourceAPI mirrors seedProject but gives the jira "platform"
// source an api: field pointing at apiURL, instead of relying on
// JIRA_BASE_URL.
func seedProjectWithSourceAPI(t *testing.T, apiURL string) string {
	t.Helper()
	root := t.TempDir()
	canaryDir := filepath.Join(root, ".canary")
	if err := os.MkdirAll(canaryDir, 0o750); err != nil {
		t.Fatal(err)
	}

	yaml := `project:
  name: Demo
  key: CBIN
sources:
  - name: core
    type: flatfile
    key: CBIN
  - name: platform
    type: jira
    key: PLAT
    api: ` + apiURL + `
`
	if err := os.WriteFile(filepath.Join(canaryDir, "project.yaml"), []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(canaryDir, "canary.db")
	if err := storage.MigrateDB(dbPath, "all"); err != nil {
		t.Fatalf("MigrateDB: %v", err)
	}
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	toks := []*storage.Token{
		{ReqID: "CBIN-105", Feature: "Scanner", Aspect: "Engine", Status: "TESTED", FilePath: "scan.go"},
		{ReqID: "PLAT-42", Feature: "Sync", Aspect: "Engine", Status: "IMPL", FilePath: "sync.go"},
	}
	for _, tok := range toks {
		if err := db.UpsertToken(tok); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// TestCANARY_CBIN_306_Sync_ApplyWithSourceAPI_BaseURLFallback proves that
// when JIRA_BASE_URL is unset but the configured jira source carries an
// api: field, that field is used as the BaseURL fallback (env still
// supplies Email/Token) and the apply path actually reaches the JIRA
// client's httptest server — not just the plan-only path.
func TestCANARY_CBIN_306_Sync_ApplyWithSourceAPI_BaseURLFallback(t *testing.T) {
	var createCalls, transitionGETs, transitionPOSTs, searchCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/rest/api/3/issue":
			createCalls++
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"1","key":"CP-1"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/rest/api/3/search":
			searchCalls++
			_, _ = w.Write([]byte(`{"issues":[{"key":"PLAT-42","fields":{"status":{"name":"To Do"}}}],"total":1,"startAt":0,"maxResults":50}`))
		case r.Method == http.MethodGet && r.URL.Path == "/rest/api/3/issue/PLAT-42/transitions":
			transitionGETs++
			_, _ = w.Write([]byte(`{"transitions":[{"id":"21","name":"go","to":{"name":"In Progress"}}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/rest/api/3/issue/PLAT-42/transitions":
			transitionPOSTs++
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	root := seedProjectWithSourceAPI(t, srv.URL)
	chdir(t, root)

	// Deliberately no JIRA_BASE_URL: only email+token from env, BaseURL
	// must come from the source's api: field.
	t.Setenv("JIRA_BASE_URL", "")
	t.Setenv("JIRA_EMAIL", "agent@example.com")
	t.Setenv("JIRA_API_TOKEN", "sekret")

	out, err := execSync(t, "--apply", "--project", "CP", "--plan", ".canary/plan.json")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "CANARY_TICKET_SYNC created=1 transitioned=1 remap_pending=0") {
		t.Fatalf("output = %q", out)
	}
	if createCalls != 1 || searchCalls == 0 || transitionGETs != 1 || transitionPOSTs != 1 {
		t.Errorf("call counts: create=%d search=%d transitionGET=%d transitionPOST=%d (expected apply path to reach the httptest server via source.API fallback)", createCalls, searchCalls, transitionGETs, transitionPOSTs)
	}

	// The apply path must also have written the remote-status cache after
	// its successful fetch.
	cacheData, cerr := os.ReadFile(filepath.Join(root, ".canary", "remote-status.json"))
	if cerr != nil {
		t.Fatalf("read remote-status cache: %v", cerr)
	}
	var cache external.Cache
	if uerr := json.Unmarshal(cacheData, &cache); uerr != nil {
		t.Fatalf("unmarshal remote-status cache: %v", uerr)
	}
	if cache.Statuses["PLAT-42"] != "To Do" {
		t.Errorf("cache.Statuses[PLAT-42] = %q, want \"To Do\"", cache.Statuses["PLAT-42"])
	}
	if cache.FetchedAt == "" {
		t.Error("cache.FetchedAt is empty, want an RFC3339 timestamp")
	}
}

// TestCANARY_ENG_3959_Status_Refresh_NoCreds proves `canary ticket status
// --refresh` degrades gracefully without credentials: exit 0, the documented
// CANARY_TICKET_STATUS line, and the cache file left untouched (not created,
// not overwritten).
func TestCANARY_ENG_3959_Status_Refresh_NoCreds(t *testing.T) {
	root := seedProject(t)
	chdir(t, root)
	t.Setenv("JIRA_BASE_URL", "")
	t.Setenv("JIRA_EMAIL", "")
	t.Setenv("JIRA_API_TOKEN", "")

	out, err := execStatus(t, "--refresh")
	if err != nil {
		t.Fatalf("Execute must never error without credentials, got: %v", err)
	}
	if strings.TrimSpace(out) != "CANARY_TICKET_STATUS cached=0 reason=no_credentials" {
		t.Fatalf("output = %q", out)
	}
	if fileExists(filepath.Join(root, ".canary", "remote-status.json")) {
		t.Error("cache file must not be created when --refresh has no credentials")
	}
}

// TestCANARY_ENG_3959_Status_Refresh_NoCreds_PreservesExistingCache proves
// that --refresh without credentials leaves an existing cache file
// byte-for-byte untouched.
func TestCANARY_ENG_3959_Status_Refresh_NoCreds_PreservesExistingCache(t *testing.T) {
	root := seedProject(t)
	chdir(t, root)
	if err := external.SaveCache(root, map[string]string{"PLAT-1": "Done"}, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	before, berr := os.ReadFile(filepath.Join(root, ".canary", "remote-status.json"))
	if berr != nil {
		t.Fatal(berr)
	}

	t.Setenv("JIRA_BASE_URL", "")
	t.Setenv("JIRA_EMAIL", "")
	t.Setenv("JIRA_API_TOKEN", "")

	if _, err := execStatus(t, "--refresh"); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	after, aerr := os.ReadFile(filepath.Join(root, ".canary", "remote-status.json"))
	if aerr != nil {
		t.Fatal(aerr)
	}
	if string(before) != string(after) {
		t.Errorf("cache changed despite no credentials:\nbefore=%s\nafter=%s", before, after)
	}
}

// TestCANARY_ENG_3959_Status_Refresh_WithCreds proves `canary ticket status
// --refresh` fetches and caches without computing or applying a sync plan
// (no create_issue/transition calls reach the server), against an httptest
// JIRA server.
func TestCANARY_ENG_3959_Status_Refresh_WithCreds(t *testing.T) {
	root := seedProject(t)
	chdir(t, root)

	var searchCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/rest/api/3/search" {
			searchCalls++
			_, _ = w.Write([]byte(`{"issues":[{"key":"PLAT-42","fields":{"status":{"name":"Done"}}}],"total":1,"startAt":0,"maxResults":50}`))
			return
		}
		t.Errorf("unexpected request: %s %s (status --refresh must only fetch, never create/transition)", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	t.Setenv("JIRA_BASE_URL", srv.URL)
	t.Setenv("JIRA_EMAIL", "agent@example.com")
	t.Setenv("JIRA_API_TOKEN", "sekret")
	t.Setenv("CANARY_TEST_TIMESTAMP", "2026-08-29T00:00:00Z")

	out, err := execStatus(t, "--refresh", "--project", "PLAT")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "CANARY_TICKET_STATUS cached=1 fetched_at=2026-08-29T00:00:00Z") {
		t.Fatalf("output = %q", out)
	}
	if searchCalls != 1 {
		t.Errorf("searchCalls = %d, want 1", searchCalls)
	}

	cacheData, cerr := os.ReadFile(filepath.Join(root, ".canary", "remote-status.json"))
	if cerr != nil {
		t.Fatalf("read cache: %v", cerr)
	}
	var cache external.Cache
	if uerr := json.Unmarshal(cacheData, &cache); uerr != nil {
		t.Fatalf("unmarshal cache: %v", uerr)
	}
	if cache.Statuses["PLAT-42"] != "Done" {
		t.Errorf("cache.Statuses[PLAT-42] = %q, want Done", cache.Statuses["PLAT-42"])
	}
	if cache.FetchedAt != "2026-08-29T00:00:00Z" {
		t.Errorf("cache.FetchedAt = %q, want pinned CANARY_TEST_TIMESTAMP value", cache.FetchedAt)
	}
}

// TestCANARY_ENG_3959_Status_Plain_NoCache proves plain `canary ticket
// status` (no --refresh) reports an absent cache without touching the
// network or erroring.
func TestCANARY_ENG_3959_Status_Plain_NoCache(t *testing.T) {
	root := seedProject(t)
	chdir(t, root)

	out, err := execStatus(t)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.TrimSpace(out) != "CANARY_TICKET_STATUS cached=0 reason=no_cache" {
		t.Fatalf("output = %q", out)
	}
}

// TestCANARY_ENG_3959_Status_Plain_ReportsCache proves plain `canary ticket
// status` reports the cache's entry count, fetched_at, and age from disk
// with no network access.
func TestCANARY_ENG_3959_Status_Plain_ReportsCache(t *testing.T) {
	root := seedProject(t)
	chdir(t, root)
	if err := external.SaveCache(root, map[string]string{"PLAT-1": "Done", "PLAT-2": "To Do"}, time.Date(2026, 8, 28, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CANARY_TEST_TIMESTAMP", "2026-08-29T00:00:00Z") // 24h later

	out, err := execStatus(t)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "CANARY_TICKET_STATUS cached=2 fetched_at=2026-08-28T00:00:00Z age=24h0m0s") {
		t.Fatalf("output = %q", out)
	}
}
