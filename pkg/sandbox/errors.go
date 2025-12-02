package sandbox

type SandboxError string

const (
	ErrTLE              SandboxError = "TLE"
	ErrMLE              SandboxError = "MLE"
	ErrCompilationError SandboxError = "CompilationError"
	ErrRuntimeError     SandboxError = "RuntimeError"
	ErrSandboxError     SandboxError = "SandboxError"
	ErrInternalError    SandboxError = "InternalError"
)

var (
	MsgTLE              = "Time Limit Exceeded: the program ran longer than allowed."
	MsgMLE              = "Memory Limit Exceeded: the program used more memory than allowed."
	MsgCompilationError = "Compilation failed. See error details."
	MsgRuntimeError     = "Runtime Error: the program crashed during execution."
	MsgSandboxError     = "Sandbox Error: execution environment failed."
	MsgInternalError    = "Internal Error: something went wrong on the server."
)
