package sandbox

import "context"

type Runner interface {
	Run(ctx context.Context, basePath, code, input, language string) (string, string, SandboxError, string)
}
