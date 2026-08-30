// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package storage

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"
)

func reserveTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := OpenRW(filepath.Join(t.TempDir(), ".canary", "canary.db"))
	if err != nil {
		t.Fatalf("OpenRW: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestReserveIDSequence proves the series is contiguous and per-prefix.
func TestReserveIDSequence(t *testing.T) {
	db := reserveTestDB(t)

	for _, want := range []string{"BUG-API-001", "BUG-API-002", "BUG-API-003"} {
		got, err := db.ReserveID("proj", "BUG-API")
		if err != nil {
			t.Fatalf("ReserveID: %v", err)
		}
		if got != want {
			t.Fatalf("ReserveID = %q, want %q", got, want)
		}
	}

	got, err := db.ReserveID("proj", "BUG-CLI")
	if err != nil {
		t.Fatalf("ReserveID: %v", err)
	}
	if got != "BUG-CLI-001" {
		t.Fatalf("second series started at %q, want BUG-CLI-001", got)
	}

	// A different project counts separately.
	got, err = db.ReserveID("other", "BUG-API")
	if err != nil {
		t.Fatalf("ReserveID: %v", err)
	}
	if got != "BUG-API-001" {
		t.Fatalf("other project's series started at %q, want BUG-API-001", got)
	}
}

// TestReserveIDConcurrent proves the allocation is atomic under contention.
func TestReserveIDConcurrent(t *testing.T) {
	db := reserveTestDB(t)

	const n = 50
	var wg sync.WaitGroup
	var mu sync.Mutex
	seen := map[string]bool{}
	errs := make([]error, 0, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := db.ReserveID("proj", "BUG-API")
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				return
			}
			if seen[id] {
				errs = append(errs, errors.New("duplicate id "+id))
				return
			}
			seen[id] = true
		}()
	}
	wg.Wait()

	for _, err := range errs {
		t.Error(err)
	}
	if len(seen) != n {
		t.Fatalf("got %d unique ids, want %d", len(seen), n)
	}
}

// TestReserveIDSkipsIndexedIDs proves an id already carried by a token is
// never re-issued, including one that entered the index from source rather
// than through a reservation.
func TestReserveIDSkipsIndexedIDs(t *testing.T) {
	db := reserveTestDB(t)

	if err := db.UpsertToken(&Token{
		ReqID: "BUG-API-042", Feature: "from source", Aspect: "API", Status: "OPEN",
		FilePath: "a.go", LineNumber: 1, ProjectID: "proj",
	}); err != nil {
		t.Fatalf("UpsertToken: %v", err)
	}

	got, err := db.ReserveID("proj", "BUG-API")
	if err != nil {
		t.Fatalf("ReserveID: %v", err)
	}
	if got != "BUG-API-043" {
		t.Fatalf("ReserveID = %q, want BUG-API-043", got)
	}
}

// TestReserveIDIgnoresForeignSuffixes proves an id that merely shares the
// prefix but is not a member of the numeric series does not move the count.
func TestReserveIDIgnoresForeignSuffixes(t *testing.T) {
	db := reserveTestDB(t)

	if err := db.UpsertToken(&Token{
		ReqID: "BUG-API-9x", Feature: "not a member", Aspect: "API", Status: "OPEN",
		FilePath: "a.go", LineNumber: 1, ProjectID: "proj",
	}); err != nil {
		t.Fatalf("UpsertToken: %v", err)
	}

	got, err := db.ReserveID("proj", "BUG-API")
	if err != nil {
		t.Fatalf("ReserveID: %v", err)
	}
	if got != "BUG-API-001" {
		t.Fatalf("ReserveID = %q, want BUG-API-001", got)
	}
}

// TestReserveIDRequiresProject proves an unscoped reservation is refused.
func TestReserveIDRequiresProject(t *testing.T) {
	db := reserveTestDB(t)
	if _, err := db.ReserveID("  ", "BUG-API"); !errors.Is(err, ErrProjectRequired) {
		t.Fatalf("err = %v, want ErrProjectRequired", err)
	}
	if _, err := db.ReserveID("proj", ""); err == nil {
		t.Fatal("empty prefix must be refused")
	}
}

// TestReserveIDRefusesReadOnly proves a reader cannot allocate.
func TestReserveIDRefusesReadOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".canary", "canary.db")
	rw, err := OpenRW(path)
	if err != nil {
		t.Fatalf("OpenRW: %v", err)
	}
	if err := rw.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	ro, err := OpenRO(path)
	if err != nil {
		t.Fatalf("OpenRO: %v", err)
	}
	defer func() { _ = ro.Close() }()

	if _, err := ro.ReserveID("proj", "BUG-API"); !errors.Is(err, ErrReadOnlyReservation) {
		t.Fatalf("err = %v, want ErrReadOnlyReservation", err)
	}
}
