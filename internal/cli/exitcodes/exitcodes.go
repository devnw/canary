package exitcodes

// Canonical exit codes.
const (
	CodeOK      = 0
	CodeVerify  = 2 // verification/staleness failure
	CodeParseIO = 3 // parse / IO error
)

// Coder allows errors to carry exit codes.
type Coder interface {
	error
	ExitCode() int
}

type exitErr struct {
	code int
	msg  string
}

func (e *exitErr) Error() string { return e.msg }
func (e *exitErr) ExitCode() int { return e.code }

// New creates a coded error.
func New(code int, msg string) error { return &exitErr{code: code, msg: msg} }
