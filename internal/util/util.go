package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"go.mongodb.org/mongo-driver/bson/primitive"
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

func WriteContentToFile(path string, content []byte, permission os.FileMode) error {
	if err := os.WriteFile(path, content, permission); err != nil {
		return err
	}
	return nil
}

func DeleteFolder(path string) error {
	if path == "" || path == "/" {
		return fmt.Errorf("unsafe delete path: %q", path)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	return os.RemoveAll(abs)
}

func IsValidObjectID(id string) (primitive.ObjectID, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	return objectId, err
}
