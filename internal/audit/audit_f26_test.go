// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"os"
	"regexp"
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

// TestAuditF26 covers F-26: the CI configuration must pin every dependency to
// an immutable reference and confine release credentials to the release job.
// A mutable `:latest` image tag or a `@v1.43.0` component include is a supply
// chain hole; an `id_tokens`/Vault block on a build or test job leaks release
// credentials to code that has no business holding them. This asserts the
// static shape a hardened pipeline must have.
func TestAuditF26(t *testing.T) {
	b, err := os.ReadFile(repoPath(t, ".gitlab-ci.yml"))
	if err != nil {
		t.Fatalf("read .gitlab-ci.yml: %v", err)
	}
	s := string(b)
	if strings.Contains(s, ":latest") {
		t.Fatal("mutable :latest tag in CI")
	}
	sha40 := regexp.MustCompile(`@[0-9a-f]{40}$`)
	for _, line := range strings.Split(s, "\n") {
		l := strings.TrimSpace(line)
		if strings.HasPrefix(l, "component:") && !sha40.MatchString(l) {
			t.Fatalf("component not SHA-pinned: %s", l)
		}
		if strings.HasPrefix(l, "image:") && !strings.Contains(l, "@sha256:") && !strings.Contains(l, "$") {
			t.Fatalf("image not digest-pinned: %s", l)
		}
	}
	if strings.Contains(s, "credentials loaded for user") {
		t.Fatal("credential echo present")
	}
	// id_tokens must appear only under the release job -- parse yaml minimally.
	var doc map[string]any
	if err := yaml.Unmarshal(b, &doc); err != nil {
		t.Fatalf("parse yaml: %v", err)
	}
	for name, v := range doc {
		job, ok := v.(map[string]any)
		if !ok {
			continue
		}
		if _, has := job["id_tokens"]; has && !strings.Contains(name, "release") {
			t.Fatalf("job %s holds id_tokens", name)
		}
	}
}
