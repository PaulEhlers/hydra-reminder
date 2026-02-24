# HydraReminder

A minimalist, frictionless tray application that reminds you to stand up or drink water. Runs on **Windows** and **Linux**.

## Features
- **Zero Distractions**: No popups, no sounds, no modal windows. Alerts use a simple red icon and optional blinking.
- **Minimal Footprint**: Lightweight tray application with ~0% CPU and <15MB RAM usage.
- **Smart UI**: Single-click the tray icon to check time/reset, right-click to configure durations natively.
- **Global Hotkey**: Press `Modifier + <Key>` to instantly reset your active timer from anywhere (configurable prefixes like `CTRL+SHIFT`).
- **Auto-Start**: Integrates with the OS to start on boot (Windows Registry / Linux XDG autostart).

## Platform Support

| Feature               | Windows       | Linux (X11)         |
|-----------------------|---------------|---------------------|
| Tray Icon & Menu      | ✅ Native      | ✅ libayatana        |
| Global Hotkeys        | ✅ Win32 API   | ✅ libX11            |
| Autostart             | ✅ Registry    | ✅ XDG `.desktop`    |
| Click-to-Reset        | ✅             | ✅ `dbus-monitor`    |

> **Note:** Linux support requires an X11 session. Wayland environments will block background global hotkeys. 
> Ensure you have a system tray or AppIndicator extension enabled (e.g., for GNOME).

## Developer Build Requirements

- **Go 1.22+**

### Windows
```bash
go mod tidy
go build -ldflags="-H=windowsgui" -o hydra-reminder.exe ./cmd/hydra-reminder
```

### Linux
Requires GCC and development headers for GTK3, libX11, and libayatana.

**Ubuntu / Debian**
```bash
sudo apt install -y gcc libgtk-3-dev libayatana-appindicator3-dev libx11-dev
```

**Arch Linux / CachyOS / Manjaro**
```bash
sudo pacman -S gcc gtk3 libayatana-appindicator libx11
```

**Build**
```bash
go mod tidy
go build -o hydra-reminder ./cmd/hydra-reminder
```

## Open Source Dependencies & Licenses

- **[github.com/getlantern/systray](https://github.com/getlantern/systray)**: Apache License 2.0. Cross-platform tray icon and menu.
- **[golang.org/x/sys](https://pkg.go.dev/golang.org/x/sys)**: BSD 3-Clause. Windows API access.
- **[golang.design/x/hotkey](https://pkg.go.dev/golang.design/x/hotkey)**: MIT. Cross-platform global hotkeys (Linux/macOS).
- Several indirect supporting libraries from the `getlantern` ecosystem under MIT / Apache 2.0.
