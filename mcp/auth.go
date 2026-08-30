// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package mcp

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

// CANARY: REQ=ENG-4393; FEATURE="MCPAuth"; ASPECT=Security; STATUS=TESTED; TEST=TestAuthLoopbackNoTokensAllows,TestAuthNonLoopbackWithoutTokensDenies,TestAuthRequiresBearerWhenConfigured,TestAuthReadTokenRefusedOnMutatingTool,TestAuthReadTokenRefusedOnBatchedMutatingTool,TestAuthWriteTokenMayBatchMutatingTool,TestAuthReadTokenMayBatchReadOnlyCalls,TestAuthUnparseableBatchFailsClosed,TestAuthNeverEchoesTokens,TestScopesFor,TestIsLoopbackHost; UPDATED=2026-08-30

// Environment variables that configure MCP authentication. Their VALUES are
// never logged, echoed in an error, or written to a response body: an error
// message that names the expected token is the same leak as printing it.
const (
	// EnvToken grants read and mutate.
	EnvToken = "CANARY_MCP_TOKEN"
	// EnvReadToken grants read only.
	EnvReadToken = "CANARY_MCP_READ_TOKEN"
)

// Scopes is what a request is allowed to do.
type Scopes struct {
	Read   bool
	Mutate bool
}

// authConfig is the resolved authentication policy for one server run.
type authConfig struct {
	// token grants read+mutate; readToken grants read. Either may be empty.
	token     string
	readToken string
	// loopback reports whether the listener is reachable only from this
	// machine. It is the only reason a request may be served unauthenticated.
	loopback bool
	// mutating names the tools a read-only caller may not invoke.
	mutating map[string]bool
}

// authFromEnv reads the policy from the environment.
func authFromEnv(loopback bool) authConfig {
	return authConfig{
		token:     strings.TrimSpace(os.Getenv(EnvToken)),
		readToken: strings.TrimSpace(os.Getenv(EnvReadToken)),
		loopback:  loopback,
		mutating:  mutatingTools(),
	}
}

// configured reports whether any token was supplied. When none was, the
// server is in dev mode and only loopback may reach it.
func (c authConfig) configured() bool {
	return c.token != "" || c.readToken != ""
}

// scopesFor resolves an Authorization header to what it may do.
//
// The comparison is constant-time in both directions. A byte-by-byte compare
// leaks the shared secret one character at a time to anything that can time
// the response, which on a loopback socket is every process on the box.
func (c authConfig) scopesFor(header string) (Scopes, bool) {
	if !c.configured() {
		// No tokens configured: full access, but only because the caller
		// already had to be on the loopback interface to get here (a
		// non-loopback bind without a token is refused at startup). A
		// refusal carries no scopes at all, so a caller that ignores the
		// second return cannot accidentally read permission out of it.
		if !c.loopback {
			return Scopes{}, false
		}
		return Scopes{Read: true, Mutate: true}, true
	}

	presented, ok := bearerToken(header)
	if !ok {
		return Scopes{}, false
	}

	// Both comparisons always run: returning early on the first match would
	// make "wrong full token" and "wrong read token" take different times.
	full := c.token != "" && constantTimeEqual(presented, c.token)
	read := c.readToken != "" && constantTimeEqual(presented, c.readToken)
	switch {
	case full:
		return Scopes{Read: true, Mutate: true}, true
	case read:
		return Scopes{Read: true}, true
	default:
		return Scopes{}, false
	}
}

// bearerToken extracts the credential from an Authorization header.
func bearerToken(header string) (string, bool) {
	const prefix = "bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	tok := strings.TrimSpace(header[len(prefix):])
	if tok == "" {
		return "", false
	}
	return tok, true
}

func constantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// maxAuthBodyPeek bounds how much of a request body the scope check will read
// to find the tool name. A tools/call envelope is small; anything larger is
// either not a tool call or is already over the body cap.
const maxAuthBodyPeek = 1 << 20

