// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package ticket

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
	key, err := c.CreateIssue(context.Background(), "CP", "Story", "CBIN-105: Scanner", "Feature/Engine: TESTED (scan.go)")
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
	if _, err := c.CreateIssue(context.Background(), "", "Story", "summary", "desc"); err == nil {
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
	if err := c.TransitionIssue(context.Background(), "CP-12", "DONE"); err != nil {
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
	if err := c.TransitionIssue(context.Background(), "CP-12", "Done"); err == nil {
		t.Fatal("expected an error when no transition matches the target status name")
	}
}

// TestCANARY_CBIN_306_FetchRemoteStatus_Paged exercises the token-paginated
// POST /rest/api/3/search/jql loop: two pages linked by nextPageToken.
func TestCANARY_CBIN_306_FetchRemoteStatus_Paged(t *testing.T) {
	var jqlSeen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/rest/api/3/search/jql" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		jql, _ := body["jql"].(string)
		jqlSeen = append(jqlSeen, jql)
		w.Header().Set("Content-Type", "application/json")
		if _, ok := body["nextPageToken"]; !ok {
			_, _ = w.Write([]byte(`{"issues":[{"key":"CP-1","fields":{"status":{"name":"To Do"}}},{"key":"CP-2","fields":{"status":{"name":"Done"}}}],"nextPageToken":"page2"}`))
			return
		}
		_, _ = w.Write([]byte(`{"issues":[{"key":"CP-3","fields":{"status":{"name":"In Progress"}}}]}`))
	}))
	defer srv.Close()

	c := &JiraClient{BaseURL: srv.URL, Email: "a@b.com", Token: "t"}
	got, err := FetchRemoteStatus(context.Background(), c, "CP")
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
		if jql != `project = "CP"` {
			t.Errorf("jql = %q, want project = \"CP\" (validated + quoted)", jql)
		}
	}
}

// TestJiraSearchJQLPagination proves FetchRemoteStatus uses the token-based
// POST /rest/api/3/search/jql endpoint, follows nextPageToken across pages,
// quotes the project key inside the JQL, and never touches the deprecated
// offset endpoint (GET /rest/api/3/search?jql=...).
func TestJiraSearchJQLPagination(t *testing.T) {
	var deprecatedHit bool
	var bodies []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Any hit on the old offset endpoint (path or ?jql= query) fails.
		if r.URL.Path == "/rest/api/3/search" || strings.Contains(r.URL.RawQuery, "jql") {
			deprecatedHit = true
		}
		if r.Method != http.MethodPost || r.URL.Path != "/rest/api/3/search/jql" {
			t.Errorf("unexpected request: %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		bodies = append(bodies, body)
		w.Header().Set("Content-Type", "application/json")
		if _, ok := body["nextPageToken"]; !ok {
			_, _ = w.Write([]byte(`{"issues":[{"key":"CP-1","fields":{"status":{"name":"To Do"}}}],"nextPageToken":"tok-2"}`))
			return
		}
		_, _ = w.Write([]byte(`{"issues":[{"key":"CP-2","fields":{"status":{"name":"Done"}}}]}`))
	}))
	defer srv.Close()

	c := &JiraClient{BaseURL: srv.URL, Email: "a@b.com", Token: "t"}
	got, err := FetchRemoteStatus(context.Background(), c, "CP")
	if err != nil {
		t.Fatalf("FetchRemoteStatus: %v", err)
	}
	if deprecatedHit {
		t.Error("the deprecated /rest/api/3/search offset endpoint must never be used")
	}
	if len(bodies) != 2 {
		t.Fatalf("expected exactly 2 paged POSTs, got %d", len(bodies))
	}
	if got["CP-1"] != "To Do" || got["CP-2"] != "Done" {
		t.Errorf("got = %+v, want CP-1=To Do CP-2=Done", got)
	}
	if bodies[0]["jql"] != `project = "CP"` {
		t.Errorf("page 1 jql = %v, want project = \"CP\"", bodies[0]["jql"])
	}
	if _, ok := bodies[0]["nextPageToken"]; ok {
		t.Error("first page must not carry a nextPageToken")
	}
	if bodies[1]["nextPageToken"] != "tok-2" {
		t.Errorf("page 2 nextPageToken = %v, want tok-2", bodies[1]["nextPageToken"])
	}
}

