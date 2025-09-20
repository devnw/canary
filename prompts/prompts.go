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

func All() map[string]string { return map[string]string{"init": Init, "policy": Policy, "requirements": Requirements, "evaluate": Evaluate} }
