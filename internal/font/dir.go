package font

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// TargetDir returns the destination directory where Quran fonts should be installed.
// If override is provided and non-empty, it is returned directly.
func TargetDir(override string) (string, error) {
	if override != "" {
		return filepath.Clean(override), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine user home directory: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Fonts"), nil
	case "windows":
		appData := os.Getenv("LOCALAPPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(appData, "Microsoft", "Windows", "Fonts"), nil
	default:
		// Linux / BSD and other Unix-like systems
		dataHome := os.Getenv("XDG_DATA_HOME")
		if dataHome == "" {
			dataHome = filepath.Join(home, ".local", "share")
		}
		return filepath.Join(dataHome, "fonts", "quran"), nil
	}
}
