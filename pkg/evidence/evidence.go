// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package evidence holds strict parsing of CANARY evidence records and the
// single completion function ("has every claimed requirement been proven at
// the current commit") that downstream verification consumes. It is
// self-contained: stdlib only, no imports from the rest of this repo, so any
// package may depend on it without creating a cycle.
package evidence

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"time"
)

// Record is one evidence claim: a single passing test run, tying a
// requirement's feature/aspect to a commit via a reproducible command and an
// observed result.
type Record struct {
	ProjectID      string `json:"project_id"`
	RequirementID  string `json:"requirement_id"`
	Feature        string `json:"feature"`
	Aspect         string `json:"aspect"`
	TestID         string `json:"test_id"`
	Command        string `json:"command"`
	Result         string `json:"result"`      // must be exactly "PASS"
	CommitSHA      string `json:"commit_sha"`  // exactly 40 lowercase hex
	ObservedAt     string `json:"observed_at"` // RFC3339, must be UTC (Z or +00:00)
	Runner         string `json:"runner"`
	ArtifactDigest string `json:"artifact_digest"` // "sha256:" + 64 lowercase hex
}

// File is the top-level shape of an evidence document.
type File struct {
	SchemaVersion int      `json:"schema_version"` // must be 1
	Records       []Record `json:"records"`
}

var (
	commitSHAPattern      = regexp.MustCompile(`^[0-9a-f]{40}$`)
	artifactDigestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

// Parse strictly decodes an evidence file from r. It rejects, with an error
// naming the offending record index and field: unknown fields anywhere,
// duplicate fields anywhere (encoding/json silently keeps the last value on
// a duplicate key, so this is caught with an explicit token walk before the
// struct decode), a schema_version other than 1, any empty required field, a
// result other than exactly "PASS", and a malformed commit_sha,
// artifact_digest, or observed_at.
func Parse(r io.Reader) (*File, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("evidence: read: %w", err)
	}

	// Pass 1: token-walk the raw bytes to catch duplicate keys in any
	// object -- something encoding/json will not do for us.
	if err := checkNoDuplicateKeys(data); err != nil {
		return nil, fmt.Errorf("evidence: %w", err)
	}

	// Pass 2: strict struct decode; unknown fields are a hard error.
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var f File
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("evidence: decode: %w", err)
	}
	if dec.More() {
		return nil, fmt.Errorf("evidence: trailing data after JSON value")
	}

	// Pass 3: semantic validation.
	if f.SchemaVersion != 1 {
		return nil, fmt.Errorf("evidence: schema_version: want 1, got %d", f.SchemaVersion)
	}
	for i, rec := range f.Records {
		if err := validateRecord(rec); err != nil {
			return nil, fmt.Errorf("evidence: records[%d].%w", i, err)
		}
	}

	return &f, nil
}

// Load reads and strictly parses the evidence file at path. A missing file
// returns an error wrapping fs.ErrNotExist so callers can distinguish it
// (via errors.Is) from a malformed one.
func Load(path string) (*File, error) {
	fh, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("evidence: load %s: %w", path, err)
		}
		return nil, fmt.Errorf("evidence: load %s: %w", path, err)
	}
	defer fh.Close()
	return Parse(fh)
}

// validateRecord checks the semantic rules for one record. The returned
// error, when non-nil, has the shape "<field>: <reason>" so Parse can prefix
// it with "records[<i>]." to produce the indexed diagnostic.
func validateRecord(rec Record) error {
	type field struct {
		name  string
		value string
	}
	required := []field{
		{"project_id", rec.ProjectID},
		{"requirement_id", rec.RequirementID},
		{"feature", rec.Feature},
		{"aspect", rec.Aspect},
		{"test_id", rec.TestID},
		{"command", rec.Command},
		{"runner", rec.Runner},
	}
	for _, f := range required {
		if f.value == "" {
			return fmt.Errorf("%s: must not be empty", f.name)
		}
	}
	if rec.Result != "PASS" {
		return fmt.Errorf("result: want exactly %q, got %q", "PASS", rec.Result)
	}
	if !commitSHAPattern.MatchString(rec.CommitSHA) {
		return fmt.Errorf("commit_sha: want 40 lowercase hex")
	}
	if !artifactDigestPattern.MatchString(rec.ArtifactDigest) {
		return fmt.Errorf("artifact_digest: want \"sha256:\" + 64 lowercase hex")
	}
	if err := validateObservedAtUTC(rec.ObservedAt); err != nil {
		return fmt.Errorf("observed_at: %w", err)
	}
	return nil
}

// validateObservedAtUTC requires observed_at to be a valid RFC3339 timestamp
// that is unambiguously UTC: the string must end in "Z" or "+00:00", AND the
// parsed offset must be zero (a belt-and-suspenders check against any
// alternate zero-offset spelling RFC3339 might allow).
func validateObservedAtUTC(s string) error {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fmt.Errorf("want RFC3339: %w", err)
	}
	if _, offset := t.Zone(); offset != 0 {
		return fmt.Errorf("want UTC (zero offset), got %q", s)
	}
	if !strings.HasSuffix(s, "Z") && !strings.HasSuffix(s, "+00:00") {
		return fmt.Errorf("want RFC3339 UTC ending in Z or +00:00, got %q", s)
	}
	return nil
}

// checkNoDuplicateKeys walks the raw JSON token stream and returns an error
// if any object -- at the top level or nested inside it (i.e. inside each
// record) -- repeats a key. encoding/json's normal Decode silently keeps the
// last value for a duplicate key, which would hide evidence tampering, so
// this walk runs as an independent first pass over the same bytes.
func checkNoDuplicateKeys(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dupWalkValue(dec, ""); err != nil {
		return err
	}
	return nil
}

// dupWalkValue reads exactly one JSON value from dec (consuming it fully)
// and recurses into objects/arrays checking for duplicate object keys.
// path is a human-readable location used in error messages, e.g.
// "records[3]".
func dupWalkValue(dec *json.Decoder, path string) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := tok.(json.Delim)
	if !ok {
		// scalar value (string, number, bool, null): nothing to recurse into.
		return nil
	}
	switch delim {
	case '{':
		seen := make(map[string]bool)
		for dec.More() {
			keyTok, err := dec.Token()
			if err != nil {
				return err
			}
			key, ok := keyTok.(string)
			if !ok {
				return fmt.Errorf("expected object key, got %v", keyTok)
			}
			if seen[key] {
				if path == "" {
					return fmt.Errorf("duplicate field %q", key)
				}
				return fmt.Errorf("%s: duplicate field %q", path, key)
			}
			seen[key] = true
			childPath := key
			if path != "" {
				childPath = path + "." + key
			}
			if err := dupWalkValue(dec, childPath); err != nil {
				return err
			}
		}
		// consume the closing '}'
		if _, err := dec.Token(); err != nil {
			return err
		}
	case '[':
		i := 0
		for dec.More() {
			childPath := fmt.Sprintf("%s[%d]", path, i)
			if err := dupWalkValue(dec, childPath); err != nil {
				return err
			}
			i++
		}
		// consume the closing ']'
		if _, err := dec.Token(); err != nil {
			return err
		}
	}
	return nil
}