// authMiddleware enforces the policy in front of the MCP handler.
//
// Loopback with no tokens configured is full access -- the developer default,
// where the only reachable caller is already running as the user. As soon as
// either token is set, a bearer credential is required even on loopback,
// because the reason to set one is that something other than the developer
// can reach the port.
//
// A read-scoped credential is refused (403) on a mutating tool. That check
// has to look at the request body: the tool name lives in the JSON-RPC
// envelope, not in the URL, so scoping by path alone would let a read token
// call `prioritize` -- and it has to look at every element of a batch, not
// just at a single envelope (see forbiddenTool).
func authMiddleware(cfg authConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scopes, ok := cfg.scopesFor(r.Header.Get("Authorization"))
		if !ok {
			w.Header().Set("WWW-Authenticate", `Bearer realm="canary-mcp"`)
			writeAuthError(w, http.StatusUnauthorized, "unauthorized: a valid Bearer token is required")
			return
		}

		if !scopes.Mutate && r.Method == http.MethodPost {
			buf, body, err := peekBody(r)
			if err != nil {
				writeAuthError(w, http.StatusBadRequest, "malformed request body")
				return
			}
			r.Body = body

			name, inspected := cfg.forbiddenTool(buf)
			if !inspected {
				// A batch this check could not decode may still be a batch the
				// SDK can decode, so it is refused rather than forwarded. The
				// alternative -- assume an undecodable payload is harmless --
				// is exactly the hole this branch closes.
				writeAuthError(w, http.StatusForbidden,
					"forbidden: a batched request that cannot be inspected requires the read+write token")
				return
			}
			if name != "" {
				writeAuthError(w, http.StatusForbidden,
					"forbidden: tool "+name+" requires the read+write token")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// jsonrpcEnvelope is the part of a JSON-RPC request the scope check reads.
type jsonrpcEnvelope struct {
	Method string `json:"method"`
	Params struct {
		Name string `json:"name"`
	} `json:"params"`
}

// forbiddenTool reports the mutating tool a read-scoped caller may not invoke
// in this body, and whether the body was understood well enough to say.
//
// A top-level array is a JSON-RPC batch. The SDK accepts one whenever the
// negotiated protocol version is below 2025-06-18, and it negotiates
// 2025-03-26 when the client sends no version header -- so a batch is not an
// exotic payload to be waved through, it is a second way to call every tool.
// Every element is inspected, not just the first: a read-scoped caller keeps
// the batching the protocol allows, and one mutating element anywhere forbids
// the whole payload.
//
// The second return is the fail-closed switch. A batch that will not decode
// yields false -- refused -- because an undecodable batch here may still be a
// decodable batch in the SDK. A single object that will not decode
// keeps the older behavior: it is not a tools/call this check can name, the
// SDK decides whether it is a valid request at all, and it cannot smuggle a
// second call the way an array can.
//
// A name is only ever returned when it is in the mutating set, so the 403
// message never echoes an arbitrary caller-supplied string.
func (c authConfig) forbiddenTool(buf []byte) (string, bool) {
	if firstJSONByte(buf) == '[' {
		var batch []jsonrpcEnvelope
		if err := json.Unmarshal(buf, &batch); err != nil {
			return "", false
		}
		for _, envelope := range batch {
			if name := c.mutatingCall(envelope); name != "" {
				return name, true
			}
		}
		return "", true
	}

	var envelope jsonrpcEnvelope
	if err := json.Unmarshal(buf, &envelope); err != nil {
		return "", true
	}
	return c.mutatingCall(envelope), true
}

// mutatingCall names the tool this envelope calls if a read-scoped caller may
// not call it, and "" otherwise.
func (c authConfig) mutatingCall(envelope jsonrpcEnvelope) string {
	if envelope.Method != "tools/call" || !c.mutating[envelope.Params.Name] {
		return ""
	}
	return envelope.Params.Name
}

// firstJSONByte returns the first byte of buf that JSON does not treat as
// whitespace, or 0 for an all-whitespace or empty body.
func firstJSONByte(buf []byte) byte {
	trimmed := bytes.TrimLeft(buf, " \t\r\n")
	if len(trimmed) == 0 {
		return 0
	}
	return trimmed[0]
}

// peekBody reads the request body far enough for the scope check to inspect
// it, and returns a replacement body the downstream handler can still read.
func peekBody(r *http.Request) ([]byte, io.ReadCloser, error) {
	if r.Body == nil {
		return nil, http.NoBody, nil
	}
	buf, err := io.ReadAll(io.LimitReader(r.Body, maxAuthBodyPeek))
	_ = r.Body.Close()
	replay := io.NopCloser(bytes.NewReader(buf))
	if err != nil {
		return buf, replay, err
	}
	return buf, replay, nil
}

// writeAuthError emits a JSON refusal. The message says what is required,
// never what was expected: no token value ever appears in a response.
func writeAuthError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// isLoopbackHost reports whether every address host resolves to is a loopback
// address -- the condition under which an unauthenticated server is
// acceptable.
//
// An empty host means "every interface", which is the opposite of loopback. A
// name that resolves to a mix of loopback and routable addresses is not
// loopback either: one routable address is enough to expose the server.
func isLoopbackHost(host string) bool {
	host = strings.TrimSpace(host)
	if host == "" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	addrs, err := net.LookupIP(host)
	if err != nil || len(addrs) == 0 {
		// A host that cannot be resolved cannot be shown to be loopback.
		return false
	}
	for _, ip := range addrs {
		if !ip.IsLoopback() {
			return false
		}
	}
	return true
}
