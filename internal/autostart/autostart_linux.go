//go:build linux

package autostart

import (
	"fmt"
	"os"
	"path/filepath"
)

const appName = "hydra-reminder"

func getAutostartPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "autostart", appName+".desktop"), nil
}

func Enable() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return err
	}

	autostartPath, err := getAutostartPath()
	if err != nil {
		return err
	}

	// Ensure the autostart directory exists
	autostartDir := filepath.Dir(autostartPath)
	if err := os.MkdirAll(autostartDir, 0755); err != nil {
		return err
	}

	desktopFileContent := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=Hydra Reminder
Comment=Hydration and Activity Reminder
Exec=%s
Icon=utilities-terminal
Terminal=false
Categories=Utility;
`, exePath)

	return os.WriteFile(autostartPath, []byte(desktopFileContent), 0644)
}

func Disable() error {
	autostartPath, err := getAutostartPath()
	if err != nil {
		return err
	}

	err = os.Remove(autostartPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func IsEnabled() (bool, error) {
	autostartPath, err := getAutostartPath()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(autostartPath)
	if err == nil {
		// File exists
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
