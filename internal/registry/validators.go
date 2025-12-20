package registry

import (
	"fmt"
	"strings"
	"unicode"
)

func ValidateAndSanitizeCpp(code string) error {
	// 1 Size limit
	if len(code) > 256*1024 {
		return fmt.Errorf("c++ code too large (>256KB)")
	}

	// 2 Non-printable characters
	for _, r := range code {
		if !unicode.IsPrint(r) && r != '\n' && r != '\t' {
			return fmt.Errorf("contains invalid characters")
		}
	}

	// 3 Basic language heuristics
	if !strings.Contains(code, "main(") {
		return fmt.Errorf("missing main() function")
	}
	if !strings.Contains(code, "#include") && !strings.Contains(code, "int main(") {
		return fmt.Errorf("not valid C++ source")
	}

	// 4 Dangerous keywords
	blocked := []string{
		"system(", "popen(", "execv", "fork(", "socket", "open(", "ofstream", "ifstream",
		"std::filesystem", "unistd.h", "netinet", "arpa", "winsock", "dirent.h",
	}
	for _, bad := range blocked {
		if strings.Contains(code, bad) {
			return fmt.Errorf("code contains forbidden keyword: %s", bad)
		}
	}
	return nil
}

func ValidateAndSanitizeJS(code string) error {
	// 1. Size limit
	if len(code) > 256*1024 {
		return fmt.Errorf("javascript code too large (>256KB)")
	}

	// 2. Non-printable characters
	for _, r := range code {
		if !unicode.IsPrint(r) && r != '\n' && r != '\t' {
			return fmt.Errorf("contains invalid characters")
		}
	}

	// 4. Dangerous APIs
	blocked := []string{
		"require('child_process')",
		"require(\"child_process\")",
		"exec(",
		"spawn(",
		"fork(",
		"process.exit",
		"fs.writeFile",
		"fs.unlink",
		"fs.rm",
		"net.createServer",
		"dgram.createSocket",
	}

	for _, bad := range blocked {
		if strings.Contains(code, bad) {
			return fmt.Errorf("code contains forbidden keyword: %s", bad)
		}
	}

	return nil
}

func ValidateAndSanitizePython(code string) error {
	// 1. Size limit
	if len(code) > 256*1024 {
		return fmt.Errorf("python code too large (>256KB)")
	}

	// 2. Non-printable characters
	for _, r := range code {
		if !unicode.IsPrint(r) && r != '\n' && r != '\t' {
			return fmt.Errorf("contains invalid characters")
		}
	}

	// 4. Dangerous modules / functions
	blocked := []string{
		"import os",
		"import sys",
		"subprocess",
		"eval(",
		"exec(",
		"open(",
		"__import__",
		"socket",
		"shutil",
		"pickle",
	}

	for _, bad := range blocked {
		if strings.Contains(code, bad) {
			return fmt.Errorf("code contains forbidden keyword: %s", bad)
		}
	}

	return nil
}

func ValidateAndSanitizeJava(code string) error {
	// 1. Size limit
	if len(code) > 256*1024 {
		return fmt.Errorf("java code too large (>256KB)")
	}

	// 2. Non-printable characters
	for _, r := range code {
		if !unicode.IsPrint(r) && r != '\n' && r != '\t' {
			return fmt.Errorf("contains invalid characters")
		}
	}

	// 3. Must contain class and main
	if !strings.Contains(code, "class ") {
		return fmt.Errorf("missing class declaration")
	}
	if !strings.Contains(code, "public static void main") {
		return fmt.Errorf("missing main method")
	}

	// 4. Dangerous APIs
	blocked := []string{
		"Runtime.getRuntime",
		"ProcessBuilder",
		"System.exit",
		"java.io.File",
		"java.nio.file",
		"java.net",
		"Thread.sleep",
		"Executors",
		"ForkJoinPool",
	}

	for _, bad := range blocked {
		if strings.Contains(code, bad) {
			return fmt.Errorf("code contains forbidden keyword: %s", bad)
		}
	}

	return nil
}