// TestJiraSearchInvalidProjectKey proves an invalid project key is rejected
// before any request is made (no network call at all).
func TestJiraSearchInvalidProjectKey(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &JiraClient{BaseURL: srv.URL, Email: "a@b.com", Token: "t"}
	// A JQL-injection attempt in the project key must never reach the wire.
	if _, err := FetchRemoteStatus(context.Background(), c, `CP" OR "1"="1`); err == nil {
		t.Fatal("expected an error for an invalid project key")
	}
	if calls != 0 {
		t.Errorf("made %d request(s) for an invalid project key, want 0", calls)
	}
}

// TestJiraRetry429ThenSuccess covers the bounded retry policy: default
// backoff (250ms, 500ms), a small Retry-After honored, and a large
// Retry-After aborting immediately.
func TestJiraRetry429ThenSuccess(t *testing.T) {
	t.Run("default backoff 250ms then 500ms", func(t *testing.T) {
		var slept []time.Duration
		var calls int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			if calls <= 2 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		c := &JiraClient{BaseURL: srv.URL, Email: "a", Token: "t", Sleep: func(d time.Duration) { slept = append(slept, d) }}
		if err := c.do(context.Background(), http.MethodGet, "/x", nil, nil); err != nil {
			t.Fatalf("do: %v", err)
		}
		if calls != 3 {
			t.Errorf("calls = %d, want 3 (two 429s then success)", calls)
		}
		want := []time.Duration{250 * time.Millisecond, 500 * time.Millisecond}
		if len(slept) != len(want) {
			t.Fatalf("slept = %v, want %v", slept, want)
		}
		for i := range want {
			if slept[i] != want[i] {
				t.Errorf("slept[%d] = %v, want %v", i, slept[i], want[i])
			}
		}
	})

	t.Run("Retry-After 2 honored as 2s", func(t *testing.T) {
		var slept []time.Duration
		var calls int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			if calls == 1 {
				w.Header().Set("Retry-After", "2")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		c := &JiraClient{BaseURL: srv.URL, Email: "a", Token: "t", Sleep: func(d time.Duration) { slept = append(slept, d) }}
		if err := c.do(context.Background(), http.MethodGet, "/x", nil, nil); err != nil {
			t.Fatalf("do: %v", err)
		}
		if len(slept) != 1 || slept[0] != 2*time.Second {
			t.Errorf("slept = %v, want [2s] (Retry-After overrides backoff)", slept)
		}
	})

	t.Run("Retry-After 6 aborts immediately", func(t *testing.T) {
		var slept []time.Duration
		var calls int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			w.Header().Set("Retry-After", "6")
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer srv.Close()

		c := &JiraClient{BaseURL: srv.URL, Email: "a", Token: "t", Sleep: func(d time.Duration) { slept = append(slept, d) }}
		err := c.do(context.Background(), http.MethodGet, "/x", nil, nil)
		if err == nil {
			t.Fatal("expected an error when Retry-After exceeds the 5s ceiling")
		}
		if len(slept) != 0 {
			t.Errorf("slept = %v, want none (a >5s Retry-After must abort without sleeping)", slept)
		}
		if calls != 1 {
			t.Errorf("calls = %d, want 1 (no retry on an over-limit Retry-After)", calls)
		}
	})
}

// TestJiraNoRetryOn400 proves a non-retryable 4xx is returned at once, with
// no retries and no backoff.
func TestJiraNoRetryOn400(t *testing.T) {
	var slept []time.Duration
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	c := &JiraClient{BaseURL: srv.URL, Email: "a", Token: "t", Sleep: func(d time.Duration) { slept = append(slept, d) }}
	if err := c.do(context.Background(), http.MethodGet, "/x", nil, nil); err == nil {
		t.Fatal("expected an error on 400")
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (400 is never retried)", calls)
	}
	if len(slept) != 0 {
		t.Errorf("slept = %v, want none", slept)
	}
}

// TestJiraBodyCap proves an over-cap response body is rejected with the cap
// error, and the body itself never appears in that error.
func TestJiraBodyCap(t *testing.T) {
	huge := strings.Repeat("A", 3<<20) // 3 MiB, over the 2 MiB cap
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(huge))
	}))
	defer srv.Close()

	c := &JiraClient{BaseURL: srv.URL, Email: "a", Token: "t"}
	err := c.do(context.Background(), http.MethodGet, "/x", nil, nil)
	if err == nil {
		t.Fatal("expected an error for an over-cap response body")
	}
	if !strings.Contains(err.Error(), "2 MiB") {
		t.Errorf("error = %q, want it to mention the 2 MiB cap", err.Error())
	}
	if strings.Contains(err.Error(), "AAAA") {
		t.Errorf("error leaked the response body: %q", err.Error())
	}
}

