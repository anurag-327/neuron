package docker

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/anurag-327/neuron/pkg/sandbox"
)

//
// -----------------------------------------------------------------------------
//  Language metadata (single source of truth)
// -----------------------------------------------------------------------------

type FileNames struct {
	BaseName string // "main" or "Main"
	FullName string // "main.cpp"
	PathBase string // "/host/job/main"
	PathFull string // "/host/job/main.cpp"
}

type LangConfig struct {
	DockerImage string
	BaseName    string // main, Main
	Ext         string // cpp, py, java, go, js
	Cmd         func(n FileNames) string
}

var Langs = map[string]LangConfig{
	"cpp": {
		DockerImage: "gcc:latest",
		BaseName:    "main",
		Ext:         "cpp",
		Cmd: func(n FileNames) string {
			return fmt.Sprintf("g++ %s -o %s && ./%s",
				n.FullName, n.BaseName, n.BaseName)
		},
	},
	"go": {
		DockerImage: "golang:1.23-alpine",
		BaseName:    "main",
		Ext:         "go",
		Cmd: func(n FileNames) string {
			return fmt.Sprintf("go run %s", n.FullName)
		},
	},
	"python": {
		DockerImage: "python:3.12-alpine",
		BaseName:    "main",
		Ext:         "py",
		Cmd: func(n FileNames) string {
			return fmt.Sprintf("python3 %s", n.FullName)
		},
	},
	"java": {
		DockerImage: "openjdk:21-jdk-slim",
		BaseName:    "Main",
		Ext:         "java",
		Cmd: func(n FileNames) string {
			return fmt.Sprintf("javac %s && java %s",
				n.FullName, n.BaseName)
		},
	},
	"js": {
		DockerImage: "node:22-alpine",
		BaseName:    "main",
		Ext:         "js",
		Cmd: func(n FileNames) string {
			return fmt.Sprintf("node %s", n.FullName)
		},
	},
}

//
// -----------------------------------------------------------------------------
//  Helper Functions
// -----------------------------------------------------------------------------

func GetLanguageConfig(language string) (LangConfig, error) {
	cfg, ok := Langs[language]
	if !ok {
		return LangConfig{}, fmt.Errorf("unsupported language: %s", language)
	}
	return cfg, nil
}

func BuildFileNames(basePath string, cfg LangConfig) FileNames {
	full := cfg.BaseName + "." + cfg.Ext
	return FileNames{
		BaseName: cfg.BaseName,
		FullName: full,
		PathBase: filepath.Join(basePath, cfg.BaseName),
		PathFull: filepath.Join(basePath, full),
	}
}

func DetectError(language, stdout, stderr string) (sandbox.SandboxError, string) {

	// Combine for signatures that may appear in stdout (JS, Go, Python, Java)
	c := stdout + "\n" + stderr
	s := stderr // compiler errors always in stderr

	switch language {

	// ---------------------------------------------------------------------
	// C++ (g++)
	// ---------------------------------------------------------------------
	case "cpp":
		if strings.Contains(s, "error:") ||
			strings.Contains(s, "fatal error:") ||
			strings.Contains(s, "undefined reference") {
			return sandbox.ErrCompilationError, sandbox.MsgCompilationError
		}
		// C++ runtime errors only printed in stderr (segfaults, abort)
		if isMeaningfulRuntimeErrorCPP(stderr) {
			return sandbox.ErrRuntimeError, sandbox.MsgRuntimeError
		}

	// ---------------------------------------------------------------------
	// Go (go run)
	// ---------------------------------------------------------------------
	case "go":
		if strings.Contains(s, "undefined:") ||
			strings.Contains(s, "cannot use") ||
			strings.Contains(s, "no required module") {
			return sandbox.ErrCompilationError, sandbox.MsgCompilationError
		}
		if strings.Contains(c, "panic:") ||
			strings.Contains(c, "runtime error:") {
			return sandbox.ErrRuntimeError, sandbox.MsgRuntimeError
		}

	// ---------------------------------------------------------------------
	// Python
	// ---------------------------------------------------------------------
	case "python":
		if strings.Contains(s, "SyntaxError") ||
			strings.Contains(s, "IndentationError") {
			return sandbox.ErrCompilationError, sandbox.MsgCompilationError
		}
		if strings.Contains(c, "Traceback (most recent call last):") {
			return sandbox.ErrRuntimeError, sandbox.MsgRuntimeError
		}

	// ---------------------------------------------------------------------
	// Java
	// ---------------------------------------------------------------------
	case "java":
		if strings.Contains(s, "error:") ||
			strings.Contains(s, "cannot find symbol") ||
			strings.Contains(s, "symbol not found") {
			return sandbox.ErrCompilationError, sandbox.MsgCompilationError
		}
		if strings.Contains(c, "Exception in thread") {
			return sandbox.ErrRuntimeError, sandbox.MsgRuntimeError
		}

	// ---------------------------------------------------------------------
	// JavaScript (Node.js)
	// ---------------------------------------------------------------------
	case "js":
		if strings.Contains(s, "SyntaxError:") {
			return sandbox.ErrCompilationError, sandbox.MsgCompilationError
		}
		if strings.Contains(c, "TypeError:") ||
			strings.Contains(c, "ReferenceError:") ||
			strings.Contains(c, "UnhandledPromiseRejectionWarning") {
			return sandbox.ErrRuntimeError, sandbox.MsgRuntimeError
		}
	}

	// ---------------------------------------------------------------------
	// Fallback:
	// Return runtime error ONLY if stderr contains meaningful error text.
	// ---------------------------------------------------------------------
	if isMeaningfulRuntimeErrorGeneric(stderr) {
		return sandbox.ErrRuntimeError, sandbox.MsgRuntimeError
	}

	return "", ""
}

func isMeaningfulRuntimeErrorCPP(stderr string) bool {
	// C++ runtime crashes often include these
	return strings.Contains(stderr, "Segmentation fault") ||
		strings.Contains(stderr, "core dumped") ||
		strings.Contains(stderr, "abort")
}

func isMeaningfulRuntimeErrorGeneric(stderr string) bool {
	// Ignore warnings
	if strings.Contains(stderr, "warning") || strings.Contains(stderr, "Warning") {
		return false
	}
	if strings.TrimSpace(stderr) == "" {
		return false
	}
	return true
}
