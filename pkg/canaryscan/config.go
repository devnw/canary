package canaryscan

import "regexp"

// Config holds scanner run options. Used by both the standalone tools/canary
// binary and the canary scan CLI subcommand so scan works from any CWD.
type Config struct {
	Root        string
	Out         string
	CSV         string
	VerifyPath  string
	Strict      bool
	SkipRegex   *regexp.Regexp
	UpdateStale bool
	ProjectOnly bool
	// StaleDays overrides the staleness window (in days) used by --strict and
	// --update-stale. When <= 0, the effective threshold falls back to
	// .canary/project.yaml's verification.staleness_days, then to 30 days.
	StaleDays int
}
