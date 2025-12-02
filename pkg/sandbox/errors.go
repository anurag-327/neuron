package sandbox

type SandboxError string
type SandboxErrorMessage string

const (
	ErrTLE              SandboxError = "TLE"
	ErrMLE              SandboxError = "MLE"
	ErrCompilationError SandboxError = "CompilationError"
	ErrRuntimeError     SandboxError = "RuntimeError"
	ErrSandboxError     SandboxError = "SandboxError"
	ErrInternalError    SandboxError = "InternalError"
)

var (
	MsgTLE              SandboxErrorMessage = "Time Limit Exceeded: the program ran longer than allowed."
	MsgMLE              SandboxErrorMessage = "Memory Limit Exceeded: the program used more memory than allowed."
	MsgCompilationError SandboxErrorMessage = "Compilation failed. See error details."
	MsgRuntimeError     SandboxErrorMessage = "Runtime Error: the program crashed during execution."
	MsgSandboxError     SandboxErrorMessage = "Sandbox Error: execution environment failed."
	MsgInternalError    SandboxErrorMessage = "Internal Error: something went wrong on the server."
)
