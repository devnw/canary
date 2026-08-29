// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package ticket

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCANARY_CBIN_306_JiraClient_CreateIssue(t *testing.T) {
	var gotAuth, gotMethod, gotPath string
	var gotBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"10001","key":"CP-12","self":"https://example.atlassian.net/rest/api/3/issue/10001"}`))
	}))
	defer srv.Close()

	c := &JiraClient{BaseURL: srv.URL, Email: "agent@example.com", Token: "sekret"}
	key, err := c.CreateIssue("CP", "Story", "CBIN-105: Scanner", "Feature/Engine: TESTED (scan.go)")
	if err != nil {
		t.Fatalf("CreateIssue: %v", err)
	}
	if key != "CP-12" {
		t.Errorf("key = %q, want CP-12", key)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/rest/api/3/issue" {
		t.Errorf("path = %q, want /rest/api/3/issue", gotPath)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("agent@example.com:sekret"))
	if gotAuth != wantAuth {
		t.Errorf("Authorization = %q, want %q", gotAuth, wantAuth)
	}

	fields, ok := gotBody["fields"].(map[string]any)
	if !ok {
		t.Fatalf("request body missing fields: %+v", gotBody)
	}
	if fields["summary"] != "CBIN-105: Scanner" {
		t.Errorf("summary = %v, want CBIN-105: Scanner", fields["summary"])
	}
	proj, ok := fields["project"].(map[string]any)
	if !ok || proj["key"] != "CP" {
		t.Errorf("project = %v, want key=CP", fields["project"])
	}
	issuetype, ok := fields["issuetype"].(map[string]any)
	if !ok || issuetype["name"] != "Story" {
		t.Errorf("issuetype = %v, want name=Story", fields["issuetype"])
	}
	desc, ok := fields["description"].(map[string]any)
	if !ok {
		t.Fatalf("description missing or wrong shape: %+v", fields["description"])
	}
	if desc["type"] != "doc" {
		t.Errorf("description.type = %v, want doc (ADF)", desc["type"])
	}
	content, ok := desc["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("description.content missing/empty: %+v", desc)
	}
	para, ok := content[0].(map[string]any)
	if !ok || para["type"] != "paragraph" {
		t.Errorf("description.content[0] = %v, want a paragraph node", content[0])
	}
}

func TestCANARY_CBIN_306_JiraClient_CreateIssue_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errorMessages":["project is required"]}`))
	}))
	defer srv.Close()

	c := &JiraClient{BaseURL: srv.URL, Email: "a@b.com", Token: "t"}
	if _, err := c.CreateIssue("", "Story", "summary", "desc"); err == nil {
		t.Fatal("expected an error on a non-2xx response")
	}
}

func TestCANARY_CBIN_306_JiraClient_TransitionIssue_ResolvesIDByName(t *testing.T) {
	var postBody map[string]any
	var getPath, postPath, postMethod string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			getPath = r.URL.Path
			_, _ = w.Write([]byte(`{"transitions":[
				{"id":"11","name":"Start Progress","to":{"name":"In Progress"}},
				{"id":"31","name":"Done","to":{"name":"done"}}
			]}`))
		case http.MethodPost:
			postPath = r.URL.Path
			postMethod = r.Method
			_ = json.NewDecoder(r.Body).Decode(&postBody)
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer srv.Close()

	c := &JiraClient{BaseURL: srv.URL, Email: "a@b.com", Token: "t"}

	// Case-insensitive match against "done".
	if err := c.TransitionIssue("CP-12", "DONE"); err != nil {
		t.Fatalf("TransitionIssue: %v", err)
	}
	if getPath != "/rest/api/3/issue/CP-12/transitions" {
		t.Errorf("GET path = %q, want /rest/api/3/issue/CP-12/transitions", getPath)
	}
	if postPath != "/rest/api/3/issue/CP-12/transitions" || postMethod != http.MethodPost {
		t.Errorf("POST path/method = %q/%q, want /rest/api/3/issue/CP-12/transitions POST", postPath, postMethod)
	}
	transition, ok := postBody["transition"].(map[string]any)
	if !ok || transition["id"] != "31" {
		t.Errorf("posted transition = %v, want id=31 (matched \"done\" case-insensitively)", postBody["transition"])
	}
}

func TestCANARY_CBIN_306_JiraClient_TransitionIssue_NoMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"transitions":[{"id":"11","name":"Start Progress","to":{"name":"In Progress"}}]}`))
	}))
	defer srv.Close()

	c := &JiraClient{BaseURL: srv.URL, Email: "a@b.com", Token: "t"}
	if err := c.TransitionIssue("CP-12", "Done"); err == nil {
		t.Fatal("expected an error when no transition matches the target status name")
	}
}

func TestCANARY_CBIN_306_FetchRemoteStatus_Paged(t *testing.T) {
	var jqlSeen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jqlSeen = append(jqlSeen, r.URL.Query().Get("jql"))
		startAt := r.URL.Query().Get("startAt")
		w.Header().Set("Content-Type", "application/json")
		switch startAt {
		case "", "0":
			_, _ = w.Write([]byte(`{"issues":[{"key":"CP-1","fields":{"status":{"name":"To Do"}}},{"key":"CP-2","fields":{"status":{"name":"Done"}}}],"total":3,"startAt":0,"maxResults":2}`))
		default:
			_, _ = w.Write([]byte(`{"issues":[{"key":"CP-3","fields":{"status":{"name":"In Progress"}}}],"total":3,"startAt":2,"maxResults":2}`))
		}
	}))
	defer srv.Close()

	c := &JiraClient{BaseURL: srv.URL, Email: "a@b.com", Token: "t"}
	got, err := FetchRemoteStatus(c, "CP")
	if err != nil {
		t.Fatalf("FetchRemoteStatus: %v", err)
	}
	want := map[string]string{"CP-1": "To Do", "CP-2": "Done", "CP-3": "In Progress"}
	if len(got) != len(want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("got[%q] = %q, want %q", k, got[k], v)
		}
	}
	if len(jqlSeen) < 2 {
		t.Fatalf("expected at least 2 paged requests, got %d", len(jqlSeen))
	}
	for _, jql := range jqlSeen {
		if !strings.Contains(jql, "project") || !strings.Contains(jql, "CP") {
			t.Errorf("jql = %q, want it to reference project CP", jql)
		}
	}
}
