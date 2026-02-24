package tray

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/getlantern/systray"

	"hydra-reminder/internal/autostart"
	"hydra-reminder/internal/config"
	"hydra-reminder/internal/hotkey"
	"hydra-reminder/internal/timer"
)

var (
	IconStopped []byte
	IconRunning []byte
	IconAlert   []byte
)

type TrayApp struct {
	cfg          *config.Config
	timerManager *timer.Manager
	blinkTicker  *time.Ticker
	blinkDone    chan struct{}
	uiChan       chan func() // channel to serialize UI updates
	iconIsAlert  bool
	timeTicker   *time.Ticker
	timeItem     *systray.MenuItem
}

func NewApp(cfg *config.Config) *TrayApp {
	return &TrayApp{
		cfg:    cfg,
		uiChan: make(chan func(), 10),
	}
}

func (t *TrayApp) SetTimerManager(tm *timer.Manager) {
	t.timerManager = tm
}

func (t *TrayApp) Run(iconStopped, iconRunning, iconAlert []byte) {
	IconStopped = iconStopped
	IconRunning = iconRunning
	IconAlert = iconAlert

	systray.Run(t.onReady, t.onExit)
}

func (t *TrayApp) saveConfig() {
	if err := config.Save(t.cfg); err != nil {
		log.Printf("Failed to save config: %v", err)
	}
}

