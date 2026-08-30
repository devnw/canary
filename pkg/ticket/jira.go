// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package ticket

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"devnw.dev/canary/pkg/config"
)

// ErrDecodeResponse is the sentinel wrapped by a JSON decode failure on an
// otherwise-successful (2xx) response. The printed error string is always
// static (method + safePath + this reason) — the underlying json error's
// text is deliberately discarded because it can quote a fragment of the
// offending response body (for example an overflowing numeric literal),
// which must never reach a log line.
var ErrDecodeResponse = errors.New("decode response")

// maxResponseBytes caps every JIRA response body read; a body larger than
// this is rejected without ever being buffered in full or echoed into an
// error, so a hostile or misbehaving endpoint can neither exhaust memory nor
// smuggle its payload into a log line.
const maxResponseBytes = 2 << 20 // 2 MiB

// maxRetryAfter bounds how long a Retry-After header may ask us to wait: a
// value at or under this is honored, a larger one aborts the call
// immediately rather than parking a sync for minutes.
const maxRetryAfter = 5 * time.Second

// JiraClient talks to the JIRA Cloud REST API (v3) using basic auth
// (email + API token). Every method is safe for concurrent use.
// CANARY: REQ=CP-279; FEATURE="TicketSync"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_306_JiraClient_CreateIssue,TestCANARY_CBIN_306_JiraClient_CreateIssue_ErrorStatus,TestCANARY_CBIN_306_JiraClient_TransitionIssue_ResolvesIDByName,TestCANARY_CBIN_306_JiraClient_TransitionIssue_NoMatch,TestCANARY_CBIN_306_FetchRemoteStatus_Paged; UPDATED=2026-08-30
// CANARY: REQ=CP-279; FEATURE="TicketSync"; ASPECT=Security; STATUS=TESTED; TEST=TestJiraSearchJQLPagination,TestJiraSearchInvalidProjectKey,TestJiraRetry429ThenSuccess,TestJiraNoRetryOn400,TestJiraBodyCap,TestJiraErrorRedaction,TestJiraStringRedactsToken,TestJiraTimeout15s,TestAuditF16; UPDATED=2026-08-30
type JiraClient struct {
	BaseURL string
	Email   string
	Token   string

	// HTTPClient lets tests point the client at an httptest server; nil
	// defaults to a client with a bounded 15s timeout.
	HTTPClient *http.Client

	// Sleep injects the backoff wait so tests can assert on (and skip) the
	// retry delays; nil defaults to time.Sleep.
	Sleep func(time.Duration)
}

// String renders the client with its Token redacted, so a stray %v (in a log
// line or wrapped error) can never spill the API token.
func (c *JiraClient) String() string {
	return fmt.Sprintf("JiraClient{BaseURL: %s, Email: %s, Token: [redacted]}", c.BaseURL, c.Email)
}

func (c *JiraClient) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 15 * time.Second}
}

func (c *JiraClient) sleep(d time.Duration) {
	if c.Sleep != nil {
		c.Sleep(d)
		return
	}
	time.Sleep(d)
}

func (c *JiraClient) authHeader() string {
	raw := c.Email + ":" + c.Token
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(raw))
}

// pathOnly strips any query string from path so error messages never leak
// query parameters (for example a JQL expression carried on the URL).
func pathOnly(path string) string {
	if i := strings.IndexByte(path, '?'); i >= 0 {
		return path[:i]
	}
	return path
}

// parseRetryAfter parses a Retry-After header expressed in delta-seconds.
// Only the numeric form is honored; an HTTP-date (or anything unparseable)
// reports ok=false so the caller falls back to its own backoff schedule.
func parseRetryAfter(v string) (time.Duration, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0, false
	}
	return time.Duration(n) * time.Second, true
}

