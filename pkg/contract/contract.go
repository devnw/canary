// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package contract carries the two facts a refused CLI contract needs on
// both sides of the command boundary: the sentinel a command returns once it
// has printed its JSON refusal, and the exit status main gives it.
//
// It is a package of its own, and not part of pkg/cmds/internal/utils, only
// because cmd/canary lives outside that internal tree and must be able to
// recognise the sentinel Execute hands back.
package contract

import "errors"

// CANARY: REQ=CBIN-307; FEATURE="ProjectScopeCLI"; ASPECT=CLI; STATUS=TESTED; TEST=TestAuditF08CLI,TestProjectRequiredNext,TestProjectRequiredDepsCheck; UPDATED=2026-08-30

// ExitCode is the exit status a contract refusal uses. It is distinct from 1
// so a caller can tell "canary refused because you asked for something it
// will not do" from "the command failed".
const ExitCode = 2

// ErrFailed is the sentinel a command returns once it has already printed its
// contract line. main maps it to ExitCode; nothing prints it, because the
// JSON line on stdout is the entire message.
var ErrFailed = errors.New("contract refused")
