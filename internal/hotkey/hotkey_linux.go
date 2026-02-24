//go:build linux

package hotkey

import (
	"log"
	"sync"

	"golang.design/x/hotkey"
)

var (
	mu             sync.Mutex
	currentHotkey  *hotkey.Hotkey
	hotkeyCallback func()
)

// Init sets the callback function to be executed when the hotkey is triggered.
func Init(callback func()) {
	hotkeyCallback = callback
}

// mapWindowsModifiers converts Windows modifier bitmask to golang.design/x/hotkey Linux modifiers.
// Windows: MOD_ALT=0x0001, MOD_CONTROL=0x0002, MOD_SHIFT=0x0004, MOD_WIN=0x0008
// Linux (X11): ModCtrl=(1<<2), ModShift=(1<<0), Mod1=(1<<3) for Alt, Mod4=(1<<6) for Super
func mapWindowsModifiers(modifiers uint32) []hotkey.Modifier {
	var mods []hotkey.Modifier

	if modifiers&0x0001 != 0 { // MOD_ALT
		mods = append(mods, hotkey.Mod1) // X11 Mod1 = Alt
	}
	if modifiers&0x0002 != 0 { // MOD_CONTROL
		mods = append(mods, hotkey.ModCtrl)
	}
	if modifiers&0x0004 != 0 { // MOD_SHIFT
		mods = append(mods, hotkey.ModShift)
	}
	if modifiers&0x0008 != 0 { // MOD_WIN
		mods = append(mods, hotkey.Mod4) // X11 Mod4 = Super/Win
	}

	return mods
}

// mapWindowsVK converts a Windows virtual key code (A-Z: 0x41-0x5A) to an X11 keysym.
// X11 keysyms for a-z are 0x0061-0x007a (lowercase ASCII).
func mapWindowsVK(vk uint32) hotkey.Key {
	if vk >= 0x41 && vk <= 0x5A {
		// Convert uppercase VK code to lowercase X11 keysym
		return hotkey.Key(vk + 0x20) // 'A'(0x41) -> 'a'(0x61)
	}
	// Fallback: pass through (works for numbers and some special keys)
	return hotkey.Key(vk)
}

// Register registers a system-wide hotkey.
func Register(modifiers uint32, key uint32) error {
	mu.Lock()
	defer mu.Unlock()

	// Unregister any existing hotkey
	if currentHotkey != nil {
		if err := currentHotkey.Unregister(); err != nil {
			log.Printf("Failed to unregister previous hotkey: %v", err)
		}
		currentHotkey = nil
	}

	mods := mapWindowsModifiers(modifiers)
	xkey := mapWindowsVK(key)

	hk := hotkey.New(mods, xkey)

	if err := hk.Register(); err != nil {
		return err
	}

	currentHotkey = hk

	// Start listening for hotkey events
	go func() {
		for {
			_, ok := <-hk.Keydown()
			if !ok {
				return
			}
			if hotkeyCallback != nil {
				go hotkeyCallback()
			}
		}
	}()

	return nil
}

// Unregister unregisters the currently registered hotkey.
func Unregister() {
	mu.Lock()
	defer mu.Unlock()

	if currentHotkey != nil {
		if err := currentHotkey.Unregister(); err != nil {
			log.Printf("Failed to unregister hotkey: %v", err)
		}
		currentHotkey = nil
	}
}
