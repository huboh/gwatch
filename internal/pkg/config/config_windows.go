//go:build windows

package config

import "path"

var (
	// defaultBinPath defines the path to built binary
	defaultBinPath = path.Join(rootDir, "bin", "main.exe")
)
