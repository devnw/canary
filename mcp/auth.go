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

// CANARY: REQ=CP-283; FEATURE="MCPAuth"; ASPECT=Security; STATUS=TESTED; TEST=TestAuthLoopbackNoTokensAllows,TestAuthNonLoopbackWithoutTokensDenies,TestAuthRequiresBearerWhenConfigured,TestAuthReadTokenRefusedOnMutatingTool,TestAuthNeverEchoesTokens,TestScopesFor,TestIsLoopbackHost; UPDATED=2026-08-30

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
// call `prioritize`.
func authMiddleware(cfg authConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scopes, ok := cfg.scopesFor(r.Header.Get("Authorization"))
		if !ok {
			w.Header().Set("WWW-Authenticate", `Bearer realm="canary-mcp"`)
			writeAuthError(w, http.StatusUnauthorized, "unauthorized: a valid Bearer token is required")
			return
		}

		if !scopes.Mutate && r.Method == http.MethodPost {
			name, body, err := peekToolName(r)
			if err != nil {
				writeAuthError(w, http.StatusBadRequest, "malformed request body")
				return
			}
			r.Body = body
			if name != "" && cfg.mutating[name] {
				writeAuthError(w, http.StatusForbidden,
					"forbidden: tool "+name+" requires the read+write token")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// peekToolName reads the JSON-RPC envelope far enough to learn which tool is
// being called, and returns a body the downstream handler can still read.
//
// A body that is not a tools/call -- initialization, a list, a batch, an
// unparseable blob -- yields an empty name and is passed through untouched.
// Rejecting it here would make this middleware a second, worse JSON-RPC
// implementation; the SDK is the one that decides what a valid request is.
func peekToolName(r *http.Request) (string, io.ReadCloser, error) {
	if r.Body == nil {
		return "", http.NoBody, nil
	}
	buf, err := io.ReadAll(io.LimitReader(r.Body, maxAuthBodyPeek))
	_ = r.Body.Close()
	if err != nil {
		return "", io.NopCloser(bytes.NewReader(buf)), err
	}
	replay := io.NopCloser(bytes.NewReader(buf))

	var envelope struct {
		Method string `json:"method"`
		Params struct {
			Name string `json:"name"`
		} `json:"params"`
	}
	if err := json.Unmarshal(buf, &envelope); err != nil {
		return "", replay, nil
	}
	if envelope.Method != "tools/call" {
		return "", replay, nil
	}
	return envelope.Params.Name, replay, nil
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
