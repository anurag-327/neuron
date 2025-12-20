package fileUtils

import (
	"fmt"
	"os"
	"path/filepath"
)

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
