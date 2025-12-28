package docker

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/anurag-327/neuron/internal/models"
	"github.com/anurag-327/neuron/internal/registry"
)

func DetectLanguageError(language, stdout, stderr string) (models.SandboxError, string) {

	s := stderr
	c := stdout + "\n" + stderr // Some runtime errors print to stdout

	// C++
	if language == "cpp" {
		// Compiler errors (from g++)
		if strings.Contains(s, "error:") ||
			strings.Contains(s, "fatal error:") ||
			strings.Contains(s, "undefined reference") {
			return models.ErrCompilationError, models.MsgCompilationError
		}

		// Runtime crash detection
		if strings.Contains(s, "Segmentation fault") ||
			strings.Contains(s, "core dumped") ||
			strings.Contains(s, "abort") ||
			strings.Contains(s, "floating point exception") {
			return models.ErrRuntimeError, models.MsgRuntimeError
		}

	}

	// Go
	if language == "go" {
		if strings.Contains(s, "undefined:") ||
			strings.Contains(s, "cannot use") ||
			strings.Contains(s, "no required module") {
			return models.ErrCompilationError, models.MsgCompilationError
		}

		if strings.Contains(c, "panic:") ||
			strings.Contains(c, "runtime error:") {
			return models.ErrRuntimeError, models.MsgRuntimeError
		}
	}

	// Python
	if language == "python" {
		if strings.Contains(s, "SyntaxError") ||
			strings.Contains(s, "IndentationError") {
			return models.ErrCompilationError, models.MsgCompilationError
		}

		if strings.Contains(c, "Traceback (most recent call last):") {
			return models.ErrRuntimeError, models.MsgRuntimeError
		}
	}

	// Java
	if language == "java" {
		if strings.Contains(s, "error:") ||
			strings.Contains(s, "cannot find symbol") ||
			strings.Contains(s, "symbol not found") {
			return models.ErrCompilationError, models.MsgCompilationError
		}

		if strings.Contains(c, "Exception in thread") {
			return models.ErrRuntimeError, models.MsgRuntimeError
		}
	}

	// JavaScript (Node.js)
	if language == "javascript" {
		if strings.Contains(s, "SyntaxError:") {
			return models.ErrCompilationError, models.MsgCompilationError
		}

		if strings.Contains(c, "TypeError:") ||
			strings.Contains(c, "ReferenceError:") ||
			strings.Contains(c, "UnhandledPromiseRejectionWarning") {
			return models.ErrRuntimeError, models.MsgRuntimeError
		}
	}

	// LAST RESORT CHECK — strict runtime detection
	// Only treat stderr as runtime error if it contains *real* crash signals.

	if isMeaningfulRuntimeErrorGeneric(stderr) {
		return models.ErrRuntimeError, models.MsgRuntimeError
	}

	// Everything OK
	return models.ErrRuntimeError, models.MsgRuntimeError
}

func isMeaningfulRuntimeErrorGeneric(stderr string) bool {
	s := strings.ToLower(stderr)

	// ignore if nothing meaningful
	if strings.TrimSpace(s) == "" {
		return false
	}

	// ignore logs like [info], [debug], etc.
	if strings.Contains(s, "[info]") ||
		strings.Contains(s, "[debug]") ||
		strings.Contains(s, "note:") {
		return false
	}

	// ignore warnings
	if strings.Contains(s, "warning") {
		return false
	}

	// real runtime errors
	crashPatterns := []string{
		"segmentation fault",
		"core dumped",
		"panic:",
		"runtime error",
		"traceback (most recent call last):",
		"exception in thread",
		"nullpointerexception",
		"typeerror:",
		"referenceerror:",
		"indexerror:",
		"valueerror:",
		"abort",
		"illegal instruction",
		"floating point exception",
	}

	for _, pat := range crashPatterns {
		if strings.Contains(s, pat) {
			return true
		}
	}

	return false
}

func BuildFileNames(basePath string, cfg registry.LanguageConfig) registry.FileNames {
	full := cfg.BaseName + "." + cfg.Ext
	return registry.FileNames{
		BaseName: cfg.BaseName,
		FullName: full,
		PathBase: filepath.Join(basePath, cfg.BaseName),
		PathFull: filepath.Join(basePath, full),
	}
}

type ResultResponse struct {
	Status       string              `json:"status"`
	ErrorType    models.SandboxError `json:"error_type"`
	ErrorMessage string              `json:"error_message"`
	ExitCode     int64               `json:"error_code"`
	Stdout       string              `json:"stdout"`
	Stderr       string              `json:"stderr"`
}

func ProcessResult(lang string, status int64, stdout, stderr string, jobID string) ResultResponse {
	// Truncate first to prevent massive strings from hitting regex or downstream DB
	const MaxOutputSize = 256 * 1024 // 256KB
	stdout = TruncateOutput(stdout, MaxOutputSize)
	stderr = TruncateOutput(stderr, MaxOutputSize)

	cleanStderr := SanitizeOutput(stderr, jobID)
	cleanStdout := SanitizeOutput(stdout, jobID)

	switch status {
	case 0: // Success
		return ResultResponse{Status: "success", ErrorType: "", ErrorMessage: "", ExitCode: status, Stdout: cleanStdout, Stderr: cleanStderr}

	case 1: // Runtime Error
		errorType, message := DetectLanguageError(lang, cleanStdout, cleanStderr)
		return ResultResponse{Status: "success", ErrorType: errorType, ErrorMessage: message, ExitCode: status, Stdout: cleanStdout, Stderr: cleanStderr}

	case 124, 143, 137: // SIGTERM / OOM / Timeout
		// For TLE/OOM, output is often huge/infinite or irrelevant. Discard it.
		return ResultResponse{Status: "success", ErrorType: models.ErrTLE, ErrorMessage: models.MsgTLE, ExitCode: status, Stdout: "", Stderr: ""}

	case 139, 136, 134: // Segmentation Fault (SIGSEGV)
		return ResultResponse{Status: "success", ErrorType: models.ErrRuntimeError, ErrorMessage: models.MsgRuntimeError, ExitCode: status, Stdout: cleanStdout, Stderr: cleanStderr}

	default:
		fmt.Println("Unknown exit code:", status)
		errorType, message := DetectLanguageError(lang, cleanStdout, cleanStderr)
		return ResultResponse{
			Status:       "error",
			ErrorType:    errorType,
			ErrorMessage: message,
			ExitCode:     status,
			Stdout:       cleanStdout,
			Stderr:       cleanStderr,
		}
	}
}

func SanitizeOutput(input string, jobID string) string {
	// Matches /sandbox/job_.../ and replaces with ./
	re := regexp.MustCompile(`/sandbox/[^/ \n]+/`)
	return re.ReplaceAllString(input, "./")
}

func TruncateOutput(input string, limit int) string {
	if len(input) > limit {
		return input[:limit] + "\n... [Output Truncated]"
	}
	return input
}
