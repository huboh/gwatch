//go:build !windows

package config

import (
	"path/filepath"
)

var (
	// defaultBinPath defines the path to built binary
	defaultBinPath = filepath.Join(rootDir, "bin", "main")
)
