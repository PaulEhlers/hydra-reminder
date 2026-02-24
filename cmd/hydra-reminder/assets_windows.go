//go:build windows || darwin

package main

import _ "embed"

//go:embed assets/icon_stopped.ico
var iconStopped []byte

//go:embed assets/icon_running.ico
var iconRunning []byte

//go:embed assets/icon_alert.ico
var iconAlert []byte
