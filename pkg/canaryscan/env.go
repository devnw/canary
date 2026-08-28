package canaryscan

import (
	"os"
	"time"
)

// RefTimeFromEnv returns time from CANARY_TEST_TIMESTAMP (RFC3339) for tests; zero if unset or invalid.
func RefTimeFromEnv() time.Time {
	s := os.Getenv("CANARY_TEST_TIMESTAMP")
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t.UTC()
}
