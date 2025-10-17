// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-118; FEATURE="EmbeddedTemplates"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17
package embedded

import "embed"

//go:embed all:base/.canary
//go:embed all:.canary
//go:embed base/.canaryignore
var CanaryFS embed.FS
