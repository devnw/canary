// CANARY: REQ=CBIN-118; FEATURE="EmbeddedTemplates"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
package embedded

import "embed"

//go:embed all:base/.canary
//go:embed base/.canaryignore
var CanaryFS embed.FS
