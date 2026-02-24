//go:build linux || darwin

package tray

// monitorMenuOpen on Linux/macOS doesn't easily have a global "menu opened" hook
// like Windows #32768 class. The getlantern/systray library doesn't expose
// a cross-platform 'menu opened' event.
// We'll leave this as a no-op on non-Windows for now, meaning the "auto reset when alerting
// and menu opened" feature is Windows-only. Users on Linux can still click the reset button.
func (t *TrayApp) monitorMenuOpen() {
	// No-op on linux/macOS
}
