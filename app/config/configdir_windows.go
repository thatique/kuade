// +build windows
package config

import (
	"os"
	"path/filepath"
)

func ConfigHomeFor(module string) string {
	configHome := os.Getenv("LOCALAPPDATA")
	if configHome == "" {
		// try APPDATA
		configHome = os.Getenv("APPDATA")
		if configHome == "" {
			// If still empty, use the default path
			userName := os.Getenv("USERNAME")
			configHome = filepath.Join("C:/", "Users", userName, "AppData", "Local")
		}
	}

	return filepath.Join(configHome, module, "config")
}
