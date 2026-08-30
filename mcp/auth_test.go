// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// echoHandler stands in for the MCP handler: it reports that it was reached
// and echoes the body it received, which is how the tests prove the scope
// check replaces the body it consumed.
func echoHandler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})
}

func callToolBody(name string) string {
	return `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"` + name + `","arguments":{}}}`
}

// post drives one request through the middleware against a real loopback
// listener (httptest binds 127.0.0.1:0), returning status and body.
func post(t *testing.T, h http.Handler, authHeader, body string) (int, string) {
	t.Helper()
	srv := httptest.NewServer(h)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	return resp.StatusCode, string(out)
}

// TestAuthLoopbackNoTokensAllows proves the developer default: a loopback
// server with no tokens configured serves everything, including a mutating
// tool.
func TestAuthLoopbackNoTokensAllows(t *testing.T) {
	cfg := authConfig{loopback: true, mutating: mutatingTools()}
	h := authMiddleware(cfg, echoHandler(t))

	code, body := post(t, h, "", callToolBody("prioritize"))
	if code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (loopback dev mode); body=%s", code, body)
	}
	if body != callToolBody("prioritize") {
		t.Fatalf("body was not replayed to the handler: %q", body)
	}
}

// TestAuthNonLoopbackWithoutTokensDenies proves the middleware does not fall
// open if it is ever reached on a non-loopback listener with no tokens. The
// command refuses that configuration at startup; this is the second lock.
func TestAuthNonLoopbackWithoutTokensDenies(t *testing.T) {
	cfg := authConfig{loopback: false, mutating: mutatingTools()}
	h := authMiddleware(cfg, echoHandler(t))

	if code, body := post(t, h, "", callToolBody("list")); code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body=%s", code, body)
	}
}

// TestAuthRequiresBearerWhenConfigured proves that configuring a token makes
// it required even on loopback: the reason to set one is that something other
// than the developer can reach the port.
func TestAuthRequiresBearerWhenConfigured(t *testing.T) {
	cfg := authConfig{token: "s3cret", loopback: true, mutating: mutatingTools()}
	h := authMiddleware(cfg, echoHandler(t))

	if code, _ := post(t, h, "", callToolBody("list")); code != http.StatusUnauthorized {
		t.Errorf("missing credential: status = %d, want 401", code)
	}
	if code, _ := post(t, h, "Bearer wrong", callToolBody("list")); code != http.StatusUnauthorized {
		t.Errorf("wrong credential: status = %d, want 401", code)
	}
	if code, _ := post(t, h, "s3cret", callToolBody("list")); code != http.StatusUnauthorized {
		t.Errorf("credential without the Bearer scheme: status = %d, want 401", code)
	}
	if code, _ := post(t, h, "Bearer s3cret", callToolBody("prioritize")); code != http.StatusOK {
		t.Errorf("full token on a mutating tool: status = %d, want 200", code)
	}
	// The scheme is case-insensitive per RFC 7235.
	if code, _ := post(t, h, "bearer s3cret", callToolBody("list")); code != http.StatusOK {
		t.Errorf("lowercase scheme: status = %d, want 200", code)
	}
}

// TestAuthReadTokenRefusedOnMutatingTool proves the read-only credential can
// read and cannot write -- and that the distinction is drawn from the tool
// name in the JSON-RPC envelope, since every call arrives at the same path.
func TestAuthReadTokenRefusedOnMutatingTool(t *testing.T) {
	cfg := authConfig{token: "full", readToken: "ro", loopback: true, mutating: mutatingTools()}
	h := authMiddleware(cfg, echoHandler(t))

	if code, _ := post(t, h, "Bearer ro", callToolBody("list")); code != http.StatusOK {
		t.Errorf("read token on a read tool: status = %d, want 200", code)
	}
	for _, tool := range []string{"prioritize", "bug-create"} {
		code, body := post(t, h, "Bearer ro", callToolBody(tool))
		if code != http.StatusForbidden {
			t.Errorf("read token on %s: status = %d, want 403", tool, code)
		}
		if !strings.Contains(body, tool) {
			t.Errorf("403 for %s does not name the tool: %s", tool, body)
		}
	}
	if code, _ := post(t, h, "Bearer full", callToolBody("bug-create")); code != http.StatusOK {
		t.Errorf("full token on a mutating tool: status = %d, want 200", code)
	}
}

