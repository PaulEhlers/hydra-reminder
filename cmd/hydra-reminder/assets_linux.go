//go:build linux

package main

import _ "embed"

//go:embed assets/icon_stopped.png
var iconStopped []byte

//go:embed assets/icon_running.png
var iconRunning []byte

//go:embed assets/icon_alert.png
var iconAlert []byte
