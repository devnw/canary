// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package ticket

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// JiraClient talks to the JIRA Cloud REST API (v3) using basic auth
// (email + API token). Every method is safe for concurrent use.
// CANARY: REQ=CP-279; FEATURE="TicketSync"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_306_JiraClient_CreateIssue,TestCANARY_CBIN_306_JiraClient_CreateIssue_ErrorStatus,TestCANARY_CBIN_306_JiraClient_TransitionIssue_ResolvesIDByName,TestCANARY_CBIN_306_JiraClient_TransitionIssue_NoMatch,TestCANARY_CBIN_306_FetchRemoteStatus_Paged; UPDATED=2026-08-29
type JiraClient struct {
	BaseURL string
	Email   string
	Token   string

	// HTTPClient lets tests point the client at an httptest server with a
	// short timeout; nil defaults to a client with a bounded timeout.
	HTTPClient *http.Client
}

func (c *JiraClient) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func (c *JiraClient) authHeader() string {
	raw := c.Email + ":" + c.Token
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(raw))
}

// do issues one JIRA REST call: method + path (path must start with "/" and
// may include a query string), an optional JSON body, decoding a JSON
// response into out (when non-nil). A non-2xx response becomes an error
// carrying the status code and response body.
func (c *JiraClient) do(method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("jira: encode request: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, strings.TrimRight(c.BaseURL, "/")+path, reader)
	if err != nil {
		return fmt.Errorf("jira: build request: %w", err)
	}
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("jira: %s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("jira: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("jira: %s %s: status %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(data)))
	}

	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("jira: decode %s %s response: %w", method, path, err)
		}
	}
	return nil
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
func (c *JiraClient) CreateIssue(project, issueType, summary, description string) (string, error) {
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
	if err := c.do(http.MethodPost, "/rest/api/3/issue", body, &out); err != nil {
		return "", err
	}
	if out.Key == "" {
		return "", fmt.Errorf("jira: create issue: response had no key")
	}
	return out.Key, nil
}

// TransitionIssue moves key to the transition whose target status name
// (transitions[].to.name) matches toStatusName case-insensitively.
func (c *JiraClient) TransitionIssue(key, toStatusName string) error {
	var list struct {
		Transitions []struct {
			ID string `json:"id"`
			To struct {
				Name string `json:"name"`
			} `json:"to"`
		} `json:"transitions"`
	}
	if err := c.do(http.MethodGet, "/rest/api/3/issue/"+url.PathEscape(key)+"/transitions", nil, &list); err != nil {
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
	return c.do(http.MethodPost, "/rest/api/3/issue/"+url.PathEscape(key)+"/transitions", body, nil)
}

// searchPageSize bounds each /search page; FetchRemoteStatus loops until
// every issue in project has been seen.
const searchPageSize = 50

// FetchRemoteStatus pages through JIRA's search endpoint
// (GET /rest/api/3/search) for every issue in project, returning a map of
// issue key -> current status name. Used only under `--apply`, immediately
// before ComputePlan, so transition actions are proposed against real
// remote state.
func FetchRemoteStatus(c *JiraClient, project string) (map[string]string, error) {
	out := map[string]string{}
	startAt := 0
	for {
		var page struct {
			Issues []struct {
				Key    string `json:"key"`
				Fields struct {
					Status struct {
						Name string `json:"name"`
					} `json:"status"`
				} `json:"fields"`
			} `json:"issues"`
			Total int `json:"total"`
		}

		q := url.Values{}
		q.Set("jql", "project="+project)
		q.Set("startAt", fmt.Sprintf("%d", startAt))
		q.Set("maxResults", fmt.Sprintf("%d", searchPageSize))
		q.Set("fields", "status")
		path := "/rest/api/3/search?" + q.Encode()

		if err := c.do(http.MethodGet, path, nil, &page); err != nil {
			return nil, err
		}
		for _, iss := range page.Issues {
			out[iss.Key] = iss.Fields.Status.Name
		}
		startAt += len(page.Issues)
		if len(page.Issues) == 0 || startAt >= page.Total {
			break
		}
	}
	return out, nil
}
