package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// GetCurrentDir get current dir of the executable
func GetCurrentDir() string {
	path, _ := os.Executable()
	return filepath.Dir(path)
}

// PathToURI Path to file uri
func PathToURI(path string) string {
	// convert to the absolute path
	if !filepath.IsAbs(path) {
		if _path, err := filepath.Abs(path); err == nil {
			path = _path
		}
	}

	// convert to unix style
	path = filepath.ToSlash(path)
	// convert a windows' driver to lower case, and remove the colon: "C:" -> "/c"
	if len(path) >= 2 && path[1] == ':' && ((path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z')) {
		path = "/" + strings.ToLower(path[0:1]) + path[2:]
	}

	return "file://" + path
}
