// +build !windows

package config

import (
	"os"
	"path/filepath"
)

func ConfigHomeFor(module string) string {
	if xdgPath := os.Getenv("XDG_CONFIG_HOME"); xdgPath != "" {
		return filepath.Join(xdgPath, module)
	}

	return filepath.Join(os.Getenv("HOME"), ".config", module)
}
