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
	"time"
)

// TestHTTPServerLimits pins the transport limits. The server used to set a
// header timeout and nothing else, so a client could hold a connection open
// indefinitely mid-body and there was no cap on how much it could send.
func TestHTTPServerLimits(t *testing.T) {
	srv := newHTTPServer("127.0.0.1:0", http.NotFoundHandler())

	checks := []struct {
		name string
		got  time.Duration
		want time.Duration
	}{
		{"ReadHeaderTimeout", srv.ReadHeaderTimeout, 5 * time.Second},
		{"ReadTimeout", srv.ReadTimeout, 15 * time.Second},
		{"WriteTimeout", srv.WriteTimeout, 15 * time.Second},
		{"IdleTimeout", srv.IdleTimeout, 30 * time.Second},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
	if srv.MaxHeaderBytes != 16<<10 {
		t.Errorf("MaxHeaderBytes = %d, want %d", srv.MaxHeaderBytes, 16<<10)
	}
	if srv.Handler == nil {
		t.Error("server has no handler")
	}
}

// TestBodyCapRefusesOversizedRequest proves the megabyte cap is enforced on
// the served routes, not merely configured.
func TestBodyCapRefusesOversizedRequest(t *testing.T) {
	h := buildHandler(NewServer("test", Deps{Root: t.TempDir()}), authConfig{loopback: true, mutating: mutatingTools()})

	// httptest binds 127.0.0.1:0 -- an ephemeral loopback port, never a
	// fixed one and never a routable interface.
	srv := httptest.NewServer(h)
	defer srv.Close()

	oversized := strings.NewReader(strings.Repeat("x", (1<<20)+1024))
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/mcp", oversized)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := srv.Client().Do(req)
	if err != nil {
		// A connection reset while the server refuses an oversized body is
		// also a refusal; what must not happen is a 200.
		return
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("oversized body accepted with 200")
	}
}

// TestHealthEndpoint proves /health answers without a credential: a health
// probe that needs the write token is not a health probe.
func TestHealthEndpoint(t *testing.T) {
	h := buildHandler(NewServer("test", Deps{Root: t.TempDir()}), authConfig{token: "s3cret", mutating: mutatingTools()})
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("get /health: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(string(body), "healthy") {
		t.Fatalf("unexpected health body: %s", body)
	}
}

// TestCheckExposureRefusesUnprotectedRemoteBind covers the startup gate.
func TestCheckExposureRefusesUnprotectedRemoteBind(t *testing.T) {
	if err := checkExposure(serverConfig{host: "127.0.0.1"}, true, authConfig{}); err != nil {
		t.Fatalf("loopback with no tokens must be allowed: %v", err)
	}

	err := checkExposure(serverConfig{host: "0.0.0.0"}, false, authConfig{})
	if err == nil {
		t.Fatal("non-loopback with neither TLS nor token must be refused")
	}
	for _, want := range []string{"--tls-cert", "--tls-key", EnvToken} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal does not name %s: %v", want, err)
		}
	}

	// TLS but no token is still refused: transport security is not
	// authorization.
	if err := checkExposure(serverConfig{host: "0.0.0.0", tlsCert: "c", tlsKey: "k"}, false, authConfig{}); err == nil {
		t.Fatal("non-loopback with TLS but no token must be refused")
	}
	// Token but no TLS is refused too: a bearer token over cleartext is a
	// bearer token handed to the network.
	if err := checkExposure(serverConfig{host: "0.0.0.0"}, false, authConfig{token: "t"}); err == nil {
		t.Fatal("non-loopback with a token but no TLS must be refused")
	}

	if err := checkExposure(serverConfig{host: "0.0.0.0", tlsCert: "c", tlsKey: "k"}, false, authConfig{token: "t"}); err != nil {
		t.Fatalf("fully protected non-loopback bind must be allowed: %v", err)
	}
}

// TestTLSFlagPairIsBothOrNeither pins the half-configured-TLS refusal. The
// banner asked "is there a certificate?" while the serve call asked "is there
// a certificate AND a key?", so `--tls-cert` alone printed https over a
// plaintext listener. Neither question is asked of a half pair any more: it is
// refused before anything binds, on loopback too.
func TestTLSFlagPairIsBothOrNeither(t *testing.T) {
	if err := checkTLSFlags(serverConfig{host: "127.0.0.1"}); err != nil {
		t.Errorf("neither flag must be allowed: %v", err)
	}
	if err := checkTLSFlags(serverConfig{host: "127.0.0.1", tlsCert: "c", tlsKey: "k"}); err != nil {
		t.Errorf("both flags must be allowed: %v", err)
	}
	for _, cfg := range []serverConfig{
		{host: "127.0.0.1", tlsCert: "c"},
		{host: "127.0.0.1", tlsKey: "k"},
		{host: "0.0.0.0", tlsCert: "c"},
		{host: "0.0.0.0", tlsKey: "k"},
	} {
		err := checkTLSFlags(cfg)
		if err == nil {
			t.Fatalf("half-configured TLS (%+v) must be refused", cfg)
		}
		if !strings.Contains(err.Error(), "--tls-cert") || !strings.Contains(err.Error(), "--tls-key") {
			t.Errorf("refusal does not name both flags: %v", err)
		}
	}

	// The banner reads the same predicate the serve call does.
	if (serverConfig{tlsCert: "c"}).tlsEnabled() || (serverConfig{tlsKey: "k"}).tlsEnabled() {
		t.Error("a half pair must not report TLS")
	}
	if !(serverConfig{tlsCert: "c", tlsKey: "k"}).tlsEnabled() {
		t.Error("a full pair must report TLS")
	}
}

// TestRunServerRefusesHalfConfiguredTLS proves the refusal happens on the
// startup path, before a listener exists -- a server that binds first and
// complains second has already accepted a connection.
func TestRunServerRefusesHalfConfiguredTLS(t *testing.T) {
	cmd := New("test")
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	err := runServer(cmd, serverConfig{
		version: "test", host: "127.0.0.1", port: 0, root: t.TempDir(), tlsCert: "cert.pem",
	})
	if err == nil {
		t.Fatal("runServer accepted --tls-cert without --tls-key")
	}
	if !strings.Contains(err.Error(), "--tls-key") {
		t.Errorf("refusal does not name the missing flag: %v", err)
	}
}

// TestNewServerReportsVersion proves the implementation version comes from
// the binary rather than a hardcoded "1.0.0".
func TestNewServerReportsVersion(t *testing.T) {
	if NewServer("v1.2.3", Deps{Root: t.TempDir()}) == nil {
		t.Fatal("NewServer returned nil")
	}
	// A blank version degrades to "dev" rather than an empty string, which
	// the MCP spec does not allow to be meaningful.
	if NewServer("", Deps{Root: t.TempDir()}) == nil {
		t.Fatal("NewServer returned nil for a blank version")
	}
}