// TestAuthNonToolCallPassesThrough proves the scope check does not become a
// second, worse JSON-RPC implementation: anything that is not a tools/call
// (initialization, a listing, an unparseable body) reaches the SDK with its
// body intact and is judged there.
func TestAuthNonToolCallPassesThrough(t *testing.T) {
	cfg := authConfig{readToken: "ro", loopback: true, mutating: mutatingTools()}
	h := authMiddleware(cfg, echoHandler(t))

	for _, body := range []string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`,
		`not json at all`,
	} {
		code, echoed := post(t, h, "Bearer ro", body)
		if code != http.StatusOK {
			t.Errorf("body %q: status = %d, want 200", body, code)
		}
		if echoed != body {
			t.Errorf("body %q was not replayed intact: %q", body, echoed)
		}
	}
}

// TestAuthNeverEchoesTokens proves no refusal discloses a configured token.
// An error that names the expected credential is the same leak as printing
// it.
func TestAuthNeverEchoesTokens(t *testing.T) {
	const full, ro = "full-secret-value", "read-secret-value"
	cfg := authConfig{token: full, readToken: ro, loopback: true, mutating: mutatingTools()}
	h := authMiddleware(cfg, echoHandler(t))

	for _, header := range []string{"", "Bearer wrong"} {
		_, body := post(t, h, header, callToolBody("list"))
		if strings.Contains(body, full) || strings.Contains(body, ro) {
			t.Fatalf("refusal disclosed a configured token: %s", body)
		}
	}
	_, body := post(t, h, "Bearer "+ro, callToolBody("prioritize"))
	if strings.Contains(body, full) || strings.Contains(body, ro) {
		t.Fatalf("403 disclosed a configured token: %s", body)
	}
}

// TestScopesFor covers the scope resolution table directly.
func TestScopesFor(t *testing.T) {
	tests := []struct {
		name     string
		cfg      authConfig
		header   string
		wantOK   bool
		wantRead bool
		wantMut  bool
	}{
		{"loopback dev mode", authConfig{loopback: true}, "", true, true, true},
		{"non-loopback dev mode is refused", authConfig{loopback: false}, "", false, false, false},
		{"full token", authConfig{token: "a", loopback: true}, "Bearer a", true, true, true},
		{"read token", authConfig{readToken: "b", loopback: true}, "Bearer b", true, true, false},
		{"read token cannot mutate", authConfig{token: "a", readToken: "b", loopback: true}, "Bearer b", true, true, false},
		{"unknown token", authConfig{token: "a", loopback: true}, "Bearer c", false, false, false},
		{"empty bearer", authConfig{token: "a", loopback: true}, "Bearer ", false, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.cfg.scopesFor(tt.header)
			if ok != tt.wantOK || got.Read != tt.wantRead || got.Mutate != tt.wantMut {
				t.Fatalf("scopesFor(%q) = %+v, %v; want read=%v mutate=%v ok=%v",
					tt.header, got, ok, tt.wantRead, tt.wantMut, tt.wantOK)
			}
		})
	}
}

// TestIsLoopbackHost covers the exposure decision.
func TestIsLoopbackHost(t *testing.T) {
	loopback := []string{"127.0.0.1", "127.0.0.53", "::1", "localhost"}
	exposed := []string{"", "0.0.0.0", "::", "192.168.1.10", "example.invalid"}

	for _, h := range loopback {
		if !isLoopbackHost(h) {
			t.Errorf("isLoopbackHost(%q) = false, want true", h)
		}
	}
	for _, h := range exposed {
		if isLoopbackHost(h) {
			t.Errorf("isLoopbackHost(%q) = true, want false", h)
		}
	}
}
