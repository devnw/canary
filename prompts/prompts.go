// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.


package prompts

import _ "embed"

//go:embed sys/init.md
var Init string

//go:embed sys/policy.md
var Policy string

//go:embed sys/requirements.md
var Requirements string

//go:embed sys/evaluate.md
var Evaluate string

func All() map[string]string {
	return map[string]string{"init": Init, "policy": Policy, "requirements": Requirements, "evaluate": Evaluate}
}
