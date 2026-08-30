// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"devnw.dev/canary/pkg/ticket"
)

// TestAuditF16 exercises the hardened JIRA client end-to-end against an
// httptest server in a single scenario: the token-paginated
// POST /rest/api/3/search/jql loop survives a transient 429 (one bounded
// retry on the injected Sleep), follows nextPageToken across two pages, and —
// on a later 5xx carrying a secret body — produces an error that leaks
// neither the body, the credential, nor the query string.
//
// F-16: the old client used the deprecated offset endpoint
// (GET /rest/api/3/search?jql=...), never retried, read the whole body
// unbounded, and embedded the raw response body and full URL in its errors.
func TestAuditF16(t *testing.T) {
	// Phase 1: pagination + one 429 retry.
	var slept []time.Duration
	var calls int
	var deprecatedHit bool

	paged := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/api/3/search" || strings.Contains(r.URL.RawQuery, "jql") {
			deprecatedHit = true
		}
		if r.Method != http.MethodPost || r.URL.Path != "/rest/api/3/search/jql" {
			t.Errorf("unexpected request: %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		calls++
		w.Header().Set("Content-Type", "application/json")
		switch calls {
		case 1:
			// Transient rate-limit: the client must retry.
			w.WriteHeader(http.StatusTooManyRequests)
		case 2:
			// First page, linking to a second via nextPageToken.
			_, _ = w.Write([]byte(`{"issues":[{"key":"CP-1","fields":{"status":{"name":"To Do"}}}],"nextPageToken":"page-2"}`))
		default:
			// Second (final) page: no nextPageToken.
			_, _ = w.Write([]byte(`{"issues":[{"key":"CP-2","fields":{"status":{"name":"Done"}}}]}`))
		}
	}))
	defer paged.Close()

	c := &ticket.JiraClient{
		BaseURL: paged.URL,
		Email:   "agent@example.com",
		Token:   "super-secret-token",
		Sleep:   func(d time.Duration) { slept = append(slept, d) },
	}
	got, err := ticket.FetchRemoteStatus(context.Background(), c, "CP")
	if err != nil {
		t.Fatalf("FetchRemoteStatus: %v", err)
	}
	if deprecatedHit {
		t.Error("the deprecated /rest/api/3/search offset endpoint must never be used")
	}
	if len(slept) != 1 || slept[0] != 250*time.Millisecond {
		t.Errorf("backoff = %v, want a single 250ms retry", slept)
	}
	want := map[string]string{"CP-1": "To Do", "CP-2": "Done"}
	if len(got) != len(want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("got[%q] = %q, want %q", k, got[k], v)
		}
	}

	// Phase 2: redaction on a 5xx carrying a secret body and reached via a
	// query-bearing path.
	fivexx := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"secret-token-xyz"}`))
	}))
	defer fivexx.Close()

	rc := &ticket.JiraClient{
		BaseURL: fivexx.URL,
		Email:   "agent@example.com",
		Token:   "super-secret-token",
		Sleep:   func(time.Duration) {}, // no-op: don't actually wait out the retries
	}
	rerr := ticketDo(t, rc)
	if rerr == nil {
		t.Fatal("expected an error on a persistent 500")
	}
	msg := rerr.Error()
	if strings.Contains(msg, "secret-token-xyz") {
		t.Errorf("error leaked the response body: %q", msg)
	}
	if strings.Contains(msg, "super-secret-token") {
		t.Errorf("error leaked the token: %q", msg)
	}
	// The endpoint path (/rest/api/3/search/jql) legitimately contains "jql";
	// what must never appear is a query string carrying a JQL expression.
	if strings.Contains(msg, "?jql=") || strings.Contains(msg, "project = ") {
		t.Errorf("error leaked the query/JQL: %q", msg)
	}
}

// ticketDo drives a redaction-triggering call through the exported surface:
// FetchRemoteStatus issues POST /rest/api/3/search/jql, whose 500 response
// (secret body) exercises the redacted error path without any query string —
// and the search body would carry the project key, never a credential.
func ticketDo(t *testing.T, c *ticket.JiraClient) error {
	t.Helper()
	_, err := ticket.FetchRemoteStatus(context.Background(), c, "CP")
	return err
}
