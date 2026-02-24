//go:build windows

package tray

import (
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"hydra-reminder/internal/timer"
)

func (t *TrayApp) monitorMenuOpen() {
	user32 := windows.NewLazySystemDLL("user32.dll")
	findWindow := user32.NewProc("FindWindowW")
	getWindowThreadProcessId := user32.NewProc("GetWindowThreadProcessId")

	myPid := uint32(windows.GetCurrentProcessId())
	var lastOpen bool

	for {
		time.Sleep(100 * time.Millisecond)

		// #32768 is the standard window class for menus
		hwnd, _, _ := findWindow.Call(
			uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("#32768"))),
			0,
		)

		if hwnd != 0 {
			var pid uint32
			getWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

			// Check if the menu window belongs to our process
			if pid == myPid {
				if !lastOpen {
					lastOpen = true
					// User opened menu, reset the timer ONLY if alerting
					if t.timerManager != nil && t.timerManager.GetState() == timer.StateAlerting {
						t.timerManager.Reset()
					}
				}
			} else {
				lastOpen = false
			}
		} else {
			lastOpen = false
		}
	}
}