// readCapped reads at most maxResponseBytes from r. A body that would exceed
// the cap is rejected with a static error that never includes the body
// itself.
func readCapped(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read response")
	}
	if len(data) > maxResponseBytes {
		return nil, fmt.Errorf("response exceeds 2 MiB")
	}
	return data, nil
}

// retryBackoffs is the fixed backoff schedule for retryable (429/5xx)
// responses: three retries at 250ms, 500ms, then 1s.
var retryBackoffs = []time.Duration{250 * time.Millisecond, 500 * time.Millisecond, time.Second}

// do issues one JIRA REST call: method + path (path must start with "/" and
// may include a query string), an optional JSON body, decoding a JSON
// response into out (when non-nil).
//
// Retries: a 429 or 5xx response is retried up to len(retryBackoffs) times
// on the fixed backoff schedule. A parseable Retry-After header overrides the
// backoff when it is <= maxRetryAfter; a larger Retry-After aborts the call
// immediately without sleeping. Every other 4xx is returned at once, never
// retried.
//
// Redaction: no error from this method ever contains a response body, a URL
// carrying a query string, or a credential — only the method, the path with
// its query stripped, the status code, and a short static reason.
func (c *JiraClient) do(ctx context.Context, method, path string, body, out any) error {
	var rawBody []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("jira: encode request: %w", err)
		}
		rawBody = b
	}

	safePath := pathOnly(path)
	target := strings.TrimRight(c.BaseURL, "/") + path

	for attempt := 0; ; attempt++ {
		var reader io.Reader
		if rawBody != nil {
			reader = bytes.NewReader(rawBody)
		}
		req, err := http.NewRequestWithContext(ctx, method, target, reader)
		if err != nil {
			return fmt.Errorf("jira: build request: %w", err)
		}
		req.Header.Set("Authorization", c.authHeader())
		req.Header.Set("Accept", "application/json")
		if rawBody != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient().Do(req)
		if err != nil {
			// Never embed err: it carries the full URL (query and all).
			return fmt.Errorf("jira: %s %s: request failed", method, safePath)
		}

		// Retryable: 429 or any 5xx.
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			retryAfter, haveRetryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
			_ = resp.Body.Close()

			if haveRetryAfter && retryAfter > maxRetryAfter {
				return fmt.Errorf("jira: %s %s: status %d: retry-after exceeds limit", method, safePath, resp.StatusCode)
			}
			if attempt >= len(retryBackoffs) {
				return fmt.Errorf("jira: %s %s: status %d: retries exhausted", method, safePath, resp.StatusCode)
			}
			delay := retryBackoffs[attempt]
			if haveRetryAfter {
				delay = retryAfter
			}
			c.sleep(delay)
			continue
		}

		data, rerr := readCapped(resp.Body)
		_ = resp.Body.Close()
		if rerr != nil {
			return fmt.Errorf("jira: %s %s: %v", method, safePath, rerr)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("jira: %s %s: status %d", method, safePath, resp.StatusCode)
		}

		if out != nil && len(data) > 0 {
			if err := json.Unmarshal(data, out); err != nil {
				// Never embed err: its text can quote a fragment of data
				// (e.g. an overflowing numeric literal) verbatim.
				return fmt.Errorf("jira: %s %s: %w", method, safePath, ErrDecodeResponse)
			}
		}
		return nil
	}
}

// adfParagraph wraps plain text as a minimal single-paragraph Atlassian
// Document Format node — the shape JIRA Cloud's v3 API requires for
// rich-text fields such as description.
func adfParagraph(text string) map[string]any {
	return map[string]any{
		"type":    "doc",
		"version": 1,
		"content": []any{
			map[string]any{
				"type": "paragraph",
				"content": []any{
					map[string]any{"type": "text", "text": text},
				},
			},
		},
	}
}

