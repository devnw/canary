// Copyright (c) 2024 by CodePros.
//
// This software is proprietary information of CodePros.
// Unauthorized use, copying, modification, distribution, and/or
// disclosure is strictly prohibited, except as provided under the terms
// of the commercial license agreement you have entered into with
// CodePros.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact CodePros at info@codepros.org.

// CANARY: REQ=CBIN-118; FEATURE="EmbeddedTemplates"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
package embedded

import "embed"

//go:embed all:base/.canary
//go:embed base/.canaryignore
var CanaryFS embed.FS
