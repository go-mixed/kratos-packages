package utils

import (
	"os"
	"path/filepath"
)

// GetCurrentDir get current dir of the executable
func GetCurrentDir() string {
	path, _ := os.Executable()
	return filepath.Dir(path)
}
