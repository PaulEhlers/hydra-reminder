package main

import (
	"log"
	"os"

	"hydra-reminder/internal/config"
	"hydra-reminder/internal/hotkey"
	"hydra-reminder/internal/timer"
	"hydra-reminder/internal/tray"

	_ "embed"
)



func main() {
	log.SetOutput(os.Stderr)

	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		cfg = config.DefaultConfig()
	}

	app := tray.NewApp(cfg)

	tm := timer.NewManager(
		func() {
			app.OnRunning()
		},
		func() {
			app.OnAlert()
		},
		func() {
			app.OnStop()
		},
	)

	// Since app needs the timer manager, we can set it. We'll modify tray slightly or access the field if exported,
	// but it isn't exported. We can just add a SetTimerManager method, or we bypass that by defining a cyclic init.
	// Actually, wait, tray internal doesn't export timerManager.
	// Let's create an Initialization method or we just modify tray package here by calling an exported Set method.
	// But it's easier to just initialize them properly.
	// Let's modify tray.NewApp to not take tm, but add SetTimerManager

	hotkey.Init(func() {
		tm.Reset()
	})

	app.SetTimerManager(tm)
	app.Run(iconStopped, iconRunning, iconAlert)
}
