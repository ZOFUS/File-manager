package fs

import (
	"errors"
	"path/filepath"
	"secure-fm/config"
	"strings"
)

var BaseDir string

func InitFS(cfg *config.Config) {
	BaseDir = cfg.SandboxPath
	// Ensure BaseDir is absolute
	abs, err := filepath.Abs(BaseDir)
	if err == nil {
		BaseDir = abs
	}
}

// ResolvePath ensures the path is within the sandbox
func ResolvePath(userPath string) (string, error) {
	// Clean the path to remove .. and .
	// cleanedPath := filepath.Clean(userPath) // Unused

	// If it's a relative path, join it with BaseDir
	// But we need to be careful. If user provides "../something", Join might keep it relative if we are not careful.
	// Best strategy: Join BaseDir + userPath, then Clean, then check prefix.

	fullPath := filepath.Join(BaseDir, userPath)
	cleanedFullPath := filepath.Clean(fullPath)

	// Check if it starts with BaseDir
	if !strings.HasPrefix(cleanedFullPath, BaseDir) {
		return "", errors.New("access denied: path traversal attempt")
	}

	return cleanedFullPath, nil
}
