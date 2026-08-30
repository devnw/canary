// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"devnw.dev/canary/pkg/storage"
)

// The scope check inspects EVERY element of a JSON-RPC batch. A read-scoped
// caller may still batch, but only read tools; one mutating element forbids
// the whole payload, and a batch that cannot be parsed is refused rather than
// waved through. See the design note on authConfig.forbiddenTool.

// batchBody wraps one or more tools/call envelopes in a JSON-RPC batch array.
func batchBody(names ...string) string {
	parts := make([]string, 0, len(names))
	for _, n := range names {
		parts = append(parts, callToolBody(n))
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// TestAuthReadTokenRefusedOnBatchedMutatingTool is the regression test for the
// read-scope bypass: the scope check unmarshalled the body into a single
// object, a top-level array failed that unmarshal, the tool name came back
// empty, and the batch was passed through untouched. The SDK accepts batches
// under the 2025-03-26 protocol it defaults to when no version header is sent,
// so `[ {"method":"tools/call","params":{"name":"prioritize"}} ]` from a
// read-only token really did write to the index.
//
// This drives the whole server, not just the middleware, so the assertion is
// on the database rather than on a status code alone.
func TestAuthReadTokenRefusedOnBatchedMutatingTool(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, ".canary", "canary.db")

	db, err := storage.OpenRW(dbPath)
	if err != nil {
		t.Fatalf("OpenRW: %v", err)
	}
	seedPriority(t, db, "default", "TEST-001", 5)
	if cerr := db.Close(); cerr != nil {
		t.Fatalf("close: %v", cerr)
	}

	h := buildHandler(NewServer("test", Deps{Root: root}), authConfig{
		token: "full", readToken: "ro", loopback: true, mutating: mutatingTools(),
	})
	// httptest binds 127.0.0.1:0 -- an ephemeral loopback port.
	srv := httptest.NewServer(h)
	defer srv.Close()

	call := func(body, session string) (int, string, string) {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/mcp", strings.NewReader(body))
		if err != nil {
			t.Fatalf("new request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set("Authorization", "Bearer ro")
		if session != "" {
			req.Header.Set("Mcp-Session-Id", session)
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
		return resp.StatusCode, string(out), resp.Header.Get("Mcp-Session-Id")
	}

	// Initialization and the initialized notification are read-scope traffic;
	// the read token is allowed to establish a session.
	code, body, session := call(`{"jsonrpc":"2.0","id":1,"method":"initialize",`+
		`"params":{"protocolVersion":"2025-03-26","capabilities":{},`+
		`"clientInfo":{"name":"probe","version":"1"}}}`, "")
	if code != http.StatusOK {
		t.Fatalf("initialize: status = %d, want 200; body=%s", code, body)
	}
	if session == "" {
		t.Fatal("initialize returned no session id")
	}
	if code, body, _ = call(`{"jsonrpc":"2.0","method":"notifications/initialized"}`, session); code != http.StatusAccepted {
		t.Fatalf("initialized: status = %d, want 202; body=%s", code, body)
	}

	batch := `[{"jsonrpc":"2.0","id":2,"method":"tools/call","params":` +
		`{"name":"prioritize","arguments":{"reqId":"TEST-001","priority":1}}}]`
	code, body, _ = call(batch, session)
	if code != http.StatusForbidden {
		t.Errorf("batched prioritize from a read token: status = %d, want 403; body=%s", code, body)
	}
	if !strings.Contains(body, "prioritize") {
		t.Errorf("403 does not name the tool: %s", body)
	}

	check, err := storage.OpenRO(dbPath)
	if err != nil {
		t.Fatalf("OpenRO: %v", err)
	}
	defer func() { _ = check.Close() }()
	tokens, err := check.GetTokensByReqID("default", "TEST-001")
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("want one seeded token, got %d", len(tokens))
	}
	if tokens[0].Priority != 5 {
		t.Fatalf("the batch wrote through the read token: priority = %d, want 5", tokens[0].Priority)
	}
}

// TestAuthWriteTokenMayBatchMutatingTool proves the fix scopes to the read
// credential: a read+write token batches whatever it likes.
func TestAuthWriteTokenMayBatchMutatingTool(t *testing.T) {
	cfg := authConfig{token: "full", readToken: "ro", loopback: true, mutating: mutatingTools()}
	h := authMiddleware(cfg, echoHandler(t))

	body := batchBody("prioritize", "bug-create")
	code, echoed := post(t, h, "Bearer full", body)
	if code != http.StatusOK {
		t.Fatalf("full token on a batch: status = %d, want 200; body=%s", code, echoed)
	}
	if echoed != body {
		t.Fatalf("batch body was not replayed intact: %q", echoed)
	}
}

// TestAuthReadTokenMayBatchReadOnlyCalls documents the chosen design: batches
// are inspected element by element rather than refused outright, so a
// read-scoped caller keeps the batching the 2025-03-26 protocol allows as long
// as no element mutates.
func TestAuthReadTokenMayBatchReadOnlyCalls(t *testing.T) {
	cfg := authConfig{token: "full", readToken: "ro", loopback: true, mutating: mutatingTools()}
	h := authMiddleware(cfg, echoHandler(t))

	body := batchBody("list", "search")
	code, echoed := post(t, h, "Bearer ro", body)
	if code != http.StatusOK {
		t.Fatalf("read token on a read-only batch: status = %d, want 200; body=%s", code, echoed)
	}
	if echoed != body {
		t.Fatalf("batch body was not replayed intact: %q", echoed)
	}

	// A mutating element anywhere in the batch forbids the whole payload --
	// not only the first position.
	for _, mixed := range []string{
		batchBody("list", "prioritize"),
		batchBody("bug-create", "list"),
	} {
		if code, body := post(t, h, "Bearer ro", mixed); code != http.StatusForbidden {
			t.Errorf("mixed batch %s: status = %d, want 403; body=%s", mixed, code, body)
		}
	}
}

// TestAuthUnparseableBatchFailsClosed proves the middleware never treats "I
// could not read this batch" as "this batch is harmless". A top-level array it
// cannot decode is refused, because the SDK might still decode it.
func TestAuthUnparseableBatchFailsClosed(t *testing.T) {
	cfg := authConfig{token: "full", readToken: "ro", loopback: true, mutating: mutatingTools()}
	h := authMiddleware(cfg, echoHandler(t))

	for _, body := range []string{
		`[{"jsonrpc":"2.0","method":`,
		`[1,2,3]`,
		`["tools/call"]`,
		`  [ {"method":"tools/call","params":{"name":"prioritize"}} `,
		"\n\t[{\"method\":\"tools/call\",\"params\":[]}]",
	} {
		code, echoed := post(t, h, "Bearer ro", body)
		if code != http.StatusForbidden {
			t.Errorf("unparseable batch %q: status = %d, want 403; body=%s", body, code, echoed)
		}
		if strings.Contains(echoed, "ro") && !strings.Contains(echoed, "forbidden") {
			t.Errorf("refusal body looks wrong: %s", echoed)
		}
	}
}
