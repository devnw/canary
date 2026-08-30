// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package audit

import (
	"sync"
	"testing"

	"devnw.dev/canary/pkg/storage"
)

// TestAuditF18 proves ID allocation is transactional: a hundred concurrent
// reservations of the same series hand out a hundred distinct ids.
//
// The read-max-then-increment generator this replaces could not do that. Two
// callers that read the same maximum both computed the same "next" number and
// both wrote it, so two bugs filed at the same moment shared an id -- and the
// second UpsertToken silently overwrote the first, because req_id is part of
// the row's identity.
func TestAuditF18(t *testing.T) {
	db := openTestDB(t)

	const n = 100
	var wg sync.WaitGroup
	ids := make(chan string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := db.ReserveID("p", "BUG-API")
			if err != nil {
				t.Errorf("ReserveID: %v", err)
				return
			}
			ids <- id
		}()
	}
	wg.Wait()
	close(ids)

	seen := map[string]bool{}
	for id := range ids {
		if seen[id] {
			t.Fatalf("duplicate id %s", id)
		}
		seen[id] = true
	}
	if len(seen) != n {
		t.Fatalf("got %d unique ids, want %d", len(seen), n)
	}
	if !seen["BUG-API-001"] || !seen["BUG-API-100"] {
		t.Fatalf("reserved ids are not the contiguous 001..100 series: %d ids, missing endpoints", len(seen))
	}
}

// TestAuditF18SeedsPastExistingTokens proves a reservation never re-issues an
// id that is already carried by an indexed token. The reservation table is
// new; the bug series it allocates from is not, so the first reservation in a
// database that predates it has to start above what the index already holds.
func TestAuditF18SeedsPastExistingTokens(t *testing.T) {
	db := openTestDB(t)

	for _, id := range []string{"BUG-API-001", "BUG-API-002", "BUG-API-005"} {
		if err := db.UpsertToken(&storage.Token{
			ReqID: id, Feature: "existing " + id, Aspect: "API", Status: "OPEN",
			FilePath: "main.go", LineNumber: 1, UpdatedAt: "2026-01-01",
			ProjectID: "default",
		}); err != nil {
			t.Fatalf("seed %s: %v", id, err)
		}
	}

	got, err := db.ReserveID("default", "BUG-API")
	if err != nil {
		t.Fatalf("ReserveID: %v", err)
	}
	if got != "BUG-API-006" {
		t.Fatalf("ReserveID = %q, want BUG-API-006 (must clear the highest indexed id)", got)
	}

	// A different series in the same project starts its own count.
	got, err = db.ReserveID("default", "BUG-CLI")
	if err != nil {
		t.Fatalf("ReserveID: %v", err)
	}
	if got != "BUG-CLI-001" {
		t.Fatalf("ReserveID = %q, want BUG-CLI-001", got)
	}
}

// TestAuditF18RefusesUnscopedReservation proves a reservation is refused
// without a project. Numbering is per project; an unscoped reservation would
// silently allocate out of whichever rows happened to share the database.
func TestAuditF18RefusesUnscopedReservation(t *testing.T) {
	db := openTestDB(t)
	if _, err := db.ReserveID("", "BUG-API"); err == nil {
		t.Fatal("ReserveID with no project must be refused")
	}
}
