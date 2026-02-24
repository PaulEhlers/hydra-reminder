//go:build linux || darwin

package tray

import (
	"bufio"
	"os/exec"
	"strings"

	"hydra-reminder/internal/timer"
)

// monitorMenuOpen on Linux uses dbus-monitor to detect when the user interacts
// with the tray app indicator (e.g. clicks it to open the menu).
func (t *TrayApp) monitorMenuOpen() {
	// Monitor the StatusNotifierItem interface for interaction methods
	// Use stdbuf -oL to prevent block buffering when piping stdout.
	cmd := exec.Command("stdbuf", "-oL", "dbus-monitor", "--session", "interface='org.kde.StatusNotifierItem'")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err := cmd.Start(); err != nil {
		return
	}
	defer cmd.Process.Kill()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		// The host environment calls these methods when the user interacts with the tray icon.
		if strings.Contains(line, "member=AboutToShow") ||
			strings.Contains(line, "member=ContextMenu") ||
			strings.Contains(line, "member=Activate") ||
			strings.Contains(line, "member=SecondaryActivate") ||
			strings.Contains(line, "member=SecondaryClick") ||
			strings.Contains(line, "member=Scroll") {

			if t.timerManager != nil && t.timerManager.GetState() == timer.StateAlerting {
				t.timerManager.Reset()
			}
		}
	}
}
