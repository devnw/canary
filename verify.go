// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.


package main

// CANARY: REQ=CBIN-102; FEATURE="VerifyGate"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_102_CLI_Verify; BENCH=BenchmarkCANARY_CBIN_102_CLI_Verify; OWNER=canary; UPDATED=2025-09-20
import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// claim means "this REQ is claimed Implemented/Complete in GAP"
type claim struct {
	REQ         string
	Implemented bool
	RawLine     string
}

var reqRe = regexp.MustCompile(`REQ[\-‑]GQL[\-‑](\d{3,})`)
var implementedRe = regexp.MustCompile(`\b(Implemented|Complete|✅)\b`)
var notImplRe = regexp.MustCompile(`\b(STUB|NOT IMPLEMENTED|❌|◻)\b`)

func ParseGAPClaims(path string) (map[string]claim, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	claims := map[string]claim{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		ids := reqRe.FindAllStringSubmatch(line, -1)
		if len(ids) == 0 {
			continue
		}
		for _, m := range ids {
			id := "REQ-GQL-" + m[1]
			c := claims[id]
			c.REQ = id
			c.RawLine = line
			// A very conservative read: mark Implemented if line suggests it and no obvious NOT markers.
			if implementedRe.MatchString(line) && !notImplRe.MatchString(line) {
				c.Implemented = true
			}
			claims[id] = c
		}
	}
	return claims, sc.Err()
}

func VerifyClaims(rep report, claims map[string]claim) error {
	var errs []string
	evidence := map[string]bool{} // REQ -> has TESTED/BENCHED
	for _, r := range rep.Requirements {
		ok := false
		for _, f := range r.Features {
			if f.Status == "TESTED" || f.Status == "BENCHED" {
				ok = true
				break
			}
		}
		evidence[r.ID] = ok
	}
	for id, c := range claims {
		if !c.Implemented {
			continue
		}
		if !evidence[normalizeReq(id)] {
			errs = append(errs, fmt.Sprintf("REQ=%s claimed Implemented without CANARY TESTED/BENCHED (%s)", id, strings.TrimSpace(c.RawLine)))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}
