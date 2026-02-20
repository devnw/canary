package canary

const CommentPrefix = "<!--"
const CommentSuffix = "-->"

const CanaryKey = "CANARY"

const CanaryStart = "START"
const CanaryEnd = "END"

type Canary struct {
	Key   string
	Start string
	End   string
}
