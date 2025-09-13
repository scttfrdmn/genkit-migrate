package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func IsGoFile(path string) bool {
	return strings.HasSuffix(path, ".go")
}

func IsGenKitFile(content string) bool {
	return strings.Contains(content, "genkit") ||
		strings.Contains(content, "firebase/genkit")
}

func GetRelativePath(basePath, fullPath string) (string, error) {
	return filepath.Rel(basePath, fullPath)
}

func WriteFileWithDir(path, content string) error {
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func CopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", src, err)
	}

	return WriteFileWithDir(dst, string(data))
}