// CreateIssue creates a new issue of issueType in project with summary and
// description (wrapped as an ADF paragraph), returning the created issue's
// key (e.g. "CP-12").
func (c *JiraClient) CreateIssue(ctx context.Context, project, issueType, summary, description string) (string, error) {
	body := map[string]any{
		"fields": map[string]any{
			"project":     map[string]any{"key": project},
			"issuetype":   map[string]any{"name": issueType},
			"summary":     summary,
			"description": adfParagraph(description),
		},
	}
	var out struct {
		Key string `json:"key"`
	}
	if err := c.do(ctx, http.MethodPost, "/rest/api/3/issue", body, &out); err != nil {
		return "", err
	}
	if out.Key == "" {
		return "", fmt.Errorf("jira: create issue: response had no key")
	}
	return out.Key, nil
}

// TransitionIssue moves key to the transition whose target status name
// (transitions[].to.name) matches toStatusName case-insensitively.
func (c *JiraClient) TransitionIssue(ctx context.Context, key, toStatusName string) error {
	var list struct {
		Transitions []struct {
			ID string `json:"id"`
			To struct {
				Name string `json:"name"`
			} `json:"to"`
		} `json:"transitions"`
	}
	if err := c.do(ctx, http.MethodGet, "/rest/api/3/issue/"+url.PathEscape(key)+"/transitions", nil, &list); err != nil {
		return err
	}

	var id string
	for _, tr := range list.Transitions {
		if strings.EqualFold(tr.To.Name, toStatusName) {
			id = tr.ID
			break
		}
	}
	if id == "" {
		return fmt.Errorf("jira: %s: no transition to status %q found", key, toStatusName)
	}

	body := map[string]any{"transition": map[string]any{"id": id}}
	return c.do(ctx, http.MethodPost, "/rest/api/3/issue/"+url.PathEscape(key)+"/transitions", body, nil)
}

// searchPageSize bounds each /search/jql page; FetchRemoteStatus loops until
// the endpoint stops returning a nextPageToken.
const searchPageSize = 50

// maxSearchPages guards the token-paginated search loop against an endpoint
// that never stops handing back a nextPageToken.
const maxSearchPages = 200

// FetchRemoteStatus pages through JIRA's token-based search endpoint
// (POST /rest/api/3/search/jql) for every issue in project, returning a map
// of issue key -> current status name. Used only under `--apply`, immediately
// before ComputePlan, so transition actions are proposed against real remote
// state.
//
// The project key is validated against config.SourceKeyPattern before any
// request is made, and quoted inside the JQL (project = "KEY") so it can
// never break out of the clause. Pagination follows nextPageToken, capped at
// maxSearchPages.
func FetchRemoteStatus(ctx context.Context, c *JiraClient, project string) (map[string]string, error) {
	if !config.SourceKeyPattern.MatchString(project) {
		return nil, fmt.Errorf("jira: invalid project key")
	}

	out := map[string]string{}
	jql := fmt.Sprintf("project = %q", project) // project = "KEY"
	nextPageToken := ""

	for page := 0; ; page++ {
		if page >= maxSearchPages {
			return nil, fmt.Errorf("jira: search exceeded %d pages", maxSearchPages)
		}

		body := map[string]any{
			"jql":        jql,
			"maxResults": searchPageSize,
			"fields":     []string{"status"},
		}
		if nextPageToken != "" {
			body["nextPageToken"] = nextPageToken
		}

		var resp struct {
			Issues []struct {
				Key    string `json:"key"`
				Fields struct {
					Status struct {
						Name string `json:"name"`
					} `json:"status"`
				} `json:"fields"`
			} `json:"issues"`
			NextPageToken string `json:"nextPageToken"`
		}

		if err := c.do(ctx, http.MethodPost, "/rest/api/3/search/jql", body, &resp); err != nil {
			return nil, err
		}
		for _, iss := range resp.Issues {
			out[iss.Key] = iss.Fields.Status.Name
		}
		if resp.NextPageToken == "" {
			break
		}
		nextPageToken = resp.NextPageToken
	}
	return out, nil
}