func (t *TrayApp) onReady() {
	systray.SetIcon(IconStopped)
	systray.SetTitle("HydraReminder - Stopped")
	systray.SetTooltip("HydraReminder - Stopped")

	// Set up menus
	mStartStop := systray.AddMenuItem("Start / Stop", "Toggle timer")
	mReset := systray.AddMenuItem("Reset Timer", "Reset timer to current duration")

	systray.AddSeparator()

	t.timeItem = systray.AddMenuItem("Time Remaining: --:--", "Time left before alert")
	t.timeItem.Disable()

	systray.AddSeparator()

	mDuration := systray.AddMenuItem("Timer Duration", "Set timer duration")
	mDir10s := mDuration.AddSubMenuItemCheckbox("10 seconds (Debug)", "", t.cfg.DurationMinutes == 0) // We'll map duration 0 to 10s
	mDir15 := mDuration.AddSubMenuItemCheckbox("15 min", "", t.cfg.DurationMinutes == 15)
	mDir30 := mDuration.AddSubMenuItemCheckbox("30 min", "", t.cfg.DurationMinutes == 30)
	mDir45 := mDuration.AddSubMenuItemCheckbox("45 min", "", t.cfg.DurationMinutes == 45)
	mDir60 := mDuration.AddSubMenuItemCheckbox("60 min", "", t.cfg.DurationMinutes == 60)

	systray.AddSeparator()

	mStyle := systray.AddMenuItemCheckbox("Blink Mode", "Toggle icon blink on alert", t.cfg.AlertStyle == "blink")

	systray.AddSeparator()

	mHotkeyMenu := systray.AddMenuItem("Global Hotkeys", "Configure hotkeys")
	
	mPrefixMenu := mHotkeyMenu.AddSubMenuItem("Prefix Shortcut...", "")
	mPrefCtrlAlt := mPrefixMenu.AddSubMenuItemCheckbox("CTRL + ALT", "", t.cfg.HotkeyModifiers == 0x0003)
	mPrefCtrlShift := mPrefixMenu.AddSubMenuItemCheckbox("CTRL + SHIFT", "", t.cfg.HotkeyModifiers == 0x0006)
	mPrefSuperShift := mPrefixMenu.AddSubMenuItemCheckbox("SUPER + SHIFT", "", t.cfg.HotkeyModifiers == 0x000C)

	mHotkeyEnable := mHotkeyMenu.AddSubMenuItemCheckbox("Enable Hotkey", "", t.cfg.HotkeyEnabled)

	mSetReset := mHotkeyMenu.AddSubMenuItem("Set Reset Key...", "")

	var resetItems []*systray.MenuItem

	for c := 'A'; c <= 'Z'; c++ {
		charStr := string(c)

		ri := mSetReset.AddSubMenuItemCheckbox(charStr, "", t.cfg.HotkeyResetKey == uint32(c))
		resetItems = append(resetItems, ri)
	}

	enabled, _ := autostart.IsEnabled()
	// Update config to match reality in case registry differs from config
	t.cfg.Autostart = enabled
	t.saveConfig()

	mAutostart := systray.AddMenuItemCheckbox("Enable Autostart", "Run on Windows startup", enabled)

	systray.AddSeparator()

	mHelp := systray.AddMenuItem("Help", "How to use HydraReminder")
	mHelpState := mHelp.AddSubMenuItem("States: Grey=Stopped, Green=Running, Red=Alert", "")
	mHelpState.Disable()
	mHelpReset := mHelp.AddSubMenuItem("Reset: Click tray icon or use Reset Timer", "")
	mHelpReset.Disable()
	mHelpBlink := mHelp.AddSubMenuItem("Blink Mode: Flashes icon red/green when alert triggers", "")
	mHelpBlink.Disable()
	mHelpHotkey := mHelp.AddSubMenuItem("Hotkeys: Start with CTRL+ALT+<Key> (configurable)", "")
	mHelpHotkey.Disable()

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit HydraReminder")

	// Start a goroutine to process UI actions linearly
	go func() {
		for action := range t.uiChan {
			action()
		}
	}()

	var lastRadioChange time.Time

	// Event handling
	go func() {
		for {
			select {
			case <-mStartStop.ClickedCh:
				t.timerManager.Toggle()
			case <-mReset.ClickedCh:
				t.timerManager.Reset()
			case <-mDir10s.ClickedCh:
				t.handleDurationClick(&lastRadioChange, 0, mDir10s, mDir15, mDir30, mDir45, mDir60)
			case <-mDir15.ClickedCh:
				t.handleDurationClick(&lastRadioChange, 15, mDir10s, mDir15, mDir30, mDir45, mDir60)
			case <-mDir30.ClickedCh:
				t.handleDurationClick(&lastRadioChange, 30, mDir10s, mDir15, mDir30, mDir45, mDir60)
			case <-mDir45.ClickedCh:
				t.handleDurationClick(&lastRadioChange, 45, mDir10s, mDir15, mDir30, mDir45, mDir60)
			case <-mDir60.ClickedCh:
				t.handleDurationClick(&lastRadioChange, 60, mDir10s, mDir15, mDir30, mDir45, mDir60)
			case <-mStyle.ClickedCh:
				if t.cfg.AlertStyle == "color" {
					t.cfg.AlertStyle = "blink"
					mStyle.Check()
				} else {
					t.cfg.AlertStyle = "color"
					mStyle.Uncheck()
				}
				t.saveConfig()
			case <-mPrefCtrlAlt.ClickedCh:
				t.setHotkeyModifier(&lastRadioChange, 0x0003, mPrefCtrlAlt, mPrefCtrlShift, mPrefSuperShift)
			case <-mPrefCtrlShift.ClickedCh:
				t.setHotkeyModifier(&lastRadioChange, 0x0006, mPrefCtrlAlt, mPrefCtrlShift, mPrefSuperShift)
			case <-mPrefSuperShift.ClickedCh:
				t.setHotkeyModifier(&lastRadioChange, 0x000C, mPrefCtrlAlt, mPrefCtrlShift, mPrefSuperShift)
			case <-mHotkeyEnable.ClickedCh:
				t.cfg.HotkeyEnabled = !t.cfg.HotkeyEnabled
				if t.cfg.HotkeyEnabled {
					mHotkeyEnable.Check()
					hotkey.Register(t.cfg.HotkeyModifiers, t.cfg.HotkeyResetKey)
				} else {
					mHotkeyEnable.Uncheck()
					hotkey.Unregister()
				}
				t.saveConfig()
			case <-mAutostart.ClickedCh:
				t.cfg.Autostart = !t.cfg.Autostart
				if t.cfg.Autostart {
					mAutostart.Check()
					autostart.Enable()
				} else {
					mAutostart.Uncheck()
					autostart.Disable()
				}
				t.saveConfig()
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()

	var hotkeyMu sync.Mutex
	var lastHotkeyChange time.Time

	// Handle dynamic hotkey menus
	for i, item := range resetItems {
		go func(index int, mi *systray.MenuItem) {
			for range mi.ClickedCh {
				hotkeyMu.Lock()
				if time.Since(lastHotkeyChange) < 150*time.Millisecond {
					hotkeyMu.Unlock()
					continue
				}
				lastHotkeyChange = time.Now()
				hotkeyMu.Unlock()

				t.cfg.HotkeyResetKey = uint32('A' + index)
				t.saveConfig()

				for _, other := range resetItems {
					if other != mi {
						other.Uncheck()
					}
				}
				mi.Check()

				if t.cfg.HotkeyEnabled {
					hotkey.Register(t.cfg.HotkeyModifiers, t.cfg.HotkeyResetKey)
				}
			}
		}(i, item)
	}

	// Ticker for updating Time Remaining UI
	t.timeTicker = time.NewTicker(time.Second)
	go func() {
		for range t.timeTicker.C {
			if t.timerManager == nil {
				continue
			}
			state := t.timerManager.GetState()
			switch state {
			case timer.StateStopped:
				t.timeItem.SetTitle("Time Remaining: Stopped")
			case timer.StateAlerting:
				t.timeItem.SetTitle("Time Remaining: 00:00 (Alert!)")
			default:
				rem := t.timerManager.TimeRemaining()
				mins := int(rem.Minutes())
				secs := int(rem.Seconds()) % 60
				t.timeItem.SetTitle(fmt.Sprintf("Time Remaining: %02d:%02d", mins, secs))
			}
		}
	}()

	// Start polling to detect menu open for auto-reset
	go t.monitorMenuOpen()

	// Left click hook. In getlantern/systray, left-clicking icon emits an event on some channels,
	// but currently it does not have a dedicated LeftClick action on Windows.
	// Oh actually systray doesn't natively expose left vs right mouse click handling uniformly.
	// That's fine, the standard menu actions are enough for a minimalist application.

	// Register initial hotkeys
	if t.cfg.HotkeyEnabled {
		err := hotkey.Register(t.cfg.HotkeyModifiers, t.cfg.HotkeyResetKey)
		if err != nil {
			log.Printf("Failed to register hotkeys: %v", err)
		}
	}

	// Start timer immediately based on config, or leave stopped if previously stopped,
	// but the project requirements said: "Start timer immediately based on config".
	// Since we now have "Stopped" state, we should actually explicitly start it so it goes green.
	var dur time.Duration
	if t.cfg.DurationMinutes == 0 {
		dur = 10 * time.Second
	} else {
		dur = time.Duration(t.cfg.DurationMinutes) * time.Minute
	}

	// Start it immediately
	t.timerManager.Start(dur)
	// Make sure UI updates to running
	t.OnRunning()
}

func (t *TrayApp) handleDurationClick(lastChange *time.Time, mins int, items ...*systray.MenuItem) {
	if time.Since(*lastChange) < 150*time.Millisecond {
		return
	}
	*lastChange = time.Now()
	t.setDuration(mins, items...)
}

func (t *TrayApp) setDuration(mins int, items ...*systray.MenuItem) {
	for i, item := range items {
		opts := []int{0, 15, 30, 45, 60}
		if opts[i] == mins {
			item.Check()
		} else {
			item.Uncheck()
		}
	}
	t.cfg.DurationMinutes = mins
	t.saveConfig()

	var dur time.Duration
	if mins == 0 {
		dur = 10 * time.Second
	} else {
		dur = time.Duration(mins) * time.Minute
	}

	t.timerManager.Start(dur)
	t.OnRunning()
}

func (t *TrayApp) setHotkeyModifier(lastChange *time.Time, modifier uint32, items ...*systray.MenuItem) {
	if time.Since(*lastChange) < 150*time.Millisecond {
		return
	}
	*lastChange = time.Now()

	opts := []uint32{0x0003, 0x0006, 0x000C}
	for i, item := range items {
		if opts[i] == modifier {
			item.Check()
		} else {
			item.Uncheck()
		}
	}

	t.cfg.HotkeyModifiers = modifier
	t.saveConfig()
	if t.cfg.HotkeyEnabled {
		hotkey.Register(t.cfg.HotkeyModifiers, t.cfg.HotkeyResetKey)
	}
}

func (t *TrayApp) OnRunning() {
	t.uiChan <- func() {
		t.stopBlinking()
		systray.SetIcon(IconRunning)
		systray.SetTooltip("HydraReminder - Running")
	}
}

func (t *TrayApp) OnAlert() {
	t.uiChan <- func() {
		if t.cfg.AlertStyle == "blink" {
			t.startBlinking()
		} else {
			// Just swap color
			systray.SetIcon(IconAlert)
			systray.SetTooltip("Stand Up / Drink Water!")
		}
	}
}

func (t *TrayApp) OnStop() {
	t.uiChan <- func() {
		t.stopBlinking()
		systray.SetIcon(IconStopped)
		systray.SetTooltip("HydraReminder - Stopped")
	}
}

func (t *TrayApp) startBlinking() {
	t.stopBlinking() // Ensure any existing is stopped
	t.blinkTicker = time.NewTicker(500 * time.Millisecond)
	t.blinkDone = make(chan struct{})
	t.iconIsAlert = true
	systray.SetIcon(IconAlert)
	systray.SetTooltip("Stand Up / Drink Water!")

	go func() {
		for {
			select {
			case <-t.blinkDone:
				return
			case <-t.blinkTicker.C:
				t.uiChan <- func() {
					if t.iconIsAlert {
						systray.SetIcon(IconRunning)
					} else {
						systray.SetIcon(IconAlert)
					}
					t.iconIsAlert = !t.iconIsAlert
				}
			}
		}
	}()
}

func (t *TrayApp) stopBlinking() {
	if t.blinkTicker != nil {
		t.blinkTicker.Stop()
		t.blinkTicker = nil
	}
	if t.blinkDone != nil {
		close(t.blinkDone)
		t.blinkDone = nil
	}
}

func (t *TrayApp) onExit() {
	hotkey.Unregister()
	os.Exit(0)
}
