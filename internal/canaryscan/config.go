package canaryscan

import "regexp"

// Config holds scanner run options. Used by both the standalone tools/canary
// binary and the canary scan CLI subcommand so scan works from any CWD.
type Config struct {
	Root        string
	Out         string
	CSV         string
	VerifyPath   string
	Strict      bool
	SkipRegex   *regexp.Regexp
	UpdateStale bool
	ProjectOnly bool
}