// TestJiraErrorRedaction proves an error never carries the response body, a
// credential, or a URL query string (for example ?jql=...).
func TestJiraErrorRedaction(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"secret":"secret-token-xyz"}`))
	}))
	defer srv.Close()

	// No-op Sleep so the 5xx retry schedule does not actually wait.
	c := &JiraClient{BaseURL: srv.URL, Email: "agent@example.com", Token: "super-secret-token", Sleep: func(time.Duration) {}}
	err := c.do(context.Background(), http.MethodGet, "/rest/api/3/search?jql=project%3DCP", nil, nil)
	if err == nil {
		t.Fatal("expected an error on 500")
	}
	msg := err.Error()
	if strings.Contains(msg, "secret-token-xyz") {
		t.Errorf("error leaked the response body: %q", msg)
	}
	if strings.Contains(msg, "super-secret-token") {
		t.Errorf("error leaked the token: %q", msg)
	}
	if strings.Contains(msg, "?jql=") || strings.Contains(msg, "jql") {
		t.Errorf("error leaked the query string: %q", msg)
	}
}

// TestJiraDecodeErrorRedaction proves a malformed 2xx JSON response body
// never leaks a body fragment into the returned decode error. It targets an
// int field with an overflowing numeric literal — encoding/json's own error
// for that case quotes the literal verbatim (e.g. "json: cannot unmarshal
// number 999...999 into Go struct field .code of type int"), which stands in
// here for a secret-looking token that must never reach the caller.
func TestJiraDecodeErrorRedaction(t *testing.T) {
	const secretLikeToken = "94024242424242424242424242424242424242"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code": ` + secretLikeToken + `}`))
	}))
	defer srv.Close()

	c := &JiraClient{BaseURL: srv.URL, Email: "a", Token: "t"}
	var out struct {
		Code int `json:"code"`
	}
	err := c.do(context.Background(), http.MethodGet, "/x", nil, &out)
	if err == nil {
		t.Fatal("expected a decode error for an overflowing numeric field")
	}
	if strings.Contains(err.Error(), secretLikeToken) {
		t.Errorf("decode error leaked a response body fragment: %q", err.Error())
	}
	if !errors.Is(err, ErrDecodeResponse) {
		t.Errorf("error does not wrap ErrDecodeResponse: %v", err)
	}
}

// TestParseRetryAfter table-tests parseRetryAfter's edge cases: the 5s
// boundary (allowed), zero (allowed), a non-numeric HTTP-date (falls back),
// and a negative value (falls back) — none of which should be confused with
// each other.
func TestParseRetryAfter(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantOK  bool
		wantDur time.Duration
	}{
		{"boundary 5s is allowed", "5", true, 5 * time.Second},
		{"zero is allowed", "0", true, 0},
		{"http-date falls back", "Wed, 21 Oct 2015 07:28:00 GMT", false, 0},
		{"negative falls back", "-3", false, 0},
		{"empty falls back", "", false, 0},
		{"non-numeric falls back", "soon", false, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, ok := parseRetryAfter(tc.in)
			if ok != tc.wantOK {
				t.Errorf("ok = %v, want %v", ok, tc.wantOK)
			}
			if ok && d != tc.wantDur {
				t.Errorf("d = %v, want %v", d, tc.wantDur)
			}
		})
	}
}

// TestJiraRetryAfterEdgeCases exercises parseRetryAfter's edge cases through
// the full retry loop in do, using the injected Sleep so no wall-clock time
// passes: a boundary 5s Retry-After is honored verbatim, "0" is honored as a
// zero-length sleep, and both a non-numeric (HTTP-date) and a negative
// Retry-After fall back to the fixed backoff schedule rather than producing
// a negative or otherwise bogus sleep.
func TestJiraRetryAfterEdgeCases(t *testing.T) {
	cases := []struct {
		name       string
		retryAfter string // header value on the single 429 response; "" omits the header
		wantSlept  []time.Duration
		wantCalls  int
	}{
		{
			name:       "Retry-After 5 boundary honored",
			retryAfter: "5",
			wantSlept:  []time.Duration{5 * time.Second},
			wantCalls:  2,
		},
		{
			name:       "Retry-After 0 honored as zero sleep",
			retryAfter: "0",
			wantSlept:  []time.Duration{0},
			wantCalls:  2,
		},
		{
			name:       "Retry-After HTTP-date falls back to backoff",
			retryAfter: "Wed, 21 Oct 2015 07:28:00 GMT",
			wantSlept:  []time.Duration{250 * time.Millisecond},
			wantCalls:  2,
		},
		{
			name:       "Retry-After negative falls back to backoff",
			retryAfter: "-3",
			wantSlept:  []time.Duration{250 * time.Millisecond},
			wantCalls:  2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var slept []time.Duration
			var calls int
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if calls == 1 {
					if tc.retryAfter != "" {
						w.Header().Set("Retry-After", tc.retryAfter)
					}
					w.WriteHeader(http.StatusTooManyRequests)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			c := &JiraClient{BaseURL: srv.URL, Email: "a", Token: "t", Sleep: func(d time.Duration) { slept = append(slept, d) }}
			if err := c.do(context.Background(), http.MethodGet, "/x", nil, nil); err != nil {
				t.Fatalf("do: %v", err)
			}
			if calls != tc.wantCalls {
				t.Errorf("calls = %d, want %d", calls, tc.wantCalls)
			}
			if len(slept) != len(tc.wantSlept) {
				t.Fatalf("slept = %v, want %v", slept, tc.wantSlept)
			}
			for i := range tc.wantSlept {
				if slept[i] != tc.wantSlept[i] {
					t.Errorf("slept[%d] = %v, want %v", i, slept[i], tc.wantSlept[i])
				}
				if slept[i] < 0 {
					t.Errorf("slept[%d] = %v, must never be negative", i, slept[i])
				}
			}
		})
	}
}

// TestJiraStringRedactsToken proves the client's %v / String() form never
// prints the Token.
func TestJiraStringRedactsToken(t *testing.T) {
	c := &JiraClient{BaseURL: "https://example.atlassian.net", Email: "a@b.com", Token: "super-secret-token"}
	s := c.String()
	if strings.Contains(s, "super-secret-token") {
		t.Errorf("String() leaked the token: %q", s)
	}
	if !strings.Contains(s, "[redacted]") {
		t.Errorf("String() = %q, want a [redacted] marker", s)
	}
	if v := "" + c.String(); strings.Contains(v, "super-secret-token") {
		t.Errorf("%%v leaked the token: %q", v)
	}
}

// TestJiraTimeout15s proves the default HTTP client carries a 15s timeout.
func TestJiraTimeout15s(t *testing.T) {
	c := &JiraClient{BaseURL: "https://example.atlassian.net", Email: "a@b.com", Token: "t"}
	if got := c.httpClient().Timeout; got != 15*time.Second {
		t.Errorf("default client timeout = %v, want 15s", got)
	}
}
